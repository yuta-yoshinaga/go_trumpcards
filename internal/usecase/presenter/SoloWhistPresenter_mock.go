//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSoloWhistPresenter ソロ・ホイストのプレゼンターモック
type MockSoloWhistPresenter struct {
	MockGamePresenter[interfaces.SoloWhistGame]
}

// HintOutput モック
func (_m *MockSoloWhistPresenter) HintOutput(g interfaces.SoloWhistGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
