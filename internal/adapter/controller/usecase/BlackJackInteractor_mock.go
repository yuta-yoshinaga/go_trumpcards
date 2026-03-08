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
func (_m *MockBlackJackInteractor) Bet(amount, ppBet, t3Bet, handCount int) string {
	ret := _m.Called(amount, ppBet, t3Bet, handCount)
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

// EarlySurrender モック
func (_m *MockBlackJackInteractor) EarlySurrender() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// DeclineEarlySurrender モック
func (_m *MockBlackJackInteractor) DeclineEarlySurrender() string {
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

// SetDeckPenetration モック
func (_m *MockBlackJackInteractor) SetDeckPenetration(penetration int) string {
	ret := _m.Called(penetration)
	return ret.Get(0).(string)
}

// SetCpuPlayerCount モック
func (_m *MockBlackJackInteractor) SetCpuPlayerCount(count int) string {
	ret := _m.Called(count)
	return ret.Get(0).(string)
}

// SetSurrenderRule モック
func (_m *MockBlackJackInteractor) SetSurrenderRule(rule int) string {
	ret := _m.Called(rule)
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockBlackJackInteractor) ResetWithConfig(dealerHitsSoft17 bool, cpuPlayerCount int, countingEnabled bool, doubleAfterSplit bool, countingSystem int, deckPenetration int, surrenderRule int) string {
	ret := _m.Called(dealerHitsSoft17, cpuPlayerCount, countingEnabled, doubleAfterSplit, countingSystem, deckPenetration, surrenderRule)
	return ret.Get(0).(string)
}
