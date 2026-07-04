//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPrimeroPresenter はプリメロ (Primero) プレゼンターモック。
type MockPrimeroPresenter struct {
	MockGamePresenter[interfaces.PrimeroGame]
}

// HintOutput モック
func (_m *MockPrimeroPresenter) HintOutput(g interfaces.PrimeroGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
