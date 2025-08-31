package server

import (
	"fmt"
	"log/slog"

	"github.com/nextpkg/goav/amf"
	"github.com/nextpkg/goav/rtmp/comm"
	"github.com/pkg/errors"
)

// 客户端向服务器发送 publish 命令用来发布带有名字的流。
// 任何客户端都可以使用这个 名字播放这个流和接收发布的音频，视频和数据信息
// 格式: <命令名称: publish>,<事务ID: 0>,<命令对象: nil>,<发布流名称>,<发布流类型>
func (s *ConnServer) handlePublish(args []interface{}) error {
	for k, v := range args {
		switch v := v.(type) {
		case float64:
			s.transactionID = uint32(v)
		case amf.Object:
		case string:
			switch k {
			case 2:
				// 发布流的名称
				s.publish.Name = v
				slog.Debug("publish name:", "name", s.publish.Name)
			case 3:
				/**
				发布流的类型。设置为"live", "record", "append"
				live: 发布直播流，数据不会被录制
				record: 发布流并且数据被录制在一个新文件中。这个文件被存储在服务器应用程序的子目录中, 如果文件已经存在, 那么它将会被覆盖
				append: 发布流并且数据被追加在文件的后面，如果文件不存在，它将会被创建
				*/
				switch v {
				case comm.PublishLive, comm.PublishRecord, comm.PublishAppend:
					s.publish.Type = v
				default:
					return errors.New("invalid publish type")
				}
			}
		}
	}

	return nil
}

// 客户端发送个命令给服务器用以播放一个流。
// 格式: <命令名称: play>,<事务ID: 0>,<命令对象: nil>,<流名称: string>,<起始时间: number>,<可选回放时间段: number>
func (s *ConnServer) handlePlay(args []interface{}) error {
	for k, v := range args {
		switch v := v.(type) {
		case float64:
			switch k {
			case 0:
				// 事务的 ID 设置为 0
				s.transactionID = uint32(v)
			case 3:
				// start: 一个可选的参数,它指定了起始的时间以秒为单位
				start := int(v)

				switch {
				case start <= -2:
					/**
					该订阅者第一次尝试播放流名称字段指定的直播流。
					如果该名称的直播流没有被找到，它会播放同名的录播流。
					如果没有该名称的录播流，该订阅者等到该名称的直播流可用时再播放。
					*/
					slog.Debug("unrealized", "start", start)
				case start == -1, start == 0:
					/**
					只有流名称字段指定的直播流才会被播放。
					*/
				case start > 0:
					/**
					一个在流名称字段指定的录播流会被播放，起始时间是Start字段指定的时间。
					如果找不到该记录流，播放列表的下一个条目会被播放。
					*/
					slog.Debug("unrealized", "start", start)
				}
			case 4:
				// duration: 一个可选的参数，它指定了回放的持续时间以秒为单位
				duration := int(v)
				switch {
				case duration <= -1:
					/**
					直播流被播放直到它不可用，或者录播流被播放直到结束
					*/
				case duration == 0:
					/**
					它会播放录播流中Start字段指定开始时间的一帧，它假定Start字段的值大于等于0
					*/
					slog.Debug("unrealized", "duration", duration)
				case duration > 0:
					/**
					它会播放Duration字段指定该段时间的直播流。
					之后，它能够播放Duration字段指定该段时间的录播流。
					*/
					slog.Debug("unrealized", "duration", duration)
				}
			}
		case amf.Object:
		case string:
			// 需要播放的流的名字
			s.publish.Name = v
		case bool:
			// 一个可选的布尔值或布尔数字，指定示是否刷新之前的播放列表
			slog.Debug("unrealized", "play", "reset")
		}
	}

	return nil
}

// 获取指定视频的长度
// 格式: <命令名称: getStreamLength>,<事务ID: number>,<发布流名称>
func (s *ConnServer) handleGetStreamLength(args []interface{}) error {
	for _, v := range args {
		switch v := v.(type) {
		case float64:
			s.transactionID = uint32(v)
		case string:
			// 未实现：通过发布流名称获取流的长度，RTMP如需支持历史视频，此处应实现按流的名字查找历史视频长度
			_ = v
			s.duration = 0
		}
	}

	return nil
}

