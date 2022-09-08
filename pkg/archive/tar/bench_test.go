package files2tar

import (
	"bufio"
	"context"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	s2k "github.com/takanoriyanagitani/go-sql2keyval"

	f2k "github.com/takanoriyanagitani/go-fs2keyval"
)

func BenchmarkTar(b *testing.B) {
	var ITEST_FS2KV_TAR_DIRNAME string = os.Getenv("ITEST_FS2KV_TAR_DIRNAME")

	if len(ITEST_FS2KV_TAR_DIRNAME) < 1 {
		b.Skip("skipping real filesystem test")
	}

	e := os.MkdirAll(ITEST_FS2KV_TAR_DIRNAME, 0755)
	if nil != e {
		b.Fatalf("Unable to create benchmark test dir: %v", e)
	}

	b.Run("Files2TarBuilderExResNew", func(b *testing.B) {
		var prefix string = filepath.Join(ITEST_FS2KV_TAR_DIRNAME, "Files2TarBuilderExResNew")

		e := os.MkdirAll(prefix, 0755)
		if nil != e {
			b.Fatalf("Unable to create dir: %v", e)
		}

		b.Run("empty", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				f, e := os.Create(filepath.Join(prefix, "empty.tar"))
				if nil != e {
					b.Fatalf("Unable to create tar file: %v", e)
				}
				defer f.Close()

				var w *bufio.Writer = bufio.NewWriter(f)
				var setter f2k.SetFiles = Files2TarBuilderExResNew(w)

				b.ResetTimer()
				e = setter(context.Background(), s2k.IterEmptyNew[f2k.Result[f2k.FileEx]]())
				if nil != e {
					b.Fatalf("Unable to set: %v", e)
				}

				e = w.Flush()
				if nil != e {
					b.Fatalf("Unable to flush: %v", e)
				}

				e = f.Sync()
				if nil != e {
					b.Fatalf("Unable to save to storage: %v", e)
				}
			}
		})

		b.Run("single file", func(b *testing.B) {
			dat := make([]byte, 8192)
			_, e := rand.Read(dat)
			if nil != e {
				b.Fatalf("Unable to populate random data: %v", e)
			}

			for i := 0; i < b.N; i++ {
				f, e := os.Create(filepath.Join(prefix, "single-file.tar"))
				if nil != e {
					b.Fatalf("Unable to create tar file: %v", e)
				}
				defer f.Close()

				var w *bufio.Writer = bufio.NewWriter(f)
				var setter f2k.SetFiles = Files2TarBuilderExResNew(w)

				files := []f2k.Result[f2k.FileEx]{
					f2k.ResultOk(f2k.FileExNew(f2k.MemFileNew("single.txt", dat, 0644), "./path/to/single.txt")),
				}

				b.ResetTimer()
				e = setter(context.Background(), s2k.IterFromArray(files))
				if nil != e {
					b.Fatalf("Unable to set: %v", e)
				}

				e = w.Flush()
				if nil != e {
					b.Fatalf("Unable to flush: %v", e)
				}

				e = f.Sync()
				if nil != e {
					b.Fatalf("Unable to save to storage: %v", e)
				}
			}
		})

		b.Run("single batch", func(b *testing.B) {
			dat := make([]byte, 8192)
			_, e := rand.Read(dat)
			if nil != e {
				b.Fatalf("Unable to populate random data: %v", e)
			}

			for i := 0; i < b.N; i++ {
				f, e := os.Create(filepath.Join(prefix, "single-batch.tar"))
				if nil != e {
					b.Fatalf("Unable to create tar file: %v", e)
				}
				defer f.Close()

				var w *bufio.Writer = bufio.NewWriter(f)
				var setter f2k.SetFiles = Files2TarBuilderExResNew(w)

				var btch s2k.Batch = s2k.BatchNew("single", []byte("key00"), dat)

				var files s2k.Iter[f2k.Result[f2k.FileEx]] = f2k.BatchIter2FilesUuid(s2k.IterFromArray([]s2k.Batch{btch}))

				b.ResetTimer()
				e = setter(context.Background(), files)
				if nil != e {
					b.Fatalf("Unable to set: %v", e)
				}

				e = w.Flush()
				if nil != e {
					b.Fatalf("Unable to flush: %v", e)
				}

				e = f.Sync()
				if nil != e {
					b.Fatalf("Unable to save to storage: %v", e)
				}
			}
		})

		b.Run("10x mini batches", func(b *testing.B) {
			rgen := func() []byte {
				dat := make([]byte, 8192)
				_, e := rand.Read(dat)
				if nil != e {
					panic(e)
				}
				return dat
			}

			for i := 0; i < b.N; i++ {
				f, e := os.Create(filepath.Join(prefix, "10x-mini-batches.tar"))
				if nil != e {
					b.Fatalf("Unable to create tar file: %v", e)
				}
				defer f.Close()

				var w *bufio.Writer = bufio.NewWriter(f)
				var setter f2k.SetFiles = Files2TarBuilderExResNew(w)

				var files s2k.Iter[f2k.Result[f2k.FileEx]] = f2k.BatchIter2FilesUuid(s2k.IterFromArray([]s2k.Batch{
					s2k.BatchNew("data_2022_09_08_0000f00ddeadbeafface864299792458", []byte("10:48:31.0Z"), rgen()),
					s2k.BatchNew("data_2022_09_08_1000f00ddeadbeafface864299792458", []byte("10:48:31.0Z"), rgen()),
					s2k.BatchNew("data_2022_09_08_2000f00ddeadbeafface864299792458", []byte("10:48:31.0Z"), rgen()),
					s2k.BatchNew("data_2022_09_08_3000f00ddeadbeafface864299792458", []byte("10:48:31.0Z"), rgen()),
					s2k.BatchNew("data_2022_09_08_4000f00ddeadbeafface864299792458", []byte("10:48:31.0Z"), rgen()),
					s2k.BatchNew("data_2022_09_08_5000f00ddeadbeafface864299792458", []byte("10:48:31.0Z"), rgen()),
					s2k.BatchNew("data_2022_09_08_6000f00ddeadbeafface864299792458", []byte("10:48:31.0Z"), rgen()),
					s2k.BatchNew("data_2022_09_08_7000f00ddeadbeafface864299792458", []byte("10:48:31.0Z"), rgen()),
					s2k.BatchNew("data_2022_09_08_8000f00ddeadbeafface864299792458", []byte("10:48:31.0Z"), rgen()),
					s2k.BatchNew("data_2022_09_08_9000f00ddeadbeafface864299792458", []byte("10:48:31.0Z"), rgen()),
				}))

				b.ResetTimer()
				e = setter(context.Background(), files)
				if nil != e {
					b.Fatalf("Unable to set: %v", e)
				}

				e = w.Flush()
				if nil != e {
					b.Fatalf("Unable to flush: %v", e)
				}

				e = f.Sync()
				if nil != e {
					b.Fatalf("Unable to save to storage: %v", e)
				}
			}
		})
	})
}
