//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTrucoPresenter トゥルコプレゼンターモック
type MockTrucoPresenter struct {
	MockGamePresenter[interfaces.TrucoGame]
}

// HintOutput モック
func (_m *MockTrucoPresenter) HintOutput(t interfaces.TrucoGame) string {
	ret := _m.Called(t)
	return ret.Get(0).(string)
}
