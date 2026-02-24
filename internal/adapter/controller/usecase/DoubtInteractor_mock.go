package usecase

import "github.com/stretchr/testify/mock"

// MockDoubtInteractor ダウトインタラクターモック
type MockDoubtInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockDoubtInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockDoubtInteractor) Play(cardIndices []int, claimedValue int) string {
	ret := _m.Called(cardIndices, claimedValue)
	return ret.Get(0).(string)
}

// ResolveDoubt モック
func (_m *MockDoubtInteractor) ResolveDoubt(doubterIndices []int) string {
	ret := _m.Called(doubterIndices)
	return ret.Get(0).(string)
}

// SkipDoubt モック
func (_m *MockDoubtInteractor) SkipDoubt() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
