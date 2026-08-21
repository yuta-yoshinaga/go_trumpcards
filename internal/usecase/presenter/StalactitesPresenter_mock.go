//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockStalactitesPresenter フリーセルプレゼンターモック
type MockStalactitesPresenter struct {
	MockGamePresenter[interfaces.StalactitesGame]
}

// HintOutput モック
func (_m *MockStalactitesPresenter) HintOutput(f interfaces.StalactitesGame) string {
	ret := _m.Called(f)
	return ret.Get(0).(string)
}
