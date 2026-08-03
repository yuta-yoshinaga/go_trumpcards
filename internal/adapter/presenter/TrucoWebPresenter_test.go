//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTrucoGame() *domain.Truco {
	g := domain.NewDefaultTruco()
	g.Reset()
	return g
}

func unmarshalTrucoOut(t *testing.T, s string) controller.TrucoWebOutput {
	t.Helper()
	var out controller.TrucoWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return out
}

func TestTrucoWebPresenter_Output_InitialState(t *testing.T) {
	p := new(presenter.TrucoWebPresenter)
	g := newTrucoGame()
	out := unmarshalTrucoOut(t, p.Output(g, nil))

	assert.Equal(t, 2, len(out.Players))
	assert.False(t, out.GameEndFlag)
	assert.Equal(t, domain.TrucoDefaultMatchTarget, out.MatchTarget)
	assert.Len(t, out.MatchPoints, 2)
	assert.Equal(t, -1, out.WinnerIdx)
	assert.Equal(t, 1, out.HandStake)
}

func TestTrucoWebPresenter_Output_HumanCardsShownCPUHidden(t *testing.T) {
	p := new(presenter.TrucoWebPresenter)
	g := newTrucoGame()
	out := unmarshalTrucoOut(t, p.Output(g, nil))

	human := out.Players[0]
	assert.True(t, human.IsHuman)
	assert.Len(t, human.Cards, human.CardCount)

	cpu := out.Players[1]
	assert.False(t, cpu.IsHuman)
	for _, c := range cpu.Cards {
		assert.Empty(t, c.Design, "CPU card design should be hidden")
	}
}

func TestTrucoWebPresenter_Output_GameEnd(t *testing.T) {
	p := new(presenter.TrucoWebPresenter)
	g := newTrucoGame()
	g.SetGameEndFlag(true)
	g.SetPlayerMatchPoints(0, 15)
	// winnerIdx is set via the engine; emulate p0 win by reaching target through advanceHand
	g.SetPhase(domain.TrucoPhaseGameEnd)

	out := unmarshalTrucoOut(t, p.Output(g, nil))
	assert.True(t, out.GameEndFlag)
}

func TestTrucoWebPresenter_Output_RespondPhase(t *testing.T) {
	p := new(presenter.TrucoWebPresenter)
	g := newTrucoGame()
	g.SetPhase(domain.TrucoPhaseRespond)
	g.SetPendingLevel(domain.TrucoLevelTruco)
	g.SetTrucoCallerIdx(1)
	g.SetResponderIdx(0)

	out := unmarshalTrucoOut(t, p.Output(g, nil))
	assert.Equal(t, "truco.respondPhase", out.MessageCode)
	assert.Equal(t, int(domain.TrucoPhaseRespond), out.Phase)
}

func TestTrucoWebPresenter_Output_HandEnd(t *testing.T) {
	p := new(presenter.TrucoWebPresenter)
	g := newTrucoGame()
	g.SetPhase(domain.TrucoPhaseHandEnd)
	g.SetHandWinnerIdx(0)
	g.SetHandStake(2)

	out := unmarshalTrucoOut(t, p.Output(g, nil))
	assert.Equal(t, "truco.handEnd.p0", out.MessageCode)
}

func TestTrucoWebPresenter_Output_Error(t *testing.T) {
	p := new(presenter.TrucoWebPresenter)
	g := newTrucoGame()
	out := unmarshalTrucoOut(t, p.Output(g, errors.New("boom")))
	assert.Equal(t, "boom", out.Message)
}

func TestTrucoWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.TrucoWebPresenter)
	g := newTrucoGame()
	g.SetPhase(domain.TrucoPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)

	out := unmarshalTrucoOut(t, p.HintOutput(g))
	assert.NotNil(t, out.Hint)
}

func TestTrucoWebPresenter_HintOutput_None(t *testing.T) {
	p := new(presenter.TrucoWebPresenter)
	g := newTrucoGame()
	g.SetPhase(domain.TrucoPhaseTrickEnd)
	out := unmarshalTrucoOut(t, p.HintOutput(g))
	assert.Nil(t, out.Hint)
}

func TestTrucoWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TrucoWebPresenter)
	g := newTrucoGame()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
//
// Output 側にゲートは置きません。Truco.GetHint() が「人間の手番で、かつ
// 行動を選べる状態か」を自分で確かめて nil を返します。
func TestTrucoWebPresenterOutputCarriesTheHint(t *testing.T) {
	g := newTrucoGame()
	g.SetPhase(domain.TrucoPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	if g.GetHint() == nil {
		t.Fatal("fixture must actually produce a hint")
	}

	result := new(presenter.TrucoWebPresenter).Output(g, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}
