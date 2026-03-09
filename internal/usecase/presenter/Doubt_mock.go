package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockDoubtPresenter ダウトプレゼンターモック
type MockDoubtPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockDoubtPresenter) Output(d interfaces.DoubtGame, lastErr error) string {
	ret := _m.Called(d, lastErr)
	return ret.Get(0).(string)
}

// ActionLogOutput モック
func (_m *MockDoubtPresenter) ActionLogOutput(d interfaces.DoubtGame) string {
	ret := _m.Called(d)
	return ret.Get(0).(string)
}
