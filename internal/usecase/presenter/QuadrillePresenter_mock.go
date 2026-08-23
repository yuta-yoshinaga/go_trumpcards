//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockQuadrillePresenter カドリール (Quadrille) のプレゼンターモック
type MockQuadrillePresenter struct {
	MockGamePresenter[interfaces.QuadrilleGame]
}

// HintOutput モック
func (_m *MockQuadrillePresenter) HintOutput(g interfaces.QuadrilleGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
