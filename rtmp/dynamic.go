// Package core dynamic 动态转播，使用TCP连接两端进行数据交换
package core

import (
	"git.code.oa.com/idc/vdn/v4/control"
	"git.code.oa.com/trpc-go/trpc-go/log"
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
	chkChan chan *ChunkStream
}

// NewDynamicRelay 构造动态转播实例
func NewDynamicRelay(pullURL string, pushURL string) *DynamicRelay {
	puller := NewConnClientByURL(pullURL)
	if puller == nil {
		log.Error("Init client connection failed.Pull url=", pullURL)
		return nil
	}
	pusher := NewConnClientByURL(pushURL)
	if pusher == nil {
		log.Error("Init client connection failed.Push url=", pushURL)
		return nil
	}

	return &DynamicRelay{
		puller:  puller,
		pusher:  pusher,
		pullURL: pullURL,
		pushURL: pushURL,
		control: control.NewControl(true),
		chkChan: make(chan *ChunkStream, maxQueue),
	}
}

// RenewConnection 重新建立连接
func (d *DynamicRelay) RenewConnection() {
	puller := NewConnClientByURL(d.pullURL)
	if puller == nil {
		log.Error("renew connection failed.Pull url=", d.pullURL)
		return
	}
	d.puller = puller

	pusher := NewConnClientByURL(d.pushURL)
	if pusher == nil {
		log.Error("renew connection failed.Push url=", d.pushURL)
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
	d.chkChan = make(chan *ChunkStream, maxQueue)
	d.control.Restart()

	// 使用puller从数据源pull数据, push到pusher中
	go d.pull()
	go d.push()

	log.Infof("Relay (%s) ===> (%s) is turned on", d.pullURL, d.pushURL)
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
		log.Error(errors.Wrap(err, "close puller failed"))
	}
}

// 从服务器中拉取数据, 转发到队列中
func (d *DynamicRelay) pull() {
	for {
		if d.IsDone() {
			d.closePuller()
			return
		}

		cs := &ChunkStream{}

		// 读取chunk
		err := d.puller.Read(cs)
		if err != nil {
			d.Stop()
			log.Tracef("err=%s,pull URL=%s", err, d.pullURL)
			continue
		}

		// 在channel空间还足够的时候, 向channel内发送数据
		if len(d.chkChan) > maxQueue {
			d.Stop()
			log.Errorf("packet queue saturated,puller='%s'", d.pullURL)
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
				log.Tracef("err=%s,push URL=%s", err, d.pushURL)
			}
			return
		case cs := <-d.chkChan:
			err := d.pusher.Write(cs)
			if err != nil {
				log.Tracef("err=%s,push URL=%s", err, d.pushURL)
				d.Stop()
			}
		}
	}
}
