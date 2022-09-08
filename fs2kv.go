package fs2kv

import (
	"context"
	"io/fs"
	"unicode/utf8"

	s2k "github.com/takanoriyanagitani/go-sql2keyval"
)

type KeyValidator func(key []byte) bool

type PairValidator func(p s2k.Pair) bool

type BatchValidator func(b s2k.Batch) bool

type FileEx interface {
	File() fs.File // base name only
	Name() string  // full path
}

type fileEx struct {
	raw fs.File
	nam string
}

func (f fileEx) File() fs.File { return f.raw }
func (f fileEx) Name() string  { return f.nam }

func FileExNew(raw fs.File, nam string) FileEx {
	return fileEx{
		raw,
		nam,
	}
}

var fileExNew func(raw fs.File) func(nam string) FileEx = Curry(FileExNew)

func FileExFromStd(raw fs.File) Result[FileEx] {
	var rfi Result[fs.FileInfo] = File2Info(raw)
	var rnm Result[string] = ResultMap(rfi, func(fi fs.FileInfo) string { return fi.Name() })
	var nm2ex func(nam string) FileEx = fileExNew(raw)
	return ResultMap(rnm, nm2ex)
}

type SetFsFileBatch func(ctx context.Context, many s2k.Iter[fs.File]) error
type SetFilesBatch func(ctx context.Context, many s2k.Iter[FileEx]) error
type SetFiles func(ctx context.Context, files s2k.Iter[Result[FileEx]]) error

type BatchIter2Fs func(ctx context.Context, many s2k.Iter[s2k.Batch]) error

func batchIter2fsBuilderNew(bi2f BatchIter2Files) func(SetFiles) BatchIter2Fs {
	return func(f2b SetFiles) BatchIter2Fs {
		return func(ctx context.Context, many s2k.Iter[s2k.Batch]) error {
			var files s2k.Iter[Result[FileEx]] = bi2f(many)
			return f2b(ctx, files)
		}
	}
}

var BatchIter2fsBuilderUuid func(SetFiles) BatchIter2Fs = batchIter2fsBuilderNew(BatchIter2FilesUuid)

var Utf8validator KeyValidator = utf8.Valid

type Bytes2string func(b []byte) s2k.Option[string]

func ErrorWarn(funcError func() error, funcWarn func() error) error {
	e := funcError()
	if nil == e {
		return funcWarn()
	}
	defer funcWarn()
	return e
}

func Error1st(f []func() error) error {
	return s2k.IterReduce(s2k.IterFromArray(f), nil, func(e error, item func() error) error {
		if nil == e {
			return item()
		}
		return e
	})
}

func IfOk(e error, f func() error) error {
	return Error1st([]func() error{
		func() error { return e },
		f,
	})
}

func file2info(f fs.File) (fs.FileInfo, error) { return f.Stat() }

var File2Info func(fs.File) Result[fs.FileInfo] = ResultBuilderNew1(file2info)
