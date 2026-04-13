//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockWhistPresenter ホイストプレゼンターモック
type MockWhistPresenter struct {
	MockGamePresenter[interfaces.WhistGame]
}

// HintOutput モック
func (_m *MockWhistPresenter) HintOutput(w interfaces.WhistGame) string {
	ret := _m.Called(w)
	return ret.Get(0).(string)
}
