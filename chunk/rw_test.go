package chunk

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Test struct {
	i     int
	value uint32
	bytes []byte
}

var (
	bigEndianTest = []Test{
		{1, 0x01, []byte{0x01}},
		{2, 0x0102, []byte{0x01, 0x02}},
		{3, 0x010203, []byte{0x01, 0x02, 0x03}},
		{4, 0x01020304, []byte{0x01, 0x02, 0x03, 0x04}},
	}
	lowEndianTest = []Test{
		{1, 0x01, []byte{0x01}},
		{2, 0x0102, []byte{0x02, 0x01}},
		{3, 0x010203, []byte{0x03, 0x02, 0x01}},
		{4, 0x01020304, []byte{0x04, 0x03, 0x02, 0x01}},
	}
)

func TestReader(t *testing.T) {
	at := assert.New(t)

	// 准备buf, 一次性读完buf
	buf := bytes.NewBufferString("abc")
	r := NewReadWriter(buf, 1024)
	b := make([]byte, 3)
	n, err := r.Read(b)
	at.Nil(err)
	at.Equal(3, n)

	// buf已被读完，再次读取会失败
	_, err = r.Read(b)
	at.Equal(io.EOF, err)

	// 再次准备buf, 因为目标字符超过已有字符数，读buf会失败
	buf.WriteString("123")
	b = make([]byte, 4)
	n, err = r.Read(b)
	at.Error(io.EOF, err)
	at.Equal(3, n)
}

func TestReaderUintBE(t *testing.T) {
	at := assert.New(t)
	for _, test := range bigEndianTest {
		buf := bytes.NewBuffer(test.bytes)
		r := NewReadWriter(buf, 1024)

		n, err := r.ReadUintBE(test.i)
		at.Equal(nil, err, "test %d", test.i)
		at.Equal(test.value, n, "test %d", test.i)
	}
}

func TestReaderUintLE(t *testing.T) {
	at := assert.New(t)
	for _, test := range lowEndianTest {
		buf := bytes.NewBuffer(test.bytes)
		r := NewReadWriter(buf, 1024)

		n, err := r.ReadUintLE(test.i)
		at.Equal(nil, err, "test %d", test.i)
		at.Equal(test.value, n, "test %d", test.i)
	}
}

func TestWriter(t *testing.T) {
	at := assert.New(t)
	buf := bytes.NewBuffer(nil)
	w := NewReadWriter(buf, 1024)
	b := []byte{1, 2, 3}

	// 写数据会成功
	n, err := w.Write(b)
	at.Nil(err)
	at.Equal(3, n)
}

func TestWriteUintBE(t *testing.T) {
	at := assert.New(t)
	for _, test := range bigEndianTest {
		buf := bytes.NewBuffer(nil)
		r := NewReadWriter(buf, 1024)

		err := r.WriteUintBE(test.value, test.i)
		at.Equal(nil, err, "test %d", test.i)

		err = r.Flush()
		at.Equal(nil, err, "test %d", test.i)
		at.Equal(test.bytes, buf.Bytes(), "test %d", test.i)
	}
}

func TestWriteUintLE(t *testing.T) {
	at := assert.New(t)
	for _, test := range lowEndianTest {
		buf := bytes.NewBuffer(nil)
		r := NewReadWriter(buf, 1024)

		err := r.WriteUintLE(test.value, test.i)
		at.Equal(nil, err, "test %d", test.i)

		err = r.Flush()
		at.Equal(nil, err, "test %d", test.i)
		at.Equal(test.bytes, buf.Bytes(), "test %d", test.i)
	}
}
