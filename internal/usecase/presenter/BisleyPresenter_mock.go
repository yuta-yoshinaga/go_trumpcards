//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBisleyPresenter ビズリー プレゼンターモック
type MockBisleyPresenter struct {
	MockGamePresenter[interfaces.BisleyGame]
}

// HintOutput モック
func (_m *MockBisleyPresenter) HintOutput(b interfaces.BisleyGame) string {
	ret := _m.Called(b)
	return ret.Get(0).(string)
}
