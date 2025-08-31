package client

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/nextpkg/goav/amf"
	"github.com/nextpkg/goav/rtmp/ce"
	"github.com/nextpkg/goav/rtmp/chunk"
	"github.com/nextpkg/goav/rtmp/comm"
	"github.com/nextpkg/goav/rtmp/message"
	"github.com/pkg/errors"
)

// ConnClient RTMP客户端
type ConnClient struct {
	connect       comm.ConnectInfo // 客户端参数
	publish       comm.PublishInfo // 客户端参数
	current       string           // 通信：客户端正在执行的指令名称
	streamID      uint32           // 通信：客户端获得的流的ID
	Conn          *message.Conn    // 通信：RTMP服务
	TransactionID int              // 通信：事务ID
}

// NewConnClientByURL 使用RTMP链接获取RTMP客户端
func NewConnClientByURL(rtmpURL string, rtmpDialTimeout time.Duration) *ConnClient {
	u, err := url.Parse(rtmpURL)
	if err != nil || len(u.Host) == 0 || u.Scheme != "rtmp" {
		slog.Error("invalid rtmp url,host or protocol")
		return nil
	}

	address := u.Host
	if u.Port() == "" {
		address = u.Host + ":1935"
	}

	conn, err := net.DialTimeout("tcp", address, rtmpDialTimeout*time.Second)
	if err != nil {
		slog.Error(errors.Wrapf(err, "Dial rtmp server failed,host='%s'", address).Error())
		return nil
	}

	// path
	ps := strings.Split(strings.TrimLeft(u.Path, "/"), "/")
	if len(ps) != 2 {
		slog.Error("invalid rtmp path")
		return nil
	}

	c := message.NewConn(conn, comm.ConnBufSize)

	// RTMP服务
	return NewConnClient(ps[0], ps[1], c)
}

// NewConnClient RTMP客户端
func NewConnClient(app, instance string, conn *message.Conn) *ConnClient {
	return &ConnClient{
		Conn: conn,
		connect: comm.ConnectInfo{
			App:   app,
			TcURL: "rtmp://" + conn.LocalAddr().String() + "/" + app,
		},
		publish: comm.PublishInfo{
			Name: instance,
			Type: comm.PublishLive,
		},
		TransactionID: 1,
	}
}

// GetInfo 返回Info的字段值
func (c *ConnClient) GetInfo() (app string, instance string) {
	return c.connect.App, c.publish.Name
}

// GetPublish 获取发布流的信息
func (c *ConnClient) GetPublish() *comm.PublishInfo {
	return &c.publish
}

// GetConnect 获取RTMP连接的信息
func (c *ConnClient) GetConnect() *comm.ConnectInfo {
	return &c.connect
}

// ===================================

// Write 向服务端发送数据
func (c *ConnClient) Write(cs *chunk.ChunkStream) error {
	// 接收到chunk后，对metadata进行整理
	err := cs.Unpack()
	if err != nil {
		return errors.Wrap(err, "unpack and get data failed")
	}

	return c.Conn.Write(cs)
}

// Read 从服务端读取数据
func (c *ConnClient) Read(cs *chunk.ChunkStream) error {
	return c.Conn.Read(cs)
}

// Close 关闭连接管道
func (c *ConnClient) Close() error {
	return c.Conn.Close()
}

// Flush 强制清空缓冲
func (c *ConnClient) Flush() error {
	return c.Conn.Flush()
}

// ===================================

