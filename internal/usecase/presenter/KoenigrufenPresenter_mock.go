//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockKoenigrufenPresenter ケーニッヒルーフェン (Königrufen) のプレゼンターモック
type MockKoenigrufenPresenter struct {
	MockGamePresenter[interfaces.KoenigrufenGame]
}

// HintOutput モック
func (_m *MockKoenigrufenPresenter) HintOutput(g interfaces.KoenigrufenGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
