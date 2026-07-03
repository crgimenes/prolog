package main

import (
	_ "embed"
	"os"

	"github.com/crgimenes/prolog"
)

//go:embed hello.pl
var hello string

func main() {
	p := prolog.New(nil, os.Stdout)
	err := p.Exec(hello)
	if err != nil {
		panic(err)
	}
}
