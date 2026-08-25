//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockDehlaPakadPresenter はデーラ・パカドのプレゼンターモック。
type MockDehlaPakadPresenter struct {
	MockGamePresenter[interfaces.DehlaPakadGame]
}

// HintOutput モック
func (_m *MockDehlaPakadPresenter) HintOutput(g interfaces.DehlaPakadGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
