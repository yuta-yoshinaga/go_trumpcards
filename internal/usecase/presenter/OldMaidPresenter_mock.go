package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockOldMaidPresenter ババ抜きプレゼンターモック
type MockOldMaidPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockOldMaidPresenter) Output(om *domain.OldMaid, lastErr error) string {
	ret := _m.Called(om, lastErr)
	return ret.Get(0).(string)
}
