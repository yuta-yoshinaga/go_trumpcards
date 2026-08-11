//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockEstimationPresenter エスティメーションプレゼンターモック
type MockEstimationPresenter struct {
	MockGamePresenter[interfaces.EstimationGame]
}

// HintOutput モック
func (_m *MockEstimationPresenter) HintOutput(e interfaces.EstimationGame) string {
	return _m.Called(e).Get(0).(string)
}
