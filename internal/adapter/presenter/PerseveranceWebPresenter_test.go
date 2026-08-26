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

func setupPerseveranceWebMockDefaults(bg *interfaces.MockPerseveranceGame) {
	bg.On("GetPhase").Return(domain.PerseverancePhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
	bg.On("CanUndo").Return(false).Maybe()
	bg.On("IsStalemate").Return(false).Maybe()
	bg.On("GetRedealsLeft").Return(domain.PerseveranceMaxRedeals).Maybe()
	bg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.PerseveranceTableauCnt][]*domain.PerseveranceTableauCard
	for i := range domain.PerseveranceTableauCnt {
		tableau[i] = make([]*domain.PerseveranceTableauCard, 4)
		for j := range 4 {
			tableau[i][j] = &domain.PerseveranceTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+1, false),
				FaceUp: true,
			}
		}
	}
	bg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.PerseveranceFoundationCnt][]*domain.Card
	bg.On("GetFoundation").Return(foundation).Maybe()
}

func parsePerseveranceOutput(t *testing.T, jsonStr string) *controller.PerseveranceWebOutput {
	t.Helper()
	var out controller.PerseveranceWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupPerseveranceOutputMock は Output 用の既定を組む。**Output() も受動ヒントを
// 埋めるようになった** (#4483) ので GetHint を呼べるようにする。共有ヘルパーに
// 置くと、先に登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupPerseveranceOutputMock(g *interfaces.MockPerseveranceGame) {
	setupPerseveranceWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestPerseveranceWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		setupPerseveranceOutputMock(bg)
		p := new(PerseveranceWebPresenter)

		result := parsePerseveranceOutput(t, p.Output(bg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Len(t, result.Tableau, domain.PerseveranceTableauCnt)
		assert.Len(t, result.Foundation, domain.PerseveranceFoundationCnt)
		assert.Equal(t, "perseverance.playing", result.MessageCode)
	})

	t.Run("all face up", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		setupPerseveranceOutputMock(bg)
		p := new(PerseveranceWebPresenter)

		result := parsePerseveranceOutput(t, p.Output(bg, nil))
		for _, col := range result.Tableau {
			for _, tc := range col {
				assert.True(t, tc.FaceUp)
				assert.NotNil(t, tc.Card)
			}
		}
	})

	t.Run("error message", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		setupPerseveranceOutputMock(bg)
		p := new(PerseveranceWebPresenter)

		result := parsePerseveranceOutput(t, p.Output(bg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		setupPerseveranceOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.PerseverancePhaseGameClear)

		p := new(PerseveranceWebPresenter)
		result := parsePerseveranceOutput(t, p.Output(bg, nil))
		assert.Equal(t, "perseverance.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		setupPerseveranceOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.PerseverancePhaseGameOver)

		p := new(PerseveranceWebPresenter)
		result := parsePerseveranceOutput(t, p.Output(bg, nil))
		assert.Equal(t, "perseverance.gameOver", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		setupPerseveranceOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)

		p := new(PerseveranceWebPresenter)
		result := parsePerseveranceOutput(t, p.Output(bg, nil))
		assert.Equal(t, "perseverance.stalemate", result.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない。ここが埋まっていないと
// フロントの `state.hint` は常に undefined で、それを読む分岐が全部死ぬ (#4483)。
func TestPerseveranceWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.PerseveranceHint{FromCol: 2, CardIndex: 1, ToZone: "foundation", ToCol: 0}

	bg := new(interfaces.MockPerseveranceGame)
	setupPerseveranceWebMockDefaults(bg)
	bg.On("GetHint").Return(hint).Maybe()

	result := parsePerseveranceOutput(t, new(PerseveranceWebPresenter).Output(bg, nil))
	if result.Hint == nil {
		t.Fatal("Output must carry the hint -- the frontend reads state.hint")
	}
	assert.Equal(t, 2, result.Hint.FromCol)
	assert.Equal(t, 1, result.Hint.CardIndex)
}

func TestPerseveranceWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		setupPerseveranceWebMockDefaults(bg)
		bg.On("GetHint").Return(&domain.PerseveranceHint{
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(PerseveranceWebPresenter)
		result := parsePerseveranceOutput(t, p.HintOutput(bg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "foundation", result.Hint.ToZone)
		assert.Equal(t, "perseverance.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		setupPerseveranceWebMockDefaults(bg)
		bg.On("GetHint").Return((*domain.PerseveranceHint)(nil))

		p := new(PerseveranceWebPresenter)
		result := parsePerseveranceOutput(t, p.HintOutput(bg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "perseverance.noHint", result.MessageCode)
	})
}

func TestPerseveranceWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		bg.On("GetPhase").Return(domain.PerseverancePhasePlaying)

		bg.On("GetGameEndFlag").Return(false)
		p := new(PerseveranceWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		bg.On("GetPhase").Return(domain.PerseverancePhaseGameOver)
		bg.On("GetGameEndFlag").Return(true)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(PerseveranceWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "move")
	})
}

// #5581: Web には `targets` に当たる操作が無く、置ける先は盤面から
// `perseveranceLegalTargets` が作る。TargetsOutput は通常の盤面をそのまま返すこと ──
// CUI 専用の応答が Web の経路に紛れ込まないように。
func TestPerseveranceWebPresenter_TargetsOutputIsTheOrdinaryBoard(t *testing.T) {
	g := domain.NewDefaultPerseverance()
	g.Reset()
	p := new(PerseveranceWebPresenter)

	assert.JSONEq(t, p.Output(g, nil), p.TargetsOutput(g, 3))
	// 列番号で答えが変わらないこと (盤面を返すだけなので)。
	assert.JSONEq(t, p.TargetsOutput(g, 0), p.TargetsOutput(g, 12))
}
