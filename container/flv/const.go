package flv

// AVC type
const (
	// AvcSeqHdr AVC sequence header(1)
	AvcSeqHdr = iota
	// AvcNalu AVC NALU(2)
	AvcNalu
	// AvcEndOfSeq AVC end of sequence (lower level NALU sequence ender is not required or supported)(3)
	AvcEndOfSeq
)

// AAC type
const (
	// AacSeqHdr aac sequence header(1)
	AacSeqHdr = iota
	// AacRaw aac raw(2)
	AacRaw
)

// Frame type
const (
	/*
		1: keyframe (for AVC, a seekable frame)
		2: inter frame (for AVC, a non-seekable frame)
		3: disposable inter frame (H.263 only)
		4: generated keyframe (reserved for server use only)
		5: video info/command frame
	*/
	// KeyFrame 关键帧(1)
	KeyFrame = iota + 1
	// InterFrame 分片帧(2)
	InterFrame
)

// Meta Data
const (
	// MetaDataAMF0 AMF0协议号
	MetaDataAMF0 = 0x12
	// MetaDataAMF3 AMF3协议号
	MetaDataAMF3 = 0xf
)

// AvcH264 H264的CodecID
const AvcH264 = 7

// Sound
const (
	SoundLinearPcmPlatformEndian = iota
	SoundADPCM
	SoundMP3
	SoundLinearPcmLittleEndian
	SoundNellymoser16KHzMono
	SoundNellymoser8KHzMono
	SoundNellymoser
	SoundG711ALawLogarithmicPCM
	SoundG711MuLawLogarithmicPCM
	SoundReserved
	SoundAAC
	SoundSpeex
	SoundMP38KHz             = 14
	SoundDeviceSpecificSound = 15
)

// SoundRate
const (
	SoundRate5500Hz = iota
	SoundRate11000Hz
	SoundRate22000Hz
	SoundRate44100Hz
)

// SoundSize
const (
	SoundSize8BitSamples  = 0
	SoundSize16BitSamples = 1
)

// SoundType
const (
	SoundTypeMono = iota
	SoundTypeStereo
)
