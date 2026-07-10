//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockNinetyNinePresenter ナインティナインプレゼンターモック
type MockNinetyNinePresenter struct {
	MockGamePresenter[interfaces.NinetyNineGame]
}

// HintOutput モック
func (_m *MockNinetyNinePresenter) HintOutput(o interfaces.NinetyNineGame) string {
	ret := _m.Called(o)
	return ret.Get(0).(string)
}
