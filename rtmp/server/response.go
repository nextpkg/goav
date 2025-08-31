package server

import (
	"github.com/nextpkg/goav/amf"
	"github.com/nextpkg/goav/rtmp/chunk"
	"github.com/nextpkg/goav/rtmp/comm"
	"github.com/pkg/errors"
)

// 服务器发往客户端的命令
func (s *ConnServer) rspGetStreamLength(cs *chunk.ChunkStream) error {
	transactionID := s.transactionID
	duration := s.duration

	return s.sendCmdMsg(cs.Csid, cs.StreamID, comm.RespResult, transactionID, nil, duration)
}

// RespCreateStream 服务器发往客户端的命令
func (s *ConnServer) RespCreateStream(cs *chunk.ChunkStream) error {
	// 响应所属的命令 ID
	transactionID := s.transactionID

	// 返回值是一个流 ID 或者一个错误信息对象
	streamID := s.StreamID

	return s.sendCmdMsg(cs.Csid, cs.StreamID, comm.RespResult, transactionID, nil, streamID)
}

// RespPublish 服务端响应客户端的publish请求
func (s *ConnServer) RespPublish(cs *chunk.ChunkStream) error {
	// 响应所属的命令 ID
	transactionID := 0

	event := make(amf.Object)
	event["level"] = comm.LevelStatus          // 消息的等级取 "warning" 或 "status" 或 "error"
	event["code"] = comm.CodePublishStart      // 消息代码
	event["description"] = "Start publishing." // 消息的自然语言的描述, 这个 Info Object 字段可能(MAY)包含其它适当的属性代码

	return s.sendCmdMsg(cs.Csid, cs.StreamID, comm.OnStatus, transactionID, nil, event)
}

// RespPlay 服务端响应客户端的play请求
func (s *ConnServer) RespPlay(cs *chunk.ChunkStream) error {
	err := s.Conn.SetBegin(1)
	if err != nil {
		return errors.Wrap(err, "set begin failed")
	}

	// 发送 start 命令开始一个流
	event := make(amf.Object)
	event["level"] = comm.LevelStatus
	event["code"] = comm.CodePlayStart
	event["description"] = "Started playing stream."

	err = s.sendCmdMsg(cs.Csid, cs.StreamID, comm.OnStatus, 0, nil, event)
	if err != nil {
		return errors.Wrap(err, "send command message failed")
	}

	return nil
}

// RespConnect 服务端将处理结果回应给客户端
func (s *ConnServer) RespConnect(cs *chunk.ChunkStream) error {
	// 响应所属的命令 ID
	transactionID := s.transactionID

	c := s.Conn.NewWindowAckSize(s.Conn.WindowAckSize)
	if err := s.Conn.Write(c); err != nil {
		return err
	}

	c = s.Conn.NewSetPeerBandwidth(s.Conn.RemoteChunkSize)
	if err := s.Conn.Write(c); err != nil {
		return err
	}

	c = s.Conn.NewSetChunkSize(s.Conn.ChunkSize)
	if err := s.Conn.Write(c); err != nil {
		return err
	}

	resp := make(amf.Object)
	resp["fmsVer"] = "FMS/3,0,1,123"
	resp["capabilities"] = 31

	event := make(amf.Object)
	event["level"] = comm.LevelStatus
	event["code"] = comm.CodeConnectSuccess
	event["description"] = "Connection succeeded."
	event["objectEncoding"] = s.connect.ObjectEncoding

	return s.sendCmdMsg(cs.Csid, cs.StreamID, comm.RespResult, transactionID, resp, event)
}
