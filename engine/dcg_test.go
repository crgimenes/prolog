package engine

import (
	"context"
	"testing"
)

func TestVM_Phrase(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		var called bool
		vm := VM{
			procedures: map[procedureIndicator]procedure{
				{name: NewAtom("a"), arity: 2}: Predicate2(func(_ *VM, s0, s Term, k Cont, env *Env) *Promise {
					called = true
					return k(env)
				}),
			},
		}

		s0, s := NewVariable(), NewVariable()
		ok, err := Phrase(&vm, NewAtom("a"), s0, s, Success, nil).Force(context.Background())
		noError(t, err)
		isTrue(t, ok)

		isTrue(t, called)
	})

	t.Run("failed", func(t *testing.T) {
		s0, s := NewVariable(), NewVariable()
		var vm VM
		_, err := Phrase(&vm, Integer(0), s0, s, Success, nil).Force(context.Background())
		hasError(t, err)
	})
}
