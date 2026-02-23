package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockBlackJackPresenter ブラックジャックプレゼンターモック
type MockBlackJackPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockBlackJackPresenter) Output(bj *domain.BlackJack, lastErr error) string {
	ret := _m.Called(bj, lastErr)
	return ret.Get(0).(string)
}
