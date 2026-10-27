package errtest

import (
	"errors"
	"fmt"
)

func NewThing(flag bool) error {
	if flag {
		return fmt.Errorf("newthing: %w", errors.New("x"))
	}
	return nil
}

type svc struct{ store *obj }

func (s *svc) look(t string) (string, string, int, bool, error) { return s.store.lookup(t) }

func (s *svc) verify(t string) (tok, error) {
	email, _, _, found, err := s.store.lookup(t)
	if err != nil {
		return tok{}, err
	}
	if !found {
		return tok{}, newRej("c", "unknown")
	}
	if len(t) > 3 {
		return tok{}, wrapf(err2(t), 1, "code", "bad %q", t)
	}
	return tok{email: email}, nil
}

func err2(t string) error { return errors.New("e2") }

func wrapf(err error, fam int, code, format string, args ...any) *efErr {
	return &efErr{code: code, msg: err.Error()}
}
