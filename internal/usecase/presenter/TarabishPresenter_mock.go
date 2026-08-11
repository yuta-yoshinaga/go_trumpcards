//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTarabishPresenter タラビッシュプレゼンターモック
type MockTarabishPresenter struct {
	MockGamePresenter[interfaces.TarabishGame]
}

// HintOutput モック
func (_m *MockTarabishPresenter) HintOutput(t interfaces.TarabishGame) string {
	return _m.Called(t).Get(0).(string)
}
