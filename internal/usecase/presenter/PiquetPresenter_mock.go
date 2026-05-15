//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPiquetPresenter Piquetプレゼンターモック
type MockPiquetPresenter struct {
	MockGamePresenter[interfaces.PiquetGame]
}

// HintOutput モック
func (_m *MockPiquetPresenter) HintOutput(p interfaces.PiquetGame) string {
	ret := _m.Called(p)
	return ret.Get(0).(string)
}
