//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMariasPresenter マリアーシュのプレゼンターモック
type MockMariasPresenter struct {
	MockGamePresenter[interfaces.MariasGame]
}

// HintOutput モック
func (_m *MockMariasPresenter) HintOutput(g interfaces.MariasGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
