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

func TestBasraWebPresenter_Output(t *testing.T) {
	g := domain.NewDefaultBasra()
	g.Reset()
	g.SetCurrentTurn(0)

	p := new(presenter.BasraWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.BasraPhasePlay), decoded["phase"])
	players, ok := decoded["players"].([]any)
	require.True(t, ok)
	assert.Len(t, players, domain.BasraPlayerCnt)
	assert.Contains(t, decoded, "tableCards")
	assert.Contains(t, decoded, "playableIndices")
	assert.Contains(t, decoded, "captureOptions")
}

func TestBasraWebPresenter_Error(t *testing.T) {
	g := domain.NewDefaultBasra()
	g.Reset()
	p := new(presenter.BasraWebPresenter)
	out := p.Output(g, errors.New("boom"))
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "boom", decoded["message"])
}

func TestBasraWebPresenter_GameEnd(t *testing.T) {
	g := domain.NewDefaultBasra()
	g.Reset()
	g.SetPhase(domain.BasraPhaseGameEnd)

	p := new(presenter.BasraWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "basra.result.scores", decoded["messageCode"])
	assert.Contains(t, decoded, "lastDealDetail")
}

func TestBasraWebPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultBasra()
	g.Reset()
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))

	p := new(presenter.BasraWebPresenter)
	out := p.HintOutput(g)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, decoded, "hint")
}

func TestBasraWebPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultBasra()
	g.Reset()
	p := new(presenter.BasraWebPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestBasraWebPresenterOutputCarriesTheHint(t *testing.T) {
	// Reset 直後にヒントが出る。300 回試して nil 0 件で確認済み。
	g := domain.NewDefaultBasra()
	g.Reset()
	require.NotNil(t, g.GetHint(), "fixture must actually produce a hint")

	out := new(presenter.BasraWebPresenter).Output(g, nil)
	assert.Contains(t, out, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, out, "basra.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestBasraWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g := domain.NewDefaultBasra()
	g.Reset()
	assert.Contains(t, new(presenter.BasraWebPresenter).HintOutput(g), "basra.hintRequested")
}
