#!/usr/bin/env python3
"""docgate: the docs-first gate. Standard library only. Run with: python3 -I tools/docgate.py ...

pet-project is a kit (skills, standards, templates, this gate). The documents it produces
(ADRs, RFCs, specs, runbooks, benchmarks) live inside a TARGET PROJECT, registered in
targets.local.json (git-ignored, machine-specific).

Commands (optional first option: --project NAME, default is the active target)
  init NAME PATH [--code-paths a/,b/] [--no-baseline] [--no-vendor]
                                        Start a new project: create PATH/docs/ skeleton, seed two
                                        baseline ADRs (Proposed) into an empty docs/adr, register it as
                                        the active target, and vendor the kit into it (see vendor).
                                        It never edits the user's Claude Code permission settings.
  vendor [--force]                      Copy the kit's skills, gate, standards, templates, process doc and
                                        hook settings into the active target so it works on its own
                                        (also run by init; use it again to update). Managed files are
                                        overwritten; settings.json and CLAUDE.md only if missing or --force.
  list                                  Show registered targets (* = active)
  use NAME                              Make NAME the active target
  target                                Print the active target (name and path)
  status                                List the target's documents and active spec
  check                                 Validate the target's ADRs, RFCs, specs (exit 1 on problems)
  new <adr|rfc|spec> "<title>"          Scaffold the next numbered document in the target
  guard                                 Claude Code PreToolUse hook (JSON on stdin). Exit 2 blocks.
                                        Code-path edits inside a registered target need an
                                        Accepted active spec whose ADRs/RFCs are Accepted and
                                        the file must be inside the spec's scope.
"""
from __future__ import annotations

import datetime
import json
import os
import re
import sys
from pathlib import Path

FRAMEWORK = Path(os.environ.get("DOCGATE_FRAMEWORK") or Path(__file__).resolve().parent.parent).resolve()

DEFAULT_CODE_PATHS = ["src/", "internal/", "cmd/", "pkg/", "api/", "services/", "deploy/"]
DOC_DIRS = ["adr", "rfc", "specs", "runbooks", "alerts", "benchmarks", "work"]

KINDS = {
    "adr": {
        "dir": "docs/adr",
        "sections": ["Context", "Decision", "Alternatives considered", "Consequences", "Operational impact"],
    },
    "rfc": {
        "dir": "docs/rfc",
        "sections": [
            "Summary", "Motivation", "Detailed design", "Security", "Performance",
            "Observability and alerting", "Alternatives considered", "Rollout", "Open questions",
        ],
    },
    "spec": {
        "dir": "docs/specs",
        "sections": [
            "Goal", "Scope", "Acceptance criteria", "Domain model", "Security",
            "Performance budget", "Observability and alerting", "Test plan",
        ],
    },
}

STATUSES = {"Proposed", "Accepted", "Rejected", "Superseded", "Deprecated", "Implemented"}
APPROVED = {"Accepted", "Implemented"}
BASE_KEYS = ["id", "title", "status", "date"]
PLACEHOLDER = re.compile(r"\bTODO\b|\bTBD\b|\{\{|\}\}")


class GateError(Exception):
    pass


# ---------------------------------------------------------------- targets

def targets_file() -> Path:
    return Path(os.environ.get("DOCGATE_TARGETS") or FRAMEWORK / "targets.local.json")


STANDALONE_FILE = ".docgate.json"


def standalone_config():
    """A vendored project carries its own config (name and code paths) so it needs no registry."""
    f = FRAMEWORK / STANDALONE_FILE
    if not f.is_file():
        return None
    cfg = json.loads(f.read_text(encoding="utf-8"))
    name = cfg.get("name")
    if not name:
        raise GateError(f"{STANDALONE_FILE} needs a 'name'")
    return name, {"path": str(FRAMEWORK), "code_paths": cfg.get("code_paths") or DEFAULT_CODE_PATHS}


def load_targets() -> dict:
    f = targets_file()
    if not f.is_file():
        sa = standalone_config()
        if sa:
            return {"active": sa[0], "projects": {sa[0]: sa[1]}}
        return {"active": None, "projects": {}}
    data = json.loads(f.read_text(encoding="utf-8"))
    data.setdefault("active", None)
    data.setdefault("projects", {})
    return data


def save_targets(data: dict) -> None:
    targets_file().write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")


