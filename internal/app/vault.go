package app

import (
	"github.com/Velarvo/velarvo-desktop/internal/apperrors"
	"github.com/Velarvo/velarvo-desktop/internal/types"
	"github.com/Velarvo/velarvo-desktop/internal/vault"
)

func (a *App) GetVaultState() types.APIResponse[vault.VaultState] {
	service, err := a.requireVault()
	if err != nil {
		return errResponse[vault.VaultState](string(apperrors.CodeVaultNotInitialized), "", err)
	}
	state, err := service.State(a.ctx)
	if err != nil {
		return errResponse[vault.VaultState](string(apperrors.CodeVaultReadStateFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", state)
}

func (a *App) SetupVault(req vault.SetupRequest) types.APIResponse[vault.VaultState] {
	service, err := a.requireVault()
	if err != nil {
		return errResponse[vault.VaultState](string(apperrors.CodeVaultNotInitialized), "", err)
	}
	state, err := service.Setup(a.ctx, req)
	if err != nil {
		return errResponse[vault.VaultState](string(apperrors.CodeVaultSetupFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", state)
}

func (a *App) UnlockVault(req vault.UnlockRequest) types.APIResponse[vault.VaultState] {
	service, err := a.requireVault()
	if err != nil {
		return errResponse[vault.VaultState](string(apperrors.CodeVaultNotInitialized), "", err)
	}
	state, err := service.Unlock(a.ctx, req)
	if err != nil {
		return errResponse[vault.VaultState](string(apperrors.CodeVaultUnlockFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", state)
}

func (a *App) LockVault() types.APIResponse[vault.VaultState] {
	service, err := a.requireVault()
	if err != nil {
		return errResponse[vault.VaultState](string(apperrors.CodeVaultNotInitialized), "", err)
	}

	service.Lock()
	state, err := service.State(a.ctx)
	if err != nil {
		return errResponse[vault.VaultState](string(apperrors.CodeVaultLockFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", state)
}

func (a *App) requireVault() (*vault.Service, error) {
	if a.vault == nil {
		return nil, ErrVaultServiceUnavailable
	}
	return a.vault, nil
}
