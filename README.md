# prolog

[![test](https://github.com/crgimenes/prolog/actions/workflows/ci.yml/badge.svg)](https://github.com/crgimenes/prolog/actions/workflows/ci.yml)

An embeddable ISO-ish Prolog interpreter for Go, with a `database/sql`-style API:
`Exec` loads clauses, `Query` runs a goal, `Scan` pulls the variable bindings
back into a Go struct.

I use it as the semantics oracle in [filo](https://github.com/crgimenes/filo)'s
conformance tests. Filo's evaluation rules are written as Prolog relations, an
executable spec. Thousands of expressions get evaluated on both sides and
compared. That job set the priorities here: get the arithmetic and term
machinery right, and never let an untrusted program take down the host.

Running untrusted input is a first-class concern. The parser bounds its
recursion depth, so a malformed term like `[*` can't blow the stack, and the
integer/rune conversions range-check before they narrow. Both were real bugs,
found by fuzzing and fixed.

## Install

```console
go get -u github.com/crgimenes/prolog
```

## Usage

### Instantiate an interpreter

```go
p := prolog.New(os.Stdin, os.Stdout) // Or prolog.New(nil, nil) if you don't need user_input/user_output.
```

For a sandbox with no built-in predicates at all, start from a bare interpreter
(see [examples/sandboxing/main.go](examples/sandboxing/main.go)):

```go
p := new(prolog.Interpreter)
```

### Load a program

```go
if err := p.Exec(`
	human(socrates).       % A fact.
	mortal(X) :- human(X). % A rule.
`); err != nil {
	panic(err)
}
```

Like `database/sql`, `?` is a placeholder for injecting Go values as Prolog data:

```go
if err := p.Exec(`human(?).`, "socrates"); err != nil { // Same as p.Exec(`human("socrates").`)
	panic(err)
}
```

### Run a query

```go
sols, err := p.Query(`mortal(?).`, "socrates") // Same as p.Query(`mortal("socrates").`)
if err != nil {
	panic(err)
}
defer sols.Close()

for sols.Next() {
	fmt.Printf("Yes.\n") // ==> Yes.
}

if err := sols.Err(); err != nil {
	panic(err)
}
```

To read the variable bindings out of each solution, scan into a struct whose
field names match the query variables:

```go
sols, err := p.Query(`mortal(Who).`)
if err != nil {
	panic(err)
}
defer sols.Close()

for sols.Next() {
	var s struct {
		Who string
	}
	if err := sols.Scan(&s); err != nil {
		panic(err)
	}
	fmt.Printf("Who = %s\n", s.Who) // ==> Who = socrates
}

if err := sols.Err(); err != nil {
	panic(err)
}
```

Two behaviors worth knowing before you rely on them. `Scan` does not convert
between Integer and Float: a Prolog integer won't fill a `float64` field, and a
float won't fill an `int64`. And `mod` is floored while `rem` is truncated;
`rem` matches Go's `math.Mod`.

## Examples

The [examples/](examples/) directory covers embedding Prolog into Go, calling Go
back from Prolog, DCG, the Towers of Hanoi, custom initialization, and a
locked-down sandbox.

## Acknowledgments

This started as a fork of [ichiban/prolog](https://github.com/ichiban/prolog) by
Yutaka Ichibangase, and the `database/sql`-style API is his design. It has since
diverged into its own codebase, with its own bug fixes, hardening, and
priorities, and is not affiliated with or endorsed by the original project. It
carries breaking changes without notice. The original MIT license and copyright
are kept in [LICENSE](LICENSE). Thanks to the upstream authors for the
foundation.
