package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type SavedFile struct {
	StoredName string
	Path       string
	Size       int64
}

type Storage interface {
	Save(context.Context, io.Reader, string) (*SavedFile, error)
	Open(context.Context, string) (io.ReadCloser, error)
	Delete(context.Context, string) error
}

type Local struct{ root string }

func NewLocal(root string) (*Local, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absolute, 0o750); err != nil {
		return nil, fmt.Errorf("create storage directory: %w", err)
	}
	return &Local{root: absolute}, nil
}

func (l *Local) Save(_ context.Context, source io.Reader, originalName string) (*SavedFile, error) {
	extension := strings.ToLower(filepath.Ext(filepath.Base(originalName)))
	storedName := uuid.NewString() + extension
	target := filepath.Join(l.root, storedName)
	if !isInside(l.root, target) {
		return nil, fmt.Errorf("invalid storage path")
	}
	file, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o640)
	if err != nil {
		return nil, err
	}
	size, copyErr := io.Copy(file, source)
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(target)
		return nil, copyErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return &SavedFile{StoredName: storedName, Path: target, Size: size}, nil
}

func (l *Local) Open(_ context.Context, path string) (io.ReadCloser, error) {
	cleaned, err := filepath.Abs(path)
	if err != nil || !isInside(l.root, cleaned) {
		return nil, fmt.Errorf("invalid storage path")
	}
	return os.Open(cleaned)
}

func (l *Local) Delete(_ context.Context, path string) error {
	cleaned, err := filepath.Abs(path)
	if err != nil || !isInside(l.root, cleaned) {
		return fmt.Errorf("invalid storage path")
	}
	if err := os.Remove(cleaned); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func isInside(root, target string) bool {
	relative, err := filepath.Rel(root, target)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
