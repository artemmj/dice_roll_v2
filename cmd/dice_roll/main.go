package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"dice_roll_v2/api"
	"dice_roll_v2/config"
	drgen "dice_roll_v2/gen/go"
	"dice_roll_v2/storage"
	"dice_roll_v2/storage/postgres"

	"google.golang.org/grpc"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()    // грузим конфиг
	log := setupLogger(cfg.Env) // грузим логгер

	pg_storage, _ := postgres.New(cfg.PostgresConnStr)
	log.Debug("Инициализирован postgres.Storage...")

	sessionStorage := storage.NewInMemoryStorage(log)
	log.Debug("Инициализировано in-memory хранилище...")

	service := api.NewService(sessionStorage, *pg_storage, log)
	server := grpc.NewServer()
	drgen.RegisterDiceRollGameAPIServer(server, service)
	// Запуск сервера
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Error("failed to listen: %v", slog.Any("err", err))
	}

	log.Debug(fmt.Sprintf("Стартовал gRPC сервер на порту %d", cfg.GRPC.Port))
	if err := server.Serve(lis); err != nil {
		log.Error("failed to serve: %v", slog.Any("err", err))
	}
}

// Функция выбирает логгер в зависимости от окружения
func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
