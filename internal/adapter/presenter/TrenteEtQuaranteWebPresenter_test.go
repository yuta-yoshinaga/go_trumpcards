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

func TestTrenteEtQuaranteWebPresenter_Output(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()

	p := new(presenter.TrenteEtQuaranteWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.TrenteEtQuarantePhaseBet), decoded["phase"])
	assert.Contains(t, decoded, "noirRow")
	assert.Contains(t, decoded, "rougeRow")
	assert.Contains(t, decoded, "chips")
	assert.Contains(t, decoded, "config")
}

func TestTrenteEtQuaranteWebPresenter_Error(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	p := new(presenter.TrenteEtQuaranteWebPresenter)
	out := p.Output(g, errors.New("boom"))
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "boom", decoded["message"])
}

func TestTrenteEtQuaranteWebPresenter_Result(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	require.NoError(t, g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 100))

	p := new(presenter.TrenteEtQuaranteWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.TrenteEtQuarantePhaseResult), decoded["phase"])
	assert.NotEmpty(t, decoded["messageCode"])
	assert.Contains(t, decoded, "winningRow")
	assert.Contains(t, decoded, "payout")
	assert.Contains(t, decoded, "currentBet")
}

func TestTrenteEtQuaranteWebPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()

	p := new(presenter.TrenteEtQuaranteWebPresenter)
	out := p.HintOutput(g)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, decoded, "hint")
}

func TestTrenteEtQuaranteWebPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	require.NoError(t, g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 100))
	p := new(presenter.TrenteEtQuaranteWebPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
//
// Output 側にゲートは置きません。TrenteEtQuarante.GetHint() が「人間の手番で、かつ
// 行動を選べる状態か」を自分で確かめて nil を返します。
func TestTrenteEtQuaranteWebPresenterOutputCarriesTheHint(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	if g.GetHint() == nil {
		t.Fatal("fixture must actually produce a hint")
	}

	result := new(presenter.TrenteEtQuaranteWebPresenter).Output(g, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}
