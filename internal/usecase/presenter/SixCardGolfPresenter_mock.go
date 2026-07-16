//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSixCardGolfPresenter シックスカードゴルフプレゼンターモック
type MockSixCardGolfPresenter struct {
	MockGamePresenter[interfaces.SixCardGolfGame]
}

// HintOutput モック
func (_m *MockSixCardGolfPresenter) HintOutput(g interfaces.SixCardGolfGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
