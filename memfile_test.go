package fs2kv

import (
	"bytes"
	"io"
	"io/fs"
	"testing"
)

func checkerBuilder[T any](comp func(got, expected T) bool) func(t *testing.T, got, expected T) {
	return func(t *testing.T, got, expected T) {
		if !comp(got, expected) {
			t.Errorf("Unexpected value got.\n")
			t.Errorf("expected: %v\n", expected)
			t.Fatalf("got:      %v\n", got)
		}
	}
}

func checker[T comparable](t *testing.T, got, expected T) {
	checkerBuilder(func(a, b T) bool { return a == b })(t, got, expected)
}

func checkerMsg(t *testing.T, ok func() bool, msg string) {
	if !ok() {
		t.Errorf(msg)
	}
}

func checkErr(e error, ng func(e error)) {
	if nil != e {
		ng(e)
	}
}

func checkResult[T any](r Result[T], emsg func(e error)) {
	if nil != r.Error() {
		emsg(r.Error())
	}
}

var checkBytes func(t *testing.T, got, expected []byte) = checkerBuilder(func(got, expected []byte) bool {
	return 0 == bytes.Compare(got, expected)
})

func TestMemfile(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		var mf *MemFile = MemFileNew("emptyfile", nil, 0644)
		checker(t, mf.IsDir(), false)
		checker(t, mf.Mode(), 0644)
		checker(t, mf.Name(), "emptyfile")
		checker(t, mf.Size(), 0)

		checkerMsg(t, func() bool { return nil == mf.Sys() }, "Non nil sys")
		checkerMsg(t, func() bool { return nil == mf.Close() }, "Non nil error")
		checkerMsg(t, func() bool { return 0 < mf.ModTime().Unix() }, "Negative unixtime.")

		var rfi Result[fs.FileInfo] = ResultNew(mf.Stat())
		checkResult(rfi, func(e error) { t.Errorf("Unable to get file info: %v", e) })

		var rb Result[[]byte] = ResultNew(io.ReadAll(mf))
		checkResult(rb, func(e error) { t.Fatalf("Unable to read: %v", e) })

		checkBytes(t, rb.Value(), nil)
	})
}
