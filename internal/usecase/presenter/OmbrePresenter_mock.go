//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockOmbrePresenter オンブル (Ombre) のプレゼンターモック
type MockOmbrePresenter struct {
	MockGamePresenter[interfaces.OmbreGame]
}

// HintOutput モック
func (_m *MockOmbrePresenter) HintOutput(g interfaces.OmbreGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
