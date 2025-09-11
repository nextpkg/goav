package server

import (
	"github.com/nextpkg/goav/amf"
	"github.com/nextpkg/goav/chunk"
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

// RespPlay responds to client's play request according to RTMP protocol
func (s *ConnServer) RespPlay(cs *chunk.ChunkStream) error {
	// Send Stream Begin control message first (RTMP protocol requirement)
	beginMsg := chunk.SetBegin(cs.StreamID)
	err := s.Conn.Write(beginMsg)
	if err != nil {
		return errors.Wrap(err, "send stream begin message failed")
	}

	// Send onStatus message to notify stream start
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

// RespConnect responds to client's connect request and sends control messages
func (s *ConnServer) RespConnect(cs *chunk.ChunkStream) error {
	// 响应所属的命令 ID
	transactionID := s.transactionID

	// Send Window Acknowledgement Size control message
	winAckMsg := chunk.NewWindowAckSize(s.Conn.GetWindowAckSize())
	err := s.Conn.Write(winAckMsg)
	if err != nil {
		return errors.Wrap(err, "send window ack size message failed")
	}

	// Send Set Peer Bandwidth control message
	peerBandwidthMsg := chunk.NewSetPeerBandwidth(s.Conn.GetRemoteChunkSize())
	err = s.Conn.Write(peerBandwidthMsg)
	if err != nil {
		return errors.Wrap(err, "send peer bandwidth message failed")
	}

	// Send Set Chunk Size control message
	chunkSizeMsg := chunk.NewSetChunkSize(s.Conn.GetChunkSize())
	err = s.Conn.Write(chunkSizeMsg)
	if err != nil {
		return errors.Wrap(err, "send chunk size message failed")
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
