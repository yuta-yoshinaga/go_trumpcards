//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockUltiPresenter ウルティ (Ulti) のプレゼンターモック
type MockUltiPresenter struct {
	MockGamePresenter[interfaces.UltiGame]
}

// HintOutput モック
func (_m *MockUltiPresenter) HintOutput(g interfaces.UltiGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
