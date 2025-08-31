package parser

import (
	"errors"
	"fmt"
	"io"

	"github.com/nextpkg/goav/packet"
	"github.com/nextpkg/goav/parser/aac"
	"github.com/nextpkg/goav/parser/h264"
	"github.com/nextpkg/goav/parser/mp3"
)

// CodecParser 解析器
type CodecParser struct {
	aac  *aac.Parser
	mp3  *mp3.Parser
	h264 *h264.Parser
}

// NewCodecParser [音频/视频]新建解析器
func NewCodecParser() *CodecParser {
	return &CodecParser{}
}

// Parse [音频/视频]解码（转换flv中的媒体流的格式）
func (c *CodecParser) Parse(p *packet.Packet, w io.Writer) error {
	if p.Header == nil || p.Media == nil {
		return errors.New("parser use nil packet header or nil packet media")
	}

	// 根据媒体类型使用不同的解码器解码
	switch p.Type {
	case packet.PktVideo:
		// 根据视频编码器做不同的处理
		vh := p.Header.(packet.VideoPacketHeader)
		if vh.IsCodecAvc() {
			// 初始化一个h264解析器
			if c.h264 == nil {
				c.h264 = h264.NewParser()
			}

			// 将H264打包格式转换为 Annex-b 的网络流格式, 写入w中
			return c.h264.Parse(p.Media, vh.IsSeqHdr(), w)
		}

		// 默认返回错误
		return fmt.Errorf("unexpected video codec number: %d", vh.CodecID())
	case packet.PktAudio:
		// 根据音频编码器做不同处理
		ah := p.Header.(packet.AudioPacketHeader)
		if ah.IsSoundAAC() {
			if c.aac == nil {
				c.aac = aac.NewParser()
			}

			return c.aac.Parse(p.Media, ah.AACType(), w)
		}
		if ah.IsSoundMP3() {
			if c.mp3 == nil {
				c.mp3 = mp3.NewParser()
			}

			return c.mp3.Parse(p.Media)
		}

		// 默认返回错误
		return fmt.Errorf("unexpected audio codec number: %d", ah.SoundFormat())
	}

	// 默认返回错误
	return fmt.Errorf("unexpected media type number: %d", p.Type)
}

// SampleRate [音频]采样率
func (c *CodecParser) SampleRate() (int, error) {
	if c.aac == nil && c.mp3 == nil {
		return 0, errors.New("unexpected audio codec, support aac or mp3 only")
	}

	if c.aac != nil {
		return c.aac.SampleRate(), nil
	}

	return c.mp3.SampleRate(), nil
}
