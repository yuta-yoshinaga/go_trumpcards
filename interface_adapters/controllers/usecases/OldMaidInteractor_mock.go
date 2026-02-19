package usecases

import "github.com/stretchr/testify/mock"

// MockOldMaidInteractor ババ抜きインタラクターモック
type MockOldMaidInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockOldMaidInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Draw モック
func (_m *MockOldMaidInteractor) Draw(cardIdx int) string {
	ret := _m.Called(cardIdx)
	return ret.Get(0).(string)
}
