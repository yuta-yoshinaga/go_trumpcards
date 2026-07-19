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

func TestGoStopWebPresenter_Output(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	g.SetCurrentTurn(0)

	p := new(presenter.GoStopWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.GoStopPhasePlay), decoded["phase"])
	players, ok := decoded["players"].([]any)
	require.True(t, ok)
	assert.Len(t, players, domain.GoStopPlayerCnt)
	assert.Contains(t, decoded, "fieldCards")
	assert.Contains(t, decoded, "playableIndices")
	assert.Contains(t, decoded, "captureOptions")
}

// TestGoStopWebPresenter_ProceduralFaces は花札の札が deck:"hanafuda" + 絵文字グリフ +
// ラベルで手続き描画用にシリアライズされることを検証する (ADR-0033)。
func TestGoStopWebPresenter_ProceduralFaces(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	g.SetCurrentTurn(0)

	p := new(presenter.GoStopWebPresenter)
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

func TestGoStopWebPresenter_Error(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	p := new(presenter.GoStopWebPresenter)
	out := p.Output(g, errors.New("boom"))
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "boom", decoded["message"])
}

func TestGoStopWebPresenter_GameEnd(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	g.SetPhase(domain.GoStopPhaseGameEnd)

	p := new(presenter.GoStopWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "gostop.result.scores", decoded["messageCode"])
}

// humanCpuIdx は人間/CPU の座席インデックスを返すヘルパ。
func humanCpuIdx(g interface {
	GetPlayerCnt() int
	GetPlayer(int) *domain.GoStopPlayer
}) (human, cpu int) {
	human, cpu = -1, -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if g.GetPlayer(i).GetIsHuman() {
			human = i
		} else {
			cpu = i
		}
	}
	return human, cpu
}

// TestGoStopWebPresenter_OutputHint は通常 Output() が人間手番でヒントを含み、CPU 手番では
// 含まないことを検証する (#3519: 死んでいたヒントトグルを機能させる)。
func TestGoStopWebPresenter_OutputHint(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	human, cpu := humanCpuIdx(g)
	require.GreaterOrEqual(t, human, 0)
	require.GreaterOrEqual(t, cpu, 0)
	p := new(presenter.GoStopWebPresenter)

	// 人間手番のプレイフェーズ: hint が載る。
	g.SetCurrentTurn(human)
	g.SetPhase(domain.GoStopPhasePlay)
	var withHint struct {
		Hint *struct {
			Reason string `json:"reason"`
		} `json:"hint"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &withHint))
	require.NotNil(t, withHint.Hint)
	assert.NotEmpty(t, withHint.Hint.Reason)

	// CPU 手番: hint は載らない。
	g.SetCurrentTurn(cpu)
	var cpuTurn map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &cpuTurn))
	assert.NotContains(t, cpuTurn, "hint")

	// エラー時: hint は載らない。
	g.SetCurrentTurn(human)
	var errOut map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, errors.New("boom"))), &errOut))
	assert.NotContains(t, errOut, "hint")
}

func TestGoStopWebPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	g.SetCurrentTurn(0)

	p := new(presenter.GoStopWebPresenter)
	out := p.HintOutput(g)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, decoded, "hint")
}

func TestGoStopWebPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	p := new(presenter.GoStopWebPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// TestGoStopWebPresenter_RoundResult は stop で確定した勝者ラウンドが lastRoundResult
// として出力されることを検証する。
func TestGoStopWebPresenter_RoundResult(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	g.SetCurrentTurn(0)
	g.SetPhase(domain.GoStopPhaseGoDecision)
	g.GetPlayer(0).AddCaptured([]*domain.Card{
		domain.NewCard(1, 1, false), domain.NewCard(3, 1, false), domain.NewCard(12, 1, false),
	})
	require.NoError(t, g.PlayerDecide(false)) // stop → RoundEnd

	p := new(presenter.GoStopWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, decoded, "lastRoundResult")
	assert.NotNil(t, decoded["lastRoundResult"])
}

// TestGoStopWebPresenter_DecisionPhase は決断フェーズで pendingBreakdown/pendingPoints が
// 出力に含まれることを検証する。
func TestGoStopWebPresenter_DecisionPhase(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	g.SetCurrentTurn(0)
	g.SetPhase(domain.GoStopPhaseGoDecision)

	p := new(presenter.GoStopWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, decoded, "pendingBreakdown")
	assert.Equal(t, float64(domain.GoStopPhaseGoDecision), decoded["phase"])
}
