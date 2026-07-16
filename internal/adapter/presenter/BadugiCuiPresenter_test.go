package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBadugiCuiPresenter_Output_ContainsGameTitle(t *testing.T) {
	pres := new(presenter.BadugiCuiPresenter)
	bd, _ := makeBadugiForPresenter()
	out := pres.Output(bd, nil)
	assert.Contains(t, out, "Badugi")
	assert.Contains(t, out, "ディーラー")
	assert.Contains(t, out, "ポット")
}

func TestBadugiCuiPresenter_Output_ShowsDrawCounter(t *testing.T) {
	pres := new(presenter.BadugiCuiPresenter)
	bd, _ := makeBadugiForPresenter()
	bd.SetDrawIndex(2)
	out := pres.Output(bd, nil)
	assert.Contains(t, out, "ドロー: 2/3")
}

func TestBadugiCuiPresenter_Output_ShowsPreDrawLabel(t *testing.T) {
	pres := new(presenter.BadugiCuiPresenter)
	bd, _ := makeBadugiForPresenter()
	bd.SetDrawIndex(0)
	out := pres.Output(bd, nil)
	assert.Contains(t, out, "プリドロー")
}

func TestBadugiCuiPresenter_Output_ShowsHumanHand(t *testing.T) {
	pres := new(presenter.BadugiCuiPresenter)
	bd, players := makeBadugiForPresenter()
	bd.SetPhase(domain.BadugiPhaseDeal)
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 4, false))

	out := pres.Output(bd, nil)
	// 4 indices 0..3 appear in the human-hand line.
	for i := range 4 {
		assert.Contains(t, out, "["+string(rune('0'+i))+"]")
	}
}

func TestBadugiCuiPresenter_Output_ShowsBetLimits(t *testing.T) {
	pres := new(presenter.BadugiCuiPresenter)

	// Fixed limit (default) → min raise only, no cap line.
	bd, _ := makeBadugiForPresenter()
	out := pres.Output(bd, nil)
	assert.Contains(t, out, "最小レイズ:")
	assert.NotContains(t, out, "上限")

	// Pot-limit → a max bet is shown.
	bdPot, _ := makeBadugiForPresenter()
	cfg := bdPot.GetConfig()
	cfg.BettingLimit = domain.BettingLimitPotLimit
	bdPot.SetConfig(cfg)
	outPot := pres.Output(bdPot, nil)
	assert.Contains(t, outPot, "上限:")
	assert.NotContains(t, outPot, "上限: なし")

	// No-limit → unlimited note.
	bdNo, _ := makeBadugiForPresenter()
	cfg2 := bdNo.GetConfig()
	cfg2.BettingLimit = domain.BettingLimitNoLimit
	bdNo.SetConfig(cfg2)
	outNo := pres.Output(bdNo, nil)
	assert.Contains(t, outNo, "上限: なし")
}

func TestBadugiCuiPresenter_HintOutput_StandPat(t *testing.T) {
	pres := new(presenter.BadugiCuiPresenter)
	bd, players := makeBadugiForPresenter()
	bd.SetPhase(domain.BadugiPhaseDraw)
	bd.SetCurrentTurn(0)
	// A complete 4-card Badugi (all ranks and suits distinct).
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 4, false))

	out := pres.HintOutput(bd)
	assert.Contains(t, out, "スタンドパット")
}

func TestBadugiCuiPresenter_HintOutput_Exchange(t *testing.T) {
	pres := new(presenter.BadugiCuiPresenter)
	bd, players := makeBadugiForPresenter()
	bd.SetPhase(domain.BadugiPhaseDraw)
	bd.SetCurrentTurn(0)
	// Two spades + two hearts: best Badugi subset is only 2 cards, so the
	// other two should be flagged for exchange.
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))

	out := pres.HintOutput(bd)
	assert.Contains(t, out, "交換")
}

func TestBadugiCuiPresenter_HintOutput_NoneOutsideDraw(t *testing.T) {
	pres := new(presenter.BadugiCuiPresenter)
	bd, _ := makeBadugiForPresenter()
	bd.SetPhase(domain.BadugiPhaseDeal)

	assert.Contains(t, pres.HintOutput(bd), "ヒントはありません")
}

func TestBadugiCuiPresenter_Output_ShowsFoldedBadge(t *testing.T) {
	pres := new(presenter.BadugiCuiPresenter)
	bd, players := makeBadugiForPresenter()
	players[1].SetFolded(true)
	out := pres.Output(bd, nil)
	assert.Contains(t, out, "フォールド")
}

func TestBadugiCuiPresenter_Output_EndPhaseShowsHandName(t *testing.T) {
	pres := new(presenter.BadugiCuiPresenter)
	bd, players := makeBadugiForPresenter()
	bd.SetPhase(domain.BadugiPhaseEnd)
	bd.SetGameEndFlag(true)
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignClover, 4, false))
	_ = players[1].EvalHand()
	bd.SetRoundResults([]domain.BadugiResult{{PlayerIdx: 1, HandSize: 4, HandName: "Badugi", WonAmount: 100}})

	out := pres.Output(bd, nil)
	assert.Contains(t, out, "Badugi")
	assert.Contains(t, out, "ゲーム終了")
	assert.Contains(t, strings.ToLower(out), "100")
}

func TestBadugiCuiPresenter_Output_ShowsCpuAction(t *testing.T) {
	pres := new(presenter.BadugiCuiPresenter)
	bd, _ := makeBadugiForPresenter()
	bd.SetCpuActions([]domain.BadugiCpuAction{
		{PlayerIdx: 1, Action: domain.BadugiActionRaise, Amount: 20, DrawIndex: 0, RoundLabel: "pre-draw"},
	})
	out := pres.Output(bd, nil)
	assert.Contains(t, out, "Player 1")
	assert.Contains(t, out, "pre-draw")
}

func TestBadugiCuiPresenter_Output_ShowsCpuExchange(t *testing.T) {
	pres := new(presenter.BadugiCuiPresenter)
	bd, _ := makeBadugiForPresenter()
	bd.SetCpuExchanges([]domain.BadugiCpuExchange{
		{PlayerIdx: 2, DrawIndex: 1, ExchangeCount: 2},
	})
	out := pres.Output(bd, nil)
	assert.Contains(t, out, "Player 2")
	assert.Contains(t, out, "draw 1")
	assert.Contains(t, out, "2枚")
}

func TestBadugiCuiPresenter_Output_ErrorSurfaces(t *testing.T) {
	pres := new(presenter.BadugiCuiPresenter)
	bd, _ := makeBadugiForPresenter()
	out := pres.Output(bd, errors.New("oops"))
	assert.Contains(t, out, "oops")
}

func TestBadugiCuiPresenter_ActionLogOutput(t *testing.T) {
	pres := new(presenter.BadugiCuiPresenter)
	bd, _ := makeBadugiForPresenter()
	out := pres.ActionLogOutput(bd)
	assert.NotEmpty(t, out)
}
