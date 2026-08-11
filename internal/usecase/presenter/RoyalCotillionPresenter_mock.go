//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockRoyalCotillionPresenter ロイヤルコティヨン プレゼンターモック
type MockRoyalCotillionPresenter struct {
	MockGamePresenter[interfaces.RoyalCotillionGame]
}

// HintOutput モック
func (_m *MockRoyalCotillionPresenter) HintOutput(c interfaces.RoyalCotillionGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
