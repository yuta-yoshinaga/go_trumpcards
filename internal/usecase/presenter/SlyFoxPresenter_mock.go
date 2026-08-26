//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSlyFoxPresenter スライ・フォックス プレゼンターモック
type MockSlyFoxPresenter struct {
	MockGamePresenter[interfaces.SlyFoxGame]
}

// HintOutput モック
func (_m *MockSlyFoxPresenter) HintOutput(c interfaces.SlyFoxGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
