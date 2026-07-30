//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockChineseTenPresenter 撿紅點 プレゼンターモック
type MockChineseTenPresenter struct {
	MockGamePresenter[interfaces.ChineseTenGame]
}

// HintOutput モック
func (_m *MockChineseTenPresenter) HintOutput(c interfaces.ChineseTenGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
