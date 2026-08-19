//go:build test

package presenter_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupMinchiateCuiMock() *interfaces.MockMinchiateGame {
	m := new(interfaces.MockMinchiateGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.MinchiatePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTeamScores").Return([2]int{0, 0})
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetHint").Return(nil).Maybe()
	return m
}

func setupMinchiateCuiMockWithPlayers() (*interfaces.MockMinchiateGame, []*domain.MinchiatePlayer) {
	m := setupMinchiateCuiMock()
	players := makeMinchiatePlayers()
	m.On("GetPlayerCnt").Return(domain.MinchiatePlayerCnt)
	for i := 0; i < domain.MinchiatePlayerCnt; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestMinchiateCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.MinchiateCuiPresenter)

	t.Run("play phase shows the hand and the teams", func(t *testing.T) {
		m, players := setupMinchiateCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(1, 14, false))
		out := p.Output(m, nil)
		assert.Contains(t, out, "ミンキアーテ")
		assert.Contains(t, out, "[0]")
		assert.Contains(t, out, "チーム")
		assert.Contains(t, out, "[親]", "the dealer must be marked")
	})

	// 切札が 40 枚あることは 21 枚のタロー系との最大の差で、「まだ上に何枚残って
	// いるか」の見積もりに直結する。毎フレーム出す。
	t.Run("the 40-trump count is stated on every frame", func(t *testing.T) {
		m, _ := setupMinchiateCuiMockWithPlayers()
		assert.Contains(t, p.Output(m, nil), "切札は40枚")
	})

	// cuiCardStr は design 5/6 を UNKNOWN に落とすので、切札とマットが全部同じ
	// 文字列になってしまう。専用の描画が要る。
	t.Run("trumps and the Matto are named, not UNKNOWN", func(t *testing.T) {
		m, players := setupMinchiateCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.MinchiateTrumpDesign, 19, false))
		players[0].AddCard(domain.NewCard(domain.MinchiateMattoDesign, domain.MinchiateMattoValue, false))
		out := p.Output(m, nil)
		assert.NotContains(t, out, "UNKNOWN")
		assert.Contains(t, out, "マット")
		// **切札は番号だけでなく呼び名も出す。**40 枚あるので番号だけでは何の札か
		// 分からない。19 は Aria。
		assert.Contains(t, out, "切札19")
		assert.Contains(t, out, domain.MinchiateTrumpName(19))
	})

	t.Run("scarto prompt names the count", func(t *testing.T) {
		m, _ := setupMinchiateCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MinchiatePhaseScarto)
		out := p.Output(m, nil)
		assert.Contains(t, out, fmt.Sprintf("%d 枚を捨てて", domain.MinchiateSurplus))
		assert.Contains(t, out, "scarto")
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupMinchiateCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MinchiatePhaseTrickEnd)
		assert.Contains(t, p.Output(m, nil), "トリック終了")
	})

	t.Run("round end lists every player's trick count", func(t *testing.T) {
		m, players := setupMinchiateCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MinchiatePhaseRoundEnd)
		players[0].AddTrick([]*domain.Card{domain.NewCard(1, 7, false)})
		assert.Contains(t, p.Output(m, nil), "各プレイヤーのトリック数")
	})

	t.Run("trick cards are rendered", func(t *testing.T) {
		m, _ := setupMinchiateCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.MinchiateTrumpDesign, 12, false)},
		})
		assert.Contains(t, p.Output(m, nil), domain.MinchiateTrumpName(12))
	})

	t.Run("error block is rendered", func(t *testing.T) {
		m, _ := setupMinchiateCuiMockWithPlayers()
		assert.Contains(t, p.Output(m, errors.New("boom")), "boom")
	})

	t.Run("game end names the winning team", func(t *testing.T) {
		m, _ := setupMinchiateCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(1)
		assert.Contains(t, p.Output(m, nil), "チーム1の勝ち")
	})

	// 同点は winnerTeam = -1。チーム -1 を勝者として書いてはならない。
	t.Run("a draw is announced as a draw", func(t *testing.T) {
		m, _ := setupMinchiateCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(-1)
		out := p.Output(m, nil)
		assert.Contains(t, out, "引き分け")
		assert.NotContains(t, out, "-1")
	})

	t.Run("nil player is skipped", func(t *testing.T) {
		m := setupMinchiateCuiMock()
		m.On("GetPlayerCnt").Return(1)
		m.On("GetPlayer", 0).Return((*domain.MinchiatePlayer)(nil))
		assert.NotPanics(t, func() { p.Output(m, nil) })
	})
}

func TestMinchiateCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.MinchiateCuiPresenter)

	// ドメインが返しうる 6 種すべてが訳に解決すること。#4660 では 3 種が生キーで
	// 画面に出る状態のまま通っていた。
	t.Run("every emitted reason resolves to prose", func(t *testing.T) {
		for _, reason := range []string{
			"lead_low", "lead_trump", "play_matto", "follow_trump", "follow_low",
		} {
			m, players := setupMinchiateCuiMockWithPlayers()
			players[0].AddCard(domain.NewCard(1, 9, false))
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
			m.On("GetHint").Return(&domain.MinchiateHint{CardIndices: []int{0}, Reason: reason})
			out := p.HintOutput(m)
			assert.NotContains(t, out, reason, "reason %q printed as a raw key", reason)
			assert.Contains(t, out, "[0]")
		}
	})

	t.Run("hint with no card indices still renders the reason", func(t *testing.T) {
		m, _ := setupMinchiateCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.MinchiateHint{Reason: "lead_low"})
		assert.Contains(t, p.HintOutput(m), "-")
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupMinchiateCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.MinchiateHint)(nil))
		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})
}

func TestMinchiateCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MinchiateCuiPresenter)
	m := new(interfaces.MockMinchiateGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You play Angelo"},
	})
	assert.Contains(t, p.ActionLogOutput(m), "play")
}

// #5715: マットは「トリックを取らず、フォロー義務も免れ、リードスートも定めない」
// という他の全カードと違う挙動なのに、その説明はヒント文言の中にしか無く、
// ヒントを切っている人には伝わらなかった (切札40枚の注記は常設なのに)。
func TestMinchiateCuiPresenter_AlwaysExplainsTheMatto(t *testing.T) {
	p := new(presenter.MinchiateCuiPresenter)
	g := domain.NewDefaultMinchiate()
	g.Reset()

	out := p.Output(g, nil)

	assert.Contains(t, out, i18n.T("minchiate.mattoNote"))
	// 切札の注記と並んで、常に出ること。
	assert.Contains(t, out, i18n.T("minchiate.trumpCountNote"))
}
