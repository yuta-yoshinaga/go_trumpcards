//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockGrandfathersClockPresenter グランドファーザーズ・クロック プレゼンターモック
type MockGrandfathersClockPresenter struct {
	MockGamePresenter[interfaces.GrandfathersClockGame]
}

// HintOutput モック
func (_m *MockGrandfathersClockPresenter) HintOutput(gc interfaces.GrandfathersClockGame) string {
	ret := _m.Called(gc)
	return ret.Get(0).(string)
}
