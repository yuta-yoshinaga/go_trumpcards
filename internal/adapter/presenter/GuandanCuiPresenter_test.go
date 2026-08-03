//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// **レベル札が A より強いことがこのゲームの肝。**画面に書いていないと読めない。
func TestGuandanCuiPresenter_ExplainsTheLevelCards(t *testing.T) {
	o := defaultGuandanOpts()
	o.level = 5
	out := new(presenter.GuandanCuiPresenter).Output(setupGuandanMock(o), nil)
	assert.Contains(t, out, "5のレベル札はAより強く、黒ジョーカーより弱い")
	assert.Contains(t, out, "♥の2枚はワイルド")
}

// **レベルは 2〜A。**数字のままでは J/Q/K/A が読めない。
func TestGuandanCuiPresenter_LabelsTheFaceLevels(t *testing.T) {
	for level, want := range map[int]string{5: "5", 11: "J", 12: "Q", 13: "K", domain.GuandanMaxLevel: "A"} {
		o := defaultGuandanOpts()
		o.level = level
		out := new(presenter.GuandanCuiPresenter).Output(setupGuandanMock(o), nil)
		assert.Contains(t, out, "この局のレベル: "+want)
	}
}

func TestGuandanCuiPresenter_HidesTheOtherHands(t *testing.T) {
	out := new(presenter.GuandanCuiPresenter).Output(setupGuandanMock(defaultGuandanOpts()), nil)
	// 人間の手札は添字つきで見える。
	assert.Contains(t, out, "0:SPADE 2")
	// **味方の手札も伏せる。**
	assert.Contains(t, out, "非公開 1枚")
	assert.Contains(t, out, "<- 手番")
}

func TestGuandanCuiPresenter_RevealsEveryHandAtGameEnd(t *testing.T) {
	o := defaultGuandanOpts()
	o.gameEnd = true
	o.winner = 0
	out := new(presenter.GuandanCuiPresenter).Output(setupGuandanMock(o), nil)
	assert.NotContains(t, out, "非公開")
	assert.Contains(t, out, "チーム0の勝ち")
}

func TestGuandanCuiPresenter_TableState(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		out := new(presenter.GuandanCuiPresenter).Output(setupGuandanMock(defaultGuandanOpts()), nil)
		assert.Contains(t, out, "流れています")
	})

	t.Run("a combination is named and sized", func(t *testing.T) {
		o := defaultGuandanOpts()
		o.combo = &domain.GuandanCombo{Kind: domain.GuandanComboTube, Rank: 5, Size: 6}
		out := new(presenter.GuandanCuiPresenter).Output(setupGuandanMock(o), nil)
		assert.Contains(t, out, "チューブ")
		assert.Contains(t, out, "(6枚)")
		assert.Contains(t, out, "席1が出しました")
	})

	// 役の名前は全種そろっていること (未知の役が "-" で出ると読めない)。
	t.Run("every combination has a label", func(t *testing.T) {
		kinds := []domain.GuandanComboKind{
			domain.GuandanComboSingle, domain.GuandanComboPair, domain.GuandanComboTriple,
			domain.GuandanComboFullHouse, domain.GuandanComboStraight, domain.GuandanComboPlate,
			domain.GuandanComboTube, domain.GuandanComboBomb, domain.GuandanComboStraightFlush,
			domain.GuandanComboJokerBomb,
		}
		for _, k := range kinds {
			o := defaultGuandanOpts()
			o.combo = &domain.GuandanCombo{Kind: k, Size: 1}
			out := new(presenter.GuandanCuiPresenter).Output(setupGuandanMock(o), nil)
			assert.NotContains(t, out, "場: - (")
		}
		o := defaultGuandanOpts()
		o.combo = &domain.GuandanCombo{Kind: domain.GuandanComboNone, Size: 1}
		out := new(presenter.GuandanCuiPresenter).Output(setupGuandanMock(o), nil)
		assert.Contains(t, out, "場: - (")
	})
}

