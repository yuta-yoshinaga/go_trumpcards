//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCatchTenPresenter Catch the Ten プレゼンターモック
type MockCatchTenPresenter struct {
	MockGamePresenter[interfaces.CatchTenGame]
}

// HintOutput モック
func (_m *MockCatchTenPresenter) HintOutput(g interfaces.CatchTenGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
