package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	drgen "dice_roll_v2/gen/go"
	"dice_roll_v2/models"
	"dice_roll_v2/storage"
	"dice_roll_v2/utils"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *DiceRollService) CreateSession(
	ctx context.Context,
	req *drgen.CreateSessionRequest,
) (*drgen.CreateSessionResponse, error) {
	const op = "DiceRollService.CreateSession"
	log := s.log.With("op", op)
	log.Debug("Создаю сессию...")

	clientSeed := req.GetClientSeed()
	if clientSeed == "" {
		return nil, status.Error(codes.InvalidArgument, "Нужен параметр client_seed")
	}

	session := &storage.GameSession{
		ID:         uuid.NewString(),
		ServerSeed: utils.MustGenerateRandomHex(32),
		ClientSeed: clientSeed,
		CreatedAt:  time.Now(),
	}

	if err := s.sessionStorage.Save(ctx, session); err != nil {
		return nil, status.Error(codes.Internal, "Ошибка при попытке созданиия сессии")
	}
	hash := sha256.Sum256([]byte(session.ServerSeed))

	return &drgen.CreateSessionResponse{
		SessionId:      session.ID,
		ServerSeedHash: hex.EncodeToString(hash[:]),
	}, nil
}

func (s *DiceRollService) Play(ctx context.Context, req *drgen.PlayRequest) (*drgen.PlayResponse, error) {
	const op = "DiceRollService.Play"
	log := s.log.With("op", op)
	log.Debug("Начинаю игру...")

	session, err := s.sessionStorage.Get(ctx, req.GetSessionId())
	if err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			return nil, status.Error(codes.NotFound, "Сессия не найдена")
		}
		return nil, status.Error(codes.Internal, "Ошибка при попытке получить сессию")
	}

	// Обновляем nonce
	session.Nonce++
	if err := s.sessionStorage.Update(ctx, session); err != nil {
		return nil, status.Error(codes.Internal, "failed to update session")
	}

	generator := s.selectRandomGenerator()
	seed := s.computeRollSignature(session.ServerSeed, session.ClientSeed, session.Nonce)
	playerRoll := generator.Generate(seed)
	serverRoll := generator.Generate(append(seed, 0x01)) // Добавляем вариативность
	winner := s.determineWinner(playerRoll, serverRoll)
	log.Debug(
		"Данные успешно сгенерированы...",
		slog.Any("playerRoll", playerRoll),
		slog.Any("serverRoll", serverRoll),
		slog.Any("winner", winner),
	)

	_, err = s.storage.SaveGameResults(ctx, s.log, models.GameResult{
		CreatedAt:  time.Now().Format(time.RFC3339),
		ServerRoll: serverRoll,
		PlayerRoll: playerRoll,
		Winner:     winner,
		Roller:     generator.Name(),
	})
	if err != nil {
		s.log.Error("Ошибка при попытке сохранить результат в БД: %v", slog.Any("err", err))
	}

	return &drgen.PlayResponse{
		CreatedAt:     time.Now().Format(time.RFC3339Nano),
		ServerRoll:    serverRoll,
		PlayerRoll:    playerRoll,
		Winner:        winner,
		Roller:        generator.Name(),
		ServerSeed:    session.ServerSeed,
		ClientSeed:    session.ClientSeed,
		Nonce:         int32(session.Nonce),
		GeneratorUsed: generator.Name(),
	}, nil
}

func (s *DiceRollService) VerifyRoll(ctx context.Context, req *drgen.VerifyRequest) (*drgen.VerifyResponse, error) {
	const op = "DiceRollService.VerifyRoll"
	log := s.log.With("op", op)
	log.Debug("Верифицирую результат...")

	generator, exists := s.generators[req.GetGeneratorName()]
	if !exists {
		return nil, status.Error(codes.InvalidArgument, "Неподдерживаемый генератор")
	}

	seed := s.computeRollSignature(
		req.GetServerSeed(),
		req.GetClientSeed(),
		int(req.GetNonce()),
	)
	expected := generator.Generate(seed)
	return &drgen.VerifyResponse{IsValid: expected == req.GetExpectedRoll()}, nil
}
