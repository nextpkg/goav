// Package core 读写rtmp分块数据
// 协议参考: https://chenlichao.gitbooks.io/rtmp-zh_cn/content/
package core

import (
	"encoding/binary"
	"fmt"

	"git.code.oa.com/idc/vdn/v4/packet"
	"github.com/moggle-mog/goav/amf"
	"github.com/pkg/errors"
)

// ChunkStream rtmp的块
type ChunkStream struct {
	format    uint32 // 块类型
	formatTmp uint32 // 当前块类型
	csid      uint32 // 块流ID
	timestamp uint32 // 绝对时间戳
	length    uint32 // 消息长度，类型0和类型1的块包含此字段，表示消息的总长度
	typeID    uint32 // 消息类型ID，类型0和类型1的块包含此字段，表示消息的类型
	streamID  uint32 // 消息流ID，类型0的块包含此字段，表示消息流ID, 相同块流中的消息属于同一个消息流
	timeDelta uint32 // 增量时间戳, 类型1和类型2的块包含此字段，表示前一个块的timestamp字段和当前块timestamp间的差值
	extend    bool   // 指示是否有扩展时间戳, 扩展时间戳存储的是完整值
	index     int    // 指示读取chunk已经读到第index个字节
	remain    int    // chunk待读取的字节数
	got       bool   // got=true: 标识一个chunk已经读取完毕
	data      []byte // 块数据
}

// 返回chunk是否读取完
func (cs *ChunkStream) intact() bool {
	return cs.got
}

// 新建一个chunk data
func (cs *ChunkStream) newChunkData(slab *Slab) error {
	if cs.length == 0 {
		return errors.New("chunk length==0")
	}

	cs.got = false
	cs.index = 0
	cs.remain = int(cs.length) /* 块承载的有效数据 */

	cs.data = slab.Get(cs.remain)
	if cs.data == nil {
		return errors.New("get nil slab")
	}

	return nil
}

func (cs *ChunkStream) fillCSID(r *ReadWriter) error {
	// 补全CSID
	switch cs.csid {
	case 0:
		// read csid(0)
		id, err := r.ReadUintLE(1)
		if err != nil {
			return err
		}

		cs.csid = id + 64
	case 1:
		// read csid(1)
		id, err := r.ReadUintLE(2)
		if err != nil {
			return err
		}

		cs.csid = id + 64
	}

	return nil
}

func (cs *ChunkStream) handleFmt0(r *ReadWriter, slab *Slab) error {
	var err error

	// Basic Header
	cs.format = cs.formatTmp

	// Message Header(read format(0) timestamp)
	cs.timestamp, err = r.ReadUintBE(3)
	if err != nil {
		return err
	}

	// read format(0) length
	cs.length, err = r.ReadUintBE(3)
	if err != nil {
		return err
	}

	// read format(0) type id
	cs.typeID, err = r.ReadUintBE(1)
	if err != nil {
		return err
	}

	// read format(0) stream id
	cs.streamID, err = r.ReadUintLE(4)
	if err != nil {
		return err
	}

	// Extended Timestamp
	if cs.timestamp == 0xffffff {
		// read format(0) extended timestamp id
		cs.timestamp, err = r.ReadUintBE(4)
		if err != nil {
			return err
		}

		cs.extend = true
	} else {
		cs.extend = false
	}

	// read format(0) malloc data field
	err = cs.newChunkData(slab)
	if err != nil {
		return err
	}

	return nil
}

func (cs *ChunkStream) handleFmt1(r *ReadWriter, slab *Slab) error {
	var err error

	// Basic Header
	cs.format = cs.formatTmp

	// Message Header(read format(1) time delta)
	cs.timeDelta, err = r.ReadUintBE(3)
	if err != nil {
		return err
	}

	// read format(1) length
	cs.length, err = r.ReadUintBE(3)
	if err != nil {
		return err
	}

	// read format(1) type id
	cs.typeID, err = r.ReadUintBE(1)
	if err != nil {
		return err
	}

	// Extended Timestamp
	if cs.timeDelta == 0xffffff {
		// read format(1) extended time delta
		cs.timeDelta, err = r.ReadUintBE(4)
		if err != nil {
			return err
		}

		cs.extend = true
	} else {
		cs.extend = false
	}
	cs.timestamp += cs.timeDelta

	// read format(1) malloc data field
	err = cs.newChunkData(slab)
	if err != nil {
		return err
	}

	return nil
}

