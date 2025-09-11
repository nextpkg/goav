package server

import (
	"bytes"
	"net"
	"sync"
	"testing"

	"github.com/nextpkg/goav/amf"
	"github.com/nextpkg/goav/chunk"
	"github.com/nextpkg/goav/rtmp/comm"
	"github.com/stretchr/testify/assert"
)

func TestSendMsg(t *testing.T) {
	at := assert.New(t)

	i, o := net.Pipe()

	server := NewConnServer(chunk.NewConn(i, chunk.DefaultOption), 128)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		at.Nil(server.sendCmdMsg(1, 1, comm.RespResult))
		at.Nil(server.sendDataMsg(2, 2, comm.RespError))
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

	server := NewConnServer(chunk.NewConn(nil, chunk.DefaultOption), 128)

	// play
	at.Nil(server.handlePlay([]interface{}{float64(123), "app"}))
	at.Equal(uint32(123), server.transactionID)
	at.Equal("app", server.publish.Name)
}

func TestPublish(t *testing.T) {
	at := assert.New(t)

	server := NewConnServer(chunk.NewConn(nil, chunk.DefaultOption), 128)

	// publish
	at.Nil(server.handlePublish([]interface{}{float64(123), " ", "app", "live"}))
	at.Equal(uint32(123), server.transactionID)
	at.Equal("app", server.publish.Name)
	at.Equal("live", server.publish.Type)
}

func TestCreateStream(t *testing.T) {
	at := assert.New(t)

	server := NewConnServer(chunk.NewConn(nil, chunk.DefaultOption), 128)

	// createStream
	at.Nil(server.handleCreateStream([]interface{}{float64(123)}))
	at.Equal(uint32(123), server.transactionID)
}

func TestConnect(t *testing.T) {
	at := assert.New(t)

	server := NewConnServer(chunk.NewConn(nil, chunk.DefaultOption), 128)

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

	server := NewConnServer(chunk.NewConn(nil, chunk.DefaultOption), 128)

	at.Nil(server.handleDataMsg(&chunk.ChunkStream{}))

	buf := bytes.NewBuffer(nil)
	err := amf.NewEnDecAMF0().EncodeBatch(buf, amf.SetDataFrame, amf.OnMetaData, amf.Object{
		"a": "1",
		"b": "2",
	})
	at.Nil(err)

	cs := &chunk.ChunkStream{
		TypeID: 18,
		Data:   buf.Bytes(),
	}
	at.Nil(server.handleDataMsg(cs))

	at.Equal("1", server.publish.MetaData["a"])
	at.Equal("2", server.publish.MetaData["b"])
}
