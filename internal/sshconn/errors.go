package sshconn

import "errors"

var (
	ErrLocked            = errors.New("vault is locked")
	ErrNotFound          = errors.New("ssh connection not found")
	ErrWorkspaceRequired = errors.New("ssh connection requires a workspace")
	ErrWorkspaceNotFound = errors.New("ssh connection workspace not found")
	ErrInvalidInput      = errors.New("invalid ssh connection input")
	ErrCipherNotReady    = errors.New("ssh connection cipher is not configured")
	ErrNoPassword        = errors.New("ssh connection has no stored password")
	ErrConnectFailed     = errors.New("ssh connection failed")
	ErrNotConnected      = errors.New("ssh connection is not connected")
	ErrTerminalFailed    = errors.New("ssh terminal failed")
	ErrSessionNotFound   = errors.New("ssh terminal session not found")

	errProbeTimeout = errors.New("ssh os probe timed out")
)
