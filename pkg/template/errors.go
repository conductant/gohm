package template

import (
	"errors"
)

var (
	ErrMissingTemplateFunc = errors.New("no-template-func")
	ErrBadTemplateFunc     = errors.New("err-bad-template-func")
)

func ErrNotSupported(protocol string) error {
	return errors.New("not-supported-" + protocol)
}
