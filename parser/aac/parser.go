package aac

import (
	"errors"
	"fmt"
	"io"

	"github.com/nextpkg/goav/container/flv"
)

// Parser aac解析器
type Parser struct {
	gotSpecific bool
	adtsHeader  []byte
	cfgInfo     *mpegCfgInfo
}

// NewParser aac解析器
func NewParser() *Parser {
	return &Parser{
		gotSpecific: false,
		adtsHeader:  make([]byte, adtsHeaderLen),
		cfgInfo:     &mpegCfgInfo{},
	}
}

// Parse 根据包类型提取和填充数据
func (p *Parser) Parse(b []byte, types uint8, w io.Writer) error {
	if len(b) == 0 {
		return errors.New("no data to parse or nil writer")
	}

	// 根据包类型提取和填充数据
	switch types {
	case flv.AacSeqHdr:
		return p.specificInfo(b)
	case flv.AacRaw:
		return p.addADTSToFrame(b, w)
	}

	return fmt.Errorf("invalid packet type(%d)", types)
}

// SampleRate 根据sampleRateIndex, 返回对应的sampleRate, 类似: map[sampleRateIndex] = sampleRate
func (p *Parser) SampleRate() int {
	rate := 44100

	if int(p.cfgInfo.sampleRateIndex) < len(aacRates) {
		rate = aacRates[p.cfgInfo.sampleRateIndex]
	}

	return rate
}

// 从aac sequence header 中提取specific config信息, 填充到 p.cfgInfo 中
// audio specific config
func (p *Parser) specificInfo(src []byte) error {
	if len(src) < 2 {
		return errors.New("audio mpeg-specific, len(src)<2")
	}

	// 填充数据
	p.gotSpecific = true
	p.cfgInfo.objectType = (src[0] >> 3) & 0xff                    /* [0:4]编码类型 */
	p.cfgInfo.sampleRateIndex = ((src[0] & 0x07) << 1) | src[1]>>7 /* [0:3]音频采样率索引值 */
	p.cfgInfo.channel = (src[1] >> 3) & 0x0f                       /* [0:3]音频输出声道 */

	// 兼容性报告
	if p.cfgInfo.sampleRateIndex == 0xf {
		return errors.New("incompatible extension sampling frequency")
	}

	return nil
}

// 向音频原始帧中插入adts(audio data transport stream)头, 形成adts帧, 写入w中
func (p *Parser) addADTSToFrame(src []byte, w io.Writer) error {
	if len(src) == 0 || !p.gotSpecific {
		return fmt.Errorf("audio data invalid, data size(%d), has specific config(%v)", len(src), p.gotSpecific)
	}

	// 音频帧大小
	aacFrameLen := uint16(len(src))

	// 含adts帧头的音频帧大小(protection_absent=1时, adts头大小为7字节)
	frameLen := (aacFrameLen + 7) & 0x1fff

	// first write adts header (params: ID: mpeg-4, protection_absent: no crc)
	p.adtsHeader[0] = 0xff /* [0:7]syncword+ */
	p.adtsHeader[1] = 0xf1 /* [0:3]+syncword, [4]ID:0 for MPEG-4;1 for MPEG-2; [5:6]layer, [7]protection_absent*/

	/*
		[0:1]profile,
		[2:5]sampling_frequency_index,
		[6]private_bit,
		[7]channel_configuration+
	*/
	p.adtsHeader[2] = (p.cfgInfo.objectType - 1) << 6
	p.adtsHeader[2] |= p.cfgInfo.sampleRateIndex << 2
	p.adtsHeader[2] |= p.cfgInfo.channel >> 2

	/*
		[0:1]+channel_configuration,
		[2]original_copy,
		[3]home,
		[4]copyright_identification_bit,
		[5]copyright_identification_start,
		[6:7]aac_frame_length+
	*/
	p.adtsHeader[3] = (p.cfgInfo.channel & 0x3) << 6
	p.adtsHeader[3] |= byte(frameLen >> 11)

	p.adtsHeader[4] = byte((frameLen & 0x7ff) >> 3) /* [0:7]+aac_frame_length+ */

	/*
		[0:2]+aac_frame_length,
		[3:7]adts_buffer_fullness+
	*/
	p.adtsHeader[5] = byte((frameLen & 0x7) << 5)
	p.adtsHeader[5] |= 0x1f

	/*
		[0:5]+adts_buffer_fullness,
		[6:7]number_of_raw_data_blocks_in_frame
	*/
	p.adtsHeader[6] = 0xfc

	// 填充adts header
	_, err := w.Write(p.adtsHeader)
	if err != nil {
		return err
	}

	// 填充body
	_, err = w.Write(src)
	if err != nil {
		return err
	}

	return nil
}
