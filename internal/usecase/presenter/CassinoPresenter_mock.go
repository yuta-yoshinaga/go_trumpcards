//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCassinoPresenter カシノプレゼンターモック。
type MockCassinoPresenter struct {
	MockGamePresenter[interfaces.CassinoGame]
}

// HintOutput モック
func (_m *MockCassinoPresenter) HintOutput(cg interfaces.CassinoGame) string {
	ret := _m.Called(cg)
	return ret.Get(0).(string)
}
