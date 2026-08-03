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

func setupCallBreakWebMock() *interfaces.MockCallBreakGame {
	m := new(interfaces.MockCallBreakGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetSpadesBroken").Return(false)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.CallBreakPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultCallBreakConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupCallBreakWebMockWithPlayers() (*interfaces.MockCallBreakGame, []*domain.CallBreakPlayer) {
	m := setupCallBreakWebMock()
	players := makeCallBreakPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	m.On("GetValidPlayIndices", 0).Return([]int{})
	return m, players
}

func TestCallBreakWebPresenter_Output(t *testing.T) {
	p := new(presenter.CallBreakWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupCallBreakWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.CallBreakWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 1, resObj.Phase) // CallBreakPhasePlay
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.False(t, resObj.SpadesBroken)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Equal(t, "callbreak.playPhase.lead", resObj.MessageCode)
		assert.Equal(t, domain.CallBreakDefaultMaxRounds, resObj.Config.MaxRounds)
	})

	t.Run("error in last action", func(t *testing.T) {
		m, _ := setupCallBreakWebMockWithPlayers()
		result := p.Output(m, errors.New("bad play"))
		var resObj controller.CallBreakWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "bad play", resObj.Message)
	})

	t.Run("bid phase message code", func(t *testing.T) {
		m, _ := setupCallBreakWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CallBreakPhaseBid)
		var resObj controller.CallBreakWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &resObj))
		assert.Equal(t, "callbreak.bidPhase", resObj.MessageCode)
	})

	t.Run("follow phase shows follow code", func(t *testing.T) {
		m, _ := setupCallBreakWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 9, false)},
		})
		var resObj controller.CallBreakWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &resObj))
		assert.Equal(t, "callbreak.playPhase.follow", resObj.MessageCode)
		assert.Len(t, resObj.CurrentTrick, 1)
	})

	t.Run("trick end and round end codes", func(t *testing.T) {
		for phase, code := range map[domain.CallBreakPhase]string{
			domain.CallBreakPhaseTrickEnd: "callbreak.trickEnd",
			domain.CallBreakPhaseRoundEnd: "callbreak.roundEnd",
		} {
			m, _ := setupCallBreakWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.On("GetPhase").Return(phase)
			var resObj controller.CallBreakWebOutput
			assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &resObj))
			assert.Equal(t, code, resObj.MessageCode)
		}
	})

	t.Run("game ended human winner", func(t *testing.T) {
		m, _ := setupCallBreakWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		var resObj controller.CallBreakWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "callbreak.result.humanWin", resObj.MessageCode)
	})

	t.Run("game ended cpu winner", func(t *testing.T) {
		m, _ := setupCallBreakWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(2)
		var resObj controller.CallBreakWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &resObj))
		assert.Equal(t, "callbreak.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, "2", resObj.MessageParams["cpuId"])
	})

	t.Run("valid play indices reflect human player", func(t *testing.T) {
		m, _ := setupCallBreakWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetValidPlayIndices")
		m.On("GetValidPlayIndices", 0).Return([]int{0, 2, 4})
		var resObj controller.CallBreakWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &resObj))
		assert.Equal(t, []int{0, 2, 4}, resObj.ValidPlayIndices)
	})

	t.Run("valid play indices default to empty slice when nil", func(t *testing.T) {
		m, _ := setupCallBreakWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetValidPlayIndices")
		m.On("GetValidPlayIndices", 0).Return(([]int)(nil))
		var resObj controller.CallBreakWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &resObj))
		assert.NotNil(t, resObj.ValidPlayIndices)
		assert.Equal(t, 0, len(resObj.ValidPlayIndices))
	})
}

func TestCallBreakWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.CallBreakWebPresenter)

	t.Run("nil hint", func(t *testing.T) {
		m, _ := setupCallBreakWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.CallBreakHint)(nil))
		var resObj controller.CallBreakWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.HintOutput(m)), &resObj))
		assert.Nil(t, resObj.Hint)
	})

	t.Run("bid hint", func(t *testing.T) {
		m, _ := setupCallBreakWebMockWithPlayers()
		bid := 4
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.CallBreakHint{Bid: &bid, Reason: "strategic_bid"})
		var resObj controller.CallBreakWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.HintOutput(m)), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.NotNil(t, resObj.Hint.Bid)
		assert.Equal(t, 4, *resObj.Hint.Bid)
		assert.Equal(t, "strategic_bid", resObj.Hint.Reason)
	})
}

func TestCallBreakWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.CallBreakWebPresenter)
	m := new(interfaces.MockCallBreakGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "x"},
	})
	out := p.ActionLogOutput(m)
	assert.Contains(t, out, "\"actionType\":\"play\"")
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestCallBreakWebPresenterOutputCarriesTheHint(t *testing.T) {
	idx := 0
	cbg, _ := setupCallBreakWebMockWithPlayers()
	cbg.ExpectedCalls = removeMockCall(cbg.ExpectedCalls, "GetHint")
	cbg.On("GetHint").Return(&domain.CallBreakHint{CardIndex: &idx, Reason: "follow_suit"})

	result := new(presenter.CallBreakWebPresenter).Output(cbg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "callbreak.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestCallBreakWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	idx := 0
	cbg, _ := setupCallBreakWebMockWithPlayers()
	cbg.ExpectedCalls = removeMockCall(cbg.ExpectedCalls, "GetHint")
	cbg.On("GetHint").Return(&domain.CallBreakHint{CardIndex: &idx, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.CallBreakWebPresenter).HintOutput(cbg), "callbreak.hintRequested")

	none, _ := setupCallBreakWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.CallBreakHint)(nil))
	assert.Contains(t, new(presenter.CallBreakWebPresenter).HintOutput(none), "callbreak.noHint")
}
