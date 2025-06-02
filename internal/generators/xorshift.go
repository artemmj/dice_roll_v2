package generators

import "encoding/binary"

type XorshiftGenerator struct{}

func (g *XorshiftGenerator) Generate(seed []byte) int32 {
	state := binary.BigEndian.Uint64(seed)
	state ^= state << 13
	state ^= state >> 7
	state ^= state << 17
	return int32(state%6) + 1
}

func (g *XorshiftGenerator) Name() string {
	return "XorshiftGenerator"
}
