//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBeleagueredCastlePresenter Beleaguered Castle プレゼンターモック
type MockBeleagueredCastlePresenter struct {
	MockGamePresenter[interfaces.BeleagueredCastleGame]
}

// HintOutput モック
func (_m *MockBeleagueredCastlePresenter) HintOutput(bc interfaces.BeleagueredCastleGame) string {
	ret := _m.Called(bc)
	return ret.Get(0).(string)
}
