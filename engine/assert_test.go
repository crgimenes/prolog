package engine

import (
	"bytes"
	"reflect"
	"regexp"
	"testing"
)

// Minimal, non-fatal test assertions in the stdlib style. They report with
// t.Errorf and keep going, matching the behavior the tests were written
// against.

func objectsAreEqual(want, got any) bool {
	if want == nil || got == nil {
		return want == got
	}
	if w, ok := want.([]byte); ok {
		g, ok := got.([]byte)
		if !ok {
			return false
		}
		return bytes.Equal(w, g)
	}
	return reflect.DeepEqual(want, got)
}

func equal(t testing.TB, want, got any) {
	t.Helper()
	if !objectsAreEqual(want, got) {
		t.Errorf("not equal\nwant: %#v\ngot:  %#v", want, got)
	}
}

func noError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func hasError(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("expected an error, got nil")
	}
}

func isTrue(t testing.TB, v bool) {
	t.Helper()
	if !v {
		t.Errorf("expected true, got false")
	}
}

func isFalse(t testing.TB, v bool) {
	t.Helper()
	if v {
		t.Errorf("expected false, got true")
	}
}

func fail(t testing.TB, msg string) {
	t.Helper()
	t.Error(msg)
}

func hasLen(t testing.TB, obj any, n int) {
	t.Helper()
	v := reflect.ValueOf(obj)
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		if v.Len() != n {
			t.Errorf("length: got %d, want %d", v.Len(), n)
		}
	default:
		t.Errorf("hasLen: %T has no length", obj)
	}
}

func isNil(t testing.TB, obj any) {
	t.Helper()
	if !isNilValue(obj) {
		t.Errorf("expected nil, got %#v", obj)
	}
}

func isNilValue(obj any) bool {
	if obj == nil {
		return true
	}
	v := reflect.ValueOf(obj)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func isEmpty(t testing.TB, obj any) {
	t.Helper()
	if !objIsEmpty(obj) {
		t.Errorf("expected empty, got %#v", obj)
	}
}

func objIsEmpty(obj any) bool {
	if obj == nil {
		return true
	}
	v := reflect.ValueOf(obj)
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return v.Len() == 0
	case reflect.Pointer:
		if v.IsNil() {
			return true
		}
		return objIsEmpty(v.Elem().Interface())
	default:
		return v.IsZero()
	}
}

func matchRegexp(t testing.TB, pattern any, s string) {
	t.Helper()
	var re *regexp.Regexp
	switch p := pattern.(type) {
	case *regexp.Regexp:
		re = p
	case string:
		re = regexp.MustCompile(p)
	default:
		t.Errorf("matchRegexp: unsupported pattern %T", pattern)
		return
	}
	if !re.MatchString(s) {
		t.Errorf("expected %q to match %s", s, re.String())
	}
}

func implements(t testing.TB, iface any, obj any) {
	t.Helper()
	ifaceType := reflect.TypeOf(iface).Elem()
	if obj == nil || !reflect.TypeOf(obj).Implements(ifaceType) {
		t.Errorf("%T does not implement %s", obj, ifaceType)
	}
}
