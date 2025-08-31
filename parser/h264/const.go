package h264

// 帧类型
const (
	iFrame byte = iota
	pFrame
	bFrame
)

// nalu 类型
const (
	naluTypeNotDefine byte = 0
	naluTypeSlice     byte = 1  // slice_layer_without_partioning_rbsp() sliceheader
	naluTypeDpa       byte = 2  // slice_data_partition_a_layer_rbsp( ), slice_header
	naluTypeDpb       byte = 3  // slice_data_partition_b_layer_rbsp( )
	naluTypeDpc       byte = 4  // slice_data_partition_c_layer_rbsp( )
	naluTypeIdr       byte = 5  // slice_layer_without_partitioning_rbsp( ),sliceheader
	naluTypeSei       byte = 6  // sei_rbsp( )
	naluTypeSps       byte = 7  // seq_parameter_set_rbsp( )
	naluTypePps       byte = 8  // pic_parameter_set_rbsp( )
	naluTypeAud       byte = 9  // access_unit_delimiter_rbsp( )
	naluTypeEOSeq     byte = 10 // end_of_seq_rbsp( )
	naluTypeEOStream  byte = 11 // end_of_stream_rbsp( )
	naluTypeFiller    byte = 12 // filler_data_rbsp( )
)

const (
	naluBytesLen int = 4
	maxSpsPpsLen int = 2 * 1024
)

var startCode = []byte{0x00, 0x00, 0x00, 0x01}
var naluAud = []byte{0x00, 0x00, 0x00, 0x01, 0x09, 0xf0} // 音频nalu

// [AVCC]序列头
type sequenceHeader struct {
	configurationVersion byte // 8bits
	avcProfileIndication byte // 8bits
	profileCompatility   byte // 8bits
	avcLevelIndication   byte // 8bits
	reserved1            byte // 6bits
	naluLen              byte // 2bits
	reserved2            byte // 3bits
	spsNum               byte // 5bits
	ppsNum               byte // 8bits
	spsLen               int
	ppsLen               int
}
