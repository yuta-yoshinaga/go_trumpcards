//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockGoFishPresenter Go Fishプレゼンターモック
type MockGoFishPresenter struct {
	MockGamePresenter[interfaces.GoFishGame]
}

// HintOutput モック
func (_m *MockGoFishPresenter) HintOutput(game interfaces.GoFishGame) string {
	ret := _m.Called(game)
	return ret.Get(0).(string)
}
