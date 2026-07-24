//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPresidentPresenter プレジデントプレゼンターモック
type MockPresidentPresenter struct {
	MockGamePresenter[interfaces.PresidentGame]
}

// HintOutput モック
func (_m *MockPresidentPresenter) HintOutput(pg interfaces.PresidentGame) string {
	ret := _m.Called(pg)
	return ret.Get(0).(string)
}
