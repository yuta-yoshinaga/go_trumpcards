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

func setupMusWebMock() *interfaces.MockMusGame {
	m := new(interfaces.MockMusGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetPhase").Return(domain.MusPhaseMus)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetManoIdx").Return(0)
	m.On("GetBetTeam").Return(0)
	m.On("GetPendingStake").Return(0)
	m.On("GetLastBettorTeam").Return(-1)
	m.On("GetMusTurn").Return(0)
	m.On("GetDiscardTurn").Return(0)
	m.On("GetMusCycle").Return(0)
	m.On("GetAmarrakos").Return([domain.MusTeamCnt]int{0, 0})
	for ri := 0; ri < domain.MusRoundCnt; ri++ {
		m.On("GetResult", ri).Return(domain.MusRoundResult{Kind: domain.MusResultPending, Stake: 0, Team: -1})
	}
	m.On("GetWinnerTeam").Return(-1)
	m.On("IsHumanTurn").Return(true)
	m.On("GetConfig").Return(domain.DefaultMusConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupMusWebMockWithPlayers() (*interfaces.MockMusGame, []*domain.MusPlayer) {
	m := setupMusWebMock()
	players := makeMusPlayers()
	m.On("GetPlayerCnt").Return(domain.MusPlayerCnt)
	for i, p := range players {
		m.On("GetPlayer", i).Return(p)
	}
	return m, players
}

func TestMusWebPresenter_Output(t *testing.T) {
	p := new(presenter.MusWebPresenter)

	t.Run("initial state mus phase", func(t *testing.T) {
		m, players := setupMusWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))

		result := p.Output(m, nil)
		var resObj controller.MusWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, domain.MusPlayerCnt)
		assert.Equal(t, int(domain.MusPhaseMus), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, "mus.musPhase", resObj.MessageCode)
		// human cards visible, CPU hidden
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupMusWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.MusWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.MusCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 40, resObj.Config.TargetAmarrakos)
	})

	t.Run("discard phase message code", func(t *testing.T) {
		m, _ := setupMusWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseDiscard)
		result := p.Output(m, nil)
		var resObj controller.MusWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "mus.discardPhase", resObj.MessageCode)
	})

	t.Run("grande phase message code", func(t *testing.T) {
		m, _ := setupMusWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseGrande)
		result := p.Output(m, nil)
		var resObj controller.MusWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "mus.grandePhase", resObj.MessageCode)
	})

	t.Run("chica phase message code", func(t *testing.T) {
		m, _ := setupMusWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseChica)
		result := p.Output(m, nil)
		var resObj controller.MusWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "mus.chicaPhase", resObj.MessageCode)
	})

	t.Run("pares phase message code", func(t *testing.T) {
		m, _ := setupMusWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhasePares)
		result := p.Output(m, nil)
		var resObj controller.MusWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "mus.paresPhase", resObj.MessageCode)
	})

	t.Run("juego phase message code", func(t *testing.T) {
		m, _ := setupMusWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseJuego)
		result := p.Output(m, nil)
		var resObj controller.MusWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "mus.juegoPhase", resObj.MessageCode)
	})

	t.Run("showdown phase message code", func(t *testing.T) {
		m, _ := setupMusWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseShowdown)
		result := p.Output(m, nil)
		var resObj controller.MusWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "mus.showdownPhase", resObj.MessageCode)
	})

	t.Run("round end reveals all hands", func(t *testing.T) {
		m, players := setupMusWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseRoundEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		result := p.Output(m, nil)
		var resObj controller.MusWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "mus.roundEnd", resObj.MessageCode)
		// CPU cards revealed at round end
		assert.Len(t, resObj.Players[1].Cards, 1)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupMusWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.MusWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human team wins", func(t *testing.T) {
		m, _ := setupMusWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0) // team 0 includes player 0 (human)
		result := p.Output(m, nil)
		var resObj controller.MusWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "mus.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu team wins", func(t *testing.T) {
		m, _ := setupMusWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(1) // team 1: CPU
		result := p.Output(m, nil)
		var resObj controller.MusWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "mus.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"team": "1"}, resObj.MessageParams)
	})

	t.Run("betting phase sets canPaso/canEnvido flags when human turn", func(t *testing.T) {
		m, _ := setupMusWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseGrande)
		result := p.Output(m, nil)
		var resObj controller.MusWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.CanPaso)
		assert.True(t, resObj.CanEnvido)
		assert.True(t, resObj.CanOrdago)
		assert.False(t, resObj.CanQuiero)
		assert.False(t, resObj.CanNoQuiero)
	})

	t.Run("betting phase with pending stake enables quiero/noquiero", func(t *testing.T) {
		m, _ := setupMusWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPendingStake")
		m.On("GetPhase").Return(domain.MusPhaseGrande)
		m.On("GetPendingStake").Return(4)
		result := p.Output(m, nil)
		var resObj controller.MusWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.False(t, resObj.CanPaso)
		assert.True(t, resObj.CanQuiero)
		assert.True(t, resObj.CanNoQuiero)
	})

	t.Run("humanTeam is 0 for player 0 human", func(t *testing.T) {
		m, _ := setupMusWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.MusWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 0, resObj.HumanTeam)
	})
}

func TestMusWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.MusWebPresenter)

	t.Run("hint with indices", func(t *testing.T) {
		m, _ := setupMusWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.MusHint{Indices: []int{1, 3}, Reason: "discard_low"})
		result := p.HintOutput(m)
		var resObj controller.MusWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{1, 3}, resObj.Hint.Indices)
		assert.Equal(t, "discard_low", resObj.Hint.Reason)
	})

	t.Run("mus hint", func(t *testing.T) {
		m, _ := setupMusWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.MusHint{Mus: true, Reason: "mus_exchange"})
		result := p.HintOutput(m)
		var resObj controller.MusWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.True(t, resObj.Hint.Mus)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupMusWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.MusHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.MusWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestMusWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MusWebPresenter)
	m := new(interfaces.MockMusGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "mus", Detail: "You wants mus"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"mus"`)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestMusWebPresenterOutputCarriesTheHint(t *testing.T) {
	msg, _ := setupMusWebMockWithPlayers()
	msg.ExpectedCalls = removeMockCall(msg.ExpectedCalls, "GetHint")
	msg.On("GetHint").Return(&domain.MusHint{Mus: false, Action: 0, Amount: 0, Indices: []int{0}, Reason: "follow_suit"})

	result := new(presenter.MusWebPresenter).Output(msg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "mus.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestMusWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	msg, _ := setupMusWebMockWithPlayers()
	msg.ExpectedCalls = removeMockCall(msg.ExpectedCalls, "GetHint")
	msg.On("GetHint").Return(&domain.MusHint{Mus: false, Action: 0, Amount: 0, Indices: []int{0}, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.MusWebPresenter).HintOutput(msg), "mus.hintRequested")

	none, _ := setupMusWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.MusHint)(nil))
	assert.Contains(t, new(presenter.MusWebPresenter).HintOutput(none), "mus.noHint")
}
