//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockEscobaPresenter エスコバプレゼンターモック。
type MockEscobaPresenter struct {
	MockGamePresenter[interfaces.EscobaGame]
}

// HintOutput モック
func (_m *MockEscobaPresenter) HintOutput(g interfaces.EscobaGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
