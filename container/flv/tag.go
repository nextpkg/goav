// Package flv 解析flv tag, 提取音视频中的一些元数据
package flv

import (
	"errors"
	"fmt"

	"github.com/nextpkg/goav/packet"
)

// Tag Header
type flvTag struct {
	fType     uint8
	dataSize  uint32
	timeStamp uint32
	streamID  uint32 // always 0
}

// Tag Data
type mediaTag struct {
	/*
		SoundFormat: UB[4]
		0 = Linear PCM, platform endian
		1 = ADPCM
		2 = MP3
		3 = Linear PCM, little endian
		4 = Nellymoser 16-kHz mono
		5 = Nellymoser 8-kHz mono
		6 = Nellymoser
		7 = G.711 A-law logarithmic PCM
		8 = G.711 mu-law logarithmic PCM
		9 = reserved
		10 = AAC
		11 = Speex
		14 = MP3 8-Khz
		15 = Device-specific sound
		Formats 7, 8, 14, and 15 are reserved for internal use
		AAC is supported in Flash Player 9,0,115,0 and higher.
		Speex is supported in Flash Player 10 and higher.
	*/
	soundFormat uint8

	/*
		SoundRate: UB[2]
		Sampling rate
		0 = 5.5-kHz For AAC: always 3
		1 = 11-kHz
		2 = 22-kHz
		3 = 44-kHz
	*/
	soundRate uint8

	/*
		SoundSize: UB[1]
		0 = snd8Bit
		1 = snd16Bit
		Size of each sample.
		This parameter only pertains to uncompressed formats.
		Compressed formats always decode to 16 bits internally
	*/
	soundSize uint8

	/*
		SoundType: UB[1]
		0 = sndMono
		1 = sndStereo
		Mono or stereo sound For Nellymoser: always 0
		For AAC: always 1
	*/
	soundType uint8

	/*
		0: AAC sequence header
		1: AAC raw
	*/
	aacType uint8

	/*
		1: keyframe (for AVC, a seekable frame)
		2: inter frame (for AVC, a non-seekable frame)
		3: disposable inter frame (H.263 only)
		4: generated keyframe (reserved for server use only)
		5: video info/command frame
	*/
	frameType uint8

	/*
		1: JPEG (currently unused)
		2: Sorenson H.263
		3: Screen video
		4: On2 VP6
		5: On2 VP6 with alpha channel
		6: Screen video version 2
		7: AVC
	*/
	codecID uint8

	/*
		0: AVC sequence header
		1: AVC NALU
		2: AVC end of sequence (lower level NALU sequence ender is not required or supported)
	*/
	avcType uint8

	compositionTime int32
}

// Tag Flv Body
type Tag struct {
	flv   flvTag
	media mediaTag
}

// parseVideoHeader [视频]解析 Flv包体 内的 Tag数据头部, 将 Tag数据头部 赋值给 Tag媒体结构, 并返回已处理的字节数
func (tag *Tag) parseVideoHeader(b []byte) (int, error) {
	if len(b) < 5 {
		return 0, errors.New("incomplete video header, len(b) < 5")
	}

	var n int

	// [1] 帧类型 和 编码ID
	flags := b[0]
	tag.media.frameType = flags >> 4
	tag.media.codecID = flags & 0xf
	n++

	// 只处理关键帧或者非关键帧
	if tag.media.frameType == InterFrame || tag.media.frameType == KeyFrame {
		// [2] H264包类型
		tag.media.avcType = b[1]

		// [3:5] 时间戳
		for i := 2; i < 5; i++ {
			tag.media.compositionTime = tag.media.compositionTime<<8 + int32(b[i])
		}
		n += 4
	}

	return n, nil
}

// parseAudioHeader [音频]解析 Flv包体 内的 Tag数据头部, 将 Tag数据头部 赋值给 Tag媒体结构, 并返回已处理的字节数
func (tag *Tag) parseAudioHeader(b []byte) (int, error) {
	if len(b) < 2 {
		return 0, errors.New("incomplete audio header, len(b) < 2")
	}

	var n int

	// [1] 音频格式, 码率, 大小以及声音类型
	flags := b[0]
	tag.media.soundFormat = flags >> 4
	tag.media.soundRate = (flags >> 2) & 0x3
	tag.media.soundSize = (flags >> 1) & 0x1
	tag.media.soundType = flags & 0x1
	n++

	// 根据音频格式做不同解析
	switch tag.media.soundFormat {
	case SoundAAC:
		// [2] aac包类型
		tag.media.aacType = b[1]
		n++
	case SoundMP3:
	default:
		return 0, fmt.Errorf("unexpected sound format number: %d", tag.media.soundFormat)
	}

	return n, nil
}

// SoundFormat [音频]返回音频格式
func (tag *Tag) SoundFormat() uint8 {
	return tag.media.soundFormat
}

// IsSoundAAC [音频:aac]判断音频格式是否是aac
func (tag *Tag) IsSoundAAC() bool {
	return tag.media.soundFormat == SoundAAC
}

// IsSoundMP3 [音频:mp3]判断音频格式是否是mp3
func (tag *Tag) IsSoundMP3() bool {
	return tag.media.soundFormat == SoundMP3
}

// AACType [音频:aac]返回aac的包类型
func (tag *Tag) AACType() uint8 {
	return tag.media.aacType
}

// IsAACSeqHdr [音频:aac]判断音频包类型是否是序列头
func (tag *Tag) IsAACSeqHdr() bool {
	return tag.media.aacType == AacSeqHdr
}

// IsCodecAvc [视频:h264]判断解码器是不是H264
func (tag *Tag) IsCodecAvc() bool {
	return tag.media.codecID == AvcH264
}

// IsKeyFrame [视频:h264]判断数据是否是关键帧
func (tag *Tag) IsKeyFrame() bool {
	return tag.media.frameType == KeyFrame
}

// IsInterFrame [视频:h264]判断数据是否是普通数据帧
func (tag *Tag) IsInterFrame() bool {
	return tag.media.frameType == InterFrame
}

// IsSeqHdr [视频:h264]判断数据是否是关键帧同时还是包序列的头
func (tag *Tag) IsSeqHdr() bool {
	return tag.media.frameType == KeyFrame && tag.media.avcType == AvcSeqHdr
}

// IsEndOfSeq [视频:h264]判断数据是否是关键帧同时还是包序列的尾
func (tag *Tag) IsEndOfSeq() bool {
	return tag.media.frameType == KeyFrame && tag.media.avcType == AvcEndOfSeq
}

// CodecID [视频]返回 CodecID
func (tag *Tag) CodecID() uint8 {
	return tag.media.codecID
}

// CompositionTime [视频:h264]返回 CompositionTime
func (tag *Tag) CompositionTime() int32 {
	if tag.media.avcType == AvcNalu {
		return tag.media.compositionTime
	}
	return 0
}

// ParseMediaTagHeader [音视频]解析视频, 音频中的头部数据, 将数据填充到 Tag.mediat; b是Tag Data
func (tag *Tag) ParseMediaTagHeader(b []byte, mediaType int) (int, error) {
	// 根据媒体类型做不同解析，这里是解析器封装，错误直接透传
	switch mediaType {
	case packet.PktVideo:
		return tag.parseVideoHeader(b)
	case packet.PktAudio:
		return tag.parseAudioHeader(b)
	case packet.PktMetadata:
		return 0, nil
	}

	return 0, errors.New("unexpected media type")
}
