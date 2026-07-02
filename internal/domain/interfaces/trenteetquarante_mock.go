//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTrenteEtQuaranteGame はトラント・エ・カラント (Trente et Quarante) のゲームモック。
type MockTrenteEtQuaranteGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockTrenteEtQuaranteGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockTrenteEtQuaranteGame) NextRound() { _m.Called() }

// PlaceBet モック
func (_m *MockTrenteEtQuaranteGame) PlaceBet(bet domain.TrenteEtQuaranteBet, stake int) error {
	ret := _m.Called(bet, stake)
	return ret.Error(0)
}

// GetConfig モック
func (_m *MockTrenteEtQuaranteGame) GetConfig() domain.TrenteEtQuaranteConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TrenteEtQuaranteConfig)
}

// SetConfig モック
func (_m *MockTrenteEtQuaranteGame) SetConfig(cfg domain.TrenteEtQuaranteConfig) { _m.Called(cfg) }

// GetPhase モック
func (_m *MockTrenteEtQuaranteGame) GetPhase() domain.TrenteEtQuarantePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.TrenteEtQuarantePhase)
}

// GetGameEndFlag モック
func (_m *MockTrenteEtQuaranteGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetCurrentBet モック
func (_m *MockTrenteEtQuaranteGame) GetCurrentBet() domain.TrenteEtQuaranteBet {
	ret := _m.Called()
	return ret.Get(0).(domain.TrenteEtQuaranteBet)
}

// GetStake モック
func (_m *MockTrenteEtQuaranteGame) GetStake() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetNoirRow モック
func (_m *MockTrenteEtQuaranteGame) GetNoirRow() []*domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

// GetRougeRow モック
func (_m *MockTrenteEtQuaranteGame) GetRougeRow() []*domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

// GetNoirTotal モック
func (_m *MockTrenteEtQuaranteGame) GetNoirTotal() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRougeTotal モック
func (_m *MockTrenteEtQuaranteGame) GetRougeTotal() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinningRow モック
func (_m *MockTrenteEtQuaranteGame) GetWinningRow() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetFirstCardRed モック
func (_m *MockTrenteEtQuaranteGame) GetFirstCardRed() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRefait モック
func (_m *MockTrenteEtQuaranteGame) GetRefait() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetResult モック
func (_m *MockTrenteEtQuaranteGame) GetResult() domain.TrenteEtQuaranteResult {
	ret := _m.Called()
	return ret.Get(0).(domain.TrenteEtQuaranteResult)
}

// GetPayout モック
func (_m *MockTrenteEtQuaranteGame) GetPayout() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetChips モック
func (_m *MockTrenteEtQuaranteGame) GetChips() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundNumber モック
func (_m *MockTrenteEtQuaranteGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRemainingDeck モック
func (_m *MockTrenteEtQuaranteGame) GetRemainingDeck() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockTrenteEtQuaranteGame) GetPlayer() *domain.TrenteEtQuarantePlayer {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.TrenteEtQuarantePlayer)
	}
	return nil
}

// GetHint モック
func (_m *MockTrenteEtQuaranteGame) GetHint() *domain.TrenteEtQuaranteHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.TrenteEtQuaranteHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockTrenteEtQuaranteGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
