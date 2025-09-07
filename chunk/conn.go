// Package chunk rtmp 收发RTMP分块功能
package chunk

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Option RTMP 连接选项
type Option struct {
	/**
	每个chunk最大的长度，这是理论上的最大长度，用来度量过长的chunk
	注意：这个参数可能会阻碍码率非常高的流的连接
	*/
	maxChunkSize uint32
	/**
	超时时间，限制网络IO的读写超时
	*/
	timeout time.Duration
}

var DefaultOption = Option{
	maxChunkSize: 100 * 1024 * 1024,
	timeout:      10 * time.Second,
}

// Conn 收发RTMP分块
type Conn struct {
	net.Conn
	/**
	rtmp传输的块大小，设置RTMP的最大chunk尺寸。
	该值越小，CPU使用率越高，延时越低；反之该值越大，CPU使用率越高，延时越高。
	*/
	ChunkSize       uint32 // 本地的块大小，按此标准写chunk
	RemoteChunkSize uint32 // 远程的块大小，按此标准读chunk
	/**
	接收端确认窗口，设置RTMP确认包的窗口大小。
	该值显示的是对端已接收的字节数；用来实现控制发送速率而降低拥塞可能的技术，值越大，窗口越大，带宽越大。
	*/
	WindowAckSize       uint32 // 接收端接收的最大字节数(确认窗口)
	RemoteWindowAckSize uint32 // 发送端发送的最大字节数(确认窗口)
	received            uint32 // 已接收的字节数(大于0xf0000000则归零)
	ackReceived         uint32 // 服务端已接收的字节数（但未确认）
	sent                uint32 // 已发送的字节数(大于0xf0000000则归零)
	ackSent             uint32 // 客户端已接收的字节数
	bandwidthLimitType  byte   // 带宽限制类型
	Option              Option
	Rw                  *ReadWriter             // 网络缓冲
	slab                *Slab                   // chunk内存分配器
	Chunks              map[uint32]*ChunkStream // chunk接收
	csID                uint32
	rLocker             sync.Mutex
	wLocker             sync.Mutex
}

// NewConn RTMP连接器
func NewConn(conn net.Conn, bufSize int) *Conn {
	return &Conn{
		Conn:                conn,
		ChunkSize:           128,       // 起始值一定是128
		RemoteChunkSize:     128,       // 起始值一定是128
		WindowAckSize:       2500000,   // 默认本地确认窗口(字节)
		RemoteWindowAckSize: 2500000,   // 默认远程确认窗口(字节)
		bandwidthLimitType:  0,         // 带宽限制, 类型: Soft limit
		slab:                NewSlab(), // 内存分配器
		Rw:                  NewReadWriter(conn, bufSize),
		Chunks:              make(map[uint32]*ChunkStream),
		Option:              DefaultOption,
	}
}

// InitSlab 初始化内存管理器的阈值
func (c *Conn) InitSlab(min, max int) {
	c.slab.Init(min, max)
}

// Read 【读锁】读取一个完整chunk的数据
func (c *Conn) Read(cs *ChunkStream) error {
	for {
		c.rLocker.Lock()

		// 只读不写
		ncs, err := c.getIntactChunk()
		if err != nil {
			c.rLocker.Unlock()
			return err
		}

		c.rLocker.Unlock()

		// 客户端或服务器在收到数据后必须向对端发送与视窗大小相等的确认消息
		c.ack(ncs.Length)

		// 处理控制消息，消息不再传递给上层
		if !c.handleControlMsg(ncs) {
			*cs = *ncs
			return nil
		}
	}
}

