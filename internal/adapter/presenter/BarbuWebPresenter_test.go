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

func bcard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

func TestBarbuWebPresenter_HintOutput(t *testing.T) {
	b := domain.NewDefaultBarbu()
	b.Reset()
	p := new(presenter.BarbuWebPresenter)
	// HintOutput mirrors Output (the GUI computes its own hint client-side).
	var out controller.BarbuWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.HintOutput(b)), &out))
	assert.Equal(t, domain.BarbuPhaseSelectContract, out.Phase)
}

func TestBarbuWebPresenter_Output(t *testing.T) {
	b := domain.NewDefaultBarbu()
	b.Reset()
	p := new(presenter.BarbuWebPresenter)

	var out controller.BarbuWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(b, nil)), &out))
	assert.Equal(t, domain.BarbuPhaseSelectContract, out.Phase)
	assert.Equal(t, domain.BarbuTotalDeals, out.TotalDeals)
	assert.Len(t, out.Players, domain.BarbuPlayerCnt)
	assert.Len(t, out.UsedContracts, domain.BarbuContractCnt)
	// human (player 0) cards are visible; CPUs hidden
	assert.NotEmpty(t, out.Players[0].Cards)
	assert.Empty(t, out.Players[1].Cards)
}

func TestBarbuWebPresenter_ErrorMessage(t *testing.T) {
	b := domain.NewDefaultBarbu()
	b.Reset()
	p := new(presenter.BarbuWebPresenter)
	var out controller.BarbuWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(b, errors.New("boom"))), &out))
	assert.Equal(t, "boom", out.Message)
}

func TestBarbuWebPresenter_DominoPlayable(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetContract(domain.BarbuContractDominoes, -1)
	b.BarbuTestSetPhase(domain.BarbuPhasePlay)
	b.BarbuTestSetCurrentPlayer(0) // human
	b.BarbuTestSetHand(0, []*domain.Card{bcard(domain.CardDesignSpade, 7), bcard(domain.CardDesignSpade, 2)})
	p := new(presenter.BarbuWebPresenter)
	var out controller.BarbuWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(b, nil)), &out))
	assert.Equal(t, []int{0}, out.DominoPlayable) // only the 7
	assert.Len(t, out.TablePlaced, 5)
}

func TestBarbuWebPresenter_GameEnd(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetGameEnd(true)
	b.BarbuTestSetContract(domain.BarbuContractTrumps, domain.CardDesignSpade)
	b.GetPlayer(0).AddScore(20)
	p := new(presenter.BarbuWebPresenter)
	var out controller.BarbuWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(b, nil)), &out))
	assert.True(t, out.GameEndFlag)
	assert.Equal(t, "barbu.result.scores", out.MessageCode)
	assert.Contains(t, out.MessageParams["scores"], "0:20")
	assert.NotEmpty(t, out.RoundWinners)
}

func TestBarbuWebPresenter_DealHistory(t *testing.T) {
	b := domain.NewDefaultBarbu()
	b.Reset()
	// Record one completed deal.
	b.BarbuTestSetContract(domain.BarbuContractNoTricks, -1)
	b.BarbuTestAddTrick(0, []*domain.Card{bcard(domain.CardDesignSpade, 5)})
	b.BarbuTestFinishDeal()

	p := new(presenter.BarbuWebPresenter)
	var out controller.BarbuWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(b, nil)), &out))
	require.Len(t, out.DealHistory, 1)
	assert.Equal(t, domain.BarbuContractNoTricks, out.DealHistory[0].Contract)
	assert.Equal(t, 0, out.DealHistory[0].DealerIdx)
}

func TestBarbuWebPresenter_ActionLog(t *testing.T) {
	b := domain.NewDefaultBarbu()
	b.Reset()
	p := new(presenter.BarbuWebPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(b))
}

// **フォロー義務の可視化 (#4804)。**ドミノ以外の 6 契約では、Web に出せる札の
// 情報が無かった。
func TestBarbuWebPresenter_PlayableIndices(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetContract(domain.BarbuContractNoTricks, -1)
	b.BarbuTestSetPhase(domain.BarbuPhasePlay)
	b.BarbuTestSetCurrentPlayer(0) // human
	b.BarbuTestSetHand(0, []*domain.Card{
		bcard(domain.CardDesignSpade, 7),
		bcard(domain.CardDesignHeart, 2),
		bcard(domain.CardDesignSpade, 3),
	})
	p := new(presenter.BarbuWebPresenter)

	// リード時は全部出せる。
	var lead controller.BarbuWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(b, nil)), &lead))
	assert.Equal(t, []int{0, 1, 2}, lead.PlayableIndices)
	assert.Empty(t, lead.DominoPlayable, "契約が違うのでドミノ用は空のまま")

	// ♠ リードが場にあるなら ♠ だけ。
	b.BarbuTestSetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: bcard(domain.CardDesignSpade, 13)}})
	var follow controller.BarbuWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(b, nil)), &follow))
	assert.Equal(t, []int{0, 2}, follow.PlayableIndices)
}
