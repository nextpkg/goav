package message

import (
	"net"
	"testing"

	"github.com/nextpkg/goav/rtmp/chunk"
	"github.com/nextpkg/goav/rtmp/comm"
	"github.com/nextpkg/goav/rtmp/slab"
	"github.com/stretchr/testify/assert"
)

func TestSetBegin(t *testing.T) {
	at := assert.New(t)

	i, o := net.Pipe()
	conn := &Conn{
		Conn:                i,
		Slab:                slab.NewSlab(),
		Rw:                  comm.NewReadWriter(i, 1024),
		ChunkSize:           128,
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Chunks:              make(map[uint32]*chunk.ChunkStream),
	}

	go func() {
		err := conn.SetBegin(12)
		at.Nil(err)

		err = conn.Flush()
		at.Nil(err)
	}()

	p := make([]byte, 1024)
	n, err := o.Read(p)
	at.Nil(err)

	at.Equal([]byte{
		0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x6, 0x4,
		0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0xc,
	}, p[:n])
}

func TestSetEOF(t *testing.T) {
	at := assert.New(t)

	i, o := net.Pipe()
	conn := &Conn{
		Conn:                i,
		Slab:                slab.NewSlab(),
		Rw:                  comm.NewReadWriter(i, 1024),
		ChunkSize:           128,
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Chunks:              make(map[uint32]*chunk.ChunkStream),
	}

	go func() {
		err := conn.SetEOF(12)
		at.Nil(err)

		err = conn.Flush()
		at.Nil(err)
	}()

	p := make([]byte, 1024)
	n, err := o.Read(p)
	at.Nil(err)

	at.Equal([]byte{
		0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x6, 0x4,
		0x1, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0,
		0x0, 0xc,
	}, p[:n])
}

func TestSetDry(t *testing.T) {
	at := assert.New(t)

	i, o := net.Pipe()
	conn := &Conn{
		Conn:                i,
		Slab:                slab.NewSlab(),
		Rw:                  comm.NewReadWriter(i, 1024),
		ChunkSize:           128,
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Chunks:              make(map[uint32]*chunk.ChunkStream),
	}

	go func() {
		err := conn.SetDry(12)
		at.Nil(err)

		err = conn.Flush()
		at.Nil(err)
	}()

	p := make([]byte, 1024)
	n, err := o.Read(p)
	at.Nil(err)

	at.Equal([]byte{
		0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x6, 0x4,
		0x1, 0x0, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0,
		0x0, 0xc,
	}, p[:n])
}

func TestSetBufferLen(t *testing.T) {
	at := assert.New(t)

	i, o := net.Pipe()
	conn := &Conn{
		Conn:                i,
		Slab:                slab.NewSlab(),
		Rw:                  comm.NewReadWriter(i, 1024),
		ChunkSize:           128,
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Chunks:              make(map[uint32]*chunk.ChunkStream),
	}

	go func() {
		err := conn.SetBufferLen(12, 1024)
		at.Nil(err)

		err = conn.Flush()
		at.Nil(err)
	}()

	p := make([]byte, 1024)
	n, err := o.Read(p)
	at.Nil(err)

	at.Equal([]byte{
		0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa, 0x4,
		0x1, 0x0, 0x0, 0x0, 0x0, 0x3, 0x0, 0x0,
		0x0, 0xc, 0x0, 0x0, 0x4, 0x0,
	}, p[:n])
}

func TestSetRecorded(t *testing.T) {
	at := assert.New(t)

	i, o := net.Pipe()
	conn := &Conn{
		Conn:                i,
		Slab:                slab.NewSlab(),
		Rw:                  comm.NewReadWriter(i, 1024),
		ChunkSize:           128,
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Chunks:              make(map[uint32]*chunk.ChunkStream),
	}

	go func() {
		err := conn.SetRecorded(12)
		at.Nil(err)

		err = conn.Flush()
		at.Nil(err)
	}()

	p := make([]byte, 1024)
	n, err := o.Read(p)
	at.Nil(err)

	at.Equal([]byte{
		0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x6, 0x4,
		0x1, 0x0, 0x0, 0x0, 0x0, 0x4, 0x0, 0x0,
		0x0, 0xc,
	}, p[:n])
}

func TestSetPingRequest(t *testing.T) {
	at := assert.New(t)

	i, o := net.Pipe()
	conn := &Conn{
		Conn:                i,
		Slab:                slab.NewSlab(),
		Rw:                  comm.NewReadWriter(i, 1024),
		ChunkSize:           128,
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Chunks:              make(map[uint32]*chunk.ChunkStream),
	}

	go func() {
		err := conn.SetPingRequest(12345)
		at.Nil(err)

		err = conn.Flush()
		at.Nil(err)
	}()

	p := make([]byte, 1024)
	n, err := o.Read(p)
	at.Nil(err)

	at.Equal([]byte{
		0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x6, 0x4,
		0x1, 0x0, 0x0, 0x0, 0x0, 0x6, 0x0, 0x0,
		0x30, 0x39,
	}, p[:n])
}

func TestSetPingResponse(t *testing.T) {
	at := assert.New(t)

	i, o := net.Pipe()
	conn := &Conn{
		Conn:                i,
		Slab:                slab.NewSlab(),
		Rw:                  comm.NewReadWriter(i, 1024),
		ChunkSize:           128,
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Chunks:              make(map[uint32]*chunk.ChunkStream),
	}

	go func() {
		err := conn.SetPingResponse(12345)
		at.Nil(err)

		err = conn.Flush()
		at.Nil(err)
	}()

	p := make([]byte, 1024)
	n, err := o.Read(p)
	at.Nil(err)

	at.Equal([]byte{
		0x2, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x6, 0x4,
		0x1, 0x0, 0x0, 0x0,
		0x0, 0x7, 0x0, 0x0,
		0x30, 0x39,
	}, p[:n])
}
