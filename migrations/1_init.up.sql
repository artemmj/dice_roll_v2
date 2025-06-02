CREATE TABLE IF NOT EXISTS game_results (
    id         SERIAL PRIMARY KEY,
    created_at TEXT    NOT NULL,
    server     INTEGER NOT NULL,
    player     INTEGER NOT NULL,
    winner     TEXT    NOT NULL,
    roller     TEXT    NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_id ON game_results (id);
