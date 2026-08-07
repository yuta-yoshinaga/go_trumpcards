//go:build test

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// addPyramidExposedExpectations marks only the bottom row as exposed, so tests
// exercise both the exposed (coordinate) and blocked ([--]) card formats.
func addPyramidExposedExpectations(pg *interfaces.MockPyramidGame) {
	for row := range domain.PyramidRowCnt {
		for col := range row + 1 {
			pg.On("IsExposed", row, col).Return(row == domain.PyramidRowCnt-1).Maybe()
			pg.On("IsRemovableKing", row, col).Return(false).Maybe()
		}
	}
}

func setupPyramidCuiMock() *interfaces.MockPyramidGame {
	pg := new(interfaces.MockPyramidGame)
	pg.On("GetPhase").Return(domain.PyramidPhasePlaying).Maybe()
	pg.On("GetMoveCount").Return(0).Maybe()
	pg.On("GetStockCount").Return(24).Maybe()
	pg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	pg.On("IsStalemate").Return(false).Maybe()
	pg.On("IsWasteKingRemovable").Return(false).Maybe()
	addPyramidExposedExpectations(pg)

	var pyramid [domain.PyramidRowCnt][]*domain.PyramidCard
	for row := range domain.PyramidRowCnt {
		pyramid[row] = make([]*domain.PyramidCard, row+1)
		for col := range row + 1 {
			pyramid[row][col] = &domain.PyramidCard{
				Card:    domain.NewCard(domain.CardDesignSpade, (row+col)%13+1, false),
				Removed: false,
			}
		}
	}
	pg.On("GetPyramid").Return(pyramid).Maybe()
	return pg
}

func TestPyramidCuiPresenterOutput_Playing(t *testing.T) {
	pg := setupPyramidCuiMock()
	p := &PyramidCuiPresenter{}

	result := p.Output(pg, nil)
	assert.Contains(t, result, "Pyramid")
	assert.Contains(t, result, "Stock: 24枚")
	assert.Contains(t, result, "手数: 0")
}

func TestPyramidCuiPresenterOutput_ExposedVsBlocked(t *testing.T) {
	pg := setupPyramidCuiMock()
	p := &PyramidCuiPresenter{}

	result := p.Output(pg, nil)
	// Exposed bottom-row cards keep their (row,col) coordinate prefix.
	assert.Contains(t, result, "(6,0)")
	// Blocked upper-row cards hide coordinates behind the [--] marker.
	assert.Contains(t, result, "[--]")
	// The apex (row 0) is never exposed while children remain, so its
	// coordinate must not appear.
	assert.NotContains(t, result, "(0,0)")
}

func TestPyramidCuiPresenterOutput_GameClear(t *testing.T) {
	pg := setupPyramidCuiMock()
	pg.ExpectedCalls = nil
	pg.On("GetPhase").Return(domain.PyramidPhaseGameClear).Maybe()
	pg.On("GetMoveCount").Return(15).Maybe()
	pg.On("GetStockCount").Return(0).Maybe()
	pg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	pg.On("IsStalemate").Return(false).Maybe()

	var pyramid [domain.PyramidRowCnt][]*domain.PyramidCard
	for row := range domain.PyramidRowCnt {
		pyramid[row] = make([]*domain.PyramidCard, row+1)
		for col := range row + 1 {
			pyramid[row][col] = &domain.PyramidCard{
				Card:    domain.NewCard(domain.CardDesignSpade, 1, false),
				Removed: true,
			}
		}
	}
	pg.On("GetPyramid").Return(pyramid).Maybe()

	p := &PyramidCuiPresenter{}
	result := p.Output(pg, nil)
	assert.Contains(t, result, "ゲームクリア")
	assert.Contains(t, result, "15")
}

func TestPyramidCuiPresenterOutput_GameOver(t *testing.T) {
	pg := setupPyramidCuiMock()
	pg.ExpectedCalls = nil
	pg.On("GetPhase").Return(domain.PyramidPhaseGameOver).Maybe()
	pg.On("GetMoveCount").Return(5).Maybe()
	pg.On("GetStockCount").Return(0).Maybe()
	pg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	pg.On("IsStalemate").Return(false).Maybe()
	addPyramidExposedExpectations(pg)

	var pyramid [domain.PyramidRowCnt][]*domain.PyramidCard
	for row := range domain.PyramidRowCnt {
		pyramid[row] = make([]*domain.PyramidCard, row+1)
		for col := range row + 1 {
			pyramid[row][col] = &domain.PyramidCard{Card: domain.NewCard(domain.CardDesignSpade, 1, false), Removed: false}
		}
	}
	pg.On("GetPyramid").Return(pyramid).Maybe()

	p := &PyramidCuiPresenter{}
	result := p.Output(pg, nil)
	assert.Contains(t, result, "ゲームオーバー")
}

func TestPyramidCuiPresenterHintOutput_WithHint(t *testing.T) {
	pg := new(interfaces.MockPyramidGame)
	pg.On("GetHint").Return(&domain.PyramidHint{Type: "king", Row1: 6, Col1: 0, Row2: -1, Col2: -1})

	p := &PyramidCuiPresenter{}
	result := p.HintOutput(pg)
	assert.Contains(t, result, "キング除去")
}

func TestPyramidCuiPresenterHintOutput_NoHint(t *testing.T) {
	pg := new(interfaces.MockPyramidGame)
	pg.On("GetHint").Return((*domain.PyramidHint)(nil))

	p := &PyramidCuiPresenter{}
	result := p.HintOutput(pg)
	assert.Contains(t, result, "ヒントはありません")
}

