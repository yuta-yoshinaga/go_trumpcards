//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestChinesePokerWebPresenter_Output(t *testing.T) {
	pp := &ChinesePokerWebPresenter{}
	cp := domain.NewDefaultChinesePoker()
	_ = cp.Bet(100)

	result := pp.Output(cp, nil)
	var output controller.ChinesePokerWebOutput
	err := json.Unmarshal([]byte(result), &output)
	require.NoError(t, err)
	assert.Equal(t, domain.ChinesePokerPhaseSetHands, output.Phase)
	assert.Equal(t, 100, output.Bet)
	assert.Len(t, output.PlayerCards, 13)
}

func TestChinesePokerWebPresenter_Output_WithError(t *testing.T) {
	pp := &ChinesePokerWebPresenter{}
	cp := domain.NewDefaultChinesePoker()
	testErr := domain.NewDomainError(domain.ErrWrongPhase, "test error")

	result := pp.Output(cp, testErr)
	var output controller.ChinesePokerWebOutput
	err := json.Unmarshal([]byte(result), &output)
	require.NoError(t, err)
	assert.Contains(t, output.Message, "test error")
}

func TestChinesePokerWebPresenter_Output_EndPhaseWin(t *testing.T) {
	pp := &ChinesePokerWebPresenter{}
	cp := domain.NewDefaultChinesePoker()
	cp.SetGameEndFlag(true)
	cp.SetResult(domain.GameResultWin)
	cp.SetScoop(false)

	result := pp.Output(cp, nil)
	var output controller.ChinesePokerWebOutput
	err := json.Unmarshal([]byte(result), &output)
	require.NoError(t, err)
	assert.Equal(t, "chinesepoker.result.playerWins", output.MessageCode)
}

func TestChinesePokerWebPresenter_Output_EndPhaseScoopWin(t *testing.T) {
	pp := &ChinesePokerWebPresenter{}
	cp := domain.NewDefaultChinesePoker()
	cp.SetGameEndFlag(true)
	cp.SetResult(domain.GameResultWin)
	cp.SetScoop(true)

	result := pp.Output(cp, nil)
	var output controller.ChinesePokerWebOutput
	err := json.Unmarshal([]byte(result), &output)
	require.NoError(t, err)
	assert.Equal(t, "chinesepoker.result.playerScoop", output.MessageCode)
}

func TestChinesePokerWebPresenter_Output_EndPhaseLose(t *testing.T) {
	pp := &ChinesePokerWebPresenter{}
	cp := domain.NewDefaultChinesePoker()
	cp.SetGameEndFlag(true)
	cp.SetResult(domain.GameResultLose)
	cp.SetScoop(false)

	result := pp.Output(cp, nil)
	var output controller.ChinesePokerWebOutput
	err := json.Unmarshal([]byte(result), &output)
	require.NoError(t, err)
	assert.Equal(t, "chinesepoker.result.dealerWins", output.MessageCode)
}

func TestChinesePokerWebPresenter_Output_EndPhaseScoopLose(t *testing.T) {
	pp := &ChinesePokerWebPresenter{}
	cp := domain.NewDefaultChinesePoker()
	cp.SetGameEndFlag(true)
	cp.SetResult(domain.GameResultLose)
	cp.SetScoop(true)

	result := pp.Output(cp, nil)
	var output controller.ChinesePokerWebOutput
	err := json.Unmarshal([]byte(result), &output)
	require.NoError(t, err)
	assert.Equal(t, "chinesepoker.result.dealerScoop", output.MessageCode)
}

func TestChinesePokerWebPresenter_ActionLogOutput(t *testing.T) {
	pp := &ChinesePokerWebPresenter{}
	cp := domain.NewDefaultChinesePoker()
	result := pp.ActionLogOutput(cp)
	assert.Contains(t, result, "[")
}

// #5615: CUI は #4717 から `GetSuggestedArrangement()` の具体的な分け方 (どの札を
// どの列へ) とファウル警告を出しているのに、Web の HintOutput は Output() を
// そのまま返すだけで、ページはフロント独自のランク降順スライスを使っていた。
// 同じ盤面で違う答えが出る状態だった。
func TestChinesePokerWebPresenterCarriesTheSuggestedArrangement(t *testing.T) {
	pp := &ChinesePokerWebPresenter{}
	cp := domain.NewDefaultChinesePoker()
	require.NoError(t, cp.Bet(100))

	var out controller.ChinesePokerWebOutput
	require.NoError(t, json.Unmarshal([]byte(pp.HintOutput(cp)), &out))

	want := cp.GetSuggestedArrangement()
	require.NotNil(t, want, "セットハンドで13枚そろっていれば必ず案が出る")
	require.NotNil(t, out.SuggestedArrangement)
	// **ドメインの答えをそのまま運ぶ。**presenter で並べ直すと、CUI と Web が
	// 同じ手札で違う分け方を勧めることになる。
	assert.Equal(t, want.Front, out.SuggestedArrangement.Front)
	assert.Equal(t, want.Middle, out.SuggestedArrangement.Middle)
	assert.Equal(t, want.Back, out.SuggestedArrangement.Back)
	assert.Equal(t, want.Foul, out.SuggestedArrangement.Foul)
	// 3/5/5 の形も固定しておく (取り違えると列がずれる)。
	assert.Len(t, out.SuggestedArrangement.Front, 3)
	assert.Len(t, out.SuggestedArrangement.Middle, 5)
	assert.Len(t, out.SuggestedArrangement.Back, 5)
}

// セットハンドでない (= 13枚そろっていない) ときは載せない。
// 空の配列を返すと、フロントには「前列に置く札が無い」と読める。
func TestChinesePokerWebPresenterOmitsTheArrangementOutsideSetHands(t *testing.T) {
	pp := &ChinesePokerWebPresenter{}
	cp := domain.NewDefaultChinesePoker()

	var out controller.ChinesePokerWebOutput
	require.NoError(t, json.Unmarshal([]byte(pp.HintOutput(cp)), &out))
	assert.Nil(t, out.SuggestedArrangement)
}
