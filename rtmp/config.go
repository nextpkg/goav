package core

import "git.code.oa.com/idc/vdn/v4/base"

const (
	codePublishStart   = "NetStream.Publish.Start"
	codePlayStart      = "NetStream.Play.Start"
	codeConnectSuccess = "NetConnection.Connect.Success"
	flashVer           = "FMLE/3.0 (compatible; Lavf58.12.100)"
)

// rtmp command
const (
	Connect         = "connect"
	CreateStream    = "createStream"
	Publish         = "publish"
	Play            = "play"
	ReleaseStream   = "releaseStream"
	FcPublish       = "FCPublish"
	FCUnpublish     = "FCUnpublish"
	GetStreamLength = "getStreamLength"
	DeleteStream    = "deleteStream"
)

const (
	publishLive   = "live"
	publishRecord = "record"
	publishAppend = "append"
)

const (
	levelWarning = "warning"
	levelStatus  = "status"
	levelError   = "error"
)

const (
	respResult = "_result" // 结果是正确的
	respError  = "_error"  // 结果是错误的
	onStatus   = "onStatus"
	onBWDone   = "onBWDone"
)

// ConnBufSize 连接缓冲大小
const ConnBufSize = 4 * 1024

// readWriteCloser RTMP读写接口
type readWriteCloser interface {
	// Read IO操作
	Read(*ChunkStream) error
	Write(*ChunkStream) error
	Close() error
	Flush() error

	// GetInfo 功能辅助
	GetInfo() (app, instance string)
	GetPublish() *base.PublishInfo
	GetConnect() *base.ConnectInfo
}
