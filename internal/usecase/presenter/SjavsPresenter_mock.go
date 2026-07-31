//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSjavsPresenter シャウス プレゼンターモック
type MockSjavsPresenter struct {
	MockGamePresenter[interfaces.SjavsGame]
}

// HintOutput モック
func (_m *MockSjavsPresenter) HintOutput(c interfaces.SjavsGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
