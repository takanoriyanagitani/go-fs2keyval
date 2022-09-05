package batch2zip

import (
	"archive/zip"
	"context"
	"io"

	s2k "github.com/takanoriyanagitani/go-sql2keyval"

	f2k "github.com/takanoriyanagitani/go-fs2keyval"
)

type zipWriter struct {
	wtr io.Writer
	err error
}

func (z *zipWriter) Write(b []byte) (int, error) { return z.wtr.Write(b) }

type path2writerSet struct {
	p2h  path2header
	h2wb header2writerBuilder
}

type path2header func(p string) *zip.FileHeader
type header2writer func(h *zip.FileHeader) zipWriter
type header2writerBuilder func(w *zip.Writer) header2writer

func path2headerBuilder(method uint16) path2header {
	return func(p string) *zip.FileHeader {
		return &zip.FileHeader{
			Name:   p,
			Method: method,
		}
	}
}

var header2writerBuilderStoreNew header2writerBuilder = func(w *zip.Writer) header2writer {
	return func(h *zip.FileHeader) (zw zipWriter) {
		zw.wtr, zw.err = w.CreateHeader(h)
		return
	}
}

var header2writerBuilderRawNew header2writerBuilder = func(w *zip.Writer) header2writer {
	return func(h *zip.FileHeader) (zw zipWriter) {
		zw.wtr, zw.err = w.CreateRaw(h)
		return
	}
}

var path2headerStore path2header = path2headerBuilder(zip.Store)
var path2headerRaw path2header = func(p string) *zip.FileHeader { return &zip.FileHeader{Name: p} }

var path2writerSetStore = path2writerSet{
	p2h:  path2headerStore,
	h2wb: header2writerBuilderStoreNew,
}

var path2writerSetRaw = path2writerSet{
	p2h:  path2headerRaw,
	h2wb: header2writerBuilderRawNew,
}

type path2writerBuilder func(z *zip.Writer) func(p string) zipWriter

func path2writerBuilderNew(p2w path2writerSet) path2writerBuilder {
	return func(z *zip.Writer) func(p string) zipWriter {
		return s2k.Compose(p2w.p2h, p2w.h2wb(z))
	}
}

var path2writerBuilderStore path2writerBuilder = path2writerBuilderNew(path2writerSetStore)
var path2writerBuilderRaw path2writerBuilder = path2writerBuilderNew(path2writerSetRaw)

type filelike2zipBuilder func(z *zip.Writer) func(ctx context.Context, f f2k.FileLike) error

func filelike2zipBuilderNew(p2wb path2writerBuilder) filelike2zipBuilder {
	return func(z *zip.Writer) func(ctx context.Context, f f2k.FileLike) error {
		var p2w func(p string) zipWriter = p2wb(z)
		return func(ctx context.Context, f f2k.FileLike) error {
			var zw zipWriter = p2w(f.Path)
			_, e := zw.Write(f.Val)
			return e
		}
	}
}

var filelike2zipBuilderStore filelike2zipBuilder = filelike2zipBuilderNew(path2writerBuilderStore)
var filelike2zipBuilderRaw filelike2zipBuilder = filelike2zipBuilderNew(path2writerBuilderRaw)

type filelikeIter2zipBuilder func(z *zip.Writer) f2k.SetFilelikeBatch

func filelikeIter2zipBuilderNew(f2zb filelike2zipBuilder) filelikeIter2zipBuilder {
	return func(z *zip.Writer) f2k.SetFilelikeBatch {
		var f2z func(ctx context.Context, f f2k.FileLike) error = f2zb(z)
		return func(ctx context.Context, many s2k.Iter[f2k.FileLike]) error {
			return s2k.IterReduce(many, nil, func(e error, f f2k.FileLike) error {
				if nil != e {
					return e
				}
				return f2z(ctx, f)
			})
		}
	}
}

var filelikeIter2zipBuilderStore filelikeIter2zipBuilder = filelikeIter2zipBuilderNew(filelike2zipBuilderStore)
var filelikeIter2zipBuilderRaw filelikeIter2zipBuilder = filelikeIter2zipBuilderNew(filelike2zipBuilderRaw)

func filelikeIter2FsBuilder(b filelikeIter2zipBuilder) func(file io.Writer) f2k.SetFilelikeBatch {
	return func(file io.Writer) f2k.SetFilelikeBatch {
		return func(ctx context.Context, many s2k.Iter[f2k.FileLike]) error {
			var zw *zip.Writer = zip.NewWriter(file)
			var sfb f2k.SetFilelikeBatch = b(zw)
			e := sfb(ctx, many)
			if nil == e {
				return zw.Close()
			}
			defer zw.Close()
			return e
		}
	}
}

var FilelikeIter2FsStored func(file io.Writer) f2k.SetFilelikeBatch = filelikeIter2FsBuilder(filelikeIter2zipBuilderStore)
var FilelikeIter2FsRaw func(file io.Writer) f2k.SetFilelikeBatch = filelikeIter2FsBuilder(filelikeIter2zipBuilderRaw)
