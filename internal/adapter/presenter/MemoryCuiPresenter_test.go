package presenter

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func newMockMemoryGame() *interfaces.MockMemoryGame {
	return new(interfaces.MockMemoryGame)
}

func setupMemoryMockDefaults(mg *interfaces.MockMemoryGame) {
	mg.On("GetPlayerCnt").Return(4).Maybe()
	mg.On("GetGameEndFlag").Return(false).Maybe()
	mg.On("GetPhase").Return(domain.MemoryPhaseFlip1).Maybe()
	mg.On("GetCurrentPlayerIdx").Return(0).Maybe()
	mg.On("GetFirstFlipPos").Return(-1).Maybe()
	mg.On("GetSecondFlipPos").Return(-1).Maybe()
	mg.On("GetLastMatchResult").Return(false).Maybe()
	mg.On("GetWinnerIdx").Return(-1).Maybe()
	mg.On("GetTurnNumber").Return(0).Maybe()

	board := make([]*domain.MemoryBoardCard, domain.MemoryBoardSize)
	for i := 0; i < domain.MemoryBoardSize; i++ {
		board[i] = &domain.MemoryBoardCard{
			Card:   domain.NewCard(domain.CardDesignSpade, (i%13)+1, false),
			FaceUp: false,
			Taken:  false,
		}
	}
	mg.On("GetBoard").Return(board).Maybe()

	for i := 0; i < 4; i++ {
		p := domain.NewMemoryPlayer(i == 0)
		mg.On("GetPlayer", i).Return(p).Maybe()
	}
}

func TestMemoryCuiPresenterOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	t.Run("initial state", func(t *testing.T) {
		mg := newMockMemoryGame()
		setupMemoryMockDefaults(mg)
		p := new(MemoryCuiPresenter)

		result := p.Output(mg, nil)
		assert.Contains(t, result, "Memory (神経衰弱)")
		assert.Contains(t, result, "あなた: 0ペア")
		assert.Contains(t, result, "CPU 1: 0ペア")
		assert.Contains(t, result, "1枚目を選んでください")
		assert.Contains(t, result, "f <pos>")
		// Progress shows all 26 pairs remaining; the seen-cell count starts at 0.
		assert.Contains(t, result, "残り 26 ペア / 全 26 ペア")
		assert.Contains(t, result, "既出セル数: 0")
	})

	t.Run("highlights a face-up cell", func(t *testing.T) {
		mg := newMockMemoryGame()
		setupMemoryMockDefaults(mg)
		mg.ExpectedCalls = nil
		mg.On("GetPlayerCnt").Return(4)
		mg.On("GetGameEndFlag").Return(false)
		mg.On("GetPhase").Return(domain.MemoryPhaseFlip2)
		mg.On("GetCurrentPlayerIdx").Return(0)
		mg.On("GetLastMatchResult").Return(false)
		mg.On("GetWinnerIdx").Return(-1)
		board := make([]*domain.MemoryBoardCard, domain.MemoryBoardSize)
		for i := 0; i < domain.MemoryBoardSize; i++ {
			board[i] = &domain.MemoryBoardCard{Card: domain.NewCard(domain.CardDesignSpade, (i%13)+1, false)}
		}
		board[0].FaceUp = true // the flipped card
		mg.On("GetBoard").Return(board)
		for i := 0; i < 4; i++ {
			mg.On("GetPlayer", i).Return(domain.NewMemoryPlayer(i == 0))
		}

		// Render with colour enabled so the highlight (yellow) is present.
		color.SetNoColor(false)
		defer color.SetNoColor(true)
		p := new(MemoryCuiPresenter)
		result := p.Output(mg, nil)
		expected := color.Yellow(fmt.Sprintf("%-10s", cuiCardStr(domain.NewCard(domain.CardDesignSpade, 1, false))))
		assert.Contains(t, result, expected)
	})

	t.Run("marks visited face-down cells distinctly with a legend", func(t *testing.T) {
		mg := newMockMemoryGame()
		setupMemoryMockDefaults(mg)
		mg.ExpectedCalls = nil // clear defaults
		mg.On("GetPlayerCnt").Return(4)
		mg.On("GetGameEndFlag").Return(false)
		mg.On("GetPhase").Return(domain.MemoryPhaseFlip1)
		mg.On("GetCurrentPlayerIdx").Return(0)
		mg.On("GetLastMatchResult").Return(false)
		mg.On("GetWinnerIdx").Return(-1)

		board := make([]*domain.MemoryBoardCard, domain.MemoryBoardSize)
		for i := 0; i < domain.MemoryBoardSize; i++ {
			board[i] = &domain.MemoryBoardCard{
				Card:   domain.NewCard(domain.CardDesignSpade, (i%13)+1, false),
				FaceUp: false,
				Taken:  false,
			}
		}
		// Position 0 was seen before (face-down but visited).
		board[0].Visited = true
		mg.On("GetBoard").Return(board)
		for i := 0; i < 4; i++ {
			mg.On("GetPlayer", i).Return(domain.NewMemoryPlayer(i == 0))
		}

		p := new(MemoryCuiPresenter)
		result := p.Output(mg, nil)
		assert.Contains(t, result, "*?") // visited, face-down
		assert.Contains(t, result, "??") // unvisited cells
		assert.Contains(t, result, "凡例")
	})

	t.Run("flip2 phase", func(t *testing.T) {
		mg := newMockMemoryGame()
		setupMemoryMockDefaults(mg)
		mg.ExpectedCalls = nil // clear defaults
		mg.On("GetPlayerCnt").Return(4)
		mg.On("GetGameEndFlag").Return(false)
		mg.On("GetPhase").Return(domain.MemoryPhaseFlip2)
		mg.On("GetCurrentPlayerIdx").Return(0)
		mg.On("GetLastMatchResult").Return(false)
		mg.On("GetWinnerIdx").Return(-1)

		board := make([]*domain.MemoryBoardCard, domain.MemoryBoardSize)
		for i := 0; i < domain.MemoryBoardSize; i++ {
			board[i] = &domain.MemoryBoardCard{
				Card:   domain.NewCard(domain.CardDesignSpade, (i%13)+1, false),
				FaceUp: false,
				Taken:  false,
			}
		}
		// Card at position 0 is face up
		board[0].FaceUp = true
		mg.On("GetBoard").Return(board)
		for i := 0; i < 4; i++ {
			mg.On("GetPlayer", i).Return(domain.NewMemoryPlayer(i == 0))
		}

		p := new(MemoryCuiPresenter)
		result := p.Output(mg, nil)
		assert.Contains(t, result, "2枚目を選んでください")
		assert.Contains(t, result, "SPADE 1") // face up card
	})

	t.Run("result phase match", func(t *testing.T) {
		mg := newMockMemoryGame()
		setupMemoryMockDefaults(mg)
		mg.ExpectedCalls = nil
		mg.On("GetPlayerCnt").Return(4)
		mg.On("GetGameEndFlag").Return(false)
		mg.On("GetPhase").Return(domain.MemoryPhaseResult)
		mg.On("GetCurrentPlayerIdx").Return(0)
		mg.On("GetLastMatchResult").Return(true)
		mg.On("GetWinnerIdx").Return(-1)

		board := make([]*domain.MemoryBoardCard, domain.MemoryBoardSize)
		for i := 0; i < domain.MemoryBoardSize; i++ {
			board[i] = &domain.MemoryBoardCard{
				Card:   domain.NewCard(domain.CardDesignSpade, (i%13)+1, false),
				FaceUp: false,
				Taken:  false,
			}
		}
		mg.On("GetBoard").Return(board)
		for i := 0; i < 4; i++ {
			mg.On("GetPlayer", i).Return(domain.NewMemoryPlayer(i == 0))
		}

		p := new(MemoryCuiPresenter)
		result := p.Output(mg, nil)
		assert.Contains(t, result, "ペアが揃いました！")
		assert.Contains(t, result, "n・・・次へ")
	})

	t.Run("result phase miss", func(t *testing.T) {
		mg := newMockMemoryGame()
		setupMemoryMockDefaults(mg)
		mg.ExpectedCalls = nil
		mg.On("GetPlayerCnt").Return(4)
		mg.On("GetGameEndFlag").Return(false)
		mg.On("GetPhase").Return(domain.MemoryPhaseResult)
		mg.On("GetCurrentPlayerIdx").Return(0)
		mg.On("GetLastMatchResult").Return(false)
		mg.On("GetWinnerIdx").Return(-1)

		board := make([]*domain.MemoryBoardCard, domain.MemoryBoardSize)
		for i := 0; i < domain.MemoryBoardSize; i++ {
			board[i] = &domain.MemoryBoardCard{
				Card:   domain.NewCard(domain.CardDesignSpade, (i%13)+1, false),
				FaceUp: false,
				Taken:  false,
			}
		}
		mg.On("GetBoard").Return(board)
		for i := 0; i < 4; i++ {
			mg.On("GetPlayer", i).Return(domain.NewMemoryPlayer(i == 0))
		}

		p := new(MemoryCuiPresenter)
		result := p.Output(mg, nil)
		assert.Contains(t, result, "残念、不一致です。")
	})

	t.Run("game end human wins", func(t *testing.T) {
		mg := newMockMemoryGame()
		setupMemoryMockDefaults(mg)
		mg.ExpectedCalls = nil
		mg.On("GetPlayerCnt").Return(4)
		mg.On("GetGameEndFlag").Return(true)
		mg.On("GetPhase").Return(domain.MemoryPhaseFlip1)
		mg.On("GetWinnerIdx").Return(0)

		board := make([]*domain.MemoryBoardCard, domain.MemoryBoardSize)
		for i := 0; i < domain.MemoryBoardSize; i++ {
			board[i] = &domain.MemoryBoardCard{
				Card:   domain.NewCard(domain.CardDesignSpade, (i%13)+1, false),
				FaceUp: false,
				Taken:  true,
			}
		}
		mg.On("GetBoard").Return(board)
		for i := 0; i < 4; i++ {
			mg.On("GetPlayer", i).Return(domain.NewMemoryPlayer(i == 0))
		}

		p := new(MemoryCuiPresenter)
		result := p.Output(mg, nil)
		assert.Contains(t, result, "ゲーム終了！ あなたの勝利です！")
	})

	t.Run("game end CPU wins", func(t *testing.T) {
		mg := newMockMemoryGame()
		setupMemoryMockDefaults(mg)
		mg.ExpectedCalls = nil
		mg.On("GetPlayerCnt").Return(4)
		mg.On("GetGameEndFlag").Return(true)
		mg.On("GetPhase").Return(domain.MemoryPhaseFlip1)
		mg.On("GetWinnerIdx").Return(2)

		board := make([]*domain.MemoryBoardCard, domain.MemoryBoardSize)
		for i := 0; i < domain.MemoryBoardSize; i++ {
			board[i] = &domain.MemoryBoardCard{
				Card:   domain.NewCard(domain.CardDesignSpade, (i%13)+1, false),
				FaceUp: false,
				Taken:  true,
			}
		}
		mg.On("GetBoard").Return(board)
		for i := 0; i < 4; i++ {
			mg.On("GetPlayer", i).Return(domain.NewMemoryPlayer(i == 0))
		}

		p := new(MemoryCuiPresenter)
		result := p.Output(mg, nil)
		assert.Contains(t, result, "ゲーム終了！ CPU 2の勝利です！")
	})

	t.Run("error message", func(t *testing.T) {
		mg := newMockMemoryGame()
		setupMemoryMockDefaults(mg)

		p := new(MemoryCuiPresenter)
		result := p.Output(mg, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("taken card shows blank", func(t *testing.T) {
		mg := newMockMemoryGame()
		setupMemoryMockDefaults(mg)
		mg.ExpectedCalls = nil
		mg.On("GetPlayerCnt").Return(4)
		mg.On("GetGameEndFlag").Return(false)
		mg.On("GetPhase").Return(domain.MemoryPhaseFlip1)
		mg.On("GetCurrentPlayerIdx").Return(0)
		mg.On("GetWinnerIdx").Return(-1)
		mg.On("GetLastMatchResult").Return(false)

		board := make([]*domain.MemoryBoardCard, domain.MemoryBoardSize)
		for i := 0; i < domain.MemoryBoardSize; i++ {
			board[i] = &domain.MemoryBoardCard{
				Card:   domain.NewCard(domain.CardDesignSpade, (i%13)+1, false),
				FaceUp: false,
				Taken:  i == 0, // position 0 taken
			}
		}
		mg.On("GetBoard").Return(board)
		for i := 0; i < 4; i++ {
			mg.On("GetPlayer", i).Return(domain.NewMemoryPlayer(i == 0))
		}

		p := new(MemoryCuiPresenter)
		result := p.Output(mg, nil)
		assert.Contains(t, result, "[ 0]    ") // taken card blank
	})
}

func TestMemoryCuiPresenterActionLog(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	t.Run("game not ended", func(t *testing.T) {
		mg := newMockMemoryGame()
		mg.On("GetGameEndFlag").Return(false)

		p := new(MemoryCuiPresenter)
		result := p.ActionLogOutput(mg)
		assert.NotEmpty(t, result)
	})

	t.Run("game ended with entries", func(t *testing.T) {
		mg := newMockMemoryGame()
		mg.On("GetGameEndFlag").Return(true)
		mg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "match", Detail: "ペア獲得"},
		})
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		mg.On("GetPlayer", mock.Anything).Return(domain.NewMemoryPlayer(true)).Maybe()

		p := new(MemoryCuiPresenter)
		result := p.ActionLogOutput(mg)
		assert.Contains(t, result, "match")
	})

	t.Run("game ended with nil log", func(t *testing.T) {
		mg := newMockMemoryGame()
		mg.On("GetGameEndFlag").Return(true)
		var nilLog []*domain.ActionLogEntry
		mg.On("GetActionLog").Return(nilLog)
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		mg.On("GetPlayer", mock.Anything).Return(domain.NewMemoryPlayer(true)).Maybe()

		p := new(MemoryCuiPresenter)
		result := p.ActionLogOutput(mg)
		assert.NotEmpty(t, result)
	})
}

