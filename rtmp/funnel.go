// Package core 数据漏斗: 数据单方面由上层写入，终端接收
package core

import (
	"sync"
	"time"

	"git.code.oa.com/idc/vdn/v4/base"
	"git.code.oa.com/idc/vdn/v4/metrics"
	"git.code.oa.com/idc/vdn/v4/packet"
	"git.code.oa.com/trpc-go/trpc-go/log"
	"go.uber.org/atomic"
)

var online atomic.Int64

func init() {
	go metrics.ReportLoop(func() {
		metrics.GaugeGlobal(metrics.Playback, float64(online.Load()))
	})
}

// 必须大于1
const maxQueueLen = 4096

// Funnel 数据漏斗，数据单方面由上层写入，终端接收
type Funnel struct {
	Terminal

	// 数据统计
	*base.Active
	*base.Stat
	checkin int64

	// 流程控制
	wg     sync.WaitGroup
	status atomic.Bool

	// 数据通道
	pktChan chan *packet.Packet
}

// NewFunnel 终端执行顺序：Before() -> [loop]Write() -> After() -> Close()
func NewFunnel(t Terminal) *Funnel {
	ret := &Funnel{
		Terminal: t,
		Active:   base.NewRwAlive(),
		Stat:     base.NewStat(),
		checkin:  time.Now().Unix(),
		pktChan:  make(chan *packet.Packet, maxQueueLen),
	}

	ret.wg.Add(1)
	go ret.processor()
	online.Inc()

	return ret
}

// 处理流进来的数据
func (f *Funnel) processor() {
	f.Terminal.Before()
	defer func() {
		f.Terminal.After()
		f.wg.Done()
	}()

	for {
		p := <-f.pktChan
		if p == nil {
			break
		}

		// 复制一个新的结构体，避免由多客户同时读导致的读写竞态
		pkt := *p

		// (公式: pts = dts + baseline)
		pkt.Baseline = pkt.TimeStamp + f.GetBaseTime()

		// 将数据传给客户端，如果客户端返回错误则认为是不可恢复的错误，直接关闭客户端
		err := f.Terminal.Write(&pkt)
		if err != nil {
			log.Debug(err)
			f.Close()
			break
		}

		// 更新状态信息
		f.Active.Keepalive()
		f.Active.SetMediaTime(&pkt)

		// 更新统计数据
		f.Stat.Update(&pkt)

		metrics.Average(f.Info().App, f.Info().Instance, metrics.PlaybackBW,
			float64(f.Stat.VideoBps()+f.Stat.AudioBps())*8,
		)
	}
}

// Close 关闭
func (f *Funnel) Close() {
	if !f.status.CAS(false, true) {
		return
	}

	online.Dec()

	// 使用nil作为结束符而不是直接close管道，能有效保证结束后未被处理的数据包能被妥善处理
	if len(f.pktChan) < maxQueueLen {
		f.pktChan <- nil
		return
	}

	log.Error("close funnel failed,because queue is saturated")
}

// Write 写数据
func (f *Funnel) Write(pkt *packet.Packet) error {
	if f.status.Load() {
		return ErrWriterWasCanceled
	}

	if len(f.pktChan) < maxQueueLen-1 {
		f.pktChan <- pkt
		return nil
	}

	return ErrQueueSaturated
}

// Wait 等待关闭
func (f *Funnel) Wait() {
	f.wg.Wait()
}

// Checkin 登记时间
func (f *Funnel) Checkin() int64 {
	return f.checkin
}
