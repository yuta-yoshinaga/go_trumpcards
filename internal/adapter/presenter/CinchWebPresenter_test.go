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
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: bcard(domain.CardDesignHeart, 9)}})

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

// TestCinchWebPresenter_BidPhase は bid フェーズ (human 手番) 出力で playableIndices が
// 空 (play 以外) になることと playableIndices ブランチを網羅する。
func TestCinchWebPresenter_BidPhase(t *testing.T) {
	g := domain.NewDefaultCinch()
	g.Reset() // bid フェーズ、human 手番
	p := new(presenter.CinchWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.CinchPhaseBid), decoded["phase"])
	// bid フェーズでは playableIndices は空配列。
	pi, ok := decoded["playableIndices"].([]any)
	require.True(t, ok)
	assert.Empty(t, pi)
	assert.True(t, decoded["isHumanTurn"].(bool))
}

// TestCinchWebPresenter_HintOutput_TrumpAndCard は name-trump / play フェーズでの
// ヒント出力 (TrumpSuit / CardIndices) を網羅する。
func TestCinchWebPresenter_HintOutput_TrumpAndCard(t *testing.T) {
	p := new(presenter.CinchWebPresenter)

	// name-trump フェーズ、human が勝者。
	gt := domain.NewDefaultCinch()
	gt.Reset()
	gt.SetPhase(domain.CinchPhaseNameTrump)
	gt.SetBidWinnerIdx(0)
	var trumpHint map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.HintOutput(gt)), &trumpHint))
	assert.Contains(t, trumpHint, "hint")

	// play フェーズ、human 手番。
	gp := domain.NewDefaultCinch()
	gp.Reset()
	gp.SetPhase(domain.CinchPhasePlay)
	gp.SetTrumpSuit(domain.CardDesignHeart)
	gp.SetBidWinnerIdx(0)
	gp.SetCurrentTurn(0)
	gp.SetLeadPlayerIdx(0)
	gp.GetPlayer(0).Reset()
	gp.GetPlayer(0).AddCard(bcard(domain.CardDesignHeart, 1))
	gp.GetPlayer(0).AddCard(bcard(domain.CardDesignSpade, 2))
	var playHint map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.HintOutput(gp)), &playHint))
	assert.Contains(t, playHint, "hint")
}
