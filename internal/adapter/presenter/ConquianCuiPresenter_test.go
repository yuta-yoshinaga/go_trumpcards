//go:build test

package presenter_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func makeConquianPlayers() []*domain.ConquianPlayer {
	return []*domain.ConquianPlayer{
		domain.NewConquianPlayer(true),
		domain.NewConquianPlayer(false),
	}
}

func setupConquianCuiMock(phase domain.ConquianPhase, ended bool, winner int) (*interfaces.MockConquianGame, []*domain.ConquianPlayer) {
	m := new(interfaces.MockConquianGame)
	players := makeConquianPlayers()
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(20)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(ended)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(winner)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetTookDiscard").Return(false)
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestConquianCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ConquianCuiPresenter)

	t.Run("draw phase shows header and hand", func(t *testing.T) {
		m, players := setupConquianCuiMock(domain.ConquianPhaseDraw, false, -1)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Conquian (コンキャン)")
		assert.Contains(t, result, "ラウンド: 1")
	})

	t.Run("meld phase shows meld and melded cards", func(t *testing.T) {
		m, players := setupConquianCuiMock(domain.ConquianPhaseMeld, false, -1)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[0].AddMeld([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 5, false),
		})
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		// No discard taken → no forced-use warning.
		assert.NotContains(t, result, "必ず")
	})

	t.Run("forced-use warning shown after taking a discard", func(t *testing.T) {
		m, _ := setupConquianCuiMock(domain.ConquianPhaseMeld, false, -1)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTookDiscard")
		m.On("GetTookDiscard").Return(true)

		result := p.Output(m, nil)
		assert.Contains(t, result, "必ず")
	})

	t.Run("discard top is displayed", func(t *testing.T) {
		m, _ := setupConquianCuiMock(domain.ConquianPhaseDraw, false, -1)
		m.ExpectedCalls = nil
		players := makeConquianPlayers()
		m.On("GetRoundNumber").Return(1)
		m.On("GetDrawPileCount").Return(20)
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignClover, 7, false))
		m.On("GetGameEndFlag").Return(false)
		m.On("GetPhase").Return(domain.ConquianPhaseDraw)
		m.On("GetCurrentPlayerIdx").Return(0)
		m.On("GetWinnerIdx").Return(-1)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		m.On("GetPlayerCnt").Return(2)
		m.On("GetPlayer", 0).Return(players[0])
		m.On("GetPlayer", 1).Return(players[1])
		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札")
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupConquianCuiMock(domain.ConquianPhaseRoundEnd, false, -1)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end with winner", func(t *testing.T) {
		m, _ := setupConquianCuiMock(domain.ConquianPhaseGameEnd, true, 0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("game end draw", func(t *testing.T) {
		m, _ := setupConquianCuiMock(domain.ConquianPhaseGameEnd, true, -1)
		result := p.Output(m, nil)
		assert.Contains(t, result, "引き分け")
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupConquianCuiMock(domain.ConquianPhaseDraw, false, -1)
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestConquianCuiPresenter_ActionLogOutput(t *testing.T) {
	m, _ := setupConquianCuiMock(domain.ConquianPhaseDraw, false, -1)
	p := new(presenter.ConquianCuiPresenter)
	assert.NotPanics(t, func() { p.ActionLogOutput(m) })
}

// #5664: 勝利は「メルドで 11 枚並べきる」こと。Web は進捗バーで合計と残りを出し、
// 残り2枚以下を警告色にしているのに、CUI はメルドの中身を並べるだけで、**毎回
// 目で数えないと勝利がどれだけ近いか分からなかった。**
func TestConquianCuiPresenter_ShowsMeldProgress(t *testing.T) {
	p := new(presenter.ConquianCuiPresenter)
	card := func(v int) *domain.Card { return domain.NewCard(domain.CardDesignSpade, v, false) }

	// n 枚ぶんのメルドを作る (3枚組 + 端数)。
	meldsOf := func(n int) [][]*domain.Card {
		var melds [][]*domain.Card
		for n > 0 {
			size := 3
			if n < 3 {
				size = n
			}
			meld := make([]*domain.Card, size)
			for i := range meld {
				meld[i] = card(i + 1)
			}
			melds = append(melds, meld)
			n -= size
		}
		return melds
	}

	t.Run("counts the melded cards toward the target", func(t *testing.T) {
		m, players := setupConquianCuiMock(domain.ConquianPhaseMeld, false, -1)
		players[0].SetMelds(meldsOf(6))

		out := p.Output(m, nil)

		assert.Contains(t, out, i18n.Tf("conquian.meldProgress",
			"count", "6", "total", strconv.Itoa(domain.ConquianMeldTarget)))
	})

	// **残り2枚以下は Web と同じ基準で強調する。**
	t.Run("warns in the final stretch", func(t *testing.T) {
		m, players := setupConquianCuiMock(domain.ConquianPhaseMeld, false, -1)
		players[0].SetMelds(meldsOf(domain.ConquianMeldTarget - 2))

		out := p.Output(m, nil)

		assert.Contains(t, out, i18n.Tf("conquian.meldRemaining", "count", "2"))
	})

	t.Run("stays quiet while more than two are missing", func(t *testing.T) {
		m, players := setupConquianCuiMock(domain.ConquianPhaseMeld, false, -1)
		players[0].SetMelds(meldsOf(domain.ConquianMeldTarget - 3))

		out := p.Output(m, nil)

		// 警告の閾値は残り2枚以下。3枚残っているうちは出さない。
		prefix, _, ok := strings.Cut(i18n.Tf("conquian.meldRemaining", "count", "\x00"), "\x00")
		require.True(t, ok)
		require.NotEmpty(t, strings.TrimSpace(prefix))
		assert.NotContains(t, out, prefix)
		// 進捗そのものは出る。
		assert.Contains(t, out, i18n.Tf("conquian.meldProgress",
			"count", strconv.Itoa(domain.ConquianMeldTarget-3), "total", strconv.Itoa(domain.ConquianMeldTarget)))
	})

	// 既存のメルド内容の表示は残す。
	t.Run("keeps listing what is in each meld", func(t *testing.T) {
		m, players := setupConquianCuiMock(domain.ConquianPhaseMeld, false, -1)
		players[0].SetMelds(meldsOf(3))

		out := p.Output(m, nil)

		assert.Contains(t, out, "SPADE 1")
	})
}
