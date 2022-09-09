package fs2kv

import (
	"fmt"
	"os"
	"path/filepath"
)

const InstanceNameGetterEnvKeyDefault = "FS2KEYVAL_INSTANCE_NAME"

// A InstanceNew creates new instance.
// Single instance may have many databases.
type InstanceNew func(name string) error

// A InstanceBuilderFsAuto creates new instance and returns its name.
type InstanceBuilderFsAuto func(dirname string) Result[string]

type InstanceGetterEnv func() (instanceName Result[string])

func getenv(key string) (val Result[string]) {
	var v string = os.Getenv(key)
	return ResultFromBool(
		func() string { return v },
		0 < len(v),
		func() error { return fmt.Errorf("empty val for key: %s", key) },
	)
}

func instanceGetterEnvBuilderNew(key string) InstanceGetterEnv {
	return func() Result[string] {
		return getenv(key)
	}
}

var InstanceGetterEnvDefault InstanceGetterEnv = instanceGetterEnvBuilderNew(
	InstanceNameGetterEnvKeyDefault,
)

type InstanceNameProviderFs func() (dirname Result[string])

func instanceBuilderFsAutoNew(i InstanceNew) func(InstanceNameProviderFs) InstanceBuilderFsAuto {
	return func(nameBuilder InstanceNameProviderFs) InstanceBuilderFsAuto {
		return func(dirname string) Result[string] {
			var instanceName Result[string] = nameBuilder()
			var fullPath Result[string] = ResultMap(
				instanceName,
				func(s string) string { return filepath.Join(dirname, s) },
			)
			var created Result[error] = ResultMap(fullPath, i)
			return ResultFlatMap(created, func(e error) Result[string] {
				return ResultFlatMap(instanceName, func(s string) Result[string] {
					return ResultNew(s, e)
				})
			})
		}
	}
}

func instanceBuilderNewFsFullPath(mode os.FileMode) InstanceNew {
	return func(fullpath string) error {
		return os.MkdirAll(fullpath, mode)
	}
}

func instanceNameProviderEnvBuilder(key string) InstanceNameProviderFs {
	return func() Result[string] {
		return getenv(key)
	}
}

var instanceNameProviderEnv InstanceNameProviderFs = instanceNameProviderEnvBuilder(
	InstanceNameGetterEnvKeyDefault,
)

var instanceBuilderFsFullPathDefault InstanceNew = instanceBuilderNewFsFullPath(0755)

var instanceNameProviderUuid InstanceNameProviderFs = InstanceNameProviderFs(uuidDirnameBuilder)

// InstanceBuilderFsAutoDefault creates new instance(dir).
// Instance name will be auto generated(uuid).
var InstanceBuilderFsAutoDefault InstanceBuilderFsAuto = instanceBuilderFsAutoNew(
	instanceBuilderFsFullPathDefault,
)(instanceNameProviderUuid)

var InstanceBuilderFsEnvDefault InstanceBuilderFsAuto = instanceBuilderFsAutoNew(
	instanceBuilderFsFullPathDefault,
)(instanceNameProviderEnv)
