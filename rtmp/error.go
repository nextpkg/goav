package core

import "errors"

var (
	// ErrWriterWasCanceled writer被取消
	ErrWriterWasCanceled = errors.New("writer was canceled")
	// ErrQueueSaturated 队列满
	ErrQueueSaturated = errors.New("queue saturated")
)
