package comm

import (
	"github.com/nextpkg/goav/packet"
)

// WriteBaser is the basement of `write`
type WriteBaser interface {
	packet.Writer
	Close()
}

// WriteCloser 高级写接口
type WriteCloser interface {
	WriteBaser
	AliveWriter
}

// ReadCloser 高级读接口
type ReadCloser interface {
	packet.Reader
	Close()

	InfoBaser
	ActiveBaser
	GetPublish() *PublishInfo
	GetConnect() *ConnectInfo
}
