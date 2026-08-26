//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockJulepePresenter フレペプレゼンターモック
type MockJulepePresenter struct {
	MockGamePresenter[interfaces.JulepeGame]
}

// HintOutput モック
func (_m *MockJulepePresenter) HintOutput(r interfaces.JulepeGame) string {
	return _m.Called(r).Get(0).(string)
}