func TestGuandanCuiPresenter_Tribute(t *testing.T) {
	t.Run("pending and returned are told apart", func(t *testing.T) {
		o := defaultGuandanOpts()
		o.phase = domain.GuandanPhaseTribute
		o.tributes = []*domain.GuandanTribute{
			{From: 3, To: 0, Card: gdTestCard(domain.CardDesignSpade, 1)},
			{From: 2, To: 1, Card: gdTestCard(domain.CardDesignHeart, 13), Returned: gdTestCard(domain.CardDesignClover, 2)},
			nil,
		}
		out := new(presenter.GuandanCuiPresenter).Output(setupGuandanMock(o), nil)
		assert.Contains(t, out, "還貢待ち")
		assert.Contains(t, out, "還貢: CLOVER 2")
		assert.Contains(t, out, "t <添字>")
	})

	// **赤ジョーカー 2 枚で貢は流れる。**理由が出ないと不可解に見える。
	t.Run("a cancelled tribute explains itself", func(t *testing.T) {
		o := defaultGuandanOpts()
		o.phase = domain.GuandanPhaseTribute
		o.cancelled = true
		out := new(presenter.GuandanCuiPresenter).Output(setupGuandanMock(o), nil)
		assert.Contains(t, out, "抗貢")
		assert.Contains(t, out, "赤ジョーカー2枚")
		assert.Contains(t, out, "n ・・・次の局へ")
	})
}

func TestGuandanCuiPresenter_HandEnd(t *testing.T) {
	// **1 着 2 着の独占は +4。**通常の +1 と区別して出す。
	t.Run("first and second is called out", func(t *testing.T) {
		o := defaultGuandanOpts()
		o.phase = domain.GuandanPhaseHandEnd
		o.level = 6
		o.lastResult = &domain.GuandanHandResult{
			WinnerTeam: 0, Advance: domain.GuandanAdvanceFirstSecond, FirstSecond: true,
		}
		out := new(presenter.GuandanCuiPresenter).Output(setupGuandanMock(o), nil)
		assert.Contains(t, out, "1着2着の独占")
		assert.Contains(t, out, "4段階昇級")
	})

	t.Run("an ordinary advance", func(t *testing.T) {
		o := defaultGuandanOpts()
		o.phase = domain.GuandanPhaseHandEnd
		o.lastResult = &domain.GuandanHandResult{WinnerTeam: 1, Advance: domain.GuandanAdvanceFirstFourth}
		out := new(presenter.GuandanCuiPresenter).Output(setupGuandanMock(o), nil)
		assert.Contains(t, out, "1段階昇級")
		assert.NotContains(t, out, "独占")
	})

	t.Run("no result yet still prompts for the next hand", func(t *testing.T) {
		o := defaultGuandanOpts()
		o.phase = domain.GuandanPhaseHandEnd
		out := new(presenter.GuandanCuiPresenter).Output(setupGuandanMock(o), nil)
		assert.Contains(t, out, "n ・・・次の局へ")
	})
}

// **着順が見えないと次局の貢が読めない。**
func TestGuandanCuiPresenter_ShowsTheFinishingOrder(t *testing.T) {
	o := defaultGuandanOpts()
	o.finished = []int{2, 0}
	out := new(presenter.GuandanCuiPresenter).Output(setupGuandanMock(o), nil)
	assert.Contains(t, out, "(T0 着順2)")
	assert.Contains(t, out, "着順1")
	assert.Contains(t, out, "着順-")
}

func TestGuandanCuiPresenter_Error(t *testing.T) {
	out := new(presenter.GuandanCuiPresenter).Output(setupGuandanMock(defaultGuandanOpts()), errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestGuandanCuiPresenter_ActionLogOutput(t *testing.T) {
	out := new(presenter.GuandanCuiPresenter).ActionLogOutput(setupGuandanMock(defaultGuandanOpts()))
	assert.True(t, strings.TrimSpace(out) != "")
}
