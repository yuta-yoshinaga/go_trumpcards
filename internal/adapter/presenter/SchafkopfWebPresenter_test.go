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

func setupSchafkopfWebMock() *interfaces.MockSchafkopfGame {
	m := new(interfaces.MockSchafkopfGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SchafkopfPhasePick)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetPickerIdx").Return(-1)
	m.On("GetPartnerIdx").Return(-1)
	m.On("GetCalledSuit").Return(0)
	m.On("IsPartnerRevealed").Return(false)
	m.On("GetPassCount").Return(0)
	m.On("GetContract").Return(domain.SchafkopfContractRufspiel)
	m.On("GetSoloSuit").Return(0)
	m.On("GetBeatableContracts").Return([]domain.SchafkopfContract{
		domain.SchafkopfContractRufspiel, domain.SchafkopfContractWenz, domain.SchafkopfContractSolo,
	})
	m.On("GetCallableSuits").Return([]int(nil))
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("GetRoundPickerPoints").Return(0)
	m.On("GetRoundMultiplier").Return(1)
	m.On("GetRoundPickerWon").Return(false)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultSchafkopfConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupSchafkopfWebMockWithPlayers() (*interfaces.MockSchafkopfGame, []*domain.SchafkopfPlayer) {
	m := setupSchafkopfWebMock()
	players := makeSchafkopfPlayers()
	m.On("GetPlayerCnt").Return(5)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	m.On("GetPlayer", 4).Return(players[4])
	return m, players
}

func TestSchafkopfWebPresenter_Output(t *testing.T) {
	p := new(presenter.SchafkopfWebPresenter)

	t.Run("initial state pick phase", func(t *testing.T) {
		m, players := setupSchafkopfWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))

		result := p.Output(m, nil)
		var resObj controller.SchafkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 5)
		assert.Equal(t, int(domain.SchafkopfPhasePick), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Equal(t, -1, resObj.PickerIdx)
		assert.Equal(t, -1, resObj.PartnerIdx)
		assert.Equal(t, "schafkopf.pickPhase", resObj.MessageCode)
		// human cards visible, CPU hidden
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupSchafkopfWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.SchafkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.SchafkopfCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 2, resObj.Config.BaseChips)
		assert.Equal(t, 20, resObj.Config.StartChips)
		assert.Equal(t, 40, resObj.Config.TargetChips)
	})

	t.Run("call phase returns callable suits", func(t *testing.T) {
		m, _ := setupSchafkopfWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCallableSuits")
		m.On("GetPhase").Return(domain.SchafkopfPhaseCall)
		m.On("GetCallableSuits").Return([]int{domain.CardDesignClover, domain.CardDesignSpade})
		result := p.Output(m, nil)
		var resObj controller.SchafkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "schafkopf.callPhase", resObj.MessageCode)
		assert.Equal(t, []int{domain.CardDesignClover, domain.CardDesignSpade}, resObj.CallableSuits)
	})

	t.Run("play phase lead message", func(t *testing.T) {
		m, _ := setupSchafkopfWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SchafkopfPhasePlay)
		result := p.Output(m, nil)
		var resObj controller.SchafkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "schafkopf.playPhase.lead", resObj.MessageCode)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupSchafkopfWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetPhase").Return(domain.SchafkopfPhasePlay)
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 7, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.SchafkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "schafkopf.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupSchafkopfWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SchafkopfPhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.SchafkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "schafkopf.trickEnd", resObj.MessageCode)
	})

	t.Run("partner idx hidden during play until revealed", func(t *testing.T) {
		m, _ := setupSchafkopfWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPartnerIdx")
		m.On("GetPhase").Return(domain.SchafkopfPhasePlay)
		m.On("GetPartnerIdx").Return(3)
		// IsPartnerRevealed is false in default mock
		result := p.Output(m, nil)
		var resObj controller.SchafkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, -1, resObj.PartnerIdx)
	})

	t.Run("partner idx shown when partner revealed", func(t *testing.T) {
		m, _ := setupSchafkopfWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPartnerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsPartnerRevealed")
		m.On("GetPhase").Return(domain.SchafkopfPhasePlay)
		m.On("GetPartnerIdx").Return(3)
		m.On("IsPartnerRevealed").Return(true)
		result := p.Output(m, nil)
		var resObj controller.SchafkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 3, resObj.PartnerIdx)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupSchafkopfWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.SchafkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupSchafkopfWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		var resObj controller.SchafkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "schafkopf.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu wins", func(t *testing.T) {
		m, _ := setupSchafkopfWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(2)
		result := p.Output(m, nil)
		var resObj controller.SchafkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "schafkopf.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "2"}, resObj.MessageParams)
	})
}

func TestSchafkopfWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.SchafkopfWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupSchafkopfWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.SchafkopfHint{CardIndices: []int{2}, Suit: 0, Pick: false, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.SchafkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("pick hint", func(t *testing.T) {
		m, _ := setupSchafkopfWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.SchafkopfHint{Pick: true, Reason: "pick_take"})
		result := p.HintOutput(m)
		var resObj controller.SchafkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.True(t, resObj.Hint.Pick)
		assert.Equal(t, "pick_take", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupSchafkopfWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.SchafkopfHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.SchafkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestSchafkopfWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SchafkopfWebPresenter)
	m := new(interfaces.MockSchafkopfGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "pick", Detail: "You picks up the blind"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"pick"`)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestSchafkopfWebPresenterOutputCarriesTheHint(t *testing.T) {
	shg, _ := setupSchafkopfWebMockWithPlayers()
	shg.ExpectedCalls = removeMockCall(shg.ExpectedCalls, "GetHint")
	shg.On("GetHint").Return(&domain.SchafkopfHint{CardIndices: []int{0}, Suit: 1, Pick: false, Reason: "follow_suit"})

	result := new(presenter.SchafkopfWebPresenter).Output(shg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "schafkopf.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestSchafkopfWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	shg, _ := setupSchafkopfWebMockWithPlayers()
	shg.ExpectedCalls = removeMockCall(shg.ExpectedCalls, "GetHint")
	shg.On("GetHint").Return(&domain.SchafkopfHint{CardIndices: []int{0}, Suit: 1, Pick: false, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.SchafkopfWebPresenter).HintOutput(shg), "schafkopf.hintRequested")

	none, _ := setupSchafkopfWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.SchafkopfHint)(nil))
	assert.Contains(t, new(presenter.SchafkopfWebPresenter).HintOutput(none), "schafkopf.noHint")
}
