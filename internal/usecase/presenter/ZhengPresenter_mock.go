//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockZhengPresenter 争上游プレゼンターモック
type MockZhengPresenter struct{ mock.Mock }

func (_m *MockZhengPresenter) Output(g interfaces.ZhengGame, err error) string {
	return _m.Called(g, err).String(0)
}
func (_m *MockZhengPresenter) ActionLogOutput(g interfaces.ZhengGame) string {
	return _m.Called(g).String(0)
}
func (_m *MockZhengPresenter) HintOutput(g interfaces.ZhengGame) string {
	return _m.Called(g).String(0)
}
