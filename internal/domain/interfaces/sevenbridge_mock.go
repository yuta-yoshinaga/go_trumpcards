//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSevenBridgeGame セブンブリッジゲームのモック
type MockSevenBridgeGame struct {
	mock.Mock
}

func (m *MockSevenBridgeGame) Reset()                       { m.Called() }
func (m *MockSevenBridgeGame) NextRound()                   { m.Called() }
func (m *MockSevenBridgeGame) PlayerDrawFromStock() error   { return m.Called().Error(0) }
func (m *MockSevenBridgeGame) PlayerClaimPon(c []int) error { return m.Called(c).Error(0) }
func (m *MockSevenBridgeGame) PlayerClaimChi(c []int) error { return m.Called(c).Error(0) }
func (m *MockSevenBridgeGame) PlayerMeld(c []int) error     { return m.Called(c).Error(0) }
func (m *MockSevenBridgeGame) PlayerLayoff(t, mi, ci int) error {
	return m.Called(t, mi, ci).Error(0)
}
func (m *MockSevenBridgeGame) PlayerDiscard(i int) error { return m.Called(i).Error(0) }
func (m *MockSevenBridgeGame) SuggestMeld(playerIdx int) []int {
	ret := m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}
func (m *MockSevenBridgeGame) SuggestDiscard(playerIdx int) int {
	return m.Called(playerIdx).Int(0)
}
func (m *MockSevenBridgeGame) CpuPlay() { m.Called() }
func (m *MockSevenBridgeGame) GetConfig() domain.SevenBridgeConfig {
	return m.Called().Get(0).(domain.SevenBridgeConfig)
}
func (m *MockSevenBridgeGame) SetConfig(cfg domain.SevenBridgeConfig) { m.Called(cfg) }
func (m *MockSevenBridgeGame) GetGameEndFlag() bool                   { return m.Called().Bool(0) }
func (m *MockSevenBridgeGame) GetPhase() domain.SevenBridgePhase {
	return m.Called().Get(0).(domain.SevenBridgePhase)
}
func (m *MockSevenBridgeGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockSevenBridgeGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockSevenBridgeGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockSevenBridgeGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockSevenBridgeGame) GetDiscardPile() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockSevenBridgeGame) GetDrawPileCount() int { return m.Called().Int(0) }
func (m *MockSevenBridgeGame) GetWinnerIdx() int     { return m.Called().Int(0) }
func (m *MockSevenBridgeGame) GetPlayerCnt() int     { return m.Called().Int(0) }
func (m *MockSevenBridgeGame) GetPlayer(i int) *domain.SevenBridgePlayer {
	return m.Called(i).Get(0).(*domain.SevenBridgePlayer)
}
func (m *MockSevenBridgeGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockSevenBridgeGame) GetRoundWinnerIdx() int   { return m.Called().Int(0) }
func (m *MockSevenBridgeGame) GetClaimedThisTurn() bool { return m.Called().Bool(0) }