class Project:
    def __init__(self, name: str, entry: dict):
        self.name = name
        self.root = Path(entry["path"]).expanduser().resolve()
        self.code_paths = entry.get("code_paths") or DEFAULT_CODE_PATHS


def all_projects() -> list:
    return [Project(n, e) for n, e in load_targets()["projects"].items()]


def get_project(name: str | None = None) -> Project:
    data = load_targets()
    name = name or data["active"]
    if not name:
        raise GateError("no active target project. Run: python3 -I tools/docgate.py init <name> <path>")
    if name not in data["projects"]:
        raise GateError(f"unknown target '{name}'. Registered: {', '.join(data['projects']) or '(none)'}")
    return Project(name, data["projects"][name])


# ---------------------------------------------------------------- parsing

def parse(text: str):
    """Return (meta, body). meta is None when there is no frontmatter."""
    m = re.match(r"\A---\n(.*?)\n---\n?(.*)\Z", text, re.S)
    if not m:
        return None, text
    meta = {}
    for line in m.group(1).splitlines():
        if ":" in line and not line.lstrip().startswith("#"):
            key, value = line.split(":", 1)
            meta[key.strip()] = value.strip()
    return meta, m.group(2)


def strip_comments(body: str) -> str:
    return re.sub(r"<!--.*?-->", "", body, flags=re.S)


def sections(body: str) -> dict:
    out: dict = {}
    current = None
    for line in strip_comments(body).splitlines():
        m = re.match(r"##\s+(.+?)\s*$", line)
        if m:
            current = m.group(1)
            out[current] = []
        elif current is not None:
            out[current].append(line)
    return {k: "\n".join(v).strip() for k, v in out.items()}


def split_list(value: str) -> list:
    items = [x.strip() for x in (value or "").split(",")]
    return [x for x in items if x and x.lower() != "none"]


# ---------------------------------------------------------------- documents

def doc_files(kind: str, root: Path) -> list:
    d = root / KINDS[kind]["dir"]
    if not d.is_dir():
        return []
    return sorted(d.glob("[0-9][0-9][0-9][0-9]-*.md"))


def find_doc(kind: str, doc_id: str, root: Path):
    for p in doc_files(kind, root):
        if p.name.startswith(doc_id + "-"):
            return p
    return None


def check_alerting(section: str) -> list:
    """Alerting gate: SLIs/SLOs and alerts with runbook links, or an explicit reasoned 'not applicable'."""
    if re.search(r"Alerting:\s*not applicable\s*[—-]\s*\S.{18,}", section):
        return []
    problems = []
    if "SLI" not in section or "SLO" not in section:
        problems.append("Observability and alerting: define SLIs and SLOs (or write 'Alerting: not applicable — <reason>')")
    rows = [l for l in section.splitlines() if l.strip().startswith("|") and "runbooks/" in l]
    if not rows:
        problems.append("Observability and alerting: at least one alert row must link a runbook (docs/runbooks/...)")
    return problems


def validate(path: Path, kind: str) -> list:
    problems = []
    meta, body = parse(path.read_text(encoding="utf-8"))
    if meta is None:
        return ["missing frontmatter (--- block)"]
    keys = BASE_KEYS + (["scope"] if kind == "spec" else [])
    for key in keys:
        if not meta.get(key):
            problems.append(f"frontmatter: '{key}' is required")
    if kind == "spec" and "adrs" not in meta:
        problems.append("frontmatter: 'adrs' is required (comma list of ADR ids, or 'none')")
    if meta.get("id") and not path.name.startswith(meta["id"] + "-"):
        problems.append(f"frontmatter: id '{meta['id']}' does not match filename '{path.name}'")
    status = meta.get("status", "")
    if status and status not in STATUSES:
        problems.append(f"status '{status}' is not one of {sorted(STATUSES)}")
    if status == "Implemented" and kind != "spec":
        problems.append("status 'Implemented' is only valid for specs")
    secs = sections(body)
    for name in KINDS[kind]["sections"]:
        if not secs.get(name):
            problems.append(f"section '## {name}' is missing or empty")
    if status in APPROVED:
        for key in ("approved_by", "approved_date"):
            if not meta.get(key):
                problems.append(f"approved documents need frontmatter '{key}'")
        if PLACEHOLDER.search(strip_comments(body)):
            problems.append("approved document still contains TODO / TBD / {{placeholder}}")
        if kind in ("rfc", "spec"):
            problems.extend(check_alerting(secs.get("Observability and alerting", "")))
    return problems


