package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDeuceToSevenCuiPresenter_Output_ContainsGameTitle(t *testing.T) {
	pres := new(presenter.DeuceToSevenCuiPresenter)
	dt, _ := makeDeuceToSevenForPresenter()
	out := pres.Output(dt, nil)
	assert.Contains(t, out, "2-7 Triple Draw")
	assert.Contains(t, out, "ディーラー")
	assert.Contains(t, out, "ポット")
}

func TestDeuceToSevenCuiPresenter_Output_ShowsDrawCounter(t *testing.T) {
	pres := new(presenter.DeuceToSevenCuiPresenter)
	dt, _ := makeDeuceToSevenForPresenter()
	dt.SetDrawIndex(2)
	out := pres.Output(dt, nil)
	assert.Contains(t, out, "ドロー: 2/3")
}

func TestDeuceToSevenCuiPresenter_Output_ShowsPreDrawLabel(t *testing.T) {
	pres := new(presenter.DeuceToSevenCuiPresenter)
	dt, _ := makeDeuceToSevenForPresenter()
	dt.SetDrawIndex(0)
	out := pres.Output(dt, nil)
	assert.Contains(t, out, "プリドロー")
}

func TestDeuceToSevenCuiPresenter_Output_ShowsHumanHand(t *testing.T) {
	pres := new(presenter.DeuceToSevenCuiPresenter)
	dt, players := makeDeuceToSevenForPresenter()
	dt.SetPhase(domain.DeuceToSevenPhaseDeal)
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 4, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))

	out := pres.Output(dt, nil)
	// 5 indices 0..4 appear in the human-hand line.
	for i := range 5 {
		assert.Contains(t, out, "["+string(rune('0'+i))+"]")
	}
}

func TestDeuceToSevenCuiPresenter_Output_ShowsFoldedBadge(t *testing.T) {
	pres := new(presenter.DeuceToSevenCuiPresenter)
	dt, players := makeDeuceToSevenForPresenter()
	players[1].SetFolded(true)
	out := pres.Output(dt, nil)
	assert.Contains(t, out, "フォールド")
}

func TestDeuceToSevenCuiPresenter_Output_EndPhaseShowsHandName(t *testing.T) {
	pres := new(presenter.DeuceToSevenCuiPresenter)
	dt, players := makeDeuceToSevenForPresenter()
	dt.SetPhase(domain.DeuceToSevenPhaseEnd)
	dt.SetGameEndFlag(true)
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 4, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	_ = players[1].EvalHand()
	dt.SetRoundResults([]domain.DeuceToSevenResult{{PlayerIdx: 1, HandRank: domain.PokerHandHighCard, HandName: "High Card", WonAmount: 100}})

	out := pres.Output(dt, nil)
	assert.Contains(t, out, "High Card")
	assert.Contains(t, out, "ゲーム終了")
	assert.Contains(t, strings.ToLower(out), "100")
}

func TestDeuceToSevenCuiPresenter_Output_ShowsCpuAction(t *testing.T) {
	pres := new(presenter.DeuceToSevenCuiPresenter)
	dt, _ := makeDeuceToSevenForPresenter()
	dt.SetCpuActions([]domain.DeuceToSevenCpuAction{
		{PlayerIdx: 1, Action: domain.DeuceToSevenActionRaise, Amount: 20, DrawIndex: 0, RoundLabel: "pre-draw"},
	})
	out := pres.Output(dt, nil)
	assert.Contains(t, out, "Player 1")
	assert.Contains(t, out, "pre-draw")
}

func TestDeuceToSevenCuiPresenter_Output_ShowsCpuExchange(t *testing.T) {
	pres := new(presenter.DeuceToSevenCuiPresenter)
	dt, _ := makeDeuceToSevenForPresenter()
	dt.SetCpuExchanges([]domain.DeuceToSevenCpuExchange{
		{PlayerIdx: 2, DrawIndex: 1, ExchangeCount: 2},
	})
	out := pres.Output(dt, nil)
	assert.Contains(t, out, "Player 2")
	assert.Contains(t, out, "draw 1")
	assert.Contains(t, out, "2枚")
}

func TestDeuceToSevenCuiPresenter_Output_ErrorSurfaces(t *testing.T) {
	pres := new(presenter.DeuceToSevenCuiPresenter)
	dt, _ := makeDeuceToSevenForPresenter()
	out := pres.Output(dt, errors.New("oops"))
	assert.Contains(t, out, "oops")
}

func TestDeuceToSevenCuiPresenter_ActionLogOutput(t *testing.T) {
	pres := new(presenter.DeuceToSevenCuiPresenter)
	dt, _ := makeDeuceToSevenForPresenter()
	out := pres.ActionLogOutput(dt)
	assert.NotEmpty(t, out)
}
