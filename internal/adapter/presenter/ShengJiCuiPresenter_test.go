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

// **切札は切札スートだけではない。**これが読めないと序列が分からない。
func TestShengJiCuiPresenter_ExplainsTheWholeTrumpGroup(t *testing.T) {
	out := new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(defaultShengJiOpts()), nil)
	assert.Contains(t, out, "切札は切札スートだけではありません")
	assert.Contains(t, out, "全スートの5（レベル札）とジョーカー4枚も切札")
	assert.Contains(t, out, "赤ジョーカー > 黒ジョーカー")
}

// **点を集めるのは守備側。**目標が読めないと打ちようがない。
func TestShengJiCuiPresenter_SaysTheDefendersCollect(t *testing.T) {
	out := new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(defaultShengJiOpts()), nil)
	assert.Contains(t, out, "点を集めるのは守備側")
	assert.Contains(t, out, "現在 35 点")
	assert.Contains(t, out, "80 点で宣言側が交代")
	assert.Contains(t, out, "総得点 200 点")
}

// **レベルは 2〜A。**数字のままでは J/Q/K/A が読めない。
func TestShengJiCuiPresenter_LabelsLevelsAndSuits(t *testing.T) {
	out := new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(defaultShengJiOpts()), nil)
	assert.Contains(t, out, "この局のレベル: 5")
	assert.Contains(t, out, "切札: ♠")

	o := defaultShengJiOpts()
	o.trumpSuit = domain.ShengJiNoTrump
	out = new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(o), nil)
	assert.Contains(t, out, "切札: 無主")

	// **レベルは 2〜A。**J/Q/K/A は数字のままでは読めない。
	for level, want := range map[int]string{7: "7", 11: "J", 12: "Q", 13: "K", domain.ShengJiMaxLevel: "A"} {
		o := defaultShengJiOpts()
		o.level = level
		out := new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(o), nil)
		assert.Contains(t, out, "この局のレベル: "+want)
	}
}

func TestShengJiCuiPresenter_ShowsSidesAndHidesOtherHands(t *testing.T) {
	out := new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(defaultShengJiOpts()), nil)
	assert.Contains(t, out, "0:SPADE 2")
	assert.Contains(t, out, "非公開 1枚")
	assert.Contains(t, out, "宣言側")
	assert.Contains(t, out, "守備側")
	assert.Contains(t, out, "<- 手番")
}

func TestShengJiCuiPresenter_RevealsEveryHandAtGameEnd(t *testing.T) {
	o := defaultShengJiOpts()
	o.gameEnd = true
	o.winner = 0
	out := new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(o), nil)
	assert.NotContains(t, out, "非公開")
	assert.Contains(t, out, "チーム0の勝ち")
}

func TestShengJiCuiPresenter_Trick(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		out := new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(defaultShengJiOpts()), nil)
		assert.Contains(t, out, "まだ誰も出していません")
	})

	t.Run("the lead and each seat's cards", func(t *testing.T) {
		o := defaultShengJiOpts()
		o.trick = [][]*domain.Card{
			{sjiTestCard(domain.CardDesignHeart, 7), sjiTestCard(domain.CardDesignHeart, 7)},
			{sjiTestCard(domain.CardDesignHeart, 9), sjiTestCard(domain.CardDesignHeart, 9)},
		}
		o.leadCombo = &domain.ShengJiCombo{Kind: domain.ShengJiComboPair, Rank: 7, Size: 2}
		out := new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(o), nil)
		assert.Contains(t, out, "対子")
		assert.Contains(t, out, "（2枚）がリードされています")
		// リード席は 1 なので、2 番目に出したのは席 2。
		assert.Contains(t, out, "席1:")
		assert.Contains(t, out, "席2:")
	})

	t.Run("every combination has a label", func(t *testing.T) {
		for _, kind := range []domain.ShengJiComboKind{
			domain.ShengJiComboSingle, domain.ShengJiComboPair, domain.ShengJiComboTractor,
		} {
			o := defaultShengJiOpts()
			o.trick = [][]*domain.Card{{sjiTestCard(domain.CardDesignHeart, 7)}}
			o.leadCombo = &domain.ShengJiCombo{Kind: kind, Rank: 7, Size: 1}
			out := new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(o), nil)
			assert.NotContains(t, out, "場: - (")
		}
	})
}

