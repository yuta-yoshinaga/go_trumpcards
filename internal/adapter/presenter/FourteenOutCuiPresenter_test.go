//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// foCuiColumns は 12 列の盤を作る。指定した列だけ中身を持つ。
func foCuiColumns(cols ...[]*domain.Card) [][]*domain.Card {
	board := make([][]*domain.Card, domain.FourteenOutColumnCnt)
	for i := range board {
		if i < len(cols) {
			board[i] = cols[i]
		}
	}
	return board
}

func foCuiCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

func setupFourteenOutCuiMockDefaults(g *interfaces.MockFourteenOutGame) {
	g.On("GetPhase").Return(domain.FourteenOutPhasePlaying).Maybe()
	g.On("GetRemovedCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("GetColumns").Return(foCuiColumns(
		[]*domain.Card{foCuiCard(domain.CardDesignSpade, 3), foCuiCard(domain.CardDesignSpade, 7)},
	)).Maybe()
	g.On("CountRemovablePairs").Return(0).Maybe()
}

func newFourteenOutCuiMock(phase domain.FourteenOutPhase, removed int, stalemate bool) *interfaces.MockFourteenOutGame {
	g := new(interfaces.MockFourteenOutGame)
	g.On("GetPhase").Return(phase).Maybe()
	g.On("GetRemovedCount").Return(removed).Maybe()
	g.On("IsStalemate").Return(stalemate).Maybe()
	g.On("CountRemovablePairs").Return(0).Maybe()
	g.On("GetColumns").Return(foCuiColumns()).Maybe()
	return g
}

func TestFourteenOutCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockFourteenOutGame)
		setupFourteenOutCuiMockDefaults(g)
		result := new(FourteenOutCuiPresenter).Output(g, nil)

		assert.Contains(t, result, i18n.T("fourteenout.helpTitle"))
		// **列の中身も末尾も見える。**末尾しか動かせないので、どれが末尾かが
		// 分からないと次の一手が読めない。
		assert.Contains(t, result, cuiCardStr(foCuiCard(domain.CardDesignSpade, 7)))
		assert.Contains(t, result, cuiCardStr(foCuiCard(domain.CardDesignSpade, 3)))
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockFourteenOutGame)
		setupFourteenOutCuiMockDefaults(g)
		result := new(FourteenOutCuiPresenter).Output(g, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("stalemate", func(t *testing.T) {
		g := newFourteenOutCuiMock(domain.FourteenOutPhasePlaying, 40, true)
		assert.Contains(t, new(FourteenOutCuiPresenter).Output(g, nil), i18n.T("fourteenout.stalemate"))
	})

	t.Run("game clear", func(t *testing.T) {
		g := newFourteenOutCuiMock(domain.FourteenOutPhaseGameClear, 52, false)
		assert.Contains(t, new(FourteenOutCuiPresenter).Output(g, nil), i18n.T("fourteenout.gameClear"))
	})

	t.Run("game over", func(t *testing.T) {
		g := newFourteenOutCuiMock(domain.FourteenOutPhaseGameOver, 20, false)
		assert.Contains(t, new(FourteenOutCuiPresenter).Output(g, nil), i18n.T("fourteenout.gameOver"))
	})

	t.Run("marks an empty column", func(t *testing.T) {
		g := newFourteenOutCuiMock(domain.FourteenOutPhasePlaying, 2, false)
		assert.Contains(t, new(FourteenOutCuiPresenter).Output(g, nil),
			i18n.Tf("fourteenout.columnEmpty", "col", "0"))
	})
}

