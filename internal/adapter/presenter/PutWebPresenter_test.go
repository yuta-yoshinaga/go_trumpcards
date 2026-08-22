//go:build test && (!js || !wasm || extra4)

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

func newPutGame() *domain.Put {
	g := domain.NewDefaultPut()
	g.Reset()
	return g
}

func unmarshalPutOut(t *testing.T, s string) controller.PutWebOutput {
	t.Helper()
	var out controller.PutWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return out
}

func TestPutWebPresenter_Output_InitialState(t *testing.T) {
	p := new(presenter.PutWebPresenter)
	g := newPutGame()
	out := unmarshalPutOut(t, p.Output(g, nil))

	assert.Equal(t, 2, len(out.Players))
	assert.False(t, out.GameEndFlag)
	assert.Equal(t, domain.PutDefaultMatchTarget, out.MatchTarget)
	assert.Len(t, out.MatchPoints, 2)
	assert.Equal(t, -1, out.WinnerIdx)
	assert.Equal(t, 1, out.HandStake)
}

func TestPutWebPresenter_Output_HumanCardsShownCPUHidden(t *testing.T) {
	p := new(presenter.PutWebPresenter)
	g := newPutGame()
	out := unmarshalPutOut(t, p.Output(g, nil))

	human := out.Players[0]
	assert.True(t, human.IsHuman)
	assert.Len(t, human.Cards, human.CardCount)

	cpu := out.Players[1]
	assert.False(t, cpu.IsHuman)
	for _, c := range cpu.Cards {
		assert.Empty(t, c.Design, "CPU card design should be hidden")
	}
}

func TestPutWebPresenter_Output_GameEnd(t *testing.T) {
	p := new(presenter.PutWebPresenter)
	g := newPutGame()
	g.SetGameEndFlag(true)
	g.SetPlayerMatchPoints(0, 15)
	// winnerIdx is set via the engine; emulate p0 win by reaching target through advanceHand
	g.SetPhase(domain.PutPhaseGameEnd)

	out := unmarshalPutOut(t, p.Output(g, nil))
	assert.True(t, out.GameEndFlag)
}

func TestPutWebPresenter_Output_RespondPhase(t *testing.T) {
	p := new(presenter.PutWebPresenter)
	g := newPutGame()
	g.SetPhase(domain.PutPhaseRespond)
	g.SetPendingLevel(domain.PutLevelPut)
	g.SetPutCallerIdx(1)
	g.SetResponderIdx(0)

	out := unmarshalPutOut(t, p.Output(g, nil))
	assert.Equal(t, "put.respondPhase", out.MessageCode)
	assert.Equal(t, int(domain.PutPhaseRespond), out.Phase)
}

func TestPutWebPresenter_Output_HandEnd(t *testing.T) {
	p := new(presenter.PutWebPresenter)
	g := newPutGame()
	g.SetPhase(domain.PutPhaseHandEnd)
	g.SetHandWinnerIdx(0)
	g.SetHandStake(2)

	out := unmarshalPutOut(t, p.Output(g, nil))
	assert.Equal(t, "put.handEnd.p0", out.MessageCode)
}

func TestPutWebPresenter_Output_Error(t *testing.T) {
	p := new(presenter.PutWebPresenter)
	g := newPutGame()
	out := unmarshalPutOut(t, p.Output(g, errors.New("boom")))
	assert.Equal(t, "boom", out.Message)
}

func TestPutWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.PutWebPresenter)
	g := newPutGame()
	g.SetPhase(domain.PutPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)

	out := unmarshalPutOut(t, p.HintOutput(g))
	assert.NotNil(t, out.Hint)
}

func TestPutWebPresenter_HintOutput_None(t *testing.T) {
	p := new(presenter.PutWebPresenter)
	g := newPutGame()
	g.SetPhase(domain.PutPhaseTrickEnd)
	out := unmarshalPutOut(t, p.HintOutput(g))
	assert.Nil(t, out.Hint)
}

func TestPutWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PutWebPresenter)
	g := newPutGame()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
//
// Output 側にゲートは置きません。Put.GetHint() が「人間の手番で、かつ
// 行動を選べる状態か」を自分で確かめて nil を返します。
func TestPutWebPresenterOutputCarriesTheHint(t *testing.T) {
	g := newPutGame()
	g.SetPhase(domain.PutPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	if g.GetHint() == nil {
		t.Fatal("fixture must actually produce a hint")
	}

	result := new(presenter.PutWebPresenter).Output(g, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}
