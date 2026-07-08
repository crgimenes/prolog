package main

import (
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("go_string", func(t *testing.T) {
		p := New(nil, nil)
		if err := p.QuerySolution(`go_string("foo", '"foo"').`).Err(); err != nil {
			t.Fatal(err)
		}
	})
}
