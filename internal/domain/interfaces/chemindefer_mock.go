//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockChemindeFerGame シュマン・ド・フェールゲームモック
type MockChemindeFerGame struct {
	mock.Mock
}

func (m *MockChemindeFerGame) Reset() {
	m.Called()
}

func (m *MockChemindeFerGame) SetStake(amount int) error {
	return m.Called(amount).Error(0)
}

func (m *MockChemindeFerGame) PlaceBet(seatIdx, amount int) error {
	return m.Called(seatIdx, amount).Error(0)
}

func (m *MockChemindeFerGame) PunterDraw() error { return m.Called().Error(0) }

func (m *MockChemindeFerGame) PunterStand() error { return m.Called().Error(0) }

func (m *MockChemindeFerGame) BankerDraw() error { return m.Called().Error(0) }

func (m *MockChemindeFerGame) BankerStand() error { return m.Called().Error(0) }

func (m *MockChemindeFerGame) PassBank() error { return m.Called().Error(0) }

func (m *MockChemindeFerGame) NextRound() error { return m.Called().Error(0) }

func (m *MockChemindeFerGame) GiveUp() { m.Called() }

func (m *MockChemindeFerGame) CpuPlay() { m.Called() }

func (m *MockChemindeFerGame) GetConfig() domain.ChemindeFerConfig {
	return m.Called().Get(0).(domain.ChemindeFerConfig)
}

func (m *MockChemindeFerGame) SetConfig(cfg domain.ChemindeFerConfig) { m.Called(cfg) }

func (m *MockChemindeFerGame) GetPhase() domain.ChemindeFerPhase {
	return m.Called().Get(0).(domain.ChemindeFerPhase)
}

func (m *MockChemindeFerGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockChemindeFerGame) IsHumanTurn() bool { return m.Called().Bool(0) }

func (m *MockChemindeFerGame) GetBankerIdx() int { return m.Called().Int(0) }

func (m *MockChemindeFerGame) GetBetTurn() int { return m.Called().Int(0) }

func (m *MockChemindeFerGame) GetStake() int { return m.Called().Int(0) }

func (m *MockChemindeFerGame) GetRemainingStake() int { return m.Called().Int(0) }

func (m *MockChemindeFerGame) GetTotalBet() int { return m.Called().Int(0) }

func (m *MockChemindeFerGame) StakeRangeFor(seatIdx int) (int, int) {
	args := m.Called(seatIdx)
	return args.Int(0), args.Int(1)
}

func (m *MockChemindeFerGame) BetRangeFor(seatIdx int) (int, int) {
	args := m.Called(seatIdx)
	return args.Int(0), args.Int(1)
}

func (m *MockChemindeFerGame) GetRepresentativeIdx() int { return m.Called().Int(0) }

func (m *MockChemindeFerGame) PunterMayChoose() bool { return m.Called().Bool(0) }

func (m *MockChemindeFerGame) GetBankerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockChemindeFerGame) GetPunterHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockChemindeFerGame) GetBankerTotal() int { return m.Called().Int(0) }

func (m *MockChemindeFerGame) GetPunterTotal() int { return m.Called().Int(0) }

func (m *MockChemindeFerGame) GetPunterDrew() bool { return m.Called().Bool(0) }

func (m *MockChemindeFerGame) GetResult() domain.ChemindeFerResult {
	return m.Called().Get(0).(domain.ChemindeFerResult)
}

func (m *MockChemindeFerGame) GetRoundNumber() int { return m.Called().Int(0) }

func (m *MockChemindeFerGame) GetPlayer(i int) *domain.ChemindeFerPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.ChemindeFerPlayer)
}

func (m *MockChemindeFerGame) GetRemainingCards() int { return m.Called().Int(0) }

func (m *MockChemindeFerGame) GetHint() *domain.ChemindeFerHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.ChemindeFerHint)
}

func (m *MockChemindeFerGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}
