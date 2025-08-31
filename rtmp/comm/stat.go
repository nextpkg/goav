package comm

import (
	"time"

	"github.com/nextpkg/goav/packet"
	"go.uber.org/atomic"
)

// 更新读写统计数据的时间, 单位: 毫秒
const statDelayTime int64 = 5000

// Stat is connection statistic
type Stat struct {
	videoLen      atomic.Uint64 // [当前]总接收到视频字节数
	audioLen      atomic.Uint64 // [当前]总接收到音频字节数
	lastVideoLen  uint64        // [上次]总接收到视频字节数
	lastAudioLen  uint64        // [上次]总接收到音频字节数
	videoBps      atomic.Uint64 // 视频速率(每秒字节)
	audioBps      atomic.Uint64 // 音频速率(每秒字节)
	lastTimestamp int64         // [上次]更新时间(单位: 毫秒)
	initTimestamp int64         // 统计服务被初始化的时间
}

// NewStat 统计音视频读写速率
func NewStat() *Stat {
	return &Stat{
		lastTimestamp: time.Now().UnixNano() / 1e6, // ms
		initTimestamp: time.Now().Unix(),
	}
}

// Update 更新统计值
func (s *Stat) Update(p *packet.Packet) {
	DataLen := uint64(len(p.Data))

	// 收集已接收的音视频的长度
	switch p.Type {
	case packet.PktVideo:
		s.videoLen.Add(DataLen)
	case packet.PktAudio:
		s.audioLen.Add(DataLen)
	}

	now := time.Now().UnixNano() / 1e6

	// 计算视频和音频速率
	if (now - s.lastTimestamp) >= statDelayTime {
		interval := (now - s.lastTimestamp) / 1000

		s.videoBps.Store((s.videoLen.Load() - s.lastVideoLen) / uint64(interval))
		s.audioBps.Store((s.audioLen.Load() - s.lastAudioLen) / uint64(interval))

		s.lastVideoLen = s.videoLen.Load()
		s.lastAudioLen = s.audioLen.Load()
		s.lastTimestamp = now
	}
}

// Duration is the stream duration until now
func (s *Stat) Duration() int64 {
	return time.Now().Unix() - s.initTimestamp
}

// VideoLen is video size until now
func (s *Stat) VideoLen() uint64 {
	return s.videoLen.Load()
}

// AudioLen is audio size until now
func (s *Stat) AudioLen() uint64 {
	return s.audioLen.Load()
}

// VideoBps is video rate until now
func (s *Stat) VideoBps() uint64 {
	return s.videoBps.Load()
}

// AudioBps is audio rate until now
func (s *Stat) AudioBps() uint64 {
	return s.audioBps.Load()
}
