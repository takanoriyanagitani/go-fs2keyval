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

func TestTar2files(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		buf := make([]byte, 1024)
		var rdr *bytes.Reader = bytes.NewReader(buf)

		var getter f2k.GetFiles = Tar2Files(rdr)
		var got f2k.Result[s2k.Iter[f2k.FileEx]] = getter(context.Background())

		checker(t, got.IsOk(), true)

		var ifiles s2k.Iter[f2k.FileEx] = got.Value()
		var ofile s2k.Option[f2k.FileEx] = ifiles()

		checker(t, ofile.HasValue(), false)
	})

	t.Run("single valid file(invalid batch)", func(t *testing.T) {
		t.Parallel()

		var vtar bytes.Buffer
		var twtr *tar.Writer = tar.NewWriter(&vtar)

		var tbdy []byte = []byte("data_2022_09_13_f00ddeadbeaffacecafe864299792458")
		var thdr *tar.Header = &tar.Header{
			Name: "cafef00ddeadbeafface864299792458/bucket.txt",
			Mode: 0644,
			Size: int64(len(tbdy)),
		}

		e := twtr.WriteHeader(thdr)
		if nil != e {
			t.Fatalf("Unable to write tar header: %v", e)
		}

		_, e = twtr.Write(tbdy)
		if nil != e {
			t.Fatalf("Unable to write tar body: %v", e)
		}

		e = twtr.Close()
		if nil != e {
			t.Fatalf("Unable to finalize tar: %v", e)
		}

		var brdr *bytes.Reader = bytes.NewReader(vtar.Bytes())
		var t2f f2k.GetFiles = Tar2Files(brdr)
		var rfiles f2k.Result[s2k.Iter[f2k.FileEx]] = t2f(context.Background())
		checker(t, rfiles.IsOk(), true)

		var ifiles s2k.Iter[f2k.FileEx] = rfiles.Value()
		var of s2k.Option[f2k.FileEx] = ifiles()
		checker(t, of.HasValue(), true)

		var f f2k.FileEx = of.Value()

		checker(t, f.Name(), "cafef00ddeadbeafface864299792458/bucket.txt")

		var raw fs.File = f.File()
		s, e := raw.Stat()
		if nil != e {
			t.Fatalf("Unable to get virtual file stat: %v", e)
		}

		checker(t, s.Size(), int64(len(tbdy)))

		got, e := io.ReadAll(raw)
		if nil != e {
			t.Fatalf("Unable to read virtual file: %v", e)
		}

		checkBytes(t, got, tbdy)
	})
}
