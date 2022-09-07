package fs2kv

func Identity[T any](t T) T { return t }

func Curry[T, U, V any](f func(T, U) V) func(T) func(U) V {
	return func(t T) func(U) V {
		return func(u U) V {
			return f(t, u)
		}
	}
}

func CurryError[T, U, V any](f func(T, U) (V, error)) func(T) func(U) (V, error) {
	return func(t T) func(U) (V, error) {
		return func(u U) (V, error) {
			return f(t, u)
		}
	}
}
