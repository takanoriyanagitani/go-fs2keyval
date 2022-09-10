package fs2kv

import (
	"os"
	"path/filepath"
	"testing"
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
	})
}
