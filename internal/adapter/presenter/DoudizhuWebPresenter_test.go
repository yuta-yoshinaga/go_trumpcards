package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newDoudizhuForPresenter() *domain.Doudizhu {
	config := domain.DefaultDoudizhuConfig()
	players := []*domain.DoudizhuPlayer{
		domain.NewDoudizhuPlayer(true),
		domain.NewDoudizhuPlayer(false),
		domain.NewDoudizhuPlayer(false),
	}
	return domain.NewDoudizhu(domain.NewTrumpCards(domain.DoudizhuJokerCount), players, config)
}

func TestDoudizhuWebPresenter_Output_BidPhase(t *testing.T) {
	dg := newDoudizhuForPresenter()
	dg.Reset()

	p := new(presenter.DoudizhuWebPresenter)
	out := p.Output(dg, nil)

	var resp controller.DoudizhuWebOutput
	require.NoError(t, json.Unmarshal([]byte(out), &resp))
	assert.Equal(t, "bid", resp.Phase)
	assert.Len(t, resp.Players, domain.DoudizhuPlayerCnt)
	assert.False(t, resp.GameEndFlag)
}

func TestDoudizhuWebPresenter_Output_PlayPhase(t *testing.T) {
	dg := newDoudizhuForPresenter()
	dg.SetPhase(domain.DoudizhuPhasePlay)
	dg.SetLandlordIdx(0)
	dg.SetBaseBid(2)
	dg.SetCurrentTurn(0)
	dg.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	dg.SetTableCombo(&domain.DoudizhuCombo{Type: domain.DoudizhuComboSingle, Rank: 10, Length: 1, Cards: []*domain.Card{domain.NewCard(domain.CardDesignHeart, 10, false)}})

	p := new(presenter.DoudizhuWebPresenter)
	out := p.Output(dg, nil)

	var resp controller.DoudizhuWebOutput
	require.NoError(t, json.Unmarshal([]byte(out), &resp))
	assert.Equal(t, "play", resp.Phase)
	assert.Equal(t, 0, resp.LandlordIdx)
	assert.Equal(t, 2, resp.BaseBid)
	assert.Equal(t, "single", resp.TableCombo)
	assert.Len(t, resp.TableCards, 1)
}

func TestDoudizhuWebPresenter_Output_Error(t *testing.T) {
	dg := newDoudizhuForPresenter()
	dg.Reset()

	p := new(presenter.DoudizhuWebPresenter)
	out := p.Output(dg, errors.New("boom"))

	var resp controller.DoudizhuWebOutput
	require.NoError(t, json.Unmarshal([]byte(out), &resp))
	assert.Equal(t, "boom", resp.Message)
}

func TestDoudizhuWebPresenter_Output_GameEnd(t *testing.T) {
	dg := newDoudizhuForPresenter()
	dg.SetPhase(domain.DoudizhuPhasePlay)
	dg.SetLandlordIdx(0)
	dg.SetBaseBid(1)
	dg.SetCurrentTurn(0)
	dg.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	dg.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	dg.GetPlayer(2).AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
	require.NoError(t, dg.PlayerPlay([]int{0}))

	p := new(presenter.DoudizhuWebPresenter)
	out := p.Output(dg, nil)

	var resp controller.DoudizhuWebOutput
	require.NoError(t, json.Unmarshal([]byte(out), &resp))
	assert.True(t, resp.GameEndFlag)
	assert.Equal(t, "end", resp.Phase)
	assert.NotEmpty(t, resp.Message)
}

func TestDoudizhuWebPresenter_ActionLogOutput(t *testing.T) {
	dg := newDoudizhuForPresenter()
	dg.Reset()

	p := new(presenter.DoudizhuWebPresenter)
	out := p.ActionLogOutput(dg)
	assert.NotEmpty(t, out)
}
