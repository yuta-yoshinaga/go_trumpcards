//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTappTarockPresenter タップ・タロックのプレゼンターモック。
type MockTappTarockPresenter struct {
	MockGamePresenter[interfaces.TappTarockGame]
}

// HintOutput モック
func (_m *MockTappTarockPresenter) HintOutput(g interfaces.TappTarockGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
