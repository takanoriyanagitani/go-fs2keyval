package fs2kv

import (
	"io"
	"io/fs"
	"regexp"
	"strings"
	"testing"

	s2k "github.com/takanoriyanagitani/go-sql2keyval"
)

func TestBatch2File(t *testing.T) {
	t.Parallel()

	type patChecker func(s string) bool

	patCheckerBuilderNewMust := func(pat string) patChecker {
		var r *regexp.Regexp = regexp.MustCompile(pat)
		return r.MatchString
	}

	var bucketChecker patChecker = patCheckerBuilderNewMust("^[0-9a-f-]{36}/bucket$")
	var keyChecker patChecker = patCheckerBuilderNewMust("^[0-9a-f-]{36}/key$")
	var valChecker patChecker = patCheckerBuilderNewMust("^[0-9a-f-]{36}/val$")

	t.Run("BatchIter2FilesUuid", func(t *testing.T) {
		t.Parallel()

		t.Run("empty", func(t *testing.T) {
			t.Parallel()

			var ib s2k.Iter[s2k.Batch] = s2k.IterEmptyNew[s2k.Batch]()
			var ir s2k.Iter[Result[FileEx]] = BatchIter2FilesUuid(ib)
			var or s2k.Option[Result[FileEx]] = ir()
			checker(t, or.HasValue(), false)
		})

		t.Run("multi", func(t *testing.T) {
			t.Parallel()

			var testdata []s2k.Batch = []s2k.Batch{
				s2k.BatchNew("data_2022_09_07_cafef00ddeadbeafface864299792458", []byte("k"), []byte("v")),
				s2k.BatchNew("data_2022_09_07_facecafef00ddeadbeaf864299792458", []byte("l"), []byte("m")),
			}
			var ib s2k.Iter[s2k.Batch] = s2k.IterFromArray(testdata)
			var ir s2k.Iter[Result[FileEx]] = BatchIter2FilesUuid(ib)

			var results []Result[FileEx] = ir.ToArray()
			checker(t, len(results), 6)

			for i := range testdata {
				var b s2k.Batch = testdata[i]

				var rbucket Result[FileEx] = results[0+3*i]
				var rkey Result[FileEx] = results[1+3*i]
				var rval Result[FileEx] = results[2+3*i]

				checker(t, rbucket.IsOk(), true)
				checker(t, rkey.IsOk(), true)
				checker(t, rval.IsOk(), true)

				var fbucket FileEx = rbucket.Value()
				var fkey FileEx = rkey.Value()
				var fval FileEx = rval.Value()

				checker(t, bucketChecker(fbucket.Name()), true)
				checker(t, keyChecker(fkey.Name()), true)
				checker(t, valChecker(fval.Name()), true)

				var sbucket []string = strings.SplitN(fbucket.Name(), "/", 2)
				var skey []string = strings.SplitN(fkey.Name(), "/", 2)
				var sval []string = strings.SplitN(fval.Name(), "/", 2)

				checker(t, len(sbucket), 2)
				checker(t, len(skey), 2)
				checker(t, len(sval), 2)

				var cprefix string = sbucket[0]

				checker(t, cprefix, sbucket[0])
				checker(t, cprefix, skey[0])
				checker(t, cprefix, sval[0])

				var rfbucket fs.File = fbucket.File()
				var rfkey fs.File = fkey.File()
				var rfval fs.File = fval.File()

				bb, eb := io.ReadAll(rfbucket)
				bk, ek := io.ReadAll(rfkey)
				bv, ev := io.ReadAll(rfval)

				checkErr(eb, func(e error) { t.Errorf("Unable to read bucket: %v", e) })
				checkErr(ek, func(e error) { t.Errorf("Unable to read key: %v", e) })
				checkErr(ev, func(e error) { t.Errorf("Unable to read val: %v", e) })

				checker(t, 0 < len(bb), true)
				checker(t, 0 < len(bk), true)
				checker(t, 0 < len(bv), true)

				var p s2k.Pair = b.Pair()

				checkBytes(t, bb, []byte(b.Bucket()))
				checkBytes(t, bk, p.Key)
				checkBytes(t, bv, p.Val)
			}

		})
	})
}
