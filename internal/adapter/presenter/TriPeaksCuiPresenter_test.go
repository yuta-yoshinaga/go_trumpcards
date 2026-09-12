//go:build test

package presenter

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func triPeaksCuiMockWithLayout(layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard) *interfaces.MockTriPeaksGame {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksCuiMockDefaults(tg)
	tg.ExpectedCalls = filterCalls(tg.ExpectedCalls, "GetLayout")
	tg.On("GetLayout").Return(layout).Maybe()
	tg.On("IsExposed", mock.Anything, mock.Anything).Return(true).Maybe()
	return tg
}

func triPeaksRemainingLine(output string) string {
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "山の残り:") {
			return line
		}
	}
	return ""
}

func TestTriPeaksCuiPresenterOutput_PeakRemaining(t *testing.T) {
	makeLayout := func(cards ...struct {
		row, col int
		removed  bool
	}) [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard {
		var layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
		for i, card := range cards {
			layout[card.row][card.col] = &domain.TriPeaksCard{
				Card:    domain.NewCard(domain.CardDesignSpade, i+1, true),
				Removed: card.removed,
			}
		}
		return layout
	}

	t.Run("counts two different layouts", func(t *testing.T) {
		tests := []struct {
			name   string
			layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
			want   string
		}{
			{
				name: "one card in each peak",
				layout: makeLayout(
					struct {
						row, col int
						removed  bool
					}{3, 0, false},
					struct {
						row, col int
						removed  bool
					}{2, 3, false},
					struct {
						row, col int
						removed  bool
					}{1, 6, false},
				),
				want: "山の残り: 左 1 / 中 1 / 右 1",
			},
			{
				name: "removed cards and column boundaries",
				layout: makeLayout(
					struct {
						row, col int
						removed  bool
					}{3, 0, false},
					struct {
						row, col int
						removed  bool
					}{3, 2, true},
					struct {
						row, col int
						removed  bool
					}{3, 3, false},
					struct {
						row, col int
						removed  bool
					}{3, 5, false},
					struct {
						row, col int
						removed  bool
					}{3, 6, false},
				),
				want: "山の残り: 左 1 / 中 2 / 右 1",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				out := (&TriPeaksCuiPresenter{}).Output(triPeaksCuiMockWithLayout(tt.layout), nil)
				assert.Equal(t, tt.want, triPeaksRemainingLine(out))
				assert.NotContains(t, out, "{{")
			})
		}
	})
}

func TestTriPeaksCuiPresenterOutput_PeakRemainingInitialLayoutTotals28(t *testing.T) {
	game := domain.NewDefaultTriPeaks()
	game.Reset()
	tg := triPeaksCuiMockWithLayout(game.GetLayout())

	out := (&TriPeaksCuiPresenter{}).Output(tg, nil)
	assert.Equal(t, "山の残り: 左 9 / 中 9 / 右 10", triPeaksRemainingLine(out))

	// The three counts must account for every tableau card. This is the only
	// thing that catches a column-to-peak mapping that disagrees with
	// peakOfColumn in TriPeaksPage.tsx -- the two live in different languages
	// and never appear in each other's tests, so a drift would leave both green.
	// Read the numbers back out of the rendered line rather than restating them.
	var sum int
	for _, field := range regexp.MustCompile(`\d+`).FindAllString(triPeaksRemainingLine(out), -1) {
		n, err := strconv.Atoi(field)
		assert.NoError(t, err)
		sum += n
	}
	assert.Equal(t, domain.TriPeaksTableauCnt, sum)
}

func TestTriPeaksCuiPresenterOutput_PeakRemainingMarksOnlyEmptyPeaks(t *testing.T) {
	var layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
	layout[3][3] = &domain.TriPeaksCard{Card: domain.NewCard(domain.CardDesignSpade, 1, true)}
	layout[3][6] = &domain.TriPeaksCard{Card: domain.NewCard(domain.CardDesignHeart, 2, true)}

	out := (&TriPeaksCuiPresenter{}).Output(triPeaksCuiMockWithLayout(layout), nil)
	assert.Equal(t, "山の残り: 左 0 ✓ / 中 1 / 右 1", triPeaksRemainingLine(out))
	assert.Contains(t, out, "SPADE 1")
	assert.NotContains(t, out, "中 1 ✓")
	assert.NotContains(t, out, "右 1 ✓")
	assert.NotContains(t, out, "{{")
}

