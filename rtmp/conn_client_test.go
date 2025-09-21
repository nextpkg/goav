package core

import (
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConnClient(t *testing.T) {
	at := assert.New(t)

	i, o := net.Pipe()

	client := &ConnClient{
		conn: &Conn{
			Conn:                o,
			rw:                  NewReadWriter(o, 1024),
			slab:                NewSlab(),
			chunkSize:           128,
			remoteChunkSize:     128,
			windowAckSize:       2500000,
			remoteWindowAckSize: 2500000,
			chunks:              make(map[uint32]*ChunkStream),
			option:              defaultOption,
		},
		transactionID: 0,
	}
	server := &ConnServer{
		streamID: 1,
		conn: &Conn{
			Conn:                i,
			rw:                  NewReadWriter(i, 1024),
			slab:                NewSlab(),
			chunkSize:           128,
			remoteChunkSize:     128,
			windowAckSize:       2500000,
			remoteWindowAckSize: 2500000,
			chunks:              make(map[uint32]*ChunkStream),
			option:              defaultOption,
		},
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		at.Nil(client.Connect())
		at.Nil(client.CreateStream())
		at.Nil(client.Play())

		at.Nil(client.Connect())
		at.Nil(client.CreateStream())
		at.Nil(client.Publish())
	}()

	at.Nil(server.CommandLinkup())
	server.done = false
	at.Nil(server.CommandLinkup())

	wg.Wait()
}
