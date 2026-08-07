//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupGolfCuiMockDefaults(gg *interfaces.MockGolfGame) {
	gg.On("GetPhase").Return(domain.GolfPhasePlaying).Maybe()
	gg.On("GetMoveCount").Return(0).Maybe()
	gg.On("GetStockCount").Return(16).Maybe()
	gg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	gg.On("IsStalemate").Return(false).Maybe()

	var layout [domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard
	// Add some cards
	for c := range domain.GolfColCnt {
		for r := range domain.GolfRowCnt {
			layout[c][r] = &domain.GolfCard{
				Card:    domain.NewCard(domain.CardDesignSpade, (c*5+r)%13+1, false),
				Removed: false,
			}
		}
	}
	gg.On("GetLayout").Return(layout).Maybe()
	for c := range domain.GolfColCnt {
		for r := range domain.GolfRowCnt {
			if r == domain.GolfRowCnt-1 {
				gg.On("IsExposed", c, r).Return(true).Maybe()
			} else {
				gg.On("IsExposed", c, r).Return(false).Maybe()
			}
		}
	}
}

func TestGolfCuiPresenterOutput_Playing(t *testing.T) {
	gg := new(interfaces.MockGolfGame)
	setupGolfCuiMockDefaults(gg)
	p := &GolfCuiPresenter{}

	result := p.Output(gg, nil)
	assert.Contains(t, result, "Golf")
	assert.Contains(t, result, "Stock: 16枚")
	assert.Contains(t, result, "手数: 0")
}

func TestGolfCuiPresenterOutput_PlayableMarker(t *testing.T) {
	gg := new(interfaces.MockGolfGame)
	gg.On("GetPhase").Return(domain.GolfPhasePlaying).Maybe()
	gg.On("GetMoveCount").Return(0).Maybe()
	gg.On("GetStockCount").Return(10).Maybe()
	gg.On("IsStalemate").Return(false).Maybe()
	// Waste top is K(13): adjacent to Q(12) normally and to A(1) via K-A wrap.
	gg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 13, false)}).Maybe()

	bottom := domain.GolfRowCnt - 1
	var layout [domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard
	layout[0][bottom] = &domain.GolfCard{Card: domain.NewCard(domain.CardDesignSpade, 12, false)} // adjacent -> playable
	layout[1][bottom] = &domain.GolfCard{Card: domain.NewCard(domain.CardDesignSpade, 1, false)}  // K-A wrap -> playable
	layout[2][bottom] = &domain.GolfCard{Card: domain.NewCard(domain.CardDesignSpade, 8, false)}  // not adjacent
	gg.On("GetLayout").Return(layout).Maybe()
	gg.On("IsExposed", 0, bottom).Return(true).Maybe()
	gg.On("IsExposed", 1, bottom).Return(true).Maybe()
	gg.On("IsExposed", 2, bottom).Return(true).Maybe()
	gg.On("IsExposed", mock.Anything, mock.Anything).Return(false).Maybe()

	p := &GolfCuiPresenter{}
	result := p.Output(gg, nil)
	assert.Contains(t, result, "(0)SPADE 12*")   // adjacent
	assert.Contains(t, result, "(1)SPADE 1*")    // K-A wrap
	assert.Contains(t, result, "(2)SPADE 8")     // exposed, not adjacent
	assert.NotContains(t, result, "(2)SPADE 8*") // no marker on non-adjacent
}

func TestGolfCuiPresenterOutput_Error(t *testing.T) {
	gg := new(interfaces.MockGolfGame)
	setupGolfCuiMockDefaults(gg)
	p := &GolfCuiPresenter{}

	result := p.Output(gg, errors.New("test error"))
	assert.Contains(t, result, "test error")
}