func setupTriPeaksCuiMockDefaults(tg *interfaces.MockTriPeaksGame) {
	tg.On("GetPhase").Return(domain.TriPeaksPhasePlaying).Maybe()
	tg.On("GetMoveCount").Return(0).Maybe()
	tg.On("CanUndo").Return(false).Maybe()
	tg.On("GetScore").Return(0).Maybe()
	tg.On("GetCombo").Return(0).Maybe()
	tg.On("GetStockCount").Return(23).Maybe()
	tg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	tg.On("IsStalemate").Return(false).Maybe()
	tg.On("UndoToEscape").Return(0).Maybe()

	var layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
	// Add some cards at row 3
	for c := range 10 {
		layout[3][c] = &domain.TriPeaksCard{
			Card:    domain.NewCard(domain.CardDesignSpade, c%13+1, false),
			Removed: false,
		}
		// Even columns exposed, odd columns blocked — exercises both formats.
		tg.On("IsExposed", 3, c).Return(c%2 == 0).Maybe()
	}
	tg.On("GetLayout").Return(layout).Maybe()
}

func TestTriPeaksCuiPresenterOutput_Playing(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksCuiMockDefaults(tg)
	p := &TriPeaksCuiPresenter{}

	result := p.Output(tg, nil)
	assert.Contains(t, result, "TriPeaks")
	assert.Contains(t, result, "Stock: 23枚")
	assert.Contains(t, result, "手数: 0")
	// No waste top -> nothing playable; stock remains, so draw is recommended.
	assert.Contains(t, result, "今出せるカード: 0枚")
	assert.Contains(t, result, "ドロー推奨")
}

func TestTriPeaksCuiPresenterOutput_PlayableAndBlocked(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	tg.On("GetPhase").Return(domain.TriPeaksPhasePlaying).Maybe()
	tg.On("GetMoveCount").Return(0).Maybe()
	tg.On("CanUndo").Return(false).Maybe()
	tg.On("GetScore").Return(0).Maybe()
	tg.On("GetCombo").Return(0).Maybe()
	tg.On("GetStockCount").Return(10).Maybe()
	tg.On("IsStalemate").Return(false).Maybe()
	// Waste top is a 2: adjacent to Ace(1) and 3, with K-A wrap also possible.
	tg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, false)}).Maybe()

	var layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
	// (3,0)=A exposed & adjacent->playable; (3,1)=5 blocked; (3,2)=9 exposed but not adjacent.
	layout[3][0] = &domain.TriPeaksCard{Card: domain.NewCard(domain.CardDesignSpade, 1, false)}
	layout[3][1] = &domain.TriPeaksCard{Card: domain.NewCard(domain.CardDesignSpade, 5, false)}
	layout[3][2] = &domain.TriPeaksCard{Card: domain.NewCard(domain.CardDesignSpade, 9, false)}
	tg.On("GetLayout").Return(layout).Maybe()
	tg.On("IsExposed", 3, 0).Return(true).Maybe()
	tg.On("IsExposed", 3, 1).Return(false).Maybe()
	tg.On("IsExposed", 3, 2).Return(true).Maybe()

	p := &TriPeaksCuiPresenter{}
	result := p.Output(tg, nil)
	// (3,0) is exposed and ±1 from the waste top -> playable marker.
	assert.Contains(t, result, "(3,0)")
	assert.Contains(t, result, "*")
	// (3,1) is blocked -> coordinates hidden.
	assert.Contains(t, result, "[--]")
	// (3,2) is exposed but not adjacent -> plain coordinate format, no marker.
	assert.Contains(t, result, "(3,2)")
	// Exactly one exposed card is playable, so no draw recommendation.
	assert.Contains(t, result, "今出せるカード: 1枚")
	assert.NotContains(t, result, "ドロー推奨")
}

