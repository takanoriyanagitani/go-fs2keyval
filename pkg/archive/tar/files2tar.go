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

func file2hdr(f f2k.FileEx) f2k.Result[*tar.Header] {
	var rfi f2k.Result[fs.FileInfo] = f2k.File2Info(f.File())
	var rh f2k.Result[*tar.Header] = f2k.ResultFlatMap(rfi, info2header)
	return rh.Map(func(th *tar.Header) *tar.Header {
		th.Name = f.Name()
		return th
	})
}

func file2tarExRes(_ctx context.Context, fe f2k.Result[f2k.FileEx], tw *tar.Writer) error {
	var th f2k.Result[*tar.Header] = f2k.ResultFlatMap(fe, file2hdr)
	return th.TryForEach(func(h *tar.Header) error {
		return f2k.Error1st([]func() error{
			func() error { return tw.WriteHeader(h) },
			func() error {
				var re f2k.Result[error] = f2k.ResultMap(fe, func(f f2k.FileEx) error {
					_, e := io.Copy(tw, f.File())
					return e
				})
				return re.UnwrapOrElse(func(e error) error {
					return e
				})
			},
		})
	})
}

func files2tarExRes(ctx context.Context, files s2k.Iter[f2k.Result[f2k.FileEx]], tw *tar.Writer) error {
	return s2k.IterReduce(files, nil, func(e error, f f2k.Result[f2k.FileEx]) error {
		return f2k.IfOk(e, func() error {
			return file2tarExRes(ctx, f, tw)
		})
	})
}

func files2tarWriterExRes(ctx context.Context, files s2k.Iter[f2k.Result[f2k.FileEx]], w io.Writer) error {
	var tw *tar.Writer = tar.NewWriter(w)
	return f2k.ErrorWarn(
		func() error { return files2tarExRes(ctx, files, tw) },
		func() error { return tw.Close() },
	)
}

func Files2TarBuilderNew(w io.Writer) f2k.SetFsFileBatch {
	return func(ctx context.Context, many s2k.Iter[fs.File]) error {
		var mapd s2k.Iter[f2k.Result[f2k.FileEx]] = s2k.IterMap(many, f2k.FileExFromStd)
		return files2tarWriterExRes(ctx, mapd, w)
	}
}

func Files2TarBuilderExNew(w io.Writer) f2k.SetFilesBatch {
	return func(ctx context.Context, many s2k.Iter[f2k.FileEx]) error {
		var mapd s2k.Iter[f2k.Result[f2k.FileEx]] = f2k.ResultWrapIter(many)
		return files2tarWriterExRes(ctx, mapd, w)
	}
}
