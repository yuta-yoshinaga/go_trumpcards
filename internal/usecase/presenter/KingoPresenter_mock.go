//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockKingoPresenter キンゴプレゼンターモック
type MockKingoPresenter struct {
	MockGamePresenter[interfaces.KingoGame]
}

// HintOutput モック
func (_m *MockKingoPresenter) HintOutput(s interfaces.KingoGame) string {
	return _m.Called(s).Get(0).(string)
}
