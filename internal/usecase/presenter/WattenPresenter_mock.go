//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockWattenPresenter ヴァッテンプレゼンターモック
type MockWattenPresenter struct {
	MockGamePresenter[interfaces.WattenGame]
}

// HintOutput モック
func (_m *MockWattenPresenter) HintOutput(g interfaces.WattenGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
