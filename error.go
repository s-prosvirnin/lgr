package main

import (
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/pkg/errors"
)

type ErrType string

func (t ErrType) String() string {
	return string(t)
}
func (t ErrType) IsEmpty() bool {
	return len(t) == 0
}
func (t ErrType) IsErrorOfType(err error) bool {
	if err == nil && t.IsEmpty() {
		return true
	}
	for err != nil {
		var tErr Error
		ok := errors.As(err, &tErr)
		if ok && tErr.Type() == t {
			return true
		}
		err = errors.Unwrap(err)
	}

	return false
}

type Error interface {
	WithParams(params ...KVParam) Error
	WithCtx(ctx context.Context) Error
	WithType(t ErrType) Error
	WithMsgWrap(message string) Error
	Type() ErrType
	Error() string
	Params() []KVParam
	IsExpected(inspectedErr error) bool
	IsOfType(errType ErrType) bool
}

func ErrExtract(err error) Error {
	var tErr Error
	ok := errors.As(err, &tErr)
	if ok {
		return tErr
	}

	return nil
}

func NewErr(err error) Error {
	var tErr *errorDto
	ok := errors.As(err, &tErr)
	if ok {
		return &errorDto{err: tErr.err, errType: tErr.Type(), params: tErr.Params()}
	}
	return &errorDto{err: err}
}

func NewErrFromMsg(message string) Error {
	return &errorDto{err: errors.New(message)}
}

type errorDto struct {
	err     error
	errType ErrType
	params  []KVParam
}

func (e errorDto) WithCtx(ctx context.Context) Error {
	return e.WithParams(getContextParams(ctx)...)
}

func (e errorDto) WithParams(params ...KVParam) Error {
	for _, param := range e.params {
		params = append(params, param)
	}

	return &errorDto{err: e.err, errType: e.errType, params: params}
}

func (e errorDto) WithType(t ErrType) Error {
	return &errorDto{err: e.err, params: e.Params(), errType: t}
}

func (e errorDto) WithMsgWrap(message string) Error {
	return NewErr(errors.Wrap(e.err, message)).WithParams(e.Params()...).WithType(e.Type())
}

func (e errorDto) Type() ErrType {
	return e.errType
}

func (e errorDto) Error() string {
	if e.err == nil {
		return "ERROR!!! native error is nil"
	}
	return e.err.Error()
}

func (e errorDto) Unwrap() error {
	return errors.Unwrap(e.err)
}

func (e errorDto) IsExpected(inspectedErr error) bool {
	var tInspectedErr errorDto
	ok := errors.As(inspectedErr, &tInspectedErr)
	if ok && !e.Type().IsEmpty() {
		if e.Type() != tInspectedErr.Type() || e.Type().IsErrorOfType(tInspectedErr.err) {
			return false
		}
		return true
	}
	return errors.Is(e.err, inspectedErr)
}

func (e errorDto) IsOfType(errType ErrType) bool {
	return e.Type() == errType
}

func (e errorDto) Params() []KVParam {
	return slices.Clone(e.params)
}

func (e errorDto) Format(s fmt.State, verb rune) {
	io.WriteString(s, e.Error())
}
