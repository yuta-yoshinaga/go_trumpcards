//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockNapoleonGame ナポレオンゲームモック
type MockNapoleonGame struct {
	mock.Mock
}

func (m *MockNapoleonGame) Reset() {
	m.Called()
}

func (m *MockNapoleonGame) NextRound() {
	m.Called()
}

func (m *MockNapoleonGame) PlayerBid(bid int) error {
	args := m.Called(bid)
	return args.Error(0)
}

func (m *MockNapoleonGame) CpuBid() {
	m.Called()
}

func (m *MockNapoleonGame) PlayerDeclareTrump(suit int, adjSuit int, adjVal int) error {
	args := m.Called(suit, adjSuit, adjVal)
	return args.Error(0)
}

func (m *MockNapoleonGame) CpuDeclareTrump() {
	m.Called()
}

func (m *MockNapoleonGame) PlayerExchangeKitty(discardIndex int) error {
	args := m.Called(discardIndex)
	return args.Error(0)
}

func (m *MockNapoleonGame) CpuExchangeKitty() {
	m.Called()
}

func (m *MockNapoleonGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockNapoleonGame) CpuPlay() {
	m.Called()
}

func (m *MockNapoleonGame) ResolveTrick() {
	m.Called()
}

func (m *MockNapoleonGame) NextTrick() {
	m.Called()
}

func (m *MockNapoleonGame) ScoreRound() {
	m.Called()
}

func (m *MockNapoleonGame) GetConfig() domain.NapoleonConfig {
	args := m.Called()
	return args.Get(0).(domain.NapoleonConfig)
}

func (m *MockNapoleonGame) SetConfig(cfg domain.NapoleonConfig) {
	m.Called(cfg)
}

func (m *MockNapoleonGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockNapoleonGame) GetPhase() domain.NapoleonPhase {
	args := m.Called()
	return args.Get(0).(domain.NapoleonPhase)
}

func (m *MockNapoleonGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockNapoleonGame) IsHumanBidTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockNapoleonGame) IsHumanDeclareTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockNapoleonGame) IsHumanExchangeTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockNapoleonGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNapoleonGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNapoleonGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNapoleonGame) GetCurrentTrick() []*domain.NapoleonTrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.NapoleonTrickCard)
}

func (m *MockNapoleonGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNapoleonGame) GetAdjutantCard() *domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.Card)
}

func (m *MockNapoleonGame) GetNapoleonIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNapoleonGame) GetAdjutantIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNapoleonGame) GetAdjutantRevealed() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockNapoleonGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNapoleonGame) GetBidPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNapoleonGame) GetKitty() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockNapoleonGame) GetHighestBid() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNapoleonGame) GetHighestBidder() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNapoleonGame) GetWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

// GetHumanIdx は人間プレイヤーのインデックスを返すモック。
func (m *MockNapoleonGame) GetHumanIdx() int {
	return m.Called().Int(0)
}

func (m *MockNapoleonGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNapoleonGame) GetPlayer(i int) *domain.NapoleonPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.NapoleonPlayer)
}

func (m *MockNapoleonGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockNapoleonGame) GetHint() *domain.NapoleonHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.NapoleonHint)
}

func (m *MockNapoleonGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}
