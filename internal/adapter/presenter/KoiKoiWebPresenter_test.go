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

func TestKoiKoiWebPresenter_Output(t *testing.T) {
	g := domain.NewDefaultKoiKoi()
	g.Reset()
	g.SetCurrentTurn(0)

	p := new(presenter.KoiKoiWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.KoiKoiPhasePlay), decoded["phase"])
	players, ok := decoded["players"].([]any)
	require.True(t, ok)
	assert.Len(t, players, domain.KoiKoiPlayerCnt)
	assert.Contains(t, decoded, "fieldCards")
	assert.Contains(t, decoded, "playableIndices")
	assert.Contains(t, decoded, "captureOptions")
}

// TestKoiKoiWebPresenter_ProceduralFaces は花札の札が deck:"hanafuda" + 絵文字グリフ +
// ラベルで手続き描画用にシリアライズされることを検証する (ADR-0033)。
func TestKoiKoiWebPresenter_ProceduralFaces(t *testing.T) {
	g := domain.NewDefaultKoiKoi()
	g.Reset()
	g.SetCurrentTurn(0)

	p := new(presenter.KoiKoiWebPresenter)
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
	// 人間の手札も花札の記述子を持つ。
	require.NotEmpty(t, parsed.Players[0].Cards)
	assert.Equal(t, "hanafuda", parsed.Players[0].Cards[0].Deck)
	assert.NotEmpty(t, parsed.Players[0].Cards[0].Glyph)
}

func TestKoiKoiWebPresenter_Error(t *testing.T) {
	g := domain.NewDefaultKoiKoi()
	g.Reset()
	p := new(presenter.KoiKoiWebPresenter)
	out := p.Output(g, errors.New("boom"))
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "boom", decoded["message"])
}

func TestKoiKoiWebPresenter_GameEnd(t *testing.T) {
	g := domain.NewDefaultKoiKoi()
	g.Reset()
	g.SetPhase(domain.KoiKoiPhaseGameEnd)

	p := new(presenter.KoiKoiWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "koikoi.result.scores", decoded["messageCode"])
}

// TestKoiKoiWebPresenter_Output_HintOnHumanTurn は通常の Output() が人間の手番 (プレイ
// フェーズ) でヒントを埋めることを検証する (#3516: フロントエンドのヒントトグルが機能する
// 前提条件)。
func TestKoiKoiWebPresenter_Output_HintOnHumanTurn(t *testing.T) {
	g := domain.NewDefaultKoiKoi()
	g.Reset()
	g.SetCurrentTurn(0) // player 0 は人間

	p := new(presenter.KoiKoiWebPresenter)
	var decoded struct {
		Hint *struct {
			CardIndex  int    `json:"cardIndex"`
			FieldIndex int    `json:"fieldIndex"`
			KoiKoi     int    `json:"koikoi"`
			Reason     string `json:"reason"`
		} `json:"hint"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &decoded))
	require.NotNil(t, decoded.Hint, "human play turn must include a hint")
	assert.NotEmpty(t, decoded.Hint.Reason)
}

// TestKoiKoiWebPresenter_Output_NoHintOnCpuTurn は CPU の手番では Output() がヒントを
// 出力しない (omitempty で省かれる) ことを検証する (#3516)。
func TestKoiKoiWebPresenter_Output_NoHintOnCpuTurn(t *testing.T) {
	g := domain.NewDefaultKoiKoi()
	g.Reset()
	g.SetCurrentTurn(1) // player 1 は CPU

	p := new(presenter.KoiKoiWebPresenter)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &decoded))
	assert.NotContains(t, decoded, "hint", "cpu turn must not include a hint")
}

func TestKoiKoiWebPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultKoiKoi()
	g.Reset()
	g.SetCurrentTurn(0)

	p := new(presenter.KoiKoiWebPresenter)
	out := p.HintOutput(g)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, decoded, "hint")
}

func TestKoiKoiWebPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultKoiKoi()
	g.Reset()
	p := new(presenter.KoiKoiWebPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// TestKoiKoiWebPresenter_RoundResult は shobu で確定した勝者ラウンドが lastRoundResult
// として出力されることを検証する (buildBase の LastRoundResult ブランチ)。
func TestKoiKoiWebPresenter_RoundResult(t *testing.T) {
	g := domain.NewDefaultKoiKoi()
	g.Reset()
	g.SetCurrentTurn(0)
	g.SetPhase(domain.KoiKoiPhaseKoiKoiDecision)
	// 三光 (松/桜/桐の光 = 5 点)。
	g.GetPlayer(0).AddCaptured([]*domain.Card{
		domain.NewCard(1, 1, false), domain.NewCard(3, 1, false), domain.NewCard(12, 1, false),
	})
	require.NoError(t, g.PlayerDecide(false)) // shobu → RoundEnd

	p := new(presenter.KoiKoiWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, decoded, "lastRoundResult")
	assert.NotNil(t, decoded["lastRoundResult"])
}

// TestKoiKoiWebPresenter_DecisionPhase は決断フェーズで pendingYaku が出力に含まれることを
// 検証する。
func TestKoiKoiWebPresenter_DecisionPhase(t *testing.T) {
	g := domain.NewDefaultKoiKoi()
	g.Reset()
	g.SetCurrentTurn(0)
	g.SetPhase(domain.KoiKoiPhaseKoiKoiDecision)

	p := new(presenter.KoiKoiWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, decoded, "pendingYaku")
	assert.Equal(t, float64(domain.KoiKoiPhaseKoiKoiDecision), decoded["phase"])
}
