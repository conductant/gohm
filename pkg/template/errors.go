package template

import (
	"errors"
)

func ErrNotSupported(protocol string) error {
	return errors.New("not-supported-" + protocol)
}
