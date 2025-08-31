package chunk

import (
	"bytes"
	"testing"

	"github.com/nextpkg/goav/rtmp/comm"
	"github.com/nextpkg/goav/rtmp/slab"
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

	rw := comm.NewReadWriter(bytes.NewBuffer(data), 1024)
	cs := &ChunkStream{}

	for {
		basicHeader, _ := rw.ReadUintBE(1)
		cs.FormatTmp = basicHeader >> 6
		cs.Csid = basicHeader & 0x3f

		err := cs.ReadChunk(rw, 128, slab.NewSlab())
		at.Nil(err)

		if cs.remain == 0 {
			break
		}
	}
	at.Equal(307, int(cs.Length))
	at.Equal(9, int(cs.TypeID))
	at.Equal(1, int(cs.StreamID))
	at.Equal(307, len(cs.Data))
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

	rw = comm.NewReadWriter(bytes.NewBuffer(data), 1024)
	cs = &ChunkStream{}

	for i := 0; i < 3; i++ {
		basicHeader, _ := rw.ReadUintBE(1)
		cs.FormatTmp = basicHeader >> 6
		cs.Csid = basicHeader & 0x3f
		err := cs.ReadChunk(rw, 128, slab.NewSlab())
		at.Nil(err)
	}
	at.Equal(307, int(cs.Length))
	at.Equal(9, int(cs.TypeID))
	at.Equal(1, int(cs.StreamID))
	at.Equal(307, len(cs.Data))
	at.Equal(true, cs.extend)
	at.Equal(5, int(cs.Timestamp))
	at.Equal(0, int(cs.remain))
}

func TestChunkWrite(t *testing.T) {
	at := assert.New(t)

	cs := &ChunkStream{}
	cs.Length = 307
	cs.TypeID = 9
	cs.Csid = 4
	cs.Timestamp = 40
	cs.Data = make([]byte, 307)

	buf := bytes.NewBuffer(nil)
	w := comm.NewReadWriter(buf, 1024)

	err := cs.WriteChunk(w, 128)
	at.Equal(nil, err)

	_ = w.Flush()
	at.Equal(321, len(buf.Bytes()))
}
