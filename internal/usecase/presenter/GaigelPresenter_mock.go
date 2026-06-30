//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockGaigelPresenter ガイゲルプレゼンターモック
type MockGaigelPresenter struct {
	MockGamePresenter[interfaces.GaigelGame]
}

// HintOutput モック
func (_m *MockGaigelPresenter) HintOutput(g interfaces.GaigelGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
