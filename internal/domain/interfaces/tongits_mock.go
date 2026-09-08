//go:build test
// +build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTongitsGame モック
type MockTongitsGame struct {
	mock.Mock
}

func (m *MockTongitsGame) Reset()                            { m.Called() }
func (m *MockTongitsGame) NextRound()                        { m.Called() }
func (m *MockTongitsGame) PlayerDrawFromStock() error        { return m.Called().Error(0) }
func (m *MockTongitsGame) PlayerDrawFromDiscard() error      { return m.Called().Error(0) }
func (m *MockTongitsGame) PlayerDiscard(cardIndex int) error { return m.Called(cardIndex).Error(0) }
func (m *MockTongitsGame) PlayerKnock(cardIndex int) error   { return m.Called(cardIndex).Error(0) }
func (m *MockTongitsGame) CpuPlay()                          { m.Called() }
func (m *MockTongitsGame) ScoreRound()                       { m.Called() }
func (m *MockTongitsGame) GetConfig() domain.TongitsConfig {
	return m.Called().Get(0).(domain.TongitsConfig)
}
func (m *MockTongitsGame) SetConfig(cfg domain.TongitsConfig) { m.Called(cfg) }
func (m *MockTongitsGame) GetGameEndFlag() bool               { return m.Called().Bool(0) }
func (m *MockTongitsGame) GetPhase() domain.TongitsPhase {
	return m.Called().Get(0).(domain.TongitsPhase)
}
func (m *MockTongitsGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockTongitsGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockTongitsGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockTongitsGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockTongitsGame) GetDrawPileCount() int { return m.Called().Int(0) }
func (m *MockTongitsGame) GetWinnerIdx() int     { return m.Called().Int(0) }
func (m *MockTongitsGame) GetPlayerCnt() int     { return m.Called().Int(0) }
func (m *MockTongitsGame) GetPlayer(i int) *domain.TongitsPlayer {
	return m.Called(i).Get(0).(*domain.TongitsPlayer)
}

// GetBestDeadwood は1枚捨てて到達できる最小デッドウッドを返すモック。
func (m *MockTongitsGame) GetBestDeadwood(playerIdx int) (int, int) {
	ret := m.Called(playerIdx)
	return ret.Int(0), ret.Int(1)
}

func (m *MockTongitsGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockTongitsGame) GetKnockerIdx() int { return m.Called().Int(0) }
func (m *MockTongitsGame) GetKnockerMelds() [][]*domain.Card {
	return m.Called().Get(0).([][]*domain.Card)
}
func (m *MockTongitsGame) GetKnockerDeadwood() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockTongitsGame) GetOpponentMelds() [][]*domain.Card {
	return m.Called().Get(0).([][]*domain.Card)
}
func (m *MockTongitsGame) GetOpponentDeadwood() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockTongitsGame) GetIsTongits() bool  { return m.Called().Bool(0) }
func (m *MockTongitsGame) GetIsUndercut() bool { return m.Called().Bool(0) }