func TestFourteenOutCuiPresenter_HintOutput(t *testing.T) {
	t.Run("names both cards", func(t *testing.T) {
		g := new(interfaces.MockFourteenOutGame)
		nine := foCuiCard(domain.CardDesignSpade, 9)
		five := foCuiCard(domain.CardDesignHeart, 5)
		g.On("Hint").Return(&domain.FourteenOutHint{
			Action: domain.FourteenOutHintActionRemove, FromCol: 1, ToCol: 2,
		})
		g.On("GetColumns").Return(foCuiColumns(nil, []*domain.Card{nine}, []*domain.Card{five}))

		result := new(FourteenOutCuiPresenter).HintOutput(g)
		assert.Contains(t, result, cuiCardStr(nine))
		assert.Contains(t, result, cuiCardStr(five))
	})

	// 列が空でも落ちない (nil ガード)。列番号だけを出す。
	t.Run("falls back to column numbers when a tail is unreadable", func(t *testing.T) {
		g := new(interfaces.MockFourteenOutGame)
		g.On("Hint").Return(&domain.FourteenOutHint{
			Action: domain.FourteenOutHintActionRemove, FromCol: 1, ToCol: 2,
		})
		g.On("GetColumns").Return(foCuiColumns())

		result := new(FourteenOutCuiPresenter).HintOutput(g)
		assert.Contains(t, result, i18n.Tf("fourteenout.hintLineRemove", "col1", "1", "col2", "2"))
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockFourteenOutGame)
		g.On("Hint").Return((*domain.FourteenOutHint)(nil))
		assert.Contains(t, new(FourteenOutCuiPresenter).HintOutput(g), i18n.T("fourteenout.noHint"))
	})
}

func TestFourteenOutCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockFourteenOutGame)
		g.On("GetPhase").Return(domain.FourteenOutPhasePlaying)
		assert.NotEmpty(t, new(FourteenOutCuiPresenter).ActionLogOutput(g))
	})

	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockFourteenOutGame)
		g.On("GetPhase").Return(domain.FourteenOutPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "remove", Detail: "test"},
		})
		assert.NotEmpty(t, new(FourteenOutCuiPresenter).ActionLogOutput(g))
	})
}

// #5587: 取り除ける組の数はこのゲームの判断材料そのもの。Web は常時カウンタで
// 出しているのに、CUI は盤を目で走査させていた。
func TestFourteenOutCuiPresenter_ShowsTheRemovablePairCount(t *testing.T) {
	i18n.SetLang("ja")
	build := func(phase domain.FourteenOutPhase, pairs int) string {
		g := new(interfaces.MockFourteenOutGame)
		g.On("GetPhase").Return(phase).Maybe()
		g.On("GetRemovedCount").Return(0).Maybe()
		g.On("IsStalemate").Return(false).Maybe()
		g.On("GetColumns").Return(foCuiColumns()).Maybe()
		g.On("CountRemovablePairs").Return(pairs).Maybe()
		return new(FourteenOutCuiPresenter).Output(g, nil)
	}

	// **取り除きの後に数が変わること**は、プレゼンタが毎回ドメインに訊いて
	// いれば自動的に成り立つ。別の数を返せば別の数が出る。
	assert.Contains(t, build(domain.FourteenOutPhasePlaying, 4),
		i18n.Tf("fourteenout.removablePairs", "count", "4"))
	assert.Contains(t, build(domain.FourteenOutPhasePlaying, 0),
		i18n.Tf("fourteenout.removablePairs", "count", "0"))

	// **0 のときだけ色を変える。**Fourteen Out に補充は無いので、0 はそのまま
	// 敗北の合図。両方向を見ないと、色付けを落としても消しても気づけない。
	origNoColor := color.NoColor()
	color.SetNoColor(false)
	defer color.SetNoColor(origNoColor)
	assert.Contains(t, build(domain.FourteenOutPhasePlaying, 0),
		color.Yellow(i18n.Tf("fourteenout.removablePairs", "count", "0")))
	assert.NotContains(t, build(domain.FourteenOutPhasePlaying, 4),
		color.Yellow(i18n.Tf("fourteenout.removablePairs", "count", "4")))

	// 終局後は出さない。取り除ける組の話をする局面ではない。
	for _, ph := range []domain.FourteenOutPhase{domain.FourteenOutPhaseGameClear, domain.FourteenOutPhaseGameOver} {
		assert.NotContains(t, build(ph, 4), i18n.Tf("fourteenout.removablePairs", "count", "4"))
	}
}
