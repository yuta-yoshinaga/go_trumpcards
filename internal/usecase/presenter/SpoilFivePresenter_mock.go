//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSpoilFivePresenter スポイル・ファイブのプレゼンターモック
type MockSpoilFivePresenter struct {
	MockGamePresenter[interfaces.SpoilFiveGame]
}

// HintOutput モック
func (_m *MockSpoilFivePresenter) HintOutput(g interfaces.SpoilFiveGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
