package core

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSetChunkSize(t *testing.T) {
	at := assert.New(t)

	conn := &Conn{
		slab:                NewSlab(),
		rw:                  NewReadWriter(bytes.NewBuffer(nil), 1024),
		remoteChunkSize:     128,
		windowAckSize:       2500000,
		remoteWindowAckSize: 2500000,
		option:              defaultOption,
		chunks:              make(map[uint32]*ChunkStream),
	}

	cs := conn.NewSetChunkSize(100)
	at.Equal(uint32(2), cs.csid)
	at.Equal(uint32(0), cs.format)
	at.Equal(uint32(1), cs.typeID)
	at.Equal(uint32(0), cs.streamID)
	at.Equal([]byte{0x0, 0x0, 0x0, 0x64}, cs.data)
}

func TestNewAbort(t *testing.T) {
	at := assert.New(t)

	conn := &Conn{
		slab:                NewSlab(),
		rw:                  NewReadWriter(bytes.NewBuffer(nil), 1024),
		remoteChunkSize:     128,
		windowAckSize:       2500000,
		remoteWindowAckSize: 2500000,
		option:              defaultOption,
		chunks:              make(map[uint32]*ChunkStream),
	}

	cs := conn.NewAbort(6)
	at.Equal(uint32(2), cs.csid)
	at.Equal(uint32(0), cs.format)
	at.Equal(uint32(2), cs.typeID)
	at.Equal(uint32(0), cs.streamID)
	at.Equal([]byte{0x0, 0x0, 0x0, 0x6}, cs.data)
}

func TestNewAck(t *testing.T) {
	at := assert.New(t)

	conn := &Conn{
		slab:                NewSlab(),
		rw:                  NewReadWriter(bytes.NewBuffer(nil), 1024),
		remoteChunkSize:     128,
		windowAckSize:       2500000,
		remoteWindowAckSize: 2500000,
		option:              defaultOption,
		chunks:              make(map[uint32]*ChunkStream),
	}

	cs := conn.NewAck(1024)
	at.Equal(uint32(2), cs.csid)
	at.Equal(uint32(0), cs.format)
	at.Equal(uint32(3), cs.typeID)
	at.Equal(uint32(0), cs.streamID)
	at.Equal([]byte{0x0, 0x0, 0x4, 0x0}, cs.data)
}

func TestNewWindowAckSize(t *testing.T) {
	at := assert.New(t)

	conn := &Conn{
		slab:                NewSlab(),
		rw:                  NewReadWriter(bytes.NewBuffer(nil), 1024),
		remoteChunkSize:     128,
		windowAckSize:       2500000,
		remoteWindowAckSize: 2500000,
		option:              defaultOption,
		chunks:              make(map[uint32]*ChunkStream),
	}

	cs := conn.NewWindowAckSize(1024)
	at.Equal(uint32(2), cs.csid)
	at.Equal(uint32(0), cs.format)
	at.Equal(uint32(5), cs.typeID)
	at.Equal(uint32(0), cs.streamID)
	at.Equal([]byte{0x0, 0x0, 0x4, 0x0}, cs.data)
}

func TestNewSetPeerBandwidth(t *testing.T) {
	at := assert.New(t)

	conn := &Conn{
		slab:                NewSlab(),
		rw:                  NewReadWriter(bytes.NewBuffer(nil), 1024),
		remoteChunkSize:     128,
		windowAckSize:       2500000,
		remoteWindowAckSize: 2500000,
		option:              defaultOption,
		chunks:              make(map[uint32]*ChunkStream),
	}

	cs := conn.NewSetPeerBandwidth(1024)
	at.Equal(uint32(2), cs.csid)
	at.Equal(uint32(0), cs.format)
	at.Equal(uint32(6), cs.typeID)
	at.Equal(uint32(0), cs.streamID)
	at.Equal([]byte{0x0, 0x0, 0x4, 0x0, 0x2}, cs.data)
}
