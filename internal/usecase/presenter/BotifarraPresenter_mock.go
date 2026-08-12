//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBotifarraPresenter ボティファラプレゼンターモック
type MockBotifarraPresenter struct {
	MockGamePresenter[interfaces.BotifarraGame]
}

// HintOutput モック
func (_m *MockBotifarraPresenter) HintOutput(s interfaces.BotifarraGame) string {
	return _m.Called(s).Get(0).(string)
}
