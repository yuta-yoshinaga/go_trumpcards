//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockYukonPresenter ユーコンプレゼンターモック
type MockYukonPresenter struct {
	mock.Mock
}

func (_m *MockYukonPresenter) Output(y interfaces.YukonGame, lastErr error) string {
	ret := _m.Called(y, lastErr)
	return ret.String(0)
}

func (_m *MockYukonPresenter) HintOutput(y interfaces.YukonGame) string {
	ret := _m.Called(y)
	return ret.String(0)
}

func (_m *MockYukonPresenter) ActionLogOutput(y interfaces.YukonGame) string {
	ret := _m.Called(y)
	return ret.String(0)
}
