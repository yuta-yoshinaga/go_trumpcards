package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockBlackJackPresenter ブラックジャックプレゼンターモック
type MockBlackJackPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockBlackJackPresenter) Output(bj interfaces.BlackJackGame, lastErr error) string {
	ret := _m.Called(bj, lastErr)
	return ret.Get(0).(string)
}

// ActionLogOutput モック
func (_m *MockBlackJackPresenter) ActionLogOutput(bj interfaces.BlackJackGame) string {
	ret := _m.Called(bj)
	return ret.Get(0).(string)
}
