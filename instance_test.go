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

		t.Run("abs", func(t *testing.T){
			var name string = filepath.Join(prefix, "abs.d")

			var r Result[string] = InstanceBuilderFsAutoDefault(name)
			checker(t, r.IsOk(), true)

			var instance string = r.Value()
			checker(t, len(instance), 36)
		})
	})
}
