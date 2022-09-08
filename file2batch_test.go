package fs2kv

import (
	"testing"

	s2k "github.com/takanoriyanagitani/go-sql2keyval"
)

func TestFile2Batch(t *testing.T) {
	t.Parallel()

	t.Run("FileIter2batchIterDefault", func(t *testing.T) {
		t.Parallel()

		t.Run("empty", func(t *testing.T) {
			t.Parallel()

			var files s2k.Iter[FileEx] = s2k.IterEmptyNew[FileEx]()
			var results s2k.Iter[Result[s2k.Batch]] = FileIter2batchIterDefault(files)

			var or s2k.Option[Result[s2k.Batch]] = results()
			checker(t, or.HasValue(), false)
		})

		t.Run("invalid files(bucket only)", func(t *testing.T) {
			t.Parallel()

			var files s2k.Iter[FileEx] = s2k.IterFromArray([]FileEx{
				FileExNew(MemFileNew("bucket", []byte("data_2022_09_07_cafef00ddeadbeafface864299792458"), 0644), "f00ddeadbeaffacecafe864299792458/bucket"),
			})
			var results s2k.Iter[Result[s2k.Batch]] = FileIter2batchIterDefault(files)
			var or s2k.Option[Result[s2k.Batch]] = results()
			checker(t, or.HasValue(), false)
		})

		t.Run("invalid files(val missing)", func(t *testing.T) {
			t.Parallel()

			var files s2k.Iter[FileEx] = s2k.IterFromArray([]FileEx{
				FileExNew(MemFileNew("bucket", []byte("data_2022_09_07_cafef00ddeadbeafface864299792458"), 0644), "f00ddeadbeaffacecafe864299792458/bucket"),
				FileExNew(MemFileNew("key", []byte("16:57:51.0Z"), 0644), "f00ddeadbeaffacecafe864299792458/key"),
			})
			var results s2k.Iter[Result[s2k.Batch]] = FileIter2batchIterDefault(files)
			var or s2k.Option[Result[s2k.Batch]] = results()
			checker(t, or.HasValue(), false)
		})

		t.Run("invalid files(dirname unmatch)", func(t *testing.T) {
			t.Parallel()

			dname1 := "f00ddeadbeaffacecafe864299792458"
			dname2 := "deadbeaffacecafef00d864299792458"

			var files s2k.Iter[FileEx] = s2k.IterFromArray([]FileEx{
				FileExNew(MemFileNew("bucket", []byte("data_2022_09_07_cafef00ddeadbeafface864299792458"), 0644), dname1+"/bucket"),
				FileExNew(MemFileNew("key", []byte("16:57:51.0Z"), 0644), dname1+"/key"),
				FileExNew(MemFileNew("val", []byte("content"), 0644), dname2+"/val"),
			})
			var results s2k.Iter[Result[s2k.Batch]] = FileIter2batchIterDefault(files)
			var or s2k.Option[Result[s2k.Batch]] = results()
			checker(t, or.HasValue(), true)

			var rb Result[s2k.Batch] = or.Value()
			checker(t, rb.IsOk(), false)
		})

		t.Run("invalid files(invalid basename)", func(t *testing.T) {
			t.Parallel()

			dname1 := "f00ddeadbeaffacecafe864299792458"

			var files s2k.Iter[FileEx] = s2k.IterFromArray([]FileEx{
				FileExNew(MemFileNew("bucket", []byte("data_2022_09_07_cafef00ddeadbeafface864299792458"), 0644), dname1+"/bucket"),
				FileExNew(MemFileNew("key", []byte("16:57:51.0Z"), 0644), dname1+"/key"),
				FileExNew(MemFileNew("val", []byte("content"), 0644), dname1+"/key"),
			})
			var results s2k.Iter[Result[s2k.Batch]] = FileIter2batchIterDefault(files)
			var or s2k.Option[Result[s2k.Batch]] = results()
			checker(t, or.HasValue(), true)

			var rb Result[s2k.Batch] = or.Value()
			checker(t, rb.IsOk(), false)
		})

		t.Run("valid files(single batch)", func(t *testing.T) {
			t.Parallel()

			dname1 := "f00ddeadbeaffacecafe864299792458"

			var files s2k.Iter[FileEx] = s2k.IterFromArray([]FileEx{
				FileExNew(MemFileNew("bucket", []byte("data_2022_09_07_cafef00ddeadbeafface864299792458"), 0644), dname1+"/bucket"),
				FileExNew(MemFileNew("key", []byte("16:57:51.0Z"), 0644), dname1+"/key"),
				FileExNew(MemFileNew("val", []byte("content"), 0644), dname1+"/val"),
			})
			var results s2k.Iter[Result[s2k.Batch]] = FileIter2batchIterDefault(files)
			var or s2k.Option[Result[s2k.Batch]] = results()
			checker(t, or.HasValue(), true)

			var rb Result[s2k.Batch] = or.Value()
			checker(t, rb.IsOk(), true)

			var b s2k.Batch = rb.Value()
			checker(t, b.Bucket(), "data_2022_09_07_cafef00ddeadbeafface864299792458")

			var p s2k.Pair = b.Pair()
			checkBytes(t, p.Key, []byte("16:57:51.0Z"))
			checkBytes(t, p.Val, []byte("content"))

			checker(t, results().HasValue(), false)
		})

		t.Run("valid files(multi batch)", func(t *testing.T) {
			t.Parallel()

			dname1 := "f00ddeadbeaffacecafe864299792458"
			dname2 := "deadbeaffacecafef00d864299792458"

			var files s2k.Iter[FileEx] = s2k.IterFromArray([]FileEx{
				FileExNew(MemFileNew("bucket", []byte("data_2022_09_07_cafef00ddeadbeafface864299792458"), 0644), dname1+"/bucket"),
				FileExNew(MemFileNew("key", []byte("16:57:51.0Z"), 0644), dname1+"/key"),
				FileExNew(MemFileNew("val", []byte("content"), 0644), dname1+"/val"),

				FileExNew(MemFileNew("bucket", []byte("data_2022_09_07_beaffacecafef00ddead864299792458"), 0644), dname2+"/bucket"),
				FileExNew(MemFileNew("key", []byte("16:57:51.1Z"), 0644), dname2+"/key"),
				FileExNew(MemFileNew("val", []byte("data"), 0644), dname2+"/val"),
			})
			var results s2k.Iter[Result[s2k.Batch]] = FileIter2batchIterDefault(files)

			chk := func(bucket string, key, val []byte) {
				var or s2k.Option[Result[s2k.Batch]] = results()
				checker(t, or.HasValue(), true)

				var rb Result[s2k.Batch] = or.Value()
				checker(t, rb.IsOk(), true)

				var b s2k.Batch = rb.Value()
				checker(t, b.Bucket(), bucket)

				var p s2k.Pair = b.Pair()
				checkBytes(t, p.Key, key)
				checkBytes(t, p.Val, val)
			}

			chk("data_2022_09_07_cafef00ddeadbeafface864299792458", []byte("16:57:51.0Z"), []byte("content"))
			chk("data_2022_09_07_beaffacecafef00ddead864299792458", []byte("16:57:51.1Z"), []byte("data"))

			checker(t, results().HasValue(), false)

		})
	})
}
