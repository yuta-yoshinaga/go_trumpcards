//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSchafkopfPresenter シャーフコップのプレゼンターモック
type MockSchafkopfPresenter struct {
	MockGamePresenter[interfaces.SchafkopfGame]
}

// HintOutput モック
func (_m *MockSchafkopfPresenter) HintOutput(g interfaces.SchafkopfGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
