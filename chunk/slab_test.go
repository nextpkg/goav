package chunk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const max = 128 * 1024
const min = max / 4

func TestNewSlab(t *testing.T) {
	at := assert.New(t)

	s := NewSlab()
	s.Init(min, max)
	at.NotNil(s)

	need1 := 10
	buf := s.Get(need1)
	at.Equal(need1, len(buf))
	at.Equal(need1, s.pos)
	at.Equal(max, len(s.buf))

	need2 := 50 * 1024
	buf = s.Get(need2)
	at.Equal(need2, len(buf))
	at.Equal(need1+need2, s.pos)
	at.Equal(max, len(s.buf))

	need3 := 80 * 1024
	buf = s.Get(need3)
	at.Equal(need3, len(buf))
	at.Equal(need1+need2, s.pos)
	at.Equal(max, len(s.buf))

	need4 := 200 * 1024
	buf = s.Get(need4)
	at.Equal(need4, len(buf))
	at.Equal(need1+need2, s.pos)
	at.Equal(max, len(s.buf))
}

func allocSlab(b *testing.B, size int) {
	b.RunParallel(func(pb *testing.PB) {
		var buf []byte

		s := NewSlab()
		s.Init(min, max)

		for pb.Next() {
			buf = s.Get(size * 1024)
		}

		_ = buf
	})
}

func mallocDirect(b *testing.B, size int) {
	b.RunParallel(func(pb *testing.PB) {
		var buf []byte

		for pb.Next() {
			buf = make([]byte, size*1024)
		}

		_ = buf
	})
}

func BenchmarkNewSlab_30K_direct(b *testing.B) {
	mallocDirect(b, 30)
}

func BenchmarkNewSlab_30K_slab(b *testing.B) {
	allocSlab(b, 30)
}

func BenchmarkNewSlab_60K_direct(b *testing.B) {
	mallocDirect(b, 60)
}

func BenchmarkNewSlab_60K_slab(b *testing.B) {
	allocSlab(b, 60)
}

func BenchmarkNewSlab_90K_direct(b *testing.B) {
	mallocDirect(b, 90)
}

func BenchmarkNewSlab_90K_slab(b *testing.B) {
	allocSlab(b, 60)
}
