//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSultanPresenter スルタンプレゼンターモック
type MockSultanPresenter struct {
	MockGamePresenter[interfaces.SultanGame]
}

// HintOutput モック
func (_m *MockSultanPresenter) HintOutput(su interfaces.SultanGame) string {
	ret := _m.Called(su)
	return ret.Get(0).(string)
}
