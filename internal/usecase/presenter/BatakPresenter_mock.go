//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBatakPresenter Batak プレゼンターモック
type MockBatakPresenter struct {
	MockGamePresenter[interfaces.BatakGame]
}

// HintOutput モック
func (_m *MockBatakPresenter) HintOutput(cb interfaces.BatakGame) string {
	ret := _m.Called(cb)
	return ret.Get(0).(string)
}
