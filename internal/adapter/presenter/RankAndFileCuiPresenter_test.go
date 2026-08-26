//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupRankAndFileCuiMockDefaults(fg *interfaces.MockRankAndFileGame) {
	fg.On("GetPhase").Return(domain.RankAndFilePhasePlaying).Maybe()
	fg.On("GetMoveCount").Return(0).Maybe()
	fg.On("GetStockCount").Return(64).Maybe()
	fg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	fg.On("IsStalemate").Return(false).Maybe()
	fg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
	for i := range domain.RankAndFileTableauCnt {
		tableau[i] = make([]*domain.RankAndFileTableauCard, 4)
		for j := range 4 {
			tableau[i][j] = &domain.RankAndFileTableauCard{
				Card: domain.NewCard(domain.CardDesignSpade, j+1, false),
				// 配りどおり。全部 true にすると、この既定盤では
				// 伏せ札の描画が一度も通らない。
				FaceUp: j == 3,
			}
		}
	}
	fg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.RankAndFileFoundationCnt][]*domain.Card
	fg.On("GetFoundation").Return(foundation).Maybe()
}

func TestRankAndFileCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		fg := new(interfaces.MockRankAndFileGame)
		setupRankAndFileCuiMockDefaults(fg)
		p := new(RankAndFileCuiPresenter)

		result := p.Output(fg, nil)
		assert.Contains(t, result, "Rank and File")
		assert.Contains(t, result, "Foundation")
		assert.Contains(t, result, "Stock: 64枚")
		assert.Contains(t, result, "Waste: [空]")
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "手数: 0")
		assert.Contains(t, result, "操作: m で移動")
	})

	t.Run("with waste card", func(t *testing.T) {
		fg := new(interfaces.MockRankAndFileGame)
		setupRankAndFileCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetWaste")
		fg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})

		p := new(RankAndFileCuiPresenter)
		result := p.Output(fg, nil)
		assert.Contains(t, result, "Waste:")
		assert.NotContains(t, result, "Waste: [空]")
	})

	t.Run("with error", func(t *testing.T) {
		fg := new(interfaces.MockRankAndFileGame)
		setupRankAndFileCuiMockDefaults(fg)
		p := new(RankAndFileCuiPresenter)

		result := p.Output(fg, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game clear", func(t *testing.T) {
		fg := new(interfaces.MockRankAndFileGame)
		setupRankAndFileCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetPhase")
		fg.On("GetPhase").Return(domain.RankAndFilePhaseGameClear)

		p := new(RankAndFileCuiPresenter)
		result := p.Output(fg, nil)
		assert.Contains(t, result, "ゲームクリア")
		assert.NotContains(t, result, "操作: m で移動")
	})

	t.Run("game over", func(t *testing.T) {
		fg := new(interfaces.MockRankAndFileGame)
		setupRankAndFileCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetPhase")
		fg.On("GetPhase").Return(domain.RankAndFilePhaseGameOver)

		p := new(RankAndFileCuiPresenter)
		result := p.Output(fg, nil)
		assert.Contains(t, result, "ゲームオーバー")
		assert.NotContains(t, result, "操作: m で移動")
	})

	t.Run("stalemate", func(t *testing.T) {
		fg := new(interfaces.MockRankAndFileGame)
		setupRankAndFileCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "IsStalemate")
		fg.On("IsStalemate").Return(true)
		fg.On("UndoToEscape").Return(0).Maybe()

		p := new(RankAndFileCuiPresenter)
		result := p.Output(fg, nil)
		assert.Contains(t, result, "手詰まり")
	})

	t.Run("empty column", func(t *testing.T) {
		fg := new(interfaces.MockRankAndFileGame)
		setupRankAndFileCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetTableau")
		var emptyTableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		fg.On("GetTableau").Return(emptyTableau)

		p := new(RankAndFileCuiPresenter)
		result := p.Output(fg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("foundation with cards", func(t *testing.T) {
		fg := new(interfaces.MockRankAndFileGame)
		setupRankAndFileCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetFoundation")
		var foundation [domain.RankAndFileFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		fg.On("GetFoundation").Return(foundation)

		p := new(RankAndFileCuiPresenter)
		result := p.Output(fg, nil)
		assert.Contains(t, result, "SPADE 1")
	})
}

func TestRankAndFileCuiPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		fg := new(interfaces.MockRankAndFileGame)
		fg.On("GetHint").Return(&domain.RankAndFileHint{
			FromZone:  "tableau",
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(RankAndFileCuiPresenter)
		result := p.HintOutput(fg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "タブロー列0")
		assert.Contains(t, result, "ファンデーション")
	})

	// #5525: ストックだけ残っている局面は行き詰まりではないので、
	// 「ヒントはありません」ではなく「引け」と言う。
	t.Run("stock hint tells the player to draw", func(t *testing.T) {
		fg := new(interfaces.MockRankAndFileGame)
		fg.On("GetHint").Return(&domain.RankAndFileHint{
			FromZone:  "stock",
			FromCol:   -1,
			CardIndex: -1,
			ToZone:    "waste",
			ToCol:     -1,
		})

		result := new(RankAndFileCuiPresenter).HintOutput(fg)
		assert.Contains(t, result, i18n.T("rankandfile.hintDraw"))
		assert.NotContains(t, result, i18n.T("cuiHintNone"))
		// 移動ヒントの体裁 (「A → B」) には落とさない。列 -1 が漏れる。
		assert.NotContains(t, result, "-1")
	})

	t.Run("waste hint", func(t *testing.T) {
		fg := new(interfaces.MockRankAndFileGame)
		fg.On("GetHint").Return(&domain.RankAndFileHint{
			FromZone:  "waste",
			FromCol:   -1,
			CardIndex: -1,
			ToZone:    "tableau",
			ToCol:     3,
		})

		p := new(RankAndFileCuiPresenter)
		result := p.HintOutput(fg)
		assert.Contains(t, result, "ウェイスト")
		assert.Contains(t, result, "タブロー列3")
	})

	t.Run("no hint", func(t *testing.T) {
		fg := new(interfaces.MockRankAndFileGame)
		fg.On("GetHint").Return((*domain.RankAndFileHint)(nil))

		p := new(RankAndFileCuiPresenter)
		result := p.HintOutput(fg)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestRankAndFileCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		fg := new(interfaces.MockRankAndFileGame)
		fg.On("GetPhase").Return(domain.RankAndFilePhasePlaying)

		p := new(RankAndFileCuiPresenter)
		result := p.ActionLogOutput(fg)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		fg := new(interfaces.MockRankAndFileGame)
		fg.On("GetPhase").Return(domain.RankAndFilePhaseGameOver)
		fg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "draw", Detail: "test"},
		})

		p := new(RankAndFileCuiPresenter)
		result := p.ActionLogOutput(fg)
		assert.Contains(t, result, "draw")
	})
}

