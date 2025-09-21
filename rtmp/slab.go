package core

import "git.code.oa.com/trpc-go/trpc-go/log"

const maxAllocSize = 256 * 1024
const minAllocSize = maxAllocSize / 4

// slabStat 缓存块统计
type slabStat struct {
	max    int
	medium int
	min    int
}

// slabMark 缓存块水位
type slabMark struct {
	max int
	min int
}

// Slab 高速缓存块
type Slab struct {
	pos  int
	buf  []byte
	stat slabStat
	mark slabMark
}

// NewSlab 高速缓存块
func NewSlab() *Slab {
	return &Slab{
		pos: maxAllocSize,
		mark: slabMark{
			max: maxAllocSize,
			min: minAllocSize,
		},
	}
}

// Init 设置分配内存的大小（如需使用，需在Get()被调用前执行）
func (s *Slab) Init(min, max int) {
	if min > max {
		log.Error("min>max")
	}

	if max <= 0 {
		max = maxAllocSize
	}

	if min <= 0 {
		min = max / 4
	}

	s.mark.min = min
	s.mark.max = max
}

// Get 从缓存池子中取出need个byte的空间
func (s *Slab) Get(need int) []byte {
	if need > s.mark.max {
		s.stat.max++
		return make([]byte, need)
	}

	remain := s.mark.max - s.pos
	if need > remain {
		if remain > s.mark.min {
			s.stat.medium++
			return make([]byte, need)
		}

		s.stat.min++
		s.pos = 0
		s.buf = make([]byte, s.mark.max)
	}

	slice := s.buf[s.pos : s.pos+need]
	s.pos += need

	return slice
}
