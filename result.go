package fs2kv

import (
	s2k "github.com/takanoriyanagitani/go-sql2keyval"
)

type Result[T any] interface {
	Value() T
	Error() error
	TryForEach(f func(T) error) error
	UnwrapOrElse(f func() T) T
	Map(f func(T) T) Result[T]
}

type resultOk[T any] struct{ val T }

func (r resultOk[T]) Value() T                         { return r.val }
func (r resultOk[T]) Error() error                     { return nil }
func (r resultOk[T]) TryForEach(f func(T) error) error { return f(r.val) }
func (r resultOk[T]) UnwrapOrElse(_ func() T) T        { return r.val }
func (r resultOk[T]) Map(f func(T) T) Result[T]        { return ResultNew(f(r.val), nil) }

type resultNg[T any] struct{ err error }

func (r resultNg[T]) Value() (t T)                     { return }
func (r resultNg[T]) Error() error                     { return r.err }
func (r resultNg[T]) TryForEach(_ func(T) error) error { return r.err }
func (r resultNg[T]) UnwrapOrElse(f func() T) T        { return f() }
func (r resultNg[T]) Map(_ func(T) T) Result[T]        { return r }

func ResultNew[T any](val T, err error) Result[T] {
	if nil == err {
		return resultOk[T]{val}
	}
	return resultNg[T]{err}
}

func ResultBuilderNew0[T any](f func() (T, error)) func() Result[T] {
	return func() Result[T] {
		t, e := f()
		return ResultNew(t, e)
	}
}

func resultPartialBuilder10[T, U any](f func(U) (T, error)) func(U) func() (T, error) {
	return func(u U) func() (T, error) {
		return func() (T, error) {
			return f(u)
		}
	}
}

func ResultBuilderNew1[T, U any](f func(U) (T, error)) func(U) Result[T] {
	var rp func(U) func() (T, error) = resultPartialBuilder10(f)
	return func(u U) Result[T] {
		return s2k.Compose(
			rp,
			ResultBuilderNew0[T],
		)(u)()
	}
}

func ResultFlatMap[T, U any](t Result[T], f func(T) Result[U]) Result[U] {
	if nil == t.Error() {
		return f(t.Value())
	}
	return resultNg[U]{err: t.Error()}
}

func ResultMap[T, U any](t Result[T], f func(T) U) Result[U] {
	return ResultFlatMap(t, func(val T) Result[U] {
		var u U = f(val)
		return ResultNew(u, nil)
	})
}

func ResultCompose[T, U, V any](f func(T) Result[U], g func(U) Result[V]) func(T) Result[V] {
	return func(t T) Result[V] {
		var ru Result[U] = f(t)
		return ResultFlatMap(ru, g)
	}
}
