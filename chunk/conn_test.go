package chunk

import (
	"bytes"
	"io"
	"net"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestReadNormal(t *testing.T) {
	at := assert.New(t)

	// chunk 0
	data := []byte{
		0x06, 0x00, 0x00, 0x00, 0x00, 0x01, 0x33, 0x09, 0x01, 0x00, 0x00, 0x00,
	}
	data1 := make([]byte, 128)
	data = append(data, data1...)

	// chunk3
	data = append(data, 0xc6)
	data = append(data, data1...)

	// chunk3
	data2 := make([]byte, 51)
	data = append(data, 0xc6)
	data = append(data, data2...)

	conn := &Conn{
		slab:                NewSlab(),
		Rw:                  NewReadWriter(bytes.NewBuffer(data), 1024),
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Option:              DefaultOption,
		Chunks:              make(map[uint32]*ChunkStream),
	}

	var cs ChunkStream
	err := conn.Read(&cs)
	at.Equal(nil, err)
	at.Equal(6, int(cs.Csid))
	at.Equal(307, int(cs.Length))
	at.Equal(9, int(cs.TypeID))
}

func TestCrossReading(t *testing.T) {
	at := assert.New(t)
	data1 := make([]byte, 128)
	data2 := make([]byte, 51)

	// video 1 -- chunk0 -- 128bytes
	videoData := []byte{
		0x06, 0x00, 0x00, 0x00, 0x00, 0x01, 0x33, 0x09, 0x01, 0x00, 0x00, 0x00,
	}
	videoData = append(videoData, data1...)

	// video 2 -- chunk3 -- 128bytes
	videoData = append(videoData, 0xc6)
	videoData = append(videoData, data1...)

	// audio 1 -- chunk0 -- 128bytes
	audioData := []byte{
		0x04, 0x00, 0x00, 0x00, 0x00, 0x01, 0x33, 0x08, 0x01, 0x00, 0x00, 0x00,
	}
	videoData = append(videoData, audioData...)
	videoData = append(videoData, data1...)

	// audio 2 -- chunk3 -- 128bytes
	videoData = append(videoData, 0xc4)
	videoData = append(videoData, data1...)

	// video 3 -- chunk3 -- 128bytes
	videoData = append(videoData, 0xc6)
	videoData = append(videoData, data2...)

	// audio 3 -- chunk3 -- 128bytes
	videoData = append(videoData, 0xc4)
	videoData = append(videoData, data2...)

	conn := &Conn{
		slab:                NewSlab(),
		Rw:                  NewReadWriter(bytes.NewBuffer(videoData), 1024),
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Chunks:              make(map[uint32]*ChunkStream),
		Option:              DefaultOption,
	}

	// video 1
	var cs ChunkStream
	err := conn.Read(&cs)
	at.Nil(err)
	at.Equal(307, int(cs.Length))
	at.Equal(9, int(cs.TypeID))

	// audio2
	err = conn.Read(&cs)
	at.Nil(err)
	at.Equal(307, int(cs.Length))
	at.Equal(8, int(cs.TypeID))

	err = conn.Read(&cs)
	at.Equal(io.EOF, errors.Cause(err))
}

func TestSetChunksizeForWrite(t *testing.T) {
	at := assert.New(t)

	i, _ := net.Pipe()

	buf := bytes.NewBuffer(nil)
	conn := &Conn{
		Conn:                i,
		slab:                NewSlab(),
		Rw:                  NewReadWriter(buf, 1024),
		ChunkSize:           128,
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Chunks:              make(map[uint32]*ChunkStream),
		Option:              DefaultOption,
	}

	// 共146字节
	// chunk0:12字节+128字节=140字节
	// chunk3:1字节+5字节=6字节
	audioChunk := ChunkStream{
		Format:    0,
		Csid:      4,
		Timestamp: 40,
		Length:    133,
		StreamID:  1,
		TypeID:    0x8,
		Data:      make([]byte, 133),
	}
	audioChunk.Data[0] = 0xff
	audioChunk.Data[128] = 0xff
	err := conn.Write(&audioChunk)
	at.Nil(err)
	at.Nil(conn.Flush())
	at.Equal(146, buf.Len())

	// 设置chunk size
	buf.Reset()
	commandChunk := ChunkStream{
		Format:    0,
		Csid:      2,
		Timestamp: 0,
		Length:    4,
		StreamID:  1,
		TypeID:    IDSetChunkSize,
		Data:      []byte{0x00, 0x00, 0x00, 0x96},
	}
	err = conn.Write(&commandChunk)
	at.Nil(err)
	at.Nil(conn.Flush())

	buf.Reset()
	err = conn.Write(&audioChunk)
	at.Nil(err)
	at.Nil(conn.Flush())
	at.Equal(146, buf.Len())
}

func TestSetChunksize(t *testing.T) {
	at := assert.New(t)

	// 视频消息
	data := []byte{
		0x06, 0x00, 0x00, 0x00, 0x00, 0x01, 0x33, 0x09, 0x01, 0x00, 0x00, 0x00,
	}

	// chunk0
	data1 := make([]byte, 128)
	data = append(data, data1...)

	// chunk3
	data = append(data, 0xc6)
	data = append(data, data1...)

	// chunk3
	data = append(data, 0xc6)
	data2 := make([]byte, 51)
	data = append(data, data2...)

	rw := bytes.NewBuffer(data)
	conn := &Conn{
		slab:                NewSlab(),
		Rw:                  NewReadWriter(rw, 1024),
		ChunkSize:           128,
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Chunks:              make(map[uint32]*ChunkStream),
		Option:              DefaultOption,
	}

	cs := &ChunkStream{}
	err := conn.Read(cs)
	at.Nil(err)
	at.Equal(6, int(cs.Csid))
	at.Equal(9, int(cs.TypeID))
	at.Equal(1, int(cs.StreamID))
	at.Equal(307, len(cs.Data))

	// 设置块格式(Set Chunk Size(1))，如不设置，会导致接下来的数据读取错乱
	n, err := rw.Write([]byte{
		0x02, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x04, 0x01,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x96,
	})
	at.Nil(err)
	at.Equal(16, n)

	// 音频消息
	data = []byte{
		0x06, 0x00, 0x00, 0x00,
		0x00, 0x01, 0x33, 0x08,
		0x01, 0x00, 0x00, 0x00,
	}

	// chunk0
	data1 = make([]byte, 150)
	data = append(data, data1...)

	// chunk1
	data = append(data, 0xc6)
	data = append(data, data1...)

	// chunk2
	data = append(data, 0xc6)
	data2 = make([]byte, 7)
	data = append(data, data2...)

	n, err = rw.Write(data)
	at.Nil(err)
	at.Equal(321, n)

	cs = &ChunkStream{}
	err = conn.Read(cs)
	at.Equal(nil, err)
	at.Equal(6, int(cs.Csid))
	at.Equal(8, int(cs.TypeID))
	at.Equal(1, int(cs.StreamID))
	at.Equal(307, len(cs.Data))

	// 最后什么都没有了
	err = conn.Read(cs)
	at.Equal(io.EOF, errors.Cause(err))
}

func TestWrite(t *testing.T) {
	at := assert.New(t)

	i, _ := net.Pipe()

	buf := bytes.NewBuffer(nil)
	conn := &Conn{
		Conn:                i,
		slab:                NewSlab(),
		Rw:                  NewReadWriter(buf, 128),
		ChunkSize:           128,
		RemoteChunkSize:     128,
		WindowAckSize:       2500000,
		RemoteWindowAckSize: 2500000,
		Chunks:              make(map[uint32]*ChunkStream),
	}

	// 音频消息
	audioChunk := ChunkStream{
		Csid:      3,
		Timestamp: 40,
		Length:    3,
		TypeID:    8,
		Data:      []byte{0x01, 0x02, 0x03},
	}
	err := conn.Write(&audioChunk)
	at.Nil(err)
	at.Nil(conn.Flush())
	at.Equal([]byte{
		0x04, 0x00, 0x00, 0x28,
		0x00, 0x00, 0x03, 0x08,
		0x00, 0x00, 0x00, 0x00,
		0x01, 0x02, 0x03,
	}, buf.Bytes())

	// 改变时间戳和数据长度
	buf.Reset()
	audioChunk = ChunkStream{
		Csid:      3,
		Timestamp: 80,
		Length:    4,
		TypeID:    8,
		Data:      []byte{0x01, 0x02, 0x03, 0x04},
	}
	err = conn.Write(&audioChunk)
	at.Nil(err)
	at.Nil(conn.Flush())
	at.Equal([]byte{
		0x04, 0x00, 0x00, 0x50,
		0x00, 0x00, 0x04, 0x08,
		0x00, 0x00, 0x00, 0x00,
		0x01, 0x02, 0x03, 0x04,
	}, buf.Bytes())

	// 只改变时间戳
	buf.Reset()
	audioChunk.Timestamp = 160
	err = conn.Write(&audioChunk)
	at.Nil(err)
	at.Nil(conn.Flush())
	at.Equal([]byte{
		0x04, 0x00, 0x00, 0xa0,
		0x00, 0x00, 0x04, 0x08,
		0x00, 0x00, 0x00, 0x00,
		0x01, 0x02, 0x03, 0x04,
	}, buf.Bytes())
}

func TestHandleControlMsg(t *testing.T) {
	at := assert.New(t)

	conn := &Conn{}
	cs := NewSetPeerBandwidth(1024)

	at.True(conn.handleControlMsg(cs))
	at.Equal(uint32(1024), conn.ackReceived)

	cs = NewSetChunkSize(4096)
	at.True(conn.handleControlMsg(cs))
	at.Equal(uint32(4096), conn.RemoteChunkSize)
}
