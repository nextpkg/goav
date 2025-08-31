// Package ts Mux, PAT和PMT方法可以生成对应的ts包
package ts

import (
	"bytes"
	"fmt"
	"io"

	"github.com/nextpkg/goav/container/ts/table"
	"github.com/nextpkg/goav/packet"
)

const (
	tsDefaultDataLen = 184
	tsPacketLen      = 188
)

const (
	videoPID = 0x100
	audioPID = 0x101
)

// Muxer TS复用器
type Muxer struct {
	videoCc  byte /* 包递增计数器 */
	audioCc  byte /* 包递增计数器 */
	patCc    byte /* 包递增计数器 */
	pmtCc    byte /* 包递增计数器 */
	sdt      [tsPacketLen]byte
	pat      [tsPacketLen]byte
	pmt      [tsPacketLen]byte
	tsPacket [tsPacketLen]byte
}

// NewMuxer TS复用器
func NewMuxer() *Muxer {
	return &Muxer{}
}

// Mux 复用TS流(使用到: p.Header(FLV信息), p.data(FLV数据),p.Media(音视频数据), p.Timestamp)
// 视频数据含有B帧时, pts需要在dts的基础上加偏移量; 如果不含B帧, 则pts=dts
func (muxer *Muxer) Mux(p *packet.Packet, dts, pts int64, w io.Writer) error {
	var pid int
	var header = p.Header
	var isKeyFrame bool

	switch p.Type {
	case packet.PktVideo:
		pid = videoPID

		vh := header.(packet.VideoPacketHeader)
		isKeyFrame = vh.IsKeyFrame()
	case packet.PktAudio:
		pid = audioPID
	default:
		return fmt.Errorf("support audio and video only,type=%d", p.Type)
	}

	// 生成pes头, 获取头的长度以及pes包总长度
	pes := table.NewPes()
	pesHeaderLen := pes.GeneratePesHeader(p.Type, len(p.Media), pts, dts)
	pesTotalLen := len(p.Media) + pesHeaderLen

	// 填充ts头
	pes.TsHeader[1] = byte(pid >> 8)
	pes.TsHeader[2] = byte(pid)

	headIndex := 0
	dataIndex := 0

	// 填充ts体
	firstPes := true
	for pesTotalLen > 0 {
		// 复用头部
		copy(muxer.tsPacket[0:4], pes.TsHeader)

		// 首包标识
		if firstPes {
			muxer.tsPacket[1] |= 0x40
		}

		// 更新音视频计数器
		switch p.Type {
		case packet.PktVideo:
			muxer.videoCc++
			if muxer.videoCc > 0xf {
				muxer.videoCc = 0
			}

			muxer.tsPacket[3] |= muxer.videoCc
		case packet.PktAudio:
			muxer.audioCc++
			if muxer.audioCc > 0xf {
				muxer.audioCc = 0
			}

			muxer.tsPacket[3] |= muxer.audioCc
		}

		// 去除包头4个字节, 从第5个字节开始算
		i := byte(4)

		// 首个视频关键帧需要加pcr的自适应域
		if firstPes && isKeyFrame {
			// 既有负载也有附加区域
			muxer.tsPacket[3] |= 0x20

			// 自适应域长度
			muxer.tsPacket[4] = 0x7

			// 包含PCR
			muxer.tsPacket[5] = 0x50

			// 写入PCR
			pes.WritePcr(muxer.tsPacket[6:], dts)

			i += 1 + muxer.tsPacket[4]
		}

		// ts包体数据的长度(去除包头)
		maxPayloadLen := tsPacketLen - i

		// 计算待填充的长度
		if pesTotalLen < tsDefaultDataLen {
			// 除去pes外, 还应填充的无效字节长度
			remainBytes := maxPayloadLen - byte(pesTotalLen)

			// pes的有效数据长度
			maxPayloadLen = byte(pesTotalLen)

			// 首包或尾包有自适应域, 填充无效字符
			muxer.tsPacket[3] |= 0x20
			pes.WriteStuff(muxer.tsPacket[i:], remainBytes)

			i += remainBytes
		}

		// 先填充pes包头
		if pesHeaderLen > 0 {
			// ts中还可以用来承载的空间
			remainToWrite := maxPayloadLen

			// 计算出实际还需要的空间
			if byte(pesHeaderLen) <= remainToWrite {
				remainToWrite = byte(pesHeaderLen)
			}

			copy(muxer.tsPacket[i:], pes.PesHeader[headIndex:headIndex+int(remainToWrite)])

			// pes包头+pes包体总长度
			pesTotalLen -= int(remainToWrite)

			// pes包头总长度
			pesHeaderLen -= int(remainToWrite)

			// 单个ts包还可以承载的空间
			maxPayloadLen -= remainToWrite

			// pes包头读取位置索引
			headIndex += int(remainToWrite)

			i += remainToWrite
		}

		// 如果还有剩余的空间, 则继续填充pes包体
		if maxPayloadLen > 0 {
			if dataIndex+int(maxPayloadLen) > len(p.Media) {
				return fmt.Errorf("index is too long(%d + %d > %d)", dataIndex, maxPayloadLen, len(p.Media))
			}

			copy(muxer.tsPacket[i:], p.Media[dataIndex:dataIndex+int(maxPayloadLen)])
			dataIndex += int(maxPayloadLen)
			pesTotalLen -= int(maxPayloadLen)
		}

		// 写出准备好的ts数据
		_, err := w.Write(muxer.tsPacket[:])
		if err != nil {
			return err
		}

		firstPes = false
	}

	return nil
}

