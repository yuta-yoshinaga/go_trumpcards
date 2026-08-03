//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func makeAluettePlayers() []*domain.AluettePlayer {
	ps := make([]*domain.AluettePlayer, domain.AluettePlayerCnt)
	ps[0] = domain.NewAluettePlayer(true)
	for i := 1; i < domain.AluettePlayerCnt; i++ {
		ps[i] = domain.NewAluettePlayer(false)
	}
	return ps
}

func setupAluetteWebMock() *interfaces.MockAluetteGame {
	m := new(interfaces.MockAluetteGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.AluettePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTeamScores").Return([2]int{0, 0})
	m.On("GetRoundTricks").Return([domain.AluettePlayerCnt]int{})
	m.On("GetLastTrickWinner").Return(-1)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("GetConfig").Return(domain.DefaultAluetteConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()
	return m
}

func setupAluetteWebMockWithPlayers() (*interfaces.MockAluetteGame, []*domain.AluettePlayer) {
	m := setupAluetteWebMock()
	players := makeAluettePlayers()
	m.On("GetPlayerCnt").Return(domain.AluettePlayerCnt)
	for i := 0; i < domain.AluettePlayerCnt; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestAluetteWebPresenter_Output(t *testing.T) {
	p := new(presenter.AluetteWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupAluetteWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(1, 13, false))
		players[1].AddCard(domain.NewCard(2, 6, false))

		var res controller.AluetteWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &res))
		assert.Len(t, res.Players, domain.AluettePlayerCnt)
		assert.Equal(t, "aluette.playPhase.lead", res.MessageCode)
		assert.Len(t, res.Players[0].Cards, 1)
		assert.Len(t, res.Players[1].Cards, 0, "CPU hands stay hidden")
		assert.Equal(t, []int{0}, res.PlayableIndices)
		assert.True(t, res.Players[0].IsDealer)
	})

	// 対面同士が組む。席のチーム番号がそのまま出ていないと、味方のトリックを
	// 奪いに行く手が「正しく」見えてしまう。
	t.Run("teams pair the opposite seats", func(t *testing.T) {
		m, _ := setupAluetteWebMockWithPlayers()
		var res controller.AluetteWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &res))
		assert.Equal(t, res.Players[0].Team, res.Players[2].Team)
		assert.Equal(t, res.Players[1].Team, res.Players[3].Team)
		assert.NotEqual(t, res.Players[0].Team, res.Players[1].Team)
	})

	// **序列表をレスポンスに載せる。**どの 6 枚がリュエットかを知らずにこの
	// ゲームは遊べないのに、表をフロントに複製すればいずれドメインとずれる。
	t.Run("every response carries the luette table in strength order", func(t *testing.T) {
		m, players := setupAluetteWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(4, 3, false)) // Monsieur — 表と同じ札を手札に置く
		var res controller.AluetteWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &res))
		table := domain.AluetteLuetteTable()
		assert.Len(t, res.Luettes, len(table))
		assert.Equal(t, "Monsieur", res.Luettes[0].Name)

		// Monsieur は手札に置いた金貨の 3。札側の表記とそのまま突き合う。
		assert.Equal(t, res.Players[0].Cards[0].Design, res.Luettes[0].Design)

		prev := 1 << 30
		for i, l := range res.Luettes {
			assert.Equal(t, table[i].Name, l.Name, "序列がドメインの表とずれている")
			assert.Equal(t, table[i].Value, l.Value)
			// **スートは盤面の札と同じ表記。**数値で送ると画面側が対応表を持つ。
			// リュエットは聖杯(HEART)と金貨(DIAMOND)の 2 スートにしか無い。
			assert.Contains(t, []string{"HEART", "DIAMOND"}, l.Design,
				"%s のスートが札と同じ表記で届いていない", l.Name)
			r := domain.AluetteRank(domain.NewCard(table[i].Design, l.Value, true))
			assert.Less(t, r, prev, "%s がテーブル順どおりに弱くなっていない", l.Name)
			prev = r
		}
	})

	// **48 枚は標準デッキの部分集合。**10 が抜けているだけなので PNG アートが
	// そのまま使える。手続き記述子を送ると既存アートを捨てることになる。
	t.Run("cards use the standard art, not the procedural path", func(t *testing.T) {
		m, players := setupAluetteWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(4, 3, false))
		players[0].AddCard(domain.NewCard(1, 13, false))

		var res controller.AluetteWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &res))
		for _, c := range res.Players[0].Cards {
			assert.Empty(t, c.Deck, "standard 52-card art must not be replaced")
		}
	})

	t.Run("follow / trick end / round end message codes", func(t *testing.T) {
		for phase, want := range map[domain.AluettePhase]string{
			domain.AluettePhaseTrickEnd: "aluette.trickEnd",
			domain.AluettePhaseRoundEnd: "aluette.roundEnd",
		} {
			m, _ := setupAluetteWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.On("GetPhase").Return(phase)
			var res controller.AluetteWebOutput
			assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &res))
			assert.Equal(t, want, res.MessageCode)
		}

		m, _ := setupAluetteWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(3, 3, false)},
		})
		var res controller.AluetteWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &res))
		assert.Equal(t, "aluette.playPhase.follow", res.MessageCode)
		assert.Equal(t, "HEART", res.CurrentTrick[0].Card.Design)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupAluetteWebMockWithPlayers()
		var res controller.AluetteWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, errors.New("boom"))), &res))
		assert.Equal(t, "boom", res.Message)
		assert.Empty(t, res.MessageCode)
	})

	// 勝敗はチーム単位。人間の席番号ではなく、人間の属するチームで判定する。
	t.Run("game end reports the human's team, a rival team, or a draw", func(t *testing.T) {
		cases := map[int]string{
			0:  "aluette.result.humanWin",
			1:  "aluette.result.cpuWin",
			-1: "aluette.result.draw",
		}
		for winner, want := range cases {
			m, _ := setupAluetteWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
			m.On("GetGameEndFlag").Return(true)
			m.On("GetWinnerTeam").Return(winner)
			var res controller.AluetteWebOutput
			assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &res))
			assert.Equal(t, want, res.MessageCode, "winner team %d", winner)
		}
	})

	t.Run("no playable indices outside the human play turn", func(t *testing.T) {
		m, _ := setupAluetteWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanTurn")
		m.On("IsHumanTurn").Return(false)
		var res controller.AluetteWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &res))
		assert.Empty(t, res.PlayableIndices)
	})

	t.Run("nil playable indices become an empty slice", func(t *testing.T) {
		m, _ := setupAluetteWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayableIndices")
		m.On("GetPlayableIndices", 0).Return(([]int)(nil))
		var res controller.AluetteWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &res))
		assert.NotNil(t, res.PlayableIndices)
		assert.Empty(t, res.PlayableIndices)
	})

	t.Run("nil player is skipped", func(t *testing.T) {
		m := setupAluetteWebMock()
		m.On("GetPlayerCnt").Return(1)
		m.On("GetPlayer", 0).Return((*domain.AluettePlayer)(nil))
		assert.NotPanics(t, func() { p.Output(m, nil) })
	})
}

func TestAluetteWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.AluetteWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupAluetteWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.AluetteHint{CardIndices: []int{2}, Reason: "play_luette"})
		var res controller.AluetteWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.HintOutput(m)), &res))
		assert.Equal(t, []int{2}, res.Hint.CardIndices)
		assert.Equal(t, "play_luette", res.Hint.Reason)
		assert.Equal(t, "aluette.hintRequested", res.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupAluetteWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.AluetteHint)(nil))
		var res controller.AluetteWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.HintOutput(m)), &res))
		assert.Nil(t, res.Hint)
		assert.Equal(t, "aluette.noHint", res.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestAluetteWebPresenterOutputCarriesTheHint(t *testing.T) {
	m, _ := setupAluetteWebMockWithPlayers()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
	m.On("GetHint").Return(&domain.AluetteHint{CardIndices: []int{0}, Reason: "lead_low"})

	out := new(presenter.AluetteWebPresenter).Output(m, nil)
	assert.Contains(t, out, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	assert.NotContains(t, out, "aluette.hintRequested")
}

func TestAluetteWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.AluetteWebPresenter)
	m := new(interfaces.MockAluetteGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "Monsieur"},
	})
	assert.Contains(t, p.ActionLogOutput(m), `"actionType":"play"`)
}
