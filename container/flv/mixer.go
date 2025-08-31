package flv

import (
	"bytes"
	"io"

	"github.com/nextpkg/goav/amf"
	"github.com/nextpkg/goav/packet"
)

type cache struct {
	metadata  *bytes.Buffer // 用户自定义头部信息
	avcSeqHdr *bytes.Buffer // 用来缓存AVC的序列头
	aacSeqHdr *bytes.Buffer // 用来缓存AAC的序列头
	types     *packet.Types // 媒体类型
}

// Mixer flv音视频混合器
type Mixer struct {
	flv   io.Writer
	muxer *muxer
	cache *cache
}

// NewMixer flv音视频混合器
func NewMixer(w io.Writer) *Mixer {
	return &Mixer{
		flv:   w,
		muxer: newMuxer(),
		cache: &cache{
			metadata:  bytes.NewBuffer(make([]byte, 0, 512)),
			avcSeqHdr: bytes.NewBuffer(make([]byte, 0, 512)),
			aacSeqHdr: bytes.NewBuffer(make([]byte, 0, 512)),
			types:     packet.NewTypes(),
		},
	}
}

// SetWriter 设置输出
func (m *Mixer) SetWriter(w io.Writer) {
	m.flv = w
}

// Mux 将数据转换为FLV格式
func (m *Mixer) Mux(p *packet.Packet, pts uint32) error {
	return m.muxer.mux(p, pts, m.flv)
}

// SaveMetadata 保存元数据
func (m *Mixer) SaveMetadata(md amf.Object) error {
	m.cache.metadata.Reset()

	mdPkt, err := m.muxer.metadata(md)
	if err != nil {
		return err
	}

	_, err = m.cache.metadata.Write(mdPkt)
	if err != nil {
		return err
	}

	return nil
}

// SaveAVCHeader 保存AVC序列头
func (m *Mixer) SaveAVCHeader(p *packet.Packet) error {
	m.cache.types.IsVideo()
	m.cache.avcSeqHdr.Reset()

	// 复位缓存(序列头pts为0, 作为视频的开端)
	return m.muxer.mux(p, 0, m.cache.avcSeqHdr)
}

// SaveAACHeader 保存AAC序列头
func (m *Mixer) SaveAACHeader(p *packet.Packet) error {
	m.cache.types.IsAudio()
	m.cache.aacSeqHdr.Reset()

	// 复位缓存(序列头pts为0, 作为音频的开端)
	return m.muxer.mux(p, 0, m.cache.aacSeqHdr)
}

// SetFlvHeader 设置FLV的头部数据，每次调用都会增加TS头部计数
func (m *Mixer) SetFlvHeader() error {
	mt := m.cache.types.ToSlice()

	header, err0 := m.muxer.header(mt...)
	if err0 != nil {
		return err0
	}

	_, err := m.flv.Write(header)
	if err != nil {
		return err
	}

	metadata := m.cache.metadata
	avcCache := m.cache.avcSeqHdr
	aacCache := m.cache.aacSeqHdr

	// 元数据
	if metadata.Len() > 0 {
		_, err = m.flv.Write(metadata.Bytes())
		if err != nil {
			return err
		}
	}

	// 视频序列头(h.264)
	if avcCache.Len() > 0 {
		_, err = m.flv.Write(avcCache.Bytes())
		if err != nil {
			return err
		}
	}

	// 音频序列头(aac)
	if aacCache.Len() > 0 {
		_, err = m.flv.Write(aacCache.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}
