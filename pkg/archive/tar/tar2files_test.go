package files2tar

import (
	"bytes"
	"context"
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
}
