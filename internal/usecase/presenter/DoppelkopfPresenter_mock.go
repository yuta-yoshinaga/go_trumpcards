//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockDoppelkopfPresenter ドッペルコップのプレゼンターモック
type MockDoppelkopfPresenter struct {
	MockGamePresenter[interfaces.DoppelkopfGame]
}

// HintOutput モック
func (_m *MockDoppelkopfPresenter) HintOutput(g interfaces.DoppelkopfGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
