package presenters

import (
	"github.com/yuta-yoshinaga/go_trumpcards/entities"

	"github.com/stretchr/testify/mock"
)

// MockOldMaidPresenter ババ抜きプレゼンターモック
type MockOldMaidPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockOldMaidPresenter) Output(om *entities.OldMaid) string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
