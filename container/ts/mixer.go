// Package ts muxer的上层封装，加入了音视频同步
package ts

import (
	"bytes"
	"io"

	"github.com/nextpkg/goav/amf"
	"github.com/nextpkg/goav/container/ts/table"
	"github.com/nextpkg/goav/packet"
	"github.com/nextpkg/goav/parser"
)

type cache struct {
	metadata  *bytes.Buffer // 用来缓存元数据
	avcSeqHdr *bytes.Buffer // 用来缓存AVC的序列头
	aacSeqHdr *bytes.Buffer // 用来缓存AAC的序列头
	media     *bytes.Buffer // 媒体数据
	types     *packet.Types // 媒体类型
}

// Mixer ts音视频混合器
type Mixer struct {
	ts    io.Writer
	cache *cache

	// 音视频解码
	muxer  *Muxer
	parser *parser.CodecParser

	// 音视频同步
	pts, dts int64
	sync     *sync
}

// NewMixer ts音视频混合器
func NewMixer(w io.Writer) *Mixer {
	return &Mixer{
		ts: w,
		cache: &cache{
			metadata:  bytes.NewBuffer(make([]byte, 0, 512)),
			avcSeqHdr: bytes.NewBuffer(make([]byte, 0, 512)),
			aacSeqHdr: bytes.NewBuffer(make([]byte, 0, 512)),
			media:     bytes.NewBuffer(make([]byte, 0, 512)),
			types:     packet.NewTypes(),
		},
		muxer:  NewMuxer(),
		parser: parser.NewCodecParser(),
		sync:   newSync(10),
	}
}

// SetWriter 设置输出
func (m *Mixer) SetWriter(w io.Writer) {
	m.ts = w
}

func (m *Mixer) parse(p *packet.Packet, cache *bytes.Buffer) error {
	cache.Reset()

	// 解析FLV数据，得到媒体数据
	err := m.parser.Parse(p, cache)
	if err != nil {
		return err
	}

	p.Media = cache.Bytes()
	return nil
}

// Mux 转换为ts格式（需要使用p.Media）
func (m *Mixer) Mux(p *packet.Packet) error {
	err := m.parse(p, m.cache.media)
	if err != nil {
		return err
	}

	return m.muxer.Mux(p, m.dts, m.pts, m.ts)
}

// SaveMetadata 保存元数据
func (m *Mixer) SaveMetadata(md amf.Object) error {
	provider, ok := md["Provider"].(string)
	if !ok {
		provider = "undefined"
	}
	service, ok1 := md["Service"].(string)
	if !ok1 {
		service = "undefined"
	}

	desc := table.NewDescriptor()

	// 填充描述信息(service type固定为1)
	err := desc.Service(1, provider, service)
	if err != nil {
		return err
	}

	m.cache.metadata = desc.GetBuffer()
	return nil
}

// SaveAVCHeader 保存AVC序列头（flv->avc sequence header）
func (m *Mixer) SaveAVCHeader(p *packet.Packet) error {
	m.cache.types.IsVideo()

	err := m.parse(p, m.cache.avcSeqHdr)
	if err != nil {
		return err
	}

	m.cache.avcSeqHdr.Reset()
	return m.muxer.Mux(p, 0, 0, m.cache.avcSeqHdr)
}

// SaveAACHeader 保存AAC序列头（flv->aac sequence header）
func (m *Mixer) SaveAACHeader(p *packet.Packet) error {
	m.cache.types.IsAudio()

	err := m.parse(p, m.cache.aacSeqHdr)
	if err != nil {
		return err
	}

	m.cache.aacSeqHdr.Reset()
	return m.muxer.Mux(p, 0, 0, m.cache.aacSeqHdr)
}

// SetTsHeader 封装PAT和PMT
func (m *Mixer) SetTsHeader() error {
	mediaType := m.cache.types.ToSlice()
	metadata := m.cache.metadata

	sdt := m.muxer.SDT(metadata)
	pat := m.muxer.PAT()
	pmt := m.muxer.PMT(mediaType...)

	// 输出SDT表
	_, err := m.ts.Write(sdt)
	if err != nil {
		return err
	}

	// 输出PAT表
	_, err = m.ts.Write(pat)
	if err != nil {
		return err
	}

	// 输出PMT表
	_, err = m.ts.Write(pmt)
	if err != nil {
		return err
	}

	return nil
}

// Update 计算音视频的pts和dts(最终是为了音视频同步)
// 参数解释:
// pktTs: 数据包的时间(dts)
// avcTs: H264的时间增量
//
// 备注:
// 视频的PTS=DTS+时间增量
// 音频的PTS=DTS
func (m *Mixer) Update(p *packet.Packet, pktTs, avcTs uint32) error {
	m.dts = int64(pktTs * avcHZ)

	switch p.Type {
	case packet.PktVideo:
		m.pts = m.dts + int64(avcTs*avcHZ)
	case packet.PktAudio:
		// 音频采样率
		sampleRate, err := m.parser.SampleRate()
		if err != nil {
			return err
		}

		// 以DTS为基准, 校正音频PTS, 音频时间片换算成以视频为单位的时间片(1秒钟的音频长度/音频速率 = 流逝时间)
		m.sync.syncAudioTs(&m.dts, sampleRate)
		m.pts = m.dts
	}

	return nil
}
