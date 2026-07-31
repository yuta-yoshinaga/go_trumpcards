//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTrexPresenter トリックス プレゼンターモック
type MockTrexPresenter struct {
	MockGamePresenter[interfaces.TrexGame]
}

// HintOutput モック
func (_m *MockTrexPresenter) HintOutput(c interfaces.TrexGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
