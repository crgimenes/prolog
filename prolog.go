// Package prolog is a thin entry point: the implementation, and the
// documentation of each symbol, live in package core.
package prolog

import (
	"io"

	"github.com/crgimenes/prolog/core"
)

type (
	Interpreter = core.Interpreter
	Scanner     = core.Scanner
	Solution    = core.Solution
	Solutions   = core.Solutions
	TermString  = core.TermString
)

var (
	ErrClosed      = core.ErrClosed
	ErrNoSolutions = core.ErrNoSolutions
)

func New(in io.Reader, out io.Writer) *Interpreter { return core.New(in, out) }
