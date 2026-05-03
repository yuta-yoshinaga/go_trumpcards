//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPitchPresenter ピッチプレゼンターモック
type MockPitchPresenter struct {
	MockGamePresenter[interfaces.PitchGame]
}

// HintOutput モック
func (_m *MockPitchPresenter) HintOutput(p interfaces.PitchGame) string {
	ret := _m.Called(p)
	return ret.Get(0).(string)
}
