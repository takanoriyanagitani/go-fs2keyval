package fs2kv

import (
	"fmt"
	"io/fs"
	"testing"
)

func TestMain(t *testing.T) {
	t.Parallel()

	t.Run("ErrorWarn", func(t *testing.T) {
		t.Parallel()

		t.Run("no error", func(t *testing.T) {
			t.Parallel()

			ef := func() error { return nil }
			wf := func() error { return nil }
			e := ErrorWarn(ef, wf)
			checkErr(e, func(e error) { t.Errorf("Unexpected error: %v", e) })
		})

		t.Run("no critical error", func(t *testing.T) {
			t.Parallel()

			ef := func() error { return nil }
			wf := func() error { return fmt.Errorf("warning") }
			e := ErrorWarn(ef, wf)
			checkerMsg(t, func() bool { return nil != e }, "Must fail")
		})

		t.Run("critical error", func(t *testing.T) {
			t.Parallel()

			ef := func() error { return fmt.Errorf("Must fail") }
			chk := 0
			wf := func() error {
				chk += 1
				return nil
			}
			e := ErrorWarn(ef, wf)
			checkerMsg(t, func() bool { return nil != e }, "Must fail")
			checker(t, chk, 1)
		})
	})

	t.Run("Error1st", func(t *testing.T) {
		t.Parallel()

		t.Run("empty", func(t *testing.T) {
			t.Parallel()

			e := Error1st(nil)
			checkErr(e, func(e error) { t.Errorf("Unexpected error: %v", e) })
		})

		t.Run("single ok", func(t *testing.T) {
			t.Parallel()

			e := Error1st([]func() error{func() error { return nil }})
			checkErr(e, func(e error) { t.Errorf("Unexpected error: %v", e) })
		})

		t.Run("single ng", func(t *testing.T) {
			t.Parallel()

			e := Error1st([]func() error{func() error { return fmt.Errorf("Must fail") }})
			checkerMsg(t, func() bool { return nil != e }, "Must fail")
		})

		t.Run("multi ok", func(t *testing.T) {
			t.Parallel()

			e := Error1st([]func() error{
				func() error { return nil },
				func() error { return nil },
				func() error { return nil },
			})
			checkErr(e, func(e error) { t.Errorf("Unexpected error: %v", e) })
		})

		t.Run("multi ng", func(t *testing.T) {
			t.Parallel()

			e := Error1st([]func() error{
				func() error { return nil },
				func() error { return fmt.Errorf("Must fail") },
				func() error { panic("must skip") },
			})
			checkerMsg(t, func() bool { return nil != e }, "Must fail")
		})
	})

	t.Run("IfOk", func(t *testing.T) {
		t.Parallel()

		t.Run("ng", func(t *testing.T) {
			t.Parallel()

			var e1 error = fmt.Errorf("Must fail")
			var e error = IfOk(e1, nil)
			checkerMsg(t, func() bool { return nil != e }, "Must fail")
			checker(t, e.Error(), e1.Error())
		})

		t.Run("ok -> ng", func(t *testing.T) {
			t.Parallel()

			var e error = IfOk(nil, func() error { return fmt.Errorf("Must fail") })
			checkerMsg(t, func() bool { return nil != e }, "Must fail")
		})

		t.Run("ok -> ok", func(t *testing.T) {
			t.Parallel()

			var e error = IfOk(nil, func() error { return nil })
			checkErr(e, func(e error) { t.Errorf("Unexpected error: %v", e) })
		})
	})

	t.Run("File2Info", func(t *testing.T) {
		t.Parallel()

		var mf fs.File = MemFileNew("filename", []byte("content"), 0644)
		var rfi Result[fs.FileInfo] = File2Info(mf)
		checker(t, rfi.IsOk(), true)

		var fi fs.FileInfo = rfi.Value()
		checker(t, fi.Name(), "filename")
		checker(t, fi.Size(), int64(len("content")))
		checker(t, fi.Mode(), 0644)
	})
}
