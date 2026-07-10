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

func newHachiHachiRoundEndGame(t *testing.T) *domain.HachiHachi {
	t.Helper()
	g := domain.NewDefaultHachiHachi()
	cfg := domain.DefaultHachiHachiConfig()
	cfg.TargetRounds = 3
	g.SetConfig(cfg)
	g.Reset()
	for step := 0; step < 20000 && g.GetPhase() == domain.HachiHachiPhasePlay; step++ {
		if g.IsHumanTurn() {
			require.NoError(t, g.PlayerPlay(0, -1))
		} else {
			g.CpuPlay()
		}
	}
	require.Equal(t, domain.HachiHachiPhaseRoundEnd, g.GetPhase())
	return g
}

func TestHachiHachiWebPresenter_Output(t *testing.T) {
	g := domain.NewDefaultHachiHachi()
	g.Reset()
	g.SetCurrentTurn(0)

	p := new(presenter.HachiHachiWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.HachiHachiPhasePlay), decoded["phase"])
	players, ok := decoded["players"].([]any)
	require.True(t, ok)
	assert.Len(t, players, domain.HachiHachiPlayerCnt)
	assert.Contains(t, decoded, "fieldCards")
	assert.Contains(t, decoded, "playableIndices")
	assert.Contains(t, decoded, "captureOptions")
}

// TestHachiHachiWebPresenter_ProceduralFaces は花札の札が deck:"hanafuda" + 絵文字グリフ +
// ラベルで手続き描画用にシリアライズされることを検証する (ADR-0033)。
func TestHachiHachiWebPresenter_ProceduralFaces(t *testing.T) {
	g := domain.NewDefaultHachiHachi()
	g.Reset()
	g.SetCurrentTurn(0)

	p := new(presenter.HachiHachiWebPresenter)
	out := p.Output(g, nil)

	type webCard struct {
		Deck  string `json:"deck"`
		Glyph string `json:"glyph"`
		Label string `json:"label"`
		Color string `json:"color"`
	}
	type webPlayer struct {
		IsHuman bool       `json:"isHuman"`
		Cards   []*webCard `json:"cards"`
	}
	var parsed struct {
		Players    []*webPlayer `json:"players"`
		FieldCards []*webCard   `json:"fieldCards"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &parsed))

	require.NotEmpty(t, parsed.FieldCards)
	for _, c := range parsed.FieldCards {
		assert.Equal(t, "hanafuda", c.Deck)
		assert.NotEmpty(t, c.Glyph)
		assert.NotEmpty(t, c.Label)
		assert.NotEmpty(t, c.Color)
	}
	require.NotEmpty(t, parsed.Players[0].Cards)
	assert.Equal(t, "hanafuda", parsed.Players[0].Cards[0].Deck)
	assert.NotEmpty(t, parsed.Players[0].Cards[0].Glyph)
}

func TestHachiHachiWebPresenter_Error(t *testing.T) {
	g := domain.NewDefaultHachiHachi()
	g.Reset()
	p := new(presenter.HachiHachiWebPresenter)
	out := p.Output(g, errors.New("boom"))
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "boom", decoded["message"])
}

func TestHachiHachiWebPresenter_GameEnd(t *testing.T) {
	g := domain.NewDefaultHachiHachi()
	g.Reset()
	g.SetPhase(domain.HachiHachiPhaseGameEnd)

	p := new(presenter.HachiHachiWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "hachihachi.result.scores", decoded["messageCode"])
}

func TestHachiHachiWebPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultHachiHachi()
	g.Reset()
	g.SetCurrentTurn(0)

	p := new(presenter.HachiHachiWebPresenter)
	out := p.HintOutput(g)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, decoded, "hint")
}

func TestHachiHachiWebPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultHachiHachi()
	g.Reset()
	p := new(presenter.HachiHachiWebPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// TestHachiHachiWebPresenter_RoundResult はラウンド精算後の lastRoundResult が
// 3 人分の内訳付きで出力されることを検証する。
func TestHachiHachiWebPresenter_RoundResult(t *testing.T) {
	g := newHachiHachiRoundEndGame(t)
	p := new(presenter.HachiHachiWebPresenter)
	out := p.Output(g, nil)
	var parsed struct {
		LastRoundResult *struct {
			Scores []struct {
				PlayerIdx int `json:"playerIdx"`
				Delta     int `json:"delta"`
			} `json:"scores"`
			Best int `json:"best"`
		} `json:"lastRoundResult"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &parsed))
	require.NotNil(t, parsed.LastRoundResult)
	assert.Len(t, parsed.LastRoundResult.Scores, domain.HachiHachiPlayerCnt)
}
