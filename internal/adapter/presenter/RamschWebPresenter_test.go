//go:build test

package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func ramschDecode(t *testing.T, raw string) controller.RamschWebOutput {
	t.Helper()
	var out controller.RamschWebOutput
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return out
}

func TestRamschWebPresenter_Output_BaseState(t *testing.T) {
	p := new(presenter.RamschWebPresenter)
	g := domain.NewDefaultRamsch()
	g.Reset()

	out := ramschDecode(t, p.Output(g, nil))
	assert.Equal(t, int(domain.RamschPhasePlay), out.Phase)
	assert.Len(t, out.Players, domain.RamschPlayerCnt)
	assert.Equal(t, 1, out.RoundNumber)
	assert.Equal(t, 1, out.TrickNumber)
	assert.Equal(t, -1, out.LoserIdx, "ラウンド途中では敗者は決まっていない")
	assert.False(t, out.Durchmarsch)
	assert.Equal(t, -1, out.DurchmarschIdx)

	// **人間の手札だけが見える。** 相手の手札が漏れると、このゲームの
	// 「誰が点を抱えているか読む」部分が消える。
	for _, pl := range out.Players {
		if pl.IsHuman {
			assert.NotEmpty(t, pl.Cards, "自分の手札が見えない")
			continue
		}
		assert.Empty(t, pl.Cards, "player %d の手札が漏れている", pl.ID)
	}
}

// **伏せ札はラウンドが終わるまで返さない。** 最終トリックの獲得者が受け取る
// 2 枚なので、中身が見えると終盤の判断が完全情報になる。
func TestRamschWebPresenter_Output_HidesTheSkatUntilTheRoundEnds(t *testing.T) {
	p := new(presenter.RamschWebPresenter)
	g := domain.NewDefaultRamsch()
	g.Reset()

	assert.Empty(t, ramschDecode(t, p.Output(g, nil)).Skat, "プレイ中に伏せ札が見えている")

	g.SetPhase(domain.RamschPhaseRoundEnd)
	assert.Len(t, ramschDecode(t, p.Output(g, nil)).Skat, domain.RamschSkatSize,
		"ラウンド終了後も伏せ札が出ていない")
}

// 罰点の結果（敗者 / Durchmarsch）がページまで届くこと。
func TestRamschWebPresenter_Output_CarriesTheRoundOutcome(t *testing.T) {
	p := new(presenter.RamschWebPresenter)

	g := domain.NewDefaultRamsch()
	g.Reset()
	g.SetPhase(domain.RamschPhaseRoundEnd)
	g.SetCardPointsForTest([domain.RamschPlayerCnt]int{30, 78, 12})
	g.ScoreRound()

	out := ramschDecode(t, p.Output(g, nil))
	assert.Equal(t, 1, out.LoserIdx)
	assert.False(t, out.Durchmarsch)
	assert.Equal(t, 78, out.Players[1].CardPoints, "集めた点がページに届いていない")
	assert.Equal(t, -78, out.Players[1].RoundScore)

	g2 := domain.NewDefaultRamsch()
	g2.Reset()
	g2.SetPhase(domain.RamschPhaseRoundEnd)
	g2.SetCardPointsForTest([domain.RamschPlayerCnt]int{domain.RamschTotalCardPoints, 0, 0})
	g2.SetDurchmarschForTest(0)
	g2.ScoreRound()

	out2 := ramschDecode(t, p.Output(g2, nil))
	assert.True(t, out2.Durchmarsch)
	assert.Equal(t, 0, out2.DurchmarschIdx)
	assert.Equal(t, -1, out2.LoserIdx, "Durchmarsch では単独の敗者は決まらない")
}

// **受動ヒントは Output にも載る。** hint コマンド専用のレスポンスはページの
// state にマージされないので、ここで載せないと state.hint が常に undefined。
func TestRamschWebPresenter_Output_CarriesThePassiveHint(t *testing.T) {
	p := new(presenter.RamschWebPresenter)
	g := domain.NewDefaultRamsch()
	g.Reset()
	require.True(t, g.IsHumanTurn())

	out := ramschDecode(t, p.Output(g, nil))
	require.NotNil(t, out.Hint, "手番なのにヒントが載っていない")
	require.NotNil(t, out.Hint.CardIndex)
	assert.NotEmpty(t, out.Hint.Reason)

	// 手番でなければ載せない。
	g.SetCurrentPlayerIdx(1)
	assert.Nil(t, ramschDecode(t, p.Output(g, nil)).Hint)
}

func TestRamschWebPresenter_Output_ShowsTheError(t *testing.T) {
	p := new(presenter.RamschWebPresenter)
	g := domain.NewDefaultRamsch()
	g.Reset()
	out := ramschDecode(t, p.Output(g, domain.ErrWrongPhase))
	assert.Contains(t, out.Message, domain.ErrWrongPhase.Error())
}
