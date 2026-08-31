//go:build test

package presenter_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupKalookiCuiMock(phase domain.KalookiPhase, gameEnd bool) (*interfaces.MockKalookiGame, []*domain.KalookiPlayer) {
	m := new(interfaces.MockKalookiGame)
	players := []*domain.KalookiPlayer{
		domain.NewKalookiPlayer(true),
		domain.NewKalookiPlayer(false),
		domain.NewKalookiPlayer(false),
		domain.NewKalookiPlayer(false),
	}
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	m.On("GetOpeningThreshold").Return(51)
	m.On("GetDrawPileCount").Return(53)
	m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
	m.On("GetGameEndFlag").Return(gameEnd)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	winner := -1
	if gameEnd {
		winner = 0
	}
	m.On("GetWinnerIdx").Return(winner)
	m.On("GetConfig").Return(domain.DefaultKalookiConfig())
	m.On("GetRoundWinnerIdx").Return(-1)
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestKalookiCuiPresenter_Output(t *testing.T) {
	p := new(presenter.KalookiCuiPresenter)

	t.Run("draw phase", func(t *testing.T) {
		m, _ := setupKalookiCuiMock(domain.KalookiPhaseDraw, false)
		assert.NotEmpty(t, p.Output(m, nil))
	})

	t.Run("meld phase", func(t *testing.T) {
		m, _ := setupKalookiCuiMock(domain.KalookiPhaseMeld, false)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
		// Current player has not opened → opening-threshold hint is shown.
		assert.Contains(t, out, "開設には")
		assert.Contains(t, out, "51")
	})

	t.Run("round end", func(t *testing.T) {
		m, _ := setupKalookiCuiMock(domain.KalookiPhaseRoundEnd, false)
		assert.NotEmpty(t, p.Output(m, nil))
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupKalookiCuiMock(domain.KalookiPhaseGameEnd, true)
		assert.NotEmpty(t, p.Output(m, nil))
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupKalookiCuiMock(domain.KalookiPhaseDraw, false)
		assert.NotEmpty(t, p.Output(m, errors.New("err")))
	})

	t.Run("opened player with melds", func(t *testing.T) {
		m, players := setupKalookiCuiMock(domain.KalookiPhaseMeld, false)
		players[0].SetHasOpened(true)
		players[0].AppendMeld([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 5, false),
		})
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
		// Already opened → no opening-threshold hint.
		assert.NotContains(t, out, "開設には")
	})
}

func TestKalookiCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.KalookiCuiPresenter)
	m, _ := setupKalookiCuiMock(domain.KalookiPhaseDraw, false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}

// **`lo <pIdx> <mIdx> <cIdx>` が要求する番号が場のどこにも出ていなかった** ──
// 札が並ぶだけなので、狙うメルドの添字は数えて当てるしかない (#6462)。
func TestKalookiCuiPresenter_NumbersEachMeldForLayoff(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	p := new(presenter.KalookiCuiPresenter)

	// 黒スートだけで組む ── 赤スートは色コードで包まれ、期待値が装飾に依存する。
	// 中身で見分けが付く 2 つのメルドを使い、番号だけでなく **その番号のすぐ後ろに
	// そのメルドの札が並ぶこと**を見る。`mi + 1` にしたり順序を入れ替えたりする
	// 変異はこれで落ちる。
	melds := [][]*domain.Card{
		{
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignSpade, 4, false),
			domain.NewCard(domain.CardDesignSpade, 5, false),
		},
		{
			domain.NewCard(domain.CardDesignClover, 9, false),
			domain.NewCard(domain.CardDesignClover, 10, false),
			domain.NewCard(domain.CardDesignClover, 11, false),
		},
	}
	want := []string{
		"[0] SPADE 3 SPADE 4 SPADE 5",
		"[1] CLOVER 9 CLOVER 10 CLOVER 11",
	}

	withMelds := func(seat int) string {
		m, players := setupKalookiCuiMock(domain.KalookiPhaseMeld, false)
		players[seat].SetHasOpened(true)
		for _, meld := range melds {
			players[seat].AppendMeld(meld)
		}
		return p.Output(m, nil)
	}

	t.Run("numbers the human melds from zero", func(t *testing.T) {
		out := withMelds(0)
		for mi, line := range want {
			assert.Contains(t, out, line,
				"meld %d must be labelled with the index `layoff` takes", mi)
		}
		// 0 始まりであること。`mi + 1` にすると最後の番号がここに現れる。
		assert.NotContains(t, out, "["+strconv.Itoa(len(melds))+"] ")
	})

	// layoff の相手は他家なので、番号付けが人間の場だけだと使えない。
	t.Run("numbers another player's melds the same way", func(t *testing.T) {
		out := withMelds(2)
		for _, line := range want {
			assert.Contains(t, out, line)
		}
	})
}
