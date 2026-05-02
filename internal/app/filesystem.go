package app

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type ExplorerNode struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"`
	Size     int64  `json:"size,omitempty"`
	Modified int64  `json:"modified"`
}

func (a *App) GetHomeDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if home == "" {
		return "", errors.New("home directory not found")
	}

	return filepath.Clean(home), nil
}

func (a *App) ListDirectory(path string) ([]ExplorerNode, error) {
	target := strings.TrimSpace(path)
	if target == "" {
		home, err := a.GetHomeDirectory()
		if err != nil {
			return nil, err
		}
		target = home
	}

	cleanPath := filepath.Clean(target)
	entries, err := os.ReadDir(cleanPath)
	if err != nil {
		return nil, err
	}

	nodes := make([]ExplorerNode, 0, len(entries))

	for _, entry := range entries {
		entryPath := filepath.Join(cleanPath, entry.Name())

		info, infoErr := entry.Info()
		if infoErr != nil {
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

	return nodes, nil
}
