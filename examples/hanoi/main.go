package main

import (
	"flag"
	"fmt"

	"github.com/crgimenes/prolog"
	"github.com/crgimenes/prolog/engine"
)

func main() {
	var n int
	flag.IntVar(&n, "n", 3, "the number of disks")
	flag.Parse()

	i := prolog.New(nil, nil)
	err2 := i.Exec(`
hanoi(N) :- move(N, left, right, center).

move(0, _, _, _) :- !.
move(N, X, Y, Z) :-
  M is N - 1,
  move(M, X, Z, Y),
  actuate(X, Y),
  move(M, Z, Y, X).
`)
	if err2 != nil {
		panic(err2)
	}

	i.Register2(engine.NewAtom("actuate"), func(_ *engine.VM, x engine.Term, y engine.Term, k engine.Cont, env *engine.Env) *engine.Promise {
		fmt.Printf("move a disk from %s to %s.\n", env.Resolve(x), env.Resolve(y))
		return k(env)
	})

	sols, err := i.Query(`hanoi(?).`, n)
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
		panic("failed")
	}
}
