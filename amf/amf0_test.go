package amf

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnDecAMF0_Decode_String(t *testing.T) {
	at := assert.New(t)

	d := new(EnDecAMF0)
	buf := bytes.NewReader([]byte{0x02, 0x00, 0x03, 0x66, 0x6f, 0x6f})

	got, err := d.Decode(buf)
	at.Equal(nil, err)
	at.Equal("foo", got)
}

func TestEnDecAMF0_Decode_Object(t *testing.T) {
	at := assert.New(t)

	d := new(EnDecAMF0)
	buf := bytes.NewReader([]byte{
		0x03, 0x00, 0x03, 0x66,
		0x6f, 0x6f, 0x02, 0x00,
		0x03, 0x62, 0x61, 0x72,
		0x00, 0x00, 0x09})

	got, err := d.Decode(buf)
	at.Equal(nil, err)

	obj, ok := got.(Object)
	at.Equal(true, ok)
	at.Equal("bar", obj["foo"])
}

func TestEnDecAMF0_Decode_Number(t *testing.T) {
	at := assert.New(t)

	dec := new(EnDecAMF0)
	buf := bytes.NewReader([]byte{
		0x00, 0x3f, 0xf3, 0x33,
		0x33, 0x33, 0x33, 0x33,
		0x33})

	got, err := dec.Decode(buf)
	at.Equal(nil, err)
	at.Equal(float64(1.2), got)
}

func TestEnDecAMF0_Decode_True(t *testing.T) {
	at := assert.New(t)

	dec := new(EnDecAMF0)
	buf := bytes.NewReader([]byte{0x01, 0x01})

	got, err := dec.Decode(buf)
	at.Equal(nil, err)
	at.Equal(true, got)
}

func TestEnDecAMF0_Decode_False(t *testing.T) {
	at := assert.New(t)

	dec := new(EnDecAMF0)
	buf := bytes.NewReader([]byte{0x01, 0x00})

	got, err := dec.Decode(buf)
	at.Equal(nil, err)
	at.Equal(false, got)
}

func TestEnDecAMF0_Decode_Null(t *testing.T) {
	at := assert.New(t)

	dec := new(EnDecAMF0)
	buf := bytes.NewReader([]byte{0x05})

	got, err := dec.Decode(buf)
	at.Equal(nil, err)
	at.Equal(nil, got)
}

func TestEnDecAMF0_Decode_Undefined(t *testing.T) {
	at := assert.New(t)

	dec := new(EnDecAMF0)
	buf := bytes.NewReader([]byte{0x06})

	got, err := dec.Decode(buf)
	at.Equal(nil, err)
	at.Equal(nil, got)
}

func TestEnDecAMF0_Decode_Reference(t *testing.T) {
	at := assert.New(t)

	dec := new(EnDecAMF0)
	buf := bytes.NewReader([]byte{
		0x03, 0x00, 0x03, 0x66,
		0x6f, 0x6f, 0x07, 0x00,
		0x00, 0x00, 0x00, 0x09})

	got, err := dec.Decode(buf)
	at.Equal(nil, err)

	obj, ok := got.(Object)
	at.Equal(true, ok)

	_, ok2 := obj["foo"].(Object)
	at.Equal(true, ok2)
}

func TestEnDecAMF0_Decode_EcmaArray(t *testing.T) {
	at := assert.New(t)

	dec := new(EnDecAMF0)
	buf := bytes.NewReader([]byte{
		0x08, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x03, 0x66,
		0x6f, 0x6f, 0x02, 0x00,
		0x03, 0x62, 0x61, 0x72,
		0x00, 0x00, 0x09})

	// Test main interface
	got, err := dec.Decode(buf)
	at.Equal(nil, err)

	obj, ok := got.(Object)
	at.Equal(true, ok)
	at.Equal("bar", obj["foo"])
}

func TestEnDecAMF0_Decode_StrictArray(t *testing.T) {
	at := assert.New(t)

	dec := new(EnDecAMF0)
	buf := bytes.NewReader([]byte{
		0x0a, 0x00, 0x00, 0x00,
		0x03, 0x00, 0x40, 0x14,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x02, 0x00,
		0x03, 0x66, 0x6f, 0x6f,
		0x05})

	got, err := dec.Decode(buf)
	at.Equal(nil, err)

	arr, ok := got.(Array)
	at.Equal(true, ok)
	at.Equal(float64(5), arr[0])
	at.Equal("foo", arr[1])
	at.Equal(nil, arr[2])
}

func TestEnDecAMF0_Decode_Date(t *testing.T) {
	at := assert.New(t)

	dec := new(EnDecAMF0)
	buf := bytes.NewReader([]byte{
		0x0b, 0x40, 0x14, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00})

	got, err := dec.Decode(buf)
	at.Equal(nil, err)
	at.Equal(float64(5), got)
}

func TestEnDecAMF0_Decode_LongString(t *testing.T) {
	at := assert.New(t)

	dec := new(EnDecAMF0)
	buf := bytes.NewReader([]byte{
		0x0c, 0x00, 0x00, 0x00,
		0x03, 0x66, 0x6f, 0x6f})

	got, err := dec.Decode(buf)
	at.Equal(nil, err)
	at.Equal("foo", got)
}

