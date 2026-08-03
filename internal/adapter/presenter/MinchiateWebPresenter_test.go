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

func makeMinchiatePlayers() []*domain.MinchiatePlayer {
	ps := make([]*domain.MinchiatePlayer, domain.MinchiatePlayerCnt)
	ps[0] = domain.NewMinchiatePlayer(true)
	for i := 1; i < domain.MinchiatePlayerCnt; i++ {
		ps[i] = domain.NewMinchiatePlayer(false)
	}
	return ps
}

func setupMinchiateWebMock() *interfaces.MockMinchiateGame {
	m := new(interfaces.MockMinchiateGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.MinchiatePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(0)
	m.On("GetScartoSize").Return(domain.MinchiateSurplus)
	m.On("GetTeamScores").Return([2]int{0, 0})
	m.On("GetRoundTricks").Return([domain.MinchiatePlayerCnt]int{})
	m.On("GetLastTrickWinner").Return(-1)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("IsHumanScartoTurn").Return(false)
	m.On("GetConfig").Return(domain.DefaultMinchiateConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()
	return m
}

func setupMinchiateWebMockWithPlayers() (*interfaces.MockMinchiateGame, []*domain.MinchiatePlayer) {
	m := setupMinchiateWebMock()
	players := makeMinchiatePlayers()
	m.On("GetPlayerCnt").Return(domain.MinchiatePlayerCnt)
	for i := 0; i < domain.MinchiatePlayerCnt; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestMinchiateWebPresenter_Output(t *testing.T) {
	p := new(presenter.MinchiateWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupMinchiateWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(1, 14, false))
		players[1].AddCard(domain.NewCard(2, 6, false))

		var res controller.MinchiateWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &res))
		assert.Len(t, res.Players, domain.MinchiatePlayerCnt)
		assert.Equal(t, "minchiate.playPhase.lead", res.MessageCode)
		assert.Len(t, res.Players[0].Cards, 1)
		assert.Len(t, res.Players[1].Cards, 0, "CPU hands stay hidden")
		assert.Equal(t, []int{0}, res.PlayableIndices)
		assert.Equal(t, domain.MinchiateSurplus, res.ScartoCount)
		assert.True(t, res.Players[0].IsDealer)
	})

	// 対面同士が組む。席のチーム番号がそのまま出ていないと、味方のトリックを
	// 奪いに行く手が「正しく」見えてしまう。
	t.Run("teams pair the opposite seats", func(t *testing.T) {
		m, _ := setupMinchiateWebMockWithPlayers()
		var res controller.MinchiateWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &res))
		assert.Equal(t, res.Players[0].Team, res.Players[2].Team)
		assert.Equal(t, res.Players[1].Team, res.Players[3].Team)
		assert.NotEqual(t, res.Players[0].Team, res.Players[1].Team)
	})

	// 97 枚デッキには専用アートが無いので、全札が手続き記述子を持って届く必要がある。
	t.Run("cards carry the procedural face descriptor", func(t *testing.T) {
		m, players := setupMinchiateWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.MinchiateTrumpDesign, 20, false))
		players[0].AddCard(domain.NewCard(domain.MinchiateTrumpDesign, domain.MinchiateMaxTrump, false))
		players[0].AddCard(domain.NewCard(domain.MinchiateMattoDesign, domain.MinchiateMattoValue, false))
		players[0].AddCard(domain.NewCard(1, 13, false))

		var res controller.MinchiateWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &res))
		cards := res.Players[0].Cards
		for _, c := range cards {
			assert.Equal(t, "tarot", c.Deck, "every card needs the procedural path")
		}
		// **切札は番号ではなく呼び名をラベルに出す。**40 枚あるので番号だけでは
		// 「35 と 36 のどちらが強いか」以外の情報が画面から消える。
		assert.Equal(t, domain.MinchiateTrumpName(20), cards[0].Label)
		assert.Equal(t, domain.MinchiateTrumpName(domain.MinchiateMaxTrump), cards[1].Label)
		assert.NotEqual(t, cards[0].Label, cards[1].Label, "each trump needs its own name")
		assert.Equal(t, "Matto", cards[2].Label)
		assert.Equal(t, "R", cards[3].Label)
	})

	t.Run("scarto phase message code", func(t *testing.T) {
		m, _ := setupMinchiateWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanScartoTurn")
		m.On("GetPhase").Return(domain.MinchiatePhaseScarto)
		m.On("IsHumanScartoTurn").Return(true)
		var res controller.MinchiateWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &res))
		assert.Equal(t, "minchiate.scartoPhase", res.MessageCode)
		assert.True(t, res.IsHumanScarto)
		assert.Empty(t, res.PlayableIndices, "no cards are playable during the scarto")
	})

	t.Run("follow / trick end / round end message codes", func(t *testing.T) {
		for phase, want := range map[domain.MinchiatePhase]string{
			domain.MinchiatePhaseTrickEnd: "minchiate.trickEnd",
			domain.MinchiatePhaseRoundEnd: "minchiate.roundEnd",
		} {
			m, _ := setupMinchiateWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.On("GetPhase").Return(phase)
			var res controller.MinchiateWebOutput
			assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &res))
			assert.Equal(t, want, res.MessageCode)
		}

		m, _ := setupMinchiateWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.MinchiateTrumpDesign, 3, false)},
		})
		var res controller.MinchiateWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &res))
		assert.Equal(t, "minchiate.playPhase.follow", res.MessageCode)
		assert.Equal(t, "tarot", res.CurrentTrick[0].Card.Deck, "trick cards need the descriptor too")
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupMinchiateWebMockWithPlayers()
		var res controller.MinchiateWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, errors.New("boom"))), &res))
		assert.Equal(t, "boom", res.Message)
		assert.Empty(t, res.MessageCode)
	})

	// 勝敗はチーム単位。人間の席番号ではなく、人間の属するチームで判定する。
	t.Run("game end reports the human's team, a rival team, or a draw", func(t *testing.T) {
		cases := map[int]string{
			0:  "minchiate.result.humanWin",
			1:  "minchiate.result.cpuWin",
			-1: "minchiate.result.draw",
		}
		for winner, want := range cases {
			m, _ := setupMinchiateWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
			m.On("GetGameEndFlag").Return(true)
			m.On("GetWinnerTeam").Return(winner)
			var res controller.MinchiateWebOutput
			assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &res))
			assert.Equal(t, want, res.MessageCode, "winner team %d", winner)
		}
	})

	t.Run("no playable indices outside the human play turn", func(t *testing.T) {
		m, _ := setupMinchiateWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanTurn")
		m.On("IsHumanTurn").Return(false)
		var res controller.MinchiateWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &res))
		assert.Empty(t, res.PlayableIndices)
	})

	t.Run("nil player is skipped", func(t *testing.T) {
		m := setupMinchiateWebMock()
		m.On("GetPlayerCnt").Return(1)
		m.On("GetPlayer", 0).Return((*domain.MinchiatePlayer)(nil))
		assert.NotPanics(t, func() { p.Output(m, nil) })
	})
}

func TestMinchiateWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.MinchiateWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupMinchiateWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.MinchiateHint{CardIndices: []int{2}, Reason: "lead_trump"})
		var res controller.MinchiateWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.HintOutput(m)), &res))
		assert.Equal(t, []int{2}, res.Hint.CardIndices)
		assert.Equal(t, "lead_trump", res.Hint.Reason)
		assert.Equal(t, "minchiate.hintRequested", res.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupMinchiateWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.MinchiateHint)(nil))
		var res controller.MinchiateWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.HintOutput(m)), &res))
		assert.Nil(t, res.Hint)
		assert.Equal(t, "minchiate.noHint", res.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestMinchiateWebPresenterOutputCarriesTheHint(t *testing.T) {
	m, _ := setupMinchiateWebMockWithPlayers()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
	m.On("GetHint").Return(&domain.MinchiateHint{CardIndices: []int{0}, Reason: "lead_trump"})

	out := new(presenter.MinchiateWebPresenter).Output(m, nil)
	assert.Contains(t, out, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	assert.NotContains(t, out, "minchiate.hintRequested")
}

func TestMinchiateWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MinchiateWebPresenter)
	m := new(interfaces.MockMinchiateGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You play Angelo"},
	})
	assert.Contains(t, p.ActionLogOutput(m), `"actionType":"play"`)
}
