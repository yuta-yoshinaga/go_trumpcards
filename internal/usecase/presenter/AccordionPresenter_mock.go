//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockAccordionPresenter アコーディオンプレゼンターモック
type MockAccordionPresenter struct {
	mock.Mock
}

func (_m *MockAccordionPresenter) Output(a interfaces.AccordionGame, lastErr error) string {
	ret := _m.Called(a, lastErr)
	return ret.String(0)
}

func (_m *MockAccordionPresenter) HintOutput(a interfaces.AccordionGame) string {
	ret := _m.Called(a)
	return ret.String(0)
}

func (_m *MockAccordionPresenter) ActionLogOutput(a interfaces.AccordionGame) string {
	ret := _m.Called(a)
	return ret.String(0)
}
