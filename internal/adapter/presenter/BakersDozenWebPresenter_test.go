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

func setupBakersDozenWebMockDefaults(bg *interfaces.MockBakersDozenGame) {
	bg.On("GetPhase").Return(domain.BakersDozenPhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
	bg.On("CanUndo").Return(false).Maybe()
	bg.On("IsStalemate").Return(false).Maybe()
	bg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.BakersDozenTableauCnt][]*domain.BakersDozenTableauCard
	for i := range domain.BakersDozenTableauCnt {
		tableau[i] = make([]*domain.BakersDozenTableauCard, 4)
		for j := range 4 {
			tableau[i][j] = &domain.BakersDozenTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+1, false),
				FaceUp: true,
			}
		}
	}
	bg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.BakersDozenFoundationCnt][]*domain.Card
	bg.On("GetFoundation").Return(foundation).Maybe()
}

func parseBakersDozenOutput(t *testing.T, jsonStr string) *controller.BakersDozenWebOutput {
	t.Helper()
	var out controller.BakersDozenWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupBakersDozenOutputMock は Output 用の既定を組む。**Output() も受動ヒントを
// 埋めるようになった** (#4483) ので GetHint を呼べるようにする。共有ヘルパーに
// 置くと、先に登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupBakersDozenOutputMock(g *interfaces.MockBakersDozenGame) {
	setupBakersDozenWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestBakersDozenWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		setupBakersDozenOutputMock(bg)
		p := new(BakersDozenWebPresenter)

		result := parseBakersDozenOutput(t, p.Output(bg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Len(t, result.Tableau, domain.BakersDozenTableauCnt)
		assert.Len(t, result.Foundation, domain.BakersDozenFoundationCnt)
		assert.Equal(t, "bakersdozen.playing", result.MessageCode)
	})

	t.Run("all face up", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		setupBakersDozenOutputMock(bg)
		p := new(BakersDozenWebPresenter)

		result := parseBakersDozenOutput(t, p.Output(bg, nil))
		for _, col := range result.Tableau {
			for _, tc := range col {
				assert.True(t, tc.FaceUp)
				assert.NotNil(t, tc.Card)
			}
		}
	})

	t.Run("error message", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		setupBakersDozenOutputMock(bg)
		p := new(BakersDozenWebPresenter)

		result := parseBakersDozenOutput(t, p.Output(bg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		setupBakersDozenOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.BakersDozenPhaseGameClear)

		p := new(BakersDozenWebPresenter)
		result := parseBakersDozenOutput(t, p.Output(bg, nil))
		assert.Equal(t, "bakersdozen.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		setupBakersDozenOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.BakersDozenPhaseGameOver)

		p := new(BakersDozenWebPresenter)
		result := parseBakersDozenOutput(t, p.Output(bg, nil))
		assert.Equal(t, "bakersdozen.gameOver", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		setupBakersDozenOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)

		p := new(BakersDozenWebPresenter)
		result := parseBakersDozenOutput(t, p.Output(bg, nil))
		assert.Equal(t, "bakersdozen.stalemate", result.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない。ここが埋まっていないと
// フロントの `state.hint` は常に undefined で、それを読む分岐が全部死ぬ (#4483)。
func TestBakersDozenWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.BakersDozenHint{FromCol: 2, CardIndex: 1, ToZone: "foundation", ToCol: 0}

	bg := new(interfaces.MockBakersDozenGame)
	setupBakersDozenWebMockDefaults(bg)
	bg.On("GetHint").Return(hint).Maybe()

	result := parseBakersDozenOutput(t, new(BakersDozenWebPresenter).Output(bg, nil))
	if result.Hint == nil {
		t.Fatal("Output must carry the hint -- the frontend reads state.hint")
	}
	assert.Equal(t, 2, result.Hint.FromCol)
	assert.Equal(t, 1, result.Hint.CardIndex)
}

func TestBakersDozenWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		setupBakersDozenWebMockDefaults(bg)
		bg.On("GetHint").Return(&domain.BakersDozenHint{
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(BakersDozenWebPresenter)
		result := parseBakersDozenOutput(t, p.HintOutput(bg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "foundation", result.Hint.ToZone)
		assert.Equal(t, "bakersdozen.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		setupBakersDozenWebMockDefaults(bg)
		bg.On("GetHint").Return((*domain.BakersDozenHint)(nil))

		p := new(BakersDozenWebPresenter)
		result := parseBakersDozenOutput(t, p.HintOutput(bg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "bakersdozen.noHint", result.MessageCode)
	})
}

func TestBakersDozenWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		bg.On("GetPhase").Return(domain.BakersDozenPhasePlaying)

		bg.On("GetGameEndFlag").Return(false)
		p := new(BakersDozenWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		bg.On("GetPhase").Return(domain.BakersDozenPhaseGameOver)
		bg.On("GetGameEndFlag").Return(true)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(BakersDozenWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "move")
	})
}

// #5581: Web には `targets` に当たる操作が無く、置ける先は盤面から
// `bakersDozenLegalTargets` が作る。TargetsOutput は通常の盤面をそのまま返すこと ──
// CUI 専用の応答が Web の経路に紛れ込まないように。
func TestBakersDozenWebPresenter_TargetsOutputIsTheOrdinaryBoard(t *testing.T) {
	g := domain.NewDefaultBakersDozen()
	g.Reset()
	p := new(BakersDozenWebPresenter)

	assert.JSONEq(t, p.Output(g, nil), p.TargetsOutput(g, 3))
	// 列番号で答えが変わらないこと (盤面を返すだけなので)。
	assert.JSONEq(t, p.TargetsOutput(g, 0), p.TargetsOutput(g, 12))
}
