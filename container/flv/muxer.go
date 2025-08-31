// Package flv 向flv文件中写入打包好的flv数据
package flv

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/nextpkg/goav/amf"
	"github.com/nextpkg/goav/packet"
)

const tagHdrLen = 11

// Muxer FLV复用器
type muxer struct {
	flvHdr []byte
	tagHdr []byte
}

// NewMuxer FLV复用器
func newMuxer() *muxer {
	ret := &muxer{
		flvHdr: []byte{0x46, 0x4c, 0x56, 0x01, 0x00, 0x00, 0x00, 0x00, 0x09, 0, 0, 0, 0},
		tagHdr: make([]byte, tagHdrLen),
	}

	// 3字节, stream id, always 0
	copy(ret.tagHdr[8:10], []byte{0, 0, 0})

	return ret
}

// Header 生成FLV头
func (m *muxer) header(mt ...int) ([]byte, error) {
	// 复位音视频选项
	m.flvHdr[4] &= 0xFA

	for _, v := range mt {
		switch v {
		case packet.PktVideo:
			m.flvHdr[4] |= 0x01
		case packet.PktAudio:
			m.flvHdr[4] |= 0x04
		default:
			return nil, fmt.Errorf("unexpected media type number='%d'\n", v)
		}
	}

	return m.flvHdr, nil
}

// Metadata 生成Metadata
func (m *muxer) metadata(metadata amf.Object) ([]byte, error) {
	pool := bytes.NewBuffer(make([]byte, 0, 256))

	// 封装metadata amf包
	en := amf.NewEnDecAMF0()

	// SetDataFrame
	_, err := en.Encode(pool, amf.SetDataFrame)
	if err != nil {
		return nil, err
	}

	// OnMetaData
	_, err = en.Encode(pool, amf.OnMetaData)
	if err != nil {
		return nil, err
	}

	// user custom
	_, err = en.Encode(pool, metadata)
	if err != nil {
		return nil, err
	}

	// 生成METADATA包
	p := packet.Packet{
		Type: packet.PktMetadata,
		Data: pool.Bytes(),
	}
	d := bytes.NewBuffer(make([]byte, 0, 256))

	// 复用媒体数据
	err = m.mux(&p, 0, d)
	if err != nil {
		return nil, err
	}

	return d.Bytes(), nil
}

// Mux 将数据转换为FLV格式
func (m *muxer) mux(p *packet.Packet, pts uint32, w io.Writer) error {
	var typeID uint32

	// 数据预处理
	switch p.Type {
	case packet.PktVideo:
		typeID = packet.TagVideo
	case packet.PktAudio:
		typeID = packet.TagAudio
	case packet.PktMetadata:
		// 默认metadata都是AMF0协议
		typeID = packet.TagScriptDataAMF0

		// 从[SetDataFrame, data]数据中提取出data(SetDataFrame是RTMP信令, 在rtmp场景外, 应该剔除在外, 只保留数据)
		data, err := amf.DelMetaHeader(p.Data, amf.NewEnDecAMF0())
		if err != nil {
			return err
		}

		p.Data = data
	default:
		return fmt.Errorf("unexpected type=%d", p.Type)
	}

	// Flv data length
	dataLen := uint32(len(p.Data))

	/* 1字节, [0:1]Reserved, [2]Filter(unencrypted), [3:7]TagType */
	/* 3字节, DataSize, 其中包含2字节的media header */
	binary.BigEndian.PutUint32(m.tagHdr[0:4], typeID<<24+dataLen&0x00ffffff)

	/* 音频/视频时间戳(dts + baseline) */
	/* Time in milliseconds, lower 24 bits */
	/* Time in milliseconds, upper 8 bits */
	binary.BigEndian.PutUint32(m.tagHdr[4:8], (pts&0x00ffffff)<<8+pts>>24)

	// 向flv文件写入 tag header 和 tag data
	// 长度: tagHdrLen
	_, err := w.Write(m.tagHdr)
	if err != nil {
		return err
	}

	// 长度: dataLen
	_, err = w.Write(p.Data)
	if err != nil {
		return err
	}

	binary.BigEndian.PutUint32(m.tagHdr[0:4], tagHdrLen+dataLen)

	// 向flv文件写入 PreviousTagSize
	_, err = w.Write(m.tagHdr[0:4])
	if err != nil {
		return err
	}

	return nil
}
