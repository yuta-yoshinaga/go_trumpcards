//go:build test

package presenter_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupTarocchiniCuiMock() *interfaces.MockTarocchiniGame {
	m := new(interfaces.MockTarocchiniGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TarocchiniPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTeamScores").Return([2]int{0, 0})
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetHint").Return(nil).Maybe()
	m.On("GetLastTrickWinner").Return(-1).Maybe()
	m.On("GetScartoSize").Return(0).Maybe()
	return m
}

func setupTarocchiniCuiMockWithPlayers() (*interfaces.MockTarocchiniGame, []*domain.TarocchiniPlayer) {
	m := setupTarocchiniCuiMock()
	players := makeTarocchiniPlayers()
	m.On("GetPlayerCnt").Return(domain.TarocchiniPlayerCnt)
	for i := 0; i < domain.TarocchiniPlayerCnt; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestTarocchiniCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TarocchiniCuiPresenter)

	t.Run("play phase shows the hand and the teams", func(t *testing.T) {
		m, players := setupTarocchiniCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(1, 14, false))
		out := p.Output(m, nil)
		assert.Contains(t, out, "タロッキーニ")
		assert.Contains(t, out, "[0]")
		assert.Contains(t, out, "チーム")
		assert.Contains(t, out, "[親]", "the dealer must be marked")
	})

	// 「後出しが勝つ」を知らないとパパ 4 枚の使い方が判断できないので毎回出す。
	t.Run("the papi rule is stated on every frame", func(t *testing.T) {
		m, _ := setupTarocchiniCuiMockWithPlayers()
		assert.Contains(t, p.Output(m, nil), "後から出した方が勝ち")
	})

	// cuiCardStr は design 5/6 を UNKNOWN に落とすので、切札とマットが全部同じ
	// 文字列になってしまう。専用の描画が要る。
	t.Run("trumps and the Matto are named, not UNKNOWN", func(t *testing.T) {
		m, players := setupTarocchiniCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.TarocchiniTrumpDesign, 19, false))
		players[0].AddCard(domain.NewCard(domain.TarocchiniMattoDesign, domain.TarocchiniMattoValue, false))
		players[0].AddCard(domain.NewCard(domain.TarocchiniTrumpDesign, 2, false))
		out := p.Output(m, nil)
		assert.NotContains(t, out, "UNKNOWN")
		assert.Contains(t, out, "切札 19")
		assert.Contains(t, out, "マット")
		// パパは番号を出さない —— 番号だと 2 が 3 より弱いと読まれる。
		assert.Contains(t, out, "パパ")
	})

	t.Run("scarto prompt names the count", func(t *testing.T) {
		m, _ := setupTarocchiniCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TarocchiniPhaseScarto)
		out := p.Output(m, nil)
		assert.Contains(t, out, "2 枚を捨てて")
		assert.Contains(t, out, "scarto")
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupTarocchiniCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TarocchiniPhaseTrickEnd)
		assert.Contains(t, p.Output(m, nil), "トリック終了")
	})

	t.Run("round end lists every player's trick count", func(t *testing.T) {
		m, players := setupTarocchiniCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TarocchiniPhaseRoundEnd)
		players[0].AddTrick([]*domain.Card{domain.NewCard(1, 7, false)})
		assert.Contains(t, p.Output(m, nil), "各プレイヤーのトリック数")
	})

	t.Run("trick cards are rendered", func(t *testing.T) {
		m, _ := setupTarocchiniCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.TarocchiniTrumpDesign, 12, false)},
		})
		assert.Contains(t, p.Output(m, nil), "切札 12")
	})

	t.Run("error block is rendered", func(t *testing.T) {
		m, _ := setupTarocchiniCuiMockWithPlayers()
		assert.Contains(t, p.Output(m, errors.New("boom")), "boom")
	})

	t.Run("game end names the winning team", func(t *testing.T) {
		m, _ := setupTarocchiniCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(1)
		assert.Contains(t, p.Output(m, nil), "チーム1の勝ち")
	})

	// 同点は winnerTeam = -1。チーム -1 を勝者として書いてはならない。
	t.Run("a draw is announced as a draw", func(t *testing.T) {
		m, _ := setupTarocchiniCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(-1)
		out := p.Output(m, nil)
		assert.Contains(t, out, "引き分け")
		assert.NotContains(t, out, "-1")
	})

	t.Run("nil player is skipped", func(t *testing.T) {
		m := setupTarocchiniCuiMock()
		m.On("GetPlayerCnt").Return(1)
		m.On("GetPlayer", 0).Return((*domain.TarocchiniPlayer)(nil))
		assert.NotPanics(t, func() { p.Output(m, nil) })
	})
}

func TestTarocchiniCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TarocchiniCuiPresenter)

	// ドメインが返しうる 6 種すべてが訳に解決すること。#4660 では 3 種が生キーで
	// 画面に出る状態のまま通っていた。
	t.Run("every emitted reason resolves to prose", func(t *testing.T) {
		for _, reason := range []string{
			"lead_low", "lead_trump", "play_papa", "play_matto", "follow_trump", "follow_low",
		} {
			m, players := setupTarocchiniCuiMockWithPlayers()
			players[0].AddCard(domain.NewCard(1, 9, false))
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
			m.On("GetHint").Return(&domain.TarocchiniHint{CardIndices: []int{0}, Reason: reason})
			out := p.HintOutput(m)
			assert.NotContains(t, out, reason, "reason %q printed as a raw key", reason)
			assert.Contains(t, out, "[0]")
		}
	})

	t.Run("hint with no card indices still renders the reason", func(t *testing.T) {
		m, _ := setupTarocchiniCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.TarocchiniHint{Reason: "lead_low"})
		assert.Contains(t, p.HintOutput(m), "-")
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupTarocchiniCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.TarocchiniHint)(nil))
		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})
}

func TestTarocchiniCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TarocchiniCuiPresenter)
	m := new(interfaces.MockTarocchiniGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You play Papa"},
	})
	assert.Contains(t, p.ActionLogOutput(m), "play")
}

// #5716: ラウンド得点は「トリック数 + 最終トリック +2 + スカルト加点」の 3 要素で
// 決まるのに、画面はトリック数しか出しておらず、teamScores の増分と突き合わせて
// 検算できなかった。
func TestTarocchiniCuiPresenter_ShowsTheRoundBreakdown(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TarocchiniCuiPresenter)

	atRoundEnd := func(lastTrickWinner, scarto, dealer int) *interfaces.MockTarocchiniGame {
		m, players := setupTarocchiniCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TarocchiniPhaseRoundEnd)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDealerIdx")
		m.On("GetDealerIdx").Return(dealer)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastTrickWinner")
		m.On("GetLastTrickWinner").Return(lastTrickWinner)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetScartoSize")
		m.On("GetScartoSize").Return(scarto)
		for i := range players {
			players[i].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
		}
		return m
	}

	// 席 1 が最終トリック → チーム1 に +2。ディーラーは席 0 (チーム0) でスカルト 2 枚。
	t.Run("names who took the last trick and where the scarto went", func(t *testing.T) {
		out := p.Output(atRoundEnd(1, 2, 0), nil)

		assert.Contains(t, out, i18n.Tf("tarocchini.roundEndLastTrick",
			"team", i18n.Tf("tarocchini.teamName", "n", "1"),
			"bonus", strconv.Itoa(domain.TarocchiniLastTrickBonus)))
		assert.Contains(t, out, i18n.Tf("tarocchini.roundEndScarto",
			"team", i18n.Tf("tarocchini.teamName", "n", "0"),
			"count", "2"))
	})

	// 最終トリックが未確定 (-1) / スカルトなしのラウンドでは、その行を出さない。
	t.Run("omits the lines that do not apply", func(t *testing.T) {
		out := p.Output(atRoundEnd(-1, 0, 0), nil)

		assert.NotContains(t, out, strings.Split(i18n.T("tarocchini.roundEndLastTrick"), "{{")[0])
		assert.NotContains(t, out, strings.Split(i18n.T("tarocchini.roundEndScarto"), "{{")[0])
	})
}
