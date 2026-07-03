package prolog

import (
	"testing"
)

// FuzzExecDoesNotPanic feeds arbitrary text to the full pipeline (lexer,
// parser, compiler): whatever the input, the interpreter must return an error
// instead of panicking.
func FuzzExecDoesNotPanic(f *testing.F) {
	f.Add("human(socrates).")
	f.Add("mortal(X) :- human(X).")
	f.Add(":- op(1200, xfx, :-).")
	f.Add(`foo("double quoted").`)
	f.Add("bad(")
	f.Add("0'a. 0'\\n. 1.5e10.")
	f.Add("/* comment */ % line\nfact.")
	f.Add("X = [a,b|T].")

	f.Fuzz(func(t *testing.T, src string) {
		p := New(nil, nil)
		_ = p.Exec(src)
	})
}
