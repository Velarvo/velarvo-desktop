package projects

import "errors"

var (
	ErrLocked         = errors.New("vault is locked")
	ErrNotFound       = errors.New("project not found")
	ErrDuplicateName  = errors.New("project name already exists")
	ErrInvalidInput   = errors.New("invalid project input")
	ErrIconNotFound   = errors.New("project icon not found")
	ErrCipherNotReady = errors.New("project cipher is not configured")
)
