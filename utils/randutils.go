package utils

type Rand struct {
	inc, last uint64
}

func (r *Rand) Next(max uint32) uint32 {
	r.last ^= (r.last << 21)
	r.last ^= (r.last >> 29)
	r.last ^= (r.last << 4)
	r.inc += 123456789123456789
	out := uint32((r.last + r.inc) % uint64(max))
	return out
}
