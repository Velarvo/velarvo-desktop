package app

import (
	"github.com/Velarvo/velarvo-desktop/internal/apperrors"
	"github.com/Velarvo/velarvo-desktop/internal/types"
)

func errResponse[T any](code string, message string, err error) types.APIResponse[T] {
	if code == "" {
		code = string(apperrors.CodeOf(err))
	}

	return types.APIResponse[T]{
		Success:    false,
		Code:       code,
		Message:    message,
		Params:     apperrors.ParamsOf(err),
		Error:      apperrors.DebugMessage(err),
		StatusCode: 400,
	}
}

func successResponse[T any](code string, message string, data T) types.APIResponse[T] {
	return types.APIResponse[T]{
		Success:    true,
		Code:       code,
		Message:    message,
		StatusCode: 200,
		Data:       &data,
	}
}
