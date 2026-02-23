package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockPokerPresenter ポーカープレゼンターモック
type MockPokerPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockPokerPresenter) Output(p *domain.Poker) string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
