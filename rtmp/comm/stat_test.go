package comm

import (
	"testing"

	"github.com/nextpkg/goav/packet"
	"github.com/stretchr/testify/assert"
)

func TestNewStatistic(t *testing.T) {
	at := assert.New(t)

	stat := NewStat()
	stat.Update(&packet.Packet{
		Type: packet.PktVideo,
		Data: make([]byte, 1024),
	})
	at.Equal(uint64(1024), stat.videoLen.Load())

	stat.Update(&packet.Packet{
		Type: packet.PktAudio,
		Data: make([]byte, 2048),
	})
	at.Equal(uint64(1024), stat.videoLen.Load())
	at.Equal(uint64(2048), stat.audioLen.Load())
}
