package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestDoudizhuCuiPresenter_Output_BidPhase(t *testing.T) {
	dg := newDoudizhuForPresenter()
	dg.Reset()
	// Put the human on turn so the bid prompt renders.
	humanIdx := 0
	for i := 0; i < dg.GetPlayerCnt(); i++ {
		if dg.GetPlayer(i).GetIsHuman() {
			humanIdx = i
			break
		}
	}
	dg.SetCurrentTurn(humanIdx)

	p := new(presenter.DoudizhuCuiPresenter)
	out := p.Output(dg, nil)
	assert.NotEmpty(t, out)
	// The current bidder line and the human bid prompt should both appear.
	bidderPrefix := strings.SplitN(i18n.T("doudizhu.currentBidder"), "{{", 2)[0]
	assert.Contains(t, out, bidderPrefix)
	assert.Contains(t, out, i18n.T("doudizhu.promptBid"))
}

func TestDoudizhuCuiPresenter_Output_PlayPhase(t *testing.T) {
	dg := newDoudizhuForPresenter()
	dg.SetPhase(domain.DoudizhuPhasePlay)
	dg.SetLandlordIdx(0)
	dg.SetCurrentTurn(0)
	dg.SetKittyCards([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
	dg.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	dg.SetTableCombo(&domain.DoudizhuCombo{Type: domain.DoudizhuComboSingle, Rank: 10, Length: 1, Cards: []*domain.Card{domain.NewCard(domain.CardDesignHeart, 10, false)}})

	p := new(presenter.DoudizhuCuiPresenter)
	out := p.Output(dg, nil)
	assert.NotEmpty(t, out)
}

func TestDoudizhuCuiPresenter_Output_EndPhase(t *testing.T) {
	dg := newDoudizhuForPresenter()
	dg.SetPhase(domain.DoudizhuPhasePlay)
	dg.SetLandlordIdx(0)
	dg.SetBaseBid(1)
	dg.SetCurrentTurn(0)
	dg.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	dg.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	dg.GetPlayer(2).AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
	_ = dg.PlayerPlay([]int{0})

	p := new(presenter.DoudizhuCuiPresenter)
	out := p.Output(dg, nil)
	assert.NotEmpty(t, out)
}

func TestDoudizhuCuiPresenter_Output_Error(t *testing.T) {
	dg := newDoudizhuForPresenter()
	dg.Reset()

	p := new(presenter.DoudizhuCuiPresenter)
	out := p.Output(dg, errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestDoudizhuCuiPresenter_ActionLogOutput(t *testing.T) {
	dg := newDoudizhuForPresenter()
	dg.Reset()

	p := new(presenter.DoudizhuCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(dg))
}
