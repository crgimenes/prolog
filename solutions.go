package prolog

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/crgimenes/prolog/engine"
)

// ErrClosed indicates the Solutions are already closed and unable to perform the operation.
var ErrClosed = errors.New("closed")

var errConversion = errors.New("conversion failed")

// Solutions is the result of a query. Everytime the Next method is called, it searches for the next solution.
// By calling the Scan method, you can retrieve the content of the solution.
type Solutions struct {
	vm     *engine.VM
	env    *engine.Env
	vars   []engine.ParsedVariable
	more   chan<- bool
	next   <-chan *engine.Env
	err    error
	closed bool
}

// Close closes the Solutions and terminates the search for other solutions.
func (s *Solutions) Close() error {
	if s.closed {
		return ErrClosed
	}
	close(s.more)
	s.closed = true
	return nil
}

// Next prepares the next solution for reading with the Scan method. It returns true if it finds another solution,
// or false if there's no further solutions or if there's an error.
func (s *Solutions) Next() bool {
	if s.closed {
		return false
	}
	s.more <- true
	var ok bool
	s.env, ok = <-s.next
	return ok
}

// Scan copies the variable values of the current solution into the specified struct/map.
func (s *Solutions) Scan(dest any) error {
	o := reflect.ValueOf(dest)
	for o.Kind() == reflect.Pointer {
		o = o.Elem()
	}
	switch o.Kind() {
	case reflect.Struct:
		t := o.Type()

		fields := make(map[string]any, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			name := f.Name
			alias, ok := f.Tag.Lookup("prolog")
			if ok {
				name = alias
			}
			fields[name] = o.Field(i).Addr().Interface()
		}

		for _, v := range s.vars {
			n := v.Name.String()
			f, ok := fields[n]
			if !ok {
				continue
			}

			err := convertAssign(f, s.vm, v.Variable, s.env)
			if err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		t := o.Type()
		if t.Key() != reflect.TypeFor[string]() {
			return errors.New("map key is not string")
		}

		for _, v := range s.vars {
			dest := reflect.New(t.Elem())
			err := convertAssign(dest.Interface(), s.vm, v.Variable, s.env)
			if err != nil {
				return err
			}
			o.SetMapIndex(reflect.ValueOf(v.Name.String()), dest.Elem())
		}
		return nil
	default:
		return fmt.Errorf("invalid kind: %s", o.Kind())
	}
}

var atomEmptyList = engine.NewAtom("[]")

func convertAssign(dest any, vm *engine.VM, t engine.Term, env *engine.Env) error {
	switch d := dest.(type) {
	case *any:
		return convertAssignAny(d, vm, t, env)
	case *string:
		return convertAssignString(d, t, env)
	case *int:
		return convertAssignInteger(d, t, env)
	case *int8:
		return convertAssignInteger(d, t, env)
	case *int16:
		return convertAssignInteger(d, t, env)
	case *int32:
		return convertAssignInteger(d, t, env)
	case *int64:
		return convertAssignInteger(d, t, env)
	case *float32:
		return convertAssignFloat(d, t, env)
	case *float64:
		return convertAssignFloat(d, t, env)
	case Scanner:
		return d.Scan(vm, t, env)
	default:
		return convertAssignSlice(d, vm, t, env)
	}
}

func convertAssignAny(d *any, vm *engine.VM, t engine.Term, env *engine.Env) error {
	switch t := env.Resolve(t).(type) {
	case engine.Variable:
		*d = nil
		return nil
	case engine.Atom:
		if t == atomEmptyList {
			*d = []any{}
		} else {
			*d = t.String()
		}
		return nil
	case engine.Integer:
		*d = int(t)
		return nil
	case engine.Float:
		*d = float64(t)
		return nil
	case engine.Compound:
		var s []any
		iter := engine.ListIterator{List: t, Env: env}
		for iter.Next() {
			s = append(s, nil)
			err := convertAssign(&s[len(s)-1], vm, iter.Current(), env)
			if err != nil {
				return err
			}
		}
		err := iter.Err()
		if err != nil {
			return errConversion
		}
		*d = s
		return nil
	default:
		return errConversion
	}
}

func convertAssignString(d *string, t engine.Term, env *engine.Env) error {
	switch t := env.Resolve(t).(type) {
	case fmt.Stringer:
		*d = t.String()
		return nil
	default:
		return errConversion
	}
}

// convertAssignInteger converts a Prolog integer into any Go signed integer
// type. A value outside the destination's range is a conversion error, not a
// silent wrap-around.
func convertAssignInteger[D ~int | ~int8 | ~int16 | ~int32 | ~int64](d *D, t engine.Term, env *engine.Env) error {
	switch t := env.Resolve(t).(type) {
	case engine.Integer:
		if int64(D(t)) != int64(t) {
			return errConversion
		}
		*d = D(t)
		return nil
	default:
		return errConversion
	}
}

// convertAssignFloat converts a Prolog float into a Go float type; narrowing
// to float32 keeps the nearest representable value, like any Go conversion.
func convertAssignFloat[D ~float32 | ~float64](d *D, t engine.Term, env *engine.Env) error {
	switch t := env.Resolve(t).(type) {
	case engine.Float:
		*d = D(t)
		return nil
	default:
		return errConversion
	}
}

func convertAssignSlice(d any, vm *engine.VM, t engine.Term, env *engine.Env) error {
	v := reflect.ValueOf(d).Elem()

	k := v.Kind()
	if k != reflect.Slice {
		return errConversion
	}

	v.SetLen(0)
	orig := v

	iter := engine.ListIterator{List: t, Env: env}
	for iter.Next() {
		v = reflect.Append(v, reflect.Zero(v.Type().Elem()))
		dest := v.Index(v.Len() - 1).Addr().Interface()
		err := convertAssign(dest, vm, iter.Current(), env)
		if err != nil {
			return err
		}
	}
	err := iter.Err()
	if err != nil {
		return errConversion
	}

	orig.Set(v)

	return nil
}

// Err returns the error if exists.
func (s *Solutions) Err() error {
	return s.err
}

// Solution is the single result of a query.
type Solution struct {
	sols *Solutions
	err  error
}

// Scan copies the variable values of the solution into the specified struct/map.
func (s *Solution) Scan(dest any) error {
	err := s.err
	if err != nil {
		return err
	}
	return s.sols.Scan(dest)
}

// Err returns an error that occurred while querying for the Solution, if any.
func (s *Solution) Err() error {
	return s.err
}

// Scanner is an interface for custom conversion from term to Go value.
type Scanner interface {
	Scan(vm *engine.VM, term engine.Term, env *engine.Env) error
}

// TermString is a string representation of term.
type TermString string

// Scan implements Scanner interface.
func (t *TermString) Scan(vm *engine.VM, term engine.Term, env *engine.Env) error {
	var sb strings.Builder
	s := engine.NewOutputTextStream(&sb)
	_, _ = engine.WriteTerm(vm, s, term, engine.List(engine.NewAtom("quoted").Apply(engine.NewAtom("true"))), engine.Success, env).Force(context.Background())
	*t = TermString(sb.String())
	return nil
}
