//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockNapoleonPresenter ナポレオンプレゼンターモック
type MockNapoleonPresenter struct {
	MockGamePresenter[interfaces.NapoleonGame]
}

// HintOutput モック
func (_m *MockNapoleonPresenter) HintOutput(n interfaces.NapoleonGame) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}
