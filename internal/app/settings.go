package app

import (
	"github.com/Velarvo/velarvo-desktop/internal/apperrors"
	"github.com/Velarvo/velarvo-desktop/internal/settings"
	"github.com/Velarvo/velarvo-desktop/internal/types"
)

func (a *App) GetSettings() types.APIResponse[settings.Settings] {
	service, err := a.requireSettings()
	if err != nil {
		return errResponse[settings.Settings](string(apperrors.CodeSettingsNotInitialized), "", err)
	}
	snapshot, err := service.GetAll(a.ctx)
	if err != nil {
		return errResponse[settings.Settings](string(apperrors.CodeSettingsReadFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", snapshot)
}

func (a *App) SetUILanguage(code string) types.APIResponse[settings.Settings] {
	service, err := a.requireSettings()
	if err != nil {
		return errResponse[settings.Settings](string(apperrors.CodeSettingsNotInitialized), "", err)
	}
	if err := service.SetLanguage(a.ctx, settings.LanguageCode(code)); err != nil {
		return errResponse[settings.Settings](string(apperrors.CodeSettingsWriteFailed), "", err)
	}
	snapshot, err := service.GetAll(a.ctx)
	if err != nil {
		return errResponse[settings.Settings](string(apperrors.CodeSettingsReadFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", snapshot)
}

func (a *App) requireSettings() (*settings.Service, error) {
	if a.settings == nil {
		return nil, ErrSettingsServiceUnavailable
	}
	return a.settings, nil
}
