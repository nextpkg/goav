package message

import (
	"bytes"
	"testing"

	"github.com/nextpkg/goav/rtmp/chunk"
	"github.com/nextpkg/goav/rtmp/comm"
	"github.com/nextpkg/goav/rtmp/slab"
	"github.com/stretchr/testify/assert"
)

func TestNewSetChunkSize(t *testing.T) {
	at := assert.New(t)

	conn := &Conn{
		Slab:                slab.NewSlab(),
		Rw:                  comm.NewReadWriter(bytes.NewBuffer(nil), 1024),
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Option:              DefaultOption,
		Chunks:              make(map[uint32]*chunk.ChunkStream),
	}

	cs := conn.NewSetChunkSize(100)
	at.Equal(uint32(2), cs.Csid)
	at.Equal(uint32(0), cs.Format)
	at.Equal(uint32(1), cs.TypeID)
	at.Equal(uint32(0), cs.StreamID)
	at.Equal([]byte{0x0, 0x0, 0x0, 0x64}, cs.Data)
}

func TestNewAbort(t *testing.T) {
	at := assert.New(t)

	conn := &Conn{
		Slab:                slab.NewSlab(),
		Rw:                  comm.NewReadWriter(bytes.NewBuffer(nil), 1024),
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Option:              DefaultOption,
		Chunks:              make(map[uint32]*chunk.ChunkStream),
	}

	cs := conn.NewAbort(6)
	at.Equal(uint32(2), cs.Csid)
	at.Equal(uint32(0), cs.Format)
	at.Equal(uint32(2), cs.TypeID)
	at.Equal(uint32(0), cs.StreamID)
	at.Equal([]byte{0x0, 0x0, 0x0, 0x6}, cs.Data)
}

func TestNewAck(t *testing.T) {
	at := assert.New(t)

	conn := &Conn{
		Slab:                slab.NewSlab(),
		Rw:                  comm.NewReadWriter(bytes.NewBuffer(nil), 1024),
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Option:              DefaultOption,
		Chunks:              make(map[uint32]*chunk.ChunkStream),
	}

	cs := conn.NewAck(1024)
	at.Equal(uint32(2), cs.Csid)
	at.Equal(uint32(0), cs.Format)
	at.Equal(uint32(3), cs.TypeID)
	at.Equal(uint32(0), cs.StreamID)
	at.Equal([]byte{0x0, 0x0, 0x4, 0x0}, cs.Data)
}

func TestNewWindowAckSize(t *testing.T) {
	at := assert.New(t)

	conn := &Conn{
		Slab:                slab.NewSlab(),
		Rw:                  comm.NewReadWriter(bytes.NewBuffer(nil), 1024),
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Option:              DefaultOption,
		Chunks:              make(map[uint32]*chunk.ChunkStream),
	}

	cs := conn.NewWindowAckSize(1024)
	at.Equal(uint32(2), cs.Csid)
	at.Equal(uint32(0), cs.Format)
	at.Equal(uint32(5), cs.TypeID)
	at.Equal(uint32(0), cs.StreamID)
	at.Equal([]byte{0x0, 0x0, 0x4, 0x0}, cs.Data)
}

func TestNewSetPeerBandwidth(t *testing.T) {
	at := assert.New(t)

	conn := &Conn{
		Slab:                slab.NewSlab(),
		Rw:                  comm.NewReadWriter(bytes.NewBuffer(nil), 1024),
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Option:              DefaultOption,
		Chunks:              make(map[uint32]*chunk.ChunkStream),
	}

	cs := conn.NewSetPeerBandwidth(1024)
	at.Equal(uint32(2), cs.Csid)
	at.Equal(uint32(0), cs.Format)
	at.Equal(uint32(6), cs.TypeID)
	at.Equal(uint32(0), cs.StreamID)
	at.Equal([]byte{0x0, 0x0, 0x4, 0x0, 0x2}, cs.Data)
}
