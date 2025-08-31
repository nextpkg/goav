package comm

import (
	"github.com/google/uuid"
	"github.com/nextpkg/goav/packet"
)

// Info 流描述
type Info struct {
	App        string
	Instance   string
	Key        string // 格式: App/Ins
	UID        string // 每个流的唯一ID
	IsExternal bool   // 外部连接标识
}

// NewInfo 流描述
func NewInfo(app, instance string, isExternal bool) *Info {
	return &Info{
		App:        app,
		Instance:   instance,
		Key:        app + "/" + instance,
		UID:        uuid.New().String(),
		IsExternal: isExternal,
	}
}

// Copy 拷贝副本
func (i *Info) Copy() *Info {
	return &Info{
		App:        i.App,
		Instance:   i.Instance,
		Key:        i.Key,
		UID:        uuid.New().String(),
		IsExternal: i.IsExternal,
	}
}

// InfoBaser is the basement of `info`
type InfoBaser interface {
	Info() *Info
}

// InfoWriter 带特定描述的写基类
type InfoWriter interface {
	packet.Writer
	InfoBaser
}

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
