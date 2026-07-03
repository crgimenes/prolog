package main

import (
	"fmt"
	"github.com/crgimenes/prolog/engine"

	"github.com/crgimenes/prolog"
)

func main() {
	// Instantiates a new Prolog interpreter without any builtin predicates nor operators.
	p := new(prolog.Interpreter)

	// In this vanilla interpreter, even the infix operator `:-` is not defined.
	// Instead of writing `:-(mortal(X), human(X)).`, you may want to define the infix operator first.

	// To define operators, register op/3.
	p.Register3(engine.NewAtom("op"), engine.Op)

	// Then, define the infix operator with priority 1200 and specifier XFX.
	err2 := p.Exec(`:-(op(1200, xfx, :-)).`)
	if err2 != nil {
		panic(err2)
	}

	// You may also want to register other predicates or define other operators to match your use case.
	// You can use p.Register0~8 to register any builtin/custom predicates of respective arity.

	// Now you can load a Prolog program with infix `:-`.
	err3 := p.Exec(`
		human(socrates).
		mortal(X) :- human(X).
	`)
	if err3 != nil {
		panic(err3)
	}

	// Run the Prolog program.
	sols, err := p.Query(`mortal(Who).`)
	if err != nil {
		panic(err)
	}
	defer func() { _ = sols.Close() }()

	for sols.Next() {
		var s struct {
			Who string
		}
		err4 := sols.Scan(&s)
		if err4 != nil {
			panic(err4)
		}
		fmt.Printf("Who = %s\n", s.Who)
		// ==> Who = socrates
	}
}
