//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCitadelPresenter Beleaguered Castle プレゼンターモック
type MockCitadelPresenter struct {
	MockGamePresenter[interfaces.CitadelGame]
}

// HintOutput モック
func (_m *MockCitadelPresenter) HintOutput(bc interfaces.CitadelGame) string {
	ret := _m.Called(bc)
	return ret.Get(0).(string)
}
