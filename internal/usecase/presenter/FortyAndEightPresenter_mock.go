//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockFortyAndEightPresenter フォーティ・アンド・エイトプレゼンターモック
type MockFortyAndEightPresenter struct {
	MockGamePresenter[interfaces.FortyAndEightGame]
}

// HintOutput モック
func (_m *MockFortyAndEightPresenter) HintOutput(ft interfaces.FortyAndEightGame) string {
	ret := _m.Called(ft)
	return ret.Get(0).(string)
}
