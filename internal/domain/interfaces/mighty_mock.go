//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMightyGame マイティゲームモック
type MockMightyGame struct {
	mock.Mock
}

func (m *MockMightyGame) Reset() {
	m.Called()
}

func (m *MockMightyGame) NextRound() {
	m.Called()
}

func (m *MockMightyGame) PlayerBid(bid int, noTrump bool) error {
	args := m.Called(bid, noTrump)
	return args.Error(0)
}

func (m *MockMightyGame) CpuBid() {
	m.Called()
}

func (m *MockMightyGame) PlayerDeclareTrumpAndFriend(suit int, partnerSuit int, partnerVal int) error {
	args := m.Called(suit, partnerSuit, partnerVal)
	return args.Error(0)
}

func (m *MockMightyGame) CpuDeclareTrumpAndFriend() {
	m.Called()
}

func (m *MockMightyGame) PlayerExchangeKitty(discardIndices []int) error {
	args := m.Called(discardIndices)
	return args.Error(0)
}

func (m *MockMightyGame) CpuExchangeKitty() {
	m.Called()
}

func (m *MockMightyGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockMightyGame) PlayerPlayJokerLead(cardIndex int, demandSuit int) error {
	args := m.Called(cardIndex, demandSuit)
	return args.Error(0)
}

func (m *MockMightyGame) CpuPlay() {
	m.Called()
}

func (m *MockMightyGame) ResolveTrick() {
	m.Called()
}

func (m *MockMightyGame) NextTrick() {
	m.Called()
}

func (m *MockMightyGame) ScoreRound() {
	m.Called()
}

func (m *MockMightyGame) GetConfig() domain.MightyConfig {
	args := m.Called()
	return args.Get(0).(domain.MightyConfig)
}

func (m *MockMightyGame) SetConfig(cfg domain.MightyConfig) {
	m.Called(cfg)
}

func (m *MockMightyGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMightyGame) GetPhase() domain.MightyPhase {
	args := m.Called()
	return args.Get(0).(domain.MightyPhase)
}

func (m *MockMightyGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMightyGame) IsHumanBidTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMightyGame) IsHumanDeclareTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMightyGame) IsHumanExchangeTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMightyGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMightyGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMightyGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMightyGame) GetCurrentTrick() []*domain.MightyTrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.MightyTrickCard)
	}
	return nil
}

func (m *MockMightyGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMightyGame) GetPartnerCard() *domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

func (m *MockMightyGame) GetDeclarerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMightyGame) GetPartnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMightyGame) GetPartnerRevealed() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMightyGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMightyGame) GetBidPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMightyGame) GetKitty() []*domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

func (m *MockMightyGame) GetHighestBid() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMightyGame) GetHighestBidder() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMightyGame) GetWinningBidNoTrump() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMightyGame) GetWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMightyGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMightyGame) GetPlayer(i int) *domain.MightyPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.MightyPlayer)
	}
	return nil
}

func (m *MockMightyGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockMightyGame) GetHint() *domain.MightyHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.MightyHint)
	}
	return nil
}

func (m *MockMightyGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
