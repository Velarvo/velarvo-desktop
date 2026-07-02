package apperrors

import (
	"errors"
	"fmt"
	"os"

	"github.com/Velarvo/velarvo-desktop/internal/projects"
	"github.com/Velarvo/velarvo-desktop/internal/settings"
	"github.com/Velarvo/velarvo-desktop/internal/sshconn"
	"github.com/Velarvo/velarvo-desktop/internal/vault"
	"github.com/Velarvo/velarvo-desktop/internal/workspaces"
)

type Code string

const (
	CodeOK                       Code = "OK"
	CodeAuthOK                   Code = "AUTH_OK"
	CodeAuthLogoutOK             Code = "AUTH_LOGOUT_OK"
	CodeAuthRefreshOK            Code = "AUTH_REFRESH_OK"
	CodeNotAuthenticated         Code = "NOT_AUTHENTICATED"
	CodeSessionExpired           Code = "SESSION_EXPIRED"
	CodeDeviceError              Code = "DEVICE_ERROR"
	CodeRemoteCallError          Code = "REMOTE_CALL_ERROR"
	CodeKeychainError            Code = "KEYCHAIN_ERROR"
	CodeInvalidResponse          Code = "INVALID_RESPONSE"
	CodeClientError              Code = "CLIENT_ERROR"
	CodeInternalError            Code = "INTERNAL_ERROR"
	CodeServiceUnavailable       Code = "SERVICE_UNAVAILABLE"
	CodeVaultNotInitialized      Code = "VAULT_NOT_INITIALIZED"
	CodeProjectsNotInitialized   Code = "PROJECTS_NOT_INITIALIZED"
	CodeWorkspacesNotInitialized Code = "WORKSPACES_NOT_INITIALIZED"
	CodeSettingsNotInitialized   Code = "SETTINGS_NOT_INITIALIZED"

	CodeVaultNotSetup         Code = "VAULT_NOT_SETUP"
	CodeVaultAlreadySetup     Code = "VAULT_ALREADY_SETUP"
	CodeVaultLocked           Code = "VAULT_LOCKED"
	CodeInvalidMasterPassword Code = "INVALID_MASTER_PASSWORD"
	CodeWeakMasterPassword    Code = "WEAK_MASTER_PASSWORD"
	CodeVaultUnsupportedState Code = "VAULT_UNSUPPORTED_STATE"
	CodeVaultReadStateFailed  Code = "VAULT_READ_STATE_FAILED"
	CodeVaultSetupFailed      Code = "VAULT_SETUP_FAILED"
	CodeVaultUnlockFailed     Code = "VAULT_UNLOCK_FAILED"
	CodeVaultLockFailed       Code = "VAULT_LOCK_FAILED"

	CodeProjectNotFound         Code = "PROJECT_NOT_FOUND"
	CodeProjectDuplicateName    Code = "PROJECT_DUPLICATE_NAME"
	CodeProjectInvalidInput     Code = "PROJECT_INVALID_INPUT"
	CodeProjectIconNotFound     Code = "PROJECT_ICON_NOT_FOUND"
	CodeProjectCipherNotReady   Code = "PROJECT_CIPHER_NOT_READY"
	CodeProjectListFailed       Code = "PROJECT_LIST_FAILED"
	CodeProjectReadFailed       Code = "PROJECT_READ_FAILED"
	CodeProjectCreateFailed     Code = "PROJECT_CREATE_FAILED"
	CodeProjectUpdateFailed     Code = "PROJECT_UPDATE_FAILED"
	CodeProjectDeleteFailed     Code = "PROJECT_DELETE_FAILED"
	CodeProjectIconSetFailed    Code = "PROJECT_ICON_SET_FAILED"
	CodeProjectIconReadFailed   Code = "PROJECT_ICON_READ_FAILED"
	CodeProjectIconDeleteFailed Code = "PROJECT_ICON_DELETE_FAILED"

	CodeWorkspaceNotFound       Code = "WORKSPACE_NOT_FOUND"
	CodeWorkspaceProjectMissing Code = "WORKSPACE_PROJECT_MISSING"
	CodeWorkspaceDuplicateName  Code = "WORKSPACE_DUPLICATE_NAME"
	CodeWorkspaceInvalidInput   Code = "WORKSPACE_INVALID_INPUT"
	CodeWorkspaceCipherNotReady Code = "WORKSPACE_CIPHER_NOT_READY"
	CodeWorkspaceListFailed     Code = "WORKSPACE_LIST_FAILED"
	CodeWorkspaceReadFailed     Code = "WORKSPACE_READ_FAILED"
	CodeWorkspaceCreateFailed   Code = "WORKSPACE_CREATE_FAILED"
	CodeWorkspaceUpdateFailed   Code = "WORKSPACE_UPDATE_FAILED"
	CodeWorkspaceDeleteFailed   Code = "WORKSPACE_DELETE_FAILED"

	CodeSSHNotInitialized   Code = "SSH_NOT_INITIALIZED"
	CodeSSHNotFound         Code = "SSH_NOT_FOUND"
	CodeSSHInvalidInput     Code = "SSH_INVALID_INPUT"
	CodeSSHCipherNotReady   Code = "SSH_CIPHER_NOT_READY"
	CodeSSHNoPassword       Code = "SSH_NO_PASSWORD"
	CodeSSHListFailed       Code = "SSH_LIST_FAILED"
	CodeSSHReadFailed       Code = "SSH_READ_FAILED"
	CodeSSHCreateFailed     Code = "SSH_CREATE_FAILED"
	CodeSSHUpdateFailed     Code = "SSH_UPDATE_FAILED"
	CodeSSHDeleteFailed     Code = "SSH_DELETE_FAILED"
	CodeSSHConnectFailed    Code = "SSH_CONNECT_FAILED"
	CodeSSHDisconnectFailed Code = "SSH_DISCONNECT_FAILED"
	CodeSSHNotConnected     Code = "SSH_NOT_CONNECTED"
	CodeSSHTerminalFailed   Code = "SSH_TERMINAL_FAILED"

	CodeSettingsReadFailed       Code = "SETTINGS_READ_FAILED"
	CodeSettingsWriteFailed      Code = "SETTINGS_WRITE_FAILED"
	CodeSettingsUnsupportedValue Code = "SETTINGS_UNSUPPORTED_VALUE"

	CodeFilesystemHomeNotFound        Code = "FILESYSTEM_HOME_NOT_FOUND"
	CodeFilesystemPathNotFound        Code = "FILESYSTEM_PATH_NOT_FOUND"
	CodeFilesystemAccessDenied        Code = "FILESYSTEM_ACCESS_DENIED"
	CodeFilesystemUnavailable         Code = "FILESYSTEM_UNAVAILABLE"
	CodeFilesystemReadHomeFailed      Code = "FILESYSTEM_READ_HOME_FAILED"
	CodeFilesystemListDirectoryFailed Code = "FILESYSTEM_LIST_DIRECTORY_FAILED"
)