# ---------------------------------------------------------------- active spec

def active_id(root: Path):
    f = root / "docs" / "work" / "ACTIVE"
    if not f.is_file():
        return None
    for line in f.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if line and not line.startswith("#"):
            return line
    return None


def why_blocked(project: Project, rel: str):
    """Return None when editing `rel` in `project` is allowed, otherwise the reason."""
    root = project.root
    spec_id = active_id(root)
    if not spec_id:
        return (f"no active spec in project '{project.name}'. Write one with /new-spec, get it Accepted by the "
                "user (/review-doc), then put its id in docs/work/ACTIVE of the project.")
    spec_path = find_doc("spec", spec_id, root)
    if spec_path is None:
        return f"docs/work/ACTIVE names spec '{spec_id}' but docs/specs/{spec_id}-*.md does not exist."
    problems = validate(spec_path, "spec")
    if problems:
        return f"spec {spec_id} is invalid: " + "; ".join(problems)
    meta, _ = parse(spec_path.read_text(encoding="utf-8"))
    if meta.get("status") != "Accepted":
        return f"spec {spec_id} has status '{meta.get('status')}', it must be 'Accepted' (approved by the user)."
    for kind, key in (("adr", "adrs"), ("rfc", "rfcs")):
        for ref in split_list(meta.get(key, "")):
            p = find_doc(kind, ref, root)
            if p is None:
                return f"spec {spec_id} links {kind.upper()} {ref}, which does not exist."
            m, _ = parse(p.read_text(encoding="utf-8"))
            if not m or m.get("status") not in APPROVED:
                return f"{kind.upper()} {ref} is '{(m or {}).get('status')}', it must be Accepted before implementation."
    scopes = split_list(meta.get("scope", ""))
    if not any(rel.startswith(s) for s in scopes):
        return f"'{rel}' is outside the scope of spec {spec_id} (scope: {', '.join(scopes)}). Amend the spec first."
    return None


# ---------------------------------------------------------------- commands

def cmd_guard() -> int:
    try:
        data = json.load(sys.stdin)
    except Exception:
        print("docgate: could not read hook input, not blocking", file=sys.stderr)
        return 0
    file_path = (data.get("tool_input") or {}).get("file_path")
    if not file_path:
        return 0
    p = Path(file_path)
    if not p.is_absolute():
        p = Path.cwd() / p
    p = p.resolve()
    try:
        projects = all_projects()
    except Exception as exc:  # unreadable targets file: fail closed outside the kit itself
        try:
            p.relative_to(FRAMEWORK)
            return 0
        except ValueError:
            print(f"docgate: BLOCKED edit of {p}\n  reason: targets file unreadable ({exc})", file=sys.stderr)
            return 2
    for project in projects:
        try:
            rel = p.relative_to(project.root).as_posix()
        except ValueError:
            continue
        if not any(rel.startswith(c) for c in project.code_paths):
            return 0
        try:
            reason = why_blocked(project, rel)
        except Exception as exc:  # fail closed for code paths
            reason = f"gate error: {exc}"
        if reason:
            print(f"docgate: BLOCKED edit of {project.name}:{rel}\n  reason: {reason}\n"
                  "  This is docs-first: see docs/process/workflow.md.", file=sys.stderr)
            return 2
        return 0
    return 0


def cmd_check(project: Project) -> int:
    bad = 0
    for kind in KINDS:
        for p in doc_files(kind, project.root):
            problems = validate(p, kind)
            rel = p.relative_to(project.root).as_posix()
            if problems:
                bad += 1
                print(f"FAIL {rel}")
                for pr in problems:
                    print(f"     - {pr}")
            else:
                print(f"ok   {rel}")
    spec_id = active_id(project.root)
    if spec_id and find_doc("spec", spec_id, project.root) is None:
        bad += 1
        print(f"FAIL docs/work/ACTIVE names unknown spec '{spec_id}'")
    return 1 if bad else 0


def cmd_status(project: Project) -> int:
    print(f"project: {project.name}  ({project.root})")
    for kind in KINDS:
        for p in doc_files(kind, project.root):
            meta, _ = parse(p.read_text(encoding="utf-8"))
            meta = meta or {}
            print(f"{kind:4} {meta.get('id', '????')} {meta.get('status', '?'):11} {meta.get('title', p.name)}")
    print(f"active spec: {active_id(project.root) or '(none)'}")
    return 0


