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

func setupSkatWebMock() *interfaces.MockSkatGame {
	m := new(interfaces.MockSkatGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(0)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetForehandIdx").Return(1)
	m.On("GetMiddlehandIdx").Return(2)
	m.On("GetRearhandIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(-1)
	m.On("GetCurrentBid").Return(0)
	m.On("GetActiveBidActorIdx").Return(2)
	m.On("GetGameType").Return(domain.SkatGameNone)
	m.On("GetTrumpSuit").Return(0)
	m.On("PickedSkat").Return(false)
	m.On("GetDeclarerCardPoints").Return(0)
	m.On("GetDefendersCardPoints").Return(0)
	m.On("GetWinnerSide").Return(domain.SkatWinnerUndecided)
	m.On("GetGameValue").Return(0)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetLeadPlayerIdx").Return(-1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetSkat").Return(([]*domain.Card)(nil))
	m.On("GetOriginalSkat").Return(([]*domain.Card)(nil))
	m.On("GetConfig").Return(domain.DefaultSkatConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetPlayerCnt").Return(3)
	for i := 0; i < 3; i++ {
		m.On("GetPlayer", i).Return(domain.NewSkatPlayer(i == 0))
	}
	m.On("GetPhase").Return(domain.SkatPhaseBid)
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func TestSkatWebPresenter_OutputBidPhase(t *testing.T) {
	p := new(presenter.SkatWebPresenter)
	m := setupSkatWebMock()
	body := p.Output(m, nil)
	var resObj controller.SkatWebOutput
	if err := json.Unmarshal([]byte(body), &resObj); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, body)
	}
	assert.Equal(t, int(domain.SkatPhaseBid), resObj.Phase)
	assert.Equal(t, 3, len(resObj.Players))
	assert.Equal(t, "skat.bidPhase", resObj.MessageCode)
	assert.False(t, resObj.GameEndFlag)
	assert.Equal(t, 2, resObj.ActiveBidActorIdx)
}

func TestSkatWebPresenter_OutputErrorMessage(t *testing.T) {
	p := new(presenter.SkatWebPresenter)
	m := setupSkatWebMock()
	body := p.Output(m, errors.New("oops"))
	var resObj controller.SkatWebOutput
	_ = json.Unmarshal([]byte(body), &resObj)
	assert.Equal(t, "oops", resObj.Message)
}

func TestSkatWebPresenter_OutputRoundEndExposesSkat(t *testing.T) {
	p := new(presenter.SkatWebPresenter)
	m := new(interfaces.MockSkatGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(10)
	m.On("GetCurrentPlayerIdx").Return(-1)
	m.On("GetForehandIdx").Return(0)
	m.On("GetMiddlehandIdx").Return(1)
	m.On("GetRearhandIdx").Return(2)
	m.On("GetDealerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetCurrentBid").Return(18)
	m.On("GetActiveBidActorIdx").Return(-1)
	m.On("GetGameType").Return(domain.SkatGameSuit)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("PickedSkat").Return(true)
	m.On("GetDeclarerCardPoints").Return(70)
	m.On("GetDefendersCardPoints").Return(50)
	m.On("GetWinnerSide").Return(domain.SkatWinnerDeclarer)
	m.On("GetGameValue").Return(22)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	skat := []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 8, false),
	}
	m.On("GetSkat").Return(skat)
	m.On("GetOriginalSkat").Return(skat)
	m.On("GetConfig").Return(domain.DefaultSkatConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetPlayerCnt").Return(3)
	for i := 0; i < 3; i++ {
		m.On("GetPlayer", i).Return(domain.NewSkatPlayer(i == 0))
	}
	m.On("GetPhase").Return(domain.SkatPhaseRoundEnd)

	m.On("GetHint").Return(nil).Maybe()
	body := p.Output(m, nil)
	var resObj controller.SkatWebOutput
	_ = json.Unmarshal([]byte(body), &resObj)
	assert.Len(t, resObj.OriginalSkat, 2, "skat should be revealed at round end")
	assert.Equal(t, 22, resObj.GameValue)
	assert.Equal(t, "skat.roundEnd", resObj.MessageCode)
}

func TestSkatWebPresenter_HintAndActionLog(t *testing.T) {
	p := new(presenter.SkatWebPresenter)
	m := setupSkatWebMock()
	val := 1
	hint := &domain.SkatHint{Bid: &val, Reason: "strategic_bid"}
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
	m.On("GetHint").Return(hint)
	body := p.HintOutput(m)
	var resObj controller.SkatWebOutput
	_ = json.Unmarshal([]byte(body), &resObj)
	assert.NotNil(t, resObj.Hint)
	assert.Equal(t, "strategic_bid", resObj.Hint.Reason)
	assert.NotPanics(t, func() {
		_ = p.ActionLogOutput(m)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
//
// トリックテイキング系は Output 側にゲートを置きません。Skat.GetHint() が
// 「人間の手番で、かつ行動を選べる状態か」を自分で確かめて nil を返します。
func TestSkatWebPresenterOutputCarriesTheHint(t *testing.T) {
	val := 1
	skg := setupSkatWebMock()
	skg.ExpectedCalls = removeMockCall(skg.ExpectedCalls, "GetHint")
	skg.On("GetHint").Return(&domain.SkatHint{Bid: &val, Reason: "strategic_bid"})

	result := new(presenter.SkatWebPresenter).Output(skg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}
