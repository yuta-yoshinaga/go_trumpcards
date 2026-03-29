//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockTriPeaksPresenter トリピークスプレゼンターモック
type MockTriPeaksPresenter struct {
	mock.Mock
}

func (_m *MockTriPeaksPresenter) Output(t interfaces.TriPeaksGame, lastErr error) string {
	ret := _m.Called(t, lastErr)
	return ret.Get(0).(string)
}

func (_m *MockTriPeaksPresenter) HintOutput(t interfaces.TriPeaksGame) string {
	ret := _m.Called(t)
	return ret.Get(0).(string)
}

func (_m *MockTriPeaksPresenter) ActionLogOutput(t interfaces.TriPeaksGame) string {
	ret := _m.Called(t)
	return ret.Get(0).(string)
}
