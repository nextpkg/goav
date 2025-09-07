package chunk

import (
	"bytes"
	"testing"

	"github.com/nextpkg/goav/packet"
)

// TestMakeControlMsg tests the MakeControlMsg function
func TestMakeControlMsg(t *testing.T) {
	tests := []struct {
		name     string
		id       uint32
		size     uint32
		value    uint32
		expected *ChunkStream
	}{
		{
			name:  "Set Chunk Size",
			id:    1,
			size:  4,
			value: 4096,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   1,
				StreamID: 0,
				Length:   4,
				Data:     []byte{0x00, 0x00, 0x10, 0x00},
			},
		},
		{
			name:  "Abort Message",
			id:    2,
			size:  4,
			value: 3,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   2,
				StreamID: 0,
				Length:   4,
				Data:     []byte{0x00, 0x00, 0x00, 0x03},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MakeControlMsg(tt.id, tt.size, tt.value)

			if result.Format != tt.expected.Format {
				t.Errorf("Format = %d, want %d", result.Format, tt.expected.Format)
			}
			if result.Csid != tt.expected.Csid {
				t.Errorf("Csid = %d, want %d", result.Csid, tt.expected.Csid)
			}
			if result.TypeID != tt.expected.TypeID {
				t.Errorf("TypeID = %d, want %d", result.TypeID, tt.expected.TypeID)
			}
			if result.StreamID != tt.expected.StreamID {
				t.Errorf("StreamID = %d, want %d", result.StreamID, tt.expected.StreamID)
			}
			if result.Length != tt.expected.Length {
				t.Errorf("Length = %d, want %d", result.Length, tt.expected.Length)
			}
			if !bytes.Equal(result.Data, tt.expected.Data) {
				t.Errorf("Data = %v, want %v", result.Data, tt.expected.Data)
			}
		})
	}
}

// TestUnpack tests the Unpack function
func TestUnpack(t *testing.T) {
	tests := []struct {
		name    string
		cs      *ChunkStream
		wantErr bool
	}{
		{
			name: "Audio packet - no processing needed",
			cs: &ChunkStream{
				TypeID: packet.TagAudio,
				Data:   []byte{0x01, 0x02, 0x03},
				Length: 3,
			},
			wantErr: false,
		},
		{
			name: "Video packet - no processing needed",
			cs: &ChunkStream{
				TypeID: packet.TagVideo,
				Data:   []byte{0x01, 0x02, 0x03, 0x04},
				Length: 4,
			},
			wantErr: false,
		},
		{
			name: "Script data AMF0 - empty data",
			cs: &ChunkStream{
				TypeID: packet.TagScriptDataAMF0,
				Data:   []byte{},
				Length: 0,
			},
			wantErr: true,
		},
		{
			name: "Script data AMF3 - empty data",
			cs: &ChunkStream{
				TypeID: packet.TagScriptDataAMF3,
				Data:   []byte{},
				Length: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cs.Unpack()
			if (err != nil) != tt.wantErr {
				t.Errorf("Unpack() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMakeUserControlMsg tests the MakeUserControlMsg function
func TestMakeUserControlMsg(t *testing.T) {
	tests := []struct {
		name      string
		eventType uint32
		bufLen    uint32
		expected  *ChunkStream
	}{
		{
			name:      "Stream Begin",
			eventType: 0,
			bufLen:    4,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   4,
				StreamID: 1,
				Length:   6,
				Data:     []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			},
		},
		{
			name:      "Stream EOF",
			eventType: 1,
			bufLen:    4,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   4,
				StreamID: 1,
				Length:   6,
				Data:     []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x00},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MakeUserControlMsg(tt.eventType, tt.bufLen)

			if result.Format != tt.expected.Format {
				t.Errorf("Format = %d, want %d", result.Format, tt.expected.Format)
			}
			if result.Csid != tt.expected.Csid {
				t.Errorf("Csid = %d, want %d", result.Csid, tt.expected.Csid)
			}
			if result.TypeID != tt.expected.TypeID {
				t.Errorf("TypeID = %d, want %d", result.TypeID, tt.expected.TypeID)
			}
			if result.StreamID != tt.expected.StreamID {
				t.Errorf("StreamID = %d, want %d", result.StreamID, tt.expected.StreamID)
			}
			if result.Length != tt.expected.Length {
				t.Errorf("Length = %d, want %d", result.Length, tt.expected.Length)
			}
			if !bytes.Equal(result.Data, tt.expected.Data) {
				t.Errorf("Data = %v, want %v", result.Data, tt.expected.Data)
			}
		})
	}
}

// TestSetBegin tests the SetBegin function
func TestSetBegin(t *testing.T) {
	tests := []struct {
		name     string
		streamID uint32
		expected *ChunkStream
	}{
		{
			name:     "Set begin stream 1",
			streamID: 1,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   4,
				StreamID: 1,
				Length:   6,
				Data:     []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			},
		},
		{
			name:     "Set begin stream 5",
			streamID: 5,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   4,
				StreamID: 1,
				Length:   6,
				Data:     []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x05},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SetBegin(tt.streamID)

			if result.Format != tt.expected.Format {
				t.Errorf("Format = %d, want %d", result.Format, tt.expected.Format)
			}
			if result.Csid != tt.expected.Csid {
				t.Errorf("Csid = %d, want %d", result.Csid, tt.expected.Csid)
			}
			if result.TypeID != tt.expected.TypeID {
				t.Errorf("TypeID = %d, want %d", result.TypeID, tt.expected.TypeID)
			}
			if result.StreamID != tt.expected.StreamID {
				t.Errorf("StreamID = %d, want %d", result.StreamID, tt.expected.StreamID)
			}
			if result.Length != tt.expected.Length {
				t.Errorf("Length = %d, want %d", result.Length, tt.expected.Length)
			}
			if !bytes.Equal(result.Data, tt.expected.Data) {
				t.Errorf("Data = %v, want %v", result.Data, tt.expected.Data)
			}
		})
	}
}

