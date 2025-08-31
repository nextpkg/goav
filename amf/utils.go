package amf

import (
	"fmt"
	"io"
)

// Array AMF数组接口
type Array []interface{}

// Object AMF对象接口
type Object map[string]interface{}

// TypedObject AMF带类型对象接口
type TypedObject struct {
	Type   string
	Object Object
}

// Decoder 解码器
type Decoder interface {
	Decode(r io.Reader) (interface{}, error)
	DecodeBatch(r io.Reader) ([]interface{}, error)
}

// Encoder 解码器
type Encoder interface {
	Encode(w io.Writer, val interface{}) (int, error)
	EncodeBatch(w io.Writer, args ...interface{}) error
}

// 写1字节
func writeByte(w io.Writer, b byte) error {
	bytes := make([]byte, 1)
	bytes[0] = b
	n, err := writeBytes(w, bytes)
	if err != nil {
		return err
	}

	if n != 1 {
		return fmt.Errorf("write %d byte, got error: %s", n, err)
	}

	return err
}

// 写多字节
func writeBytes(w io.Writer, bytes []byte) (int, error) {
	return w.Write(bytes)
}

// 读1字节
func readByte(r io.Reader) (byte, error) {
	bytes, err := readBytes(r, 1)
	if err != nil {
		return 0x00, err
	}

	return bytes[0], nil
}

// 读多字节
func readBytes(r io.Reader, n int) ([]byte, error) {
	bytes := make([]byte, n)
	m, err := r.Read(bytes)
	if err != nil {
		return bytes, err
	}

	if m != n {
		return bytes, fmt.Errorf("expected %d bytes, but got %d bytes", m, n)
	}

	return bytes, nil
}

// 下一个字节一定是target, 否则返回错误
func nextByteMustBe(r io.Reader, target byte) error {
	// 读取一个字节
	marker, err := readByte(r)
	if err != nil {
		return err
	}

	if marker != target {
		return fmt.Errorf("expected %v got %v", target, marker)
	}

	return nil
}