// Write 【写锁】写入一个完整的chunk
func (c *Conn) Write(cs *ChunkStream) error {
	if false {
		// 监控发送的数据量，在客户端接收能力不足时应控制发送速率，直到客户端接收量恢复正常
		ackSent := atomic.LoadUint32(&c.ackSent)

		// 为避免客户端一直接收不过来，这里采用了丢包的策略
		if c.sent > c.WindowAckSize && c.sent > 3*ackSent {
			slog.Error("sent>3*ackSent", "send", c.sent, "ackSent", ackSent)
			return nil
		}
	}

	c.wLocker.Lock()
	defer c.wLocker.Unlock()

	// 只写不读
	err := cs.WriteChunk(c.Rw, c.ChunkSize)
	if err != nil {
		return err
	}

	// 发送量统计
	c.sent += cs.Length
	if c.sent >= 0xf0000000 {
		c.sent = 0
	}

	return nil
}

// Flush 【写锁】刷新缓冲
func (c *Conn) Flush() error {
	c.wLocker.Lock()
	defer c.wLocker.Unlock()

	return c.Rw.Flush()
}

// Close 关闭connection时要执行的工作
func (c *Conn) Close() error {
	if c.slab.Stat.Max > 0 || c.slab.Stat.Medium > 0 || c.slab.Stat.Min > 0 {
		slog.Debug("Connection is closed,slab stat",
			"from", c.Conn.RemoteAddr(),
			"max", c.slab.Stat.Max,
			"medium", c.slab.Stat.Medium,
			"min", c.slab.Stat.Min,
		)
	}
	return c.Conn.Close()
}

// 从网络缓冲中拼接chunk
func (c *Conn) getIntactChunk() (*ChunkStream, error) {
	var counter uint32

	// 标记数据无效，加速GC回收
	ncs, ok := c.Chunks[c.csID]
	if ok {
		ncs.Data = nil
	}

	for {
		// 放弃过长的chunk
		if counter*c.ChunkSize > c.Option.maxChunkSize {
			return nil, fmt.Errorf("chunk too large(single chunk %d > %d), discard it",
				counter*c.ChunkSize, c.Option.maxChunkSize)
		}

		// 读取一个字节, 以大端的形式保存到h中
		basicHeader, err := c.Rw.ReadUintBE(1)
		if err != nil {
			return nil, err
		}

		// chunk basic header
		format := basicHeader >> 6 /* [0,3] */
		csID := basicHeader & 0x3f /* [0,63] */

		// 潜在风险：无效的csID会有很多吗？可能chunks会变得很大
		ncs, ok = c.Chunks[csID]
		if !ok {
			ncs = &ChunkStream{
				Csid: csID,
			}
			c.Chunks[csID] = ncs
		}

		// 块消息类型
		ncs.FormatTmp = format

		// read chunk
		err = ncs.ReadChunk(c.Rw, c.RemoteChunkSize, c.slab)
		if err != nil {
			return nil, err
		}

		// store
		c.Chunks[csID] = ncs

		// 读完一个chunk
		if ncs.Intact() {
			return ncs, nil
		}

		// 分片数
		counter++
	}
}

func (c *Conn) setChunkSize(cs *ChunkStream) {
	remoteChunkSize := binary.BigEndian.Uint32(cs.Data)

	// 第一位必须是0
	if (remoteChunkSize >> 31) != 0 {
		slog.Error("incorrect control value when setting chunk size")
		return
	}

	// 理论上最大块大小不能小于1字节(实际上太小会增加CPU负载, 所以设置默认128)
	if remoteChunkSize < 128 {
		slog.Error("incorrect chunk size", "size", remoteChunkSize)
		return
	}

	// 任何块不可能比消息(最大值0xFFFFFF)大
	if c.RemoteChunkSize > 0xFFFFFF {
		remoteChunkSize = 0xFFFFFF
	}

	// 设置块大小，用于通知另一端新的最大块大小
	c.RemoteChunkSize = remoteChunkSize
	slog.Debug("remote chunk size is changed to", "size", remoteChunkSize)
}

