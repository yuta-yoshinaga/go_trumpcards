//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTrenteEtQuarantePresenter はトラント・エ・カラントプレゼンターモック。
type MockTrenteEtQuarantePresenter struct {
	MockGamePresenter[interfaces.TrenteEtQuaranteGame]
}

// HintOutput モック
func (_m *MockTrenteEtQuarantePresenter) HintOutput(g interfaces.TrenteEtQuaranteGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
