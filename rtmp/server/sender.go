package server

import (
	"bytes"

	"github.com/nextpkg/goav/amf"
	"github.com/nextpkg/goav/rtmp/chunk"
	"github.com/pkg/errors"
)

// 构造命令消息
func (s *ConnServer) sendCmdMsg(csid, streamID uint32, args ...interface{}) error {
	cs := &chunk.ChunkStream{
		Format:    0,
		Csid:      csid, // 命令消息所在通道号
		Timestamp: 0,
		TypeID:    20, // 消息类型20代表AMF0编码, 消息类型17代表AMF3编码
		StreamID:  streamID,
	}

	return s.sendMsg(cs, args...)
}

// 构造数据消息
func (s *ConnServer) sendDataMsg(csid, streamID uint32, args ...interface{}) error {
	cs := &chunk.ChunkStream{
		Format:    0,
		Csid:      csid, // 数据消息所在通道号
		Timestamp: 0,
		TypeID:    18, // 消息类型18代表AMF0编码, 消息类型15代表AMF3编码
		StreamID:  streamID,
	}

	return s.sendMsg(cs, args...)
}

// 使用AMF0格式发送命令消息
func (s *ConnServer) sendMsg(cs *chunk.ChunkStream, args ...interface{}) error {
	command := bytes.NewBuffer(nil)

	// 使用AMF0编码
	err := amf.NewEnDecAMF0().EncodeBatch(command, args...)
	if err != nil {
		return errors.Wrap(err, "amf0 batch encode failed")
	}

	if command.Len() == 0 {
		return errors.New("args are useless")
	}

	msg := command.Bytes()

	// 填充消息内容
	cs.Length = uint32(len(msg))
	cs.Data = msg

	err = s.Conn.Write(cs)
	if err != nil {
		return errors.Wrap(err, "write message failed")
	}

	return s.Conn.Flush()
}
