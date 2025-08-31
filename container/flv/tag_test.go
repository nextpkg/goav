package flv

import (
	"testing"

	"github.com/nextpkg/goav/packet"
	"github.com/stretchr/testify/assert"
)

func TestTag_ParseVideo(t *testing.T) {
	at := assert.New(t)

	// case1:
	var tag Tag

	v := []byte{
		0x17, 0x00, 0x00, 0x00, 0x00,
	}

	n, err := tag.ParseMediaTagHeader(v, packet.PktVideo)
	at.Nil(err)
	at.Equal(5, n)

	at.True(tag.IsCodecAvc())
	at.True(tag.IsKeyFrame())
	at.True(tag.IsSeqHdr())
	at.False(tag.IsEndOfSeq())
	at.Equal(byte(7), tag.CodecID())
	at.Equal(int32(0), tag.CompositionTime())

	// case2:
	tag = Tag{}
	v = []byte{
		0x27, 0x01, 0x00, 0x00, 0x00,
	}

	n, err = tag.ParseMediaTagHeader(v, packet.PktVideo)
	at.Nil(err)
	at.Equal(5, n)

	at.True(tag.IsCodecAvc())
	at.False(tag.IsKeyFrame())
	at.False(tag.IsSeqHdr())
	at.False(tag.IsEndOfSeq())
	at.Equal(byte(7), tag.CodecID())
	at.Equal(int32(0), tag.CompositionTime())
}

func TestTag_ParseAudio(t *testing.T) {
	at := assert.New(t)

	// case1:
	var tag Tag

	v := []byte{
		0xAF, 0x01,
	}

	n, err := tag.ParseMediaTagHeader(v, packet.PktAudio)
	at.Nil(err)
	at.Equal(2, n)

	at.Equal(byte(10), tag.SoundFormat())
	at.False(tag.IsSoundMP3())
	at.True(tag.IsSoundAAC())
	at.False(tag.IsAACSeqHdr())
	at.Equal(byte(1), tag.AACType())
}
