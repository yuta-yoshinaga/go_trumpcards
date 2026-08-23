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

func makeTrappolaPlayers() []*domain.TrappolaPlayer {
	return []*domain.TrappolaPlayer{
		domain.NewTrappolaPlayer(true),
		domain.NewTrappolaPlayer(false),
		domain.NewTrappolaPlayer(false),
		domain.NewTrappolaPlayer(false),
	}
}

func setupTrappolaWebMock() *interfaces.MockTrappolaGame {
	m := new(interfaces.MockTrappolaGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TrappolaPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetTeamScores").Return([domain.TrappolaTeamCnt]int{0, 0})
	m.On("GetTeamRoundThirds").Return([domain.TrappolaTeamCnt]int{0, 0})
	m.On("IsHumanTurn").Return(true)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("GetConfig").Return(domain.DefaultTrappolaConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupTrappolaWebMockWithPlayers() (*interfaces.MockTrappolaGame, []*domain.TrappolaPlayer) {
	m := setupTrappolaWebMock()
	players := makeTrappolaPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestTrappolaWebPresenter_Output(t *testing.T) {
	p := new(presenter.TrappolaWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupTrappolaWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.TrappolaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 4)
		assert.Equal(t, 0, resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, 0, resObj.Players[0].TeamID)
		assert.Equal(t, 1, resObj.Players[1].TeamID)
		assert.Equal(t, "trappola.playPhase.lead", resObj.MessageCode)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
		// human cards visible, cpu hidden
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupTrappolaWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.TrappolaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.TrappolaCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 21, resObj.Config.TargetPoints)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupTrappolaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 3, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.TrappolaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "trappola.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end / round end message codes", func(t *testing.T) {
		for phase, code := range map[domain.TrappolaPhase]string{
			domain.TrappolaPhaseTrickEnd: "trappola.trickEnd",
			domain.TrappolaPhaseRoundEnd: "trappola.roundEnd",
		} {
			m, _ := setupTrappolaWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.On("GetPhase").Return(phase)
			result := p.Output(m, nil)
			var resObj controller.TrappolaWebOutput
			assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
			assert.Equal(t, code, resObj.MessageCode)
		}
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupTrappolaWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.TrappolaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human team wins", func(t *testing.T) {
		m, _ := setupTrappolaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		var resObj controller.TrappolaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "trappola.result.humanTeamWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"team": "A"}, resObj.MessageParams)
	})

	t.Run("game end cpu team wins", func(t *testing.T) {
		m, _ := setupTrappolaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(1)
		result := p.Output(m, nil)
		var resObj controller.TrappolaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "trappola.result.cpuTeamWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"team": "B"}, resObj.MessageParams)
	})
}

func TestTrappolaWebPresenter_LastTrick(t *testing.T) {
	p := new(presenter.TrappolaWebPresenter)

	// resolvedTrickLog returns an action log for one fully-resolved trick:
	// four play entries (players 1,2,3,0) followed by the trick_win entry (winner 2).
	resolvedTrickLog := func() []*domain.ActionLogEntry {
		mk := func(pi, design, value int, at string) *domain.ActionLogEntry {
			return &domain.ActionLogEntry{
				PlayerIdx:  pi,
				ActionType: at,
				Cards:      []*domain.Card{domain.NewCard(design, value, false)},
			}
		}
		return []*domain.ActionLogEntry{
			mk(1, domain.CardDesignSpade, 3, "play"),
			mk(2, domain.CardDesignSpade, 1, "play"),
			mk(3, domain.CardDesignSpade, 5, "play"),
			mk(0, domain.CardDesignSpade, 7, "play"),
			{PlayerIdx: 2, ActionType: "trick_win"},
		}
	}

	t.Run("round start (play phase, trick 1) has empty last trick", func(t *testing.T) {
		m, _ := setupTrappolaWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.TrappolaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Empty(t, resObj.LastTrick)
		assert.Equal(t, -1, resObj.LastTrickWinner)
	})

	t.Run("populated from action log during next trick", func(t *testing.T) {
		m, _ := setupTrappolaWebMockWithPlayers()
		// Playing trick 2 now, so the resolved trick 1 is the last trick.
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrickNumber")
		m.On("GetTrickNumber").Return(2)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetActionLog")
		m.On("GetActionLog").Return(resolvedTrickLog())

		result := p.Output(m, nil)
		var resObj controller.TrappolaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.LastTrick, 4)
		assert.Equal(t, 2, resObj.LastTrickWinner)
		// Order and player mapping preserved from the play log.
		assert.Equal(t, 1, resObj.LastTrick[0].PlayerIdx)
		assert.Equal(t, 3, resObj.LastTrick[0].Card.Value)
		assert.Equal(t, 0, resObj.LastTrick[3].PlayerIdx)
	})
}

func TestTrappolaWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.TrappolaWebPresenter)

	t.Run("hint available", func(t *testing.T) {
		m, _ := setupTrappolaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.TrappolaHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.TrappolaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupTrappolaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.TrappolaHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.TrappolaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestTrappolaWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TrappolaWebPresenter)
	m := new(interfaces.MockTrappolaGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "plays ♠5"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestTrappolaWebPresenterOutputCarriesTheHint(t *testing.T) {
	trg, _ := setupTrappolaWebMockWithPlayers()
	trg.ExpectedCalls = removeMockCall(trg.ExpectedCalls, "GetHint")
	trg.On("GetHint").Return(&domain.TrappolaHint{CardIndices: []int{0}, Reason: "follow_suit"})

	result := new(presenter.TrappolaWebPresenter).Output(trg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "trappola.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestTrappolaWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	trg, _ := setupTrappolaWebMockWithPlayers()
	trg.ExpectedCalls = removeMockCall(trg.ExpectedCalls, "GetHint")
	trg.On("GetHint").Return(&domain.TrappolaHint{CardIndices: []int{0}, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.TrappolaWebPresenter).HintOutput(trg), "trappola.hintRequested")

	none, _ := setupTrappolaWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.TrappolaHint)(nil))
	assert.Contains(t, new(presenter.TrappolaWebPresenter).HintOutput(none), "trappola.noHint")
}
