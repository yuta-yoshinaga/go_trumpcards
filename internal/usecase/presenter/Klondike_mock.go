package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockKlondikePresenter クロンダイクプレゼンターモック
type MockKlondikePresenter struct {
	mock.Mock
}

func (_m *MockKlondikePresenter) Output(k interfaces.KlondikeGame, lastErr error) string {
	ret := _m.Called(k, lastErr)
	return ret.Get(0).(string)
}

func (_m *MockKlondikePresenter) HintOutput(k interfaces.KlondikeGame) string {
	ret := _m.Called(k)
	return ret.Get(0).(string)
}

func (_m *MockKlondikePresenter) ActionLogOutput(k interfaces.KlondikeGame) string {
	ret := _m.Called(k)
	return ret.Get(0).(string)
}
