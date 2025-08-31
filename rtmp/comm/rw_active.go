package comm

import (
	"log/slog"
	"time"

	"github.com/nextpkg/goav/packet"
	"go.uber.org/atomic"
)

// Active 用来辅助读写的超时处理
type Active struct {
	active        atomic.Int64
	baseMediaTime atomic.Uint32
	lastVideoTime atomic.Uint32
	lastAudioTime atomic.Uint32
}

// NewRwAlive 用来辅助读写的超时处理
func NewRwAlive() *Active {
	ra := &Active{}
	ra.active.Store(time.Now().Unix())
	return ra
}

// GetBaseTime 获取基准时间戳【migrate会用到】
func (rw *Active) GetBaseTime() uint32 {
	return rw.baseMediaTime.Load()
}

// RebaseTime [写时]设置基准时间戳: 以音频/视频时间戳数值大的一个作为基准时间戳
func (rw *Active) RebaseTime() {
	if rw.lastAudioTime.Load() > rw.lastVideoTime.Load() {
		rw.baseMediaTime = rw.lastAudioTime
	} else {
		rw.baseMediaTime = rw.lastVideoTime
	}
}

// SetMediaTime [写时]设置音视频的最后活跃时间戳(因为RTMP包有增量时间戳)
func (rw *Active) SetMediaTime(p *packet.Packet) {
	switch p.Type {
	case packet.PktVideo:
		rw.lastVideoTime.Store(p.Baseline)
	case packet.PktAudio:
		rw.lastAudioTime.Store(p.Baseline)
	}
}

// LastVideoTime is last active video timestamp
func (rw *Active) LastVideoTime() uint32 {
	return rw.lastVideoTime.Load()
}

// LastAudioTime is last active audio timestamp
func (rw *Active) LastAudioTime() uint32 {
	return rw.lastAudioTime.Load()
}

// Keepalive 保持活跃状态（当前时间）
func (rw *Active) Keepalive() {
	rw.active.Store(time.Now().Unix())
}

// IsTimeout 根据时间判断是否活跃, 活跃标准: 当前时间 - preTime < timeout
func (rw *Active) IsTimeout(rtmpReadTimeout int64) bool {
	if rtmpReadTimeout <= 0 {
		return false
	}

	elapse := time.Now().Unix() - rw.active.Load()
	if elapse <= rtmpReadTimeout {
		return false
	}

	slog.Error("rtmp IO timeout.",
		"elapse(s)", elapse,
		"limit(s)", rtmpReadTimeout,
	)
	return true
}
