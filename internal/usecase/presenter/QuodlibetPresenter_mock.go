//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockQuodlibetPresenter はクオドリベットのプレゼンターモック。
type MockQuodlibetPresenter struct {
	MockGamePresenter[interfaces.QuodlibetGame]
}

// HintOutput モック
func (_m *MockQuodlibetPresenter) HintOutput(g interfaces.QuodlibetGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
