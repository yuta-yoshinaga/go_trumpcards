//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBriscolaPresenter ブリスコラプレゼンターモック
type MockBriscolaPresenter struct {
	MockGamePresenter[interfaces.BriscolaGame]
}

// HintOutput モック
func (_m *MockBriscolaPresenter) HintOutput(b interfaces.BriscolaGame) string {
	ret := _m.Called(b)
	return ret.Get(0).(string)
}
