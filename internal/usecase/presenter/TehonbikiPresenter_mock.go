//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTehonbikiPresenter 手本引きプレゼンターモック
type MockTehonbikiPresenter struct {
	MockGamePresenter[interfaces.TehonbikiGame]
}

// HintOutput モック
func (_m *MockTehonbikiPresenter) HintOutput(s interfaces.TehonbikiGame) string {
	return _m.Called(s).Get(0).(string)
}
