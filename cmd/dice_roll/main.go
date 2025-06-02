package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"dice_roll_v2/internal/app"
	"dice_roll_v2/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()    // грузим конфиг
	log := setupLogger(cfg.Env) // грузим логгер

	// Тут надо поменять строку подключения если будет работать не в докере
	appication := app.New(log, cfg.PostgresConnStrForDocker, cfg.GRPC.Port)
	go func() {
		appication.MustRun()
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	appication.Stop()
	log.Info("Gracefully stopped")
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
