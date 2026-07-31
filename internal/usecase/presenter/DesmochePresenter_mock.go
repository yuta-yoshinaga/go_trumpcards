//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockDesmochePresenter デスモチェ プレゼンターモック
type MockDesmochePresenter struct {
	MockGamePresenter[interfaces.DesmocheGame]
}

// HintOutput モック
func (_m *MockDesmochePresenter) HintOutput(c interfaces.DesmocheGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
