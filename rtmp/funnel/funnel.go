// Package funnel 数据漏斗: 数据单方面由上层写入，终端接收
package funnel

import (
	"log/slog"
	"sync"
	"time"

	"github.com/nextpkg/goav/packet"
	"github.com/nextpkg/goav/rtmp/ce"
	"github.com/nextpkg/goav/rtmp/comm"
	"go.uber.org/atomic"
)

var online atomic.Int64

// 必须大于1
const maxQueueLen = 4096

// Funnel 数据漏斗，数据单方面由上层写入，终端接收
type Funnel struct {
	Terminal

	// 数据统计
	*comm.Active
	*comm.Stat
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
		Active:   comm.NewRwAlive(),
		Stat:     comm.NewStat(),
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
	f.Before()
	defer func() {
		f.After()
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
			slog.Debug(err.Error())
			f.Close()
			break
		}

		// 更新状态信息
		f.Keepalive()
		f.SetMediaTime(&pkt)

		// 更新统计数据
		f.Update(&pkt)
	}
}

// Close 关闭
func (f *Funnel) Close() {
	if !f.status.CompareAndSwap(false, true) {
		return
	}

	online.Dec()

	// 使用nil作为结束符而不是直接close管道，能有效保证结束后未被处理的数据包能被妥善处理
	if len(f.pktChan) < maxQueueLen {
		f.pktChan <- nil
		return
	}

	slog.Error("close funnel failed,because queue is saturated")
}

// Write 写数据
func (f *Funnel) Write(pkt *packet.Packet) error {
	if f.status.Load() {
		return ce.ErrWriterWasCanceled
	}

	if len(f.pktChan) < maxQueueLen-1 {
		f.pktChan <- pkt
		return nil
	}

	return ce.ErrQueueSaturated
}

// Wait 等待关闭
func (f *Funnel) Wait() {
	f.wg.Wait()
}

// Checkin 登记时间
func (f *Funnel) Checkin() int64 {
	return f.checkin
}
