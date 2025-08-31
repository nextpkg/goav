package gop

import (
	"errors"

	"github.com/nextpkg/goav/packet"
)

// gopCache GOP缓存
type gopCache struct {
	len   int // 当前使用到第几个新的GOP(<=cap)
	cap   int // 一共有几个GOP
	index int // 当前使用第几个GOP
	gops  []*array
}

// newGopCache GOP缓存
// num不能等于0，否则会让播放器收到无法解码的H264包
func newGopCache(num int) *gopCache {
	return &gopCache{
		len:   0,
		cap:   num,
		index: -1,
		gops:  make([]*array, num),
	}
}

// Write 写入一个包到GOP缓存中，create指示是否需要创建新的GOP
func (g *gopCache) Write(p *packet.Packet, create bool) error {
	if g.cap <= 0 || len(g.gops) <= 0 {
		return nil
	}

	// 计算index, 准备新GOP缓存
	var gop *array
	if create {
		g.index = (g.index + 1) % g.cap

		// 得到一个GOP缓冲, 用来存放GOP数据
		gop = g.gops[g.index]
		if gop == nil {
			// 新建一个GOP
			gop = newArray()
			g.len++
			g.gops[g.index] = gop
		} else {
			// 复位GOP
			gop.reset()
		}

		return gop.write(p)
	}

	// index未经过初始化
	if g.index < 0 {
		return errors.New("uninitialized gop,index<0")
	}

	// 得到一个GOP缓冲, 用来存放GOP数据
	gop = g.gops[g.index]
	if gop == nil {
		return errors.New("unexpected gop index")
	}

	return gop.write(p)
}

// SendTo 将环形GOP的数据按从旧到新的序列写入w中
func (g *gopCache) SendTo(w packet.Writer) error {
	if g.cap <= 0 || g.len <= 0 {
		return nil
	}

	baseIndex := (g.index + 1) % g.len

	for i := 0; i < g.len; i++ {
		index := (baseIndex + i) % g.len
		gop := g.gops[index]

		err := gop.sendTo(w)
		if err != nil {
			return err
		}
	}

	return nil
}
