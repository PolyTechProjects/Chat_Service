package errors

import (
	"errors"
	"fmt"
)

var (
	ErrMapping = errors.New("mapping error")

	ErrReadMessageError = errors.New("websocket read message error")

	ErrDataIntegrityViolation = fmt.Errorf("data integrity violation")

	ErrDatabaseInternalError = fmt.Errorf("database internal error")

	ErrPublishMessageError = fmt.Errorf("publish message error")

	ErrSetStatusRedis = fmt.Errorf("set user status redis error")

	ErrDropStatusRedis = fmt.Errorf("drop user status redis error")
)

func IsDatabaseInternalError(err error) bool {
	return errors.Is(err, ErrDatabaseInternalError)
}
