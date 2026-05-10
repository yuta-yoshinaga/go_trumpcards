//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockMonteCarloPresenter はモンテカルロ・ソリティアのプレゼンターモック。
type MockMonteCarloPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockMonteCarloPresenter) Output(g interfaces.MonteCarloGame, lastErr error) string {
	ret := _m.Called(g, lastErr)
	return ret.String(0)
}

// HintOutput モック
func (_m *MockMonteCarloPresenter) HintOutput(g interfaces.MonteCarloGame) string {
	ret := _m.Called(g)
	return ret.String(0)
}

// ActionLogOutput モック
func (_m *MockMonteCarloPresenter) ActionLogOutput(g interfaces.MonteCarloGame) string {
	ret := _m.Called(g)
	return ret.String(0)
}
