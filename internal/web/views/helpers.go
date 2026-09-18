package views

import "strconv"

// fmtInt renders counts for templ (templ children cannot call strconv
// directly with a plain int conversion).
func fmtInt(n int) string { return strconv.Itoa(n) }
