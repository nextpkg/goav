// Package chunk provides RTMP chunk stream interfaces and implementations
package chunk

import (
	"time"
)

// Reader defines the interface for reading chunk streams
type Reader interface {
	// Read reads a chunk stream from the connection
	Read(cs *ChunkStream) error
}

// Writer defines the interface for writing chunk streams
type Writer interface {
	// Write writes a chunk stream to the connection
	Write(cs *ChunkStream) error
	// Flush flushes any buffered data
	Flush() error
}

// ChunkReadWriter combines Reader and Writer interfaces for chunk operations
type ChunkReadWriter interface {
	Reader
	Writer
}

// Connection defines the interface for RTMP connections
type Connection interface {
	ChunkReadWriter
	// Close closes the connection
	Close() error
	// SetDeadline sets the read and write deadlines
	SetDeadline(t time.Time) error
	// SetChunkSize sets the chunk size for the connection
	SetChunkSize(size uint32)
	// GetChunkSize returns the current chunk size
	GetChunkSize() uint32
}

// StreamProcessor defines the interface for processing chunk streams
type StreamProcessor interface {
	// Process processes a chunk stream
	Process(cs *ChunkStream) error
}

// ChunkHandler defines the interface for handling different chunk types
type ChunkHandler interface {
	// HandleAudio handles audio chunks
	HandleAudio(cs *ChunkStream) error
	// HandleVideo handles video chunks
	HandleVideo(cs *ChunkStream) error
	// HandleCommand handles command chunks
	HandleCommand(cs *ChunkStream) error
	// HandleData handles data chunks
	HandleData(cs *ChunkStream) error
}

// Ensure Conn implements Connection interface
var _ Connection = (*Conn)(nil)