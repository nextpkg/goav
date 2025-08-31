// Package cache 缓存一帧数据
package gop

import (
	"github.com/nextpkg/goav/packet"
)

// frameCache 单帧缓存
type frameCache struct {
	packet *packet.Packet
}

// newFrameCache 单帧缓存
func newFrameCache() *frameCache {
	return &frameCache{}
}

// Write 缓存
func (f *frameCache) Write(p *packet.Packet) {
	f.packet = p
}

// SendTo 发送单帧数据
func (f *frameCache) SendTo(w packet.Writer) error {
	if f.packet == nil {
		return nil
	}

	return w.Write(f.packet)
}
