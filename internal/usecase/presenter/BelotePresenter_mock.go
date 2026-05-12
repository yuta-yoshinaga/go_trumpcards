//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBelotePresenter ベロートプレゼンターモック
type MockBelotePresenter struct {
	MockGamePresenter[interfaces.BeloteGame]
}

// HintOutput モック
func (_m *MockBelotePresenter) HintOutput(b interfaces.BeloteGame) string {
	ret := _m.Called(b)
	return ret.Get(0).(string)
}
