//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func makeMaoPlayers() []*domain.MaoPlayer {
	return []*domain.MaoPlayer{
		domain.NewMaoPlayer(true),
		domain.NewMaoPlayer(false),
		domain.NewMaoPlayer(false),
		domain.NewMaoPlayer(false),
	}
}

func setupMaoWebMock() *interfaces.MockMaoGame {
	m := new(interfaces.MockMaoGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(30)
	m.On("GetDirection").Return(1)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetChosenSuit").Return(-1)
	m.On("GetPenaltyDrawCount").Return(0)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.MaoPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultMaoConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetAwaitingWord").Return(false)
	m.On("GetPlayerCorrectCount").Return(0)
	m.On("GetHintUnlocked").Return(false)
	m.On("GetRuleHintKey").Return("")
	m.On("GetRulePenaltyFlag").Return(false)
	return m
}

func setupMaoWebMockWithPlayers() (*interfaces.MockMaoGame, []*domain.MaoPlayer) {
	m := setupMaoWebMock()
	players := makeMaoPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestMaoWebPresenter_Output(t *testing.T) {
	p := new(presenter.MaoWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupMaoWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		var resObj controller.MaoWebOutput
		err := json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resObj.Players))
		assert.Equal(t, 0, resObj.Phase)
		assert.Equal(t, "mao.playPhase", resObj.MessageCode)
		assert.False(t, resObj.AwaitingWord)
		assert.False(t, resObj.HintUnlocked)
		assert.Empty(t, resObj.RuleHint)
	})

	t.Run("hidden rule NEVER leaks in raw JSON", func(t *testing.T) {
		m, players := setupMaoWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		// Even when the hint IS unlocked, only the vague hint is exposed, never
		// the trigger value or the required word.
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHintUnlocked")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRuleHintKey")
		m.On("GetHintUnlocked").Return(true)
		m.On("GetRuleHintKey").Return("hintSuit")

		raw := p.Output(m, nil)
		// The web output must not contain any hidden-rule field names.
		assert.NotContains(t, raw, "hiddenRule")
		assert.NotContains(t, raw, "requiredWord")
		assert.NotContains(t, strings.ToLower(raw), "triggervalue")

		var resObj controller.MaoWebOutput
		_ = json.Unmarshal([]byte(raw), &resObj)
		assert.True(t, resObj.HintUnlocked)
		// **コードを返す。**Web サーバの i18n 言語はプロセス全体で 1 つなので、
		// ブラウザ側で訳せるようコードも載せる (#4917)。
		assert.Equal(t, "hintSuit", resObj.RuleHintCode)
		assert.Equal(t, "あるスートを出したときに言葉が必要です。", resObj.RuleHint)
	})

	t.Run("awaiting word flag and message", func(t *testing.T) {
		m, _ := setupMaoWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetAwaitingWord")
		m.On("GetAwaitingWord").Return(true)
		var resObj controller.MaoWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.True(t, resObj.AwaitingWord)
		assert.Equal(t, "mao.awaitingWord", resObj.MessageCode)
	})

	t.Run("rule penalty flag and message", func(t *testing.T) {
		m, _ := setupMaoWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRulePenaltyFlag")
		m.On("GetRulePenaltyFlag").Return(true)
		var resObj controller.MaoWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.True(t, resObj.RulePenalty)
		assert.Equal(t, "mao.rulePenalty", resObj.MessageCode)
	})

	t.Run("human cards shown, CPU hidden", func(t *testing.T) {
		m, players := setupMaoWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		var resObj controller.MaoWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupMaoWebMockWithPlayers()
		var resObj controller.MaoWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, errors.New("test error"))), &resObj)
		assert.Equal(t, "test error", resObj.Message)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupMaoWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		var resObj controller.MaoWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "mao.result.humanWin", resObj.MessageCode)
	})

	t.Run("game end CPU wins", func(t *testing.T) {
		m, _ := setupMaoWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(2)
		var resObj controller.MaoWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Equal(t, "mao.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "2"}, resObj.MessageParams)
	})

	t.Run("choose suit / must declare / round end phases", func(t *testing.T) {
		for phase, code := range map[domain.MaoPhase]string{
			domain.MaoPhaseChooseSuit:  "mao.chooseSuitPhase",
			domain.MaoPhaseMustDeclare: "mao.mustDeclarePhase",
			domain.MaoPhaseRoundEnd:    "mao.roundEnd",
		} {
			m, _ := setupMaoWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.On("GetPhase").Return(phase)
			var resObj controller.MaoWebOutput
			_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
			assert.Equal(t, code, resObj.MessageCode)
		}
	})

	t.Run("discard top populated", func(t *testing.T) {
		m, _ := setupMaoWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
		var resObj controller.MaoWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.NotNil(t, resObj.DiscardTop)
	})
}

func TestMaoWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MaoWebPresenter)
	m := new(interfaces.MockMaoGame)
	entries := []*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played SPADE 5"},
	}
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return(entries)
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}
