package conn

import (
	"github.com/nextpkg/goav/chunk"
	"github.com/nextpkg/goav/rtmp/comm"
)

// ReadWriteCloser RTMP读写接口
type ReadWriteCloser interface {
	// Read IO操作
	Read(*chunk.ChunkStream) error
	Write(*chunk.ChunkStream) error
	Close() error
	Flush() error

	// GetInfo 功能辅助
	GetInfo() (app, instance string)
	GetPublish() *comm.PublishInfo
	GetConnect() *comm.ConnectInfo
}
