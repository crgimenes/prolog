package engine

import (
	"testing"
)

func TestNewException(t *testing.T) {
	equal(t, Exception{term: NewAtom("foo").Apply(NewAtom("bar"))}, NewException(NewAtom("foo").Apply(NewAtom("bar")), nil))

	defer setMemFree(1)()
	equal(t, resourceError(resourceMemory, nil), NewException(NewAtom("foo").Apply(NewVariable(), NewVariable(), NewVariable(), NewVariable(), NewVariable(), NewVariable(), NewVariable(), NewVariable(), NewVariable()), nil))
}

func TestException_Error(t *testing.T) {
	e := Exception{term: NewAtom("foo")}
	equal(t, "foo", e.Error())
}

func TestInstantiationError(t *testing.T) {
	equal(t, Exception{
		term: atomError.Apply(atomInstantiationError, rootContext),
	}, InstantiationError(nil))
}

func TestDomainError(t *testing.T) {
	equal(t, Exception{
		term: atomError.Apply(
			atomDomainError.Apply(atomNotLessThanZero, Integer(-1)),
			rootContext,
		),
	}, DomainError(atomNotLessThanZero, Integer(-1), nil))
}

func TestTypeError(t *testing.T) {
	equal(t, Exception{
		term: atomError.Apply(
			atomTypeError.Apply(atomAtom, Integer(0)),
			rootContext,
		),
	}, TypeError(atomAtom, Integer(0), nil))
}

func TestExceptionalValue_Error(t *testing.T) {
	equal(t, "int_overflow", exceptionalValueIntOverflow.Error())
}
