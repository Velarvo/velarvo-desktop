package app

import "errors"

var (
	ErrVaultServiceUnavailable      = errors.New("vault service is not initialized")
	ErrProjectsServiceUnavailable   = errors.New("projects service is not initialized")
	ErrWorkspacesServiceUnavailable = errors.New("workspaces service is not initialized")
	ErrSettingsServiceUnavailable   = errors.New("settings service is not initialized")
	ErrSSHServiceUnavailable        = errors.New("ssh service is not initialized")
	ErrHomeDirectoryNotFound        = errors.New("home directory not found")
)
