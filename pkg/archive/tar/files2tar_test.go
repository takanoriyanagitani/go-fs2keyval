package files2tar

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"io/fs"
	"testing"

	s2k "github.com/takanoriyanagitani/go-sql2keyval"

	f2k "github.com/takanoriyanagitani/go-fs2keyval"
)

func checkerBuilder[T any](comp func(got, expected T) bool) func(t *testing.T, got, expected T) {
	return func(t *testing.T, got, expected T) {
		if !comp(got, expected) {
			t.Errorf("Unexpected value got.\n")
			t.Errorf("expected: %v\n", expected)
			t.Errorf("got:      %v\n", got)
		}
	}
}

func checker[T comparable](t *testing.T, got, expected T) {
	checkerBuilder(func(a, b T) bool { return a == b })(t, got, expected)
}

var checkBytes func(t *testing.T, got, expected []byte) = checkerBuilder(func(got, expected []byte) bool {
	return 0 == bytes.Compare(got, expected)
})

func TestAll(t *testing.T) {
	t.Parallel()

	t.Run("Files2TarBuilderNew", func(t *testing.T) {
		t.Parallel()

		t.Run("empty", func(t *testing.T) {
			var wtr bytes.Buffer
			var setter f2k.SetFsFileBatch = Files2TarBuilderNew(&wtr)
			e := setter(context.Background(), s2k.IterEmptyNew[fs.File]())
			if nil != e {
				t.Fatalf("Unable to write empty tar: %v", e)
			}

			var br *bytes.Reader = bytes.NewReader(wtr.Bytes())
			var tr *tar.Reader = tar.NewReader(br)

			_, e = tr.Next()
			if nil == e {
				t.Errorf("Must be empty")
			}
		})

		t.Run("single empty file", func(t *testing.T) {
			var wtr bytes.Buffer
			var setter f2k.SetFsFileBatch = Files2TarBuilderNew(&wtr)
			e := setter(context.Background(), s2k.IterFromArray([]fs.File{
				f2k.MemFileNew("fn", []byte("hw"), 0644),
			}))
			if nil != e {
				t.Fatalf("Unable to write empty tar: %v", e)
			}

			var br *bytes.Reader = bytes.NewReader(wtr.Bytes())
			var tr *tar.Reader = tar.NewReader(br)

			hdr, e := tr.Next()
			if nil != e {
				t.Errorf("Unable to get header: %v", e)
			}

			checker(t, hdr.Name, "fn")
			checker(t, hdr.Size, 2)

			var buf bytes.Buffer
			n, e := io.Copy(&buf, tr)
			if nil != e && e != io.EOF {
				t.Errorf("Unexpected error: %v", e)
			}

			checker(t, n, 2)
			checkBytes(t, buf.Bytes(), []byte("hw"))
		})

		t.Run("many files", func(t *testing.T) {
			var wtr bytes.Buffer
			var setter f2k.SetFsFileBatch = Files2TarBuilderNew(&wtr)
			e := setter(context.Background(), s2k.IterFromArray([]fs.File{
				f2k.MemFileNew("f1", []byte("hw"), 0644),
				f2k.MemFileNew("f2", []byte("hx"), 0644),
			}))
			if nil != e {
				t.Fatalf("Unable to write empty tar: %v", e)
			}

			var br *bytes.Reader = bytes.NewReader(wtr.Bytes())
			var tr *tar.Reader = tar.NewReader(br)

			manyCheck := func(name string, expected []byte) {
				hdr, e := tr.Next()
				if nil != e {
					t.Errorf("Unable to get header: %v", e)
				}

				checker(t, hdr.Name, name)
				checker(t, hdr.Size, 2)

				var buf bytes.Buffer
				n, e := io.Copy(&buf, tr)
				if nil != e && e != io.EOF {
					t.Errorf("Unexpected error: %v", e)
				}

				checker(t, n, 2)
				checkBytes(t, buf.Bytes(), expected)
			}

			manyCheck("f1", []byte("hw"))
			manyCheck("f2", []byte("hx"))

		})
	})
}