func TestGolfCuiPresenterOutput_Stalemate(t *testing.T) {
	gg := new(interfaces.MockGolfGame)
	setupGolfCuiMockDefaults(gg)
	gg.ExpectedCalls = nil
	gg.On("GetPhase").Return(domain.GolfPhasePlaying).Maybe()
	gg.On("GetMoveCount").Return(5).Maybe()
	gg.On("GetStockCount").Return(0).Maybe()
	gg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	gg.On("IsStalemate").Return(true).Maybe()
	var layout [domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard
	gg.On("GetLayout").Return(layout).Maybe()
	for c := range domain.GolfColCnt {
		for r := range domain.GolfRowCnt {
			gg.On("IsExposed", c, r).Return(false).Maybe()
		}
	}

	p := &GolfCuiPresenter{}
	result := p.Output(gg, nil)
	assert.Contains(t, result, "手詰まり")
}

func TestGolfCuiPresenterOutput_GameClear(t *testing.T) {
	gg := new(interfaces.MockGolfGame)
	setupGolfCuiMockDefaults(gg)
	gg.ExpectedCalls = nil
	gg.On("GetPhase").Return(domain.GolfPhaseGameClear).Maybe()
	gg.On("GetMoveCount").Return(10).Maybe()
	gg.On("GetStockCount").Return(0).Maybe()
	gg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	gg.On("IsStalemate").Return(false).Maybe()
	var layout [domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard
	gg.On("GetLayout").Return(layout).Maybe()
	for c := range domain.GolfColCnt {
		for r := range domain.GolfRowCnt {
			gg.On("IsExposed", c, r).Return(false).Maybe()
		}
	}

	p := &GolfCuiPresenter{}
	result := p.Output(gg, nil)
	assert.Contains(t, result, "ゲームクリア")
}

func TestGolfCuiPresenterOutput_GameOver(t *testing.T) {
	gg := new(interfaces.MockGolfGame)
	setupGolfCuiMockDefaults(gg)
	gg.ExpectedCalls = nil
	gg.On("GetPhase").Return(domain.GolfPhaseGameOver).Maybe()
	gg.On("GetMoveCount").Return(5).Maybe()
	gg.On("GetStockCount").Return(0).Maybe()
	gg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	gg.On("IsStalemate").Return(false).Maybe()
	var layout [domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard
	gg.On("GetLayout").Return(layout).Maybe()
	for c := range domain.GolfColCnt {
		for r := range domain.GolfRowCnt {
			gg.On("IsExposed", c, r).Return(false).Maybe()
		}
	}

	p := &GolfCuiPresenter{}
	result := p.Output(gg, nil)
	assert.Contains(t, result, "ゲームオーバー")
}

func TestGolfCuiPresenterOutput_WithWaste(t *testing.T) {
	gg := new(interfaces.MockGolfGame)
	setupGolfCuiMockDefaults(gg)
	gg.ExpectedCalls = nil
	gg.On("GetPhase").Return(domain.GolfPhasePlaying).Maybe()
	gg.On("GetMoveCount").Return(1).Maybe()
	gg.On("GetStockCount").Return(15).Maybe()
	gg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, true)}).Maybe()
	gg.On("IsStalemate").Return(false).Maybe()
	var layout [domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard
	gg.On("GetLayout").Return(layout).Maybe()
	for c := range domain.GolfColCnt {
		for r := range domain.GolfRowCnt {
			gg.On("IsExposed", c, r).Return(false).Maybe()
		}
	}

	p := &GolfCuiPresenter{}
	result := p.Output(gg, nil)
	assert.Contains(t, result, "Waste:")
	assert.NotContains(t, result, "[空]")
}

func TestGolfCuiPresenterHintOutput(t *testing.T) {
	t.Run("remove hint", func(t *testing.T) {
		gg := new(interfaces.MockGolfGame)
		gg.On("GetHint").Return(&domain.GolfHint{Type: "remove", Col: 3})

		p := &GolfCuiPresenter{}
		result := p.HintOutput(gg)
		assert.Contains(t, result, "カード除去")
	})

	t.Run("draw hint", func(t *testing.T) {
		gg := new(interfaces.MockGolfGame)
		gg.On("GetHint").Return(&domain.GolfHint{Type: "draw", Col: -1})

		p := &GolfCuiPresenter{}
		result := p.HintOutput(gg)
		assert.Contains(t, result, "ストック")
	})

	t.Run("no hint", func(t *testing.T) {
		gg := new(interfaces.MockGolfGame)
		gg.On("GetHint").Return((*domain.GolfHint)(nil))

		p := &GolfCuiPresenter{}
		result := p.HintOutput(gg)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("unknown hint type", func(t *testing.T) {
		gg := new(interfaces.MockGolfGame)
		gg.On("GetHint").Return(&domain.GolfHint{Type: "unknown"})

		p := &GolfCuiPresenter{}
		result := p.HintOutput(gg)
		assert.Contains(t, result, "不明")
	})
}

func TestGolfCuiPresenterActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		gg := new(interfaces.MockGolfGame)
		gg.On("GetPhase").Return(domain.GolfPhasePlaying)

		p := &GolfCuiPresenter{}
		result := p.ActionLogOutput(gg)
		assert.NotEmpty(t, result)
	})

	t.Run("game over", func(t *testing.T) {
		gg := new(interfaces.MockGolfGame)
		gg.On("GetPhase").Return(domain.GolfPhaseGameOver)
		gg.On("GetActionLog").Return([]*domain.ActionLogEntry{})

		p := &GolfCuiPresenter{}
		result := p.ActionLogOutput(gg)
		assert.NotEmpty(t, result)
	})
}