func TestEnDecAMF0_Decode_Unsupported(t *testing.T) {
	at := assert.New(t)

	dec := new(EnDecAMF0)
	buf := bytes.NewReader([]byte{0x0d})

	got, err := dec.Decode(buf)
	at.Equal(nil, err)
	at.Equal(nil, got)
}

func TestEnDecAMF0_Decode_XmlDocument(t *testing.T) {
	at := assert.New(t)

	dec := new(EnDecAMF0)
	buf := bytes.NewReader([]byte{
		0x0f, 0x00, 0x00, 0x00,
		0x03, 0x66, 0x6f, 0x6f})

	got, err := dec.Decode(buf)
	at.Equal(nil, err)
	at.Equal("foo", got)
}

func TestEnDecAMF0_Decode_TypedObject(t *testing.T) {
	at := assert.New(t)

	dec := new(EnDecAMF0)
	buf := bytes.NewReader([]byte{
		0x10, 0x00, 0x0F, 'o',
		'r', 'g', '.', 'a',
		'm', 'f', '.', 'A',
		'S', 'C', 'l', 'a',
		's', 's', 0x00, 0x03,
		'b', 'a', 'z', 0x05,
		0x00, 0x03, 'f', 'o',
		'o', 0x02, 0x00, 0x03,
		'b', 'a', 'r', 0x00,
		0x00, 0x09,
	})

	got, err := dec.Decode(buf)
	at.Equal(nil, err)

	tobj, ok := got.(TypedObject)
	at.Equal(true, ok)
	at.Equal("org.amf.ASClass", tobj.Type)
	at.Equal("bar", tobj.Object["foo"])
	at.Equal(nil, tobj.Object["baz"])
}

// ========================================================

func TestEnDecAMF0_Encode_Number(t *testing.T) {
	at := assert.New(t)

	enc := new(EnDecAMF0)
	buf := new(bytes.Buffer)

	n, err := enc.Encode(buf, float64(1.2))
	at.Equal(nil, err)
	at.Equal(9, n)

	expect := []byte{
		numberMarkerOnAMF0, 0x3f, 0xf3, 0x33,
		0x33, 0x33, 0x33, 0x33,
		0x33}
	at.Equal(expect, buf.Bytes())
}

func TestEnDecAMF0_Encode_True(t *testing.T) {
	at := assert.New(t)

	enc := new(EnDecAMF0)
	buf := new(bytes.Buffer)

	n, err := enc.Encode(buf, true)
	at.Equal(nil, err)
	at.Equal(2, n)

	expect := []byte{booleanMarkerOnAMF0, 0x01}
	at.Equal(expect, buf.Bytes())
}

func TestEnDecAMF0_Encode_False(t *testing.T) {
	at := assert.New(t)

	enc := new(EnDecAMF0)
	buf := new(bytes.Buffer)

	n, err := enc.Encode(buf, false)
	at.Equal(nil, err)
	at.Equal(2, n)

	expect := []byte{booleanMarkerOnAMF0, 0x00}
	at.Equal(expect, buf.Bytes())
}

func TestEnDecAMF0_Encode_String(t *testing.T) {
	at := assert.New(t)

	enc := new(EnDecAMF0)
	buf := new(bytes.Buffer)

	n, err := enc.Encode(buf, "foo")
	at.Equal(nil, err)
	at.Equal(6, n)

	expect := []byte{
		stringMarkerOnAMF0, 0x00, 0x03, 0x66,
		0x6f, 0x6f}
	at.Equal(expect, buf.Bytes())
}

func TestEnDecAMF0_Encode_Object(t *testing.T) {
	at := assert.New(t)

	enc := new(EnDecAMF0)
	buf := new(bytes.Buffer)

	obj := make(Object)
	obj["foo"] = "bar"

	n, err := enc.Encode(buf, obj)
	at.Equal(nil, err)
	at.Equal(15, n)

	expect := []byte{
		objectMarkerOnAMF0, 0x00, 0x03, 0x66,
		0x6f, 0x6f, 0x02, 0x00,
		0x03, 0x62, 0x61, 0x72,
		0x00, 0x00, 0x09}
	at.Equal(expect, buf.Bytes())
}

func TestEnDecAMF0_Encode_EcmaArray(t *testing.T) {
	at := assert.New(t)

	enc := new(EnDecAMF0)
	buf := new(bytes.Buffer)

	obj := make(Object)
	obj["foo"] = "bar"

	_, err := enc.EncodeEcmaArray(buf, obj)
	at.Equal(nil, err)

	expect := []byte{
		ecmaArrayMarkerOnAMF0, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x03, 0x66,
		0x6f, 0x6f, 0x02, 0x00,
		0x03, 0x62, 0x61, 0x72,
		0x00, 0x00, 0x09}
	at.Equal(expect, buf.Bytes())
}

