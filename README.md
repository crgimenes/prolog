*Attention*: This project started as a fork of [ichiban/prolog](https://github.com/ichiban/prolog) and is not affiliated with the original project.

It now deliberately diverges from upstream and is developed independently — it is the embeddable Prolog used across my own projects (for example as the semantics oracle in [filo](https://github.com/crgimenes/filo)'s conformance tests). It may contain breaking changes without notice.

---

### Install latest version

```console
go get -u github.com/crgimenes/prolog
```

### Usage

#### Instantiate an interpreter

```go
p := prolog.New(os.Stdin, os.Stdout) // Or `prolog.New(nil, nil)` if you don't need user_input/user_output.
```

Or, if you want a sandbox interpreter without any built-in predicates:

```go
// See examples/sandboxing/main.go for details.
p := new(prolog.Interpreter)
```

#### Load a Prolog program

```go
if err := p.Exec(`
	human(socrates).       % This is a fact.
	mortal(X) :- human(X). % This is a rule.
`); err != nil {
	panic(err)
}
```

Similar to `database/sql`, you can use placeholder `?` to insert Go data as Prolog data.

```go
if err := p.Exec(`human(?).`, "socrates"); err != nil { // Same as p.Exec(`human("socrates").`)
	panic(err)
}
```

#### Run the Prolog program

```go
sols, err := p.Query(`mortal(?).`, "socrates") // Same as p.Query(`mortal("socrates").`)
if err != nil {
	panic(err)
}
defer sols.Close()

// Iterates over solutions.
for sols.Next() {
	fmt.Printf("Yes.\n") // ==> Yes.
}

// Check if an error occurred while querying.
if err := sols.Err(); err != nil {
	panic(err)
}
```

Or, if you want to query for the variable values for each solution:

```go
sols, err := p.Query(`mortal(Who).`)
if err != nil {
	panic(err)
}
defer sols.Close()

// Iterates over solutions.
for sols.Next() {
	// Prepare a struct with fields which name corresponds with a variable in the query.
	var s struct {
		Who string
	}
	if err := sols.Scan(&s); err != nil {
		panic(err)
	}
	fmt.Printf("Who = %s\n", s.Who) // ==> Who = socrates
}

// Check if an error occurred while querying.
if err := sols.Err(); err != nil {
	panic(err)
}
```