// 客户端发送这个命令去创建一个用于消息通信的逻辑通道
// 音频、视频和metadata都是通过使用 createStream 命令创建的流通道进行传输的
// 格式: <命令名称: createStream>,<事务ID: number>,<命令对象: object>
func (s *ConnServer) handleCreateStream(args []interface{}) error {
	for _, v := range args {
		switch v := v.(type) {
		case float64:
			s.transactionID = uint32(v)
		case amf.Object:
			// 【可以不实现】这里的命令对象由客户端自定义
			slog.Debug("unrealized object")
		}
	}

	return nil
}

func (s *ConnServer) connectCommandMsg(v interface{}) {
	property := v.(amf.Object)

	// 客户端连接到服务器上的应用实例名称
	app, ok := property["app"]
	if ok {
		s.connect.App = app.(string)
	}

	// Flash 播放器版本
	flashVer, ok := property["flashVer"]
	if ok {
		s.connect.FlashVer = flashVer.(string)
	}

	// 当前连接的 SWF 源文件地址
	swfURL, ok := property["swfUrl"]
	if ok {
		s.connect.SwfURL = swfURL.(string)
	}

	// 服务端的地址
	tcURL, ok := property["tcUrl"]
	if ok {
		s.connect.TcURL = tcURL.(string)
	}

	// 如果使用代理设置为 True
	fpad, ok := property["fpad"]
	if ok {
		s.connect.FPad = fpad.(bool)
	}

	// 表明客户端支持的音频编码
	audioCodecs, ok := property["audioCodecs"]
	if ok {
		s.connect.AudioCodecs = int(audioCodecs.(float64))
	}

	// 表明客户端支持的视频频编码
	videoCodecs, ok := property["videoCodecs"]
	if ok {
		s.connect.VideoCodecs = int(videoCodecs.(float64))
	}

	// 表明支持的特殊的视频函数
	videoFunction, ok := property["videoFunction"]
	if ok {
		s.connect.VideoFunction = int(videoFunction.(float64))
	}

	// 加载 SWF 文件的网页地址
	pageURL, ok := property["pageUrl"]
	if ok {
		s.connect.PageURL = pageURL.(string)
	}

	// AMF 编码方法
	objectEncoding, ok := property["objectEncoding"]
	if ok {
		s.connect.ObjectEncoding = int(objectEncoding.(float64))
	}
}

// 解析客户端发过来的connect指令
// 客户端发送连接命令给服务器去请求连接服务器上的一个应用实例
// 消息格式: <命令名称: connect>,<事务ID: number>,<命令对象: object>,<可选用户参数: object>
func (s *ConnServer) handleConnect(args []interface{}) error {
	for k, v := range args {
		switch v := v.(type) {
		case float64:
			// 事务ID
			id := int(v)
			if id != 1 {
				return fmt.Errorf("invalid transaction ID=%d", id)
			}

			s.transactionID = 1
		case amf.Object:
			// 命令对象
			if k == 1 {

				s.connectCommandMsg(v)
				continue
			}

			// 【可以不实现】用户可选参数
			slog.Debug("unrealized optional user object")
		}
	}

	return nil
}

// 解析客户端发过来的fcUnpublish指令
// 消息格式: <命令名称: FCUnpublish>,<事务ID: number>,<命令对象: nil>,<流名称: string>
func (s *ConnServer) handleFcunpublish(args []interface{}) error {
	for _, v := range args {
		switch v := v.(type) {
		case float64:
			s.transactionID = uint32(v)
		case amf.Object:
		case string:
			publishName := v
			if publishName != s.publish.Name {
				return fmt.Errorf("invalid publish name=%s", publishName)
			}
		}
	}

	return nil
}

// 解析客户端发过来的deleteStream指令，不需要响应客户端
// 格式：<命令名称: deleteStream>,<事务ID: 0>,<命令对象: nil>,<流ID: number>
func (s *ConnServer) handleDeleteStream(args []interface{}) error {
	for k, v := range args {
		switch v := v.(type) {
		case float64:
			num := uint32(v)

			switch k {
			case 0:
				// 事务ID
				transactionID := num
				if transactionID != 0 {
					return errors.New("invalid request data")
				}
			case 2:
				// 流ID
				streamID := num
				if streamID != s.StreamID {
					return errors.New("invalid stream ID")
				}
			}
		case amf.Object:
		}
	}

	return nil
}
