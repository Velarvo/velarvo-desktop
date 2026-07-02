package app

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Velarvo/velarvo-desktop/internal/apperrors"
	"github.com/Velarvo/velarvo-desktop/internal/logger"
	"github.com/Velarvo/velarvo-desktop/internal/types"
	"go.uber.org/zap"
)

func filesystemLog() *zap.SugaredLogger {
	return logger.Named("filesystem")
}

type ExplorerNode struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"`
	Size     int64  `json:"size,omitempty"`
	Modified int64  `json:"modified"`
}

func getHomeDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		filesystemLog().Errorw("failed to resolve home directory", "error", err)
		return "", err
	}

	if home == "" {
		filesystemLog().Warn("home directory resolved to empty path")
		return "", ErrHomeDirectoryNotFound
	}

	return filepath.Clean(home), nil
}

func listDirectory(path string) ([]ExplorerNode, error) {
	target := strings.TrimSpace(path)
	if target == "" {
		home, err := getHomeDirectory()
		if err != nil {
			return nil, err
		}
		target = home
	}

	cleanPath := filepath.Clean(target)
	entries, err := os.ReadDir(cleanPath)
	if err != nil {
		filesystemLog().Errorw("failed to list directory", "path", cleanPath, "error", err)
		return nil, err
	}

	nodes := make([]ExplorerNode, 0, len(entries))

	for _, entry := range entries {
		entryPath := filepath.Join(cleanPath, entry.Name())

		info, infoErr := entry.Info()
		if infoErr != nil {
			filesystemLog().Warnw("failed to read file info, skipping entry", "path", entryPath, "error", infoErr)
			continue
		}

		nodeType := "file"
		if entry.IsDir() {
			nodeType = "directory"
		}

		modified := info.ModTime()
		if modified.IsZero() {
			modified = time.Now()
		}

		node := ExplorerNode{
			ID:       entryPath,
			Name:     entry.Name(),
			Path:     entryPath,
			Type:     nodeType,
			Modified: modified.UnixMilli(),
		}

		if nodeType == "file" {
			node.Size = info.Size()
		}

		nodes = append(nodes, node)
	}

	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Type != nodes[j].Type {
			return nodes[i].Type == "directory"
		}
		return strings.ToLower(nodes[i].Name) < strings.ToLower(nodes[j].Name)
	})

	filesystemLog().Debugw("directory listed", "path", cleanPath, "entries", len(nodes))
	return nodes, nil
}

func (a *App) GetHomeDirectory() types.APIResponse[string] {
	home, err := getHomeDirectory()
	if err != nil {
		code := string(apperrors.CodeFilesystemReadHomeFailed)
		if errors.Is(err, ErrHomeDirectoryNotFound) {
			code = string(apperrors.CodeFilesystemHomeNotFound)
		}
		return errResponse[string](code, "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", home)
}

func (a *App) ListDirectory(path string) types.APIResponse[[]ExplorerNode] {
	nodes, err := listDirectory(path)
	if err != nil {
		return errResponse[[]ExplorerNode](string(apperrors.CodeFilesystemListDirectoryFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", nodes)
}
