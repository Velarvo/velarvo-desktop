package files

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Velarvo/velarvo-desktop/internal/apperr"
)

var (
	ErrTargetExists = apperr.New("FILESYSTEM_TARGET_EXISTS", "a file with that name already exists")
	ErrInvalidPath  = apperr.New("FILESYSTEM_INVALID_PATH", "invalid path")
)

const (
	TypeFile      = "file"
	TypeDirectory = "directory"
)

type Entry struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
}

func SortEntries(entries []Entry) {
	slices.SortFunc(entries, func(a, b Entry) int {
		if a.Type != b.Type {
			if a.Type == TypeDirectory {
				return -1
			}
			return 1
		}
		return cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
}

type Service struct{}

func (Service) Home() (string, error) {
	return os.UserHomeDir()
}

func (Service) List(dir string) ([]Entry, error) {
	dir = filepath.Clean(dir)
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(dirEntries))
	for _, dirEntry := range dirEntries {
		info, err := dirEntry.Info()
		if err != nil {
			continue
		}
		entry := Entry{
			Name:     dirEntry.Name(),
			Path:     filepath.Join(dir, dirEntry.Name()),
			Type:     TypeFile,
			Size:     info.Size(),
			Modified: info.ModTime().UnixMilli(),
		}
		if dirEntry.IsDir() {
			entry.Type, entry.Size = TypeDirectory, 0
		}
		entries = append(entries, entry)
	}
	SortEntries(entries)
	return entries, nil
}

func (Service) Preview(path string) (Preview, error) {
	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if err != nil {
		return Preview{}, err
	}
	if !info.Mode().IsRegular() {
		return Preview{}, ErrNotAFile
	}

	file, err := os.Open(path)
	if err != nil {
		return Preview{}, err
	}
	defer func() { _ = file.Close() }()

	return BuildPreview(filepath.Base(path), path, info.Size(), file)
}

func (Service) Rename(from, to string) error {
	from, to = filepath.Clean(from), filepath.Clean(to)
	if from == "." || to == "." {
		return ErrInvalidPath
	}
	if from == to {
		return nil
	}

	if _, err := os.Lstat(to); err == nil {
		return ErrTargetExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(from, to)
}

func (Service) Remove(path string) error {
	path = filepath.Clean(path)
	if path == "." || path == filepath.VolumeName(path)+string(filepath.Separator) {
		return ErrInvalidPath
	}
	if _, err := os.Lstat(path); err != nil {
		return err
	}
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("remove %s: %w", path, err)
	}
	return nil
}
