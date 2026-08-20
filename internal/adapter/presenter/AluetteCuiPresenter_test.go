//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupAluetteCuiMock() *interfaces.MockAluetteGame {
	m := new(interfaces.MockAluetteGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.AluettePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTeamScores").Return([2]int{0, 0})
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetConfig").Return(domain.DefaultAluetteConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetHint").Return(nil).Maybe()
	m.On("GetRoundTricks").Return([domain.AluettePlayerCnt]int{}).Maybe()
	return m
}

func setupAluetteCuiMockWithPlayers() (*interfaces.MockAluetteGame, []*domain.AluettePlayer) {
	m := setupAluetteCuiMock()
	players := makeAluettePlayers()
	m.On("GetPlayerCnt").Return(domain.AluettePlayerCnt)
	for i := 0; i < domain.AluettePlayerCnt; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestAluetteCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.AluetteCuiPresenter)

	t.Run("play phase shows the hand and the teams", func(t *testing.T) {
		m, players := setupAluetteCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(1, 13, false))
		out := p.Output(m, nil)
		assert.Contains(t, out, "アリュエット")
		assert.Contains(t, out, "[0]")
		assert.Contains(t, out, "チーム")
		assert.Contains(t, out, "[親]", "the dealer must be marked")
	})

	// **序列表は毎フレーム出す。**6 枚を覚えていないと着手が選べないゲームで、
	// 画面のどこにも書いていなければ手札の意味が読めない。
	t.Run("the luette table is listed on every frame, strongest first", func(t *testing.T) {
		m, _ := setupAluetteCuiMockWithPlayers()
		out := p.Output(m, nil)
		last := -1
		for _, l := range domain.AluetteLuetteTable() {
			at := indexOfSubstring(out, l.Name)
			assert.GreaterOrEqual(t, at, 0, "%s が凡例に出ていない", l.Name)
			assert.Greater(t, at, last, "%s が強さ順に並んでいない", l.Name)
			last = at
		}
	})

	// **手札のリュエットには呼び名を添える。**「♦3」とだけ出ると、それがデッキ
	// 最強の Monsieur なのか、ただの 3 なのかが読み取れない。
	t.Run("a luette in hand carries its name", func(t *testing.T) {
		m, players := setupAluetteCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(4, 3, false)) // Monsieur
		players[0].AddCard(domain.NewCard(1, 3, false)) // ただの 3
		out := p.Output(m, nil)
		assert.Contains(t, out, "〈Monsieur〉")
		assert.NotContains(t, out, "UNKNOWN")
	})

	t.Run("the target score is shown with the team scores", func(t *testing.T) {
		m, _ := setupAluetteCuiMockWithPlayers()
		assert.Contains(t, p.Output(m, nil), "点先取")
	})

	// フォロー義務が無いことは、トリックテイキングとしては例外的な規則。
	// プロンプトが黙っていると、遊ぶ側は既定のマストフォローを仮定してしまう。
	t.Run("the play prompt states that there is no follow obligation", func(t *testing.T) {
		m, _ := setupAluetteCuiMockWithPlayers()
		assert.Contains(t, p.Output(m, nil), "フォロー義務はありません")
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupAluetteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.AluettePhaseTrickEnd)
		assert.Contains(t, p.Output(m, nil), "トリック終了")
	})

	t.Run("round end lists every player's trick count", func(t *testing.T) {
		m, players := setupAluetteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.AluettePhaseRoundEnd)
		players[0].AddTrick([]*domain.Card{domain.NewCard(1, 7, false)})
		assert.Contains(t, p.Output(m, nil), "各プレイヤーのトリック数")
	})

	t.Run("trick cards are rendered", func(t *testing.T) {
		m, _ := setupAluetteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(3, 2, false)}, // Borgne
		})
		assert.Contains(t, p.Output(m, nil), "〈Borgne〉")
	})

	t.Run("error block is rendered", func(t *testing.T) {
		m, _ := setupAluetteCuiMockWithPlayers()
		assert.Contains(t, p.Output(m, errors.New("boom")), "boom")
	})

	t.Run("game end names the winning team", func(t *testing.T) {
		m, _ := setupAluetteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(1)
		assert.Contains(t, p.Output(m, nil), "チーム1の勝ち")
	})

	// 同点は winnerTeam = -1。チーム -1 を勝者として書いてはならない。
	t.Run("a draw is announced as a draw", func(t *testing.T) {
		m, _ := setupAluetteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(-1)
		out := p.Output(m, nil)
		assert.Contains(t, out, "引き分け")
		assert.NotContains(t, out, "チーム-1")
	})

	t.Run("nil player is skipped", func(t *testing.T) {
		m := setupAluetteCuiMock()
		m.On("GetPlayerCnt").Return(1)
		m.On("GetPlayer", 0).Return((*domain.AluettePlayer)(nil))
		assert.NotPanics(t, func() { p.Output(m, nil) })
	})
}

func TestAluetteCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.AluetteCuiPresenter)

	// ドメインが返しうる 4 種すべてが訳に解決すること。外れた理由は
	// hintReasonStr の既定でキー文字列そのものが画面に出る (#4660)。
	t.Run("every emitted reason resolves to prose", func(t *testing.T) {
		for _, reason := range []string{
			"lead_low", "play_luette", "partner_winning", "follow_low",
		} {
			m, players := setupAluetteCuiMockWithPlayers()
			players[0].AddCard(domain.NewCard(1, 9, false))
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
			m.On("GetHint").Return(&domain.AluetteHint{CardIndices: []int{0}, Reason: reason})
			out := p.HintOutput(m)
			assert.NotContains(t, out, reason, "reason %q printed as a raw key", reason)
			assert.Contains(t, out, "[0]")
		}
	})

	t.Run("hint with no card indices still renders the reason", func(t *testing.T) {
		m, _ := setupAluetteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.AluetteHint{Reason: "lead_low"})
		assert.Contains(t, p.HintOutput(m), "-")
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupAluetteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.AluetteHint)(nil))
		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})
}

func TestAluetteCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.AluetteCuiPresenter)
	m := new(interfaces.MockAluetteGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "Monsieur"},
	})
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewAluettePlayer(true)).Maybe()
	assert.Contains(t, p.ActionLogOutput(m), "play")
}

// #5714: メーヌの勝敗は**チーム合計が 3 以上か**で決まる (4-1 でも 3-2 でも 1 点)。
// それなのに Web も CUI も個人トリック数を並べるだけで、チーム集計も勝者チームも
// 出しておらず、プレイヤーが自分で足し算する必要があった。
func TestAluetteCuiPresenter_ShowsTheTeamTally(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.AluetteCuiPresenter)

	atRoundEnd := func(tricks [domain.AluettePlayerCnt]int) *interfaces.MockAluetteGame {
		m, players := setupAluetteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.AluettePhaseRoundEnd)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundTricks")
		m.On("GetRoundTricks").Return(tricks)
		for i, n := range tricks {
			for range n {
				players[i].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
			}
		}
		return m
	}

	// 席 0/2 がチーム0、1/3 がチーム1。3-2 でもチーム0 の勝ち。
	t.Run("adds up each team and names the winner", func(t *testing.T) {
		out := p.Output(atRoundEnd([domain.AluettePlayerCnt]int{2, 1, 1, 1}), nil)

		assert.Contains(t, out, i18n.Tf("aluette.roundEndTeamTally",
			"team0", "3", "team1", "2"))
		assert.Contains(t, out, i18n.Tf("aluette.roundEndMeineWinner",
			"team", i18n.Tf("aluette.teamName", "n", "0")))
	})

	t.Run("names the other team when it takes the majority", func(t *testing.T) {
		out := p.Output(atRoundEnd([domain.AluettePlayerCnt]int{1, 3, 0, 1}), nil)

		assert.Contains(t, out, i18n.Tf("aluette.roundEndTeamTally",
			"team0", "1", "team1", "4"))
		assert.Contains(t, out, i18n.Tf("aluette.roundEndMeineWinner",
			"team", i18n.Tf("aluette.teamName", "n", "1")))
	})
}
