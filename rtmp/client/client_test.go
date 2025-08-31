package client

import (
	"net"
	"sync"
	"testing"

	"github.com/nextpkg/goav/rtmp/chunk"
	"github.com/nextpkg/goav/rtmp/comm"
	"github.com/nextpkg/goav/rtmp/message"
	"github.com/nextpkg/goav/rtmp/server"
	"github.com/nextpkg/goav/rtmp/slab"
	"github.com/stretchr/testify/assert"
)

func TestCommandLinkup(t *testing.T) {
	at := assert.New(t)

	i, o := net.Pipe()

	client := &ConnClient{
		Conn: &message.Conn{
			Conn:                o,
			Rw:                  comm.NewReadWriter(o, 1024),
			Slab:                slab.NewSlab(),
			ChunkSize:           128,
			RemoteChunkSize:     128,
			WindowAckSize:       2500000,
			RemoteWindowAckSize: 2500000,
			Chunks:              make(map[uint32]*chunk.ChunkStream),
			Option:              message.DefaultOption,
		},
		TransactionID: 0,
	}
	server := &server.ConnServer{
		StreamID: 1,
		Conn: &message.Conn{
			Conn:                i,
			Rw:                  comm.NewReadWriter(i, 1024),
			Slab:                slab.NewSlab(),
			ChunkSize:           128,
			RemoteChunkSize:     128,
			WindowAckSize:       2500000,
			RemoteWindowAckSize: 2500000,
			Chunks:              make(map[uint32]*chunk.ChunkStream),
			Option:              message.DefaultOption,
		},
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		// connectResp
		err := server.RespConnect(&chunk.ChunkStream{
			Csid:     12,
			StreamID: 12,
		})
		at.Nil(err)

		// playResp
		err = server.RespPlay(&chunk.ChunkStream{
			Csid:     12,
			StreamID: 12,
		})
		at.Nil(err)

		// createStreamResp
		err = server.RespCreateStream(&chunk.ChunkStream{
			Csid:     12,
			StreamID: 12,
		})
		at.Nil(err)

		// publishResp
		err = server.RespPublish(&chunk.ChunkStream{
			Csid:     12,
			StreamID: 12,
		})
		at.Nil(err)
	}()

	client.current = comm.Connect
	at.Nil(client.recvCmdMsg())

	client.current = comm.Play
	at.Nil(client.recvCmdMsg())

	client.current = comm.CreateStream
	at.Nil(client.recvCmdMsg())

	client.current = comm.Publish
	at.Nil(client.recvCmdMsg())

	wg.Wait()
}

func TestNewConnClient(t *testing.T) {
	at := assert.New(t)

	i, o := net.Pipe()

	client := &ConnClient{
		Conn: &message.Conn{
			Conn:                o,
			Rw:                  comm.NewReadWriter(o, 1024),
			Slab:                slab.NewSlab(),
			ChunkSize:           128,
			RemoteChunkSize:     128,
			WindowAckSize:       2500000,
			RemoteWindowAckSize: 2500000,
			Chunks:              make(map[uint32]*chunk.ChunkStream),
			Option:              message.DefaultOption,
		},
		TransactionID: 0,
	}
	server := &server.ConnServer{
		StreamID: 1,
		Conn: &message.Conn{
			Conn:                i,
			Rw:                  comm.NewReadWriter(i, 1024),
			Slab:                slab.NewSlab(),
			ChunkSize:           128,
			RemoteChunkSize:     128,
			WindowAckSize:       2500000,
			RemoteWindowAckSize: 2500000,
			Chunks:              make(map[uint32]*chunk.ChunkStream),
			Option:              message.DefaultOption,
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
	server.Done = false
	at.Nil(server.CommandLinkup())

	wg.Wait()
}
