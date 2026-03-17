package presenter

import "github.com/stretchr/testify/mock"

// MockGamePresenter ゲームプレゼンター汎用モック
type MockGamePresenter[G any] struct {
	mock.Mock
}

// Output モック
func (_m *MockGamePresenter[G]) Output(game G, lastErr error) string {
	ret := _m.Called(game, lastErr)
	return ret.Get(0).(string)
}

// ActionLogOutput モック
func (_m *MockGamePresenter[G]) ActionLogOutput(game G) string {
	ret := _m.Called(game)
	return ret.Get(0).(string)
}
