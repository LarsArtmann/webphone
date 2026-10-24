package push

import "testing"

func TestStatusOKAcceptsEvery2xxOnly(t *testing.T) {
	cases := map[int]bool{
		199: false,
		200: true,
		204: true,
		299: true,
		300: false,
		500: false,
	}
	for code, want := range cases {
		if got := statusOK(code); got != want {
			t.Fatalf("statusOK(%d) = %v, want %v", code, got, want)
		}
	}
}