// **伏せ札を伏せて出すこと。**クローン元の Forty Thieves は 40 枚すべて表向きに
// 配るので `FaceUp` を無視して描いても害が無かった。Rank and File は各列 4 枚の
// うち 3 枚が伏せなので、無視すると CUI が隠し札を全部見せてしまう。
func TestRankAndFileCuiPresenter_HidesFaceDownCards(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	fg := new(interfaces.MockRankAndFileGame)
	setupRankAndFileCuiMockDefaults(fg)
	fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetTableau")

	var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
	// 配られたとおりの列: 下 3 枚が伏せ、最上段だけが表。
	tableau[0] = []*domain.RankAndFileTableauCard{
		{Card: domain.NewCard(domain.CardDesignHeart, 7, false), FaceUp: false},
		{Card: domain.NewCard(domain.CardDesignSpade, 13, false), FaceUp: true},
	}
	fg.On("GetTableau").Return(tableau)

	p := &RankAndFileCuiPresenter{}
	out := p.Output(fg, nil)

	assert.Contains(t, out, "[0]??", "伏せ札は ?? で出す")
	// 負のコントロール: 表向きの札は今までどおり出る。?? を全部に出して
	// 通してしまわないことを見る。
	assert.Contains(t, out, cuiCardStr(domain.NewCard(domain.CardDesignSpade, 13, false)))
	assert.NotContains(t, out, cuiCardStr(domain.NewCard(domain.CardDesignHeart, 7, false)),
		"伏せ札の中身が漏れていない")
}
