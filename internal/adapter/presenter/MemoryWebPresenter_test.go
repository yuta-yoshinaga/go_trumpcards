package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupMemoryWebMockDefaults(mg *interfaces.MockMemoryGame) {
	mg.On("GetPlayerCnt").Return(4).Maybe()
	mg.On("GetGameEndFlag").Return(false).Maybe()
	mg.On("GetPhase").Return(domain.MemoryPhaseFlip1).Maybe()
	mg.On("GetCurrentPlayerIdx").Return(0).Maybe()
	mg.On("GetFirstFlipPos").Return(-1).Maybe()
	mg.On("GetSecondFlipPos").Return(-1).Maybe()
	mg.On("GetLastMatchResult").Return(false).Maybe()
	mg.On("GetWinnerIdx").Return(-1).Maybe()
	mg.On("GetTurnNumber").Return(0).Maybe()
	mg.On("GetConfig").Return(domain.DefaultMemoryConfig()).Maybe()

	var board [domain.MemoryBoardSize]*domain.MemoryBoardCard
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

func parseMemoryOutput(t *testing.T, jsonStr string) *controller.MemoryWebOutput {
	t.Helper()
	var out controller.MemoryWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestMemoryWebPresenterOutput(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		mg := newMockMemoryGame()
		setupMemoryWebMockDefaults(mg)
		p := new(MemoryWebPresenter)

		result := parseMemoryOutput(t, p.Output(mg, nil))
		assert.Len(t, result.Players, 4)
		assert.True(t, result.Players[0].IsHuman)
		assert.False(t, result.Players[1].IsHuman)
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, -1, result.WinnerIdx)
		assert.Len(t, result.Board, 52)
		assert.Equal(t, "memory.flip1", result.MessageCode)

		// Face-down cards should have nil card data
		assert.Nil(t, result.Board[0].Card)
		assert.False(t, result.Board[0].FaceUp)
		assert.False(t, result.Board[0].Taken)
	})

	t.Run("face up card shows data", func(t *testing.T) {
		mg := newMockMemoryGame()
		setupMemoryWebMockDefaults(mg)
		mg.ExpectedCalls = nil
		mg.On("GetPlayerCnt").Return(4)
		mg.On("GetGameEndFlag").Return(false)
		mg.On("GetPhase").Return(domain.MemoryPhaseFlip2)
		mg.On("GetCurrentPlayerIdx").Return(0)
		mg.On("GetFirstFlipPos").Return(5)
		mg.On("GetSecondFlipPos").Return(-1)
		mg.On("GetLastMatchResult").Return(false)
		mg.On("GetWinnerIdx").Return(-1)
		mg.On("GetTurnNumber").Return(0)
		mg.On("GetConfig").Return(domain.DefaultMemoryConfig())

		var board [domain.MemoryBoardSize]*domain.MemoryBoardCard
		for i := 0; i < domain.MemoryBoardSize; i++ {
			board[i] = &domain.MemoryBoardCard{
				Card:   domain.NewCard(domain.CardDesignSpade, (i%13)+1, false),
				FaceUp: i == 5,
				Taken:  false,
			}
		}
		mg.On("GetBoard").Return(board)
		for i := 0; i < 4; i++ {
			mg.On("GetPlayer", i).Return(domain.NewMemoryPlayer(i == 0))
		}

		p := new(MemoryWebPresenter)
		result := parseMemoryOutput(t, p.Output(mg, nil))
		assert.NotNil(t, result.Board[5].Card)
		assert.True(t, result.Board[5].FaceUp)
		assert.Equal(t, "memory.flip2", result.MessageCode)
	})

	t.Run("taken card has no card data", func(t *testing.T) {
		mg := newMockMemoryGame()
		setupMemoryWebMockDefaults(mg)
		mg.ExpectedCalls = nil
		mg.On("GetPlayerCnt").Return(4)
		mg.On("GetGameEndFlag").Return(false)
		mg.On("GetPhase").Return(domain.MemoryPhaseFlip1)
		mg.On("GetCurrentPlayerIdx").Return(0)
		mg.On("GetFirstFlipPos").Return(-1)
		mg.On("GetSecondFlipPos").Return(-1)
		mg.On("GetLastMatchResult").Return(false)
		mg.On("GetWinnerIdx").Return(-1)
		mg.On("GetTurnNumber").Return(0)
		mg.On("GetConfig").Return(domain.DefaultMemoryConfig())

		var board [domain.MemoryBoardSize]*domain.MemoryBoardCard
		for i := 0; i < domain.MemoryBoardSize; i++ {
			board[i] = &domain.MemoryBoardCard{
				Card:   domain.NewCard(domain.CardDesignSpade, (i%13)+1, false),
				FaceUp: false,
				Taken:  i == 0,
			}
		}
		mg.On("GetBoard").Return(board)
		for i := 0; i < 4; i++ {
			mg.On("GetPlayer", i).Return(domain.NewMemoryPlayer(i == 0))
		}

		p := new(MemoryWebPresenter)
		result := parseMemoryOutput(t, p.Output(mg, nil))
		assert.Nil(t, result.Board[0].Card)
		assert.True(t, result.Board[0].Taken)
	})

	t.Run("result phase matched", func(t *testing.T) {
		mg := newMockMemoryGame()
		setupMemoryWebMockDefaults(mg)
		mg.ExpectedCalls = nil
		mg.On("GetPlayerCnt").Return(4)
		mg.On("GetGameEndFlag").Return(false)
		mg.On("GetPhase").Return(domain.MemoryPhaseResult)
		mg.On("GetCurrentPlayerIdx").Return(0)
		mg.On("GetFirstFlipPos").Return(0)
		mg.On("GetSecondFlipPos").Return(1)
		mg.On("GetLastMatchResult").Return(true)
		mg.On("GetWinnerIdx").Return(-1)
		mg.On("GetTurnNumber").Return(1)
		mg.On("GetConfig").Return(domain.DefaultMemoryConfig())

		var board [domain.MemoryBoardSize]*domain.MemoryBoardCard
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

		p := new(MemoryWebPresenter)
		result := parseMemoryOutput(t, p.Output(mg, nil))
		assert.Equal(t, "memory.matched", result.MessageCode)
	})

	t.Run("result phase mismatched", func(t *testing.T) {
		mg := newMockMemoryGame()
		setupMemoryWebMockDefaults(mg)
		mg.ExpectedCalls = nil
		mg.On("GetPlayerCnt").Return(4)
		mg.On("GetGameEndFlag").Return(false)
		mg.On("GetPhase").Return(domain.MemoryPhaseResult)
		mg.On("GetCurrentPlayerIdx").Return(0)
		mg.On("GetFirstFlipPos").Return(0)
		mg.On("GetSecondFlipPos").Return(2)
		mg.On("GetLastMatchResult").Return(false)
		mg.On("GetWinnerIdx").Return(-1)
		mg.On("GetTurnNumber").Return(1)
		mg.On("GetConfig").Return(domain.DefaultMemoryConfig())

		var board [domain.MemoryBoardSize]*domain.MemoryBoardCard
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

		p := new(MemoryWebPresenter)
		result := parseMemoryOutput(t, p.Output(mg, nil))
		assert.Equal(t, "memory.mismatched", result.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		mg := newMockMemoryGame()
		mg.On("GetPlayerCnt").Return(4)
		mg.On("GetGameEndFlag").Return(true)
		mg.On("GetPhase").Return(domain.MemoryPhaseGameEnd)
		mg.On("GetCurrentPlayerIdx").Return(0)
		mg.On("GetFirstFlipPos").Return(-1)
		mg.On("GetSecondFlipPos").Return(-1)
		mg.On("GetLastMatchResult").Return(false)
		mg.On("GetWinnerIdx").Return(0)
		mg.On("GetTurnNumber").Return(26)
		mg.On("GetConfig").Return(domain.DefaultMemoryConfig())

		var board [domain.MemoryBoardSize]*domain.MemoryBoardCard
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

		p := new(MemoryWebPresenter)
		result := parseMemoryOutput(t, p.Output(mg, nil))
		assert.True(t, result.GameEndFlag)
		assert.Equal(t, "memory.result.humanWin", result.MessageCode)
		assert.Contains(t, result.Message, "あなた")
	})

	t.Run("game end CPU wins", func(t *testing.T) {
		mg := newMockMemoryGame()
		mg.On("GetPlayerCnt").Return(4)
		mg.On("GetGameEndFlag").Return(true)
		mg.On("GetPhase").Return(domain.MemoryPhaseGameEnd)
		mg.On("GetCurrentPlayerIdx").Return(0)
		mg.On("GetFirstFlipPos").Return(-1)
		mg.On("GetSecondFlipPos").Return(-1)
		mg.On("GetLastMatchResult").Return(false)
		mg.On("GetWinnerIdx").Return(2)
		mg.On("GetTurnNumber").Return(26)
		mg.On("GetConfig").Return(domain.DefaultMemoryConfig())

		var board [domain.MemoryBoardSize]*domain.MemoryBoardCard
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

		p := new(MemoryWebPresenter)
		result := parseMemoryOutput(t, p.Output(mg, nil))
		assert.Equal(t, "memory.result.cpuWin", result.MessageCode)
		assert.Equal(t, "2", result.MessageParams["cpuId"])
	})

	t.Run("captured pairs are output in rank order", func(t *testing.T) {
		mg := newMockMemoryGame()
		mg.ExpectedCalls = nil
		mg.On("GetPlayerCnt").Return(4)
		mg.On("GetGameEndFlag").Return(false)
		mg.On("GetPhase").Return(domain.MemoryPhaseFlip1)
		mg.On("GetCurrentPlayerIdx").Return(0)
		mg.On("GetFirstFlipPos").Return(-1)
		mg.On("GetSecondFlipPos").Return(-1)
		mg.On("GetLastMatchResult").Return(false)
		mg.On("GetWinnerIdx").Return(-1)
		mg.On("GetTurnNumber").Return(0)
		mg.On("GetConfig").Return(domain.DefaultMemoryConfig())

		var board [domain.MemoryBoardSize]*domain.MemoryBoardCard
		for i := 0; i < domain.MemoryBoardSize; i++ {
			board[i] = &domain.MemoryBoardCard{
				Card:   domain.NewCard(domain.CardDesignSpade, (i%13)+1, false),
				FaceUp: false,
				Taken:  false,
			}
		}
		mg.On("GetBoard").Return(board)

		human := domain.NewMemoryPlayer(true)
		// Add pairs out of rank order to verify presenter sorts them.
		human.AddPair(domain.NewCard(domain.CardDesignSpade, 7, false), domain.NewCard(domain.CardDesignHeart, 7, false))
		human.AddPair(domain.NewCard(domain.CardDesignClover, 3, false), domain.NewCard(domain.CardDesignDiamond, 3, false))
		mg.On("GetPlayer", 0).Return(human)
		for i := 1; i < 4; i++ {
			mg.On("GetPlayer", i).Return(domain.NewMemoryPlayer(false))
		}

		p := new(MemoryWebPresenter)
		result := parseMemoryOutput(t, p.Output(mg, nil))
		assert.Len(t, result.Players[0].Pairs, 2)
		// Rank-ascending: 3 before 7.
		assert.Equal(t, 3, result.Players[0].Pairs[0].Value)
		assert.Equal(t, 7, result.Players[0].Pairs[1].Value)
		// Players with no captures expose an empty (non-nil) slice.
		assert.Empty(t, result.Players[1].Pairs)
	})

	t.Run("error message", func(t *testing.T) {
		mg := newMockMemoryGame()
		setupMemoryWebMockDefaults(mg)

		p := new(MemoryWebPresenter)
		result := parseMemoryOutput(t, p.Output(mg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
		assert.Empty(t, result.MessageCode)
	})

	t.Run("config values", func(t *testing.T) {
		mg := newMockMemoryGame()
		setupMemoryWebMockDefaults(mg)
		mg.ExpectedCalls = nil
		mg.On("GetPlayerCnt").Return(4)
		mg.On("GetGameEndFlag").Return(false)
		mg.On("GetPhase").Return(domain.MemoryPhaseFlip1)
		mg.On("GetCurrentPlayerIdx").Return(0)
		mg.On("GetFirstFlipPos").Return(-1)
		mg.On("GetSecondFlipPos").Return(-1)
		mg.On("GetLastMatchResult").Return(false)
		mg.On("GetWinnerIdx").Return(-1)
		mg.On("GetTurnNumber").Return(0)
		mg.On("GetConfig").Return(domain.MemoryConfig{CpuDifficulty: domain.MemoryCpuDifficultyHard})

		var board [domain.MemoryBoardSize]*domain.MemoryBoardCard
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

		p := new(MemoryWebPresenter)
		result := parseMemoryOutput(t, p.Output(mg, nil))
		assert.Equal(t, 2, result.Config.CpuDifficulty)
	})
}

func TestMemoryWebPresenterActionLog(t *testing.T) {
	t.Run("game not ended", func(t *testing.T) {
		mg := newMockMemoryGame()
		mg.On("GetGameEndFlag").Return(false)

		p := new(MemoryWebPresenter)
		result := p.ActionLogOutput(mg)
		assert.NotEmpty(t, result)
	})

	t.Run("game ended with entries", func(t *testing.T) {
		mg := newMockMemoryGame()
		mg.On("GetGameEndFlag").Return(true)
		mg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "match", Detail: "ペア獲得"},
		})

		p := new(MemoryWebPresenter)
		result := p.ActionLogOutput(mg)
		assert.Contains(t, result, "match")
	})

	t.Run("game ended with nil log", func(t *testing.T) {
		mg := newMockMemoryGame()
		mg.On("GetGameEndFlag").Return(true)
		var nilLog []*domain.ActionLogEntry
		mg.On("GetActionLog").Return(nilLog)

		p := new(MemoryWebPresenter)
		result := p.ActionLogOutput(mg)
		assert.NotEmpty(t, result)
	})
}

func TestMemoryWebPresenterBuildResultMessageNilPlayer(t *testing.T) {
	// nil player maps to isHuman=false → CPU message
	result := buildWinnerResultMessage(5, false)
	assert.Contains(t, result, "CPU 5")
}
