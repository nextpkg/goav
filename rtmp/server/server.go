// Package server 服务端: 用于响应客户端的连接请求(但不负责传输媒体数据)
package server

import (
	"github.com/nextpkg/goav/chunk"
	"github.com/nextpkg/goav/rtmp/comm"
	"github.com/pkg/errors"
)

// ConnServer RTMP服务端
type ConnServer struct {
	Done          bool   // 标志是否完成指令消息读取的流程
	method        string // 控制指令, "play", "publish"等
	duration      int    // 流的长度，直播=0，点播>0，单位：毫秒
	StreamID      uint32
	transactionID uint32
	Conn          *chunk.Conn
	connect       comm.ConnectInfo
	publish       comm.PublishInfo
}

// NewConnServer RTMP服务端
func NewConnServer(conn *chunk.Conn, chunkSize uint32) *ConnServer {
	// 针对server调整默认的chunkSize
	conn.SetChunkSize(chunkSize)
	if conn.GetChunkSize() < 128 {
		panic("chunk size < 128")
	}

	return &ConnServer{
		StreamID: 1,
		Conn:     conn,
	}
}

// GetInfo 返回Info的字段值
func (s *ConnServer) GetInfo() (string, string) {
	return s.connect.App, s.publish.Name
}

// GetPublish 获取发布流的信息
func (s *ConnServer) GetPublish() *comm.PublishInfo {
	return &s.publish
}

// GetConnect 获取RTMP连接的信息
func (s *ConnServer) GetConnect() *comm.ConnectInfo {
	return &s.connect
}

// Method RTMP控制指令
func (s *ConnServer) Method() string {
	return s.method
}

// ===================================

// Write 向客户端写数据
func (s *ConnServer) Write(cs *chunk.ChunkStream) error {
	// 接收到chunk后，对metadata进行处理
	err := cs.Unpack()
	if err != nil {
		return errors.Wrap(err, "handle metadata failed")
	}

	return s.Conn.Write(cs)
}

// Read 从客户端读数据
func (s *ConnServer) Read(cs *chunk.ChunkStream) error {
	for {
		err := s.Conn.Read(cs)
		if err != nil {
			return err
		}

		switch cs.TypeID {
		case 15, 18:
			err = s.handleDataMsg(cs)
			if err != nil {
				return errors.Wrap(err, "handle data message failed")
			}
		case 20:
			err = s.handleCommandMsg(cs)
			if err != nil {
				return errors.Wrap(err, "handle command message failed")
			}

			continue
		}

		return nil
	}
}

// Close 关闭服务端
func (s *ConnServer) Close() error {
	return s.Conn.Close()
}

// Flush 强制清空缓冲
func (s *ConnServer) Flush() error {
	return s.Conn.Flush()
}

// ===================================

// CommandLinkup 控制指令的读与响应。读取消息直到客户端请求开始播放或者发布，一个实例只需要被执行一次
func (s *ConnServer) CommandLinkup() error {

	for {
		cs := &chunk.ChunkStream{}

		err := s.Conn.Read(cs)
		if err != nil {
			return err
		}

		err = s.handleCommandMsg(cs)
		if err != nil {
			return errors.Wrap(err, "handle command message failed")
		}

		// 已经publish或者play则退出循环
		if s.Done {
			break
		}
	}

	return nil
}
