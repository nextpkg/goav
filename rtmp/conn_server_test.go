package core

import (
	"bytes"
	"net"
	"sync"
	"testing"

	"github.com/moggle-mog/goav/amf"
	"github.com/stretchr/testify/assert"
)

func TestSendMsg(t *testing.T) {
	at := assert.New(t)

	i, o := net.Pipe()

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

		at.Nil(server.sendCmdMsg(1, 1, respResult))
		at.Nil(server.sendDataMsg(2, 2, respError))
	}()

	// cmd msg
	buf := make([]byte, 1024)
	n, err := o.Read(buf)
	at.Nil(err)
	at.Equal([]byte{
		0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa, 0x14,
		0x1, 0x0, 0x0, 0x0, 0x2, 0x0, 0x7, 0x5f,
		0x72, 0x65, 0x73, 0x75, 0x6c, 0x74,
	}, buf[:n])

	// data msg
	n, err = o.Read(buf)
	at.Nil(err)
	at.Equal([]byte{
		0x6, 0x0, 0x0, 0x0, 0x0, 0x0, 0x9, 0x12,
		0x2, 0x0, 0x0, 0x0, 0x2, 0x0, 0x6, 0x5f,
		0x65, 0x72, 0x72, 0x6f, 0x72,
	}, buf[:n])

	wg.Wait()
}

func TestPlay(t *testing.T) {
	at := assert.New(t)

	server := &ConnServer{
		streamID: 1,
		conn: &Conn{
			slab:                NewSlab(),
			chunkSize:           128,
			remoteChunkSize:     128,
			windowAckSize:       2500000,
			remoteWindowAckSize: 2500000,
			chunks:              make(map[uint32]*ChunkStream),
			option:              defaultOption,
		},
	}

	// play
	at.Nil(server.handlePlay([]interface{}{float64(123), "app"}))
	at.Equal(uint32(123), server.transactionID)
	at.Equal("app", server.publish.Name)
}

func TestPublish(t *testing.T) {
	at := assert.New(t)

	server := &ConnServer{
		streamID: 1,
		conn: &Conn{
			slab:                NewSlab(),
			chunkSize:           128,
			remoteChunkSize:     128,
			windowAckSize:       2500000,
			remoteWindowAckSize: 2500000,
			chunks:              make(map[uint32]*ChunkStream),
			option:              defaultOption,
		},
	}

	// publish
	at.Nil(server.handlePublish([]interface{}{float64(123), " ", "app", "live"}))
	at.Equal(uint32(123), server.transactionID)
	at.Equal("app", server.publish.Name)
	at.Equal("live", server.publish.Type)
}

func TestCreateStream(t *testing.T) {
	at := assert.New(t)

	server := &ConnServer{
		streamID: 1,
		conn: &Conn{
			slab:                NewSlab(),
			chunkSize:           128,
			remoteChunkSize:     128,
			windowAckSize:       2500000,
			remoteWindowAckSize: 2500000,
			chunks:              make(map[uint32]*ChunkStream),
			option:              defaultOption,
		},
	}

	// createStream
	at.Nil(server.handleCreateStream([]interface{}{float64(123)}))
	at.Equal(uint32(123), server.transactionID)
}

func TestConnect(t *testing.T) {
	at := assert.New(t)

	server := &ConnServer{
		streamID: 1,
		conn: &Conn{
			slab:                NewSlab(),
			chunkSize:           128,
			remoteChunkSize:     128,
			windowAckSize:       2500000,
			remoteWindowAckSize: 2500000,
			chunks:              make(map[uint32]*ChunkStream),
			option:              defaultOption,
		},
	}

	// connect
	at.NotNil(server.handleConnect([]interface{}{float64(123)}))

	at.Nil(server.handleConnect([]interface{}{
		float64(1),
		amf.Object{
			"app":            "app",
			"flashVer":       "123",
			"swfUrl":         "url",
			"tcUrl":          "url",
			"fpad":           true,
			"audioCodecs":    float64(13),
			"videoCodecs":    float64(14),
			"videoFunction":  float64(15),
			"pageUrl":        "url",
			"objectEncoding": float64(16),
		},
	}))
	at.Equal("app", server.connect.App)
	at.Equal("123", server.connect.FlashVer)
	at.Equal("url", server.connect.SwfURL)
	at.Equal("url", server.connect.TcURL)
	at.Equal(13, server.connect.AudioCodecs)
	at.Equal(14, server.connect.VideoCodecs)
	at.Equal(15, server.connect.VideoFunction)
	at.Equal("url", server.connect.PageURL)
	at.Equal(16, server.connect.ObjectEncoding)
}

func TestHandleDataMsg(t *testing.T) {
	at := assert.New(t)

	server := &ConnServer{
		streamID: 1,
		conn: &Conn{
			slab:                NewSlab(),
			chunkSize:           128,
			remoteChunkSize:     128,
			windowAckSize:       2500000,
			remoteWindowAckSize: 2500000,
			chunks:              make(map[uint32]*ChunkStream),
			option:              defaultOption,
		},
	}

	at.Nil(server.handleDataMsg(&ChunkStream{}))

	buf := bytes.NewBuffer(nil)
	err := amf.NewEnDecAMF0().EncodeBatch(buf, amf.SetDataFrame, amf.OnMetaData, amf.Object{
		"a": "1",
		"b": "2",
	})
	at.Nil(err)

	cs := &ChunkStream{
		typeID: 18,
		data:   buf.Bytes(),
	}
	at.Nil(server.handleDataMsg(cs))

	at.Equal("1", server.publish.MetaData["a"])
	at.Equal("2", server.publish.MetaData["b"])
}

func TestCommandLinkup(t *testing.T) {
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

		// connectResp
		err := server.rspConnect(&ChunkStream{
			csid:     12,
			streamID: 12,
		})
		at.Nil(err)

		// playResp
		err = server.rspPlay(&ChunkStream{
			csid:     12,
			streamID: 12,
		})
		at.Nil(err)

		// createStreamResp
		err = server.rspCreateStream(&ChunkStream{
			csid:     12,
			streamID: 12,
		})
		at.Nil(err)

		// publishResp
		err = server.rspPublish(&ChunkStream{
			csid:     12,
			streamID: 12,
		})
		at.Nil(err)
	}()

	client.current = Connect
	at.Nil(client.recvCmdMsg())

	client.current = Play
	at.Nil(client.recvCmdMsg())

	client.current = CreateStream
	at.Nil(client.recvCmdMsg())

	client.current = Publish
	at.Nil(client.recvCmdMsg())

	wg.Wait()
}
