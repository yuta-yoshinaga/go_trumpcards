//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockGuandanPresenter 掼蛋 (Guandan) プレゼンターモック
type MockGuandanPresenter struct {
	MockGamePresenter[interfaces.GuandanGame]
}

// CheckOutput モック
func (_m *MockGuandanPresenter) CheckOutput(g interfaces.GuandanGame, idxs []int) string {
	ret := _m.Called(g, idxs)
	return ret.Get(0).(string)
}
