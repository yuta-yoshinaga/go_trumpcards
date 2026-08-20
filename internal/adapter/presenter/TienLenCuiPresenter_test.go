//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupTienLenCuiMock() (*interfaces.MockTienLenGame, []*domain.TienLenPlayer) {
	m := new(interfaces.MockTienLenGame)
	players := makeTienLenPlayers()
	m.On("GetGameEndFlag").Return(false)
	m.On("GetTableCards").Return(([]*domain.Card)(nil))
	m.On("GetTablePlayType").Return(domain.TienLenPlayInvalid)
	m.On("GetCpuActions").Return(([]*domain.TienLenAction)(nil))
	m.On("IsHumanTurn").Return(true)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestTienLenCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.TienLenCuiPresenter)

	t.Run("initial empty table", func(t *testing.T) {
		m, players := setupTienLenCuiMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Tien Len")
		assert.Contains(t, result, "[0]")
		assert.Contains(t, result, "自由に出せます")
		assert.Contains(t, result, "あなたのターン")
		// Combo-strength rules are shown on the human's turn.
		assert.Contains(t, result, "連番>3カード>ペア>シングル")
	})

	t.Run("invalid combo error is accompanied by the combo rules", func(t *testing.T) {
		m, _ := setupTienLenCuiMock()
		result := p.Output(m, errors.New("invalid combo"))
		assert.Contains(t, result, "invalid combo")
		assert.Contains(t, result, "連番>3カード>ペア>シングル")
	})

	t.Run("table cards and cpu pass action", func(t *testing.T) {
		m, _ := setupTienLenCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTableCards")
		m.On("GetTableCards").Return([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTablePlayType")
		m.On("GetTablePlayType").Return(domain.TienLenPlaySingle)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCpuActions")
		m.On("GetCpuActions").Return([]*domain.TienLenAction{
			{PlayerIdx: 1, PlayedCards: nil},
			{PlayerIdx: 2, PlayedCards: []*domain.Card{domain.NewCard(domain.CardDesignHeart, 4, false)}},
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "場")
		assert.Contains(t, result, "CPU 1")
		// The table combo type is named alongside the cards.
		assert.Contains(t, result, i18n.Tf("tienlen.tableComboType", "type", i18n.T("tienlen.comboSingle")))
	})

	t.Run("finished player shown", func(t *testing.T) {
		m, players := setupTienLenCuiMock()
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		assert.Contains(t, p.Output(m, nil), "上がり")
	})

	t.Run("error shown", func(t *testing.T) {
		m, _ := setupTienLenCuiMock()
		assert.Contains(t, p.Output(m, errors.New("invalid play")), "invalid play")
	})

	t.Run("game ended rankings", func(t *testing.T) {
		m, players := setupTienLenCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		players[0].SetRank(1)
		players[1].SetRank(2)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
		assert.Contains(t, result, "あなた")
	})
}

func TestTienLenCuiPresenter_Output_AllComboTypeLabels(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.TienLenCuiPresenter)

	// Every table combo type renders its own localized label.
	for _, tc := range []struct {
		playType domain.TienLenPlayType
		labelKey string
	}{
		{domain.TienLenPlaySingle, "tienlen.comboSingle"},
		{domain.TienLenPlayPair, "tienlen.comboPair"},
		{domain.TienLenPlayTriple, "tienlen.comboTriple"},
		{domain.TienLenPlayStraight, "tienlen.comboStraight"},
		{domain.TienLenPlayThreePairRun, "tienlen.comboThreePairRun"},
		{domain.TienLenPlayFourOfAKind, "tienlen.comboFourOfAKind"},
		{domain.TienLenPlayInvalid, "tienlen.comboInvalid"},
	} {
		m, _ := setupTienLenCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTableCards")
		m.On("GetTableCards").Return([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTablePlayType")
		m.On("GetTablePlayType").Return(tc.playType)
		result := p.Output(m, nil)
		assert.Contains(t, result, i18n.Tf("tienlen.tableComboType", "type", i18n.T(tc.labelKey)))
	}
}

func TestTienLenCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.TienLenCuiPresenter)

	m := new(interfaces.MockTienLenGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played 1 card(s)"},
	})
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewTienLenPlayer(true)).Maybe()
	assert.Contains(t, p.ActionLogOutput(m), "play")
}

// **CPU 名も i18n と太字を通す。**順位表示と CPU 手番ログだけ `fmt.Sprintf("CPU %d")`
// で自前に組んでいて、同じ画面の他の名前と見た目も経路も違っていた (#4807)。
func TestTienLen_CpuNamesGoThroughTheSharedHelper(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(false)
	defer color.SetNoColor(origNoColor)

	bold1 := color.Bold(i18n.Tf("cuiPlayerCpu", "idx", "1"))
	p := new(presenter.TienLenCuiPresenter)

	t.Run("cpu action log", func(t *testing.T) {
		m, _ := setupTienLenCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCpuActions")
		m.On("GetCpuActions").Return([]*domain.TienLenAction{{PlayerIdx: 1, PlayedCards: nil}})
		out := p.Output(m, nil)
		assert.Contains(t, out, bold1)
		// 素の "CPU 1" が太字の外に残っていないこと。
		assert.NotContains(t, strings.ReplaceAll(out, bold1, ""), "CPU 1")
	})
}

// #5624: CUI にヒントが無く、CPU の判断材料 (findBestPlay) を人間向けに
// 取り出す経路そのものが存在しなかった。
func TestTienLenCuiPresenterHintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TienLenCuiPresenter)

	t.Run("names the cards and their indices", func(t *testing.T) {
		m := new(interfaces.MockTienLenGame)
		pl := domain.NewTienLenPlayer(true)
		pl.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		pl.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		m.On("GetHint").Return(&domain.TienLenHint{Indices: []int{1}})
		m.On("GetCurrentTurn").Return(0)
		m.On("GetPlayer", 0).Return(pl)

		out := p.HintOutput(m)
		// 番号を出す ── `p <idx>` にそのまま打ち込めること。
		assert.Contains(t, out, "[1]")
		assert.NotContains(t, out, "[0]")
	})

	t.Run("says to pass when nothing beats the table", func(t *testing.T) {
		m := new(interfaces.MockTienLenGame)
		m.On("GetHint").Return(&domain.TienLenHint{Pass: true})

		assert.Contains(t, p.HintOutput(m), i18n.T("tienlen.hintPass"))
	})

	t.Run("says nothing useful when it is not the human's turn", func(t *testing.T) {
		m := new(interfaces.MockTienLenGame)
		m.On("GetHint").Return((*domain.TienLenHint)(nil))

		assert.Contains(t, p.HintOutput(m), i18n.T("tienlen.hintNone"))
	})
}
