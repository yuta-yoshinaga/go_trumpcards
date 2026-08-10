//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockDurakPresenter ドゥラークプレゼンターモック
type MockDurakPresenter struct {
	MockGamePresenter[interfaces.DurakGame]
}

// HintOutput モック
func (_m *MockDurakPresenter) HintOutput(g interfaces.DurakGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
