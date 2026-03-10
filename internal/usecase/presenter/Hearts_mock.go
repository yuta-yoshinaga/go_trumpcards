package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockHeartsPresenter ハーツプレゼンターモック
type MockHeartsPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockHeartsPresenter) Output(h interfaces.HeartsGame, lastErr error) string {
	ret := _m.Called(h, lastErr)
	return ret.Get(0).(string)
}

// ActionLogOutput モック
func (_m *MockHeartsPresenter) ActionLogOutput(h interfaces.HeartsGame) string {
	ret := _m.Called(h)
	return ret.Get(0).(string)
}
