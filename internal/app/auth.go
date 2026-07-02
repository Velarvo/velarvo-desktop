package app

import (
	"errors"

	"github.com/Velarvo/velarvo-desktop/internal/apperrors"
	"github.com/Velarvo/velarvo-desktop/internal/device"
	"github.com/Velarvo/velarvo-desktop/internal/logger"
	"github.com/Velarvo/velarvo-desktop/internal/session"
	"github.com/Velarvo/velarvo-desktop/internal/types"
	"go.uber.org/zap"
)

func authLog() *zap.SugaredLogger {
	return logger.Named("auth")
}

func (a *App) GetAuthState() types.APIResponse[map[string]string] {
	if !a.session.IsLoggedIn() {
		return errResponse[map[string]string](string(apperrors.CodeNotAuthenticated), "", nil)
	}

	current := a.session.Get()
	if current == nil {
		return errResponse[map[string]string](string(apperrors.CodeNotAuthenticated), "", nil)
	}

	if current.UserID == "" {
		resp := a.fetchCurrentUser()
		if !resp.Success {
			refreshResp := a.Refresh()
			if !refreshResp.Success {
				return errResponse[map[string]string](string(apperrors.CodeSessionExpired), "", nil)
			}

			resp = a.fetchCurrentUser()
			if !resp.Success {
				return resp
			}
		}

		return resp
	}

	return successResponse(string(apperrors.CodeAuthOK), "", map[string]string{
		"userId":    current.UserID,
		"email":     current.Email,
		"firstName": current.FirstName,
		"lastName":  current.LastName,
	})
}

func (a *App) Register(email, firstName, lastName, password string) types.APIResponse[map[string]string] {
	resp, err := a.client.Register(&types.RegisterPayload{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Password:  password,
	})
	if err != nil {
		authLog().Errorw("register failed", "error", err)
		return errResponse[map[string]string](string(apperrors.CodeRemoteCallError), "", err)
	}
	if !resp.Success {
		return errResponse[map[string]string](resp.Code, resp.Message, errors.New(resp.Message))
	}
	if resp.Data == nil {
		return errResponse[map[string]string](string(apperrors.CodeInvalidResponse), "", nil)
	}

	return successResponse(resp.Code, resp.Message, map[string]string{
		"id":    resp.Data.ID,
		"email": resp.Data.Email,
	})
}

func (a *App) Login(email, password string) types.APIResponse[map[string]string] {
	deviceID, err := device.GetOrCreateID()
	if err != nil {
		authLog().Errorw("failed to get device ID", "error", err)
		return errResponse[map[string]string](string(apperrors.CodeDeviceError), "", err)
	}

	resp, err := a.client.Login(&types.LoginPayload{
		Email:          email,
		Password:       password,
		DeviceClientId: deviceID,
		DeviceOs:       device.GetOS(),
		DeviceName:     device.GetName(),
	})
	if err != nil {
		authLog().Errorw("login failed", "error", err)
		return errResponse[map[string]string](string(apperrors.CodeRemoteCallError), "", err)
	}
	if !resp.Success {
		return errResponse[map[string]string](resp.Code, resp.Message, errors.New(resp.Message))
	}
	if resp.Data == nil {
		return errResponse[map[string]string](string(apperrors.CodeInvalidResponse), "", nil)
	}

	if err := a.keychain.SaveTokens(resp.Data.AccessToken, resp.Data.RefreshToken); err != nil {
		authLog().Errorw("failed to save tokens", "error", err)
		return errResponse[map[string]string](string(apperrors.CodeKeychainError), "", err)
	}

	a.session.Set(&session.UserSession{
		AccessToken:  resp.Data.AccessToken,
		RefreshToken: resp.Data.RefreshToken,
		UserID:       resp.Data.User.ID,
		Email:        resp.Data.User.Email,
		FirstName:    strOrEmpty(resp.Data.User.FirstName),
		LastName:     strOrEmpty(resp.Data.User.LastName),
		DeviceID:     resp.Data.Device.ID,
	})

	return successResponse(resp.Code, resp.Message, map[string]string{
		"userId":    resp.Data.User.ID,
		"email":     resp.Data.User.Email,
		"firstName": strOrEmpty(resp.Data.User.FirstName),
		"lastName":  strOrEmpty(resp.Data.User.LastName),
		"deviceId":  resp.Data.Device.ID,
	})
}

func (a *App) Logout() types.APIResponse[map[string]string] {
	refreshToken, err := a.keychain.GetRefreshToken()
	if err == nil && refreshToken != "" {
		if _, err := a.client.Logout(&types.RefreshPayload{RefreshToken: refreshToken}); err != nil {
			authLog().Warnw("logout request failed", "error", err)
		}
	}

	if err := a.keychain.ClearTokens(); err != nil {
		authLog().Warnw("failed to clear tokens from keychain", "error", err)
	}
	a.session.Clear()

	return successResponse(string(apperrors.CodeAuthLogoutOK), "", map[string]string{})
}

func (a *App) Refresh() types.APIResponse[map[string]string] {
	refreshToken, err := a.keychain.GetRefreshToken()
	if err != nil {
		return errResponse[map[string]string](string(apperrors.CodeKeychainError), "", err)
	}

	resp, err := a.client.Refresh(&types.RefreshPayload{RefreshToken: refreshToken})
	if err != nil {
		authLog().Errorw("refresh failed", "error", err)
		return errResponse[map[string]string](string(apperrors.CodeRemoteCallError), "", err)
	}
	if !resp.Success {
		if err := a.keychain.ClearTokens(); err != nil {
			authLog().Warnw("failed to clear tokens from keychain", "error", err)
		}
		a.session.Clear()
		return errResponse[map[string]string](resp.Code, resp.Message, errors.New(resp.Message))
	}
	if resp.Data == nil {
		return errResponse[map[string]string](string(apperrors.CodeInvalidResponse), "", nil)
	}

	if err := a.keychain.SaveTokens(resp.Data.AccessToken, resp.Data.RefreshToken); err != nil {
		authLog().Errorw("failed to persist refreshed tokens", "error", err)
		return errResponse[map[string]string](string(apperrors.CodeKeychainError), "", err)
	}

	a.session.UpdateTokens(resp.Data.AccessToken, resp.Data.RefreshToken)
	return successResponse(string(apperrors.CodeAuthRefreshOK), "", map[string]string{})
}

func (a *App) fetchCurrentUser() types.APIResponse[map[string]string] {
	resp, err := a.client.GetCurrentUser()
	if err != nil {
		authLog().Errorw("failed to fetch current user", "error", err)
		return errResponse[map[string]string](string(apperrors.CodeRemoteCallError), "", err)
	}
	if !resp.Success {
		return errResponse[map[string]string](resp.Code, resp.Message, errors.New(resp.Message))
	}
	if resp.Data == nil {
		return errResponse[map[string]string](string(apperrors.CodeInvalidResponse), "", nil)
	}

	current := a.session.Get()
	if current != nil {
		current.UserID = resp.Data.ID
		current.Email = resp.Data.Email
		current.FirstName = strOrEmpty(resp.Data.FirstName)
		current.LastName = strOrEmpty(resp.Data.LastName)
	}

	return successResponse(string(apperrors.CodeAuthOK), "", map[string]string{
		"userId":    resp.Data.ID,
		"email":     resp.Data.Email,
		"firstName": strOrEmpty(resp.Data.FirstName),
		"lastName":  strOrEmpty(resp.Data.LastName),
	})
}

func strOrEmpty(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
