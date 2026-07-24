//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPokerSquaresPresenter はポーカー・スクエアズのプレゼンターモック。
type MockPokerSquaresPresenter struct {
	MockGamePresenter[interfaces.PokerSquaresGame]
}

// HintOutput はヒント出力のモック。
func (_m *MockPokerSquaresPresenter) HintOutput(p interfaces.PokerSquaresGame) string {
	ret := _m.Called(p)
	return ret.String(0)
}
