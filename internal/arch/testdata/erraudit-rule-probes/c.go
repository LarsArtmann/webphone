package errtest

type tok struct{ email string }

func newRej(code, msg string) *efErr { return &efErr{code: code, msg: msg} }

func m(o *obj, t string) (tok, error) {
	if len(t) < 16 {
		return tok{}, newRej("c", "too short")
	}
	email, uid, exp, found, err := o.lookup(t)
	if err != nil {
		return tok{}, err
	}
	if !found {
		return tok{}, newRej("c", "unknown")
	}
	_ = uid
	_ = exp
	return tok{email: email}, nil
}

func n(o *obj, t string) (tok, error) {
	_, _, _, found, err := o.lookup(t)
	if err != nil {
		return tok{}, err
	}
	if !found {
		return tok{}, newRej("c", "unknown")
	}
	return tok{}, nil
}
