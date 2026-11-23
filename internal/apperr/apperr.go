package apperr

import (
	"encoding/json"
	"errors"
	"io/fs"
)

type Error struct {
	Code    string
	Message string
}

func New(code, message string) *Error {
	return &Error{Code: code, Message: message}
}

func (e *Error) Error() string { return e.Message }

const (
	CodeNotFound     = "PATH_NOT_FOUND"
	CodeAccessDenied = "ACCESS_DENIED"
)

func Marshal(err error) []byte {
	var code string
	var appErr *Error
	switch {
	case errors.As(err, &appErr):
		code = appErr.Code
	case errors.Is(err, fs.ErrNotExist):
		code = CodeNotFound
	case errors.Is(err, fs.ErrPermission):
		code = CodeAccessDenied
	default:
		return nil
	}

	data, jsonErr := json.Marshal(map[string]string{"code": code})
	if jsonErr != nil {
		return nil
	}
	return data
}
