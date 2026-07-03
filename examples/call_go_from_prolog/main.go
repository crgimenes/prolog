package main

import (
	"fmt"
	"net/http"

	"github.com/crgimenes/prolog"
	"github.com/crgimenes/prolog/engine"
)

func main() {
	p := prolog.New(nil, nil)

	// Define a custom predicate of arity 2.
	p.Register2(engine.NewAtom("get_status"), func(_ *engine.VM, url, status engine.Term, k engine.Cont, env *engine.Env) *engine.Promise {
		// Check if the input arguments are of the types you expected.
		u, ok := env.Resolve(url).(engine.Atom)
		if !ok {
			return engine.Error(engine.TypeError(engine.NewAtom("atom"), url, env))
		}

		// Do whatever you want with the given inputs.
		resp, err := http.Get(u.String())
		if err != nil {
			return engine.Error(err)
		}

		// Return values by unification with the output arguments.
		env, ok = env.Unify(status, engine.Integer(resp.StatusCode))
		if !ok {
			return engine.Bool(false)
		}

		// Tell Prolog to continue with the given continuation and environment.
		return k(env)
	})

	// Treat a string argument as an atom.
	err2 := p.Exec(`:- set_prolog_flag(double_quotes, atom).`)
	if err2 != nil {
		panic(err2)
	}

	// Query with the custom predicate get_status/2 but parameterize the first argument.
	sols, err := p.Query(`get_status(?, Status).`, "https://httpbin.org/status/200")
	if err != nil {
		panic(err)
	}
	defer func() {
		err3 := sols.Close()
		if err3 != nil {
			panic(err3)
		}
	}()

	if !sols.Next() {
		panic("no solutions")
	}

	var s struct {
		Status int
	}
	err3 := sols.Scan(&s)
	if err3 != nil {
		panic(err3)
	}

	fmt.Printf("%+v\n", s)
}
