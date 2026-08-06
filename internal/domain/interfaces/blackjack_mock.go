//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBlackJackGame ブラックジャックゲームモック
type MockBlackJackGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockBlackJackGame) Reset() {
	_m.Called()
}

// PlayerBet モック
func (_m *MockBlackJackGame) PlayerBet(amount, ppBet, t3Bet, handCount int) error {
	ret := _m.Called(amount, ppBet, t3Bet, handCount)
	return ret.Error(0)
}

// PlayerInsurance モック
func (_m *MockBlackJackGame) PlayerInsurance() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerDeclineInsurance モック
func (_m *MockBlackJackGame) PlayerDeclineInsurance() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerHit モック
func (_m *MockBlackJackGame) PlayerHit() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerStand モック
func (_m *MockBlackJackGame) PlayerStand() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerDoubleDown モック
func (_m *MockBlackJackGame) PlayerDoubleDown() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerSplit モック
func (_m *MockBlackJackGame) PlayerSplit() error {
	ret := _m.Called()
	return ret.Error(0)
}

// GetPlayer モック
func (_m *MockBlackJackGame) GetPlayer() *domain.BlackJackPlayer {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.BlackJackPlayer); ok {
		return val
	}
	return nil
}

// GetDealer モック
func (_m *MockBlackJackGame) GetDealer() *domain.BlackJackPlayer {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.BlackJackPlayer); ok {
		return val
	}
	return nil
}

// GetPhase モック
func (_m *MockBlackJackGame) GetPhase() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetGameEndFlag モック
func (_m *MockBlackJackGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPlayerHands モック
func (_m *MockBlackJackGame) GetPlayerHands() []*domain.BlackJackHand {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.BlackJackHand); ok {
		return val
	}
	return nil
}

// GetCurrentHandIdx モック
func (_m *MockBlackJackGame) GetCurrentHandIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetInsuranceBet モック
func (_m *MockBlackJackGame) GetInsuranceBet() int {
	ret := _m.Called()
	return ret.Int(0)
}

// IsInsuranceAvailable モック
func (_m *MockBlackJackGame) IsInsuranceAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GameJudgmentForHand モック
func (_m *MockBlackJackGame) GameJudgmentForHand(handIdx int) domain.GameResult {
	ret := _m.Called(handIdx)
	return domain.GameResult(ret.Int(0))
}

// GameJudgment モック
func (_m *MockBlackJackGame) GameJudgment() domain.GameResult {
	ret := _m.Called()
	return domain.GameResult(ret.Int(0))
}

// PlayerSurrender モック
func (_m *MockBlackJackGame) PlayerSurrender() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerEarlySurrender モック
func (_m *MockBlackJackGame) PlayerEarlySurrender() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerDeclineEarlySurrender モック
func (_m *MockBlackJackGame) PlayerDeclineEarlySurrender() error {
	ret := _m.Called()
	return ret.Error(0)
}

// SetDeckCount モック
func (_m *MockBlackJackGame) SetDeckCount(count int) error {
	ret := _m.Called(count)
	return ret.Error(0)
}

// ToggleHint モック
func (_m *MockBlackJackGame) ToggleHint() {
	_m.Called()
}

// GetDeckCount モック
func (_m *MockBlackJackGame) GetDeckCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// IsHintEnabled モック
func (_m *MockBlackJackGame) IsHintEnabled() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetBasicStrategySuggestion モック
func (_m *MockBlackJackGame) GetBasicStrategySuggestion() domain.BJSuggestedAction {
	ret := _m.Called()
	return domain.BJSuggestedAction(ret.Int(0))
}

// SetConfig モック
func (_m *MockBlackJackGame) SetConfig(config domain.BlackJackConfig) error {
	ret := _m.Called(config)
	return ret.Error(0)
}

// GetConfig モック
func (_m *MockBlackJackGame) GetConfig() domain.BlackJackConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BlackJackConfig)
}

// GetRunningCount モック
func (_m *MockBlackJackGame) GetRunningCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTrueCount モック
func (_m *MockBlackJackGame) GetTrueCount() float64 {
	ret := _m.Called()
	return ret.Get(0).(float64)
}

// IsCountingEnabled モック
func (_m *MockBlackJackGame) IsCountingEnabled() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetCpuPlayers モック
func (_m *MockBlackJackGame) GetCpuPlayers() []*domain.BlackJackCpuSeat {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.BlackJackCpuSeat); ok {
		return val
	}
	return nil
}

// GetSideBetResults モック
func (_m *MockBlackJackGame) GetSideBetResults() []*domain.BJSideBetResult {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.BJSideBetResult); ok {
		return val
	}
	return nil
}

// GetPerfectPairsBet モック
func (_m *MockBlackJackGame) GetPerfectPairsBet() int {
	ret := _m.Called()
	return ret.Int(0)
}

// Get21Plus3Bet モック
func (_m *MockBlackJackGame) Get21Plus3Bet() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetDeckPenetration モック
func (_m *MockBlackJackGame) GetDeckPenetration() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetVariant モック
func (_m *MockBlackJackGame) GetVariant() *domain.BlackJackVariantConfig {
	v, _ := _m.Called().Get(0).(*domain.BlackJackVariantConfig)
	return v
}

// GetMultiHandCount モック
func (_m *MockBlackJackGame) GetMultiHandCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetBonusKeys モック
func (_m *MockBlackJackGame) GetBonusKeys() []string {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]string); ok {
		return v
	}
	return nil
}

// CanSurrenderHand モック
func (_m *MockBlackJackGame) CanSurrenderHand(handIdx int) bool {
	ret := _m.Called(handIdx)
	return ret.Bool(0)
}

// CanSurrenderCpuHand モック
func (_m *MockBlackJackGame) CanSurrenderCpuHand(cpuIdx, handIdx int) bool {
	ret := _m.Called(cpuIdx, handIdx)
	return ret.Bool(0)
}

// GetActionLog モック
func (_m *MockBlackJackGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return val
	}
	return nil
}
