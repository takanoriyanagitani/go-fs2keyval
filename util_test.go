package fs2kv

import (
	"fmt"
	"testing"
)

func TestUtil(t *testing.T) {
	t.Run("Identity", func(t *testing.T) {
		var fy func(int) int = Identity[int]

		checker(t, fy(1), 1)
		checker(t, fy(2), 2)
		checker(t, fy(3), 3)
		checker(t, fy(9), 9)
	})

	t.Run("Curry", func(t *testing.T) {
		var add func(a, b int) int = func(a, b int) int { return a + b }
		var cadd func(a int) func(b int) int = Curry(add)

		var add2 func(b int) int = cadd(2)

		checker(t, add2(40), 42)
	})

	t.Run("CurryError", func(t *testing.T) {
		var div func(a, b int) (int, error) = func(a, b int) (int, error) {
			if 0 == b {
				return 0, fmt.Errorf("Division by 0")
			}
			return a / b, nil
		}

		var cdiv func(a int) func(b int) (int, error) = CurryError(div)
		var div42 func(b int) (int, error) = cdiv(42)

		i, e := div42(42)
		checker(t, i, 1)
		checkErr(e, func(e error) { t.Errorf("Unexpected error: %v", e) })
	})
}
