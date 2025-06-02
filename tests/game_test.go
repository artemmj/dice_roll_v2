package tests

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	drv1 "dice_roll_v2/gen/go"
	"dice_roll_v2/internal/api"
	"dice_roll_v2/internal/generators"
	"dice_roll_v2/internal/storage"
	"dice_roll_v2/internal/storage/postgres"
	"dice_roll_v2/tests/suite"
)

func TestCryptoGenerator(t *testing.T) {
	seed := []byte{0x01, 0x02, 0x03}
	gen := generators.CryptoHMACGenerator{}
	roll := gen.Generate(seed)
	assert.True(t, roll >= 1 && roll <= 6)
}

func TestNewServiceWithInMemoryStorage(t *testing.T) {
	_, st := suite.New(t)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	inMemStorage := storage.NewInMemoryStorage()
	pgStorage, _ := postgres.New(st.Cfg.PostgresConnStrForDocker)
	service := api.NewService(inMemStorage, *pgStorage, log)
	assert.NotNil(t, service)
}

func createTestSession(t *testing.T, s storage.SessionStorage) string {
	session := &storage.GameSession{
		ID:         uuid.New().String(),
		ServerSeed: "test_server_seed",
		ClientSeed: "test_client_seed",
		Nonce:      0,
		CreatedAt:  time.Now(),
	}
	err := s.Save(context.Background(), session)
	require.NoError(t, err)
	return session.ID
}

func TestPlay_Success(t *testing.T) {
	_, st := suite.New(t)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	inMemStorage := storage.NewInMemoryStorage()
	pgStorage, _ := postgres.New(st.Cfg.PostgresConnStrForDocker)
	service := api.NewService(inMemStorage, *pgStorage, log)

	// Создаем тестовую сессию
	sessionID := createTestSession(t, inMemStorage)

	// Вызываем Play
	req := &drv1.PlayRequest{
		SessionId: sessionID,
	}
	resp, err := service.Play(context.Background(), req)

	// Проверяем результат
	require.NoError(t, err)
	assert.True(t, resp.PlayerRoll >= 1 && resp.PlayerRoll <= 6)
	assert.True(t, resp.ServerRoll >= 1 && resp.ServerRoll <= 6)
}

func TestPlay_SessionNotFound(t *testing.T) {
	_, st := suite.New(t)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	inMemStorage := storage.NewInMemoryStorage()
	pgStorage, _ := postgres.New(st.Cfg.PostgresConnStrForDocker)
	service := api.NewService(inMemStorage, *pgStorage, log)

	req := &drv1.PlayRequest{
		SessionId: "non_existent_session",
	}
	resp, err := service.Play(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}
