package gop

import (
	"github.com/nextpkg/goav/packet"
)

// Cache RTMP缓存器
type Cache struct {
	Gop         *gopCache
	VideoSeqHdr *frameCache
	AudioSeqHdr *frameCache
	Metadata    *frameCache
}

// NewCache RTMP缓存器
func NewCache(gop int) *Cache {
	return &Cache{
		Gop:         newGopCache(gop),
		VideoSeqHdr: newFrameCache(),
		AudioSeqHdr: newFrameCache(),
		Metadata:    newFrameCache(),
	}
}

// WriteVideo write video packet
func (c *Cache) WriteVideo(p *packet.Packet) error {
	// 缓存视频序列头
	vh := p.Header.(packet.VideoPacketHeader)
	if vh.IsSeqHdr() {
		c.VideoSeqHdr.Write(p)
		return nil
	}

	// 缓存视频数据包(不会包含序列头)
	if vh.IsKeyFrame() || vh.IsInterFrame() {
		return c.Gop.Write(p, vh.IsKeyFrame())
	}

	return nil
}

// WriteAudio write audio packet
func (c *Cache) WriteAudio(p *packet.Packet) error {
	// 缓存音频序列头
	ah := p.Header.(packet.AudioPacketHeader)

	// aac
	if ah.IsSoundAAC() && ah.IsAACSeqHdr() {
		c.AudioSeqHdr.Write(p)
	}

	return nil
}

// Write 缓存音视频数据包
func (c *Cache) Write(p *packet.Packet) error {

	switch p.Type {
	case packet.PktVideo:
		err := c.WriteVideo(p)
		if err != nil {
			return err
		}
	case packet.PktAudio:
		err := c.WriteAudio(p)
		if err != nil {
			return err
		}
	case packet.PktMetadata:
		c.Metadata.Write(p)
	}

	return nil
}

// SendTo 向writer发送完整的缓存包
func (c *Cache) SendTo(w packet.Writer) error {
	// 1. 发送metadata
	err := c.Metadata.SendTo(w)
	if err != nil {
		return err
	}

	// 2. 发送视频序列头
	err = c.VideoSeqHdr.SendTo(w)
	if err != nil {
		return err
	}

	// 3. 发送音频序列头
	err = c.AudioSeqHdr.SendTo(w)
	if err != nil {
		return err
	}

	// 4. 发送GOP
	err = c.Gop.SendTo(w)
	if err != nil {
		return err
	}

	return nil
}
