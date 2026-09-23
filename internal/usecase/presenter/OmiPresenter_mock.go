//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockOmiPresenter オミプレゼンターモック
type MockOmiPresenter struct {
	MockGamePresenter[interfaces.OmiGame]
}

// HintOutput モック
func (_m *MockOmiPresenter) HintOutput(e interfaces.OmiGame) string {
	ret := _m.Called(e)
	return ret.Get(0).(string)
}
