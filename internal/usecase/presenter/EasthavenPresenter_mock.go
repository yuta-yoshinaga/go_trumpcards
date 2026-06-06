//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockEasthavenPresenter イーストヘイブンプレゼンターモック
type MockEasthavenPresenter struct {
	mock.Mock
}

func (_m *MockEasthavenPresenter) Output(e interfaces.EasthavenGame, lastErr error) string {
	ret := _m.Called(e, lastErr)
	return ret.String(0)
}

func (_m *MockEasthavenPresenter) HintOutput(e interfaces.EasthavenGame) string {
	ret := _m.Called(e)
	return ret.String(0)
}

func (_m *MockEasthavenPresenter) ActionLogOutput(e interfaces.EasthavenGame) string {
	ret := _m.Called(e)
	return ret.String(0)
}