func TestPyramidCuiPresenterHintOutput_Pair(t *testing.T) {
	pg := new(interfaces.MockPyramidGame)
	pg.On("GetHint").Return(&domain.PyramidHint{Type: "pair", Row1: 6, Col1: 0, Row2: 6, Col2: 1})

	p := &PyramidCuiPresenter{}
	result := p.HintOutput(pg)
	assert.Contains(t, result, "ペア除去")
}

func TestPyramidCuiPresenterHintOutput_WasteKing(t *testing.T) {
	pg := new(interfaces.MockPyramidGame)
	pg.On("GetHint").Return(&domain.PyramidHint{Type: "waste_king"})

	p := &PyramidCuiPresenter{}
	result := p.HintOutput(pg)
	assert.Contains(t, result, "ウェイストのキング除去")
}

func TestPyramidCuiPresenterHintOutput_WastePair(t *testing.T) {
	pg := new(interfaces.MockPyramidGame)
	pg.On("GetHint").Return(&domain.PyramidHint{Type: "waste_pair", Row1: 6, Col1: 0})

	p := &PyramidCuiPresenter{}
	result := p.HintOutput(pg)
	assert.Contains(t, result, "ウェイスト+ピラミッド")
}

func TestPyramidCuiPresenterHintOutput_UnknownTypeFallsThrough(t *testing.T) {
	pg := new(interfaces.MockPyramidGame)
	pg.On("GetHint").Return(&domain.PyramidHint{Type: "???"})

	p := &PyramidCuiPresenter{}
	result := p.HintOutput(pg)
	assert.Contains(t, result, "不明")
}

func TestPyramidCuiPresenterOutput_StalemateAndNonEmptyWaste(t *testing.T) {
	pg := setupPyramidCuiMock()
	pg.ExpectedCalls = nil
	pg.On("GetPhase").Return(domain.PyramidPhasePlaying).Maybe()
	pg.On("GetMoveCount").Return(7).Maybe()
	pg.On("GetStockCount").Return(0).Maybe()
	// Non-nil waste with one card exercises the wasteCard branch.
	pg.On("GetWaste").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
	}).Maybe()
	pg.On("IsStalemate").Return(true).Maybe()
	pg.On("IsWasteKingRemovable").Return(false).Maybe()
	addPyramidExposedExpectations(pg)

	var pyramid [domain.PyramidRowCnt][]*domain.PyramidCard
	for row := range domain.PyramidRowCnt {
		pyramid[row] = make([]*domain.PyramidCard, row+1)
		for col := range row + 1 {
			pyramid[row][col] = &domain.PyramidCard{
				Card:    domain.NewCard(domain.CardDesignSpade, 1, false),
				Removed: false,
			}
		}
	}
	pg.On("GetPyramid").Return(pyramid).Maybe()

	p := &PyramidCuiPresenter{}
	result := p.Output(pg, nil)
	assert.Contains(t, result, "手詰まりです")
	assert.Contains(t, result, "Waste: ")
}

func TestPyramidCuiPresenterActionLogOutput(t *testing.T) {
	pg := new(interfaces.MockPyramidGame)
	pg.On("GetPhase").Return(domain.PyramidPhasePlaying)

	p := &PyramidCuiPresenter{}
	result := p.ActionLogOutput(pg)
	assert.Contains(t, result, "棋譜はありません")
}

// **キングは相方が要らず単独で消せる (#4782)。**Web は常時ハイライトして
// いるのに、CUI は数値を1枚ずつ見て 13 を自分で探すしかなかった。
func TestPyramidCuiPresenterOutput_MarksRemovableKings(t *testing.T) {
	last := domain.PyramidRowCnt - 1
	// kingAt で指定した1枚だけを「除去できるキング」にした盤面のモック。
	board := func(kingCol int, waste []*domain.Card, wasteKing bool) *interfaces.MockPyramidGame {
		pg := new(interfaces.MockPyramidGame)
		pg.On("GetPhase").Return(domain.PyramidPhasePlaying).Maybe()
		pg.On("GetMoveCount").Return(0).Maybe()
		pg.On("GetStockCount").Return(24).Maybe()
		pg.On("GetWaste").Return(waste).Maybe()
		pg.On("IsStalemate").Return(false).Maybe()
		pg.On("IsWasteKingRemovable").Return(wasteKing).Maybe()
		for row := range domain.PyramidRowCnt {
			for col := range row + 1 {
				pg.On("IsExposed", row, col).Return(row == last).Maybe()
				pg.On("IsRemovableKing", row, col).Return(row == last && col == kingCol).Maybe()
			}
		}
		var pyramid [domain.PyramidRowCnt][]*domain.PyramidCard
		for row := range domain.PyramidRowCnt {
			pyramid[row] = make([]*domain.PyramidCard, row+1)
			for col := range row + 1 {
				pyramid[row][col] = &domain.PyramidCard{
					Card:    domain.NewCard(domain.CardDesignSpade, 13, false),
					Removed: false,
				}
			}
		}
		pg.On("GetPyramid").Return(pyramid).Maybe()
		return pg
	}
	p := &PyramidCuiPresenter{}

	// **印が付くのは1枚だけ。**全部に付ける実装でも「付く」だけなら通る。
	t.Run("marks one king on the pyramid and leaves the rest bare", func(t *testing.T) {
		out := p.Output(board(0, nil, false), nil)
		assert.Equal(t, 1, strings.Count(out, "[K]"))
	})

	t.Run("marks a king on top of the waste", func(t *testing.T) {
		out := p.Output(board(-1, []*domain.Card{domain.NewCard(domain.CardDesignSpade, 13, false)}, true), nil)
		assert.Contains(t, out, "[K]")
	})

	t.Run("does not mark a non-king waste card", func(t *testing.T) {
		out := p.Output(board(-1, []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}, false), nil)
		assert.NotContains(t, out, "[K]")
	})
}
