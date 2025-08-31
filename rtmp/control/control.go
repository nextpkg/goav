package control

import (
	"context"

	"go.uber.org/atomic"
)

// Control 控制阻塞接口的流程, 包括但不限于设置超时/取消等
type Control struct {
	ctx    context.Context
	cancel context.CancelFunc
	isDone atomic.Bool // 设置标记位, 为了让某些循环避免使用管道轮询, 提高IO吞吐量
}

// NewControl 控制阻塞接口的流程, 包括但不限于设置超时/取消等
// isDone参数改变的是默认的状态，isDone=true意味着控制器初始状态是已完成，isDone=false则意味着待完成。
func NewControl(isDone bool) *Control {
	ctx, cancel := context.WithCancel(context.Background())

	ret := &Control{
		ctx:    ctx,
		cancel: cancel,
	}

	ret.isDone.Store(isDone)
	return ret
}

// Restart 重置状态为“进行中”
func (c *Control) Restart() {
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.isDone.Store(false)
}

// Cancel 取消
func (c *Control) Cancel() {
	c.isDone.Store(true)
	c.cancel()
}

// IsDone 是否在“已取消”状态
func (c *Control) IsDone() bool {
	return c.isDone.Load()
}

// Done 阻塞等待完成
func (c *Control) Done() <-chan struct{} {
	return c.ctx.Done()
}
