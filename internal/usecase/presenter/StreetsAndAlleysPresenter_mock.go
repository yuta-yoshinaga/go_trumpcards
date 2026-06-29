//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockStreetsAndAlleysPresenter Streets and Alleys プレゼンターモック
type MockStreetsAndAlleysPresenter struct {
	MockGamePresenter[interfaces.StreetsAndAlleysGame]
}

// HintOutput モック
func (_m *MockStreetsAndAlleysPresenter) HintOutput(bc interfaces.StreetsAndAlleysGame) string {
	ret := _m.Called(bc)
	return ret.Get(0).(string)
}
