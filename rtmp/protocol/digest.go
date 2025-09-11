package protocol

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"

	"github.com/pkg/errors"
)

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

// getDigest calculates digest using client key for C1/S1, then recalculates using server key
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

// createDigestC1S1 creates digest for C1/S1
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

// createDigestC2S2 creates digest for C2/S2
// Structure:
//	random-data: 1504bytes
//	digest-data: 32
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

// findDigestPos finds the position of digest-data, returns -1 if not found
// digest structure:
//	offset: 4 bytes
//	random-data: (offset)bytes
//	digest-data: 32bytes
//	random-data: (764-4-offset-32)bytes
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

// getDigestDataPos gets the position of digest data
// digest structure:
//	offset: 4 bytes
//	random-data: (offset)bytes
//	digest-data: 32bytes
//	random-data: (764-4-offset-32)bytes
func getDigestDataPos(p []byte, base int) int {
	pos := 0

	for i := 0; i < 4; i++ {
		pos += int(p[base+i])
	}

	return (pos % 728) + base + 4
}

// createDigest calculates digest using key for src data
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