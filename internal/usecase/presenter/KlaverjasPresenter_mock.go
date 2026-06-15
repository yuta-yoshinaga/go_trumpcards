//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockKlaverjasPresenter クラヴァヤスのプレゼンターモック
type MockKlaverjasPresenter struct {
	MockGamePresenter[interfaces.KlaverjasGame]
}

// HintOutput モック
func (_m *MockKlaverjasPresenter) HintOutput(g interfaces.KlaverjasGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
