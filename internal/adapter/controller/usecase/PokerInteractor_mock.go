package usecase

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

// Bet モック
func (_m *MockPokerInteractor) Bet(amount int) string {
	ret := _m.Called(amount)
	return ret.Get(0).(string)
}

// Call モック
func (_m *MockPokerInteractor) Call() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Raise モック
func (_m *MockPokerInteractor) Raise(amount int) string {
	ret := _m.Called(amount)
	return ret.Get(0).(string)
}

// Fold モック
func (_m *MockPokerInteractor) Fold() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Check モック
func (_m *MockPokerInteractor) Check() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