func TestShengJiCuiPresenter_Declare(t *testing.T) {
	t.Run("nobody has declared yet", func(t *testing.T) {
		o := defaultShengJiOpts()
		o.phase = domain.ShengJiPhaseDeclare
		o.declarable = map[int]int{domain.CardDesignHeart: 2}
		out := new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(o), nil)
		assert.Contains(t, out, "まだ誰も亮牌していません")
		assert.Contains(t, out, "無主")
		assert.Contains(t, out, "宣言できるスート: 3=♥(x2)")
		assert.Contains(t, out, "d <0-4>")
	})

	// **強い宣言だけが上書きできる。**
	t.Run("an existing declaration shows its strength", func(t *testing.T) {
		o := defaultShengJiOpts()
		o.phase = domain.ShengJiPhaseDeclare
		o.declaration = &domain.ShengJiDeclaration{Seat: 2, Suit: domain.CardDesignClover, Strength: 1}
		out := new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(o), nil)
		assert.Contains(t, out, "席2 が ♣（強さ 1）")
		assert.Contains(t, out, "強い宣言だけが上書き")
	})

	// **持っていないスートは宣言できない。**
	t.Run("no level card means nothing to declare", func(t *testing.T) {
		o := defaultShengJiOpts()
		o.phase = domain.ShengJiPhaseDeclare
		out := new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(o), nil)
		assert.Contains(t, out, "宣言できるスートがありません")
	})
}

func TestShengJiCuiPresenter_KittyPrompt(t *testing.T) {
	o := defaultShengJiOpts()
	o.phase = domain.ShengJiPhaseKitty
	out := new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(o), nil)
	assert.Contains(t, out, "底牌に8枚埋め戻します")
	assert.Contains(t, out, "得点札と切札は埋めないでください")
}

func TestShengJiCuiPresenter_HandEnd(t *testing.T) {
	t.Run("the declarers held", func(t *testing.T) {
		o := defaultShengJiOpts()
		o.phase = domain.ShengJiPhaseHandEnd
		o.lastResult = &domain.ShengJiHandResult{
			DeclarerTeam: 0, DefenderPoints: 35, DeclarerHeld: true, Advance: 2, AdvancingTeam: 0,
		}
		out := new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(o), nil)
		assert.Contains(t, out, "宣言側が守りきりました")
		assert.Contains(t, out, "守備側 35 点 / 80 点")
		assert.Contains(t, out, "2段階昇級")
		assert.Contains(t, out, "n ・・・次の局へ")
	})

	// **底牌の倍率は最終トリックを取った側にしか掛からない。**
	t.Run("the defenders took it, with the kitty", func(t *testing.T) {
		o := defaultShengJiOpts()
		o.phase = domain.ShengJiPhaseHandEnd
		o.lastResult = &domain.ShengJiHandResult{
			DeclarerTeam: 0, DefenderPoints: 120, KittyPoints: 40, KittyMultiplier: 4,
			DeclarerHeld: false, Advance: 1, AdvancingTeam: 1,
		}
		out := new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(o), nil)
		assert.Contains(t, out, "守備側が 120 点を集めました")
		assert.Contains(t, out, "宣言側が交代")
		assert.Contains(t, out, "底牌: 40 点（倍率 x4）")
	})

	t.Run("no kitty line when nothing was multiplied", func(t *testing.T) {
		o := defaultShengJiOpts()
		o.phase = domain.ShengJiPhaseHandEnd
		o.lastResult = &domain.ShengJiHandResult{
			DeclarerTeam: 0, DefenderPoints: 35, DeclarerHeld: true, Advance: 2, AdvancingTeam: 0,
		}
		out := new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(o), nil)
		assert.NotContains(t, out, "底牌:")
	})

	t.Run("no result yet still prompts for the next hand", func(t *testing.T) {
		o := defaultShengJiOpts()
		o.phase = domain.ShengJiPhaseHandEnd
		out := new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(o), nil)
		assert.Contains(t, out, "n ・・・次の局へ")
	})
}

func TestShengJiCuiPresenter_Error(t *testing.T) {
	out := new(presenter.ShengJiCuiPresenter).Output(setupShengJiMock(defaultShengJiOpts()), errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestShengJiCuiPresenter_ActionLogOutput(t *testing.T) {
	out := new(presenter.ShengJiCuiPresenter).ActionLogOutput(setupShengJiMock(defaultShengJiOpts()))
	assert.True(t, strings.TrimSpace(out) != "")
}
