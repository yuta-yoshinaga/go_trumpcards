//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPenguinPresenter ペンギンプレゼンターモック
type MockPenguinPresenter struct {
	MockGamePresenter[interfaces.PenguinGame]
}

// HintOutput モック
func (_m *MockPenguinPresenter) HintOutput(p interfaces.PenguinGame) string {
	ret := _m.Called(p)
	return ret.Get(0).(string)
}