def slugify(title: str) -> str:
    return re.sub(r"[^a-z0-9]+", "-", title.lower()).strip("-")[:60] or "untitled"


def cmd_new(project: Project, kind: str, title: str) -> int:
    if kind not in KINDS:
        print(f"unknown kind '{kind}', use adr, rfc or spec", file=sys.stderr)
        return 1
    template = project.root / "docs" / "templates" / f"{kind}.md"
    if not template.is_file():
        template = FRAMEWORK / "docs" / "templates" / f"{kind}.md"
    if not template.is_file():
        print(f"missing template for {kind}", file=sys.stderr)
        return 1
    out_dir = project.root / KINDS[kind]["dir"]
    out_dir.mkdir(parents=True, exist_ok=True)
    numbers = [int(p.name[:4]) for p in doc_files(kind, project.root)]
    doc_id = f"{(max(numbers) + 1) if numbers else 1:04d}"
    out = out_dir / f"{doc_id}-{slugify(title)}.md"
    text = (template.read_text(encoding="utf-8")
            .replace("{{id}}", doc_id)
            .replace("{{title}}", title)
            .replace("{{date}}", datetime.date.today().isoformat()))
    out.write_text(text, encoding="utf-8")
    print(out)
    return 0


def seed_baseline(root: Path) -> list:
    """Copy the kit's baseline ADRs into an empty docs/adr (never into one that already has ADRs)."""
    adr_dir = root / "docs" / "adr"
    src = FRAMEWORK / "docs" / "templates" / "baseline"
    if not src.is_dir() or any(adr_dir.glob("[0-9][0-9][0-9][0-9]-*.md")):
        return []
    created = []
    for t in sorted(src.glob("[0-9][0-9][0-9][0-9]-*.md")):
        text = t.read_text(encoding="utf-8").replace("{{date}}", datetime.date.today().isoformat())
        (adr_dir / t.name).write_text(text, encoding="utf-8")
        created.append(t.name)
    return created


VENDOR_FILES = [  # (source in kit, destination in project)
    ("tools/docgate.py", "tools/docgate.py"),
    ("docs/process/workflow.md", "docs/process/workflow.md"),
]
VENDOR_DIRS = [  # (source dir, destination dir, skip names)
    (".claude/skills", ".claude/skills", {"init-project"}),
    ("docs/standards", "docs/standards", set()),
]
VENDOR_TEMPLATES = ["adr.md", "rfc.md", "spec.md", "runbook.md"]


def localise(text: str) -> str:
    """In a vendored project the standards and process doc are local, not in 'pet-project'."""
    for old, new in (("pet-project `docs/", "`docs/"), ("pet-project docs/", "docs/"),
                     ("not into pet-project", "not into a shared kit"), ("pet-project/standards/", "standards/")):
        text = text.replace(old, new)
    return text


def vendor(project: Project, force: bool = False) -> list:
    """Copy the kit into the project so it is self-contained. Returns a list of notes."""
    import shutil
    root = project.root
    if root == FRAMEWORK:
        raise GateError("the target is the kit itself; nothing to vendor")
    notes = []

    def put(src: Path, dst: Path):
        dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(src, dst)

    for s, d in VENDOR_FILES:
        if (FRAMEWORK / s).is_file():
            put(FRAMEWORK / s, root / d)
            notes.append(d)
    for s, d, skip in VENDOR_DIRS:
        if not (FRAMEWORK / s).is_dir():
            continue
        for src in sorted((FRAMEWORK / s).rglob("*")):
            rel = src.relative_to(FRAMEWORK / s)
            if src.is_file() and rel.parts[0] not in skip and "__pycache__" not in rel.parts:
                if src.suffix == ".md" and s.startswith(".claude"):
                    dst = root / d / rel
                    dst.parent.mkdir(parents=True, exist_ok=True)
                    dst.write_text(localise(src.read_text(encoding="utf-8")), encoding="utf-8")
                else:
                    put(src, root / d / rel)
        notes.append(d + "/")
    for name in VENDOR_TEMPLATES:
        src = FRAMEWORK / "docs" / "templates" / name
        if src.is_file():
            put(src, root / "docs" / "templates" / name)
    notes.append("docs/templates/")
    # Files the project may have customised: only created when missing, unless --force.
    for s, d in ((".claude/settings.json", ".claude/settings.json"),
                 ("docs/templates/project-claude.md", "CLAUDE.md")):
        src, dst = FRAMEWORK / s, root / d
        if not src.is_file():
            continue
        if dst.exists() and not force:
            notes.append(f"{d} kept (exists; use --force to overwrite)")
        else:
            put(src, dst)
            notes.append(d)
    cfg = root / STANDALONE_FILE
    if not cfg.exists():
        cfg.write_text(json.dumps({"name": project.name, "code_paths": project.code_paths}, indent=2) + "\n",
                       encoding="utf-8")
        notes.append(STANDALONE_FILE)
    return notes


