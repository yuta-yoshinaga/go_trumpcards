//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPerseverancePresenter パーシビアランスプレゼンターモック
type MockPerseverancePresenter struct {
	MockGamePresenter[interfaces.PerseveranceGame]
}

// TargetsOutput モック
func (_m *MockPerseverancePresenter) TargetsOutput(bd interfaces.PerseveranceGame, col int) string {
	ret := _m.Called(bd, col)
	return ret.String(0)
}

// HintOutput モック
func (_m *MockPerseverancePresenter) HintOutput(bd interfaces.PerseveranceGame) string {
	ret := _m.Called(bd)
	return ret.Get(0).(string)
}
