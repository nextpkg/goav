// Package protocol provides RTMP protocol layer interfaces
package protocol

import (
	"github.com/nextpkg/goav/chunk"
)

// Handshaker defines the interface for RTMP handshake operations
type Handshaker interface {
	// HandshakeClient performs client-side RTMP handshake
	HandshakeClient(conn *chunk.Conn) error
	// HandshakeServer performs server-side RTMP handshake
	HandshakeServer(conn *chunk.Conn) error
}

// ProtocolHandler defines the interface for handling RTMP protocol operations
type ProtocolHandler interface {
	Handshaker
	// GetVersion returns the RTMP protocol version
	GetVersion() uint8
	// IsComplexHandshake returns true if complex handshake is supported
	IsComplexHandshake() bool
}

// DefaultHandshaker provides default implementation of Handshaker interface
type DefaultHandshaker struct{}

// HandshakeClient implements Handshaker interface
func (h *DefaultHandshaker) HandshakeClient(conn *chunk.Conn) error {
	return HandshakeClient(conn)
}

// HandshakeServer implements Handshaker interface
func (h *DefaultHandshaker) HandshakeServer(conn *chunk.Conn) error {
	return HandshakeServer(conn)
}

// GetVersion returns RTMP protocol version 3
func (h *DefaultHandshaker) GetVersion() uint8 {
	return 3
}

// IsComplexHandshake returns true as this implementation supports complex handshake
func (h *DefaultHandshaker) IsComplexHandshake() bool {
	return true
}

// Ensure DefaultHandshaker implements ProtocolHandler interface
var _ ProtocolHandler = (*DefaultHandshaker)(nil)