package storage

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalStorageLifecycle(t *testing.T) {
	root := t.TempDir()
	local, err := NewLocal(root)
	require.NoError(t, err)
	saved, err := local.Save(context.Background(), bytes.NewBufferString("document"), "surat.pdf")
	require.NoError(t, err)
	require.Equal(t, ".pdf", filepath.Ext(saved.StoredName))
	file, err := local.Open(context.Background(), saved.Path)
	require.NoError(t, err)
	content, err := io.ReadAll(file)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	require.Equal(t, "document", string(content))
	require.NoError(t, local.Delete(context.Background(), saved.Path))
}

func TestLocalStorageRejectsPathTraversal(t *testing.T) {
	local, err := NewLocal(t.TempDir())
	require.NoError(t, err)
	_, err = local.Open(context.Background(), filepath.Join(local.root, "..", "secret.txt"))
	require.Error(t, err)
}
