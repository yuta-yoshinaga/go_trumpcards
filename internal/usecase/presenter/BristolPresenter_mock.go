//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBristolPresenter ブリストルプレゼンターモック
type MockBristolPresenter struct {
	MockGamePresenter[interfaces.BristolGame]
}

// TargetsOutput モック
func (_m *MockBristolPresenter) TargetsOutput(b interfaces.BristolGame, zone string, col int) string {
	ret := _m.Called(b, zone, col)
	return ret.String(0)
}

// HintOutput モック
func (_m *MockBristolPresenter) HintOutput(b interfaces.BristolGame) string {
	ret := _m.Called(b)
	return ret.Get(0).(string)
}
