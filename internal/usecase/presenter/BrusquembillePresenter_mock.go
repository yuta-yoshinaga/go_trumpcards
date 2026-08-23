//go:build test && (!js || !wasm || classic)

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBrusquembillePresenter ブリュスカンビーユプレゼンターモック
type MockBrusquembillePresenter struct {
	MockGamePresenter[interfaces.BrusquembilleGame]
}

// HintOutput モック
func (_m *MockBrusquembillePresenter) HintOutput(b interfaces.BrusquembilleGame) string {
	ret := _m.Called(b)
	return ret.Get(0).(string)
}
