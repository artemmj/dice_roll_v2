package generators

import (
	"math/rand"
)

type MathRandGenerator struct{}

func (g *MathRandGenerator) Generate(seed []byte) int32 {
	return rand.Int31n(6) + 1
}

func (g *MathRandGenerator) Name() string {
	return "MathRandGenerator"
}
