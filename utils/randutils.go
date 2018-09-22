package utils

type Rand struct {
	inc, last uint32
}

func (r *Rand) Next(max uint32) uint32 {
	r.last ^= (r.last << 21)
	r.last ^= (r.last >> 29)
	r.last ^= (r.last << 4)
	r.inc += 123456789
	out := uint32((r.last + r.inc) % max)
	return out
}
