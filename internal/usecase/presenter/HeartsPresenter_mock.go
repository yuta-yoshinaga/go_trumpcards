//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockHeartsPresenter ハーツプレゼンターモック
type MockHeartsPresenter struct {
	MockGamePresenter[interfaces.HeartsGame]
}

// HintOutput モック
func (_m *MockHeartsPresenter) HintOutput(h interfaces.HeartsGame) string {
	ret := _m.Called(h)
	return ret.Get(0).(string)
}
