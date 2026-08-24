//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCoinchePresenter コワンシュプレゼンターモック
type MockCoinchePresenter struct {
	MockGamePresenter[interfaces.CoincheGame]
}

// HintOutput モック
func (_m *MockCoinchePresenter) HintOutput(b interfaces.CoincheGame) string {
	ret := _m.Called(b)
	return ret.Get(0).(string)
}
