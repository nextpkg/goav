// Package core 数据漏斗通用终端
package core

import (
	"git.code.oa.com/idc/vdn/v4/base"
	"git.code.oa.com/idc/vdn/v4/packet"
)

// Terminal 终端处理
type Terminal interface {
	base.InfoBaser
	packet.Writer
	Before()
	After()
	Name() string
}

// Universal 通用终端
type Universal struct {
	info *base.Info
}

// NewUniversal 通用终端
func NewUniversal(info *base.Info) *Universal {
	return &Universal{
		info: info,
	}
}

// Before 在主流程之前要做的工作
func (u *Universal) Before() {}

// After 在主流程之后要做的工作
func (u *Universal) After() {}

// Info 终端描述
func (u *Universal) Info() *base.Info {
	return u.info
}
