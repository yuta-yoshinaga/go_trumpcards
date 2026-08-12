//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockAndarBaharPresenter アンダーバハールプレゼンターモック
type MockAndarBaharPresenter struct {
	MockGamePresenter[interfaces.AndarBaharGame]
}

// HintOutput モック
func (_m *MockAndarBaharPresenter) HintOutput(s interfaces.AndarBaharGame) string {
	return _m.Called(s).Get(0).(string)
}
