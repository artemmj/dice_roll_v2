package suite

import (
	"context"
	"net"
	"os"
	"strconv"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	drv1 "dice_roll_v2/gen/go"
	"dice_roll_v2/internal/config"
)

const (
	grpcHost = "localhost"
)

type Suite struct {
	*testing.T                            // Потребуется для вызова методов *testing.T
	Cfg        *config.Config             // Конфигурация приложения
	GameClient drv1.DiceRollGameAPIClient // Клиент для взаимодействия с gRPC-сервером Auth
}

// New creates new test suite.
func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()   // Функция будет восприниматься как вспомогательная для тестов
	t.Parallel() // Разрешаем параллельный запуск тестов

	cfg := config.MustLoadPath(configPath())
	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)
	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})
	cc, err := grpc.NewClient(grpcAddress(cfg), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}
	// gRPC-клиент сервера Auth
	authClient := drv1.NewDiceRollGameAPIClient(cc)
	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		GameClient: authClient,
	}
}

func configPath() string {
	const key = "CONFIG_PATH"
	if v := os.Getenv(key); v != "" {
		return v
	}
	return "../config/local.yaml"
}

func grpcAddress(cfg *config.Config) string {
	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))
}
