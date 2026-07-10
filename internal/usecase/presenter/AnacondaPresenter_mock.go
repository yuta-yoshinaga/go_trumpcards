//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockAnacondaPresenter はアナコンダ (Anaconda) プレゼンターモック。
type MockAnacondaPresenter struct {
	MockGamePresenter[interfaces.AnacondaGame]
}

// HintOutput モック
func (_m *MockAnacondaPresenter) HintOutput(g interfaces.AnacondaGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
