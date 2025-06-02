package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	drgen "dice_roll_v2/gen/go"
	"dice_roll_v2/internal/generators"
	"dice_roll_v2/internal/models"
	"dice_roll_v2/internal/storage"
	"dice_roll_v2/internal/storage/postgres"
	"dice_roll_v2/internal/utils"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DiceRollService struct {
	drgen.UnimplementedDiceRollGameAPIServer
	log            *slog.Logger
	storage        postgres.Storage
	sessionStorage storage.SessionStorage
	generators     map[string]generators.RollGenerator
}

type DiceRollGameAPI interface {
	CreateSession(context.Context, *drgen.CreateSessionRequest) (*drgen.CreateSessionResponse, error)
	Play(context.Context, *drgen.PlayRequest) (*drgen.PlayResponse, error)
	VerifyRoll(context.Context, *drgen.VerifyRequest) (*drgen.VerifyResponse, error)
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

// Функция регистрации сервиса в gRPC сервере
func Register(gRPCServer *grpc.Server, drserver drgen.DiceRollGameAPIServer) {
	drgen.RegisterDiceRollGameAPIServer(gRPCServer, drserver)
}

// Функция выбирает рандомно доступный генератор
func (s *DiceRollService) selectRandomGenerator() generators.RollGenerator {
	keys := make([]string, 0, len(s.generators))
	for k := range s.generators {
		keys = append(keys, k)
	}
	return s.generators[keys[rand.Intn(len(keys))]]
}

// Функция считает хэш по переданным сидам
func (s *DiceRollService) computeRollSignature(serverSeed, clientSeed string, nonce int) []byte {
	data := fmt.Sprintf("%s:%s:%d", serverSeed, clientSeed, nonce)
	mac := hmac.New(sha256.New, []byte(serverSeed))
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

// Функция для определения победителя
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
			log.Error("Сессия не найдена!", slog.Any("err", err), slog.Any("session", session))
			return nil, status.Error(codes.NotFound, "Сессия не найдена")
		}
		log.Error("Возникла ошибка при попытке получить сессию!", slog.Any("err", err))
		return nil, status.Error(codes.Internal, "Ошибка при попытке получить сессию")
	}

	// Обновляем nonce
	session.Nonce++
	if err := s.sessionStorage.Update(ctx, session); err != nil {
		return nil, status.Error(codes.Internal, "failed to update session")
	}

	generator := s.selectRandomGenerator()
	log.Debug("Выбран генератор...", slog.Any("generator", generator.Name()))
	seed := s.computeRollSignature(session.ServerSeed, session.ClientSeed, session.Nonce)
	serverRoll := generator.Generate(append(seed, 0x01)) // Добавляем вариативность
	playerRoll := generator.Generate(seed)
	winner := s.determineWinner(playerRoll, serverRoll)
	log.Debug(
		"Игра сыграна, результаты...",
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
