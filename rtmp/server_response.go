package core

import (
	"github.com/moggle-mog/goav/amf"
	"github.com/pkg/errors"
)

// 服务器发往客户端的命令
func (s *ConnServer) rspGetStreamLength(cs *ChunkStream) error {

	transactionID := s.transactionID
	duration := s.duration

	return s.sendCmdMsg(cs.csid, cs.streamID, respResult, transactionID, nil, duration)
}

// 服务器发往客户端的命令
func (s *ConnServer) rspCreateStream(cs *ChunkStream) error {

	// 响应所属的命令 ID
	transactionID := s.transactionID

	// 返回值是一个流 ID 或者一个错误信息对象
	streamID := s.streamID

	return s.sendCmdMsg(cs.csid, cs.streamID, respResult, transactionID, nil, streamID)
}

// 服务端响应客户端的publish请求
func (s *ConnServer) rspPublish(cs *ChunkStream) error {

	// 响应所属的命令 ID
	transactionID := 0

	event := make(amf.Object)
	event["level"] = levelStatus               // 消息的等级取 "warning" 或 "status" 或 "error"
	event["code"] = codePublishStart           // 消息代码
	event["description"] = "Start publishing." // 消息的自然语言的描述, 这个 Info Object 字段可能(MAY)包含其它适当的属性代码

	return s.sendCmdMsg(cs.csid, cs.streamID, onStatus, transactionID, nil, event)
}

// 服务端响应客户端的play请求
func (s *ConnServer) rspPlay(cs *ChunkStream) error {

	err := s.conn.SetBegin(1)
	if err != nil {
		return errors.Wrap(err, "set begin failed")
	}

	// 发送 start 命令开始一个流
	event := make(amf.Object)
	event["level"] = levelStatus
	event["code"] = codePlayStart
	event["description"] = "Started playing stream."

	err = s.sendCmdMsg(cs.csid, cs.streamID, onStatus, 0, nil, event)
	if err != nil {
		return errors.Wrap(err, "send command message failed")
	}

	return nil
}

// 服务端将处理结果回应给客户端
func (s *ConnServer) rspConnect(cs *ChunkStream) error {

	// 响应所属的命令 ID
	transactionID := s.transactionID

	c := s.conn.NewWindowAckSize(s.conn.windowAckSize)
	if err := s.conn.Write(c); err != nil {
		return err
	}

	c = s.conn.NewSetPeerBandwidth(s.conn.remoteChunkSize)
	if err := s.conn.Write(c); err != nil {
		return err
	}

	c = s.conn.NewSetChunkSize(s.conn.chunkSize)
	if err := s.conn.Write(c); err != nil {
		return err
	}

	resp := make(amf.Object)
	resp["fmsVer"] = "FMS/3,0,1,123"
	resp["capabilities"] = 31

	event := make(amf.Object)
	event["level"] = levelStatus
	event["code"] = codeConnectSuccess
	event["description"] = "Connection succeeded."
	event["objectEncoding"] = s.connect.ObjectEncoding

	return s.sendCmdMsg(cs.csid, cs.streamID, respResult, transactionID, resp, event)
}
