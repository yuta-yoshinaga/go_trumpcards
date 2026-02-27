package presenter

import (
	"github.com/stretchr/testify/mock"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockHoldemPresenter テキサスホールデムプレゼンターモック
type MockHoldemPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockHoldemPresenter) Output(h interfaces.HoldemGame, lastErr error) string {
	ret := _m.Called(h, lastErr)
	return ret.Get(0).(string)
}
