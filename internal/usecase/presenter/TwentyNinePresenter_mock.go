//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTwentyNinePresenter トゥエンティナイン (29) のプレゼンターモック
type MockTwentyNinePresenter struct {
	MockGamePresenter[interfaces.TwentyNineGame]
}

// HintOutput モック
func (_m *MockTwentyNinePresenter) HintOutput(g interfaces.TwentyNineGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