// TestSetEOF tests the SetEOF function
func TestSetEOF(t *testing.T) {
	tests := []struct {
		name     string
		streamID uint32
		expected *ChunkStream
	}{
		{
			name:     "Set EOF stream 1",
			streamID: 1,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   4,
				StreamID: 1,
				Length:   6,
				Data:     []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x01},
			},
		},
		{
			name:     "Set EOF stream 3",
			streamID: 3,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   4,
				StreamID: 1,
				Length:   6,
				Data:     []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x03},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SetEOF(tt.streamID)

			if result.Format != tt.expected.Format {
				t.Errorf("Format = %d, want %d", result.Format, tt.expected.Format)
			}
			if result.Csid != tt.expected.Csid {
				t.Errorf("Csid = %d, want %d", result.Csid, tt.expected.Csid)
			}
			if result.TypeID != tt.expected.TypeID {
				t.Errorf("TypeID = %d, want %d", result.TypeID, tt.expected.TypeID)
			}
			if result.StreamID != tt.expected.StreamID {
				t.Errorf("StreamID = %d, want %d", result.StreamID, tt.expected.StreamID)
			}
			if result.Length != tt.expected.Length {
				t.Errorf("Length = %d, want %d", result.Length, tt.expected.Length)
			}
			if !bytes.Equal(result.Data, tt.expected.Data) {
				t.Errorf("Data = %v, want %v", result.Data, tt.expected.Data)
			}
		})
	}
}

// TestSetDry tests the SetDry function
func TestSetDry(t *testing.T) {
	tests := []struct {
		name     string
		streamID uint32
		expected *ChunkStream
	}{
		{
			name:     "Set dry stream 1",
			streamID: 1,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   4,
				StreamID: 1,
				Length:   6,
				Data:     []byte{0x00, 0x02, 0x00, 0x00, 0x00, 0x01},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SetDry(tt.streamID)

			if result.Format != tt.expected.Format {
				t.Errorf("Format = %d, want %d", result.Format, tt.expected.Format)
			}
			if result.Csid != tt.expected.Csid {
				t.Errorf("Csid = %d, want %d", result.Csid, tt.expected.Csid)
			}
			if result.TypeID != tt.expected.TypeID {
				t.Errorf("TypeID = %d, want %d", result.TypeID, tt.expected.TypeID)
			}
			if result.StreamID != tt.expected.StreamID {
				t.Errorf("StreamID = %d, want %d", result.StreamID, tt.expected.StreamID)
			}
			if result.Length != tt.expected.Length {
				t.Errorf("Length = %d, want %d", result.Length, tt.expected.Length)
			}
			if !bytes.Equal(result.Data, tt.expected.Data) {
				t.Errorf("Data = %v, want %v", result.Data, tt.expected.Data)
			}
		})
	}
}

