// Package archtest checks the dependency rule from the architecture standard:
// the domain imports no application or adapter code, and application code
// imports no adapter. CI runs it as an ordinary test.
package archtest

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type rule struct {
	dir         string
	includeTest bool
	forbidden   []string
}

func rules(module string) []rule {
	application := module + "/internal/application"
	adapters := module + "/internal/infrastructure"
	return []rule{
		{dir: "internal/domain/", includeTest: true, forbidden: []string{application, adapters}},
		// Application tests may use in-memory adapters as fakes of their own ports.
		{dir: "internal/application/", includeTest: false, forbidden: []string{adapters}},
	}
}

// Check returns one message per violating import, sorted by file path.
func Check(root, module string) ([]string, error) {
	var violations []string
	rs := rules(module)

	err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return fmt.Errorf("parse %s: %w", rel, err)
		}
		isTest := strings.HasSuffix(rel, "_test.go")

		for _, r := range rs {
			if !strings.HasPrefix(rel, r.dir) || (isTest && !r.includeTest) {
				continue
			}
			for _, imp := range file.Imports {
				target, err := strconv.Unquote(imp.Path.Value)
				if err != nil {
					return err
				}
				for _, banned := range r.forbidden {
					if target == banned || strings.HasPrefix(target, banned+"/") {
						violations = append(violations, fmt.Sprintf("%s imports %s", rel, target))
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(violations)
	return violations, nil
}
