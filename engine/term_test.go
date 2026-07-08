package engine

import (
	"errors"
	"testing"
)

func TestErrWriter_Write(t *testing.T) {
	var failed = errors.New("failed")

	var m mockWriter
	m.err = failed

	ew := errWriter{w: &m}
	_, err := ew.Write([]byte("foo"))
	noError(t, err)
	_, err = ew.Write([]byte("bar"))
	noError(t, err)
	_, err = ew.Write([]byte("baz"))
	noError(t, err)
	equal(t, failed, ew.err)

	// Once it has seen an error, errWriter stops touching the wrapped writer.
	equal(t, 1, m.writes)
}

func TestCompareAtomic(t *testing.T) {
	type x struct {
		mockTerm
	}
	type y struct {
		mockTerm
		val int
	}
	type z struct {
		mockTerm
	}

	cmp := func(y1 *y, y2 *y) int {
		return y1.val - y2.val
	}

	tests := []struct {
		a   *y
		t   Term
		cmp func(*y, *y) int
		o   int
	}{
		{a: &y{}, t: NewVariable(), o: 1},
		{a: &y{}, t: Float(0), o: 1},
		{a: &y{}, t: Integer(0), o: 1},
		{a: &y{}, t: Atom(0), o: 1},
		{a: &y{}, t: &x{}, o: 1},
		{a: &y{val: 1}, t: &y{val: 0}, cmp: cmp, o: 1},
		{a: &y{val: 0}, t: &y{val: 0}, cmp: cmp, o: 0},
		{a: &y{val: 0}, t: &y{val: 1}, cmp: cmp, o: -1},
		{a: &y{}, t: &z{}, o: -1},
		{a: &y{}, t: Atom(0).Apply(Integer(0)), o: -1},
	}

	for _, tt := range tests {
		equal(t, tt.o, CompareAtomic[*y](tt.a, tt.t, tt.cmp, nil))
	}
}
