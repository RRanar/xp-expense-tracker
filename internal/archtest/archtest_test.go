package archtest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RRanar/xp-expense-tracker/internal/archtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const module = "example.com/m"

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, src := range files {
		path := filepath.Join(root, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(src), 0o644))
	}
	return root
}

func TestDomainImportingApplicationOrAdapterIsAViolation(t *testing.T) {
	root := writeTree(t, map[string]string{
		"internal/domain/a/a.go": "package a\nimport _ \"" + module + "/internal/application/usecase\"\n",
		"internal/domain/b/b.go": "package b\nimport _ \"" + module + "/internal/infrastructure/memory\"\n",
		"internal/domain/c/c.go": "package c\nimport _ \"" + module + "/internal/domain/shared\"\nimport _ \"fmt\"\n",
	})

	got, err := archtest.Check(root, module)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Contains(t, got[0], "internal/domain/a/a.go")
	assert.Contains(t, got[1], "internal/domain/b/b.go")
}

func TestApplicationImportingAnAdapterIsAViolationButItsTestsMay(t *testing.T) {
	root := writeTree(t, map[string]string{
		"internal/application/usecase/u.go":      "package usecase\nimport _ \"" + module + "/internal/infrastructure/memory\"\n",
		"internal/application/usecase/u_test.go": "package usecase_test\nimport _ \"" + module + "/internal/infrastructure/memory\"\n",
	})

	got, err := archtest.Check(root, module)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Contains(t, got[0], "u.go")
	assert.NotContains(t, got[0], "u_test.go")
}

func TestInwardImportsAreAllowed(t *testing.T) {
	root := writeTree(t, map[string]string{
		"internal/application/usecase/u.go":   "package usecase\nimport _ \"" + module + "/internal/domain/expense\"\n",
		"internal/infrastructure/memory/m.go": "package memory\nimport _ \"" + module + "/internal/domain/expense\"\n",
		"internal/domain/expense/e.go":        "package expense\nimport _ \"" + module + "/internal/domain/shared\"\n",
	})

	got, err := archtest.Check(root, module)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestThisRepositoryRespectsTheDependencyRule(t *testing.T) {
	got, err := archtest.Check("../..", "github.com/RRanar/xp-expense-tracker")
	require.NoError(t, err)
	assert.Empty(t, got)
}