// TestSetBufferLen tests the SetBufferLen function
func TestSetBufferLen(t *testing.T) {
	tests := []struct {
		name     string
		streamID uint32
		buflen   uint32
		expected *ChunkStream
	}{
		{
			name:     "Set buffer length 1000 for stream 1",
			streamID: 1,
			buflen:   1000,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   4,
				StreamID: 1,
				Length:   10,
				Data:     []byte{0x00, 0x03, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x03, 0xe8},
			},
		},
		{
			name:     "Set buffer length 2048 for stream 2",
			streamID: 2,
			buflen:   2048,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   4,
				StreamID: 1,
				Length:   10,
				Data:     []byte{0x00, 0x03, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x08, 0x00},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SetBufferLen(tt.streamID, tt.buflen)

			if result.Format != tt.expected.Format {
				t.Errorf("Format = %d, want %d", result.Format, tt.expected.Format)
			}
			if result.Csid != tt.expected.Csid {
				t.Errorf("Csid = %d, want %d", result.Csid, tt.expected.Csid)
			}
			if result.TypeID != tt.expected.TypeID {
				t.Errorf("TypeID = %d, want %d", result.TypeID, tt.expected.TypeID)
			}
			if result.StreamID != tt.expected.StreamID {
				t.Errorf("StreamID = %d, want %d", result.StreamID, tt.expected.StreamID)
			}
			if result.Length != tt.expected.Length {
				t.Errorf("Length = %d, want %d", result.Length, tt.expected.Length)
			}
			if !bytes.Equal(result.Data, tt.expected.Data) {
				t.Errorf("Data = %v, want %v", result.Data, tt.expected.Data)
			}
		})
	}
}

// TestSetRecorded tests the SetRecorded function
func TestSetRecorded(t *testing.T) {
	tests := []struct {
		name     string
		streamID uint32
		expected *ChunkStream
	}{
		{
			name:     "Set recorded stream 1",
			streamID: 1,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   4,
				StreamID: 1,
				Length:   6,
				Data:     []byte{0x00, 0x04, 0x00, 0x00, 0x00, 0x01},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SetRecorded(tt.streamID)

			if result.Format != tt.expected.Format {
				t.Errorf("Format = %d, want %d", result.Format, tt.expected.Format)
			}
			if result.Csid != tt.expected.Csid {
				t.Errorf("Csid = %d, want %d", result.Csid, tt.expected.Csid)
			}
			if result.TypeID != tt.expected.TypeID {
				t.Errorf("TypeID = %d, want %d", result.TypeID, tt.expected.TypeID)
			}
			if result.StreamID != tt.expected.StreamID {
				t.Errorf("StreamID = %d, want %d", result.StreamID, tt.expected.StreamID)
			}
			if result.Length != tt.expected.Length {
				t.Errorf("Length = %d, want %d", result.Length, tt.expected.Length)
			}
			if !bytes.Equal(result.Data, tt.expected.Data) {
				t.Errorf("Data = %v, want %v", result.Data, tt.expected.Data)
			}
		})
	}
}

// TestSetPingRequest tests the SetPingRequest function
func TestSetPingRequest(t *testing.T) {
	tests := []struct {
		name      string
		timestamp uint32
		expected  *ChunkStream
	}{
		{
			name:      "Ping request with timestamp 12345",
			timestamp: 12345,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   4,
				StreamID: 1,
				Length:   6,
				Data:     []byte{0x00, 0x06, 0x00, 0x00, 0x30, 0x39},
			},
		},
		{
			name:      "Ping request with timestamp 0",
			timestamp: 0,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   4,
				StreamID: 1,
				Length:   6,
				Data:     []byte{0x00, 0x06, 0x00, 0x00, 0x00, 0x00},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SetPingRequest(tt.timestamp)

			if result.Format != tt.expected.Format {
				t.Errorf("Format = %d, want %d", result.Format, tt.expected.Format)
			}
			if result.Csid != tt.expected.Csid {
				t.Errorf("Csid = %d, want %d", result.Csid, tt.expected.Csid)
			}
			if result.TypeID != tt.expected.TypeID {
				t.Errorf("TypeID = %d, want %d", result.TypeID, tt.expected.TypeID)
			}
			if result.StreamID != tt.expected.StreamID {
				t.Errorf("StreamID = %d, want %d", result.StreamID, tt.expected.StreamID)
			}
			if result.Length != tt.expected.Length {
				t.Errorf("Length = %d, want %d", result.Length, tt.expected.Length)
			}
			if !bytes.Equal(result.Data, tt.expected.Data) {
				t.Errorf("Data = %v, want %v", result.Data, tt.expected.Data)
			}
		})
	}
}

