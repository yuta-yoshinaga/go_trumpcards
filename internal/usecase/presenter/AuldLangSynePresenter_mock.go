//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockAuldLangSynePresenter オールド・ラング・サインプレゼンターモック
type MockAuldLangSynePresenter struct {
	mock.Mock
}

func (_m *MockAuldLangSynePresenter) Output(g interfaces.AuldLangSyneGame, lastErr error) string {
	return _m.Called(g, lastErr).String(0)
}

func (_m *MockAuldLangSynePresenter) HintOutput(g interfaces.AuldLangSyneGame) string {
	return _m.Called(g).String(0)
}

func (_m *MockAuldLangSynePresenter) ActionLogOutput(g interfaces.AuldLangSyneGame) string {
	return _m.Called(g).String(0)
}
