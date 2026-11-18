BASE_BRANCH ?= master

.PHONY: check build vet test lint vuln secrets buf-lint buf-breaking

check: build vet test lint vuln secrets buf-lint

build:
	go build ./...
	CGO_ENABLED=0 go build ./...

vet:
	go vet ./...

# The dependency-rule check (internal/archtest) runs as part of the tests.
test:
	go test -race ./...

lint:
	golangci-lint run ./...

vuln:
	govulncheck ./...

secrets:
	gitleaks git --no-banner --redact .

buf-lint:
	buf lint

# Needs the base branch in the local clone (CI fetches full history).
buf-breaking:
	buf breaking --against ".git#branch=$(BASE_BRANCH)"
