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

func setupSchnapsenWebMock(trumpCard *domain.Card) *interfaces.MockSchnapsenGame {
	m := new(interfaces.MockSchnapsenGame)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SchnapsenPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTrumpCard").Return(trumpCard)
	m.On("GetDealerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(1)
	m.On("GetStockRemaining").Return(9)
	m.On("IsEndgame").Return(false)
	m.On("GetValidPlayIndices", 0).Return([]int{0})
	m.On("GetMarriageIndices", 0).Return([]int(nil))
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultSchnapsenConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupSchnapsenWebMockWithPlayers(trumpCard *domain.Card) (*interfaces.MockSchnapsenGame, []*domain.SchnapsenPlayer) {
	m := setupSchnapsenWebMock(trumpCard)
	players := []*domain.SchnapsenPlayer{
		domain.NewSchnapsenPlayer(true),
		domain.NewSchnapsenPlayer(false),
	}
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayerPoints", 0).Return(18)
	m.On("GetPlayerPoints", 1).Return(5)
	return m, players
}

func TestSchnapsenWebPresenter_Output_InitialState(t *testing.T) {
	p := new(presenter.SchnapsenWebPresenter)
	trump := domain.NewCard(domain.CardDesignSpade, 13, false)
	m, players := setupSchnapsenWebMockWithPlayers(trump)
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	got := p.Output(m, nil)
	assert.NotEmpty(t, got)

	var out controller.SchnapsenWebOutput
	assert.NoError(t, json.Unmarshal([]byte(got), &out))
	assert.Equal(t, 2, len(out.Players))
	assert.False(t, out.GameEndFlag)
	assert.Equal(t, 0, out.Phase)
	assert.Equal(t, domain.CardDesignSpade, out.TrumpSuit)
	assert.NotNil(t, out.TrumpCard)
	assert.Equal(t, 9, out.StockRemaining)
	assert.False(t, out.IsEndgame)
	assert.Equal(t, []int{0}, out.ValidPlays)
	assert.Equal(t, 18, out.Players[0].Points)
}

func TestSchnapsenWebPresenter_Output_HumanCardsShownCPUHidden(t *testing.T) {
	p := new(presenter.SchnapsenWebPresenter)
	trump := domain.NewCard(domain.CardDesignSpade, 13, false)
	m, players := setupSchnapsenWebMockWithPlayers(trump)
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	got := p.Output(m, nil)
	var out controller.SchnapsenWebOutput
	_ = json.Unmarshal([]byte(got), &out)

	assert.True(t, out.Players[0].IsHuman)
	assert.Len(t, out.Players[0].Cards, 1)
	for _, c := range out.Players[1].Cards {
		assert.Empty(t, c.Design, "CPU card design should be hidden")
	}
}

func TestSchnapsenWebPresenter_Output_GameEnd(t *testing.T) {
	p := new(presenter.SchnapsenWebPresenter)
	m := setupSchnapsenWebMock(nil)
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(domain.NewSchnapsenPlayer(true))
	m.On("GetPlayer", 1).Return(domain.NewSchnapsenPlayer(false))
	m.On("GetPlayerPoints", 0).Return(70)
	m.On("GetPlayerPoints", 1).Return(40)
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerIdx").Return(0)

	got := p.Output(m, nil)
	var out controller.SchnapsenWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.True(t, out.GameEndFlag)
	assert.Equal(t, 0, out.WinnerIdx)
	assert.Equal(t, "schnapsen.result.p0Win", out.MessageCode)
}

func TestSchnapsenWebPresenter_Output_Error(t *testing.T) {
	p := new(presenter.SchnapsenWebPresenter)
	m, _ := setupSchnapsenWebMockWithPlayers(nil)
	got := p.Output(m, errors.New("boom"))
	var out controller.SchnapsenWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.Equal(t, "boom", out.Message)
}

func TestSchnapsenWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.SchnapsenWebPresenter)
	trump := domain.NewCard(domain.CardDesignSpade, 13, false)
	m, players := setupSchnapsenWebMockWithPlayers(trump)
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	idx := 0
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
	m.On("GetHint").Return(&domain.SchnapsenHint{CardIndex: &idx, Reason: "lead_low"})

	got := p.HintOutput(m)
	var out controller.SchnapsenWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.NotNil(t, out.Hint)
	assert.Equal(t, 0, *out.Hint.CardIndex)
	assert.Equal(t, "lead_low", out.Hint.Reason)
}

func TestSchnapsenWebPresenter_HintOutput_None(t *testing.T) {
	p := new(presenter.SchnapsenWebPresenter)
	m, _ := setupSchnapsenWebMockWithPlayers(nil)
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
	m.On("GetHint").Return((*domain.SchnapsenHint)(nil))
	got := p.HintOutput(m)
	var out controller.SchnapsenWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.Nil(t, out.Hint)
}

func TestSchnapsenWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SchnapsenWebPresenter)
	m, _ := setupSchnapsenWebMockWithPlayers(nil)
	got := p.ActionLogOutput(m)
	assert.NotEmpty(t, got)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
//
// トリックテイキング系は Output 側にゲートを置きません。Schnapsen.GetHint() が
// 「人間の手番で、かつ行動を選べる状態か」を自分で確かめて nil を返します。
func TestSchnapsenWebPresenterOutputCarriesTheHint(t *testing.T) {
	trump := domain.NewCard(domain.CardDesignSpade, 13, false)
	idx := 0
	sng, players := setupSchnapsenWebMockWithPlayers(trump)
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	sng.ExpectedCalls = removeMockCall(sng.ExpectedCalls, "GetHint")
	sng.On("GetHint").Return(&domain.SchnapsenHint{CardIndex: &idx, Reason: "lead_low"})

	result := new(presenter.SchnapsenWebPresenter).Output(sng, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}