// TestMemoryCuiPresenter_ShortBoard は 52 枚未満の盤面で Output が落ちないことを見る。
//
// 盤面描画が 4 行 x 13 列固定だったため、ペア数を減らすと index out of range で
// panic した (ADR-0035)。Web 側と同じ誤りで、こちらは CUI にペア数変更コマンドが
// 無いため現状ユーザーからは到達しないが、既存テストが 52 枚しか流しておらず
// 検出できない状態だったのは同じ。
func TestMemoryCuiPresenter_ShortBoard(t *testing.T) {
	for _, pairs := range []int{domain.MemoryMinPairCount, 20} {
		mg := newMockMemoryGame()
		setupMemoryMockDefaults(mg)

		board := make([]*domain.MemoryBoardCard, pairs*2)
		for i := range board {
			board[i] = &domain.MemoryBoardCard{
				Card:   domain.NewCard(domain.CardDesignSpade, (i/2)+1, false),
				FaceUp: false,
				Taken:  false,
			}
		}
		mg.ExpectedCalls = filterOutCall(mg.ExpectedCalls, "GetBoard")
		mg.On("GetBoard").Return(board).Maybe()

		p := &MemoryCuiPresenter{}
		assert.NotPanics(t, func() { p.Output(mg, nil) }, "ペア数 %d で panic してはならない", pairs)
	}
}
