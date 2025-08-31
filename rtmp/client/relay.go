// Package client 动态转播，使用TCP连接两端进行数据交换
package client

import (
	"log/slog"
	"time"

	"github.com/nextpkg/goav/rtmp/chunk"
	"github.com/nextpkg/goav/rtmp/control"
	"github.com/pkg/errors"
)

const maxQueue = 4096

// DynamicRelay 动态转播
type DynamicRelay struct {
	puller  *ConnClient
	pusher  *ConnClient
	pullURL string
	pushURL string
	control *control.Control
	chkChan chan *chunk.ChunkStream
	timeout time.Duration
}

// NewDynamicRelay 构造动态转播实例
func NewDynamicRelay(pullURL string, pushURL string, timeout time.Duration) *DynamicRelay {
	puller := NewConnClientByURL(pullURL, timeout)
	if puller == nil {
		slog.Error("Init client connection failed.", "pullURL", pullURL)
		return nil
	}
	pusher := NewConnClientByURL(pushURL, timeout)
	if pusher == nil {
		slog.Error("Init client connection failed.", "pushURL", pushURL)
		return nil
	}

	return &DynamicRelay{
		puller:  puller,
		pusher:  pusher,
		pullURL: pullURL,
		pushURL: pushURL,
		control: control.NewControl(true),
		chkChan: make(chan *chunk.ChunkStream, maxQueue),
		timeout: timeout,
	}
}

// RenewConnection 重新建立连接
func (d *DynamicRelay) RenewConnection() {
	puller := NewConnClientByURL(d.pullURL, d.timeout)
	if puller == nil {
		slog.Error("renew connection failed.", "pullURL", d.pullURL)
		return
	}
	d.puller = puller

	pusher := NewConnClientByURL(d.pushURL, d.timeout)
	if pusher == nil {
		slog.Error("renew connection failed.", "pushURL", d.pushURL)
		return
	}
	d.pusher = pusher
}

// Start 启动转播
func (d *DynamicRelay) Start() error {
	if !d.IsDone() {
		return nil
	}

	// 启动puller客户端
	err := d.puller.StartPlay()
	if err != nil {
		return errors.Wrap(err, "start puller failed")
	}

	// 启动pusher客户端
	err = d.pusher.StartPublish()
	if err != nil {
		d.closePuller()
		return errors.Wrap(err, "start pusher failed")
	}

	// 清理旧队列中的消息, 启动控制器
	d.chkChan = make(chan *chunk.ChunkStream, maxQueue)
	d.control.Restart()

	// 使用puller从数据源pull数据, push到pusher中
	go d.pull()
	go d.push()

	slog.Info("Relay is turned on", "pullURL", d.pullURL, "pushURL", d.pushURL)
	return nil
}

// IsDone 转播服务是否在工作
func (d *DynamicRelay) IsDone() bool {
	return d.control.IsDone()
}

// Wait 等待服务完成
func (d *DynamicRelay) Wait() <-chan struct{} {
	return d.control.Done()
}

// Stop 暂停转播
func (d *DynamicRelay) Stop() {
	d.control.Cancel()
}

func (d *DynamicRelay) closePuller() {
	err := d.puller.Close()
	if err != nil {
		slog.Error("[Close] failed", "err", errors.Wrap(err, "close puller failed"))
		return
	}
}

// 从服务器中拉取数据, 转发到队列中
func (d *DynamicRelay) pull() {
	for {
		if d.IsDone() {
			d.closePuller()
			return
		}

		cs := &chunk.ChunkStream{}

		// 读取chunk
		err := d.puller.Read(cs)
		if err != nil {
			d.Stop()
			slog.Debug("[Read] failed",
				"err", errors.Wrap(err, "read failed"),
				"pullURL", d.pullURL,
			)
			continue
		}

		// 在channel空间还足够的时候, 向channel内发送数据
		if len(d.chkChan) > maxQueue {
			d.Stop()
			slog.Error("packet queue saturated.",
				"pullURL", d.pullURL,
			)
			continue
		}

		d.chkChan <- cs
	}
}

// 从队列中取数据, 向pusher推送数据
func (d *DynamicRelay) push() {
	for {
		select {
		case <-d.Wait():
			err := d.pusher.Close()
			if err != nil {
				slog.Debug("[Close] pusher failed",
					"err", err,
					"pushURL", d.pushURL,
				)
			}
			return
		case cs := <-d.chkChan:
			err := d.pusher.Write(cs)
			if err != nil {
				slog.Debug("[Write] failed",
					"err", err,
					"pushURL", d.pushURL,
				)
				d.Stop()
			}
		}
	}
}
