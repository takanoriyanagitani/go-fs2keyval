package fs2kv

import (
	"io"
)

// A DatabaseNew prepare batch setter for io.Writer.
// For example, tar driver will create tar file.
// Single database = single archive file.
// Single archive file may contain many batches.
type DatabaseNew func(io.Writer) SetFiles
