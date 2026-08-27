package ids

// murmurHash64A 实现 MurmurHash64A 算法
// 从 C++ id.cc 第 74-126 行移植，公有领域算法
func murmurHash64A(key []byte, seed uint64) uint64 {
	const m uint64 = 0xc6a4a7935bd1e995
	const r = 47

	h := seed ^ (uint64(len(key)) * m)

	nblocks := len(key) / 8
	for i := 0; i < nblocks; i++ {
		k := uint64(key[i*8]) |
			uint64(key[i*8+1])<<8 |
			uint64(key[i*8+2])<<16 |
			uint64(key[i*8+3])<<24 |
			uint64(key[i*8+4])<<32 |
			uint64(key[i*8+5])<<40 |
			uint64(key[i*8+6])<<48 |
			uint64(key[i*8+7])<<56

		k *= m
		k ^= k >> r
		k *= m

		h ^= k
		h *= m
	}

	tail := key[nblocks*8:]
	switch len(tail) {
	case 7:
		h ^= uint64(tail[6]) << 48
		fallthrough
	case 6:
		h ^= uint64(tail[5]) << 40
		fallthrough
	case 5:
		h ^= uint64(tail[4]) << 32
		fallthrough
	case 4:
		h ^= uint64(tail[3]) << 24
		fallthrough
	case 3:
		h ^= uint64(tail[2]) << 16
		fallthrough
	case 2:
		h ^= uint64(tail[1]) << 8
		fallthrough
	case 1:
		h ^= uint64(tail[0])
		h *= m
	}

	h ^= h >> r
	h *= m
	h ^= h >> r

	return h
}
