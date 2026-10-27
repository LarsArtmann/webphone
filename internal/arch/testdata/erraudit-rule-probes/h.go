package errtest

import "errtest/other"

type EnrollToken struct{ Email string }

func (s *svc) verify2(t string) (EnrollToken, error) {
	if len(t) < 16 {
		return EnrollToken{}, other.NewRejection("c", "short")
	}
	email, raw, found, err := s.store.lookup2(t)
	if err != nil {
		return EnrollToken{}, err
	}
	if !found {
		return EnrollToken{}, other.NewRejection("c", "unknown")
	}
	if !valid(raw) {
		return EnrollToken{}, other.Wrapf(err2(t), 1, "code", "invalid %q", raw)
	}
	return EnrollToken{Email: email}, nil
}

func (o *obj) lookup2(t string) (string, string, bool, error) { return t, t, false, nil }

func valid(s string) bool { return s != "" }
