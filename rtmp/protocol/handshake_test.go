package protocol

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/nextpkg/goav/chunk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockReadWriter implements chunk.ReadWriter interface for testing
type mockReadWriter struct {
	buf *bytes.Buffer
}

func newMockReadWriter() *mockReadWriter {
	return &mockReadWriter{
		buf: bytes.NewBuffer(nil),
	}
}

func (m *mockReadWriter) Read(p []byte) (int, error) {
	return m.buf.Read(p)
}

func (m *mockReadWriter) Write(p []byte) (int, error) {
	return m.buf.Write(p)
}

func (m *mockReadWriter) Flush() error {
	return nil
}

func (m *mockReadWriter) Close() error {
	return nil
}

func (m *mockReadWriter) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockReadWriter) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockReadWriter) SetWriteDeadline(t time.Time) error {
	return nil
}

func (m *mockReadWriter) LocalAddr() net.Addr {
	return nil
}

func (m *mockReadWriter) RemoteAddr() net.Addr {
	return nil
}

// TestHandshakeClient tests client-side handshake functionality
func TestHandshakeClient(t *testing.T) {
	t.Run("simple handshake success", func(t *testing.T) {
		// Create mock connection
		mockRW := newMockReadWriter()
		conn := chunk.NewConn(mockRW, chunk.DefaultOption)

		// Prepare server response (S0S1S2)
		serverResponse := make([]byte, 1+1536+1536)
		serverResponse[0] = 3 // S0: version 3
		// S1: 1536 bytes (can be random for simple handshake)
		// S2: 1536 bytes (can be random for simple handshake)

		// Write server response to mock buffer
		mockRW.buf.Write(serverResponse)

		// Test client handshake
		err := HandshakeClient(conn)
		require.NoError(t, err)
	})

	t.Run("invalid server version", func(t *testing.T) {
		// Create mock connection
		mockRW := newMockReadWriter()
		conn := chunk.NewConn(mockRW, chunk.DefaultOption)

		// Prepare invalid server response
		serverResponse := make([]byte, 1+1536+1536)
		serverResponse[0] = 2 // Invalid version

		// Write server response to mock buffer
		mockRW.buf.Write(serverResponse)

		// Test client handshake should fail
		err := HandshakeClient(conn)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected rtmp version 3")
	})
}

// TestHandshakeServer tests server-side handshake functionality
func TestHandshakeServer(t *testing.T) {
	t.Run("simple handshake success", func(t *testing.T) {
		// Create mock connection
		mockRW := newMockReadWriter()
		conn := chunk.NewConn(mockRW, chunk.DefaultOption)

		// Prepare client request (C0C1 + C2)
		clientRequest := make([]byte, 1+1536+1536)
		clientRequest[0] = 3 // C0: version 3
		// C1: 1536 bytes with version 0 for simple handshake
		// bytes 4-7 should be 0 for simple handshake
		// C2: 1536 bytes

		// Write client request to mock buffer
		mockRW.buf.Write(clientRequest)

		// Test server handshake
		err := HandshakeServer(conn)
		require.NoError(t, err)
	})

	t.Run("invalid client version", func(t *testing.T) {
		// Create mock connection
		mockRW := newMockReadWriter()
		conn := chunk.NewConn(mockRW, chunk.DefaultOption)

		// Prepare invalid client request
		clientRequest := make([]byte, 1+1536)
		clientRequest[0] = 2 // Invalid version

		// Write client request to mock buffer
		mockRW.buf.Write(clientRequest)

		// Test server handshake should fail
		err := HandshakeServer(conn)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid rtmp version")
	})
}

// TestHandshakeIntegration tests client and server handshake together
func TestHandshakeIntegration(t *testing.T) {
	t.Run("client server handshake integration", func(t *testing.T) {
		// Create pipe for bidirectional communication
		clientConn, serverConn := net.Pipe()
		defer clientConn.Close()
		defer serverConn.Close()

		// Create chunk connections
		clientChunkConn := chunk.NewConn(clientConn, chunk.DefaultOption)
		serverChunkConn := chunk.NewConn(serverConn, chunk.DefaultOption)

		// Run handshakes concurrently
		clientDone := make(chan error, 1)
		serverDone := make(chan error, 1)

		// Start client handshake
		go func() {
			clientDone <- HandshakeClient(clientChunkConn)
		}()

		// Start server handshake
		go func() {
			serverDone <- HandshakeServer(serverChunkConn)
		}()

		// Wait for both to complete
		select {
		case err := <-clientDone:
			require.NoError(t, err, "client handshake failed")
		case <-time.After(5 * time.Second):
			t.Fatal("client handshake timeout")
		}

		select {
		case err := <-serverDone:
			require.NoError(t, err, "server handshake failed")
		case <-time.After(5 * time.Second):
			t.Fatal("server handshake timeout")
		}
	})
}