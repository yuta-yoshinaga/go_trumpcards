//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockFourteenOutPresenter はフォーティーンアウト・ソリティアのプレゼンターモック。
type MockFourteenOutPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockFourteenOutPresenter) Output(g interfaces.FourteenOutGame, lastErr error) string {
	ret := _m.Called(g, lastErr)
	return ret.String(0)
}

// HintOutput モック
func (_m *MockFourteenOutPresenter) HintOutput(g interfaces.FourteenOutGame) string {
	ret := _m.Called(g)
	return ret.String(0)
}

// ActionLogOutput モック
func (_m *MockFourteenOutPresenter) ActionLogOutput(g interfaces.FourteenOutGame) string {
	ret := _m.Called(g)
	return ret.String(0)
}
