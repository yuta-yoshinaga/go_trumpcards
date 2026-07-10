//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockKoiKoiPresenter はこいこいプレゼンターモック。
type MockKoiKoiPresenter struct {
	MockGamePresenter[interfaces.KoiKoiGame]
}

// HintOutput モック
func (_m *MockKoiKoiPresenter) HintOutput(g interfaces.KoiKoiGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
