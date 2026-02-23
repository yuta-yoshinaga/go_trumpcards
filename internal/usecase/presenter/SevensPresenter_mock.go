package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockSevensPresenter 7並べプレゼンターモック
type MockSevensPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockSevensPresenter) Output(s *domain.Sevens) string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
