//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCasinoHoldemPresenter カジノホールデムプレゼンターモック
type MockCasinoHoldemPresenter struct {
	MockGamePresenter[interfaces.CasinoHoldemGame]
}

// HintOutput モック
func (_m *MockCasinoHoldemPresenter) HintOutput(g interfaces.CasinoHoldemGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
