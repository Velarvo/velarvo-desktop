package app

import "github.com/Velarvo/velarvo-desktop/internal/types"

func errResponse(code, message string, err error) types.APIResponse[map[string]string] {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	return types.APIResponse[map[string]string]{
		Success:    false,
		Code:       code,
		Message:    message,
		Error:      errStr,
		StatusCode: 400,
	}
}

func successResponse(code, message string, data map[string]string) types.APIResponse[map[string]string] {
	return types.APIResponse[map[string]string]{
		Success:    true,
		Code:       code,
		Message:    message,
		StatusCode: 200,
		Data:       &data,
	}
}
