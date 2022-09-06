package fs2kv

import (
	"bytes"
	"io/fs"
	"time"
)

type MemFile struct {
	name string
	size int64
	mode fs.FileMode
	date time.Time
	val  []byte
	rdr  *bytes.Reader
}

func MemFileNew(name string, val []byte, mode fs.FileMode) *MemFile {
	size := int64(len(val))
	date := time.Now()
	rdr := bytes.NewReader(val)
	return &MemFile{
		name,
		size,
		mode,
		date,
		val,
		rdr,
	}
}

func (m *MemFile) Name() string       { return m.name }
func (m *MemFile) Size() int64        { return m.size }
func (m *MemFile) Mode() fs.FileMode  { return m.mode }
func (m *MemFile) ModTime() time.Time { return m.date }
func (m *MemFile) IsDir() bool        { return false }
func (m *MemFile) Sys() any           { return nil }

func (m *MemFile) Stat() (fs.FileInfo, error) { return m, nil }
func (m *MemFile) Read(b []byte) (int, error) { return m.rdr.Read(b) }
func (m *MemFile) Close() error               { return nil }
