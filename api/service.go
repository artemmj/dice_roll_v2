package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"math/rand"

	proto "dice_roll_v2/gen/go"
	"dice_roll_v2/generators"
	"dice_roll_v2/storage"
	"dice_roll_v2/storage/postgres"
)

type DiceRollService struct {
	proto.UnimplementedDiceRollGameAPIServer
	log            *slog.Logger
	storage        postgres.Storage
	sessionStorage storage.SessionStorage
	generators     map[string]generators.RollGenerator
}

func NewService(sessionStorage storage.SessionStorage, storage postgres.Storage, log *slog.Logger) *DiceRollService {
	return &DiceRollService{
		log:            log,
		storage:        storage,
		sessionStorage: sessionStorage,
		generators: map[string]generators.RollGenerator{
			"CryptoHMACGenerator":  &generators.CryptoHMACGenerator{},
			"MathRandGenerator":    &generators.MathRandGenerator{},
			"ExternalAPIGenerator": &generators.ExternalAPIGenerator{},
			"XorshiftGenerator":    &generators.XorshiftGenerator{},
		},
	}
}

func (s *DiceRollService) selectRandomGenerator() generators.RollGenerator {
	keys := make([]string, 0, len(s.generators))
	for k := range s.generators {
		keys = append(keys, k)
	}
	return s.generators[keys[rand.Intn(len(keys))]]
}

func (s *DiceRollService) computeRollSignature(serverSeed, clientSeed string, nonce int) []byte {
	data := fmt.Sprintf("%s:%s:%d", serverSeed, clientSeed, nonce)
	mac := hmac.New(sha256.New, []byte(serverSeed))
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

func (s *DiceRollService) determineWinner(playerRoll, serverRoll int32) string {
	switch {
	case playerRoll > serverRoll:
		return "player"
	case serverRoll > playerRoll:
		return "server"
	default:
		return "draw"
	}
}
