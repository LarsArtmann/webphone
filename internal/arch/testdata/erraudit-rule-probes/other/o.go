package other

type E struct{ code, msg string }

func (e *E) Error() string { return e.code + ": " + e.msg }

func Wrapf(err error, fam int, code, format string, args ...any) *E {
	return &E{code: code, msg: err.Error()}
}

func NewRejection(code, msg string) *E { return &E{code: code, msg: msg} }
