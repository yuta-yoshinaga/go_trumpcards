//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupMarriageCuiMock(phase domain.MarriagePhase, gameEnd bool) (*interfaces.MockMarriageGame, []*domain.MarriagePlayer) {
	m := new(interfaces.MockMarriageGame)
	players := []*domain.MarriagePlayer{
		domain.NewMarriagePlayer(true),
		domain.NewMarriagePlayer(false),
	}
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	m.On("GetRoundNumber").Return(1)
	m.On("GetTargetRounds").Return(3)
	m.On("GetDrawPileCount").Return(60)
	m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
	m.On("GetWildJoker").Return(domain.NewCard(domain.CardDesignDiamond, 9, false))
	m.On("GetGameEndFlag").Return(gameEnd)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	winner := -1
	if gameEnd {
		winner = 0
	}
	m.On("GetWinnerIdx").Return(winner)
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("PlayerDeadwoodValue", 0).Return(10).Maybe()
	m.On("PlayerDeadwoodValue", 1).Return(20).Maybe()
	m.On("PlayerMaalValue", 0).Return(3).Maybe()
	m.On("PlayerMaalValue", 1).Return(4).Maybe()
	m.On("PlayerHasPureSequence", 0).Return(false).Maybe()
	m.On("PlayerHasPureSequence", 1).Return(false).Maybe()
	return m, players
}

func TestMarriageCuiPresenter_Output(t *testing.T) {
	p := new(presenter.MarriageCuiPresenter)

	t.Run("draw phase", func(t *testing.T) {
		m, _ := setupMarriageCuiMock(domain.MarriagePhaseDraw, false)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
		assert.Contains(t, out, "マリッジ")
		assert.Contains(t, out, "ラウンド")
		assert.Contains(t, out, "ワイルドジョーカー")
		assert.Contains(t, out, "マール3点")
		assert.NotContains(t, out, "マール4点", "通常時は CPU のマールを隠す")
	})

	t.Run("discard phase shows deadwood and unmet pure sequence", func(t *testing.T) {
		m, _ := setupMarriageCuiMock(domain.MarriagePhaseDiscard, false)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
		assert.Contains(t, out, "ディスカードフェーズ")
		assert.Contains(t, out, "デッドウッド: 10 点")
		assert.Contains(t, out, "純シーケンス未成立")
	})

	t.Run("discard phase explains the scale the deadwood number uses", func(t *testing.T) {
		m, _ := setupMarriageCuiMock(domain.MarriagePhaseDiscard, false)
		out := p.Output(m, nil)
		assert.Contains(t, out, "点数:")

		// The legend is only worth printing if it matches what actually scores.
		// A Gin Rummy player expects the ace to be 1 and a joker to cost its
		// face value; both are wrong here, which is why the line exists.
		wildRank := 5
		ace := domain.NewCard(domain.CardDesignSpade, 1, false)
		king := domain.NewCard(domain.CardDesignHeart, 13, false)
		seven := domain.NewCard(domain.CardDesignClover, 7, false)
		wild := domain.NewCard(domain.CardDesignDiamond, wildRank, false)
		assert.Equal(t, 10, domain.MarriageCardPoints(ace, wildRank), "legend says A = 10")
		assert.Equal(t, 10, domain.MarriageCardPoints(king, wildRank), "legend says K = 10")
		assert.Equal(t, 7, domain.MarriageCardPoints(seven, wildRank), "legend says 2-9 = face value")
		assert.Equal(t, 0, domain.MarriageCardPoints(wild, wildRank), "legend says wild = 0")
	})

	t.Run("discard phase shows pure sequence met", func(t *testing.T) {
		m, _ := setupMarriageCuiMock(domain.MarriagePhaseDiscard, false)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "PlayerHasPureSequence")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "PlayerDeadwoodValue")
		m.On("PlayerHasPureSequence", 0).Return(true).Maybe()
		m.On("PlayerDeadwoodValue", 0).Return(0).Maybe()
		out := p.Output(m, nil)
		assert.Contains(t, out, "デッドウッド: 0 点")
		assert.Contains(t, out, "純シーケンス成立")
		assert.NotContains(t, out, "純シーケンス未成立")
	})

	t.Run("round end", func(t *testing.T) {
		m, _ := setupMarriageCuiMock(domain.MarriagePhaseRoundEnd, false)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
		assert.Contains(t, out, "マール4点", "ラウンド終了時は CPU のマールを公開する")
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupMarriageCuiMock(domain.MarriagePhaseGameEnd, true)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
		assert.Contains(t, out, "マール4点", "ゲーム終了時は CPU のマールを公開する")
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupMarriageCuiMock(domain.MarriagePhaseDraw, false)
		out := p.Output(m, errors.New("err"))
		assert.NotEmpty(t, out)
	})
}

func TestMarriageCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MarriageCuiPresenter)
	m, _ := setupMarriageCuiMock(domain.MarriagePhaseDraw, false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}
