//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTuSacPresenter 四色牌プレゼンターモック
type MockTuSacPresenter struct {
	MockGamePresenter[interfaces.TuSacGame]
}

// HintOutput モック
func (_m *MockTuSacPresenter) HintOutput(s interfaces.TuSacGame) string {
	return _m.Called(s).Get(0).(string)
}
