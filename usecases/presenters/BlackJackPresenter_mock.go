package presenters

import (
	"go_trumpcards/entities"

	"github.com/stretchr/testify/mock"
)

// MockBlackJackPresenter ブラックジャックプレゼンターモック
type MockBlackJackPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockBlackJackPresenter) Output(bj *entities.BlackJack) string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
