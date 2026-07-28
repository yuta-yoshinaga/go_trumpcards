//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPiquetGame Piquetゲームモック
type MockPiquetGame struct {
	mock.Mock
}

func (m *MockPiquetGame) Reset()    { m.Called() }
func (m *MockPiquetGame) NextDeal() { m.Called() }
func (m *MockPiquetGame) CpuPlay()  { m.Called() }

func (m *MockPiquetGame) ExchangeElder(discardIndices []int) error {
	args := m.Called(discardIndices)
	return args.Error(0)
}

func (m *MockPiquetGame) ExchangeYounger(discardIndices []int) error {
	args := m.Called(discardIndices)
	return args.Error(0)
}

func (m *MockPiquetGame) ResolveDeclaration() (*domain.PiquetDeclarationResult, error) {
	args := m.Called()
	var res *domain.PiquetDeclarationResult
	if v := args.Get(0); v != nil {
		res = v.(*domain.PiquetDeclarationResult)
	}
	return res, args.Error(1)
}

func (m *MockPiquetGame) PlayCard(cardIdx int) error {
	args := m.Called(cardIdx)
	return args.Error(0)
}

func (m *MockPiquetGame) GetConfig() domain.PiquetConfig {
	args := m.Called()
	return args.Get(0).(domain.PiquetConfig)
}

func (m *MockPiquetGame) SetConfig(cfg domain.PiquetConfig) { m.Called(cfg) }

func (m *MockPiquetGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockPiquetGame) GetPhase() domain.PiquetPhase {
	args := m.Called()
	return args.Get(0).(domain.PiquetPhase)
}

func (m *MockPiquetGame) GetDealNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPiquetGame) GetDealsPerPartie() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPiquetGame) GetElderIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPiquetGame) GetYoungerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPiquetGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockPiquetGame) GetPlayer(idx int) *domain.PiquetPlayer {
	args := m.Called(idx)
	if v := args.Get(0); v != nil {
		return v.(*domain.PiquetPlayer)
	}
	return nil
}

func (m *MockPiquetGame) GetPlayers() []*domain.PiquetPlayer {
	args := m.Called()
	return args.Get(0).([]*domain.PiquetPlayer)
}

func (m *MockPiquetGame) GetElderTalon() []*domain.Card {
	args := m.Called()
	return args.Get(0).([]*domain.Card)
}

func (m *MockPiquetGame) GetYoungerTalon() []*domain.Card {
	args := m.Called()
	return args.Get(0).([]*domain.Card)
}

func (m *MockPiquetGame) GetExchangeTurn() domain.PiquetExchangeTurn {
	args := m.Called()
	return args.Get(0).(domain.PiquetExchangeTurn)
}

func (m *MockPiquetGame) GetElderExchangedCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPiquetGame) GetYoungerExchangedCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPiquetGame) GetElderRevealedTalon() []*domain.Card {
	args := m.Called()
	return args.Get(0).([]*domain.Card)
}

func (m *MockPiquetGame) GetYoungerRevealedTalon() []*domain.Card {
	args := m.Called()
	return args.Get(0).([]*domain.Card)
}

func (m *MockPiquetGame) GetCarteBlanche(idx int) bool {
	args := m.Called(idx)
	return args.Bool(0)
}

func (m *MockPiquetGame) GetDeclStage() domain.PiquetDeclarationKind {
	args := m.Called()
	return args.Get(0).(domain.PiquetDeclarationKind)
}

func (m *MockPiquetGame) GetDeclResults() []*domain.PiquetDeclarationResult {
	args := m.Called()
	return args.Get(0).([]*domain.PiquetDeclarationResult)
}

func (m *MockPiquetGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockPiquetGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPiquetGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPiquetGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPiquetGame) GetTricksWon(idx int) int {
	args := m.Called(idx)
	return args.Int(0)
}

func (m *MockPiquetGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPiquetGame) GetLegalPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	return args.Get(0).([]int)
}

func (m *MockPiquetGame) GetHint(playerIdx int) *domain.PiquetHint {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.(*domain.PiquetHint)
	}
	return nil
}

func (m *MockPiquetGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
