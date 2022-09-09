package fs2kv

import (
	"os"
	"path/filepath"
)

// A InstanceNew creates new instance.
// Single instance may have many databases.
type InstanceNew func(name string) error

// A InstanceBuilderFsAuto creates new instance and returns its name.
type InstanceBuilderFsAuto func(dirname string) Result[string]

func instanceBuilderFsAutoNew(i InstanceNew) func(UniqueDirnameBuilder) InstanceBuilderFsAuto {
	return func(nameBuilder UniqueDirnameBuilder) InstanceBuilderFsAuto {
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

var instanceBuilderFsFullPathDefault InstanceNew = instanceBuilderNewFsFullPath(0755)

// InstanceBuilderFsAutoDefault creates new instance(dir).
// Instance name will be auto generated(uuid).
var InstanceBuilderFsAutoDefault InstanceBuilderFsAuto = instanceBuilderFsAutoNew(
	instanceBuilderFsFullPathDefault,
)(uuidDirnameBuilder)
