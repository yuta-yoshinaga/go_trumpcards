package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockBaccaratPresenter バカラプレゼンターモック
type MockBaccaratPresenter struct {
	mock.Mock
}

func (m *MockBaccaratPresenter) Output(b interfaces.BaccaratGame, lastErr error) string {
	args := m.Called(b, lastErr)
	return args.String(0)
}

func (m *MockBaccaratPresenter) ActionLogOutput(b interfaces.BaccaratGame) string {
	args := m.Called(b)
	return args.String(0)
}
