//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCegoPresenter チェゴ (Cego) のプレゼンターモック
type MockCegoPresenter struct {
	MockGamePresenter[interfaces.CegoGame]
}

// HintOutput モック
func (_m *MockCegoPresenter) HintOutput(g interfaces.CegoGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
