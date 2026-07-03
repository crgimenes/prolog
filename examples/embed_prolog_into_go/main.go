package main

import (
	"fmt"

	"github.com/crgimenes/prolog"
)

// http://www.cse.unsw.edu.au/~billw/dictionaries/prolog/cut.html
func main() {
	p := prolog.New(nil, nil)
	err := p.Exec(`
teaches(dr_fred, history).
teaches(dr_fred, english).
teaches(dr_fred, drama).
teaches(dr_fiona, physics).

studies(alice, english).
studies(angus, english).
studies(amelia, drama).
studies(alex, physics).
`)
	if err != nil {
		panic(err)
	}

	for _, q := range []string{
		`teaches(dr_fred, Course), studies(Student, Course).`,
		`teaches(dr_fred, Course), !, studies(Student, Course).`,
		`teaches(dr_fred, Course), studies(Student, Course), !.`,
		`!, teaches(dr_fred, Course), studies(Student, Course).`,
	} {
		fmt.Printf("%s\n", q)

		sols, err := p.Query(q)
		if err != nil {
			panic(err)
		}

		for sols.Next() {
			var s struct {
				Course  string
				Student string
			}
			err2 := sols.Scan(&s)
			if err2 != nil {
				panic(err2)
			}
			fmt.Printf("\t%+v\n", s)
		}

		fmt.Printf("\n")
		err2 := sols.Close()
		if err2 != nil {
			panic(err2)
		}
	}
}