// Start 发起一个RTMP连接到服务端
func (c *ConnClient) Start(method string) error {
	// 1. 握手
	err := c.Conn.HandshakeClient()
	if err != nil {
		return errors.Wrap(err, "client handshake to server")
	}

	// 2. 交互 Connect
	err = c.Connect()
	if err != nil {
		return errors.Wrap(err, "client send CONNECT command to server")
	}

	// 3. 交互 CreateStream
	err = c.CreateStream()
	if err != nil {
		return errors.Wrap(err, "client send CREATE_STREAM command to server")
	}

	// 4. 开始 Play 或者 Publish
	switch method {
	case comm.Publish:
		if err := c.Publish(); err != nil {
			return errors.Wrap(err, "client send PUBLISH command to server")
		}
	case comm.Play:
		if err := c.Play(); err != nil {
			return errors.Wrap(err, "client send PLAY command to server")
		}
	default:
		return errors.New("unknown method")
	}

	return nil
}

// StartPlay 播放流
func (c *ConnClient) StartPlay() error {
	return c.Start(comm.Play)
}

// StartPublish 发布流
func (c *ConnClient) StartPublish() error {
	return c.Start(comm.Publish)
}

// Connect 发送Connect命令并接收响应
func (c *ConnClient) Connect() error {
	// always set to 1
	c.TransactionID = 1
	c.current = comm.Connect

	event := make(amf.Object)
	event["app"] = c.connect.App
	event["flashVer"] = comm.FlashVer
	event["tcUrl"] = c.connect.TcURL

	// ready to request
	slog.Debug("Connect chunk size before sending", "size", c.Conn.ChunkSize)

	err := c.sendCmdMsg(comm.Connect, c.TransactionID, event)
	if err != nil {
		return errors.Wrapf(err, "sendCmdMsg,transaction id=%d,app='%s',tcUrl='%s'",
			c.TransactionID, c.connect.App, c.connect.TcURL)
	}

	err = c.recvCmdMsg()
	if err != nil {
		return errors.Wrapf(err, "recvCmdMsg,transaction id=%d,app='%s',tcUrl='%s'",
			c.TransactionID, c.connect.App, c.connect.TcURL)
	}

	slog.Debug("Connect chunk size after sending", "size", c.Conn.ChunkSize)

	return nil
}

// CreateStream 发送CreateStream命令并接收响应
func (c *ConnClient) CreateStream() error {
	c.TransactionID++
	c.current = comm.CreateStream

	// 格式: <command name>,<transaction id>,<command object>
	err := c.sendCmdMsg(comm.CreateStream, c.TransactionID, nil)
	if err != nil {
		return errors.Wrapf(err, "sendCmdMsg,transaction id=%d", c.TransactionID)
	}

	err = c.recvCmdMsg()
	if err != nil {
		return errors.Wrapf(err, "recvCmdMsg,transaction id=%d", c.TransactionID)
	}

	return nil
}

// Publish 发布类型: live
func (c *ConnClient) Publish() error {
	c.current = comm.Publish
	c.TransactionID++

	// 格式: <command name>,<transaction ID>,<command object>,<publishing name>,<publishing type>
	err := c.sendCmdMsg(comm.Publish, 0, nil, c.publish.Name, comm.PublishLive)
	if err != nil {
		return errors.Wrapf(err, "sendCmdMsg,transaction id=%d,instance='%s'", c.TransactionID, c.publish.Name)
	}

	err = c.recvCmdMsg()
	if err != nil {
		return errors.Wrapf(err, "recvCmdMsg,transaction id=%d,instance='%s'", c.TransactionID, c.publish.Name)
	}

	return nil
}

// Play 直播
func (c *ConnClient) Play() error {
	c.current = comm.Play
	c.TransactionID++

	// 格式: <command name>,<transaction id>,<command object>,
	// <stream name>,<Start(可选) float>,<Duration(可选) float>,<Reset(可选) bool>
	err := c.sendCmdMsg(comm.Play, 0, nil, c.publish.Name)
	if err != nil {
		return errors.Wrapf(err, "sendCmdMsg,transaction id=%d,instance='%s'", c.TransactionID, c.publish.Name)
	}

	err = c.recvCmdMsg()
	if err != nil {
		return errors.Wrapf(err, "recvCmdMsg,transaction id=%d,instance='%s'", c.TransactionID, c.publish.Name)
	}

	return nil
}

