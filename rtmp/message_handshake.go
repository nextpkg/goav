// Package core rtmp握手过程
package core

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"

	"git.code.oa.com/trpc-go/trpc-go/log"
	"github.com/pkg/errors"
)

// 握手过程的超时时间
const handshakeTimeout = 2 * time.Second

var (
	clientFullKey = []byte{
		'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
		'F', 'l', 'a', 's', 'h', ' ', 'P', 'l', 'a', 'y', 'e', 'r', ' ',
		'0', '0', '1', // partial key
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
		0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
	serverFullKey = []byte{
		'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
		'F', 'l', 'a', 's', 'h', ' ', 'M', 'e', 'd', 'i', 'a', ' ',
		'S', 'e', 'r', 'v', 'e', 'r', ' ',
		'0', '0', '1', // partial key
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
		0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}

	clientPartialKey = clientFullKey[:30]
	serverPartialKey = serverFullKey[:36]
)

// 握手
func (c *Conn) write(data []byte) error {

	// 将协议数据写入网络缓冲中
	n, err := c.rw.Write(data)
	if err != nil {
		return errors.Wrapf(err, "write failed,wrote=%d", n)
	}

	// 使缓冲的数据写入网络通道
	err = c.rw.Flush()
	if err != nil {
		return errors.Wrap(err, "flush failed")
	}

	return nil
}

// 握手
func (c *Conn) read(data []byte) error {

	// 从网络缓冲中读取对端数据
	n, err := c.rw.Read(data)
	if err != nil {
		return errors.Wrapf(err, "read failed,read=%d", n)
	}

	return nil
}

func (c *Conn) resetTimer() {
	// reset deadline for client handshake
	err := c.Conn.SetDeadline(time.Time{})
	if err != nil {
		log.Error(errors.Wrap(err, "reset connection timer failed"))
	}
}

// HandshakeClient 客户端发起握手, 只支持简单握手
func (c *Conn) HandshakeClient() error {

	err := c.Conn.SetDeadline(time.Now().Add(handshakeTimeout))
	if err != nil {
		return errors.Wrap(err, "set deadline failed")
	}
	defer c.resetTimer()

	var random [(1 + 1536*2) * 2]byte
	C0C1C2 := random[:1536*2+1]
	S0S1S2 := random[1536*2+1:]

	C0 := C0C1C2[:1]
	C1 := C0C1C2[1 : 1536+1]
	// C2 := C0C1C2[1536+1:]

	S0 := S0S1S2[:1]
	S1 := S0S1S2[1 : 1536+1]

	// 构造C0C1
	C0[0] = 3                              // 客户端版本
	binary.BigEndian.PutUint32(C1[0:4], 0) // 这个字段包含了一个时间戳，它可能(SHOULD)作为所有本端的后续块的起始时间点
	binary.BigEndian.PutUint32(C1[4:8], 0) // 简单握手, 这个字段必需是全0

	// 这个字段包含了任意值, 应该(SHOULD)足够的随机, 也不必是加密的随机数，或者是动态数据
	var n int
	n, err = rand.Read(C1[8:])
	if err != nil {
		return errors.Wrapf(err, "get random number failed,read=%d", n)
	}

	// -> C0C1
	C0C1 := C0C1C2[:1536+1]
	if err := c.write(C0C1); err != nil {
		return errors.Wrap(err, "send C0C1")
	}

	// <- S0S1S2
	if err := c.read(S0S1S2); err != nil {
		return errors.Wrap(err, "receive S0S1S2")
	}

	if S0[0] != 3 {
		return fmt.Errorf("expected rtmp version 3, but got %d", S0[0])
	}

	// 简单握手 -> C2
	C2 := S1
	if err := c.write(C2); err != nil {
		return errors.Wrap(err, "send C2")
	}

	log.Trace("client handshake is finished")
	return nil
}

// HandshakeServer 服务端响应握手, 支持复杂握手和简单握手
func (c *Conn) HandshakeServer() error {

	err := c.Conn.SetDeadline(time.Now().Add(handshakeTimeout))
	if err != nil {
		return errors.Wrap(err, "set deadline")
	}
	defer c.resetTimer()

	var random [(1 + 1536*2) * 2]byte
	C0C1C2 := random[:1536*2+1]
	S0S1S2 := random[1536*2+1:]

	C0 := C0C1C2[:1]
	C1 := C0C1C2[1 : 1536+1]
	C2 := C0C1C2[1536+1:]

	S0 := S0S1S2[:1]
	S1 := S0S1S2[1 : 1536+1]
	S2 := S0S1S2[1536+1:]

	// <- C0C1
	C0C1 := C0C1C2[:1536+1]
	err = c.read(C0C1)
	if err != nil {
		return errors.Wrap(err, "read C0C1")
	}

	if C0[0] != 3 {
		return fmt.Errorf("invalid rtmp version=%d ", C0[0])
	}

	// 构造S0S1S2
	S0[0] = 3
	cTime := binary.BigEndian.Uint32(C1[0:4])    // 这个字段包含了一个时间戳，它可能(SHOULD)作为所有本端的后续块的起始时间点
	cVersion := binary.BigEndian.Uint32(C1[4:8]) // 0=简单握手, 其它=复杂握手
	if cVersion != 0 {

		// 复杂握手
		sTime := cTime
		sVersion := uint32(0x04050001)

		// 用client key计算C1的摘要, 再用server key计算C1摘要的摘要并返回
		digest, err := getDigest(C1, clientPartialKey, serverFullKey)
		if err != nil {
			return errors.Wrap(err, "calculate C1 digest")
		}

		// 计算S1的签名, 填充到S1中
		err = createDigestC1S1(S1, sTime, sVersion, serverPartialKey)
		if err != nil {
			return errors.Wrap(err, "calculate C1S1 digest")
		}

		// 计算S2的签名, 填充到S2中
		err = createDigestC2S2(S2, digest)
		if err != nil {
			return errors.Wrap(err, "calculate C2S2 digest")
		}
	} else {

		// 简单握手
		copy(S1, C2)
		copy(S2, C1)
	}

	// -> S0S1S2
	if err := c.write(S0S1S2); err != nil {
		return errors.Wrap(err, "send S0S1S2")
	}

	// <- C2
	if err := c.read(C2); err != nil {
		return errors.Wrap(err, "receive C2")
	}

	log.Trace("server handshake is finished")
	return nil
}

// 用client key计算C1/S1的摘要, 再用server key计算C1/S1摘要的摘要并返回
func getDigest(c1s1 []byte, clientKey []byte, serverKey []byte) ([]byte, error) {

	// schema0: |time: 4bytes|version: 4bytes|key: 764bytes|digest: 764bytes|
	digestDataPos, err := findDigestPos(c1s1, clientKey, 772)
	if err != nil {
		return nil, errors.Wrap(err, "schema0 find digest data position failed")
	}

	// schema1: |time: 4bytes|version: 4bytes|digest: 764bytes|key: 764bytes|
	if digestDataPos == -1 {

		digestDataPos, err = findDigestPos(c1s1, clientKey, 8)
		if digestDataPos == -1 {
			return nil, errors.Wrap(err, "schema1 find digest data position failed")
		}
	}

	// 根据客户端数据摘要, 用服务端的key重新计算摘要
	return createDigest(serverKey, c1s1[digestDataPos:digestDataPos+32], -1)
}

// 构造C1/S1的签名
func createDigestC1S1(c1s1 []byte, time uint32, version uint32, serverKey []byte) error {

	// 随机值
	n, err := rand.Read(c1s1[8:])
	if err != nil {
		return errors.Wrapf(err, "read C1S1 random data failed,length=%d", n)
	}

	binary.BigEndian.PutUint32(c1s1[0:4], time)
	binary.BigEndian.PutUint32(c1s1[4:8], version)

	// schema1:
	// time: 4bytes
	// version: 4bytes
	// digest: 764bytes
	// key: 764bytes
	digestDataPos := getDigestDataPos(c1s1, 8)

	var digest []byte
	digest, err = createDigest(serverKey, c1s1, digestDataPos)
	if err != nil {
		return errors.Wrap(err, "create C1S1 digest failed")
	}

	copy(c1s1[digestDataPos:], digest)
	return nil
}

// 构造C2/S2的签名
// 结构:
// 	random-data: 1504bytes
// 	digest-data: 32
func createDigestC2S2(c2s2 []byte, key []byte) error {

	n, err := rand.Read(c2s2)
	if err != nil {
		return errors.Wrapf(err, "read C2S2 random data failed,length=%d", n)
	}

	digestDataPos := len(c2s2) - 32

	var digest []byte
	digest, err = createDigest(key, c2s2, digestDataPos)
	if err != nil {
		return errors.Wrap(err, "calculate digest")
	}

	copy(c2s2[digestDataPos:], digest)
	return nil
}

// 找到digest-data的位置, 找不到则返回-1
// digest结构:
// 	offset: 4 bytes
// 	random-data: (offset)bytes
// 	digest-data: 32bytes
// 	random-data: (764-4-offset-32)bytes
func findDigestPos(c1s1 []byte, clientKey []byte, base int) (int, error) {

	// 试图找到digest-data的位置
	digestDataPos := getDigestDataPos(c1s1, base)

	// 如果不匹配, 说明找到的不是digest数据, 而是key数据
	digest, err := createDigest(clientKey, c1s1, digestDataPos)
	if err != nil {
		return 0, errors.Wrap(err, "calculate digest")
	}

	// 比较数字签名
	compare := bytes.Compare(c1s1[digestDataPos:digestDataPos+32], digest)
	if compare != 0 {
		return -1, nil
	}

	return digestDataPos, nil
}

// 获取digest数据的位置
// digest结构:
// 	offset: 4 bytes
// 	random-data: (offset)bytes
// 	digest-data: 32bytes
// 	random-data: (764-4-offset-32)bytes
func getDigestDataPos(p []byte, base int) int {

	pos := 0

	for i := 0; i < 4; i++ {
		pos += int(p[base+i])
	}

	return (pos % 728) + base + 4
}

// 使用key计算src的摘要
func createDigest(key []byte, src []byte, digestDataPos int) ([]byte, error) {

	h := hmac.New(sha256.New, key)

	if digestDataPos <= 0 {

		// 全部数据参与计算摘要
		n, err := h.Write(src)
		if err != nil {
			return nil, errors.Wrapf(err, "write hmac,length=%d", n)
		}
	} else {

		// include offset, random-data
		n, err := h.Write(src[:digestDataPos])
		if err != nil {
			return nil, errors.Wrapf(err, "write hmac include offset and random data,length=%d", n)
		}

		// include random-data
		n, err = h.Write(src[digestDataPos+32:])
		if err != nil {
			return nil, errors.Wrapf(err, "write hmac include random data,length=%d", n)
		}
	}

	return h.Sum(nil), nil
}
