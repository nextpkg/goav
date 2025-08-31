package comm

const (
	CodePublishStart   = "NetStream.Publish.Start"
	CodePlayStart      = "NetStream.Play.Start"
	CodeConnectSuccess = "NetConnection.Connect.Success"
	FlashVer           = "FMLE/3.0 (compatible; Lavf58.12.100)"
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
	PublishLive   = "live"
	PublishRecord = "record"
	PublishAppend = "append"
)

const (
	LevelWarning = "warning"
	LevelStatus  = "status"
	LevelError   = "error"
)

const (
	RespResult = "_result" // 结果是正确的
	RespError  = "_error"  // 结果是错误的
	OnStatus   = "onStatus"
	OnBWDone   = "onBWDone"
)

// ConnBufSize 连接缓冲大小
const ConnBufSize = 4 * 1024
