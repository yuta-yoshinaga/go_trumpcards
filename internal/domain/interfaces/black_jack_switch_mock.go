//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBlackJackSwitchGame ブラックジャック・スイッチゲームモック
type MockBlackJackSwitchGame struct {
	mock.Mock
}

func (m *MockBlackJackSwitchGame) Reset()                     { m.Called() }
func (m *MockBlackJackSwitchGame) PlayerBet(amount int) error { return m.Called(amount).Error(0) }
func (m *MockBlackJackSwitchGame) PlayerSwitch() error        { return m.Called().Error(0) }
func (m *MockBlackJackSwitchGame) PlayerKeep() error          { return m.Called().Error(0) }
func (m *MockBlackJackSwitchGame) PlayerHit() error           { return m.Called().Error(0) }
func (m *MockBlackJackSwitchGame) PlayerStand() error         { return m.Called().Error(0) }
func (m *MockBlackJackSwitchGame) PlayerDoubleDown() error    { return m.Called().Error(0) }

func (m *MockBlackJackSwitchGame) GetPlayer() *domain.BlackJackPlayer {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.BlackJackPlayer)
}

func (m *MockBlackJackSwitchGame) GetDealer() *domain.BlackJackPlayer {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.BlackJackPlayer)
}

// SwitchPreviewScores は 2 枚目を入れ替えた場合の両ハンドの得点を返す。
func (m *MockBlackJackSwitchGame) SwitchPreviewScores() (int, int, bool) {
	args := m.Called()
	return args.Int(0), args.Int(1), args.Bool(2)
}

func (m *MockBlackJackSwitchGame) GetHands() []*domain.BlackJackHand {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.BlackJackHand)
}

func (m *MockBlackJackSwitchGame) GetCurrentHandIdx() int { return m.Called().Int(0) }
func (m *MockBlackJackSwitchGame) GetPhase() int          { return m.Called().Int(0) }
func (m *MockBlackJackSwitchGame) GetGameEndFlag() bool   { return m.Called().Bool(0) }
func (m *MockBlackJackSwitchGame) IsSwitched() bool       { return m.Called().Bool(0) }
func (m *MockBlackJackSwitchGame) IsDealerPushed22() bool { return m.Called().Bool(0) }

func (m *MockBlackJackSwitchGame) GetHandResults() []domain.GameResult {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]domain.GameResult)
}

func (m *MockBlackJackSwitchGame) GetHandPayouts() []int {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockBlackJackSwitchGame) GetTotalPayout() int { return m.Called().Int(0) }

func (m *MockBlackJackSwitchGame) GetOverallResult() domain.GameResult {
	return m.Called().Get(0).(domain.GameResult)
}

func (m *MockBlackJackSwitchGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}
