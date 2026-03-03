package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockOldMaidPresenter ババ抜きプレゼンターモック
type MockOldMaidPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockOldMaidPresenter) Output(om interfaces.OldMaidGame, lastErr error) string {
	ret := _m.Called(om, lastErr)
	return ret.Get(0).(string)
}
