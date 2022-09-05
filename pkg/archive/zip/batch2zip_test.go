package batch2zip

import (
	"archive/zip"
	"bytes"
	"context"
	"testing"

	s2k "github.com/takanoriyanagitani/go-sql2keyval"

	f2k "github.com/takanoriyanagitani/go-fs2keyval"
)

func TestAll(t *testing.T) {
	t.Parallel()

	t.Run("FilelikeIter2FsRaw", func(t *testing.T) {
		t.Parallel()

		t.Run("empty zip", func(t *testing.T) {
			t.Parallel()

			var vfile bytes.Buffer
			var files2zip f2k.SetFilelikeBatch = FilelikeIter2FsRaw(&vfile)
			e := files2zip(context.Background(), s2k.IterEmptyNew[f2k.FileLike]())
			if nil != e {
				t.Errorf("Unable to create empty zip: %v", e)
			}

			var vzfile *bytes.Reader = bytes.NewReader(vfile.Bytes())
			_, e = zip.NewReader(vzfile, int64(vzfile.Len()))
			if nil != e {
				t.Errorf("Unable to open zip: %v", e)
			}
		})

		t.Run("single empty file", func(t *testing.T) {
			t.Parallel()

			var vfile bytes.Buffer
			var files2zip f2k.SetFilelikeBatch = FilelikeIter2FsRaw(&vfile)
			e := files2zip(context.Background(), s2k.IterFromArray([]f2k.FileLike{
				{
					Path: "./empty",
					Val:  nil,
				},
			}))
			if nil != e {
				t.Errorf("Unable to create empty zip: %v", e)
			}

			var vzfile *bytes.Reader = bytes.NewReader(vfile.Bytes())
			zr, e := zip.NewReader(vzfile, int64(vzfile.Len()))
			if nil != e {
				t.Errorf("Unable to open zip: %v", e)
			}

			f, e := zr.Open("empty")
			if nil != e {
				t.Errorf("Unable to open zip item: %v", e)
			}
			if nil == f {
				t.Errorf("File nill!!")
			} else {
				f.Close()
			}
		})
	})
}
