//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupFiftyOneMock() *interfaces.MockFiftyOneGame {
	m := new(interfaces.MockFiftyOneGame)
	m.On("GetPlayerCnt").Return(4)
	m.On("GetCurrentTurn").Return(0)
	m.On("GetPhase").Return(domain.FiftyOnePhasePlay)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetTurnNumber").Return(1)
	m.On("GetStopCallerIdx").Return(-1)
	m.On("GetLastAction").Return("")
	m.On("GetLastHandIdx").Return(-1)
	m.On("GetLastTableIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultFiftyOneConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("IsHumanTurn").Return(true)

	tableCards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
		domain.NewCard(domain.CardDesignDiamond, 12, false),
		domain.NewCard(domain.CardDesignClover, 6, false),
		domain.NewCard(domain.CardDesignSpade, 8, false),
	}
	m.On("GetTableCards").Return(tableCards)

	for i := range 4 {
		p := domain.NewFiftyOnePlayer(i == 0)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		p.AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
		p.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		m.On("GetPlayer", i).Return(p)
	}
	return m
}

// **4 スート分の合計を出す。**Web はバッジで 4 つとも常時出しているのに、CUI は
// 最良スートの 1 数値しか出さず、残り 3 スートは手札一覧から暗算する必要があった (#4866)。
func TestFiftyOneCuiPresenter_ShowsEverySuitTotal(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	m := setupFiftyOneMock()
	p := new(presenter.FiftyOneCuiPresenter)
	result := p.Output(m, nil)

	// 手札は SPADE A(11)/10, HEART 5, DIAMOND 3, CLOVER 2 -> SPADE 21 が最良で、
	// 既存の (スコア: 21) と一致する。満点 51 と合わせた「合計/51」で表示される。
	assert.Contains(t, result, "スート別: SPADE 21/51*  CLOVER 2/51  HEART 5/51  DIAMOND 3/51")
	assert.Contains(t, result, "SPADE 21/51*")
	assert.Contains(t, result, "CLOVER 2/51")
	// CPU の行には出さない (非公開ルールは変えない)。
	assert.Equal(t, 1, strings.Count(result, "スート別:"))

	// 異なる手札構成でもスート別行に /51 が正しく出ること (DIAMOND 20/51, HEART 8/51)
	m2 := setupFiftyOneMock()
	p2 := domain.NewFiftyOnePlayer(true)
	p2.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
	p2.AddCard(domain.NewCard(domain.CardDesignDiamond, 11, false))
	p2.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
	m2.ExpectedCalls = nil
	m2.On("GetPlayerCnt").Return(4)
	m2.On("GetCurrentTurn").Return(0)
	m2.On("GetPhase").Return(domain.FiftyOnePhasePlay)
	m2.On("GetGameEndFlag").Return(false)
	m2.On("GetWinnerIdx").Return(-1)
	m2.On("GetTurnNumber").Return(1)
	m2.On("GetStopCallerIdx").Return(-1)
	m2.On("GetLastAction").Return("")
	m2.On("GetLastHandIdx").Return(-1)
	m2.On("GetLastTableIdx").Return(-1)
	m2.On("GetConfig").Return(domain.DefaultFiftyOneConfig())
	m2.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m2.On("IsHumanTurn").Return(true)
	m2.On("GetTableCards").Return([]*domain.Card{})
	m2.On("GetPlayer", 0).Return(p2)
	for i := 1; i < 4; i++ {
		m2.On("GetPlayer", i).Return(domain.NewFiftyOnePlayer(false))
	}
	result2 := p.Output(m2, nil)
	assert.Contains(t, result2, "スート別: SPADE 0/51  CLOVER 0/51  HEART 8/51  DIAMOND 20/51*")
	assert.Contains(t, result2, "DIAMOND 20/51*")
	assert.Contains(t, result2, "HEART 8/51")
}

func TestFiftyOneCuiPresenter_Output_Initial(t *testing.T) {
	p := new(presenter.FiftyOneCuiPresenter)
	m := setupFiftyOneMock()

	result := p.Output(m, nil)
	assert.Contains(t, result, "Fifty-one")
	assert.Contains(t, result, "場札:")
	assert.Contains(t, result, "あなたのターン")
}

func TestFiftyOneCuiPresenter_Output_GameEnd(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	p := new(presenter.FiftyOneCuiPresenter)
	m := new(interfaces.MockFiftyOneGame)
	m.On("GetPlayerCnt").Return(4)
	m.On("GetCurrentTurn").Return(0)
	m.On("GetPhase").Return(domain.FiftyOnePhaseGameEnd)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerIdx").Return(0)
	m.On("GetStopCallerIdx").Return(-1)
	m.On("GetLastAction").Return("")
	m.On("GetConfig").Return(domain.DefaultFiftyOneConfig())
	m.On("IsHumanTurn").Return(true)
	m.On("GetTableCards").Return([]*domain.Card{})

	humanPlayer := domain.NewFiftyOnePlayer(true)
	humanPlayer.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	m.On("GetPlayer", 0).Return(humanPlayer)
	cpu1 := domain.NewFiftyOnePlayer(false)
	cpu1.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	m.On("GetPlayer", 1).Return(cpu1)
	for i := 2; i < 4; i++ {
		m.On("GetPlayer", i).Return(domain.NewFiftyOnePlayer(false))
	}

	result := p.Output(m, nil)
	assert.Contains(t, result, "あなたの勝ち")
	assert.Contains(t, result, "ゲーム終了")
	// スコア一覧に /51 が出ること (異なる合計値 2 つ: あなた 11/51点, CPU 1 7/51点)
	assert.Contains(t, result, "あなた: 11/51点")
	assert.Contains(t, result, "CPU 1: 7/51点")
}

func TestFiftyOneCuiPresenter_Output_Error(t *testing.T) {
	p := new(presenter.FiftyOneCuiPresenter)
	m := setupFiftyOneMock()

	result := p.Output(m, errors.New("invalid index"))
	assert.Contains(t, result, "invalid index")
}

func TestFiftyOneCuiPresenter_Output_StopCalled(t *testing.T) {
	p := new(presenter.FiftyOneCuiPresenter)
	m := setupFiftyOneMock()
	// Override stop caller
	m.ExpectedCalls = nil
	m.On("GetPlayerCnt").Return(4)
	m.On("GetCurrentTurn").Return(0)
	m.On("GetPhase").Return(domain.FiftyOnePhasePlay)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetStopCallerIdx").Return(2)
	m.On("GetLastAction").Return("stop")
	m.On("GetConfig").Return(domain.DefaultFiftyOneConfig())
	m.On("IsHumanTurn").Return(true)
	m.On("GetTableCards").Return([]*domain.Card{})

	for i := range 4 {
		m.On("GetPlayer", i).Return(domain.NewFiftyOnePlayer(i == 0))
	}

	result := p.Output(m, nil)
	assert.Contains(t, result, "ストップ宣言")
	// The prominent colored alert is shown before the prompt in every phase.
	assert.Contains(t, result, "最終ラウンド")
}

func TestFiftyOneCuiPresenter_Output_NoStopAlertWhenNoStop(t *testing.T) {
	p := new(presenter.FiftyOneCuiPresenter)
	m := setupFiftyOneMock()
	// setupFiftyOneMock leaves GetStopCallerIdx() == -1 (no stop declared).
	result := p.Output(m, nil)
	assert.NotContains(t, result, "最終ラウンド")
}

func TestFiftyOneCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.FiftyOneCuiPresenter)
	m := new(interfaces.MockFiftyOneGame)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})

	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}