func TestEnDecAMF0_Encode_StrictArray(t *testing.T) {
	at := assert.New(t)

	enc := new(EnDecAMF0)
	buf := new(bytes.Buffer)

	arr := make(Array, 3)
	arr[0] = float64(5)
	arr[1] = "foo"
	arr[2] = nil

	_, err := enc.encodeStrictArray(buf, arr)
	at.Equal(nil, err)

	expect := []byte{
		strictArrayMarkerOnAMF0, 0x00, 0x00, 0x00,
		0x03, 0x00, 0x40, 0x14,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x02, 0x00,
		0x03, 0x66, 0x6f, 0x6f,
		0x05}
	at.Equal(expect, buf.Bytes())
}

func TestEnDecAMF0_Encode_Null(t *testing.T) {
	at := assert.New(t)

	enc := new(EnDecAMF0)
	buf := new(bytes.Buffer)

	n, err := enc.Encode(buf, nil)
	at.Equal(nil, err)
	at.Equal(1, n)

	expect := []byte{nullMarkerOnAMF0}
	at.Equal(expect, buf.Bytes())
}

func TestEnDecAMF0_Encode_LongString(t *testing.T) {
	at := assert.New(t)

	enc := new(EnDecAMF0)
	buf := new(bytes.Buffer)

	// 构造大字符串
	testBytes := []byte("12345678")
	tbuf := new(bytes.Buffer)
	for i := 0; i < 65536; i++ {
		tbuf.Write(testBytes)
	}

	_, err := enc.Encode(buf, string(tbuf.Bytes()))
	at.Equal(nil, err)

	// 校验标记位
	mbuf := make([]byte, 1)
	_, err = buf.Read(mbuf)
	at.Equal(nil, err)
	at.Equal(uint8(longStringMarkerOnAMF0), mbuf[0])

	// 校验字符串长度
	var length uint32
	err = binary.Read(buf, binary.BigEndian, &length)
	at.Equal(nil, err)
	at.Equal(uint32(65536*8), length)

	// 校验字符串内容
	tmpBuf := make([]byte, 8)
	counter := 0
	for buf.Len() > 0 {
		n, err := buf.Read(tmpBuf)
		at.Equal(nil, err)
		at.Equal(8, n)
		at.Equal(testBytes, tmpBuf)

		counter++
	}

	// 校验是实际长度
	at.Equal(65536, counter)
}

// ========================================================

func encodeAndDecodeAMF0(val interface{}, at *assert.Assertions) (result interface{}, err error) {
	ed := NewEnDecAMF0()

	buf := new(bytes.Buffer)
	_, err = ed.Encode(buf, val)
	at.Equal(nil, err)

	result, err = ed.Decode(buf)
	at.Equal(nil, err)

	return
}

func compareAMF0(val interface{}, t *testing.T) {
	at := assert.New(t)

	result, err := encodeAndDecodeAMF0(val, at)
	at.Equal(nil, err)
	at.Equal(val, result)
}

func TestEnDecAMF0_Number(t *testing.T) {
	compareAMF0(float64(3.14159), t)
	compareAMF0(float64(124567890), t)
	compareAMF0(float64(-34.2), t)
}

func TestEnDecAMF0_String(t *testing.T) {
	compareAMF0("a pup!", t)
	compareAMF0("日本語", t)
}

func TestEnDecAMF0_Bool(t *testing.T) {
	compareAMF0(true, t)
	compareAMF0(false, t)
}

func TestEnDecAMF0_Null(t *testing.T) {
	compareAMF0(nil, t)
}

func TestEnDecAMF0_Object(t *testing.T) {
	at := assert.New(t)

	obj := make(Object)
	obj["dog"] = "alfie"
	obj["coffee"] = true
	obj["drugs"] = false
	obj["pi"] = 3.14159

	res, err := encodeAndDecodeAMF0(obj, at)
	at.Equal(nil, err)

	result, ok := res.(Object)
	at.True(ok)
	at.Equal(obj, result)
}

func TestEnDecAMF0_Array(t *testing.T) {
	at := assert.New(t)

	arr := [5]float64{1, 2, 3, 4, 5}
	res, err := encodeAndDecodeAMF0(arr, at)
	at.Equal(nil, err)

	result, ok := res.(Array)
	at.True(ok)

	for i := 0; i < len(arr); i++ {
		at.Equal(arr[i], result[i])
	}
}

// ========================================================

func TestEnDecAMF0_RealCase1(t *testing.T) {
	at := assert.New(t)

	buf := bytes.NewBuffer([]byte{
		0x02, 0x00, 0x07, 0x70, 0x75, 0x62, 0x6c,
		0x69, 0x73, 0x68, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x05, 0x02, 0x00, 0x09,
		0x63, 0x61, 0x6d, 0x73, 0x74, 0x72, 0x65, 0x61,
		0x6d, 0x02, 0x00, 0x04, 0x6c, 0x69, 0x76, 0x65,
	})
	expect := []interface{}([]interface{}{"publish", float64(0), interface{}(nil), "camstream", "live"})

	ret, err := NewEnDecAMF0().DecodeBatch(buf)
	at.Nil(err)
	at.Equal(expect, ret)
}
