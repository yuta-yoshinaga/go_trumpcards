//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockScorpionPresenter スコーピオンプレゼンターモック
type MockScorpionPresenter struct {
	mock.Mock
}

func (_m *MockScorpionPresenter) Output(s interfaces.ScorpionGame, lastErr error) string {
	ret := _m.Called(s, lastErr)
	return ret.String(0)
}

func (_m *MockScorpionPresenter) HintOutput(s interfaces.ScorpionGame) string {
	ret := _m.Called(s)
	return ret.String(0)
}

func (_m *MockScorpionPresenter) LegalMovesOutput(s interfaces.ScorpionGame, col int) string {
	ret := _m.Called(s, col)
	return ret.String(0)
}

func (_m *MockScorpionPresenter) ActionLogOutput(s interfaces.ScorpionGame) string {
	ret := _m.Called(s)
	return ret.String(0)
}
