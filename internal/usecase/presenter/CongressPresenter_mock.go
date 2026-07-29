//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCongressPresenter コングレス プレゼンターモック
type MockCongressPresenter struct {
	MockGamePresenter[interfaces.CongressGame]
}

// HintOutput モック
func (_m *MockCongressPresenter) HintOutput(c interfaces.CongressGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
