package packet

// Types 缓存媒体元数据类型(音频或者是视频)
type Types struct {
	types byte
}

// NewTypes 缓存媒体元数据类型
func NewTypes() *Types {
	return &Types{
		types: 0,
	}
}

// Reset 重置类型
func (mt *Types) Reset() {
	mt.types = 0
}

// IsVideo 标记为视频
func (mt *Types) IsVideo() {
	mt.types |= 0x1
}

// IsAudio 标记为音频
func (mt *Types) IsAudio() {
	mt.types |= 0x2
}

// ToSlice 将缓存的媒体元素类型转换为包类型
func (mt *Types) ToSlice() (types []int) {
	if mt.types&0x1 == 1 {
		types = append(types, PktVideo)
	}
	if mt.types&0x2 == 2 {
		types = append(types, PktAudio)
	}

	return types
}
