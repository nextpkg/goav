package core

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunkRead(t *testing.T) {
	at := assert.New(t)

	// chunk type: 0
	data := []byte{
		0x06, 0x00, 0x00, 0x00, 0x00, 0x01, 0x33, 0x09, 0x01, 0x00, 0x00, 0x00,
	}
	data1 := make([]byte, 128)
	data = append(data, data1...)

	// chunk type: 3
	data = append(data, 0xc6)
	data = append(data, data1...)

	// chunk type: 3
	data2 := make([]byte, 51)
	data = append(data, 0xc6)
	data = append(data, data2...)

	rw := NewReadWriter(bytes.NewBuffer(data), 1024)
	cs := &ChunkStream{}

	for {
		basicHeader, _ := rw.ReadUintBE(1)
		cs.formatTmp = basicHeader >> 6
		cs.csid = basicHeader & 0x3f

		err := cs.readChunk(rw, 128, NewSlab())
		at.Nil(err)

		if cs.remain == 0 {
			break
		}
	}
	at.Equal(307, int(cs.length))
	at.Equal(9, int(cs.typeID))
	at.Equal(1, int(cs.streamID))
	at.Equal(307, len(cs.data))
	at.Equal(0, int(cs.remain))

	// chunk type: 0
	data = []byte{
		0x06, 0xff, 0xff, 0xff,
		0x00, 0x01, 0x33, 0x09,
		0x01, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x05,
	}
	data = append(data, data1...)

	// chunk type: 3
	data = append(data, 0xc6)

	// extension timestamp
	data = append(data, []byte{0x00, 0x00, 0x00, 0x05}...)
	data = append(data, data1...)

	// chunk type: 3
	data = append(data, 0xc6)
	data = append(data, data2...)

	rw = NewReadWriter(bytes.NewBuffer(data), 1024)
	cs = &ChunkStream{}

	for i := 0; i < 3; i++ {
		basicHeader, _ := rw.ReadUintBE(1)
		cs.formatTmp = basicHeader >> 6
		cs.csid = basicHeader & 0x3f
		err := cs.readChunk(rw, 128, NewSlab())
		at.Nil(err)
	}
	at.Equal(307, int(cs.length))
	at.Equal(9, int(cs.typeID))
	at.Equal(1, int(cs.streamID))
	at.Equal(307, len(cs.data))
	at.Equal(true, cs.extend)
	at.Equal(5, int(cs.timestamp))
	at.Equal(0, int(cs.remain))
}

func TestChunkWrite(t *testing.T) {
	at := assert.New(t)

	cs := &ChunkStream{}
	cs.length = 307
	cs.typeID = 9
	cs.csid = 4
	cs.timestamp = 40
	cs.data = make([]byte, 307)

	buf := bytes.NewBuffer(nil)
	w := NewReadWriter(buf, 1024)

	err := cs.writeChunk(w, 128)
	at.Equal(nil, err)

	_ = w.Flush()
	at.Equal(321, len(buf.Bytes()))
}