func TestTriPeaksCuiPresenterOutput_Error(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksCuiMockDefaults(tg)
	p := &TriPeaksCuiPresenter{}

	result := p.Output(tg, errors.New("test error"))
	assert.Contains(t, result, "test error")
}

func TestTriPeaksCuiPresenterOutput_Stalemate(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksCuiMockDefaults(tg)
	tg.ExpectedCalls = nil
	tg.On("GetPhase").Return(domain.TriPeaksPhasePlaying).Maybe()
	tg.On("GetMoveCount").Return(5).Maybe()
	tg.On("CanUndo").Return(false).Maybe()
	tg.On("GetScore").Return(0).Maybe()
	tg.On("GetCombo").Return(0).Maybe()
	tg.On("GetStockCount").Return(0).Maybe()
	tg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	tg.On("IsStalemate").Return(true).Maybe()
	tg.On("UndoToEscape").Return(0).Maybe()
	var layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
	tg.On("GetLayout").Return(layout).Maybe()

	p := &TriPeaksCuiPresenter{}
	result := p.Output(tg, nil)
	assert.Contains(t, result, "手詰まり")
	// Nothing playable but the stock is empty, so no draw recommendation.
	assert.Contains(t, result, "今出せるカード: 0枚")
	assert.NotContains(t, result, "ドロー推奨")
}

func TestTriPeaksCuiPresenterOutput_GameClear(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksCuiMockDefaults(tg)
	tg.ExpectedCalls = nil
	tg.On("GetPhase").Return(domain.TriPeaksPhaseGameClear).Maybe()
	tg.On("GetMoveCount").Return(10).Maybe()
	tg.On("CanUndo").Return(false).Maybe()
	tg.On("GetScore").Return(0).Maybe()
	tg.On("GetCombo").Return(0).Maybe()
	tg.On("GetStockCount").Return(0).Maybe()
	tg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	tg.On("IsStalemate").Return(false).Maybe()
	var layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
	tg.On("GetLayout").Return(layout).Maybe()

	p := &TriPeaksCuiPresenter{}
	result := p.Output(tg, nil)
	assert.Contains(t, result, "ゲームクリア")
}

func TestTriPeaksCuiPresenterOutput_GameOver(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksCuiMockDefaults(tg)
	tg.ExpectedCalls = nil
	tg.On("GetPhase").Return(domain.TriPeaksPhaseGameOver).Maybe()
	tg.On("GetMoveCount").Return(5).Maybe()
	tg.On("CanUndo").Return(false).Maybe()
	tg.On("GetScore").Return(0).Maybe()
	tg.On("GetCombo").Return(0).Maybe()
	tg.On("GetStockCount").Return(0).Maybe()
	tg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	tg.On("IsStalemate").Return(false).Maybe()
	var layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
	tg.On("GetLayout").Return(layout).Maybe()

	p := &TriPeaksCuiPresenter{}
	result := p.Output(tg, nil)
	assert.Contains(t, result, "ゲームオーバー")
}

func TestTriPeaksCuiPresenterOutput_WithWaste(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksCuiMockDefaults(tg)
	tg.ExpectedCalls = nil
	tg.On("GetPhase").Return(domain.TriPeaksPhasePlaying).Maybe()
	tg.On("GetMoveCount").Return(1).Maybe()
	tg.On("CanUndo").Return(false).Maybe()
	tg.On("GetScore").Return(0).Maybe()
	tg.On("GetCombo").Return(0).Maybe()
	tg.On("GetStockCount").Return(22).Maybe()
	tg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, true)}).Maybe()
	tg.On("IsStalemate").Return(false).Maybe()
	var layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
	tg.On("GetLayout").Return(layout).Maybe()

	p := &TriPeaksCuiPresenter{}
	result := p.Output(tg, nil)
	assert.Contains(t, result, "Waste:")
	assert.NotContains(t, result, "[空]")
}

