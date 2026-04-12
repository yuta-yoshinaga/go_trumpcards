//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTwoTenJackPresenter ツーテンジャックプレゼンターモック
type MockTwoTenJackPresenter struct {
	MockGamePresenter[interfaces.TwoTenJackGame]
}

// HintOutput モック
func (_m *MockTwoTenJackPresenter) HintOutput(ttj interfaces.TwoTenJackGame) string {
	ret := _m.Called(ttj)
	return ret.Get(0).(string)
}
