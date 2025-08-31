package comm

import "github.com/google/uuid"

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
