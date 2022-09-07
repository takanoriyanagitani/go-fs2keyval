package batch2zip

import (
	"archive/zip"
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

	t.Run("Files2ZipBuilderStore", func(t *testing.T) {
		t.Parallel()

		t.Run("empty", func(t *testing.T) {
			t.Parallel()

			var wtr bytes.Buffer
			var files2z f2k.SetFsFileBatch = Files2ZipBuilderStore(&wtr)

			e := files2z(context.Background(), s2k.IterEmptyNew[fs.File]())
			if nil != e {
				t.Fatalf("Unable to create empty zip: %v", e)
			}

			var rdr *bytes.Reader = bytes.NewReader(wtr.Bytes())
			zr, e := zip.NewReader(rdr, rdr.Size())
			if nil != e {
				t.Fatalf("Invalid zip: %v", e)
			}

			checker(t, len(zr.File), 0)
		})

		t.Run("single empty file", func(t *testing.T) {
			t.Parallel()

			var wtr bytes.Buffer
			var files2z f2k.SetFsFileBatch = Files2ZipBuilderStore(&wtr)

			e := files2z(context.Background(), s2k.IterFromArray([]fs.File{
				f2k.MemFileNew("f1", nil, 0644),
			}))
			if nil != e {
				t.Fatalf("Unable to create empty zip: %v", e)
			}

			var rdr *bytes.Reader = bytes.NewReader(wtr.Bytes())
			zr, e := zip.NewReader(rdr, rdr.Size())
			if nil != e {
				t.Fatalf("Invalid zip: %v", e)
			}

			checker(t, len(zr.File), 1)

			var zf *zip.File = zr.File[0]

			checker(t, zf.Name, "f1")
			checker(t, zf.Method, zip.Store)
			checker(t, zf.UncompressedSize64, 0)
		})

		t.Run("many files", func(t *testing.T) {
			t.Parallel()

			var wtr bytes.Buffer
			var files2z f2k.SetFsFileBatch = Files2ZipBuilderStore(&wtr)

			e := files2z(context.Background(), s2k.IterFromArray([]fs.File{
				f2k.MemFileNew("f1", []byte("hw"), 0644),
				f2k.MemFileNew("f2", []byte("hx"), 0644),
			}))
			if nil != e {
				t.Fatalf("Unable to create empty zip: %v", e)
			}

			var rdr *bytes.Reader = bytes.NewReader(wtr.Bytes())
			zr, e := zip.NewReader(rdr, rdr.Size())
			if nil != e {
				t.Fatalf("Invalid zip: %v", e)
			}

			checker(t, len(zr.File), 2)

			checkerNew := func(zf *zip.File, name string, expected []byte) func(*testing.T) {
				return func(t *testing.T) {
					t.Parallel()

					checker(t, zf.Name, name)
					rc, e := zf.Open()
					if nil != e {
						t.Fatalf("Unable to open zip: %v", e)
					}
					defer rc.Close()

					got, e := io.ReadAll(rc)
					if nil != e {
						t.Fatalf("Unable to read: %v", e)
					}

					checkBytes(t, got, expected)
				}
			}

			t.Run("chk f1", checkerNew(zr.File[0], "f1", []byte("hw")))
			t.Run("chk f2", checkerNew(zr.File[1], "f2", []byte("hx")))
		})
	})
}
