package errtest

import "errors"

type efErr struct{ code, msg string }

func (e *efErr) Error() string { return e.code + ": " + e.msg }

func NewRejection(code, msg string) *efErr             { return &efErr{code: code, msg: msg} }
func WrapRejection(err error, code, msg string) *efErr { return &efErr{code: code, msg: err.Error()} }

func h() error { return NewRejection("x", "y") }

func i() error { return WrapRejection(errors.New("z"), "x", "y") }

func j(v bool) error {
	if v {
		return nil
	}
	return NewRejection("x", "y")
}

type obj struct{}

func (o *obj) lookup(h string) (string, string, int, bool, error) { return "", "", 0, false, nil }

func k(o *obj) (string, error) {
	s, _, _, found, err := o.lookup("x")
	if err != nil {
		return "", err
	}
	if !found {
		return "", &joinish{errs: []error{errors.New("a"), NewRejection("b", "c")}}
	}
	return s, nil
}

type joinish struct{ errs []error }

func (e *joinish) Error() string { return "joinish" }
