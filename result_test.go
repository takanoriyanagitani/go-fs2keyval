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

	t.Run("ResultNew", func(t *testing.T) {
		t.Parallel()

		var iok Result[int] = ResultNew(div(84, 2))
		checkResult(iok, func(e error) { t.Fatalf("Unexpected error: %v", e) })
		checker(t, iok.Value(), 42)

		var ing Result[int] = ResultNew(div(42, 0))
		checkerMsg(t, func() bool { return nil != ing.Error() }, "Error nil")
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
}
