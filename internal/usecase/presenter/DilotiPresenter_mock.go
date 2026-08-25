//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockDilotiPresenter はディロティのプレゼンターモック。
type MockDilotiPresenter struct {
	MockGamePresenter[interfaces.DilotiGame]
}

// HintOutput モック
func (_m *MockDilotiPresenter) HintOutput(g interfaces.DilotiGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
