package workspaces

import "errors"

var (
	ErrLocked         = errors.New("vault is locked")
	ErrNotFound       = errors.New("workspace not found")
	ErrProjectMissing = errors.New("workspace project not found")
	ErrDuplicateName  = errors.New("workspace name already exists in project")
	ErrInvalidInput   = errors.New("invalid workspace input")
	ErrCipherNotReady = errors.New("workspace cipher is not configured")
)