func (cs *ChunkStream) handleFmt2(r *ReadWriter, slab *Slab) error {
	var err error

	// Basic Header
	cs.format = cs.formatTmp

	// Message Header(read format(2) time delta)
	cs.timeDelta, err = r.ReadUintBE(3)
	if err != nil {
		return err
	}

	// Extended Timestamp
	if cs.timeDelta == 0xffffff {
		// read format(2) extended time delta
		cs.timeDelta, err = r.ReadUintBE(4)
		if err != nil {
			return err
		}

		cs.extend = true
	} else {
		cs.extend = false
	}
	cs.timestamp += cs.timeDelta

	// read format(2) malloc data field
	err = cs.newChunkData(slab)
	if err != nil {
		return err
	}

	return nil
}

func (cs *ChunkStream) handleFmt3(r *ReadWriter, slab *Slab) error {
	var err error

	if cs.remain == 0 {
		// 根据上一个chunk的格式来填充数据
		switch cs.format {
		case 0:

			// 当它跟在Type＝0的chunk后面时，表示和前一个chunk的时间戳都是相同的
			if cs.extend {

				// read format(3:0) extended timestamp
				cs.timestamp, err = r.ReadUintBE(4)
				if err != nil {
					return err
				}
			}
		case 1, 2:
			// 当它跟在Type＝1或者Type＝2的chunk后面时，表示和前一个chunk的时间戳的差是相同的
			var timeDelta uint32
			if cs.extend {

				// read format(3:1,2) extended time delta
				timeDelta, err = r.ReadUintBE(4)
				if err != nil {
					return err
				}
			} else {
				timeDelta = cs.timeDelta
			}
			cs.timestamp += timeDelta
		}

		// malloc format(3) data field
		err = cs.newChunkData(slab)
		if err != nil {
			return err
		}
	} else {
		// 如果有未读完的chunk data, 则试图抛弃掉扩展时间戳
		if cs.extend {
			// peek format(3) timestamp
			b, err := r.Peek(4)
			if err != nil {
				return err
			}

			if binary.BigEndian.Uint32(b) == cs.timestamp {

				// discard format(3) extended timestamp
				if n, err := r.Discard(4); err != nil || n != 4 {
					return err
				}
			}
		}
	}

	return nil
}

// 读取一个chunk(由于chunkSize的限制, 可能需要读取多次才能读取完)
func (cs *ChunkStream) readChunk(r *ReadWriter, chunkSize uint32, slab *Slab) error {
	if chunkSize <= 0 {
		return errors.New("chunk size<=0")
	}

	// 接收块类型3, 无条件(可多次出现)
	// 接收块类型0, 类型1, 块类型2的同时, 待读字节数必须为0(也即是这两类chunk只出现一次)
	if cs.formatTmp != 3 && cs.remain != 0 {
		return fmt.Errorf("invalid chunk remain=%d", cs.remain)
	}

	err := cs.fillCSID(r)
	if err != nil {
		return err
	}

	// 填充 chunk header, 为chunk data准备好buffer
	switch cs.formatTmp {
	case 0:
		if err := cs.handleFmt0(r, slab); err != nil {
			return errors.Wrap(err, "handle format 0")
		}
	case 1:
		if err := cs.handleFmt1(r, slab); err != nil {
			return errors.Wrap(err, "handle format 1")
		}
	case 2:
		if err := cs.handleFmt2(r, slab); err != nil {
			return errors.Wrap(err, "handle format 2")
		}
	case 3:
		if err := cs.handleFmt3(r, slab); err != nil {
			return errors.Wrap(err, "handle format 3")
		}
	default:
		return fmt.Errorf("invalid chunk format=%d", cs.format)
	}

	// 重新计算待读取的chunk data大小
	size := cs.remain
	if size > int(chunkSize) {
		size = int(chunkSize)
	}

	// 准备好合适的buf, 从conn中读取size个字节
	buf := cs.data[cs.index : cs.index+size]

	if _, err := r.Read(buf); err != nil {
		return errors.Wrapf(err, "read %d byte from connection", size)
	}

	cs.index += size
	cs.remain -= size
	if cs.remain == 0 {
		cs.got = true
	}

	return nil
}

// Chunk Basic Header
func (cs *ChunkStream) writeChunkBasicHeader(w *ReadWriter) error {
	h := cs.format << 6

	switch {
	case cs.csid < 64:
		// csid[0,64)
		err := w.WriteUintBE(h|cs.csid, 1)
		if err != nil {
			return err
		}
	case cs.csid-64 < 256:
		// csid[64,256)
		// 0表示块基本头为2个字节
		err := w.WriteUintBE(h|0x00, 1)
		if err != nil {
			return err
		}

		err = w.WriteUintLE(cs.csid-64, 1)
		if err != nil {
			return err
		}
	case cs.csid-64 < 65536:
		// csid[256,65536)
		// 1表示块基本头为3个字节
		err := w.WriteUintBE(h|0x01, 1)
		if err != nil {
			return err
		}

		err = w.WriteUintLE(cs.csid-64, 2)
		if err != nil {
			return err
		}
	}

	return nil
}

