package usecases

import "github.com/stretchr/testify/mock"

// MockBlackJackInteractor ブラックジャックインタラクターモック
type MockBlackJackInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockBlackJackInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Hit モック
func (_m *MockBlackJackInteractor) Hit() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Stand モック
func (_m *MockBlackJackInteractor) Stand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