// SDT make service description table
func (muxer *Muxer) SDT(desc *bytes.Buffer) []byte {
	sdt := table.NewSdt()

	// 填充TS头
	i := 0
	copy(muxer.sdt[i:], sdt.TsHeader)
	i += len(sdt.TsHeader)

	// 修正长度, 填充SDT头
	sectionLen := 17 + desc.Len()
	sdt.SdtHeader[1] |= byte(sectionLen>>8) & 0x0F
	sdt.SdtHeader[2] = byte(sectionLen)
	sdt.SdtHeader[14] |= byte((sectionLen-17)>>8) & 0x0F
	sdt.SdtHeader[15] = byte(sectionLen - 17)
	copy(muxer.sdt[i:], sdt.SdtHeader)
	i += len(sdt.SdtHeader)

	// 填充描述信息
	copy(muxer.sdt[i:], desc.Bytes())
	i += desc.Len()

	// 计算CRC32
	crc32Value := GenerateCrc32(muxer.sdt[len(sdt.TsHeader):i])
	muxer.sdt[i] = byte(crc32Value >> 24)
	i++
	muxer.sdt[i] = byte(crc32Value >> 16)
	i++
	muxer.sdt[i] = byte(crc32Value >> 8)
	i++
	muxer.sdt[i] = byte(crc32Value)
	i++

	for j := 0; j < tsPacketLen-i; j++ {
		muxer.sdt[i+j] = 0xff
	}

	return muxer.sdt[:]
}

// PAT make program associate table
func (muxer *Muxer) PAT() []byte {
	pat := table.NewPat()

	// 填写包递增计数器, 共4位, 超出则归零
	if muxer.patCc > 0xf {
		muxer.patCc = 0
	}
	pat.TsHeader[3] |= muxer.patCc & 0x0f
	muxer.patCc++

	i := 0
	copy(muxer.pat[i:], pat.TsHeader)
	i += len(pat.TsHeader)

	copy(muxer.pat[i:], pat.PatHeader)
	i += len(pat.PatHeader)

	// 计算CRC32
	crc32Value := GenerateCrc32(pat.PatHeader)
	muxer.pat[i] = byte(crc32Value >> 24)
	i++
	muxer.pat[i] = byte(crc32Value >> 16)
	i++
	muxer.pat[i] = byte(crc32Value >> 8)
	i++
	muxer.pat[i] = byte(crc32Value)
	i++

	for j := 0; j < tsPacketLen-i; j++ {
		muxer.pat[i+j] = 0xff
	}

	return muxer.pat[:]
}

// PMT make program map table, mediaType: PktVideo or PktAudio
func (muxer *Muxer) PMT(mediaType ...int) []byte {
	pmt := table.NewPmt()
	pro := table.NewProgram()

	// 填充节目信息
	var programInfo bytes.Buffer
	for _, v := range mediaType {
		switch v {
		case packet.PktVideo:
			// 视频节目参考时钟(PCR_PID)所在TS分组的PID: 0x00
			pmt.PmtHeader[9] = 0x00
			programInfo.Write(pro.Avc)
		case packet.PktAudio:
			// 音频节目参考时钟(PCR_PID)所在TS分组的PID: 0x01
			pmt.PmtHeader[9] = 0x01
			programInfo.Write(pro.Aac)
		}
	}

	// section length
	pmt.PmtHeader[2] = byte(programInfo.Len() + 9 + 4)

	// 填写包递增计数器, 共4位, 超出则归零
	if muxer.pmtCc > 0xf {
		muxer.pmtCc = 0
	}
	pmt.TsHeader[3] |= muxer.pmtCc & 0x0f
	muxer.pmtCc++

	i := 0
	copy(muxer.pmt[i:], pmt.TsHeader)
	i += len(pmt.TsHeader)

	copy(muxer.pmt[i:], pmt.PmtHeader)
	i += len(pmt.PmtHeader)

	copy(muxer.pmt[i:], programInfo.Bytes())
	i += programInfo.Len()

	// 计算CRC32
	crc32Value := GenerateCrc32(muxer.pmt[len(pmt.TsHeader):i])
	muxer.pmt[i] = byte(crc32Value >> 24)
	i++
	muxer.pmt[i] = byte(crc32Value >> 16)
	i++
	muxer.pmt[i] = byte(crc32Value >> 8)
	i++
	muxer.pmt[i] = byte(crc32Value)
	i++

	for j := 0; j < tsPacketLen-i; j++ {
		muxer.pmt[i+j] = 0xff
	}

	return muxer.pmt[:]
}
