package engine

import (
	"testing"
)

func TestListIterator_Next(t *testing.T) {
	t.Run("proper list", func(t *testing.T) {
		iter := ListIterator{List: List(NewAtom("a"), NewAtom("b"), NewAtom("c"))}
		isTrue(t, iter.Next())
		equal(t, NewAtom("a"), iter.Current())
		isTrue(t, iter.Next())
		equal(t, NewAtom("b"), iter.Current())
		isTrue(t, iter.Next())
		equal(t, NewAtom("c"), iter.Current())
		isFalse(t, iter.Next())
		noError(t, iter.Err())
	})

	t.Run("improper list", func(t *testing.T) {
		t.Run("variable", func(t *testing.T) {
			iter := ListIterator{List: PartialList(NewVariable(), NewAtom("a"), NewAtom("b"))}
			isTrue(t, iter.Next())
			equal(t, NewAtom("a"), iter.Current())
			isTrue(t, iter.Next())
			equal(t, NewAtom("b"), iter.Current())
			isFalse(t, iter.Next())
			equal(t, InstantiationError(nil), iter.Err())
		})

		t.Run("atom", func(t *testing.T) {
			iter := ListIterator{List: PartialList(NewAtom("foo"), NewAtom("a"), NewAtom("b"))}
			isTrue(t, iter.Next())
			equal(t, NewAtom("a"), iter.Current())
			isTrue(t, iter.Next())
			equal(t, NewAtom("b"), iter.Current())
			isFalse(t, iter.Next())
			equal(t, typeError(validTypeList, PartialList(NewAtom("foo"), NewAtom("a"), NewAtom("b")), nil), iter.Err())
		})

		t.Run("compound", func(t *testing.T) {
			iter := ListIterator{List: PartialList(NewAtom("f").Apply(Integer(0)), NewAtom("a"), NewAtom("b"))}
			isTrue(t, iter.Next())
			equal(t, NewAtom("a"), iter.Current())
			isTrue(t, iter.Next())
			equal(t, NewAtom("b"), iter.Current())
			isFalse(t, iter.Next())
			equal(t, typeError(validTypeList, PartialList(NewAtom("f").Apply(Integer(0)), NewAtom("a"), NewAtom("b")), nil), iter.Err())
		})

		t.Run("other", func(t *testing.T) {
			iter := ListIterator{List: PartialList(&mockTerm{}, NewAtom("a"), NewAtom("b"))}
			isTrue(t, iter.Next())
			equal(t, NewAtom("a"), iter.Current())
			isTrue(t, iter.Next())
			equal(t, NewAtom("b"), iter.Current())
			isFalse(t, iter.Next())
			equal(t, typeError(validTypeList, PartialList(&mockTerm{}, NewAtom("a"), NewAtom("b")), nil), iter.Err())
		})

		t.Run("circular list", func(t *testing.T) {
			l := NewVariable()
			const max = 500
			elems := make([]Term, 0, max)
			for range max {
				elems = append(elems, NewAtom("a"))
				env := NewEnv().bind(l, PartialList(l, elems...))
				iter := ListIterator{List: l, Env: env}
				for iter.Next() {
					equal(t, NewAtom("a"), iter.Current())
				}
				equal(t, typeError(validTypeList, l, env), iter.Err())
			}
		})
	})
}

func TestListIterator_Suffix(t *testing.T) {
	iter := ListIterator{List: List(NewAtom("a"), NewAtom("b"), NewAtom("c"))}
	equal(t, List(NewAtom("a"), NewAtom("b"), NewAtom("c")), iter.Suffix())
	isTrue(t, iter.Next())
	equal(t, List(NewAtom("b"), NewAtom("c")), iter.Suffix())
	isTrue(t, iter.Next())
	equal(t, List(NewAtom("c")), iter.Suffix())
	isTrue(t, iter.Next())
	equal(t, List(), iter.Suffix())
	isFalse(t, iter.Next())
}

func TestSeqIterator_Next(t *testing.T) {
	t.Run("sequence", func(t *testing.T) {
		iter := seqIterator{Seq: seq(atomComma, NewAtom("a"), NewAtom("b"), NewAtom("c"))}
		isTrue(t, iter.Next())
		equal(t, NewAtom("a"), iter.Current())
		isTrue(t, iter.Next())
		equal(t, NewAtom("b"), iter.Current())
		isTrue(t, iter.Next())
		equal(t, NewAtom("c"), iter.Current())
		isFalse(t, iter.Next())
	})

	t.Run("sequence with a trailing compound", func(t *testing.T) {
		iter := seqIterator{Seq: seq(atomComma, NewAtom("a"), NewAtom("b"), NewAtom("f").Apply(NewAtom("c")))}
		isTrue(t, iter.Next())
		equal(t, NewAtom("a"), iter.Current())
		isTrue(t, iter.Next())
		equal(t, NewAtom("b"), iter.Current())
		isTrue(t, iter.Next())
		equal(t, NewAtom("f").Apply(NewAtom("c")), iter.Current())
		isFalse(t, iter.Next())
	})
}

