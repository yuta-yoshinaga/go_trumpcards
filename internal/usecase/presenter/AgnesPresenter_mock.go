//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockAgnesPresenter アグネス・ソレルプレゼンターモック
type MockAgnesPresenter struct {
	MockGamePresenter[interfaces.AgnesGame]
}

// HintOutput モック
func (_m *MockAgnesPresenter) HintOutput(c interfaces.AgnesGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
