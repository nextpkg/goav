package server

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/nextpkg/goav/amf"
	"github.com/nextpkg/goav/chunk"
	"github.com/nextpkg/goav/rtmp/comm"
	"github.com/pkg/errors"
)

func (s *ConnServer) connectMsg(cs *chunk.ChunkStream, cmd []interface{}) error {

	if len(cmd) <= 1 {
		return fmt.Errorf("incomplete connect command,length=%d", len(cmd))
	}

	// 解析命令
	err := s.handleConnect(cmd[1:])
	if err != nil {
		return errors.Wrap(err, "parse connect command failed")
	}

	// 响应
	err = s.RespConnect(cs)
	if err != nil {
		return errors.Wrap(err, "response connect command failed")
	}

	s.method = comm.Connect
	return nil
}

func (s *ConnServer) createStreamMsg(cs *chunk.ChunkStream, cmd []interface{}) error {

	if len(cmd) <= 1 {
		return fmt.Errorf("incomplete createStream command,length=%d", len(cmd))
	}

	// 解析命令
	err := s.handleCreateStream(cmd[1:])
	if err != nil {
		return errors.Wrap(err, "parse 'CreateStream' failed")
	}

	// 响应
	err = s.RespCreateStream(cs)
	if err != nil {
		return errors.Wrap(err, "response createStream command failed")
	}

	s.method = comm.CreateStream
	return nil
}

func (s *ConnServer) publishMsg(cs *chunk.ChunkStream, cmd []interface{}) error {

	if len(cmd) <= 1 {
		return fmt.Errorf("incomplete publish command,length=%d", len(cmd))
	}

	// 解析命令
	err := s.handlePublish(cmd[1:])
	if err != nil {
		return errors.Wrap(err, "parse publish command failed")
	}

	// 响应
	err = s.RespPublish(cs)
	if err != nil {
		return errors.Wrap(err, "response publish command failed")
	}

	s.Done = true
	s.method = comm.Publish

	return nil
}

func (s *ConnServer) playMsg(cs *chunk.ChunkStream, cmd []interface{}) error {

	if len(cmd) <= 1 {
		return fmt.Errorf("incomplete play command,length=%d", len(cmd))
	}

	// 解析命令
	err := s.handlePlay(cmd[1:])
	if err != nil {
		return errors.Wrap(err, "parse play command failed")
	}

	// 响应
	err = s.RespPlay(cs)
	if err != nil {
		return errors.Wrap(err, "response play command failed")
	}

	s.Done = true
	s.method = comm.Play

	return nil
}

func (s *ConnServer) getStreamLengthMsg(cs *chunk.ChunkStream, cmd []interface{}) error {

	if len(cmd) <= 1 {
		return fmt.Errorf("incomplete getStreamLength command,length=%d", len(cmd))
	}

	// 直播流不需要获取长度
	err := s.handleGetStreamLength(cmd[1:])
	if err != nil {
		return errors.Wrap(err, "get stream length failed")
	}

	err = s.rspGetStreamLength(cs)
	if err != nil {
		return errors.Wrap(err, "response getStreamLength command failed")
	}

	s.method = comm.GetStreamLength

	return nil
}

func (s *ConnServer) deleteStreamMsg(cmd []interface{}) error {

	if len(cmd) <= 1 {
		return errors.New("incomplete deleteStream command")
	}

	err := s.handleDeleteStream(cmd[4:])
	if err != nil {
		return errors.Wrap(err, "parse deleteStream command failed")
	}

	s.StreamID = 0
	s.method = comm.DeleteStream

	return nil
}

func (s *ConnServer) fcUnPublishMsg(cmd []interface{}) error {
	if len(cmd) <= 1 {
		return fmt.Errorf("incomplete FCUnpublish command,length=%d", len(cmd))
	}

	err := s.handleFcunpublish(cmd[1:])
	if err != nil {
		return errors.Wrap(err, "parse FCUnpublish command failed")
	}

	// 协议为指定该如何处理这类消息，以下是自定义处理过程
	s.publish = comm.PublishInfo{}
	s.method = comm.FCUnpublish

	return nil
}