func TestAltIterator_Next(t *testing.T) {
	t.Run("alternatives", func(t *testing.T) {
		iter := altIterator{Alt: seq(atomSemiColon, NewAtom("a"), NewAtom("b"), NewAtom("c"))}
		isTrue(t, iter.Next())
		equal(t, NewAtom("a"), iter.Current())
		isTrue(t, iter.Next())
		equal(t, NewAtom("b"), iter.Current())
		isTrue(t, iter.Next())
		equal(t, NewAtom("c"), iter.Current())
		isFalse(t, iter.Next())
	})

	t.Run("alternatives with a trailing compound", func(t *testing.T) {
		iter := altIterator{Alt: seq(atomSemiColon, NewAtom("a"), NewAtom("b"), NewAtom("f").Apply(NewAtom("c")))}
		isTrue(t, iter.Next())
		equal(t, NewAtom("a"), iter.Current())
		isTrue(t, iter.Next())
		equal(t, NewAtom("b"), iter.Current())
		isTrue(t, iter.Next())
		equal(t, NewAtom("f").Apply(NewAtom("c")), iter.Current())
		isFalse(t, iter.Next())
	})

	t.Run("if then else", func(t *testing.T) {
		iter := altIterator{Alt: seq(atomSemiColon, atomThen.Apply(NewAtom("a"), NewAtom("b")), NewAtom("c"))}
		isTrue(t, iter.Next())
		equal(t, seq(atomSemiColon, atomThen.Apply(NewAtom("a"), NewAtom("b")), NewAtom("c")), iter.Current())
		isFalse(t, iter.Next())
	})
}

func TestAnyIterator_Next(t *testing.T) {
	t.Run("proper list", func(t *testing.T) {
		iter := anyIterator{Any: List(NewAtom("a"), NewAtom("b"), NewAtom("c"))}
		isTrue(t, iter.Next())
		equal(t, NewAtom("a"), iter.Current())
		isTrue(t, iter.Next())
		equal(t, NewAtom("b"), iter.Current())
		isTrue(t, iter.Next())
		equal(t, NewAtom("c"), iter.Current())
		isFalse(t, iter.Next())
		noError(t, iter.Err())
	})

	t.Run("improper list", func(t *testing.T) {
		t.Run("variable", func(t *testing.T) {
			iter := anyIterator{Any: PartialList(NewVariable(), NewAtom("a"), NewAtom("b"))}
			isTrue(t, iter.Next())
			equal(t, NewAtom("a"), iter.Current())
			isTrue(t, iter.Next())
			equal(t, NewAtom("b"), iter.Current())
			isFalse(t, iter.Next())
			equal(t, InstantiationError(nil), iter.Err())
		})

		t.Run("atom", func(t *testing.T) {
			iter := anyIterator{Any: PartialList(NewAtom("foo"), NewAtom("a"), NewAtom("b"))}
			isTrue(t, iter.Next())
			equal(t, NewAtom("a"), iter.Current())
			isTrue(t, iter.Next())
			equal(t, NewAtom("b"), iter.Current())
			isFalse(t, iter.Next())
			equal(t, typeError(validTypeList, PartialList(NewAtom("foo"), NewAtom("a"), NewAtom("b")), nil), iter.Err())
		})
	})

	t.Run("sequence", func(t *testing.T) {
		iter := anyIterator{Any: seq(atomComma, NewAtom("a"), NewAtom("b"), NewAtom("c"))}
		isTrue(t, iter.Next())
		equal(t, NewAtom("a"), iter.Current())
		isTrue(t, iter.Next())
		equal(t, NewAtom("b"), iter.Current())
		isTrue(t, iter.Next())
		equal(t, NewAtom("c"), iter.Current())
		isFalse(t, iter.Next())
		noError(t, iter.Err())
	})

	t.Run("single", func(t *testing.T) {
		iter := anyIterator{Any: NewAtom("a")}
		isTrue(t, iter.Next())
		equal(t, NewAtom("a"), iter.Current())
		isFalse(t, iter.Next())
		noError(t, iter.Err())
	})
}
