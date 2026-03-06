package usecase

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

// Bet モック
func (_m *MockBlackJackInteractor) Bet(amount, ppBet, t3Bet int) string {
	ret := _m.Called(amount, ppBet, t3Bet)
	return ret.Get(0).(string)
}

// DoubleDown モック
func (_m *MockBlackJackInteractor) DoubleDown() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Split モック
func (_m *MockBlackJackInteractor) Split() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Insurance モック
func (_m *MockBlackJackInteractor) Insurance() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// DeclineInsurance モック
func (_m *MockBlackJackInteractor) DeclineInsurance() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Surrender モック
func (_m *MockBlackJackInteractor) Surrender() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SetDeckCount モック
func (_m *MockBlackJackInteractor) SetDeckCount(count int) string {
	ret := _m.Called(count)
	return ret.Get(0).(string)
}

// ToggleHint モック
func (_m *MockBlackJackInteractor) ToggleHint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ToggleSoft17 モック
func (_m *MockBlackJackInteractor) ToggleSoft17() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ToggleCounting モック
func (_m *MockBlackJackInteractor) ToggleCounting() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ToggleDAS モック
func (_m *MockBlackJackInteractor) ToggleDAS() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SetCountingSystem モック
func (_m *MockBlackJackInteractor) SetCountingSystem(system int) string {
	ret := _m.Called(system)
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockBlackJackInteractor) ResetWithConfig(dealerHitsSoft17 bool, cpuPlayerCount int, countingEnabled bool, doubleAfterSplit bool, countingSystem int) string {
	ret := _m.Called(dealerHitsSoft17, cpuPlayerCount, countingEnabled, doubleAfterSplit, countingSystem)
	return ret.Get(0).(string)
}
