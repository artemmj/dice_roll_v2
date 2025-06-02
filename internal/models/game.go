package models

type GameResult struct {
	CreatedAt  string
	ServerRoll int32
	PlayerRoll int32
	Winner     string
	Roller     string
}
