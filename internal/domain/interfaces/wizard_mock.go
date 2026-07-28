//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockWizardGame ウィザードゲームモック
type MockWizardGame struct {
	mock.Mock
}

func (m *MockWizardGame) Reset() {
	m.Called()
}

func (m *MockWizardGame) NextRound() {
	m.Called()
}

func (m *MockWizardGame) PlayerBid(bid int) error {
	args := m.Called(bid)
	return args.Error(0)
}

func (m *MockWizardGame) CpuBid() {
	m.Called()
}

func (m *MockWizardGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockWizardGame) CpuPlay() {
	m.Called()
}

func (m *MockWizardGame) ResolveTrick() {
	m.Called()
}

func (m *MockWizardGame) NextTrick() {
	m.Called()
}

func (m *MockWizardGame) ScoreRound() {
	m.Called()
}

func (m *MockWizardGame) GetConfig() domain.WizardConfig {
	args := m.Called()
	return args.Get(0).(domain.WizardConfig)
}

func (m *MockWizardGame) SetConfig(cfg domain.WizardConfig) {
	m.Called(cfg)
}

func (m *MockWizardGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockWizardGame) GetPhase() domain.WizardPhase {
	args := m.Called()
	return args.Get(0).(domain.WizardPhase)
}

func (m *MockWizardGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockWizardGame) IsHumanBidTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockWizardGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWizardGame) GetTotalRounds() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWizardGame) GetHandSize() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWizardGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWizardGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWizardGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockWizardGame) GetTrumpCard() *domain.Card {
	args := m.Called()
	if val, ok := args.Get(0).(*domain.Card); ok {
		return val
	}
	return nil
}

func (m *MockWizardGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWizardGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWizardGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWizardGame) GetBidPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWizardGame) GetRestrictedBid() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWizardGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWizardGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWizardGame) GetPlayer(i int) *domain.WizardPlayer {
	args := m.Called(i)
	return args.Get(0).(*domain.WizardPlayer)
}

func (m *MockWizardGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	return args.Get(0).([]int)
}

// GetHint モック
func (m *MockWizardGame) GetHint() *domain.WizardHint {
	args := m.Called()
	if val, ok := args.Get(0).(*domain.WizardHint); ok {
		return val
	}
	return nil
}

func (m *MockWizardGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	return args.Get(0).([]*domain.ActionLogEntry)
}
