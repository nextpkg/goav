package table

import (
	"github.com/nextpkg/goav/packet"
)

const (
	videoSID = 0xe0
	audioSID = 0xc0
)

// Pes Ts的Pes表
type Pes struct {
	TsHeader  []byte
	PesHeader []byte
}

// NewPes 新建Pes表
func NewPes() *Pes {
	return &Pes{
		TsHeader: []byte{0x47, 0x00, 0x00, 0x10},
		PesHeader: []byte{
			0x00, 0x00, 0x01, audioSID, 0xff, 0xff, 0x80, 0x80, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00,
		},
	}
}

// GeneratePesHeader 生成pes头, 返回pes头的总长度
func (pe *Pes) GeneratePesHeader(mt int, mediaDataLen int, pts, dts int64) int {
	switch mt {
	case packet.PktVideo:
		pe.PesHeader[3] = videoSID
	case packet.PktAudio:
		pe.PesHeader[3] = audioSID
	default:
		return 0
	}

	// pts
	ptsB5 := pe.encodeTs(pe.PesHeader[7], pts)
	copy(pe.PesHeader[9:], ptsB5[:])
	if mt == packet.PktVideo && pts != dts {
		pe.PesHeader[7] |= 0x40
		pe.PesHeader[8] += 0x05

		// dts
		dtsB5 := pe.encodeTs(pe.PesHeader[7], dts)
		copy(pe.PesHeader[14:], dtsB5[:])
	}

	// pes数据包的长度
	pesDataLen := int(pe.PesHeader[8])
	size := 3 + pesDataLen + mediaDataLen
	if size > 0xffff {
		size = 0
	}
	pe.PesHeader[4] = byte(size >> 8)
	pe.PesHeader[5] = byte(size)

	// 6字节的pes固定头, 3+pesDataLen的可选头
	return 6 + 3 + pesDataLen
}

// 33位时间戳编码成40位时间戳
func (pe *Pes) encodeTs(flag byte, ts int64) [5]byte {
	var val uint16
	if ts > 0x1ffffffff {
		ts -= 0x1ffffffff
	}

	var u33 [5]byte

	// 6位
	val = uint16((flag&0xc0)>>2) | ((uint16(ts>>30) & 0x07) << 1) | 1
	u33[0] = byte(val)

	// 16位
	val = ((uint16(ts>>15) & 0x7fff) << 1) | 1
	u33[1] = byte(val >> 8)
	u33[2] = byte(val)

	// 16位
	val = (uint16(ts&0x7fff) << 1) | 1
	u33[3] = byte(val >> 8)
	u33[4] = byte(val)

	return u33
}

// WriteStuff 使用自适用域填充ts包
func (pe *Pes) WriteStuff(src []byte, remainBytes byte) {
	// 自适应域长度
	src[0] = byte(remainBytes - 1)
	if src[0] > 0 {
		src[1] = 0x00
		for i := 2; i <= int(src[0]); i++ {
			src[i] = 0xff
		}
	}
}

// WritePcr 向buf中写入pcr(buf至少应有6字节长度)
func (pe *Pes) WritePcr(buf []byte, pcr int64) {
	buf[0] = byte(pcr >> 25)
	buf[1] = byte(pcr >> 17)
	buf[2] = byte(pcr >> 9)
	buf[3] = byte(pcr >> 1)
	buf[4] = byte(((pcr & 0x1) << 7) | 0x7e)
	buf[5] = 0x00
}
