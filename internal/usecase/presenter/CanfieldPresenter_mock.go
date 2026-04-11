//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCanfieldPresenter キャンフィールドプレゼンターモック
type MockCanfieldPresenter struct {
	MockGamePresenter[interfaces.CanfieldGame]
}

// HintOutput モック
func (_m *MockCanfieldPresenter) HintOutput(c interfaces.CanfieldGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
