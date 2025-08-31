// Package message rtmp 大部分指令由客户端使用，服务端需要实现响应过程
package message

import (
	"encoding/binary"

	"github.com/nextpkg/goav/rtmp/chunk"
)

const (
	_ = iota
	idSetChunkSize
	idAbortMessage
	idAck
	idUser
	idWindowAckSize
	idSetPeerBandwidth
)

// NewSetChunkSize 构造设置块大小消息(1#)
func (c *Conn) NewSetChunkSize(size uint32) *chunk.ChunkStream {
	return c.MakeControlMsg(idSetChunkSize, 4, size)
}

// NewAbort 构造中止消息(2#)
func (c *Conn) NewAbort(csid uint32) *chunk.ChunkStream {
	return c.MakeControlMsg(idAbortMessage, 4, csid)
}

// NewAck 构造确认消息(3#)
func (c *Conn) NewAck(value uint32) *chunk.ChunkStream {
	return c.MakeControlMsg(idAck, 4, value)
}

// NewWindowAckSize 视窗大小确认(5#)
func (c *Conn) NewWindowAckSize(size uint32) *chunk.ChunkStream {
	return c.MakeControlMsg(idWindowAckSize, 4, size)
}

// NewSetPeerBandwidth 设置对等带宽(6#)
func (c *Conn) NewSetPeerBandwidth(size uint32) *chunk.ChunkStream {

	// header
	ret := c.MakeControlMsg(idSetPeerBandwidth, 5, size)

	// body
	ret.Data[4] = 2

	return ret
}

// MakeControlMsg 协议控制消息构造器
// csid=2
func (c *Conn) MakeControlMsg(id, size, value uint32) *chunk.ChunkStream {
	cs := &chunk.ChunkStream{
		Format:   0,
		Csid:     2,
		TypeID:   id, // 可选值: 1，2，3，5和6
		StreamID: 0,
		Length:   size,
		Data:     make([]byte, size),
	}

	binary.BigEndian.PutUint32(cs.Data[:size], value)

	return cs
}
