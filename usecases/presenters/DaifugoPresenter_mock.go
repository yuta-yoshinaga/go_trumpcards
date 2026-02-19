package presenters

import (
	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/stretchr/testify/mock"
)

// MockDaifugoPresenter 大富豪プレゼンターモック
type MockDaifugoPresenter struct {
	mock.Mock
}

// Output モック実装
func (m *MockDaifugoPresenter) Output(d *entities.Daifugo) string {
	args := m.Called(d)
	return args.String(0)
}
