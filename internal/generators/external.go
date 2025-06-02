package generators

type ExternalAPIGenerator struct{}

func (g *ExternalAPIGenerator) Generate(seed []byte) int32 {
	sum := 0
	for _, b := range seed {
		sum += int(b)
	}
	return int32(sum%6) + 1
}

func (g *ExternalAPIGenerator) Name() string {
	return "ExternalAPIGenerator"
}
