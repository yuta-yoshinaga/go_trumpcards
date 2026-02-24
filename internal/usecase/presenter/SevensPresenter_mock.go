package presenter

import (
	"github.com/stretchr/testify/mock"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockSevensPresenter 7並べプレゼンターモック
type MockSevensPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockSevensPresenter) Output(s interfaces.SevensGame, lastErr error) string {
	ret := _m.Called(s, lastErr)
	return ret.Get(0).(string)
}
