// Package flv 从av.Packet.data中解复用出flv的头和flv的包体内容
package flv

import (
	"github.com/nextpkg/goav/packet"
)

// Demuxer FLV 解复用器
type Demuxer struct{}

// NewDemuxer FLV 解复用器
func NewDemuxer() *Demuxer {
	return &Demuxer{}
}

// Demux flv解复用
// [音视频]解析FLV:Tag Data, 填充到 base.Packet 中
func (d *Demuxer) Demux(p *packet.Packet) error {
	var tag Tag

	// 解析出 Flv Body中Tag Data的媒体头, 填充到 Tag.Mediat 中
	n, err := tag.ParseMediaTagHeader(p.Data, p.Type)
	if err != nil {
		return err
	}

	p.Header = &tag
	p.Media = p.Data[n:]

	return nil
}
