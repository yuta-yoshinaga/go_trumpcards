package usecases

import "github.com/stretchr/testify/mock"

// MockPokerInteractor ポーカーインタラクターモック
type MockPokerInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockPokerInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Exchange モック
func (_m *MockPokerInteractor) Exchange(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

// Stand モック
func (_m *MockPokerInteractor) Stand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
