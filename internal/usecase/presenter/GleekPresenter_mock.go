//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockGleekPresenter グリーク (Gleek) のプレゼンターモック
type MockGleekPresenter struct {
	MockGamePresenter[interfaces.GleekGame]
}

// HintOutput モック
func (_m *MockGleekPresenter) HintOutput(g interfaces.GleekGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
