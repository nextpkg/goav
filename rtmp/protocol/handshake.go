// Package protocol implements RTMP protocol layer functionality.
//
// This package provides RTMP protocol-level operations, primarily focusing on
// the handshake process that establishes secure connections between RTMP clients and servers.
//
// Key features:
//   - Client and server handshake implementations
//   - Support for both simple and complex handshake modes
//   - Cryptographic digest validation for secure connections
//   - Unified interfaces for protocol operations
//
// The handshake process follows the RTMP specification:
//   1. Client sends C0+C1 to server
//   2. Server responds with S0+S1+S2
//   3. Client sends C2 to complete handshake
//
// Example usage:
//
//	err := protocol.HandshakeClient(conn)
//	if err != nil {
//		// handle handshake error
//	}
//
// This package was refactored from the chunk package to provide better separation
// of concerns and improved maintainability.
package protocol

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log/slog"
	"time"

	"github.com/nextpkg/goav/chunk"
	"github.com/pkg/errors"
)

// handshakeTimeout defines the timeout for handshake operations
const handshakeTimeout = 2 * time.Second

// HandshakeClient performs client-side RTMP handshake with simple handshake only
func HandshakeClient(conn *chunk.Conn) error {
	err := conn.SetDeadline(time.Now().Add(handshakeTimeout))
	if err != nil {
		return errors.Wrap(err, "set deadline failed")
	}
	defer resetTimer(conn)

	var random [(1 + 1536*2) * 2]byte
	C0C1C2 := random[:1536*2+1]
	S0S1S2 := random[1536*2+1:]

	C0 := C0C1C2[:1]
	C1 := C0C1C2[1 : 1536+1]
	// C2 := C0C1C2[1536+1:]

	S0 := S0S1S2[:1]
	S1 := S0S1S2[1 : 1536+1]

	// 构造C0C1
	C0[0] = 3                              // 客户端版本
	binary.BigEndian.PutUint32(C1[0:4], 0) // 这个字段包含了一个时间戳，它可能(SHOULD)作为所有本端的后续块的起始时间点
	binary.BigEndian.PutUint32(C1[4:8], 0) // 简单握手, 这个字段必需是全0

	// 这个字段包含了任意值, 应该(SHOULD)足够的随机, 也不必是加密的随机数，或者是动态数据
	var n int
	n, err = rand.Read(C1[8:])
	if err != nil {
		return errors.Wrapf(err, "get random number failed,read=%d", n)
	}

	// -> C0C1
	C0C1 := C0C1C2[:1536+1]
	if err := writeHandshakeData(conn, C0C1); err != nil {
		return errors.Wrap(err, "send C0C1")
	}

	// <- S0S1S2
	if err := readHandshakeData(conn, S0S1S2); err != nil {
		return errors.Wrap(err, "receive S0S1S2")
	}

	if S0[0] != 3 {
		return fmt.Errorf("expected rtmp version 3, but got %d", S0[0])
	}

	// 简单握手 -> C2
	C2 := S1
	if err := writeHandshakeData(conn, C2); err != nil {
		return errors.Wrap(err, "send C2")
	}

	slog.Debug("client handshake is finished")
	return nil
}

// HandshakeServer performs server-side RTMP handshake, supports both complex and simple handshake
func HandshakeServer(conn *chunk.Conn) error {
	err := conn.SetDeadline(time.Now().Add(handshakeTimeout))
	if err != nil {
		return errors.Wrap(err, "set deadline")
	}
	defer resetTimer(conn)

	var random [(1 + 1536*2) * 2]byte
	C0C1C2 := random[:1536*2+1]
	S0S1S2 := random[1536*2+1:]

	C0 := C0C1C2[:1]
	C1 := C0C1C2[1 : 1536+1]
	C2 := C0C1C2[1536+1:]

	S0 := S0S1S2[:1]
	S1 := S0S1S2[1 : 1536+1]
	S2 := S0S1S2[1536+1:]

	// <- C0C1
	C0C1 := C0C1C2[:1536+1]
	err = readHandshakeData(conn, C0C1)
	if err != nil {
		return errors.Wrap(err, "read C0C1")
	}

	if C0[0] != 3 {
		return fmt.Errorf("invalid rtmp version=%d ", C0[0])
	}

	// 构造S0S1S2
	S0[0] = 3
	cTime := binary.BigEndian.Uint32(C1[0:4])    // 这个字段包含了一个时间戳，它可能(SHOULD)作为所有本端的后续块的起始时间点
	cVersion := binary.BigEndian.Uint32(C1[4:8]) // 0=简单握手, 其它=复杂握手
	if cVersion != 0 {
		// 复杂握手
		sTime := cTime
		sVersion := uint32(0x04050001)

		// 用client key计算C1的摘要, 再用server key计算C1摘要的摘要并返回
		digest, err := getDigest(C1, clientPartialKey, serverFullKey)
		if err != nil {
			return errors.Wrap(err, "calculate C1 digest")
		}

		// 计算S1的签名, 填充到S1中
		err = createDigestC1S1(S1, sTime, sVersion, serverPartialKey)
		if err != nil {
			return errors.Wrap(err, "calculate C1S1 digest")
		}

		// 计算S2的签名, 填充到S2中
		err = createDigestC2S2(S2, digest)
		if err != nil {
			return errors.Wrap(err, "calculate C2S2 digest")
		}
	} else {
		// 简单握手
		copy(S1, C2)
		copy(S2, C1)
	}

	// -> S0S1S2
	if err := writeHandshakeData(conn, S0S1S2); err != nil {
		return errors.Wrap(err, "send S0S1S2")
	}

	// <- C2
	if err := readHandshakeData(conn, C2); err != nil {
		return errors.Wrap(err, "receive C2")
	}

	slog.Debug("server handshake is finished")
	return nil
}

// writeHandshakeData writes handshake data to connection
func writeHandshakeData(conn *chunk.Conn, data []byte) error {
	// 将协议数据写入网络缓冲中
	n, err := conn.Rw.Write(data)
	if err != nil {
		return errors.Wrapf(err, "write failed,wrote=%d", n)
	}

	// 使缓冲的数据写入网络通道
	err = conn.Rw.Flush()
	if err != nil {
		return errors.Wrap(err, "flush failed")
	}

	return nil
}

// readHandshakeData reads handshake data from connection
func readHandshakeData(conn *chunk.Conn, data []byte) error {
	// 从网络缓冲中读取对端数据
	n, err := conn.Rw.Read(data)
	if err != nil {
		return errors.Wrapf(err, "read failed,read=%d", n)
	}

	return nil
}

// resetTimer resets the connection deadline
func resetTimer(conn *chunk.Conn) {
	// reset deadline for client handshake
	err := conn.SetDeadline(time.Time{})
	if err != nil {
		slog.Error(errors.Wrap(err, "reset connection timer failed").Error())
	}
}