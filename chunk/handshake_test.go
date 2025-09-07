package chunk

import (
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandshake(t *testing.T) {
	at := assert.New(t)

	i, o := net.Pipe()

	client := &Conn{
		Conn:                i,
		slab:                NewSlab(),
		Rw:                  NewReadWriter(i, 1024),
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Option:              DefaultOption,
		Chunks:              make(map[uint32]*ChunkStream),
	}
	server := &Conn{
		Conn:                o,
		slab:                NewSlab(),
		Rw:                  NewReadWriter(o, 1024),
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Option:              DefaultOption,
		Chunks:              make(map[uint32]*ChunkStream),
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		err := client.HandshakeClient()
		at.Nil(err)
	}()

	err := server.HandshakeServer()
	at.Nil(err)

	wg.Wait()
}
