package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockMemoryPresenter 神経衰弱プレゼンターモック
type MockMemoryPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockMemoryPresenter) Output(m interfaces.MemoryGame, lastErr error) string {
	ret := _m.Called(m, lastErr)
	return ret.Get(0).(string)
}

// ActionLogOutput モック
func (_m *MockMemoryPresenter) ActionLogOutput(m interfaces.MemoryGame) string {
	ret := _m.Called(m)
	return ret.Get(0).(string)
}
