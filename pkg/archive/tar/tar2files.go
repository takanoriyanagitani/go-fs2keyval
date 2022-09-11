package files2tar

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"io/fs"

	s2k "github.com/takanoriyanagitani/go-sql2keyval"

	f2k "github.com/takanoriyanagitani/go-fs2keyval"
)

func tarItem2file(rdr io.Reader) func(hdr *tar.Header) f2k.Result[f2k.FileEx] {
	return func(hdr *tar.Header) f2k.Result[f2k.FileEx] {
		var fi fs.FileInfo = hdr.FileInfo()
		var buf bytes.Buffer
		_, e := io.Copy(&buf, rdr)
		return f2k.ResultFromBool(
			func() f2k.FileEx {
				return f2k.FileExNew(
					f2k.MemFileNew(
						fi.Name(),
						buf.Bytes(),
						0644,
					),
					hdr.Name,
				)
			},
			nil == e,
			func() error { return e },
		)
	}
}

func reader2fileIter(ctx context.Context, r *tar.Reader) s2k.Iter[f2k.Result[f2k.FileEx]] {
	var header2file func(hdr *tar.Header) f2k.Result[f2k.FileEx] = tarItem2file(r)
	return func() s2k.Option[f2k.Result[f2k.FileEx]] {
		var rh f2k.Result[*tar.Header] = f2k.ResultNew(r.Next())
		var or s2k.Option[f2k.Result[*tar.Header]] = f2k.ResultFilter(
			rh,
			func(e error) bool { return io.EOF == e },
		)
		return s2k.OptionMap(or, func(r f2k.Result[*tar.Header]) f2k.Result[f2k.FileEx] {
			return f2k.ResultFlatMap(r, header2file)
		})
	}
}

func reader2files(ctx context.Context, r *tar.Reader) f2k.Result[s2k.Iter[f2k.FileEx]] {
	var ifiles s2k.Iter[f2k.Result[f2k.FileEx]] = reader2fileIter(ctx, r)
	var rfiles f2k.Result[[]f2k.FileEx] = f2k.ResultTryUnwrapAll(ifiles)
	return f2k.ResultMap(rfiles, s2k.IterFromArray[f2k.FileEx])
}

var Tar2Files f2k.DatabaseGetBatchIter = func(rdr io.Reader) f2k.GetFiles {
	return func(ctx context.Context) f2k.Result[s2k.Iter[f2k.FileEx]] {
		return reader2files(ctx, tar.NewReader(rdr))
	}
}
