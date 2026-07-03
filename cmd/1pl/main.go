package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"sort"
	"strings"

	"golang.org/x/term"

	"github.com/crgimenes/prolog"
	"github.com/crgimenes/prolog/engine"
)

const (
	prompt          = "?- "
	contPrompt      = "|- "
	userInputPrompt = "|: "
)

var version = func() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	return info.Main.Version
}()

func main() {
	flag.Parse()

	fmt.Printf(`Top level for crgimenes/prolog %s
This is for testing purposes only!
See https://github.com/crgimenes/prolog for more details.
Type Ctrl-C or 'halt.' to exit.
`, version)

	halt := engine.Halt
	if term.IsTerminal(0) {
		oldState, err := term.MakeRaw(0)
		if err != nil {
			log.Panicf("failed to enter raw mode: %v", err)
		}
		restore := func() {
			_ = term.Restore(0, oldState)
		}
		defer restore()

		halt = func(vm *engine.VM, n engine.Term, k engine.Cont, env *engine.Env) *engine.Promise {
			restore()
			return engine.Halt(vm, n, k, env)
		}
	}

	t := term.NewTerminal(os.Stdin, prompt)
	defer fmt.Printf("\r\n")

	log.SetOutput(t)

	i := New(&userInput{t: t}, t)
	i.Register1(engine.NewAtom("halt"), halt)
	i.Unknown = func(name engine.Atom, args []engine.Term, env *engine.Env) {
		var sb strings.Builder
		s := engine.NewOutputTextStream(&sb)
		_, _ = engine.WriteTerm(&i.VM, s, name.Apply(args...), engine.List(engine.NewAtom("quoted").Apply(engine.NewAtom("true"))), engine.Success, env).Force(context.Background())
		log.Printf("UNKNOWN %s", &sb)
	}

	// Consult arguments.
	err := i.QuerySolution(`findall(F, (member(X, ?), atom_chars(F, X)), Fs), consult(Fs).`, flag.Args()).Err()
	if err != nil {
		log.Panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var buf strings.Builder
	keys := bufio.NewReader(os.Stdin)
	for {
		err := handleLine(ctx, &buf, i, t, keys)
		switch err {
		case nil:
			// Keep reading.
		case io.EOF:
			return
		default:
			log.Panic(err)
		}
	}
}

func handleLine(ctx context.Context, buf *strings.Builder, p *prolog.Interpreter, t *term.Terminal, keys *bufio.Reader) (err error) {
	line, err := t.ReadLine()
	if err != nil {
		return err
	}
	_, _ = buf.WriteString(line)
	_, _ = buf.WriteString("\n")

	sols, err := p.QueryContext(ctx, buf.String())
	switch err {
	case nil:
		buf.Reset()
		t.SetPrompt(prompt)
	case io.EOF:
		// Returns without resetting buf.
		t.SetPrompt(contPrompt)
		return nil
	default:
		log.Printf("failed to query: %v", err)
		buf.Reset()
		t.SetPrompt(prompt)
		return nil
	}
	defer func() {
		_ = sols.Close()
	}()

	var exists bool
	for sols.Next() {
		exists = true

		m := map[string]prolog.TermString{}
		_ = sols.Scan(m)

		var buf bytes.Buffer
		if len(m) == 0 {
			_, _ = fmt.Fprintf(&buf, "%t", true)
		} else {
			ls := make([]string, 0, len(m))
			for v, t := range m {
				ls = append(ls, fmt.Sprintf("%s = %s", v, t))
			}
			sort.Strings(ls)
			_, _ = fmt.Fprint(&buf, strings.Join(ls, ",\n"))
		}
		_, err2 := t.Write(buf.Bytes())
		if err2 != nil {
			return err2
		}

		r, _, err := keys.ReadRune()
		if err != nil {
			return err
		}
		if r != ';' {
			r = '.'
		}
		_, err3 := fmt.Fprintf(t, "%s\n", string(r))
		if err3 != nil {
			return err3
		}
		if r == '.' {
			break
		}
	}

	err2 := sols.Err()
	if err2 != nil {
		log.Print(err2)
		return nil
	}

	if !exists {
		_, err3 := fmt.Fprintf(t, "%t.\n", false)
		if err3 != nil {
			return err3
		}
	}

	return nil
}

type userInput struct {
	t   *term.Terminal
	buf bytes.Buffer
}

func (u *userInput) Read(p []byte) (n int, err error) {
	if u.buf.Len() == 0 {
		u.t.SetPrompt(userInputPrompt)
		defer u.t.SetPrompt(prompt)
		line, err := u.t.ReadLine()
		if err != nil {
			return 0, err
		}
		u.buf.WriteString(line + "\n")
	}

	return u.buf.Read(p)
}

func (u *userInput) Write(b []byte) (n int, err error) {
	return 0, nil
}