type AppError struct {
	Code   Code
	Err    error
	Params map[string]string
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Code)
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func Wrap(code Code, err error, params map[string]string) error {
	return &AppError{
		Code:   code,
		Err:    err,
		Params: params,
	}
}

func CodeOf(err error) Code {
	var appErr *AppError
	if errors.As(err, &appErr) && appErr.Code != "" {
		return appErr.Code
	}

	switch {
	case errors.Is(err, vault.ErrNotSetup):
		return CodeVaultNotSetup
	case errors.Is(err, vault.ErrAlreadySetup):
		return CodeVaultAlreadySetup
	case errors.Is(err, vault.ErrLocked):
		return CodeVaultLocked
	case errors.Is(err, vault.ErrInvalidMasterPassword):
		return CodeInvalidMasterPassword
	case errors.Is(err, vault.ErrWeakMasterPassword):
		return CodeWeakMasterPassword
	case errors.Is(err, projects.ErrNotFound):
		return CodeProjectNotFound
	case errors.Is(err, projects.ErrDuplicateName):
		return CodeProjectDuplicateName
	case errors.Is(err, projects.ErrInvalidInput):
		return CodeProjectInvalidInput
	case errors.Is(err, projects.ErrIconNotFound):
		return CodeProjectIconNotFound
	case errors.Is(err, projects.ErrCipherNotReady):
		return CodeProjectCipherNotReady
	case errors.Is(err, workspaces.ErrNotFound):
		return CodeWorkspaceNotFound
	case errors.Is(err, workspaces.ErrProjectMissing):
		return CodeWorkspaceProjectMissing
	case errors.Is(err, workspaces.ErrDuplicateName):
		return CodeWorkspaceDuplicateName
	case errors.Is(err, workspaces.ErrInvalidInput):
		return CodeWorkspaceInvalidInput
	case errors.Is(err, workspaces.ErrCipherNotReady):
		return CodeWorkspaceCipherNotReady
	case errors.Is(err, sshconn.ErrNotFound):
		return CodeSSHNotFound
	case errors.Is(err, sshconn.ErrInvalidInput):
		return CodeSSHInvalidInput
	case errors.Is(err, sshconn.ErrCipherNotReady):
		return CodeSSHCipherNotReady
	case errors.Is(err, sshconn.ErrNoPassword):
		return CodeSSHNoPassword
	case errors.Is(err, sshconn.ErrConnectFailed):
		return CodeSSHConnectFailed
	case errors.Is(err, sshconn.ErrNotConnected):
		return CodeSSHNotConnected
	case errors.Is(err, sshconn.ErrSessionNotFound):
		return CodeSSHTerminalFailed
	case errors.Is(err, sshconn.ErrTerminalFailed):
		return CodeSSHTerminalFailed
	case errors.Is(err, sshconn.ErrLocked):
		return CodeVaultLocked
	case errors.Is(err, settings.ErrUnsupportedLanguage):
		return CodeSettingsUnsupportedValue
	case errors.Is(err, os.ErrNotExist):
		return CodeFilesystemPathNotFound
	case errors.Is(err, os.ErrPermission):
		return CodeFilesystemAccessDenied
	case err != nil:
		return CodeInternalError
	default:
		return CodeOK
	}
}

func ParamsOf(err error) map[string]string {
	var appErr *AppError
	if errors.As(err, &appErr) && len(appErr.Params) > 0 {
		return appErr.Params
	}
	return nil
}

func DebugMessage(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%v", err)
}
