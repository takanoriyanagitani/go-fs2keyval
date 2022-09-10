package fs2kv

import (
	"os"
	"path/filepath"
	"testing"
	"sort"
)

func TestDb(t *testing.T) {
	t.Parallel()

	t.Run("DatabaseListFsEnvDefault", func(t *testing.T) {
		t.Parallel()

		var ITEST_FS2KV_DB_ENV_DIR string = os.Getenv("ITEST_FS2KV_DB_ENV_DIR")
		if len(ITEST_FS2KV_DB_ENV_DIR) < 1 {
			t.Skip("No db env test")
		}

		var FS2KEYVAL_INSTANCE_NAME string = os.Getenv(InstanceNameGetterEnvKeyDefault)
		if len(FS2KEYVAL_INSTANCE_NAME) < 1 {
			t.Skip("No db env test")
		}

		initDir := func(dirname string) error {
			e := os.RemoveAll(dirname)
			if nil != e {
				return e
			}

			return os.MkdirAll(dirname, 0755)
		}

		t.Run("empty", func(t *testing.T) {
			t.Parallel()

			var rootdir = filepath.Join(
				ITEST_FS2KV_DB_ENV_DIR,
				"empty-instance.d",
			)
			var dirname = filepath.Join(
				ITEST_FS2KV_DB_ENV_DIR,
				"empty-instance.d",
				FS2KEYVAL_INSTANCE_NAME,
			)
			e := initDir(dirname)
			if nil != e {
				t.Fatalf("Unable to initialize dir: %v", e)
			}

			var dbNames Result[[]string] = DatabaseListFsEnvDefault(1)(rootdir)
			checker(t, dbNames.IsOk(), true)
		})

		t.Run("missing", func(t *testing.T) {
			t.Parallel()

			var rootdir = filepath.Join(
				ITEST_FS2KV_DB_ENV_DIR,
				"missing-instance.d",
			)
			var dirname = filepath.Join(
				ITEST_FS2KV_DB_ENV_DIR,
				"missing-instance.d",
				FS2KEYVAL_INSTANCE_NAME,
			)
			e := os.RemoveAll(dirname)
			if nil != e {
				t.Fatalf("Unable to remove dir: %v", e)
			}

			var dbNames Result[[]string] = DatabaseListFsEnvDefault(1)(rootdir)
			checker(t, dbNames.IsOk(), false)
		})

		t.Run("invalid", func(t *testing.T) {
			t.Parallel()

			var rootdir = filepath.Join(
				ITEST_FS2KV_DB_ENV_DIR,
				"missing-instance.d",
			)
			var dirname = filepath.Join(
				ITEST_FS2KV_DB_ENV_DIR,
				"missing-instance.d",
				FS2KEYVAL_INSTANCE_NAME,
			)
			e := os.RemoveAll(dirname)
			if nil != e {
				t.Fatalf("Unable to remove dir: %v", e)
			}

			var parentName string = filepath.Dir(dirname)
			e = os.MkdirAll(parentName, 0755)
			if nil != e {
				t.Fatalf("Unable to create parent dir: %v", e)
			}

			f, e := os.Create(dirname) // regular file
			if nil != e {
				t.Fatalf("Unable to create invalid 'directory': %v", e)
			}
			defer f.Close()

			var dbNames Result[[]string] = DatabaseListFsEnvDefault(1)(rootdir)
			checker(t, dbNames.IsOk(), false)
		})

		t.Run("single database", func(t *testing.T) {
			t.Parallel()

			var rootdir = filepath.Join(
				ITEST_FS2KV_DB_ENV_DIR,
				"singledb-instance.d",
			)
			var dirname = filepath.Join(
				ITEST_FS2KV_DB_ENV_DIR,
				"singledb-instance.d",
				FS2KEYVAL_INSTANCE_NAME,
			)
			e := initDir(dirname)
			if nil != e {
				t.Fatalf("Unable to initialize dir: %v", e)
			}

			f, e := os.Create(filepath.Join(dirname, "db1.tar"))
			if nil != e {
				t.Fatalf("Unable to create dummy database tar file: %v", e)
			}
			defer f.Close()

			var dbNames Result[[]string] = DatabaseListFsEnvDefault(1)(rootdir)
			checker(t, dbNames.IsOk(), true)

			var ss []string = dbNames.Value()
			checker(t, len(ss), 1)

			checker(t, ss[0], "db1.tar")
		})

		t.Run("many databases", func(t *testing.T) {
			t.Parallel()

			var rootdir = filepath.Join(
				ITEST_FS2KV_DB_ENV_DIR,
				"manydb-instance.d",
			)
			var dirname = filepath.Join(
				ITEST_FS2KV_DB_ENV_DIR,
				"manydb-instance.d",
				FS2KEYVAL_INSTANCE_NAME,
			)
			e := initDir(dirname)
			if nil != e {
				t.Fatalf("Unable to initialize dir: %v", e)
			}

			createDummyDatabase := func(dbname string) error {
				f, e := os.Create(filepath.Join(dirname, dbname))
				if nil != e {
					return e
				}
				return f.Close()
			}

			requests := []string{
				"dummy-db1.tar",
				"dummy-db2.tar",
			}

			for _, req := range requests {
				e = createDummyDatabase(req)
				if nil != e {
					t.Fatalf("Unable to create dummy database: %v", e)
				}
			}

			var dbNames Result[[]string] = DatabaseListFsEnvDefault(2)(rootdir)
			checker(t, dbNames.IsOk(), true)

			var ss []string = dbNames.Value()
			sort.Strings(ss)
			checker(t, len(ss), 2)

			checker(t, ss[0], "dummy-db1.tar")
			checker(t, ss[1], "dummy-db2.tar")
		})

		t.Run("too many databases", func(t *testing.T) {
			t.Parallel()

			var rootdir = filepath.Join(
				ITEST_FS2KV_DB_ENV_DIR,
				"toomanydb-instance.d",
			)
			var dirname = filepath.Join(
				ITEST_FS2KV_DB_ENV_DIR,
				"toomanydb-instance.d",
				FS2KEYVAL_INSTANCE_NAME,
			)
			e := initDir(dirname)
			if nil != e {
				t.Fatalf("Unable to initialize dir: %v", e)
			}

			createDummyDatabase := func(dbname string) error {
				f, e := os.Create(filepath.Join(dirname, dbname))
				if nil != e {
					return e
				}
				return f.Close()
			}

			requests := []string{
				"dummy-db1.tar",
				"dummy-db2.tar",
			}

			for _, req := range requests {
				e = createDummyDatabase(req)
				if nil != e {
					t.Fatalf("Unable to create dummy database: %v", e)
				}
			}

			var dbNames Result[[]string] = DatabaseListFsEnvDefault(1)(rootdir)
			checker(t, dbNames.IsOk(), true)

			var ss []string = dbNames.Value()
			sort.Strings(ss)
			checker(t, len(ss), 1)

			checker(t, ss[0], "dummy-db1.tar")
		})
	})
}