// 根据chunkStream的信息, 构造chunk header并写入到w中
func (cs *ChunkStream) writeHeader(w *ReadWriter) error {
	// Chunk Basic Header
	err := cs.writeChunkBasicHeader(w)
	if err != nil {
		return err
	}

	// Chunk Message Header
	ts := cs.timestamp
	if ts > 0xffffff {
		ts = 0xffffff
	}

	// 块类型3的消息头最多含有扩展时间戳
	if cs.format == 3 {
		goto END
	}

	// 块类型2至少含有时间增量
	err = w.WriteUintBE(ts, 3)
	if err != nil {
		return err
	}

	if cs.format == 2 {
		goto END
	}

	// 在块类型2基础上, 块类型1至少含有消息长度和消息类型
	if cs.length > 0xffffff {
		return fmt.Errorf("overflow chunk length=%d", cs.length)
	}

	err = w.WriteUintBE(cs.length, 3)
	if err != nil {
		return err
	}

	err = w.WriteUintBE(cs.typeID, 1)
	if err != nil {
		return err
	}

	if cs.format == 1 {
		goto END
	}

	// 在块类型1基础上, 块类型0至少含有流ID
	err = w.WriteUintLE(cs.streamID, 4)
	if err != nil {
		return err
	}

END:
	// Extended Timestamp
	if ts >= 0xffffff {
		err = w.WriteUintBE(cs.timestamp, 4)
		if err != nil {
			return err
		}
	}

	return nil
}

// 构造chunk, 将chunk写入缓存中，chunk-0后面跟着chunk-3。chunkSize: 每个chunk除header外的最大负载
func (cs *ChunkStream) writeChunk(w *ReadWriter, chunkSize uint32) error {
	if chunkSize <= 0 {
		return errors.New("chunk size<=0")
	}

	switch cs.typeID {
	case packet.TagAudio:
		cs.csid = 4
	case packet.TagVideo, packet.TagScriptDataAMF0, packet.TagScriptDataAMF3:
		cs.csid = 6
	}

	// 已写入字节数
	var writtenLen uint32

	// 字节增量
	var increment int

	// 分块的总数
	totalNumOfChunks := cs.length / chunkSize

	// 建立并写一个chunk
	for i := 0; i <= int(totalNumOfChunks); i++ {
		// 如果已经写完, 则退出
		if writtenLen >= cs.length {
			break
		}

		// chunk类型0后跟着chunk类型3
		if i == 0 {
			cs.format = uint32(0)
		} else {
			cs.format = uint32(3)
		}

		// 接入chunk header
		err := cs.writeHeader(w)
		if err != nil {
			return err
		}

		// 计算得到真实的chunk负载
		increment = int(chunkSize)
		start := i * increment
		if len(cs.data)-start <= increment {
			increment = len(cs.data) - start
		}

		if increment < 0 {
			return errors.New("overflow chunk size")
		}

		// 写入chunk数据
		end := start + increment
		buf := cs.data[start:end]

		_, err = w.Write(buf)
		if err != nil {
			return err
		}

		// 已写入字节数
		writtenLen += uint32(increment)
	}

	if writtenLen != cs.length {
		return errors.New("incomplete chunk")
	}

	return nil
}

// 接收到chunk后，对metadata进行处理
func (cs *ChunkStream) unpack() error {
	var err error

	// 数据消息, 客户端用这个消息向对端发送 Metadata 或者任意的用户数据
	switch cs.typeID {
	case packet.TagScriptDataAMF0:
		// 从[SetDataFrame, data]数据中提取出data(SetDataFrame是通讯指令, 在rtmp场景外, 应该剔除在外, 只保留数据)
		cs.data, err = amf.DelMetaHeader(cs.data, amf.NewEnDecAMF0())
		if err != nil {
			return err
		}

		// 重新计算长度
		cs.length = uint32(len(cs.data))
	case packet.TagScriptDataAMF3:
		// 从[SetDataFrame, data]数据中提取出data(SetDataFrame是通讯指令, 在rtmp场景外, 应该剔除在外, 只保留数据)
		cs.data, err = amf.DelMetaHeader(cs.data, amf.NewEnDecAMF3())
		if err != nil {
			return err
		}

		// 重新计算长度
		cs.length = uint32(len(cs.data))
	}

	return nil
}
