package core

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
		rw:                  NewReadWriter(i, 1024),
		remoteChunkSize:     128,
		windowAckSize:       2500000,
		remoteWindowAckSize: 2500000,
		option:              defaultOption,
		chunks:              make(map[uint32]*ChunkStream),
	}
	server := &Conn{
		Conn:                o,
		slab:                NewSlab(),
		rw:                  NewReadWriter(o, 1024),
		remoteChunkSize:     128,
		windowAckSize:       2500000,
		remoteWindowAckSize: 2500000,
		option:              defaultOption,
		chunks:              make(map[uint32]*ChunkStream),
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

func TestGetDigest(t *testing.T) {
	at := assert.New(t)

	var random [(1 + 1536*2) * 2]byte
	C0C1C2 := random[:1536*2+1]

	C1 := C0C1C2[1 : 1536+1]

	C1[7] = 1
	copy(C1[776:], []byte{
		100, 0, 133, 133, 86, 221, 156, 123,
		183, 132, 97, 23, 222, 215, 55, 222,
		197, 31, 180, 100, 211, 72, 39, 151,
		37, 215, 164, 102, 84, 55, 44, 93,
	})

	data, err := getDigest(C1, clientPartialKey, serverFullKey)
	at.Nil(err)

	at.Equal([]byte{
		0x9b, 0xde, 0x63, 0xb9, 0x32, 0xf6, 0x5f, 0x6f,
		0x75, 0xcb, 0xd, 0xeb, 0x53, 0xab, 0x99, 0x63,
		0xcc, 0x56, 0xbb, 0x5e, 0xfb, 0x30, 0xf2, 0xa6,
		0x1f, 0x62, 0xeb, 0x62, 0x4e, 0x8c, 0x70, 0xb8,
	}, data)
}
