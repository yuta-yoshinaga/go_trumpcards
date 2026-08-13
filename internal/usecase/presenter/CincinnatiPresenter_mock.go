//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCincinnatiPresenter シンシナティプレゼンターモック
type MockCincinnatiPresenter struct {
	MockGamePresenter[interfaces.CincinnatiGame]
}

// HintOutput モック
func (_m *MockCincinnatiPresenter) HintOutput(s interfaces.CincinnatiGame) string {
	return _m.Called(s).Get(0).(string)
}
