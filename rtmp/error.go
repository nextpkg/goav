package rtmp

import "errors"

var (
	// ErrWriterWasCanceled writer被取消
	ErrWriterWasCanceled = errors.New("writer was canceled")
	// ErrQueueSaturated 队列满
	ErrQueueSaturated = errors.New("queue saturated")
	// ErrRespFailed 响应数据格式错误
	ErrRespFailed = errors.New("invalid server response data")
)
