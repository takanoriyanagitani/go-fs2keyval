package fs2kv

import (
	"io"
	"os"
	"path/filepath"

	s2k "github.com/takanoriyanagitani/go-sql2keyval"
)

// A DatabaseNew prepare batch setter for io.Writer.
// For example, tar driver will create tar file.
// Single database = single archive file.
// Single archive file may contain many batches.
type DatabaseNew func(io.Writer) SetFiles

// A DatabaseGetBatchIter prepare batch iter getter for io.Reader.
// For example, tar driver will get batch files from tar file.
type DatabaseGetBatchIter func(io.Reader) GetFiles

type DatabaseListFsEnv func(limit int) func(rootdir string) Result[[]string]
type DatabaseListFs func(instanceName string) DatabaseListFsEnv

type DatabaseDropFsEnv func(rootdir string) func(dbname string) error
type DatabaseDropFs func(instanceName string) DatabaseDropFsEnv

func databaseListFsEnvBuilderNew(dlf DatabaseListFs) DatabaseListFsEnv {
	return func(limit int) func(rootdir string) Result[[]string] {
		return func(rootdir string) Result[[]string] {
			var iname Result[string] = InstanceGetterEnvDefault()
			return ResultFlatMap(iname, func(s string) Result[[]string] {
				var dlfed DatabaseListFsEnv = dlf(s)
				return dlfed(limit)(rootdir)
			})
		}
	}
}

func file2dirs(limit int) func(f *os.File) Result[[]os.DirEntry] {
	return func(f *os.File) Result[[]os.DirEntry] {
		return ResultNew(f.ReadDir(limit)).OrElse(func(e error) Result[[]os.DirEntry] {
			return ResultFromBool(
				func() []os.DirEntry { return nil },
				io.EOF == e,
				func() error { return e },
			)
		})
	}
}

func fullpath2dirs(limit int) func(fullpath string) Result[[]os.DirEntry] {
	var file2items func(f *os.File) Result[[]os.DirEntry] = file2dirs(limit)
	return func(fullpath string) Result[[]os.DirEntry] {
		var fdir Result[*os.File] = ResultNew(os.Open(filepath.Clean(fullpath)))
		var items Result[[]os.DirEntry] = ResultFlatMap(fdir, file2items)
		fdir.Ok().ForEach(func(f *os.File) {
			_ = f.Close() // ignore error on close
		})
		return items
	}
}

func dirent2string(d os.DirEntry) string { return d.Name() }

func items2strings(i s2k.Iter[os.DirEntry]) s2k.Iter[string] {
	return s2k.IterMap(i, dirent2string)
}

var databaseListFsDefault DatabaseListFs = func(instanceName string) DatabaseListFsEnv {
	return func(limit int) func(rootdir string) Result[[]string] {
		var path2dirs func(p string) Result[[]os.DirEntry] = fullpath2dirs(limit)
		return func(rootdir string) Result[[]string] {
			var dirname string = filepath.Join(rootdir, instanceName)
			var items Result[[]os.DirEntry] = path2dirs(dirname)
			var ri Result[s2k.Iter[os.DirEntry]] = ResultMap(items, s2k.IterFromArray[os.DirEntry])
			var names Result[s2k.Iter[string]] = ResultMap(ri, items2strings)
			return ResultMap(names, func(i s2k.Iter[string]) []string { return i.ToArray() })
		}
	}
}

var databaseDropFsDefault DatabaseDropFs = func(instance string) DatabaseDropFsEnv {
	return func(rootdir string) func(dbname string) error {
		return func(dbname string) error {
			var parent string = filepath.Join(rootdir, instance)
			var full string = filepath.Join(parent, dbname)
			return os.RemoveAll(full)
		}
	}
}

func databaseDropFsEnvBuilderNew(d2f DatabaseDropFs) DatabaseDropFsEnv {
	return func(rootdir string) func(dbname string) error {
		return func(dbname string) error {
			var iname Result[string] = InstanceGetterEnvDefault()
			var rdrop Result[DatabaseDropFsEnv] = ResultMap(iname, d2f)
			var re Result[error] = ResultMap(rdrop, func(d DatabaseDropFsEnv) error {
				return d(rootdir)(dbname)
			})
			return re.UnwrapOrElse(Identity[error])
		}
	}
}

var DatabaseListFsEnvDefault DatabaseListFsEnv = databaseListFsEnvBuilderNew(
	databaseListFsDefault,
)

var DatabaseDropFsEnvDefault DatabaseDropFsEnv = databaseDropFsEnvBuilderNew(
	databaseDropFsDefault,
)
