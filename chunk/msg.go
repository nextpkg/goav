package chunk

import (
	"encoding/binary"
)

const (
	_ = iota
	IDSetChunkSize
	IDAbortMessage
	IDAck
	IDUser
	IDWindowAckSize
	IDSetPeerBandwidth
)

const (
	StreamBegin        = iota // 流开始, 服务端发送该事件,用来通知客户端一个流已经可以用来通讯了
	StreamEOF                 // 流结束, 服务端发送该事件,用来通知客户端其在流中请求的回放数据已经结束了
	StreamDry                 // 流枯竭, 服务端发送该事件,用来通知客户端流中已经没有更多的数据了
	StreamSetBufferLen        // 设置缓冲区大小,客户端发送该事件,用来告知服务端用来缓存流中数据的缓冲区大小(单位毫秒)
	StreamIsRecorded          // 流已录制, 服务端发送该事件,用来通知客户端指定流是一个录制流
	_                         //
	PingRequest               // ping请求, 服务端发送该事件,用来探测客户端是否处于可达状态
	PingResponse              // ping响应, 客户端用该事件回复服务端的ping请求,事件数据为收到的ping请求中携带的4字节的时间戳
)

// MakeControlMsg 协议控制消息构造器
// csid=2
func MakeControlMsg(id, size, value uint32) *ChunkStream {
	cs := &ChunkStream{
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

// NewSetChunkSize 构造设置块大小消息(1#)
func NewSetChunkSize(size uint32) *ChunkStream {
	return MakeControlMsg(IDSetChunkSize, 4, size)
}

// NewAbort 构造中止消息(2#)
func NewAbort(csid uint32) *ChunkStream {
	return MakeControlMsg(IDAbortMessage, 4, csid)
}

// NewAck 构造确认消息(3#)
func NewAck(value uint32) *ChunkStream {
	return MakeControlMsg(IDAck, 4, value)
}

// NewWindowAckSize 视窗大小确认(5#)
func NewWindowAckSize(size uint32) *ChunkStream {
	return MakeControlMsg(IDWindowAckSize, 4, size)
}

// NewSetPeerBandwidth 设置对等带宽(6#)
func NewSetPeerBandwidth(size uint32) *ChunkStream {
	// header
	ret := MakeControlMsg(IDSetPeerBandwidth, 5, size)

	// body
	ret.Data[4] = 2
	return ret
}

// MakeUserControlMsg 用户控制消息(4#)
// csid: 2
func MakeUserControlMsg(eventType, bufLen uint32) *ChunkStream {
	bufLen += 2
	ret := &ChunkStream{
		Format:   0,
		Csid:     2,
		TypeID:   4, // RTMP协议将消息类型4作为用户控制消息ID
		StreamID: 1,
		Length:   bufLen,
		Data:     make([]byte, bufLen),
	}

	ret.Data[0] = byte(eventType >> 8 & 0xff)
	ret.Data[1] = byte(eventType & 0xff)
	return ret
}

// SetBegin [用户控制消息]流开始, 事件的数据使用4个字节来表示可用的流的ID
func SetBegin(streamID uint32) *ChunkStream {
	// header
	ret := MakeUserControlMsg(StreamBegin, 4)
	// body，跳过2个字节的eventType
	binary.BigEndian.PutUint32(ret.Data[2:], streamID)
	return ret
}

// SetEOF [用户控制消息]流结束, 事件数据使用4个字节来表示回放完成的流的ID
func SetEOF(streamID uint32) *ChunkStream {
	// header
	ret := MakeUserControlMsg(StreamEOF, 4)
	// body，跳过2个字节的eventType
	binary.BigEndian.PutUint32(ret.Data[2:], streamID)
	return ret
}

// SetDry [用户控制消息]流枯竭, 事件数据用4个字节来表示枯竭的流的ID
func SetDry(streamID uint32) *ChunkStream {
	// header
	ret := MakeUserControlMsg(StreamDry, 4)
	// body，跳过2个字节的eventType
	binary.BigEndian.PutUint32(ret.Data[2:], streamID)
	return ret
}

// SetBufferLen [用户控制消息]设置缓冲区大小, 事件数据中, 前4个字节用来表示流ID, 之后的4个字节用来表示缓冲区大小(单位毫秒)
func SetBufferLen(streamID, bufInMs uint32) *ChunkStream {
	// header
	ret := MakeUserControlMsg(StreamSetBufferLen, 8)

	// body，跳过2个字节的eventType
	binary.BigEndian.PutUint32(ret.Data[2:], streamID)
	binary.BigEndian.PutUint32(ret.Data[6:], bufInMs)
	return ret
}

// SetRecorded [用户控制消息]流已录制, 事件数据用4个字节表示录制流的ID
func SetRecorded(streamID uint32) *ChunkStream {
	// header
	ret := MakeUserControlMsg(StreamIsRecorded, 4)

	// body，跳过2个字节的eventType
	binary.BigEndian.PutUint32(ret.Data[2:], streamID)
	return ret
}

// SetPingRequest [用户控制消息]ping请求, 事件数据是一个4字节的时间戳，表示服务端分发该事件时的服务器本地时间
func SetPingRequest(timestamp uint32) *ChunkStream {
	// header
	ret := MakeUserControlMsg(PingRequest, 4)

	// body，跳过2个字节的eventType
	binary.BigEndian.PutUint32(ret.Data[2:], timestamp)
	return ret
}

// SetPingResponse [用户控制消息]ping响应, 事件数据为收到的ping请求中携带的4字节的时间戳
func SetPingResponse(timestamp uint32) *ChunkStream {
	// header
	ret := MakeUserControlMsg(PingResponse, 4)

	// body，跳过2个字节的eventType
	binary.BigEndian.PutUint32(ret.Data[2:], timestamp)
	return ret
}
