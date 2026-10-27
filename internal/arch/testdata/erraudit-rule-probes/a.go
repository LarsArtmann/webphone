package errtest

import (
	"errors"
	"fmt"
)

func a() error { return nil }

func b(flag bool) error {
	if flag {
		return errors.New("b")
	}
	return nil
}

func c(flag bool) error {
	if flag {
		return fmt.Errorf("c: %w", errors.New("inner"))
	}
	return nil
}

func g() error { return errors.New("g") }

func d() error {
	if err := g(); err != nil {
		return err
	}
	return nil
}

type I interface{ Do() error }

func (s *S) Do() error { return g() }

type S struct{}
