package generators

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
)

type CryptoHMACGenerator struct{}

func (g *CryptoHMACGenerator) Generate(seed []byte) int32 {
	mac := hmac.New(sha256.New, seed)
	mac.Write([]byte("dice_roll"))
	result := mac.Sum(nil)
	number := binary.BigEndian.Uint32(result[:4])
	return int32(number%6) + 1
}

func (g *CryptoHMACGenerator) Name() string {
	return "CryptoHMACGenerator"
}
