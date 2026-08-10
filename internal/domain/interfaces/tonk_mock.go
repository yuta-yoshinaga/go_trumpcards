//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTonkGame モック
type MockTonkGame struct {
	mock.Mock
}

func (m *MockTonkGame) Reset()                            { m.Called() }
func (m *MockTonkGame) NextRound()                        { m.Called() }
func (m *MockTonkGame) PlayerDrawFromStock() error        { return m.Called().Error(0) }
func (m *MockTonkGame) PlayerDrawFromDiscard() error      { return m.Called().Error(0) }
func (m *MockTonkGame) PlayerDiscard(cardIndex int) error { return m.Called(cardIndex).Error(0) }
func (m *MockTonkGame) PlayerKnock(cardIndex int) error   { return m.Called(cardIndex).Error(0) }
func (m *MockTonkGame) CpuPlay()                          { m.Called() }
func (m *MockTonkGame) ScoreRound()                       { m.Called() }
func (m *MockTonkGame) GetConfig() domain.TonkConfig {
	return m.Called().Get(0).(domain.TonkConfig)
}
func (m *MockTonkGame) SetConfig(cfg domain.TonkConfig) { m.Called(cfg) }
func (m *MockTonkGame) GetGameEndFlag() bool            { return m.Called().Bool(0) }
func (m *MockTonkGame) GetPhase() domain.TonkPhase {
	return m.Called().Get(0).(domain.TonkPhase)
}
func (m *MockTonkGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockTonkGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockTonkGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockTonkGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockTonkGame) GetDrawPileCount() int { return m.Called().Int(0) }
func (m *MockTonkGame) GetWinnerIdx() int     { return m.Called().Int(0) }
func (m *MockTonkGame) GetPlayerCnt() int     { return m.Called().Int(0) }
func (m *MockTonkGame) GetPlayer(i int) *domain.TonkPlayer {
	return m.Called(i).Get(0).(*domain.TonkPlayer)
}

// GetBestDeadwood は1枚捨てて到達できる最小デッドウッドを返すモック。
func (m *MockTonkGame) GetBestDeadwood(playerIdx int) (int, int) {
	ret := m.Called(playerIdx)
	return ret.Int(0), ret.Int(1)
}

func (m *MockTonkGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockTonkGame) GetKnockerIdx() int { return m.Called().Int(0) }
func (m *MockTonkGame) GetKnockerMelds() [][]*domain.Card {
	return m.Called().Get(0).([][]*domain.Card)
}
func (m *MockTonkGame) GetKnockerDeadwood() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockTonkGame) GetOpponentMelds() [][]*domain.Card {
	return m.Called().Get(0).([][]*domain.Card)
}
func (m *MockTonkGame) GetOpponentDeadwood() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockTonkGame) GetIsTonk() bool     { return m.Called().Bool(0) }
func (m *MockTonkGame) GetIsUndercut() bool { return m.Called().Bool(0) }
