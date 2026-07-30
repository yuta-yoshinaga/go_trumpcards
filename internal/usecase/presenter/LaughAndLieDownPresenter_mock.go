//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockLaughAndLieDownPresenter ラフ・アンド・ライダウン プレゼンターモック
type MockLaughAndLieDownPresenter struct {
	MockGamePresenter[interfaces.LaughAndLieDownGame]
}

// HintOutput モック
func (_m *MockLaughAndLieDownPresenter) HintOutput(c interfaces.LaughAndLieDownGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