// TestSetPingResponse tests the SetPingResponse function
func TestSetPingResponse(t *testing.T) {
	tests := []struct {
		name      string
		timestamp uint32
		expected  *ChunkStream
	}{
		{
			name:      "Ping response with timestamp 12345",
			timestamp: 12345,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   4,
				StreamID: 1,
				Length:   6,
				Data:     []byte{0x00, 0x07, 0x00, 0x00, 0x30, 0x39},
			},
		},
		{
			name:      "Ping response with timestamp 0",
			timestamp: 0,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   4,
				StreamID: 1,
				Length:   6,
				Data:     []byte{0x00, 0x07, 0x00, 0x00, 0x00, 0x00},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SetPingResponse(tt.timestamp)

			if result.Format != tt.expected.Format {
				t.Errorf("Format = %d, want %d", result.Format, tt.expected.Format)
			}
			if result.Csid != tt.expected.Csid {
				t.Errorf("Csid = %d, want %d", result.Csid, tt.expected.Csid)
			}
			if result.TypeID != tt.expected.TypeID {
				t.Errorf("TypeID = %d, want %d", result.TypeID, tt.expected.TypeID)
			}
			if result.StreamID != tt.expected.StreamID {
				t.Errorf("StreamID = %d, want %d", result.StreamID, tt.expected.StreamID)
			}
			if result.Length != tt.expected.Length {
				t.Errorf("Length = %d, want %d", result.Length, tt.expected.Length)
			}
			if !bytes.Equal(result.Data, tt.expected.Data) {
				t.Errorf("Data = %v, want %v", result.Data, tt.expected.Data)
			}
		})
	}
}

// TestNewSetChunkSize tests the NewSetChunkSize function
func TestNewSetChunkSize(t *testing.T) {
	tests := []struct {
		name      string
		chunkSize uint32
		expected  *ChunkStream
	}{
		{
			name:      "Default chunk size 128",
			chunkSize: 128,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   1,
				StreamID: 0,
				Length:   4,
				Data:     []byte{0x00, 0x00, 0x00, 0x80},
			},
		},
		{
			name:      "Large chunk size 4096",
			chunkSize: 4096,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   1,
				StreamID: 0,
				Length:   4,
				Data:     []byte{0x00, 0x00, 0x10, 0x00},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewSetChunkSize(tt.chunkSize)

			if result.Format != tt.expected.Format {
				t.Errorf("Format = %d, want %d", result.Format, tt.expected.Format)
			}
			if result.Csid != tt.expected.Csid {
				t.Errorf("Csid = %d, want %d", result.Csid, tt.expected.Csid)
			}
			if result.TypeID != tt.expected.TypeID {
				t.Errorf("TypeID = %d, want %d", result.TypeID, tt.expected.TypeID)
			}
			if result.StreamID != tt.expected.StreamID {
				t.Errorf("StreamID = %d, want %d", result.StreamID, tt.expected.StreamID)
			}
			if result.Length != tt.expected.Length {
				t.Errorf("Length = %d, want %d", result.Length, tt.expected.Length)
			}
			if !bytes.Equal(result.Data, tt.expected.Data) {
				t.Errorf("Data = %v, want %v", result.Data, tt.expected.Data)
			}
		})
	}
}

// TestNewAbort tests the NewAbort function
func TestNewAbort(t *testing.T) {
	tests := []struct {
		name     string
		csid     uint32
		expected *ChunkStream
	}{
		{
			name: "Abort chunk stream 3",
			csid: 3,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   2,
				StreamID: 0,
				Length:   4,
				Data:     []byte{0x00, 0x00, 0x00, 0x03},
			},
		},
		{
			name: "Abort chunk stream 5",
			csid: 5,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   2,
				StreamID: 0,
				Length:   4,
				Data:     []byte{0x00, 0x00, 0x00, 0x05},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewAbort(tt.csid)

			if result.Format != tt.expected.Format {
				t.Errorf("Format = %d, want %d", result.Format, tt.expected.Format)
			}
			if result.Csid != tt.expected.Csid {
				t.Errorf("Csid = %d, want %d", result.Csid, tt.expected.Csid)
			}
			if result.TypeID != tt.expected.TypeID {
				t.Errorf("TypeID = %d, want %d", result.TypeID, tt.expected.TypeID)
			}
			if result.StreamID != tt.expected.StreamID {
				t.Errorf("StreamID = %d, want %d", result.StreamID, tt.expected.StreamID)
			}
			if result.Length != tt.expected.Length {
				t.Errorf("Length = %d, want %d", result.Length, tt.expected.Length)
			}
			if !bytes.Equal(result.Data, tt.expected.Data) {
				t.Errorf("Data = %v, want %v", result.Data, tt.expected.Data)
			}
		})
	}
}

