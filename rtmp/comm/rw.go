package comm

import "github.com/nextpkg/goav/packet"

// ConnectInfo "Connect"指令
type ConnectInfo struct {
	App      string `amf:"app" json:"app"`
	FlashVer string `amf:"flashVer" json:"flashVer"`
	SwfURL   string `amf:"swfUrl" json:"swfUrl"`
	// URL of the target stream. Defaults to proto://host[:port]/app
	TcURL          string `amf:"tcUrl" json:"tcUrl"`
	FPad           bool   `amf:"fpad" json:"fpad"`
	AudioCodecs    int    `amf:"audioCodecs" json:"audioCodecs"`
	VideoCodecs    int    `amf:"videoCodecs" json:"videoCodecs"`
	VideoFunction  int    `amf:"videoFunction" json:"videoFunction"`
	PageURL        string `amf:"pageUrl" json:"pageUrl"`
	ObjectEncoding int    `amf:"objectEncoding" json:"objectEncoding"`
}

// PublishInfo "Publish"指令
type PublishInfo struct {
	Name     string                 // 发布流的名称
	Type     string                 // 发布流的类型, 设置为"live", "record", "append"
	MetaData map[string]interface{} // 发布流的元数据
}

// InfoBaser is the basement of `info`
type InfoBaser interface {
	Info() *Info
}

// ActiveBaser is the basement of `alive`
type ActiveBaser interface {
	IsTimeout() bool
}

// WriteBaser is the basement of `write`
type WriteBaser interface {
	packet.Writer
	Close()
}

// WriteCloser 高级写接口
type WriteCloser interface {
	WriteBaser
	AliveWriter
}

// InfoWriter 带特定描述的写基类
type InfoWriter interface {
	packet.Writer
	InfoBaser
}

// AliveWriter 写保活
type AliveWriter interface {
	InfoBaser
	ActiveBaser
	RebaseTime()
}

// ReadCloser 高级读接口
type ReadCloser interface {
	packet.Reader
	Close()

	InfoBaser
	ActiveBaser
	GetPublish() *PublishInfo
	GetConnect() *ConnectInfo
}
