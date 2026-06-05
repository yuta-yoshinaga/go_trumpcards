//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockGongZhuPresenter 拱猪プレゼンターモック
type MockGongZhuPresenter struct {
	MockGamePresenter[interfaces.GongZhuGame]
}

// HintOutput モック
func (_m *MockGongZhuPresenter) HintOutput(g interfaces.GongZhuGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