func TestTriPeaksCuiPresenterHintOutput(t *testing.T) {
	t.Run("remove hint", func(t *testing.T) {
		tg := new(interfaces.MockTriPeaksGame)
		tg.On("GetHint").Return(&domain.TriPeaksHint{Type: "remove", Row: 3, Col: 0})

		p := &TriPeaksCuiPresenter{}
		result := p.HintOutput(tg)
		assert.Contains(t, result, "カード除去")
	})

	t.Run("draw hint", func(t *testing.T) {
		tg := new(interfaces.MockTriPeaksGame)
		tg.On("GetHint").Return(&domain.TriPeaksHint{Type: "draw", Row: -1, Col: -1})

		p := &TriPeaksCuiPresenter{}
		result := p.HintOutput(tg)
		assert.Contains(t, result, "ストック")
	})

	t.Run("no hint", func(t *testing.T) {
		tg := new(interfaces.MockTriPeaksGame)
		tg.On("GetHint").Return((*domain.TriPeaksHint)(nil))

		p := &TriPeaksCuiPresenter{}
		result := p.HintOutput(tg)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("unknown hint type", func(t *testing.T) {
		tg := new(interfaces.MockTriPeaksGame)
		tg.On("GetHint").Return(&domain.TriPeaksHint{Type: "unknown"})

		p := &TriPeaksCuiPresenter{}
		result := p.HintOutput(tg)
		assert.Contains(t, result, "不明")
	})
}

func TestTriPeaksCuiPresenterActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		tg := new(interfaces.MockTriPeaksGame)
		tg.On("GetPhase").Return(domain.TriPeaksPhasePlaying)

		p := &TriPeaksCuiPresenter{}
		result := p.ActionLogOutput(tg)
		assert.NotEmpty(t, result)
	})

	t.Run("game over", func(t *testing.T) {
		tg := new(interfaces.MockTriPeaksGame)
		tg.On("GetPhase").Return(domain.TriPeaksPhaseGameOver)
		tg.On("GetActionLog").Return([]*domain.ActionLogEntry{})

		p := &TriPeaksCuiPresenter{}
		result := p.ActionLogOutput(tg)
		assert.NotEmpty(t, result)
	})
}

// **得点は Web だけの概念だった。** 計算が useTriPeaksScore にしか無く、CUI から
// 参照する値がサーバ側に存在しなかった (#5511)。ドメインへ移したので出せる。
func TestTriPeaksCuiPresenter_ShowsTheScore(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksCuiMockDefaults(tg)
	tg.ExpectedCalls = filterCalls(tg.ExpectedCalls, "GetScore")
	tg.On("GetScore").Return(1700)

	out := (&TriPeaksCuiPresenter{}).Output(tg, nil)
	assert.Contains(t, out, i18n.Tf("tripeaks.cuiScoreLine", "score", "1700"))
	// 手数の行は残っていること。得点で置き換えると手数が消える。
	assert.Contains(t, out, "手数: ")
}

// The web badges a chain from 2 and announces it; the CUI never called
// GetCombo at all, so a run of removals looked like unrelated single moves.
func TestTriPeaksCuiPresenter_ComboLine(t *testing.T) {
	p := new(TriPeaksCuiPresenter)

	build := func(combo int) *interfaces.MockTriPeaksGame {
		tg := new(interfaces.MockTriPeaksGame)
		// Registered before the defaults so this value is the one returned.
		tg.On("GetCombo").Return(combo).Maybe()
		setupTriPeaksCuiMockDefaults(tg)
		return tg
	}

	t.Run("names a chain of two", func(t *testing.T) {
		assert.Contains(t, p.Output(build(2), nil), "2")
		assert.Contains(t, p.Output(build(2), nil), "連鎖")
	})

	t.Run("says nothing at one, which is not a chain", func(t *testing.T) {
		// Printing it at 1 would make the line permanent and meaningless; the
		// web badge starts at 2 for the same reason.
		assert.NotContains(t, p.Output(build(1), nil), "連鎖")
	})

	t.Run("says nothing at zero", func(t *testing.T) {
		assert.NotContains(t, p.Output(build(0), nil), "連鎖")
	})
}
