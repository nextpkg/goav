package table

import (
	"testing"

	"github.com/nextpkg/goav/packet"
	"github.com/stretchr/testify/assert"
)

func TestPes_GeneratePesHeader(t *testing.T) {
	at := assert.New(t)

	pes := NewPes()

	at.Equal(14, pes.GeneratePesHeader(packet.PktVideo, 1024, 0, 0))
	at.Equal([]byte{
		0x0, 0x0, 0x1, 0xe0, 0x4, 0x8, 0x80, 0x80,
		0x5, 0x21, 0x0, 0x1, 0x0, 0x1, 0x0, 0x0,
		0x0, 0x0, 0x0,
	}, pes.PesHeader)
	at.Equal([]byte{0x47, 0x0, 0x0, 0x10}, pes.TsHeader)
}
