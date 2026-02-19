package presenters

import (
	"github.com/yuta-yoshinaga/go_trumpcards/entities"

	"github.com/stretchr/testify/mock"
)

// MockPokerPresenter ポーカープレゼンターモック
type MockPokerPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockPokerPresenter) Output(p *entities.Poker) string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
