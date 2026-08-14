//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSergeantMajorPresenter サージェントメジャープレゼンターモック
type MockSergeantMajorPresenter struct {
	MockGamePresenter[interfaces.SergeantMajorGame]
}

// HintOutput モック
func (_m *MockSergeantMajorPresenter) HintOutput(s interfaces.SergeantMajorGame) string {
	return _m.Called(s).Get(0).(string)
}
