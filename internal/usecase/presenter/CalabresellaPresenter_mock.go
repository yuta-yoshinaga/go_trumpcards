//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCalabresellaPresenter カラブレセッラ (Calabresella) のプレゼンターモック
type MockCalabresellaPresenter struct {
	MockGamePresenter[interfaces.CalabresellaGame]
}

// HintOutput モック
func (_m *MockCalabresellaPresenter) HintOutput(g interfaces.CalabresellaGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
