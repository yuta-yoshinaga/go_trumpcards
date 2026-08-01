//go:build test

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

func setupKingAlbertWebMockDefaults(bg *interfaces.MockKingAlbertGame) {
	bg.On("GetPhase").Return(domain.KingAlbertPhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
	bg.On("CanUndo").Return(false).Maybe()
	bg.On("IsStalemate").Return(false).Maybe()
	bg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
	for i := range domain.KingAlbertTableauCnt {
		tableau[i] = make([]*domain.KingAlbertTableauCard, i+1)
		for j := range i + 1 {
			tableau[i][j] = &domain.KingAlbertTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	bg.On("GetTableau").Return(tableau).Maybe()

	reserve := make([]*domain.Card, domain.KingAlbertReserveCnt)
	for i := range reserve {
		reserve[i] = domain.NewCard(domain.CardDesignHeart, i+2, false)
	}
	bg.On("GetReserve").Return(reserve).Maybe()

	var foundation [domain.KingAlbertFoundationCnt][]*domain.Card
	for i := range domain.KingAlbertFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade+i, 1, false)}
	}
	bg.On("GetFoundation").Return(foundation).Maybe()
}

func parseKingAlbertOutput(t *testing.T, jsonStr string) *controller.KingAlbertWebOutput {
	t.Helper()
	var out controller.KingAlbertWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupKingAlbertOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**
// ようになった (#4483) ので GetHint を呼べるようにする。
func setupKingAlbertOutputMock(g *interfaces.MockKingAlbertGame) {
	setupKingAlbertWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestKingAlbertWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		setupKingAlbertOutputMock(bg)
		p := new(KingAlbertWebPresenter)

		result := parseKingAlbertOutput(t, p.Output(bg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Len(t, result.Tableau, domain.KingAlbertTableauCnt)
		assert.Len(t, result.Reserve, domain.KingAlbertReserveCnt)
		assert.Len(t, result.Foundation, domain.KingAlbertFoundationCnt)
		assert.Equal(t, "kingalbert.playing", result.MessageCode)
	})

	t.Run("all face up", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		setupKingAlbertOutputMock(bg)
		p := new(KingAlbertWebPresenter)

		result := parseKingAlbertOutput(t, p.Output(bg, nil))
		for _, col := range result.Tableau {
			for _, tc := range col {
				assert.True(t, tc.FaceUp)
				assert.NotNil(t, tc.Card)
			}
		}
	})

	t.Run("error message", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		setupKingAlbertOutputMock(bg)
		p := new(KingAlbertWebPresenter)

		result := parseKingAlbertOutput(t, p.Output(bg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		setupKingAlbertOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.KingAlbertPhaseGameClear)

		p := new(KingAlbertWebPresenter)
		result := parseKingAlbertOutput(t, p.Output(bg, nil))
		assert.Equal(t, "kingalbert.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		setupKingAlbertOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.KingAlbertPhaseGameOver)

		p := new(KingAlbertWebPresenter)
		result := parseKingAlbertOutput(t, p.Output(bg, nil))
		assert.Equal(t, "kingalbert.gameOver", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		setupKingAlbertOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)

		p := new(KingAlbertWebPresenter)
		result := parseKingAlbertOutput(t, p.Output(bg, nil))
		assert.Equal(t, "kingalbert.stalemate", result.MessageCode)
	})

	t.Run("depleted reserve cell serialises null", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		setupKingAlbertOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetReserve")
		bg.On("GetReserve").Return([]*domain.Card{nil, domain.NewCard(domain.CardDesignSpade, 5, false)})

		p := new(KingAlbertWebPresenter)
		result := parseKingAlbertOutput(t, p.Output(bg, nil))
		assert.Len(t, result.Reserve, 2)
		assert.Nil(t, result.Reserve[0])
		assert.NotNil(t, result.Reserve[1])
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestKingAlbertWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.KingAlbertHint{FromZone: "tableau", FromCol: 2, CardIndex: 1, ToZone: "foundation", ToCol: 0}

	bg := new(interfaces.MockKingAlbertGame)
	setupKingAlbertWebMockDefaults(bg)
	bg.On("GetHint").Return(hint).Maybe()

	result := parseKingAlbertOutput(t, new(KingAlbertWebPresenter).Output(bg, nil))
	if result.Hint == nil {
		t.Fatal("Output must carry the hint -- the frontend reads state.hint")
	}
	assert.Equal(t, 2, result.Hint.FromCol)
}

func TestKingAlbertWebPresenter_HintOutput(t *testing.T) {
	t.Run("with tableau hint", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		setupKingAlbertWebMockDefaults(bg)
		bg.On("GetHint").Return(&domain.KingAlbertHint{
			FromZone:  "tableau",
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(KingAlbertWebPresenter)
		result := parseKingAlbertOutput(t, p.HintOutput(bg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "foundation", result.Hint.ToZone)
		assert.Equal(t, "kingalbert.hintAvailable", result.MessageCode)
	})

	t.Run("with reserve hint", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		setupKingAlbertWebMockDefaults(bg)
		bg.On("GetHint").Return(&domain.KingAlbertHint{
			FromZone:  "reserve",
			FromCol:   2,
			CardIndex: -1,
			ToZone:    "tableau",
			ToCol:     1,
		})

		p := new(KingAlbertWebPresenter)
		result := parseKingAlbertOutput(t, p.HintOutput(bg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "reserve", result.Hint.FromZone)
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		setupKingAlbertWebMockDefaults(bg)
		bg.On("GetHint").Return((*domain.KingAlbertHint)(nil))

		p := new(KingAlbertWebPresenter)
		result := parseKingAlbertOutput(t, p.HintOutput(bg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "kingalbert.noHint", result.MessageCode)
	})
}

func TestKingAlbertWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		bg.On("GetPhase").Return(domain.KingAlbertPhasePlaying)
		bg.On("GetGameEndFlag").Return(false)

		p := new(KingAlbertWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		bg.On("GetPhase").Return(domain.KingAlbertPhaseGameOver)
		bg.On("GetGameEndFlag").Return(true)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(KingAlbertWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "move")
	})
}
