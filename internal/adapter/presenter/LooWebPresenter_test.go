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

func TestLooWebPresenter_Output(t *testing.T) {
	g := domain.NewDefaultLoo()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.LooPhasePlay)
	g.SetLeadPlayerIdx(0)
	g.SetCurrentTurn(0)
	g.SetTrickNumber(1)
	g.GetPlayer(0).SetPlaying(true)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: bcard(domain.CardDesignHeart, 9)}})

	p := new(presenter.LooWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.LooPhasePlay), decoded["phase"])
	assert.Equal(t, float64(domain.LooTrickCount), decoded["totalTricks"])
	assert.Equal(t, float64(domain.CardDesignHeart), decoded["trumpSuit"])
	players, ok := decoded["players"].([]any)
	require.True(t, ok)
	assert.Len(t, players, domain.LooPlayerCnt)
	assert.Contains(t, decoded, "currentTrick")
	assert.Contains(t, decoded, "pot")
}

func TestLooWebPresenter_Error(t *testing.T) {
	g := domain.NewDefaultLoo()
	g.Reset()
	p := new(presenter.LooWebPresenter)
	out := p.Output(g, errors.New("boom"))
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "boom", decoded["message"])
}

func TestLooWebPresenter_RoundEnd(t *testing.T) {
	g := domain.NewDefaultLoo()
	g.Reset()
	g.GetPlayer(0).SetPlaying(true)
	g.GetPlayer(1).SetPlaying(true)
	g.SetRoundTricks([domain.LooPlayerCnt]int{domain.LooTrickCount, 0, 0, 0})
	g.SetPhase(domain.LooPhaseRoundEnd)
	g.ScoreRound()

	p := new(presenter.LooWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "loo.result.chips", decoded["messageCode"])
	assert.Contains(t, decoded, "lastDealDetail")
}

func TestLooWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.LooWebPresenter)

	t.Run("decide hint", func(t *testing.T) {
		g := domain.NewDefaultLoo()
		g.Reset()
		g.SetDecidePlayerIdx(0)
		out := p.HintOutput(g)
		var decoded map[string]any
		require.NoError(t, json.Unmarshal([]byte(out), &decoded))
		assert.Contains(t, decoded, "hint")
	})

	t.Run("play hint", func(t *testing.T) {
		g := domain.NewDefaultLoo()
		g.Reset()
		g.SetTrumpSuit(domain.CardDesignHeart)
		g.SetPhase(domain.LooPhasePlay)
		g.SetCurrentTurn(0)
		g.SetLeadPlayerIdx(0)
		g.GetPlayer(0).SetPlaying(true)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(bcard(domain.CardDesignHeart, 1))
		g.GetPlayer(0).AddCard(bcard(domain.CardDesignSpade, 2))
		assert.NotEmpty(t, p.HintOutput(g))
	})
}

func TestLooWebPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultLoo()
	g.Reset()
	p := new(presenter.LooWebPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

func TestLooWebPresenter_TurnUp(t *testing.T) {
	g := domain.NewDefaultLoo()
	g.Reset()
	p := new(presenter.LooWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	// turn-up should be present after reset.
	assert.Contains(t, decoded, "turnUp")
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestLooWebPresenterOutputCarriesTheHint(t *testing.T) {
	// 既存の HintOutput テストと同じ状態。300 回試して nil 0 件で確認済み。
	g := domain.NewDefaultLoo()
	g.Reset()
	g.SetDecidePlayerIdx(0)
	require.NotNil(t, g.GetHint(), "fixture must actually produce a hint")

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(new(presenter.LooWebPresenter).Output(g, nil)), &decoded))
	assert.Contains(t, decoded, "hint", "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotEqual(t, "loo.hintRequested", decoded["messageCode"])
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestLooWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g := domain.NewDefaultLoo()
	g.Reset()
	g.SetDecidePlayerIdx(0)
	assert.Contains(t, new(presenter.LooWebPresenter).HintOutput(g), "loo.hintRequested")
}

// **ヒントが無いときの分岐も見る。**Output() の受動ヒントは nil のとき
// `hint` キーごと落ちる。HintOutput() は noHint を返す。codecov が
// PR #4591 でこの 2 本を未到達として報告した。
func TestLooWebPresenterWithoutAHint(t *testing.T) {
	g := domain.NewDefaultLoo()
	g.Reset()
	g.SetDecidePlayerIdx(1) // 人間は席 0。他人の判断待ちなので助言することがない
	require.Nil(t, g.GetHint(), "fixture must actually produce no hint")

	p := new(presenter.LooWebPresenter)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &decoded))
	assert.NotContains(t, decoded, "hint")

	assert.Contains(t, p.HintOutput(g), "loo.noHint")
}
