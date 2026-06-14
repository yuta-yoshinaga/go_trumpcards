//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSheepsheadPresenter シープスヘッドのプレゼンターモック
type MockSheepsheadPresenter struct {
	MockGamePresenter[interfaces.SheepsheadGame]
}

// HintOutput モック
func (_m *MockSheepsheadPresenter) HintOutput(g interfaces.SheepsheadGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
