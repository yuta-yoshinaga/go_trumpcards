package presenter

import (
	"github.com/stretchr/testify/mock"
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