// 处理命令消息，如果不是命令消息则返回nil
func (s *ConnServer) handleCommandMsg(cs *chunk.ChunkStream) error {
	// 只支持AMF0, 如果出现AMF3, 只需要把第一个字节去掉即是AMF0
	switch cs.TypeID {
	case 17:
		// AMF3
		if len(cs.Data) <= 1 {
			return errors.New("incomplete chunk data")
		}
		cs.Data = cs.Data[1:]
	case 20:
	default:
		return nil
	}

	// 按协议版本解析命令消息
	cmd, err := amf.NewEnDecAMF0().DecodeBatch(bytes.NewReader(cs.Data))
	if err != nil {
		return errors.Wrap(err, "amf0 batch decode failed")
	}

	if len(cmd) == 0 {
		return errors.New("empty command message")
	}

	// 命令名称, 只有string类型
	cmdName, ok := cmd[0].(string)
	if !ok {
		return errors.New("invalid command name format")
	}

	slog.Debug("server handle command", "name", cmdName)

	switch cmdName {
	case comm.Connect:
		return s.connectMsg(cs, cmd)
	case comm.CreateStream:
		return s.createStreamMsg(cs, cmd)
	case comm.Publish:
		return s.publishMsg(cs, cmd)
	case comm.Play:
		return s.playMsg(cs, cmd)
	case comm.ReleaseStream:
		s.transactionID = 0
		s.method = comm.ReleaseStream
		return nil
	case comm.FcPublish:
		s.publish = comm.PublishInfo{}
		s.method = comm.FcPublish
		return nil
	case comm.GetStreamLength:
		return s.getStreamLengthMsg(cs, cmd)
	case comm.DeleteStream:
		return s.deleteStreamMsg(cmd)
	case comm.FCUnpublish:
		return s.fcUnPublishMsg(cmd)
	default:
		// 遇到未实现的命令不需要抛出错误从而使连接断开，提高服务稳定性
		slog.Error("unrealized command", "message", cmdName)
		return nil
	}
}

// 处理数据消息，如果消息类型不匹配，则返回nil，TypeID=<15,18>
func (s *ConnServer) handleDataMsg(cs *chunk.ChunkStream) error {
	var err error
	var cmd amf.Array

	switch cs.TypeID {
	case 18:
		cmd, err = amf.NewEnDecAMF0().DecodeBatch(bytes.NewReader(cs.Data))
		if err != nil {
			return errors.Wrap(err, "amf0 batch decode failed")
		}
	case 15:
		cmd, err = amf.NewEnDecAMF3().DecodeBatch(bytes.NewReader(cs.Data))
		if err != nil {
			return errors.Wrap(err, "amf3 batch decode failed")
		}
	default:
		return nil
	}

	if len(cmd) == 0 {
		return errors.New("empty command message")
	}

	// 命令名称, 只有string类型
	cmdName, ok := cmd[0].(string)
	if !ok {
		return fmt.Errorf("invalid command name format")
	}

	// 格式: [SetDataFrame, OnMetaData, data]
	switch cmdName {
	case amf.SetDataFrame:
		if len(cmd) <= 1 {
			return errors.New("incomplete SetDataFrame command")
		}

		err = s.handleSetDataFrame(cmd[1:])
		if err != nil {
			return errors.Wrap(err, "set data frame failed")
		}
	}

	return nil
}

func (s *ConnServer) handleSetDataFrame(args []interface{}) error {
	subCmd, ok := args[0].(string)
	if !ok {
		return errors.New("invalid sub command name format")
	}

	switch subCmd {
	case amf.OnMetaData:
		if len(args) <= 1 {
			return errors.New("invalid OnMetaData")
		}

		object, ok := args[1].(amf.Object)
		if !ok {
			return errors.New("invalid object")
		}

		s.publish.MetaData = object
	}

	return nil
}
