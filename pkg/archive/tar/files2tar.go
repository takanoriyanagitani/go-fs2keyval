package files2tar

import (
	"archive/tar"
	"context"
	"io"
	"io/fs"

	s2k "github.com/takanoriyanagitani/go-sql2keyval"

	f2k "github.com/takanoriyanagitani/go-fs2keyval"
)

func inf2hdr(i fs.FileInfo) (*tar.Header, error) { return tar.FileInfoHeader(i, "") }

var info2header func(fs.FileInfo) f2k.Result[*tar.Header] = f2k.ResultBuilderNew1(inf2hdr)

func file2tar(_ctx context.Context, f fs.File, tw *tar.Writer) error {
	var fi f2k.Result[fs.FileInfo] = f2k.File2Info(f)
	var th f2k.Result[*tar.Header] = f2k.ResultFlatMap(fi, info2header)
	return th.TryForEach(func(h *tar.Header) error {
		return f2k.Error1st([]func() error{
			func() error { return tw.WriteHeader(h) },
			func() error {
				_, e := io.Copy(tw, f)
				return e
			},
		})
	})
}

func files2tar(ctx context.Context, files s2k.Iter[fs.File], tw *tar.Writer) error {
	return s2k.IterReduce(files, nil, func(e error, f fs.File) error {
		return f2k.IfOk(e, func() error {
			return file2tar(ctx, f, tw)
		})
	})
}

func files2tarWriter(ctx context.Context, files s2k.Iter[fs.File], w io.Writer) error {
	var tw *tar.Writer = tar.NewWriter(w)
	return f2k.ErrorWarn(
		func() error { return files2tar(ctx, files, tw) },
		func() error { return tw.Close() },
	)
}

//type SetFsFileBatch func(ctx context.Context, many s2k.Iter[fs.File]) error

func Files2TarBuilderNew(w io.Writer) f2k.SetFsFileBatch {
	return func(ctx context.Context, many s2k.Iter[fs.File]) error {
		return files2tarWriter(ctx, many, w)
	}
}
