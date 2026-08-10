//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKlaberjassGame モック
type MockKlaberjassGame struct {
	mock.Mock
}

func (m *MockKlaberjassGame) Reset()                       { m.Called() }
func (m *MockKlaberjassGame) NextDeal() error              { return m.Called().Error(0) }
func (m *MockKlaberjassGame) AcceptTrump(player int) error { return m.Called(player).Error(0) }
func (m *MockKlaberjassGame) CallTrump(player, suit int) error {
	return m.Called(player, suit).Error(0)
}
func (m *MockKlaberjassGame) Pass(player int) error          { return m.Called(player).Error(0) }
func (m *MockKlaberjassGame) Schmeiss(player int) error      { return m.Called(player).Error(0) }
func (m *MockKlaberjassGame) PlayCard(player, idx int) error { return m.Called(player, idx).Error(0) }
func (m *MockKlaberjassGame) CpuPlay()                       { m.Called() }
func (m *MockKlaberjassGame) AnswerSchmeiss(player int, accept bool) error {
	return m.Called(player, accept).Error(0)
}
func (m *MockKlaberjassGame) KlaberjassValidPlays(player int) []int {
	return m.Called(player).Get(0).([]int)
}
func (m *MockKlaberjassGame) GetConfig() domain.KlaberjassConfig {
	return m.Called().Get(0).(domain.KlaberjassConfig)
}
func (m *MockKlaberjassGame) SetConfig(cfg domain.KlaberjassConfig) { m.Called(cfg) }
func (m *MockKlaberjassGame) GetGameEndFlag() bool                  { return m.Called().Bool(0) }
func (m *MockKlaberjassGame) GetPhase() domain.KlaberjassPhase {
	return m.Called().Get(0).(domain.KlaberjassPhase)
}
func (m *MockKlaberjassGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockKlaberjassGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockKlaberjassGame) GetBidPlayerIdx() int     { return m.Called().Int(0) }
func (m *MockKlaberjassGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockKlaberjassGame) GetTrumpSuit() int        { return m.Called().Int(0) }
func (m *MockKlaberjassGame) GetTurnUpCard() *domain.Card {
	c, _ := m.Called().Get(0).(*domain.Card)
	return c
}
func (m *MockKlaberjassGame) GetMakerIdx() int { return m.Called().Int(0) }
func (m *MockKlaberjassGame) GetTrick() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockKlaberjassGame) GetTrickLeaderIdx() int    { return m.Called().Int(0) }
func (m *MockKlaberjassGame) GetTrickNumber() int       { return m.Called().Int(0) }
func (m *MockKlaberjassGame) GetHandPoints(idx int) int { return m.Called(idx).Int(0) }
func (m *MockKlaberjassGame) GetSequences(idx int) []*domain.KlaberjassSequence {
	return m.Called(idx).Get(0).([]*domain.KlaberjassSequence)
}
func (m *MockKlaberjassGame) GetSequenceWinner() int { return m.Called().Int(0) }
func (m *MockKlaberjassGame) GetBelaHolder() int     { return m.Called().Int(0) }
func (m *MockKlaberjassGame) IsBelaScored() bool     { return m.Called().Bool(0) }
func (m *MockKlaberjassGame) IsDixUsed() bool        { return m.Called().Bool(0) }
func (m *MockKlaberjassGame) IsBete() bool           { return m.Called().Bool(0) }
func (m *MockKlaberjassGame) GetSchmeissBy() int     { return m.Called().Int(0) }
func (m *MockKlaberjassGame) GetScore(idx int) int   { return m.Called(idx).Int(0) }
func (m *MockKlaberjassGame) GetDealNumber() int     { return m.Called().Int(0) }
func (m *MockKlaberjassGame) GetWinnerIdx() int      { return m.Called().Int(0) }
func (m *MockKlaberjassGame) GetPlayers() []*domain.KlaberjassPlayer {
	return m.Called().Get(0).([]*domain.KlaberjassPlayer)
}
func (m *MockKlaberjassGame) GetPlayer(idx int) *domain.KlaberjassPlayer {
	p, _ := m.Called(idx).Get(0).(*domain.KlaberjassPlayer)
	return p
}
func (m *MockKlaberjassGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}

// GetLastTrickWinner モック
func (m *MockKlaberjassGame) GetLastTrickWinner() int { return m.Called().Int(0) }
