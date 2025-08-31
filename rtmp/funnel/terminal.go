// Package funnel 数据漏斗通用终端
package funnel

import (
	"github.com/nextpkg/goav/packet"
	"github.com/nextpkg/goav/rtmp/comm"
)

// Terminal 终端处理
type Terminal interface {
	comm.InfoBaser
	packet.Writer
	Before()
	After()
	Name() string
}

// Universal 通用终端
type Universal struct {
	info *comm.Info
}

// NewUniversal 通用终端
func NewUniversal(info *comm.Info) *Universal {
	return &Universal{
		info: info,
	}
}

// Before 在主流程之前要做的工作
func (u *Universal) Before() {}

// After 在主流程之后要做的工作
func (u *Universal) After() {}

// Info 终端描述
func (u *Universal) Info() *comm.Info {
	return u.info
}