// TestGolfCuiPresenter_NineHoleScorecard confirms the CUI accumulates a 9-hole
// card across deals, the way the web GUI's useGolfNineHole does (#4784).
func TestGolfCuiPresenter_NineHoleScorecard(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	// remaining は「タブローに残す枚数」。0 ならクリア相当のスコア。
	endedGame := func(remaining int, phase domain.GolfPhase) *interfaces.MockGolfGame {
		g := new(interfaces.MockGolfGame)
		var layout [domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard
		for i := 0; i < remaining; i++ {
			layout[i%domain.GolfColCnt][i/domain.GolfColCnt] =
				&domain.GolfCard{Card: domain.NewCard(domain.CardDesignSpade, i%13+1, false)}
		}
		g.On("GetLayout").Return(layout).Maybe()
		g.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
		g.On("GetStockCount").Return(0).Maybe()
		g.On("GetMoveCount").Return(10).Maybe()
		g.On("IsStalemate").Return(false).Maybe()
		g.On("IsExposed", mock.Anything, mock.Anything).Return(false).Maybe()
		g.On("GetPhase").Return(phase).Maybe()
		return g
	}

	t.Run("no scorecard before any deal has ended", func(t *testing.T) {
		p := &GolfCuiPresenter{}
		assert.NotContains(t, p.Output(endedGame(5, domain.GolfPhasePlaying), nil), "ホール")
	})

	t.Run("records each finished deal and keeps a running total", func(t *testing.T) {
		p := &GolfCuiPresenter{}
		assert.Contains(t, p.Output(endedGame(4, domain.GolfPhaseGameOver), nil), "ホール 1/9  今回: 4  合計: 4")

		// 次のディールが始まる。
		p.Output(endedGame(9, domain.GolfPhasePlaying), nil)
		assert.Contains(t, p.Output(endedGame(3, domain.GolfPhaseGameClear), nil), "ホール 2/9  今回: 3  合計: 7")
	})

	// **1ディールは1回しか数えない。**終局後は Output が何度も呼ばれる。
	t.Run("does not count the same deal twice", func(t *testing.T) {
		p := &GolfCuiPresenter{}
		p.Output(endedGame(4, domain.GolfPhaseGameOver), nil)
		result := p.Output(endedGame(4, domain.GolfPhaseGameOver), nil)
		assert.Contains(t, result, "ホール 1/9")
		assert.NotContains(t, result, "ホール 2/9")
	})

	t.Run("announces the total once all nine holes are in", func(t *testing.T) {
		p := &GolfCuiPresenter{}
		for i := 0; i < 9; i++ {
			p.Output(endedGame(2, domain.GolfPhaseGameOver), nil)
			p.Output(endedGame(2, domain.GolfPhasePlaying), nil)
		}
		result := p.Output(endedGame(2, domain.GolfPhasePlaying), nil)
		assert.Contains(t, result, "9ホール終了")
		assert.Contains(t, result, "合計 18 打")
	})

	// 9ホールを超えて記録しない。10ディール目は無視される。
	t.Run("stops recording past the ninth hole", func(t *testing.T) {
		p := &GolfCuiPresenter{}
		for i := 0; i < 10; i++ {
			p.Output(endedGame(2, domain.GolfPhaseGameOver), nil)
			p.Output(endedGame(2, domain.GolfPhasePlaying), nil)
		}
		assert.Contains(t, p.Output(endedGame(2, domain.GolfPhasePlaying), nil), "合計 18 打")
	})
}
