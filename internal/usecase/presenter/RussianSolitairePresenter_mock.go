//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockRussianSolitairePresenter ロシアンソリティアプレゼンターモック
type MockRussianSolitairePresenter struct {
	mock.Mock
}

func (_m *MockRussianSolitairePresenter) Output(r interfaces.RussianSolitaireGame, lastErr error) string {
	ret := _m.Called(r, lastErr)
	return ret.String(0)
}

func (_m *MockRussianSolitairePresenter) HintOutput(r interfaces.RussianSolitaireGame) string {
	ret := _m.Called(r)
	return ret.String(0)
}

func (_m *MockRussianSolitairePresenter) ActionLogOutput(r interfaces.RussianSolitaireGame) string {
	ret := _m.Called(r)
	return ret.String(0)
}