func (c *ConnClient) strCmd(v interface{}) error {
	str := v.(string)

	switch c.current {
	case comm.Connect, comm.CreateStream:
		if str != comm.RespResult {
			return fmt.Errorf("connect or createStream command response result='%s'", str)
		}
	case comm.Publish:
		if str != comm.OnStatus {
			return ce.ErrRespFailed
		}
	}

	return nil
}

func (c *ConnClient) intCmd(k int, v interface{}) error {
	id := int(v.(float64))

	switch c.current {
	case comm.Connect, comm.CreateStream:
		switch k {
		case 1:
			if c.TransactionID != id {
				return ce.ErrRespFailed
			}
		case 3:
			c.streamID = uint32(id)
		}
	case comm.Publish:
		if id != 0 {
			return ce.ErrRespFailed
		}
	}

	return nil
}

func (c *ConnClient) objCmd(v interface{}) error {
	property := v.(amf.Object)

	switch c.current {
	case comm.Connect:
		// RTSP
		_, ok := property["fmsVer"]
		if ok {
			return nil
		}

		// event
		code, ok := property["code"]
		if ok {
			str := code.(string)
			if str != comm.CodeConnectSuccess {
				return ce.ErrRespFailed
			}

			return nil
		}

		return ce.ErrRespFailed
	case comm.Publish:
		// event
		code, ok := property["code"]
		if !ok {
			return ce.ErrRespFailed
		}

		str := code.(string)
		if str != comm.CodePublishStart {
			return fmt.Errorf("server return publish code='%s'", code.(string))
		}
	}

	return nil
}

// 验证服务端响应的指令消息
func (c *ConnClient) recvCmdMsg() error {
	for {
		cs := &chunk.ChunkStream{}

		err := c.Conn.Read(cs)
		if err != nil {
			return err
		}

		var cmd []interface{}

		// 只处理指令消息
		// 客户端发送这些消息来完成连接、创建流、发布、播放、暂停等操作
		// 像状态、结果这样的指令消息，用于通知发送方请求的指令状态
		switch cs.TypeID {
		case 20:
			cmd, err = amf.NewEnDecAMF0().DecodeBatch(bytes.NewReader(cs.Data))
			if err != nil {
				return errors.Wrap(err, "decode amf0")
			}
		case 17:
			cmd, err = amf.NewEnDecAMF3().DecodeBatch(bytes.NewReader(cs.Data))
			if err != nil {
				return errors.Wrap(err, "decode amf3")
			}
		default:
			continue
		}

		for k, v := range cmd {
			switch v.(type) {
			case string:
				if err := c.strCmd(v); err != nil {
					return errors.Wrap(err, "handle str command")
				}
			case float64:
				if err := c.intCmd(k, v); err != nil {
					return errors.Wrap(err, "handle int command")
				}
			case amf.Object:
				if err := c.objCmd(v); err != nil {
					return errors.Wrap(err, "handle obj command")
				}
			}
		}

		return nil
	}
}

// 构造并发送命令消息，CSID:3，消息类型: 20
func (c *ConnClient) sendCmdMsg(args ...interface{}) error {
	command := bytes.NewBuffer(nil)

	// 使用AMF0编码
	err := amf.NewEnDecAMF0().EncodeBatch(command, args...)
	if err != nil {
		return errors.Wrap(err, "encode command message")
	}

	msg := command.Bytes()

	// 构造消息
	cs := &chunk.ChunkStream{
		Format:    0,
		Csid:      3, // 命令消息所在通道号
		Timestamp: 0,
		TypeID:    20, // 消息类型20代表AMF0编码, 消息类型17代表AMF3编码
		StreamID:  c.streamID,
		Length:    uint32(len(msg)),
		Data:      msg,
	}

	if err := c.Conn.Write(cs); err != nil {
		return errors.Wrap(err, "send command message to server")
	}

	return c.Conn.Flush()
}
