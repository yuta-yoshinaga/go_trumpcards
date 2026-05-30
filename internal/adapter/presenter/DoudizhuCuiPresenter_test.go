package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDoudizhuCuiPresenter_Output_BidPhase(t *testing.T) {
	dg := newDoudizhuForPresenter()
	dg.Reset()

	p := new(presenter.DoudizhuCuiPresenter)
	out := p.Output(dg, nil)
	assert.NotEmpty(t, out)
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
