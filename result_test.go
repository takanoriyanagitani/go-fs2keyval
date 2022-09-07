package fs2kv

import (
	"fmt"
	"math"
	"os"
	"testing"

	s2k "github.com/takanoriyanagitani/go-sql2keyval"
)

func TestResult(t *testing.T) {
	t.Parallel()

	var div func(a, b int) (int, error) = func(a, b int) (int, error) {
		if 0 == b {
			return 0, fmt.Errorf("Division by 0")
		}
		return a / b, nil
	}

	var cdiv func(a int) func(b int) (int, error) = CurryError(div)
	var div42 func(b int) (int, error) = cdiv(42)

	var fdiv func(a, b float64) Result[float64] = func(a, b float64) Result[float64] {
		if 0.0 == b {
			return ResultNew(0.0, fmt.Errorf("Division by 0"))
		}
		return ResultNew(a/b, nil)
	}
	var cfdiv func(a float64) func(b float64) Result[float64] = Curry(fdiv)

	var mul func(a, b int) int = func(a, b int) int { return a * b }
	var cmul func(a int) func(b int) int = Curry(mul)
	var mul42 func(b int) int = cmul(42)

	var nan2result func(float64) Result[float64] = func(f float64) Result[float64] {
		if math.IsNaN(f) {
			return ResultNew(0.0, fmt.Errorf("NaN"))
		}
		return ResultNew(f, nil)
	}

	var sqrt2result func(float64) Result[float64] = s2k.Compose(math.Sqrt, nan2result)
	var rcp func(float64) Result[float64] = cfdiv(1.0)

	var fmul func(a, b float64) float64 = func(a, b float64) float64 { return a * b }
	var cfml func(a float64) func(b float64) float64 = Curry(fmul)
	var fml21 func(b float64) float64 = cfml(21.0)

	t.Run("ResultNew", func(t *testing.T) {
		t.Parallel()

		var iok Result[int] = ResultNew(div(84, 2))
		checkResult(iok, func(e error) { t.Fatalf("Unexpected error: %v", e) })
		checker(t, iok.Value(), 42)

		var ing Result[int] = ResultNew(div(42, 0))
		checkerMsg(t, func() bool { return nil != ing.Error() }, "Error nil")
		checker(t, ing.Value(), 0)
	})

	t.Run("ResultBuilderNew1", func(t *testing.T) {
		t.Parallel()

		var rdiv42 func(int) Result[int] = ResultBuilderNew1(div42)

		var db0 Result[int] = rdiv42(0)
		checkerMsg(t, func() bool { return nil != db0.Error() }, "Error nil")
	})

	t.Run("ResultBuilderNew0", func(t *testing.T) {
		t.Parallel()

		var whoami func() Result[string] = ResultBuilderNew0(os.Executable)
		var rs Result[string] = whoami()
		s, e := os.Executable()
		checker(t, rs.Value(), s)
		checker(t, nil == rs.Error(), nil == e)
	})

	t.Run("ResultMap", func(t *testing.T) {
		t.Parallel()

		var iok Result[int] = ResultNew(div(84, 84))
		var ri Result[int] = ResultMap(iok, mul42)
		checker(t, ri.Value(), 42)

		var ing Result[int] = ResultNew(div(42, 0))
		var r2 Result[int] = ResultMap(ing, mul42)
		checkerMsg(t, func() bool { return nil != r2.Error() }, "Must fail")
	})

	t.Run("ResultCompose", func(t *testing.T) {
		t.Parallel()

		var rsqrt func(float64) Result[float64] = ResultCompose(
			sqrt2result,
			rcp,
		)

		checker(t, rsqrt(1.00).Value(), 1.0)
		checker(t, rsqrt(4.00).Value(), 0.5)
		checker(t, rsqrt(0.25).Value(), 2.0)

		checkerMsg(t, func() bool { return nil != rsqrt(-1.0).Error() }, "Must fail")
		checkerMsg(t, func() bool { return nil != rsqrt(0.0).Error() }, "Must fail")
	})

	t.Run("TryForEach", func(t *testing.T) {
		t.Parallel()
		var ok1 Result[int] = ResultNew(42, nil)
		ignore := func(_ int) error { return nil }
		checkErr(ok1.TryForEach(ignore), func(e error) { t.Errorf("Unexpected error: %v", e) })

		alwaysErr := func(_ int) error { return fmt.Errorf("Must fail") }
		checkerMsg(t, func() bool { return nil != ok1.TryForEach(alwaysErr) }, "Must fail")

		var ng1 Result[int] = ResultNew(0, fmt.Errorf("Must fail"))
		checkerMsg(t, func() bool { return nil != ng1.TryForEach(ignore) }, "Must fail")
	})

	t.Run("UnwrapOrElse", func(t *testing.T) {
		t.Parallel()
		var ok1 Result[int] = ResultNew(42, nil)
		zero := func(error) int { return 0 }
		checker(t, ok1.UnwrapOrElse(zero), 42)

		var ng1 Result[int] = ResultNew(0, fmt.Errorf("Must fail"))
		nz := func(error) int { return 42 }
		checker(t, ng1.UnwrapOrElse(nz), 42)
	})

	t.Run("Map", func(t *testing.T) {
		t.Parallel()
		var rf Result[float64] = sqrt2result(4.0)
		var r2 Result[float64] = rf.Map(fml21)
		checker(t, r2.Value(), 42.0)

		var rng Result[float64] = sqrt2result(-1.0)
		var ng2 Result[float64] = rng.Map(fml21)
		checkerMsg(t, func() bool { return nil != ng2.Error() }, "Must fail")
	})

	t.Run("Ok", func(t *testing.T) {
		t.Parallel()
		var rok Result[int] = ResultOk(42)
		var oi s2k.Option[int] = rok.Ok()
		checker(t, oi.Value(), 42)

		var rng Result[float64] = sqrt2result(-1.0)
		var ong s2k.Option[float64] = rng.Ok()
		checker(t, ong.HasValue(), false)
	})

	t.Run("ResultNg", func(t *testing.T) {
		t.Parallel()
		var rng Result[int] = ResultNg[int](fmt.Errorf("Must fail"))
		checkerMsg(t, func() bool { return nil != rng.Error() }, "Must fail")
	})

	t.Run("ResultConv", func(t *testing.T) {
		t.Parallel()

		var rok Result[float64] = ResultOk(42.0)
		var f2i func(float64) int = func(f float64) int { return int(f) }
		var iok int = ResultConv(rok, f2i, func(_ error) int { return -1 })
		checker(t, iok, 42)

		var rng Result[float64] = sqrt2result(-9.0)
		var ing int = ResultConv(rng, f2i, func(_ error) int { return -1 })
		checker(t, ing, -1)
	})

	t.Run("ResultIter2iterResults", func(t *testing.T) {
		t.Parallel()

		t.Run("empty", func(t *testing.T) {
			t.Parallel()

			var riok Result[s2k.Iter[int]] = ResultOk(s2k.IterEmptyNew[int]())
			var irok s2k.Iter[Result[int]] = ResultIter2iterResults(riok)
			var orok s2k.Option[Result[int]] = irok()
			checker(t, orok.HasValue(), false)
		})

		t.Run("single", func(t *testing.T) {
			t.Parallel()

			var riok Result[s2k.Iter[int]] = ResultOk(s2k.IterFromArray([]int{42}))
			var irok s2k.Iter[Result[int]] = ResultIter2iterResults(riok)
			var orok s2k.Option[Result[int]] = irok()

			checker(t, orok.HasValue(), true)
			var rok Result[int] = orok.Value()
			checker(t, rok.Value(), 42)
		})

		t.Run("multi", func(t *testing.T) {
			t.Parallel()

			var riok Result[s2k.Iter[int]] = ResultOk(s2k.IterFromArray([]int{333, 634}))
			var irok s2k.Iter[Result[int]] = ResultIter2iterResults(riok)

			chk := func(expected int) {
				var orok s2k.Option[Result[int]] = irok()
				checker(t, orok.HasValue(), true)
				var rok Result[int] = orok.Value()
				checker(t, rok.Value(), expected)
			}

			chk(333)
			chk(634)

			checker(t, irok().HasValue(), false)

			var ring Result[s2k.Iter[int]] = ResultNg[s2k.Iter[int]](fmt.Errorf("Must fail"))
			var irng s2k.Iter[Result[int]] = ResultIter2iterResults(ring)

			var orng s2k.Option[Result[int]] = irng()
			checker(t, orng.HasValue(), true)
			var rng Result[int] = orng.Value()
			checkerMsg(t, func() bool { return nil != rng.Error() }, "Must fail")
		})
	})

	t.Run("ResultsFlatten", func(t *testing.T) {
		t.Parallel()

		t.Run("empty", func(t *testing.T) {
			t.Parallel()

			var iir s2k.Iter[s2k.Iter[Result[int]]] = s2k.IterEmptyNew[s2k.Iter[Result[int]]]()
			var ir s2k.Iter[Result[int]] = ResultsFlatten(iir)
			var or s2k.Option[Result[int]] = ir()
			checker(t, or.HasValue(), false)
		})

		t.Run("many", func(t *testing.T) {
			t.Parallel()

			var iir s2k.Iter[s2k.Iter[Result[int]]] = s2k.IterFromArray([]s2k.Iter[Result[int]]{
				s2k.IterFromArray([]Result[int]{
					ResultOk(599),
					ResultOk(3776),
				}),
				s2k.IterFromArray([]Result[int]{
					ResultOk(333),
					ResultOk(634),
				}),
			})

			var ir s2k.Iter[Result[int]] = ResultsFlatten(iir)
			var ii s2k.Iter[int] = s2k.IterMap(ir, func(ri Result[int]) int { return ri.UnwrapOr(0) })
			var iadd func(a, b int) int = func(a, b int) int { return a + b }
			var i int = s2k.IterReduce(ii, 0, iadd)
			checker(t, i, 599+3776+333+634)
		})
	})

}
