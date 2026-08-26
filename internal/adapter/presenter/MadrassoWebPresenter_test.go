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

func makeMadrassoPlayers() []*domain.MadrassoPlayer {
	return []*domain.MadrassoPlayer{
		domain.NewMadrassoPlayer(true),
		domain.NewMadrassoPlayer(false),
		domain.NewMadrassoPlayer(false),
		domain.NewMadrassoPlayer(false),
	}
}

func setupMadrassoWebMock() *interfaces.MockMadrassoGame {
	m := new(interfaces.MockMadrassoGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.MadrassoPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetTeamScores").Return([domain.MadrassoTeamCnt]int{0, 0})
	m.On("GetTeamRoundPoints").Return([domain.MadrassoTeamCnt]int{0, 0})
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("IsHumanTurn").Return(true)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("GetConfig").Return(domain.DefaultMadrassoConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupMadrassoWebMockWithPlayers() (*interfaces.MockMadrassoGame, []*domain.MadrassoPlayer) {
	m := setupMadrassoWebMock()
	players := makeMadrassoPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestMadrassoWebPresenter_Output(t *testing.T) {
	p := new(presenter.MadrassoWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupMadrassoWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.MadrassoWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 4)
		assert.Equal(t, 0, resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, 0, resObj.Players[0].TeamID)
		assert.Equal(t, 1, resObj.Players[1].TeamID)
		assert.Equal(t, "madrasso.playPhase.lead", resObj.MessageCode)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
		// human cards visible, cpu hidden
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupMadrassoWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.MadrassoWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.MadrassoCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, domain.DefaultMadrassoConfig().TargetPoints, resObj.Config.TargetPoints)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupMadrassoWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 3, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.MadrassoWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "madrasso.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end / round end message codes", func(t *testing.T) {
		for phase, code := range map[domain.MadrassoPhase]string{
			domain.MadrassoPhaseTrickEnd: "madrasso.trickEnd",
			domain.MadrassoPhaseRoundEnd: "madrasso.roundEnd",
		} {
			m, _ := setupMadrassoWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.On("GetPhase").Return(phase)
			result := p.Output(m, nil)
			var resObj controller.MadrassoWebOutput
			assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
			assert.Equal(t, code, resObj.MessageCode)
		}
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupMadrassoWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.MadrassoWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human team wins", func(t *testing.T) {
		m, _ := setupMadrassoWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		var resObj controller.MadrassoWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "madrasso.result.humanTeamWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"team": "A"}, resObj.MessageParams)
	})

	t.Run("game end cpu team wins", func(t *testing.T) {
		m, _ := setupMadrassoWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(1)
		result := p.Output(m, nil)
		var resObj controller.MadrassoWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "madrasso.result.cpuTeamWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"team": "B"}, resObj.MessageParams)
	})
}

func TestMadrassoWebPresenter_LastTrick(t *testing.T) {
	p := new(presenter.MadrassoWebPresenter)

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
		m, _ := setupMadrassoWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.MadrassoWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Empty(t, resObj.LastTrick)
		assert.Equal(t, -1, resObj.LastTrickWinner)
	})

	t.Run("populated from action log during next trick", func(t *testing.T) {
		m, _ := setupMadrassoWebMockWithPlayers()
		// Playing trick 2 now, so the resolved trick 1 is the last trick.
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrickNumber")
		m.On("GetTrickNumber").Return(2)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetActionLog")
		m.On("GetActionLog").Return(resolvedTrickLog())

		result := p.Output(m, nil)
		var resObj controller.MadrassoWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.LastTrick, 4)
		assert.Equal(t, 2, resObj.LastTrickWinner)
		// Order and player mapping preserved from the play log.
		assert.Equal(t, 1, resObj.LastTrick[0].PlayerIdx)
		assert.Equal(t, 3, resObj.LastTrick[0].Card.Value)
		assert.Equal(t, 0, resObj.LastTrick[3].PlayerIdx)
	})
}

func TestMadrassoWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.MadrassoWebPresenter)

	t.Run("hint available", func(t *testing.T) {
		m, _ := setupMadrassoWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.MadrassoHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.MadrassoWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupMadrassoWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.MadrassoHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.MadrassoWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestMadrassoWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MadrassoWebPresenter)
	m := new(interfaces.MockMadrassoGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "plays ♠5"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestMadrassoWebPresenterOutputCarriesTheHint(t *testing.T) {
	trg, _ := setupMadrassoWebMockWithPlayers()
	trg.ExpectedCalls = removeMockCall(trg.ExpectedCalls, "GetHint")
	trg.On("GetHint").Return(&domain.MadrassoHint{CardIndices: []int{0}, Reason: "follow_suit"})

	result := new(presenter.MadrassoWebPresenter).Output(trg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "madrasso.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestMadrassoWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	trg, _ := setupMadrassoWebMockWithPlayers()
	trg.ExpectedCalls = removeMockCall(trg.ExpectedCalls, "GetHint")
	trg.On("GetHint").Return(&domain.MadrassoHint{CardIndices: []int{0}, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.MadrassoWebPresenter).HintOutput(trg), "madrasso.hintRequested")

	none, _ := setupMadrassoWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.MadrassoHint)(nil))
	assert.Contains(t, new(presenter.MadrassoWebPresenter).HintOutput(none), "madrasso.noHint")
}
