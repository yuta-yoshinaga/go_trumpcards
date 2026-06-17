//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockThreeCardBragPresenter スリーカード・ブラグのプレゼンターモック
type MockThreeCardBragPresenter struct {
	MockGamePresenter[interfaces.ThreeCardBragGame]
}

// HintOutput モック
func (_m *MockThreeCardBragPresenter) HintOutput(g interfaces.ThreeCardBragGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
