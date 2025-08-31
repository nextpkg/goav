package gop

import (
	"errors"

	"github.com/nextpkg/goav/packet"
)

// 每个GOP的最大包数
const maxGop = 1024

type array struct {
	packets []*packet.Packet
}

func newArray() *array {
	return &array{
		packets: make([]*packet.Packet, 0, maxGop),
	}
}

func (a *array) reset() {
	a.packets = a.packets[:0]
}

// 将数据包保存到数组中
func (a *array) write(p *packet.Packet) error {
	// 如果关键帧间隔太长, 会导致在有限容量内不能完整存储视频
	if len(a.packets) > maxGop {
		return errors.New("the group of picture is too large")
	}

	a.packets = append(a.packets, p)
	return nil
}

// 将缓存的数据一次性写入w中
func (a *array) sendTo(w packet.Writer) error {
	l := len(a.packets)

	// 循环写出
	for i := 0; i < l; i++ {
		err := w.Write(a.packets[i])
		if err != nil {
			return err
		}
	}

	return nil
}
