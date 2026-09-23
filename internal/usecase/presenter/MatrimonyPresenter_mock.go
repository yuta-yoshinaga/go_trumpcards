//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMatrimonyPresenter マトリモニー プレゼンターモック
type MockMatrimonyPresenter struct {
	MockGamePresenter[interfaces.MatrimonyGame]
}

// HintOutput モック
func (_m *MockMatrimonyPresenter) HintOutput(c interfaces.MatrimonyGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
