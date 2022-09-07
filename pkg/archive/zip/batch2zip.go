package batch2zip

import (
	"context"
	"io"
	"io/fs"
	"os"

	tpzip "github.com/mholt/archiver/v3"

	s2k "github.com/takanoriyanagitani/go-sql2keyval"

	f2k "github.com/takanoriyanagitani/go-fs2keyval"
)

func file2zipBuilderNew(z *tpzip.Zip) func(context.Context, tpzip.File) error {
	return func(_ctx context.Context, f tpzip.File) error {
		return z.Write(f)
	}
}

func fileNew(FileInfo os.FileInfo, Header interface{}, ReadCloser io.ReadCloser) tpzip.File {
	return tpzip.File{
		FileInfo,
		Header,
		ReadCloser,
	}
}

func stdFile2file(sf fs.File) f2k.Result[tpzip.File] {
	var ri f2k.Result[fs.FileInfo] = f2k.File2Info(sf)
	return f2k.ResultMap(ri, func(fi fs.FileInfo) tpzip.File {
		return fileNew(fi, nil, sf)
	})
}

func stdFiles2zipBuilderNew(z *tpzip.Zip) f2k.SetFsFileBatch {
	var tf2z func(context.Context, tpzip.File) error = file2zipBuilderNew(z)
	var tc func(context.Context) func(tpzip.File) error = f2k.Curry(tf2z)
	return func(ctx context.Context, files s2k.Iter[fs.File]) error {
		return s2k.IterReduce(files, nil, func(e error, f fs.File) error {
			return f2k.IfOk(e, func() error {
				var rf f2k.Result[tpzip.File] = stdFile2file(f)
				var f2z func(tpzip.File) error = tc(ctx)
				var re f2k.Result[error] = f2k.ResultMap(rf, f2z)
				return re.UnwrapOrElse(func(e error) error {
					return e
				})
			})
		})
	}
}

type ZipConfig func(z *tpzip.Zip) *tpzip.Zip

func zipConfigBuilderNew(method tpzip.ZipCompressionMethod) ZipConfig {
	return func(z *tpzip.Zip) *tpzip.Zip {
		z.FileMethod = method
		return z
	}
}

var zipConfigDefault ZipConfig = f2k.Identity[*tpzip.Zip]
var zipConfigStore ZipConfig = zipConfigBuilderNew(tpzip.Store)

type zipWriterBuilder func(w io.Writer) f2k.Result[*tpzip.Zip]

func zipWriterBuilderNew(cfg ZipConfig) func(w io.Writer) f2k.Result[*tpzip.Zip] {
	return func(w io.Writer) f2k.Result[*tpzip.Zip] {
		var z *tpzip.Zip = cfg(tpzip.NewZip())
		e := z.Create(w)
		return f2k.ResultNew(z, e)
	}
}

var zipWriterBuilderDefault zipWriterBuilder = zipWriterBuilderNew(zipConfigDefault)
var zipWriterBuilderStore zipWriterBuilder = zipWriterBuilderNew(zipConfigStore)

func files2zipBuilderFactoryNew(b zipWriterBuilder) func(w io.Writer) f2k.SetFsFileBatch {
	return func(w io.Writer) f2k.SetFsFileBatch {
		return func(ctx context.Context, files s2k.Iter[fs.File]) error {
			var rz f2k.Result[*tpzip.Zip] = b(w)
			return rz.TryForEach(func(z *tpzip.Zip) error {
				defer z.Close()
				var sfb f2k.SetFsFileBatch = stdFiles2zipBuilderNew(z)
				return sfb(ctx, files)
			})
		}
	}
}

var Files2ZipBuilderDefault func(w io.Writer) f2k.SetFsFileBatch = files2zipBuilderFactoryNew(zipWriterBuilderDefault)
var Files2ZipBuilderStore func(w io.Writer) f2k.SetFsFileBatch = files2zipBuilderFactoryNew(zipWriterBuilderStore)
