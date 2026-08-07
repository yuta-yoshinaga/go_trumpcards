//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMississippiStudPresenter ミシシッピ・スタッドプレゼンターモック
type MockMississippiStudPresenter struct {
	MockGamePresenter[interfaces.MississippiStudGame]
}

// HintOutput モック
func (_m *MockMississippiStudPresenter) HintOutput(g interfaces.MississippiStudGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
