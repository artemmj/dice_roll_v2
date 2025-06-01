package generators

type RollGenerator interface {
	Generate(seed []byte) int32 // Возвращает число 1-6
	Name() string               // Уникальное имя генератора
}
