package vault

import "errors"

var (
	ErrAlreadySetup          = errors.New("vault is already setup")
	ErrNotSetup              = errors.New("vault is not setup")
	ErrLocked                = errors.New("vault is locked")
	ErrInvalidMasterPassword = errors.New("invalid master password")
	ErrInvalidKDFParameters  = errors.New("invalid key derivation parameters")
	ErrWeakMasterPassword    = errors.New("master password is too weak")
	ErrUnsupportedMethod     = errors.New("unlock method is not supported yet")
)
