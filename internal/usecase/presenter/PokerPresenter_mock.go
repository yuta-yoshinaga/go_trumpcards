package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockPokerPresenter ポーカープレゼンターモック
type MockPokerPresenter struct {
	MockGamePresenter[interfaces.PokerGame]
}

// OutputWithOdds モック
func (_m *MockPokerPresenter) OutputWithOdds(p interfaces.PokerGame, lastErr error, odds []domain.PokerDrawOdds) string {
	ret := _m.Called(p, lastErr, odds)
	return ret.Get(0).(string)
}
