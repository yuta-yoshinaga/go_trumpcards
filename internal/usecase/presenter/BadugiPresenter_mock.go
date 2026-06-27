//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBadugiPresenter is the testify/mock presenter for Badugi.
type MockBadugiPresenter struct {
	MockGamePresenter[interfaces.BadugiGame]
}

// HintOutput モック
func (_m *MockBadugiPresenter) HintOutput(g interfaces.BadugiGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
