package main

import (
	"flag"
	"fmt"

	"github.com/crgimenes/prolog"
)

// This example explains how to parse a simple English sentence with DCG (Definite Clause Grammar).
// You can check if it parses a sentence with `go run examples/dcg/main.go <SENTENCE>`. If it does, the program returns
// `true` otherwise `false`. Also, you can generate every possible sentence by providing a prefix
// `go run examples/dcg/main.go -prefix <PREFIX>`.
//
// e.g.)
//   $ go run examples/dcg/main.go the cat chases the mouse
//   $ go run examples/dcg/main.go -prefix the cat

func main() {
	var prefix bool
	flag.BoolVar(&prefix, "prefix", false, "prefix search mode")
	flag.Parse()

	// First, create a Prolog interpreter.
	i := prolog.New(nil, nil)

	// Then, define DCG rules with -->/2.
	err2 := i.Exec(`
:- set_prolog_flag(double_quotes, atom).

sentence --> noun_phrase, verb_phrase.
verb_phrase --> verb.
noun_phrase --> article, noun.
noun_phrase --> article, adjective, noun.
article --> [the].
adjective --> [nice].
noun --> [dog].
noun --> [cat].
verb --> [runs].
verb --> [barks].
verb --> [bites].
`)
	if err2 != nil {
		panic(err2)
	}

	// Finally, query with phrase/2.
	if prefix {
		sols, err := i.Query(`Prefix = ?, append(Prefix, _, Sentence), phrase(sentence, Sentence).`, flag.Args())
		if err != nil {
			panic(err)
		}
		defer func() {
			err3 := sols.Close()
			if err3 != nil {
				panic(err3)
			}
		}()

		for sols.Next() {
			var s struct {
				Sentence []string
			}
			err3 := sols.Scan(&s)
			if err3 != nil {
				panic(err3)
			}

			fmt.Printf("%s\n", s.Sentence)
		}
		return
	}

	sols, err := i.Query(`phrase(sentence, ?).`, flag.Args())
	if err != nil {
		panic(err)
	}
	defer func() {
		err3 := sols.Close()
		if err3 != nil {
			panic(err3)
		}
	}()

	fmt.Printf("%t\n", sols.Next())
}
