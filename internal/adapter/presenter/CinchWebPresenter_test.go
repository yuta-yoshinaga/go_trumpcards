package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestCinchWebPresenter_Output(t *testing.T) {
	g := domain.NewDefaultCinch()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetBidWinnerIdx(0)
	g.SetCurrentBid(3)
	g.SetPhase(domain.CinchPhasePlay)
	g.SetLeadPlayerIdx(0)
	g.SetCurrentTurn(0)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*domain.CinchTrickCard{{PlayerIdx: 3, Card: bcard(domain.CardDesignHeart, 9)}})

	p := new(presenter.CinchWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.CinchPhasePlay), decoded["phase"])
	assert.Equal(t, float64(domain.CinchTotalTricks), decoded["totalTricks"])
	assert.Equal(t, float64(domain.CardDesignHeart), decoded["trumpSuit"])
	players, ok := decoded["players"].([]any)
	require.True(t, ok)
	assert.Len(t, players, domain.CinchPlayerCnt)
	assert.Contains(t, decoded, "currentTrick")
}

func TestCinchWebPresenter_Error(t *testing.T) {
	g := domain.NewDefaultCinch()
	g.Reset()
	p := new(presenter.CinchWebPresenter)
	out := p.Output(g, errors.New("boom"))
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "boom", decoded["message"])
}

func TestCinchWebPresenter_GameEnd(t *testing.T) {
	g := domain.NewDefaultCinch()
	g.Reset()
	// 人為的にゲーム終了状態を作る。
	g.GetPlayer(0).AddScore(30)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetBidWinnerIdx(0)
	g.SetCurrentBid(1)
	g.SetPhase(domain.CinchPhaseRoundEnd)
	g.ScoreRound()
	require.True(t, g.GetGameEndFlag())

	p := new(presenter.CinchWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, true, decoded["gameEndFlag"])
	assert.Equal(t, "cinch.result.scores", decoded["messageCode"])
	assert.NotEmpty(t, decoded["roundWinners"])
	assert.NotNil(t, decoded["lastDealDetail"])
}

func TestCinchWebPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultCinch()
	g.Reset() // bid フェーズ, human 手番
	p := new(presenter.CinchWebPresenter)
	out := p.HintOutput(g)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, decoded, "hint")
}

func TestCinchWebPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultCinch()
	g.Reset()
	p := new(presenter.CinchWebPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
