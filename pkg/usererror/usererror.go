// Package usererror provides an error type that wraps both the original error and a
// more "user-friendly" message that details a common user-error and the fix
package usererror

import "errors"

type userError struct {
	original error
	msg      string
}

// New creates a new userError from the original error and a user-friendly message that
// details the common user-error and the fix
func New(original error, msg string) *userError {
	return &userError{
		original: original,
		msg:      msg,
	}
}

func (u *userError) Error() string {
	return u.original.Error()
}

func (u *userError) Msg() string {
	return u.msg
}

func Is(err error) (*userError, bool) {
	var u *userError
	if ok := errors.As(err, &u); !ok {
		return nil, false
	}

	return u, true
}
