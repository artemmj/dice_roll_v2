package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"dice_roll_v2/internal/config"
)

func main() {
	var migrationsPath, migrationsTable string

	flag.StringVar(&migrationsPath, "migrations-path", "./migrations", "path to migrations")
	// Таблица, в которой будет храниться информация о миграциях. Она нужна
	// для того, чтобы понимать, какие миграции уже применены, а какие нет.
	// Дефолтное значение - 'migrations'.
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")
	flag.Parse() // Выполняем парсинг флагов

	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	cfg := config.MustLoad()
	sourceURL := "file://" + migrationsPath
	// Тут тоже надо поменять строку подключения, чтобы подключиться не в докере
	databaseURL := fmt.Sprintf("%s&x-migrations-table=%s", cfg.PostgresConnStrForDocker, migrationsTable)
	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		panic(err)
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		panic(err)
	}
}
