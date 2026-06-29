//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockFlowerGardenPresenter Flower Garden プレゼンターモック
type MockFlowerGardenPresenter struct {
	MockGamePresenter[interfaces.FlowerGardenGame]
}

// HintOutput モック
func (_m *MockFlowerGardenPresenter) HintOutput(bc interfaces.FlowerGardenGame) string {
	ret := _m.Called(bc)
	return ret.Get(0).(string)
}
