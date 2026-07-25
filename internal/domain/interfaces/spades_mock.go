//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSpadesGame スペードゲームモック
type MockSpadesGame struct {
	mock.Mock
}

func (m *MockSpadesGame) Reset() {
	m.Called()
}

func (m *MockSpadesGame) NextRound() {
	m.Called()
}

func (m *MockSpadesGame) PlayerBid(bid int) error {
	args := m.Called(bid)
	return args.Error(0)
}

func (m *MockSpadesGame) CpuBid() {
	m.Called()
}

func (m *MockSpadesGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockSpadesGame) CpuPlay() {
	m.Called()
}

func (m *MockSpadesGame) ResolveTrick() {
	m.Called()
}

func (m *MockSpadesGame) NextTrick() {
	m.Called()
}

func (m *MockSpadesGame) ScoreRound() {
	m.Called()
}

func (m *MockSpadesGame) GetConfig() domain.SpadesConfig {
	args := m.Called()
	return args.Get(0).(domain.SpadesConfig)
}

func (m *MockSpadesGame) SetConfig(cfg domain.SpadesConfig) {
	m.Called(cfg)
}

func (m *MockSpadesGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockSpadesGame) GetPhase() domain.SpadesPhase {
	args := m.Called()
	return args.Get(0).(domain.SpadesPhase)
}

func (m *MockSpadesGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockSpadesGame) IsHumanBidTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockSpadesGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSpadesGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSpadesGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSpadesGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockSpadesGame) GetSpadesBroken() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockSpadesGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSpadesGame) GetBidPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSpadesGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSpadesGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSpadesGame) GetPlayer(i int) *domain.SpadesPlayer {
	args := m.Called(i)
	return args.Get(0).(*domain.SpadesPlayer)
}

// GetHint モック
func (m *MockSpadesGame) GetHint() *domain.SpadesHint {
	args := m.Called()
	if val, ok := args.Get(0).(*domain.SpadesHint); ok {
		return val
	}
	return nil
}

func (m *MockSpadesGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	return args.Get(0).([]*domain.ActionLogEntry)
}
