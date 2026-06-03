//go:build test

package interfaces

// MockBurracoGame は BurracoGame (= CanastaGame) のモック。型エイリアスにより
// Canasta のモック実装をそのまま再利用する。
type MockBurracoGame = MockCanastaGame
