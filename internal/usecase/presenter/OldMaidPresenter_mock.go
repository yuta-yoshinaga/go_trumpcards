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
func (_m *MockOldMaidPresenter) Output(om *domain.OldMaid) string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
