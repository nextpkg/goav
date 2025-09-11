package comm

import (
	"testing"
	"time"

	"github.com/nextpkg/goav/packet"
	"github.com/stretchr/testify/assert"
)

func TestNewRwAlive(t *testing.T) {
	at := assert.New(t)

	alive := NewRwAlive()

	alive.SetMediaTime(&packet.Packet{
		Type:     packet.PktVideo,
		Baseline: 2000,
	})
	alive.SetMediaTime(&packet.Packet{
		Type:     packet.PktAudio,
		Baseline: 1000,
	})

	at.Equal(uint32(2000), alive.LastVideoTime())
	at.Equal(uint32(1000), alive.LastAudioTime())

	at.Equal(uint32(0), alive.GetBaseTime())
	alive.RebaseTime()
	at.Equal(uint32(2000), alive.GetBaseTime())

	// Test timeout functionality
	// Set active time to current time minus 5 seconds (should not timeout with 10s limit)
	alive.active.Store(time.Now().Unix() - 5)
	at.False(alive.IsTimeout(10))
	
	// Set active time to current time minus 15 seconds (should timeout with 10s limit)
	alive.active.Store(time.Now().Unix() - 15)
	at.True(alive.IsTimeout(10))

	alive.Keepalive()
	at.Greater(alive.active.Load(), int64(0))
}
