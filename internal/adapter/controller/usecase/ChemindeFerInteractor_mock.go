//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockChemindeFerInteractor シュマン・ド・フェールインタラクターモック
type MockChemindeFerInteractor struct {
	mock.Mock
}

func (m *MockChemindeFerInteractor) Reset() string { return m.Called().String(0) }

func (m *MockChemindeFerInteractor) ResetWithConfig(cfg domain.ChemindeFerConfig) string {
	return m.Called(cfg).String(0)
}

func (m *MockChemindeFerInteractor) SetStake(amount int) string {
	return m.Called(amount).String(0)
}

func (m *MockChemindeFerInteractor) PlaceBet(seatIdx, amount int) string {
	return m.Called(seatIdx, amount).String(0)
}

func (m *MockChemindeFerInteractor) PunterDraw() string { return m.Called().String(0) }

func (m *MockChemindeFerInteractor) PunterStand() string { return m.Called().String(0) }

func (m *MockChemindeFerInteractor) BankerDraw() string { return m.Called().String(0) }

func (m *MockChemindeFerInteractor) BankerStand() string { return m.Called().String(0) }

func (m *MockChemindeFerInteractor) DrawOrStand(draw bool) string {
	return m.Called(draw).String(0)
}

func (m *MockChemindeFerInteractor) PassBank() string { return m.Called().String(0) }

func (m *MockChemindeFerInteractor) NextRound() string { return m.Called().String(0) }

func (m *MockChemindeFerInteractor) GiveUp() string { return m.Called().String(0) }

func (m *MockChemindeFerInteractor) GetConfig() domain.ChemindeFerConfig {
	return m.Called().Get(0).(domain.ChemindeFerConfig)
}

func (m *MockChemindeFerInteractor) Hint() string { return m.Called().String(0) }

func (m *MockChemindeFerInteractor) ActionLog() string { return m.Called().String(0) }

func (m *MockChemindeFerInteractor) Snapshot() ([]byte, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}
