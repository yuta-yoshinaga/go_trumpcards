//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockRedDogPresenter レッドドッグプレゼンターモック
type MockRedDogPresenter struct {
	MockGamePresenter[interfaces.RedDogGame]
}

// HintOutput モック
func (_m *MockRedDogPresenter) HintOutput(rd interfaces.RedDogGame) string {
	ret := _m.Called(rd)
	return ret.Get(0).(string)
}
