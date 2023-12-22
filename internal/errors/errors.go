package errors

import (
	stdErrors "errors"
	"fmt"
)

type ErrorCode string

const (
	LoginAlredyExists              ErrorCode = "login has already been taken by another user"
	ErrUnmarshallingJSON           ErrorCode = "bad request"
	WrongLoginOrPassword           ErrorCode = "wrong login/password"
	OrderAlreadyAddedByThisUser    ErrorCode = "you`ve alredy loaded this order"
	OrderAlreadyAddedByAnotherUser ErrorCode = "order has already been loaded by another user"
	NotUthenticated                ErrorCode = "not authenticated"
	InvalidOrderNumber             ErrorCode = "invalid order number"
	NoDataFound                    ErrorCode = "no data found"
	InsufficientFunds              ErrorCode = "insufficient funds"
	NotUniqueToken                 ErrorCode = "session token already exists"
)

type domainError struct {
	// We define our domainError struct, which is composed of error
	error
	errorCode ErrorCode
}

func (e domainError) Error() string {
	return fmt.Sprintf("%s: %s", e.errorCode, e.error.Error())
}

func Unwrap(err error) error {
	var dErr domainError
	if stdErrors.As(err, &dErr) {
		return stdErrors.Unwrap(dErr.error)
	}

	return stdErrors.Unwrap(err)
}

func Code(err error) ErrorCode {
	if err == nil {
		return ""
	}

	var dErr domainError
	if stdErrors.As(err, &dErr) {
		return dErr.errorCode
	}

	return ""
}

func NewDomainError(errorCode ErrorCode, format string, args ...interface{}) error {
	return domainError{
		error:     fmt.Errorf(format, args...),
		errorCode: errorCode,
	}
}

func WrapIntoDomainError(err error, errorCode ErrorCode, msg string) error {
	return domainError{
		error:     fmt.Errorf("%s: [%w]", msg, err),
		errorCode: errorCode,
	}
}