// TestNewAck tests the NewAck function
func TestNewAck(t *testing.T) {
	tests := []struct {
		name     string
		seqnum   uint32
		expected *ChunkStream
	}{
		{
			name:   "Ack sequence 1000",
			seqnum: 1000,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   3,
				StreamID: 0,
				Length:   4,
				Data:     []byte{0x00, 0x00, 0x03, 0xe8},
			},
		},
		{
			name:   "Ack sequence 0",
			seqnum: 0,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   3,
				StreamID: 0,
				Length:   4,
				Data:     []byte{0x00, 0x00, 0x00, 0x00},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewAck(tt.seqnum)

			if result.Format != tt.expected.Format {
				t.Errorf("Format = %d, want %d", result.Format, tt.expected.Format)
			}
			if result.Csid != tt.expected.Csid {
				t.Errorf("Csid = %d, want %d", result.Csid, tt.expected.Csid)
			}
			if result.TypeID != tt.expected.TypeID {
				t.Errorf("TypeID = %d, want %d", result.TypeID, tt.expected.TypeID)
			}
			if result.StreamID != tt.expected.StreamID {
				t.Errorf("StreamID = %d, want %d", result.StreamID, tt.expected.StreamID)
			}
			if result.Length != tt.expected.Length {
				t.Errorf("Length = %d, want %d", result.Length, tt.expected.Length)
			}
			if !bytes.Equal(result.Data, tt.expected.Data) {
				t.Errorf("Data = %v, want %v", result.Data, tt.expected.Data)
			}
		})
	}
}

// TestNewWindowAckSize tests the NewWindowAckSize function
func TestNewWindowAckSize(t *testing.T) {
	tests := []struct {
		name     string
		size     uint32
		expected *ChunkStream
	}{
		{
			name: "Window ack size 2500000",
			size: 2500000,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   5,
				StreamID: 0,
				Length:   4,
				Data:     []byte{0x00, 0x26, 0x25, 0xa0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewWindowAckSize(tt.size)

			if result.Format != tt.expected.Format {
				t.Errorf("Format = %d, want %d", result.Format, tt.expected.Format)
			}
			if result.Csid != tt.expected.Csid {
				t.Errorf("Csid = %d, want %d", result.Csid, tt.expected.Csid)
			}
			if result.TypeID != tt.expected.TypeID {
				t.Errorf("TypeID = %d, want %d", result.TypeID, tt.expected.TypeID)
			}
			if result.StreamID != tt.expected.StreamID {
				t.Errorf("StreamID = %d, want %d", result.StreamID, tt.expected.StreamID)
			}
			if result.Length != tt.expected.Length {
				t.Errorf("Length = %d, want %d", result.Length, tt.expected.Length)
			}
			if !bytes.Equal(result.Data, tt.expected.Data) {
				t.Errorf("Data = %v, want %v", result.Data, tt.expected.Data)
			}
		})
	}
}

// TestNewSetPeerBandwidth tests the NewSetPeerBandwidth function
func TestNewSetPeerBandwidth(t *testing.T) {
	tests := []struct {
		name      string
		bandwidth uint32
		expected  *ChunkStream
	}{
		{
			name:      "Set peer bandwidth 2500000",
			bandwidth: 2500000,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   6,
				StreamID: 0,
				Length:   5,
				Data:     []byte{0x00, 0x26, 0x25, 0xa0, 0x02},
			},
		},
		{
			name:      "Set peer bandwidth 1000000",
			bandwidth: 1000000,
			expected: &ChunkStream{
				Format:   0,
				Csid:     2,
				TypeID:   6,
				StreamID: 0,
				Length:   5,
				Data:     []byte{0x00, 0x0f, 0x42, 0x40, 0x02},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewSetPeerBandwidth(tt.bandwidth)

			if result.Format != tt.expected.Format {
				t.Errorf("Format = %d, want %d", result.Format, tt.expected.Format)
			}
			if result.Csid != tt.expected.Csid {
				t.Errorf("Csid = %d, want %d", result.Csid, tt.expected.Csid)
			}
			if result.TypeID != tt.expected.TypeID {
				t.Errorf("TypeID = %d, want %d", result.TypeID, tt.expected.TypeID)
			}
			if result.StreamID != tt.expected.StreamID {
				t.Errorf("StreamID = %d, want %d", result.StreamID, tt.expected.StreamID)
			}
			if result.Length != tt.expected.Length {
				t.Errorf("Length = %d, want %d", result.Length, tt.expected.Length)
			}
			if !bytes.Equal(result.Data, tt.expected.Data) {
				t.Errorf("Data = %v, want %v", result.Data, tt.expected.Data)
			}
		})
	}
}
