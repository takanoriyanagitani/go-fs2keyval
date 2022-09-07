package fs2kv

import (
	s2k "github.com/takanoriyanagitani/go-sql2keyval"
)

type Result[T any] interface {
	Value() T
	Error() error
	TryForEach(f func(T) error) error
	UnwrapOrElse(f func(error) T) T
	UnwrapOr(alt T) T
	Map(f func(T) T) Result[T]
	Ok() s2k.Option[T]
}

type resultOk[T any] struct{ val T }

func (r resultOk[T]) Value() T                         { return r.val }
func (r resultOk[T]) Error() error                     { return nil }
func (r resultOk[T]) TryForEach(f func(T) error) error { return f(r.val) }
func (r resultOk[T]) UnwrapOrElse(_ func(error) T) T   { return r.val }
func (r resultOk[T]) UnwrapOr(_ T) T                   { return r.val }
func (r resultOk[T]) Map(f func(T) T) Result[T]        { return ResultNew(f(r.val), nil) }
func (r resultOk[T]) Ok() s2k.Option[T]                { return s2k.OptionNew(r.val) }

func ResultOk[T any](t T) Result[T] { return ResultNew(t, nil) }

type resultNg[T any] struct{ err error }

func (r resultNg[T]) Value() (t T)                     { return }
func (r resultNg[T]) Error() error                     { return r.err }
func (r resultNg[T]) TryForEach(_ func(T) error) error { return r.err }
func (r resultNg[T]) UnwrapOrElse(f func(error) T) T   { return f(r.err) }
func (r resultNg[T]) UnwrapOr(alt T) T                 { return alt }
func (r resultNg[T]) Map(_ func(T) T) Result[T]        { return r }
func (r resultNg[T]) Ok() s2k.Option[T]                { return s2k.OptionEmptyNew[T]() }

func ResultNg[T any](err error) Result[T] { return resultNg[T]{err} }

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

func ResultConv[T, U any](r Result[T], ok func(T) U, ng func(error) U) U {
	if nil == r.Error() {
		return ok(r.Value())
	}
	return ng(r.Error())
}

func ResultIter2iterResults[T any](ri Result[s2k.Iter[T]]) s2k.Iter[Result[T]] {
	return ResultConv(
		ri,
		func(i s2k.Iter[T]) s2k.Iter[Result[T]] {
			return func() s2k.Option[Result[T]] {
				return s2k.OptionMap(i(), ResultOk[T])
			}
		},
		func(e error) s2k.Iter[Result[T]] {
			var or s2k.Option[Result[T]] = s2k.OptionNew(ResultNg[T](e))
			return s2k.IterFromOpt(or)
		},
	)
}

func ResultsFlatten[T any](results s2k.Iter[s2k.Iter[Result[T]]]) s2k.Iter[Result[T]] {
	var ia []s2k.Iter[Result[T]] = results.ToArray()
	var ra []Result[T]
	for _, i := range ia {
		for or := i(); or.HasValue(); or = i() {
			var rt Result[T] = or.Value()
			ra = append(ra, rt)
		}
	}
	return s2k.IterFromArray(ra)
}

func ResultWrapIter[T any](i s2k.Iter[T]) s2k.Iter[Result[T]] {
	return func() s2k.Option[Result[T]] {
		var ot s2k.Option[T] = i()
		return s2k.OptionMap(ot, ResultOk[T])
	}
}
