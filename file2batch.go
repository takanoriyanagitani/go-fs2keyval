package fs2kv

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"unicode/utf8"

	s2k "github.com/takanoriyanagitani/go-sql2keyval"
)

type Files2BatchIter func(s2k.Iter[FileEx]) s2k.Iter[Result[s2k.Batch]]

type file2bytes func(f fs.File) Result[[]byte]

var file2bytesBufio file2bytes = func(f fs.File) Result[[]byte] {
	var rdr *bufio.Reader = bufio.NewReader(f)
	var buf bytes.Buffer
	_, e := io.Copy(&buf, rdr)
	return ResultFromBool(buf.Bytes, nil == e, func() error { return e })
}

type bytes2bucket func([]byte) Result[string]
type bucketChecker func([]byte) bool

var bucketCheckerUtf8 bucketChecker = utf8.Valid

type bytes2string func([]byte) string

var bytes2stringDefault bytes2string = func(b []byte) string {
	var sb strings.Builder
	sb.Write(b) // error = nil(always)
	return sb.String()
}

func bytes2bucketBuilderNew(bchk bucketChecker) func(bytes2string) bytes2bucket {
	return func(b2s bytes2string) bytes2bucket {
		return func(b []byte) Result[string] {
			return ResultFromBool(
				func() string { return b2s(b) },
				bchk(b),
				func() error { return fmt.Errorf("Invalid bucket name") },
			)
		}
	}
}

var bytes2bucketDefault bytes2bucket = bytes2bucketBuilderNew(bucketCheckerUtf8)(bytes2stringDefault)

type files2batch func(bkt, key, val fs.File) Result[s2k.Batch]

func files2batchBuilderNew(f2b file2bytes) func(b2b bytes2bucket) files2batch {
	return func(b2b bytes2bucket) files2batch {
		return func(bkt, key, val fs.File) Result[s2k.Batch] {
			var rb Result[[]byte] = f2b(bkt)
			var rk Result[[]byte] = f2b(key)
			var rv Result[[]byte] = f2b(val)
			results := []Result[[]byte]{
				rb, rk, rv,
			}
			var ir s2k.Iter[Result[[]byte]] = s2k.IterFromArray(results)
			var rs Result[[][]byte] = ResultTryUnwrapAll(ir)
			var ra Result[[3][]byte] = ResultFlatMap(rs, func(s [][]byte) Result[[3][]byte] {
				var o s2k.Option[[][]byte] = s2k.OptionNew(s).
					Filter(func(sb [][]byte) bool { return 3 == len(sb) })
				var oa s2k.Option[[3][]byte] = s2k.OptionMap(o, func(sb [][]byte) [3][]byte {
					var abs [3][]byte
					abs[0] = sb[0]
					abs[1] = sb[1]
					abs[2] = sb[2]
					return abs
				})
				return ResultFromBool(oa.Value, oa.HasValue(), func() error {
					return fmt.Errorf("Invalid file")
				})
			})
			return ResultFlatMap(ra, func(a [3][]byte) Result[s2k.Batch] {
				var rb Result[string] = b2b(a[0])
				return ResultMap(rb, func(bucket string) s2k.Batch {
					return s2k.BatchNew(bucket, a[1], a[2])
				})
			})
		}
	}
}

var files2batchDefault files2batch = files2batchBuilderNew(file2bytesBufio)(bytes2bucketDefault)

type fileIter2batch func(files s2k.Iter[FileEx]) Result[s2k.Batch]

type filesEx2batch func(bkt, key, val FileEx) Result[s2k.Batch]

type namesChecker func(bkt, key, val string) bool

var nameCheckerDefault namesChecker = func(bkt, key, val string) bool {
	var sb []string = strings.SplitN(bkt, "/", 2)
	var sk []string = strings.SplitN(key, "/", 2)
	var sv []string = strings.SplitN(val, "/", 2)
	i := s2k.IterFromArray([]func() bool{
		func() bool { return 2 == len(sb) },
		func() bool { return 2 == len(sk) },
		func() bool { return 2 == len(sv) },
		func() bool { return sb[0] == sk[0] },
		func() bool { return sb[0] == sv[0] },
		func() bool { return "bucket" == sb[1] },
		func() bool { return "key" == sk[1] },
		func() bool { return "val" == sv[1] },
	})
	return s2k.IterReduce(i, true, func(b bool, f func() bool) bool {
		return b && f()
	})
}

func filesEx2batchBuilderNew(nchk namesChecker) func(files2batch) filesEx2batch {
	return func(f2b files2batch) filesEx2batch {
		return func(bkt, key, val FileEx) Result[s2k.Batch] {
			var rr Result[Result[s2k.Batch]] = ResultFromBool(
				func() Result[s2k.Batch] { return f2b(bkt.File(), key.File(), val.File()) },
				nchk(bkt.Name(), key.Name(), val.Name()),
				func() error { return fmt.Errorf("Invalid files") },
			)
			return ResultFlatMap(rr, Identity[Result[s2k.Batch]])
		}
	}
}

var filesEx2batchDefault filesEx2batch = filesEx2batchBuilderNew(nameCheckerDefault)(files2batchDefault)

func fileIter2batchBuilderNew(fe2b filesEx2batch) fileIter2batch {
	return func(files s2k.Iter[FileEx]) Result[s2k.Batch] {
		var obkt s2k.Option[FileEx] = files()
		var okey s2k.Option[FileEx] = files()
		var oval s2k.Option[FileEx] = files()
		i := s2k.IterFromArray([]func() bool{
			obkt.HasValue,
			okey.HasValue,
			oval.HasValue,
		})
		var ok bool = s2k.IterReduce(i, true, func(b bool, f func() bool) bool {
			return b && f()
		})
		var ra Result[[3]FileEx] = ResultFromBool(
			func() [3]FileEx {
				var fa [3]FileEx
				fa[0] = obkt.Value()
				fa[1] = okey.Value()
				fa[2] = oval.Value()
				return fa
			},
			ok,
			func() error { return io.EOF },
		)
		return ResultFlatMap(ra, func(fa [3]FileEx) Result[s2k.Batch] {
			return fe2b(fa[0], fa[1], fa[2])
		})
	}
}

var fileIter2batchDefault fileIter2batch = fileIter2batchBuilderNew(filesEx2batchDefault)

type FileIter2batchIter func(files s2k.Iter[FileEx]) s2k.Iter[Result[s2k.Batch]]

func fileIter2batchIterBuilderNew(fi2b fileIter2batch) FileIter2batchIter {
	return func(files s2k.Iter[FileEx]) s2k.Iter[Result[s2k.Batch]] {
		return func() s2k.Option[Result[s2k.Batch]] {
			var rb Result[s2k.Batch] = fi2b(files)
			return ResultFilter(
				rb,
				func(e error) bool { return io.EOF == e },
			)
		}
	}
}

var FileIter2batchIterDefault FileIter2batchIter = fileIter2batchIterBuilderNew(fileIter2batchDefault)
