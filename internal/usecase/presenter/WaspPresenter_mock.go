//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockWaspPresenter ワスププレゼンターモック
type MockWaspPresenter struct {
	mock.Mock
}

func (_m *MockWaspPresenter) Output(s interfaces.WaspGame, lastErr error) string {
	ret := _m.Called(s, lastErr)
	return ret.String(0)
}

func (_m *MockWaspPresenter) HintOutput(s interfaces.WaspGame) string {
	ret := _m.Called(s)
	return ret.String(0)
}

func (_m *MockWaspPresenter) LegalMovesOutput(s interfaces.WaspGame, col int) string {
	ret := _m.Called(s, col)
	return ret.String(0)
}

func (_m *MockWaspPresenter) ActionLogOutput(s interfaces.WaspGame) string {
	ret := _m.Called(s)
	return ret.String(0)
}
