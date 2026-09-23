//go:build test

package interfaces

// MockBiribaGame は BiribaGame (= CanastaGame) のモック。型エイリアスにより
// Canasta のモック実装をそのまま再利用する。
type MockBiribaGame = MockCanastaGame
