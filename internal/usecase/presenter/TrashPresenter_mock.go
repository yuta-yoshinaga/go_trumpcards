//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockTrashPresenter トラッシュプレゼンターモック
type MockTrashPresenter struct {
	mock.Mock
}

func (_m *MockTrashPresenter) Output(t interfaces.TrashGame, lastErr error) string {
	ret := _m.Called(t, lastErr)
	return ret.String(0)
}

func (_m *MockTrashPresenter) ActionLogOutput(t interfaces.TrashGame) string {
	ret := _m.Called(t)
	return ret.String(0)
}

func (_m *MockTrashPresenter) HintOutput(t interfaces.TrashGame) string {
	ret := _m.Called(t)
	return ret.String(0)
}
