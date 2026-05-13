//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMightyPresenter マイティプレゼンターモック
type MockMightyPresenter struct {
	MockGamePresenter[interfaces.MightyGame]
}

// HintOutput モック
func (_m *MockMightyPresenter) HintOutput(m interfaces.MightyGame) string {
	ret := _m.Called(m)
	return ret.Get(0).(string)
}
