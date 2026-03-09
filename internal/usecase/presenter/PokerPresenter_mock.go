package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockPokerPresenter ポーカープレゼンターモック
type MockPokerPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockPokerPresenter) Output(p interfaces.PokerGame, lastErr error) string {
	ret := _m.Called(p, lastErr)
	return ret.Get(0).(string)
}

// OutputWithOdds モック
func (_m *MockPokerPresenter) OutputWithOdds(p interfaces.PokerGame, lastErr error, odds []domain.PokerDrawOdds) string {
	ret := _m.Called(p, lastErr, odds)
	return ret.Get(0).(string)
}

// ActionLogOutput モック
func (_m *MockPokerPresenter) ActionLogOutput(p interfaces.PokerGame) string {
	ret := _m.Called(p)
	return ret.Get(0).(string)
}
