package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"

	"dice_roll_v2/internal/models"
)

type Storage struct {
	db *sql.DB
}

func New(storageConnStr string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("postgres", storageConnStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) SaveGameResults(
	ctx context.Context,
	log *slog.Logger,
	results models.GameResult,
) (models.GameResult, error) {
	const op = "storage.postgres.SaveGame"
	log = log.With(slog.String("op", op))

	created_at := results.CreatedAt
	server := results.ServerRoll
	player := results.PlayerRoll
	winner := results.Winner
	roller := results.Roller

	insertq := `INSERT INTO game_results (created_at, server, player, winner, roller)
						VALUES (($1), ($2), ($3), ($4), ($5))`
	stmt, err := s.db.Prepare(insertq)
	if err != nil {
		return models.GameResult{}, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, created_at, server, player, winner, roller)
	if err != nil {
		return models.GameResult{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("Данные игры успешно записаны в БД...")
	return models.GameResult{
		CreatedAt:  created_at,
		ServerRoll: server,
		PlayerRoll: player,
		Winner:     winner,
		Roller:     roller,
	}, nil
}