def cmd_init(name: str, path: str, code_paths: list | None, baseline: bool = True, do_vendor: bool = True) -> int:
    root = Path(path).expanduser().resolve()
    root.mkdir(parents=True, exist_ok=True)
    for d in DOC_DIRS:
        (root / "docs" / d).mkdir(parents=True, exist_ok=True)
    active = root / "docs" / "work" / "ACTIVE"
    if not active.exists():
        active.write_text(
            "# The id of the spec currently being implemented (for example: 0001).\n"
            "# Leave empty when none. The gate hook reads the first non-comment line.\n"
            "# Only set this after the spec and its ADRs are Accepted by the user.\n",
            encoding="utf-8",
        )
    seeded = seed_baseline(root) if baseline else []
    data = load_targets()
    entry = {"path": str(root)}
    if code_paths:
        entry["code_paths"] = code_paths
    data["projects"][name] = entry
    data["active"] = name
    save_targets(data)
    vendored = vendor(Project(name, entry)) if do_vendor else []
    print(f"registered '{name}' -> {root} (now the active target)")
    print(f"code paths guarded: {', '.join(code_paths or DEFAULT_CODE_PATHS)}")
    print(f"baseline ADRs: {', '.join(seeded) if seeded else 'none added'}")
    if vendored:
        print(f"vendored the kit into the project: {', '.join(vendored)}")
        print(f"the project now works on its own: cd {root} && claude  (skills, gate and standards are local)")
    else:
        print("access: this tool does not change Claude Code permissions. To let Claude edit the project from "
              f"the kit, start it with:  claude --add-dir {root}")
    print("next: /new-rfc, /new-adr or /new-spec, then /review-doc, then /implement")
    return 0


def cmd_list() -> int:
    data = load_targets()
    if not data["projects"]:
        print("no targets registered. Run: python3 -I tools/docgate.py init <name> <path>")
        return 0
    for n, e in data["projects"].items():
        mark = "*" if n == data["active"] else " "
        print(f"{mark} {n}\t{Path(e['path']).expanduser()}")
    return 0


def cmd_use(name: str) -> int:
    data = load_targets()
    if name not in data["projects"]:
        print(f"unknown target '{name}'", file=sys.stderr)
        return 1
    data["active"] = name
    save_targets(data)
    print(f"active target: {name}")
    return 0


def cmd_target() -> int:
    p = get_project()
    print(f"{p.name}\t{p.root}")
    return 0


def main(argv: list) -> int:
    args = argv[1:]
    project_name = None
    if len(args) >= 2 and args[0] == "--project":
        project_name, args = args[1], args[2:]
    if not args:
        print(__doc__)
        return 1
    cmd, rest = args[0], args[1:]
    try:
        if cmd == "guard":
            return cmd_guard()
        if cmd == "init" and len(rest) >= 2:
            code_paths = None
            if "--code-paths" in rest:
                i = rest.index("--code-paths")
                code_paths = split_list(rest[i + 1]) if i + 1 < len(rest) else None
            return cmd_init(rest[0], rest[1], code_paths, baseline="--no-baseline" not in rest,
                            do_vendor="--no-vendor" not in rest)
        if cmd in ("vendor", "update"):
            notes = vendor(get_project(project_name), force="--force" in rest)
            print("vendored: " + ", ".join(notes))
            return 0
        if cmd == "list":
            return cmd_list()
        if cmd == "use" and rest:
            return cmd_use(rest[0])
        if cmd == "target":
            return cmd_target()
        if cmd == "check":
            return cmd_check(get_project(project_name))
        if cmd == "status":
            return cmd_status(get_project(project_name))
        if cmd == "new" and len(rest) >= 2:
            return cmd_new(get_project(project_name), rest[0], " ".join(rest[1:]))
    except GateError as exc:
        print(f"docgate: {exc}", file=sys.stderr)
        return 1
    print(__doc__)
    return 1


if __name__ == "__main__":
    sys.exit(main(sys.argv))
