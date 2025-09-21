// Package core rtmp 大部分指令由服务端使用，客户端需要实现响应过程
package core

import (
	"encoding/binary"
)

const (
	streamBegin      = iota // 流开始, 服务端发送该事件,用来通知客户端一个流已经可以用来通讯了
	streamEOF               // 流结束, 服务端发送该事件,用来通知客户端其在流中请求的回放数据已经结束了
	streamDry               // 流枯竭, 服务端发送该事件,用来通知客户端流中已经没有更多的数据了
	setBufferLen            // 设置缓冲区大小,客户端发送该事件,用来告知服务端用来缓存流中数据的缓冲区大小(单位毫秒)
	streamIsRecorded        // 流已录制, 服务端发送该事件,用来通知客户端指定流是一个录制流
	_                       //
	pingRequest             // ping请求, 服务端发送该事件,用来探测客户端是否处于可达状态
	pingResponse            // ping响应, 客户端用该事件回复服务端的ping请求,事件数据为收到的ping请求中携带的4字节的时间戳
)

// SetBegin [用户控制消息]流开始, 事件的数据使用4个字节来表示可用的流的ID
func (c *Conn) SetBegin(streamID uint32) error {
	// header
	ret := c.MakeUserControlMsg(streamBegin, 4)

	// body，跳过2个字节的eventType
	binary.BigEndian.PutUint32(ret.data[2:], streamID)

	// write函数的封装，错误直接透传
	return c.Write(ret)
}

// SetEOF [用户控制消息]流结束, 事件数据使用4个字节来表示回放完成的流的ID
func (c *Conn) SetEOF(streamID uint32) error {
	// header
	ret := c.MakeUserControlMsg(streamEOF, 4)

	// body，跳过2个字节的eventType
	binary.BigEndian.PutUint32(ret.data[2:], streamID)

	// write函数的封装，错误直接透传
	return c.Write(ret)
}

// SetDry [用户控制消息]流枯竭, 事件数据用4个字节来表示枯竭的流的ID
func (c *Conn) SetDry(streamID uint32) error {
	// header
	ret := c.MakeUserControlMsg(streamDry, 4)

	// body，跳过2个字节的eventType
	binary.BigEndian.PutUint32(ret.data[2:], streamID)

	// write函数的封装，错误直接透传
	return c.Write(ret)

}

// SetBufferLen [用户控制消息]设置缓冲区大小, 事件数据中, 前4个字节用来表示流ID, 之后的4个字节用来表示缓冲区大小(单位毫秒)
func (c *Conn) SetBufferLen(streamID, bufInMs uint32) error {
	// header
	ret := c.MakeUserControlMsg(setBufferLen, 8)

	// body，跳过2个字节的eventType
	binary.BigEndian.PutUint32(ret.data[2:], streamID)
	binary.BigEndian.PutUint32(ret.data[6:], bufInMs)

	// write函数的封装，错误直接透传
	return c.Write(ret)
}

// SetRecorded [用户控制消息]流已录制, 事件数据用4个字节表示录制流的ID
func (c *Conn) SetRecorded(streamID uint32) error {
	// header
	ret := c.MakeUserControlMsg(streamIsRecorded, 4)

	// body，跳过2个字节的eventType
	binary.BigEndian.PutUint32(ret.data[2:], streamID)

	// write函数的封装，错误直接透传
	return c.Write(ret)
}

// SetPingRequest [用户控制消息]ping请求, 事件数据是一个4字节的时间戳，表示服务端分发该事件时的服务器本地时间
func (c *Conn) SetPingRequest(timestamp uint32) error {
	// header
	ret := c.MakeUserControlMsg(pingRequest, 4)

	// body，跳过2个字节的eventType
	binary.BigEndian.PutUint32(ret.data[2:], timestamp)

	// write函数的封装，错误直接透传
	return c.Write(ret)
}

// SetPingResponse [用户控制消息]ping响应, 事件数据为收到的ping请求中携带的4字节的时间戳
func (c *Conn) SetPingResponse(timestamp uint32) error {
	// header
	ret := c.MakeUserControlMsg(pingResponse, 4)

	// body，跳过2个字节的eventType
	binary.BigEndian.PutUint32(ret.data[2:], timestamp)

	// write函数的封装，错误直接透传
	return c.Write(ret)
}

// MakeUserControlMsg 用户控制消息(4#)
// csid: 2
func (c *Conn) MakeUserControlMsg(eventType, bufLen uint32) *ChunkStream {
	bufLen += 2

	ret := &ChunkStream{
		format:   0,
		csid:     2,
		typeID:   4, // RTMP协议将消息类型4作为用户控制消息ID
		streamID: 1,
		length:   bufLen,
		data:     make([]byte, bufLen),
	}

	ret.data[0] = byte(eventType >> 8 & 0xff)
	ret.data[1] = byte(eventType & 0xff)

	return ret
}
