package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockDaifugoPresenter 大富豪プレゼンターモック
type MockDaifugoPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockDaifugoPresenter) Output(dg interfaces.DaifugoGame, lastErr error) string {
	ret := _m.Called(dg, lastErr)
	return ret.Get(0).(string)
}
