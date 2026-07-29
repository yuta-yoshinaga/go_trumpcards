//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTerracePresenter テラス プレゼンターモック
type MockTerracePresenter struct {
	MockGamePresenter[interfaces.TerraceGame]
}

// HintOutput モック
func (_m *MockTerracePresenter) HintOutput(t interfaces.TerraceGame) string {
	ret := _m.Called(t)
	return ret.Get(0).(string)
}
