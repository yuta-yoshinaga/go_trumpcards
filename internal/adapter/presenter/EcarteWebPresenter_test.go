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

func setupEcarteWebMock(trumpCard *domain.Card) *interfaces.MockEcarteGame {
	m := new(interfaces.MockEcarteGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.EcartePhasePlay)
	m.On("GetNegStep").Return(domain.EcarteNegElderDecide)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTrumpCard").Return(trumpCard)
	m.On("GetDealerIdx").Return(0)
	m.On("GetElderIdx").Return(1)
	m.On("GetLeadPlayerIdx").Return(1)
	m.On("GetStockRemaining").Return(21)
	m.On("IsRefusalByDealer").Return(false)
	m.On("GetValidPlayIndices", 0).Return([]int{0})
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultEcarteConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupEcarteWebMockWithPlayers(trumpCard *domain.Card) (*interfaces.MockEcarteGame, []*domain.EcartePlayer) {
	m := setupEcarteWebMock(trumpCard)
	players := []*domain.EcartePlayer{
		domain.NewEcartePlayer(true),
		domain.NewEcartePlayer(false),
	}
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetDealPoints", 0).Return(2)
	m.On("GetDealPoints", 1).Return(1)
	m.On("GetMatchScore", 0).Return(4)
	m.On("GetMatchScore", 1).Return(3)
	return m, players
}

func TestEcarteWebPresenter_Output_InitialState(t *testing.T) {
	p := new(presenter.EcarteWebPresenter)
	trump := domain.NewCard(domain.CardDesignSpade, 13, false)
	m, players := setupEcarteWebMockWithPlayers(trump)
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	players[0].SetRoundScore(2)
	players[0].SetCumulativeScore(4)

	got := p.Output(m, nil)
	assert.NotEmpty(t, got)

	var out controller.EcarteWebOutput
	assert.NoError(t, json.Unmarshal([]byte(got), &out))
	assert.Equal(t, 2, len(out.Players))
	assert.False(t, out.GameEndFlag)
	assert.Equal(t, 1, out.Phase)
	assert.Equal(t, 0, out.NegStep)
	assert.Equal(t, 1, out.RoundNumber)
	assert.Equal(t, domain.CardDesignSpade, out.TrumpSuit)
	assert.NotNil(t, out.TrumpCard)
	assert.Equal(t, 21, out.StockRemaining)
	assert.False(t, out.RefusalByDealer)
	assert.Equal(t, 0, out.DealerIdx)
	assert.Equal(t, 1, out.ElderIdx)
	assert.Equal(t, []int{2, 1}, out.DealPoints)
	assert.Equal(t, []int{4, 3}, out.MatchScore)
	assert.Equal(t, []int{0}, out.ValidPlays)
	assert.Equal(t, 2, out.Players[0].RoundScore)
	assert.Equal(t, 4, out.Players[0].CumulativeScore)
}

func TestEcarteWebPresenter_Output_HumanCardsShownCPUHidden(t *testing.T) {
	p := new(presenter.EcarteWebPresenter)
	trump := domain.NewCard(domain.CardDesignSpade, 13, false)
	m, players := setupEcarteWebMockWithPlayers(trump)
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	got := p.Output(m, nil)
	var out controller.EcarteWebOutput
	_ = json.Unmarshal([]byte(got), &out)

	assert.True(t, out.Players[0].IsHuman)
	assert.Len(t, out.Players[0].Cards, 1)
	for _, c := range out.Players[1].Cards {
		assert.Empty(t, c.Design, "CPU card design should be hidden")
	}
}

func TestEcarteWebPresenter_Output_ExchangePhaseMessage(t *testing.T) {
	p := new(presenter.EcarteWebPresenter)
	m, _ := setupEcarteWebMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
	m.On("GetPhase").Return(domain.EcartePhaseExchange)
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetNegStep")
	m.On("GetNegStep").Return(domain.EcarteNegDealerRespond)

	got := p.Output(m, nil)
	var out controller.EcarteWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.Equal(t, 0, out.Phase)
	assert.Equal(t, 1, out.NegStep)
	assert.Equal(t, "ecarte.exchange.dealerRespond", out.MessageCode)
	// ValidPlays is empty outside the play phase.
	assert.Empty(t, out.ValidPlays)
}

func TestEcarteWebPresenter_Output_GameEnd(t *testing.T) {
	p := new(presenter.EcarteWebPresenter)
	m := setupEcarteWebMock(nil)
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(domain.NewEcartePlayer(true))
	m.On("GetPlayer", 1).Return(domain.NewEcartePlayer(false))
	m.On("GetDealPoints", 0).Return(0)
	m.On("GetDealPoints", 1).Return(0)
	m.On("GetMatchScore", 0).Return(5)
	m.On("GetMatchScore", 1).Return(3)
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerIdx").Return(0)

	got := p.Output(m, nil)
	var out controller.EcarteWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.True(t, out.GameEndFlag)
	assert.Equal(t, 0, out.WinnerIdx)
	assert.Equal(t, "ecarte.result.p0Win", out.MessageCode)
}

func TestEcarteWebPresenter_Output_Error(t *testing.T) {
	p := new(presenter.EcarteWebPresenter)
	m, _ := setupEcarteWebMockWithPlayers(nil)
	got := p.Output(m, errors.New("boom"))
	var out controller.EcarteWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.Equal(t, "boom", out.Message)
}

func TestEcarteWebPresenter_HintOutput_Card(t *testing.T) {
	p := new(presenter.EcarteWebPresenter)
	trump := domain.NewCard(domain.CardDesignSpade, 13, false)
	m, players := setupEcarteWebMockWithPlayers(trump)
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	idx := 0
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
	m.On("GetHint").Return(&domain.EcarteHint{CardIndex: &idx, Reason: "lead_high"})

	got := p.HintOutput(m)
	var out controller.EcarteWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.NotNil(t, out.Hint)
	assert.Equal(t, 0, *out.Hint.CardIndex)
	assert.Equal(t, "lead_high", out.Hint.Reason)
}

func TestEcarteWebPresenter_HintOutput_Action(t *testing.T) {
	p := new(presenter.EcarteWebPresenter)
	m, _ := setupEcarteWebMockWithPlayers(nil)
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
	m.On("GetHint").Return(&domain.EcarteHint{Action: "propose", Reason: "weak_hand"})

	got := p.HintOutput(m)
	var out controller.EcarteWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.NotNil(t, out.Hint)
	assert.Equal(t, "propose", out.Hint.Action)
	assert.Nil(t, out.Hint.CardIndex)
}

func TestEcarteWebPresenter_HintOutput_None(t *testing.T) {
	p := new(presenter.EcarteWebPresenter)
	m, _ := setupEcarteWebMockWithPlayers(nil)
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
	m.On("GetHint").Return((*domain.EcarteHint)(nil))
	got := p.HintOutput(m)
	var out controller.EcarteWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.Nil(t, out.Hint)
}

func TestEcarteWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.EcarteWebPresenter)
	m, _ := setupEcarteWebMockWithPlayers(nil)
	got := p.ActionLogOutput(m)
	assert.NotEmpty(t, got)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestEcarteWebPresenterOutputCarriesTheHint(t *testing.T) {
	idx := 0
	ecg, _ := setupEcarteWebMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
	ecg.ExpectedCalls = removeMockCall(ecg.ExpectedCalls, "GetHint")
	ecg.On("GetHint").Return(&domain.EcarteHint{CardIndex: &idx, Action: "play", Reason: "lead_low"})

	result := new(presenter.EcarteWebPresenter).Output(ecg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "ecarte.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**このゲーム群の
// hintAvailable は画面のラベルとして埋まっているので別キーを使う (#4483)。
func TestEcarteWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	idx := 0
	ecg, _ := setupEcarteWebMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
	ecg.ExpectedCalls = removeMockCall(ecg.ExpectedCalls, "GetHint")
	ecg.On("GetHint").Return(&domain.EcarteHint{CardIndex: &idx, Action: "play", Reason: "lead_low"})
	assert.Contains(t, new(presenter.EcarteWebPresenter).HintOutput(ecg), "ecarte.hintRequested")

	none, _ := setupEcarteWebMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.EcarteHint)(nil))
	assert.Contains(t, new(presenter.EcarteWebPresenter).HintOutput(none), "ecarte.noHint")
}
