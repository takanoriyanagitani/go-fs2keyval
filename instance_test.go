package fs2kv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstance(t *testing.T) {
	t.Parallel()

	var ITEST_FS2KV_INSTANCE_DIRNAME = os.Getenv("ITEST_FS2KV_INSTANCE_DIRNAME")
	if len(ITEST_FS2KV_INSTANCE_DIRNAME) < 1 {
		t.Skip("skipping instance test")
	}

	t.Run("InstanceBuilderFsAutoDefault", func(t *testing.T) {
		t.Parallel()

		var prefix string = filepath.Join(
			ITEST_FS2KV_INSTANCE_DIRNAME,
			"InstanceBuilderFsAutoDefault",
		)

		t.Run("abs", func(t *testing.T) {
			var name string = filepath.Join(prefix, "abs.d")

			var r Result[string] = InstanceBuilderFsAutoDefault(name)
			checker(t, r.IsOk(), true)

			var instance string = r.Value()
			checker(t, len(instance), 36)
		})

		t.Cleanup(func() {
			e := os.RemoveAll(prefix)
			if nil != e {
				t.Fatalf("Unable to remove dir: %v", e)
			}
		})
	})

	t.Run("InstanceBuilderFsEnvDefault", func(t *testing.T) {
		t.Parallel()

		var FS2KEYVAL_INSTANCE_NAME string = os.Getenv(InstanceNameGetterEnvKeyDefault)
		if len(FS2KEYVAL_INSTANCE_NAME) < 1 {
			t.Skip("skipping instance name test(env)")
		}

		parentDirnames := []string{
			filepath.Join(ITEST_FS2KV_INSTANCE_DIRNAME, "env.d", "test1.d"),
			filepath.Join(ITEST_FS2KV_INSTANCE_DIRNAME, "env.d", "test2.d"),
		}

		for _, pdname := range parentDirnames {
			t.Run(pdname, func(t *testing.T) {
				var iname Result[string] = InstanceBuilderFsEnvDefault(pdname)
				checker(t, iname.IsOk(), true)

				var s string = iname.Value()
				checker(t, s, FS2KEYVAL_INSTANCE_NAME)
			})
		}
	})
}
