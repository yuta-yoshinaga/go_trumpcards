//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockWhiteheadPresenter ホワイトヘッドプレゼンターモック
type MockWhiteheadPresenter struct {
	MockGamePresenter[interfaces.WhiteheadGame]
}

// HintOutput モック
func (_m *MockWhiteheadPresenter) HintOutput(k interfaces.WhiteheadGame) string {
	ret := _m.Called(k)
	return ret.Get(0).(string)
}
