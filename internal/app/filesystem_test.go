package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Velarvo/velarvo-desktop/internal/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListDirectoryReturnsSortedExplorerNodes(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	folderPath := filepath.Join(dir, "Documents")
	readmePath := filepath.Join(dir, "README.md")
	archivePath := filepath.Join(dir, "archive.zip")

	require.NoError(t, os.Mkdir(folderPath, 0o755))
	require.NoError(t, os.WriteFile(readmePath, []byte("hello"), 0o644))
	require.NoError(t, os.WriteFile(archivePath, []byte("archive"), 0o644))

	nodes, err := (&app.App{}).ListDirectory(dir)
	require.NoError(t, err)
	require.Len(t, nodes, 3)

	assertNode(t, &nodes[0], &app.ExplorerNode{
		ID:   folderPath,
		Name: "Documents",
		Path: folderPath,
		Type: "directory",
	})

	assertNode(t, &nodes[1], &app.ExplorerNode{
		ID:   archivePath,
		Name: "archive.zip",
		Path: archivePath,
		Type: "file",
		Size: int64(len("archive")),
	})

	assertNode(t, &nodes[2], &app.ExplorerNode{
		ID:   readmePath,
		Name: "README.md",
		Path: readmePath,
		Type: "file",
		Size: int64(len("hello")),
	})
}

func TestListDirectoryReturnsErrorForMissingPath(t *testing.T) {
	t.Parallel()

	missingPath := filepath.Join(t.TempDir(), "missing")

	nodes, err := (&app.App{}).ListDirectory(missingPath)

	require.Error(t, err)
	assert.Nil(t, nodes)
}

func assertNode(t *testing.T, got, want *app.ExplorerNode) {
	t.Helper()

	assert.Equal(t, want.ID, got.ID)
	assert.Equal(t, want.Name, got.Name)
	assert.Equal(t, want.Path, got.Path)
	assert.Equal(t, want.Type, got.Type)
	assert.Equal(t, want.Size, got.Size)
	assert.Positive(t, got.Modified)
}
