package fs2kv

import (
	"io/fs"
	"path"

	"github.com/google/uuid"

	s2k "github.com/takanoriyanagitani/go-sql2keyval"
)

type Batch2Files func(b s2k.Batch) Result[s2k.Iter[FileEx]]
type BatchIter2Files func(s2k.Iter[s2k.Batch]) s2k.Iter[Result[FileEx]]

type UniqueDirnameBuilder func() Result[string]

func uuid2string(u uuid.UUID) string { return u.String() }

var uuidGen4 func() Result[uuid.UUID] = ResultBuilderNew0(uuid.NewRandom)

var uuidDirnameBuilder UniqueDirnameBuilder = func() Result[string] {
	var ru Result[uuid.UUID] = uuidGen4()
	return ResultMap(ru, uuid2string)
}

func bytes2file(name string, val []byte) fs.File { return MemFileNew(name, val, 0644) }

func prefix2fileExBuilderNew(prefix string) func(name string) func(val []byte) FileEx {
	return func(name string) func([]byte) FileEx {
		return func(val []byte) FileEx {
			var f fs.File = bytes2file(name, val)
			return FileExNew(f, path.Join(prefix, name))
		}
	}
}

func batch2filesBuilderNew(dbldr UniqueDirnameBuilder) Batch2Files {
	return func(b s2k.Batch) Result[s2k.Iter[FileEx]] {
		var rp Result[string] = dbldr()
		return ResultMap(rp, func(prefix string) s2k.Iter[FileEx] {
			var nv2fe func(name string) func(val []byte) FileEx = prefix2fileExBuilderNew(prefix)
			var kv s2k.Pair = b.Pair()
			return s2k.IterFromArray([]FileEx{
				nv2fe("bucket")([]byte(b.Bucket())),
				nv2fe("key")(kv.Key),
				nv2fe("val")(kv.Val),
			})
		})
	}
}

var uuidBatch2Files Batch2Files = batch2filesBuilderNew(uuidDirnameBuilder)

func batchIter2filesBuilderNew(b2f Batch2Files) func(s2k.Iter[s2k.Batch]) s2k.Iter[Result[FileEx]] {
	return func(i s2k.Iter[s2k.Batch]) s2k.Iter[Result[FileEx]] {
		var ri s2k.Iter[Result[s2k.Iter[FileEx]]] = s2k.IterMap(i, b2f)
		var ir s2k.Iter[s2k.Iter[Result[FileEx]]] = s2k.IterMap(ri, ResultIter2iterResults[FileEx])
		return ResultsFlatten(ir)
	}
}

var BatchIter2FilesUuid BatchIter2Files = batchIter2filesBuilderNew(uuidBatch2Files)
