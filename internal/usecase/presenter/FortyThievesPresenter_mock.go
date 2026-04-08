//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockFortyThievesPresenter フォーティシーブスプレゼンターモック
type MockFortyThievesPresenter struct {
	MockGamePresenter[interfaces.FortyThievesGame]
}

// HintOutput モック
func (_m *MockFortyThievesPresenter) HintOutput(ft interfaces.FortyThievesGame) string {
	ret := _m.Called(ft)
	return ret.Get(0).(string)
}
