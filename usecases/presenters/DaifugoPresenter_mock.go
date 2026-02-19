package presenters

import (
	"github.com/yuta-yoshinaga/go_trumpcards/entities"

	"github.com/stretchr/testify/mock"
)

// MockDaifugoPresenter 大富豪プレゼンターモック
type MockDaifugoPresenter struct {
	mock.Mock
}

// Output モック
func (_m *MockDaifugoPresenter) Output(dg *entities.Daifugo) string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