func (c *Conn) setPeerBandwidth(cs *ChunkStream) {
	bandwidth := binary.BigEndian.Uint32(cs.Data)

	// 通过将已发送但尚未被确认的数据总数限制为该消息指定的视窗大小, 来实现限制输出带宽的目的
	limitType := cs.Data[4]

	// Dynamic: 如果上一个消息的限制类型为Hard，则该消息同样为Hard，否则抛弃该消息
	if limitType == 2 && c.bandwidthLimitType == 0 {
		limitType = 0
	}

	// Hard and Soft limit
	switch limitType {
	case 0:
		// Hard: 消息接收端应该将输出带宽限制为指定视窗大小
		c.ackReceived = bandwidth
	case 1:
		// Soft: 消息接收端应该将输出带宽限制为指定视窗大小和当前视窗大小中较小的值
		if bandwidth < c.ackReceived {
			c.ackReceived = bandwidth
		}
	}

	slog.Debug("bandwidth is changed to", "size=", bandwidth)
}

// 处理控制消息
func (c *Conn) handleControlMsg(cs *ChunkStream) bool {
	switch cs.TypeID {
	case IDSetChunkSize:
		c.setChunkSize(cs)
		return true
	case IDAbortMessage:
		streamID := binary.BigEndian.Uint32(cs.Data)

		/**
		一个实例一个streamID，只要配对上了就可以关闭流，这种方式可以实现用户控制服务端关闭连接。
		但由于用户主动关闭的效果差别也不大，我们也没找到被动关闭的使用场景，因此这里不实现。
		*/
		slog.Error("ignore unrealize abort message", "stream id", streamID)
		return true
	case IDAck:
		atomic.StoreUint32(&c.ackSent, binary.BigEndian.Uint32(cs.Data))
		return true
	case IDUser:
		c.handleUserMsg(cs)
		return true
	case IDWindowAckSize:
		// 客户端或服务端发送该消息来通知对端发送确认消息所使用的视窗大小
		c.RemoteWindowAckSize = binary.BigEndian.Uint32(cs.Data)
		return true
	case IDSetPeerBandwidth:
		c.setPeerBandwidth(cs)
		return true
	default:
		return false
	}
}

// 处理用户消息
func (c *Conn) handleUserMsg(cs *ChunkStream) {
	eventType := binary.BigEndian.Uint16(cs.Data)
	switch eventType {
	case StreamSetBufferLen:
		// 事件数据的前4字节表示流ID,接下来的4字节表示缓冲区的大小(单位是毫秒)
		if len(cs.Data) != 10 {
			slog.Debug("setBufferLen command data != 10", "data", len(cs.Data))
			return
		}

		streamID := binary.BigEndian.Uint32(cs.Data[2:6])
		bufferLen := binary.BigEndian.Uint32(cs.Data[6:10])

		// 【实时流可以实现】客户端指明每毫秒能接收的字节数，服务端应该按这个字节数来调整发送速率
		slog.Debug("unrealized setBufferLen command",
			"streamID", streamID,
			"bufferLen", bufferLen,
		)
	case PingResponse:
		// 事件数据是客户端接收到pingRequest请求时的4字节时间戳
		if len(cs.Data) != 6 {
			slog.Error("pingResponse command data != 6", "data", len(cs.Data))
			return
		}

		timestamp := binary.BigEndian.Uint32(cs.Data[2:6])
		slog.Debug("unrealized pingResponse command", "timestamp", timestamp)
	}
}

// 客户端或服务器在收到数据后必须向对端发送与视窗大小相等的确认消息
func (c *Conn) ack(received uint32) {
	c.received += received
	if c.received >= 0xf0000000 {
		c.received = 0
	}

	// 已接收但未确认的数据总数
	c.ackReceived += received
	if c.ackReceived >= c.RemoteWindowAckSize {
		cs := NewAck(c.ackReceived)
		err := c.Write(cs)
		if err != nil {
			slog.Error(err.Error())
		}

		c.ackReceived = 0
	}
}
