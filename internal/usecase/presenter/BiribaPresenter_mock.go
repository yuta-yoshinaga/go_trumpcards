//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBiribaPresenter ビリバプレゼンターモック
type MockBiribaPresenter struct {
	MockGamePresenter[interfaces.BiribaGame]
}

// HintOutput モック
func (_m *MockBiribaPresenter) HintOutput(g interfaces.BiribaGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
