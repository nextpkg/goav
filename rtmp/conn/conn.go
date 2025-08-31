// Package rtmp 服务端与客户端的操作界面
package conn

import (
	"fmt"
	"log/slog"

	"github.com/nextpkg/goav/container/flv"
	"github.com/nextpkg/goav/packet"
	"github.com/nextpkg/goav/rtmp/chunk"
	"github.com/nextpkg/goav/rtmp/comm"
	"github.com/nextpkg/goav/rtmp/funnel"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
)

// Reader 用以从外部连接中读取数据
type Reader struct {
	*comm.Active
	*comm.Stat
	info *comm.Info
	flv  *flv.Demuxer    // RTMP解复用
	conn ReadWriteCloser // 读写工具，注意，读写应该分别拥有独立缓冲区
}

// NewReader 用以从外部连接中读取数据, IsExternal: false
func NewReader(conn ReadWriteCloser) *Reader {
	app, instance := conn.GetInfo()
	return &Reader{
		Active: comm.NewRwAlive(),
		Stat:   comm.NewStat(),
		info:   comm.NewInfo(app, instance, false),
		flv:    flv.NewDemuxer(),
		conn:   conn,
	}
}

// Info Reader描述信息
func (r *Reader) Info() *comm.Info {
	return r.info
}

// Read 从conn读取数据
func (r *Reader) Read(p *packet.Packet) error {
	var cs *chunk.ChunkStream

ReadChunk:
	for {
		cs = &chunk.ChunkStream{}

		// 从"conn"中读取一个数据"chunk"
		err := r.conn.Read(cs)
		if err != nil {
			return err
		}

		// 更新读时间
		r.Active.Keepalive()

		// 类型转换
		switch cs.TypeID {
		case packet.TagVideo:
			p.Type = packet.PktVideo
			break ReadChunk
		case packet.TagAudio:
			p.Type = packet.PktAudio
			break ReadChunk
		case packet.TagScriptDataAMF0, packet.TagScriptDataAMF3:
			p.Type = packet.PktMetadata
			break ReadChunk
		}
	}

	p.TimeStamp = cs.Timestamp
	p.Baseline = cs.Timestamp + r.GetBaseTime()
	p.StreamID = cs.StreamID
	p.Data = cs.Data

	err := r.flv.Demux(p)
	if err != nil {
		return errors.Wrap(err, "flv demux failed")
	}

	switch p.Type {
	case packet.PktVideo:
		vh := p.Header.(packet.VideoPacketHeader)
		if !vh.IsCodecAvc() {
			return fmt.Errorf("incompatible video codec(%d)", vh.CodecID())
		}
	case packet.PktAudio:
		ah := p.Header.(packet.AudioPacketHeader)
		if !ah.IsSoundAAC() {
			return fmt.Errorf("incompatible audio codec(%d)", ah.SoundFormat())
		}
	}

	// 更新统计信息
	r.Active.SetMediaTime(p)
	r.Stat.Update(p)

	return nil
}

// Close 关闭连接
func (r *Reader) Close() {
	err := r.conn.Flush()
	if err != nil {
		slog.Debug("flush failed")
		return
	}

	err = r.conn.Close()
	if err != nil {
		slog.Debug("close failed")
		return
	}
}

// GetPublish 获取底层Publish信息
func (r *Reader) GetPublish() *comm.PublishInfo {
	return r.conn.GetPublish()
}

// GetConnect 获取底层Connect信息
func (r *Reader) GetConnect() *comm.ConnectInfo {
	return r.conn.GetConnect()
}

// Writer 提供向外输出RTMP的writer封装
type Writer struct {
	*funnel.Universal
	status atomic.Bool

	// 数据源，客户端 or 服务端 连接, 注意，读写应该分别拥有独立缓冲区
	conn ReadWriteCloser
}

// NewWriter 提供向外输出RTMP的writer封装, IsExternal: true
func NewWriter(conn ReadWriteCloser) *funnel.Funnel {
	app, instance := conn.GetInfo()
	info := comm.NewInfo(app, instance, true)

	ret := &Writer{
		Universal: funnel.NewUniversal(info),
		conn:      conn,
	}

	go ret.connRead()

	return funnel.NewFunnel(ret)
}

// 从客户端读取数据, Read逻辑会有响应客户端的指令
func (w *Writer) connRead() {
	for !w.status.Load() {
		cs := &chunk.ChunkStream{}
		err := w.conn.Read(cs)
		if err != nil {
			// rtmp reader was closed
			slog.Debug("[connRead] failed",
				"key", w.Info().Key,
				"err", err)
			break
		}
	}
	w.Close()
}

// Write 写源数据到队列
func (w *Writer) Write(p *packet.Packet) error {
	cs := &chunk.ChunkStream{
		Data:      p.Data,
		Length:    uint32(len(p.Data)),
		StreamID:  p.StreamID,
		Timestamp: p.Baseline,
	}

	switch p.Type {
	case packet.PktVideo:
		cs.TypeID = packet.TagVideo
	case packet.PktAudio:
		cs.TypeID = packet.TagAudio
	case packet.PktMetadata:
		cs.TypeID = packet.TagScriptDataAMF0
	}

	err := w.conn.Write(cs)
	if err != nil {
		w.Close()
		return errors.Wrapf(err, "[%s]rtmp write failed", w.Info().Key)
	}

	return nil
}

// After 在主流程之后要做的工作
func (w *Writer) After() {
	w.Close()

	if err := w.conn.Flush(); err != nil {
		slog.Debug("[Flush] failed", "err", err)
	}
	if err := w.conn.Close(); err != nil {
		slog.Debug("[Close] failed", "err", err)
	}
}

// Close 发起关闭流程时要处理的工作
func (w *Writer) Close() {
	w.status.Store(true)
}

// Name writer's name
func (w *Writer) Name() string {
	return "rtmp"
}
