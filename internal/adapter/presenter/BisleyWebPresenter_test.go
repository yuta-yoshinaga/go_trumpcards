//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupBisleyWebMockDefaults(bg *interfaces.MockBisleyGame) {
	bg.On("GetPhase").Return(domain.BisleyPhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
	bg.On("CanUndo").Return(false).Maybe()
	bg.On("IsStalemate").Return(false).Maybe()
	bg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.BisleyTableauCnt][]*domain.BisleyTableauCard
	for i := range domain.BisleyTableauCnt {
		tableau[i] = make([]*domain.BisleyTableauCard, bisleyTestColumnLen)
		for j := range bisleyTestColumnLen {
			tableau[i][j] = &domain.BisleyTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	bg.On("GetTableau").Return(tableau).Maybe()

	var ace [domain.BisleyFoundationCnt][]*domain.Card
	var king [domain.BisleyFoundationCnt][]*domain.Card
	for i := range domain.BisleyFoundationCnt {
		ace[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade+i, 1, false)}
		king[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade+i, domain.CardValueMax, false)}
	}
	bg.On("GetAceFoundations").Return(ace).Maybe()
	bg.On("GetKingFoundations").Return(king).Maybe()
}

func parseBisleyOutput(t *testing.T, jsonStr string) *controller.BisleyWebOutput {
	t.Helper()
	var out controller.BisleyWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupBisleyOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**
// ようになった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと
// 先に登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupBisleyOutputMock(g *interfaces.MockBisleyGame) {
	setupBisleyWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestBisleyWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		setupBisleyOutputMock(bg)
		p := new(BisleyWebPresenter)

		result := parseBisleyOutput(t, p.Output(bg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Len(t, result.Tableau, domain.BisleyTableauCnt)
		assert.Len(t, result.AceFoundations, domain.BisleyFoundationCnt)
		assert.Len(t, result.KingFoundations, domain.BisleyFoundationCnt)
		assert.Equal(t, "bisley.playing", result.MessageCode)
	})

	t.Run("all face up", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		setupBisleyOutputMock(bg)
		p := new(BisleyWebPresenter)

		result := parseBisleyOutput(t, p.Output(bg, nil))
		for _, col := range result.Tableau {
			for _, tc := range col {
				assert.True(t, tc.FaceUp)
				assert.NotNil(t, tc.Card)
			}
		}
	})

	t.Run("error message", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		setupBisleyOutputMock(bg)
		p := new(BisleyWebPresenter)

		result := parseBisleyOutput(t, p.Output(bg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		setupBisleyOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.BisleyPhaseGameClear)

		p := new(BisleyWebPresenter)
		result := parseBisleyOutput(t, p.Output(bg, nil))
		assert.Equal(t, "bisley.gameClear", result.MessageCode)
		assert.Equal(t, "0", result.MessageParams["moveCount"])
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		setupBisleyOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.BisleyPhaseGameOver)

		p := new(BisleyWebPresenter)
		result := parseBisleyOutput(t, p.Output(bg, nil))
		assert.Equal(t, "bisley.gameOver", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		setupBisleyOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)

		p := new(BisleyWebPresenter)
		result := parseBisleyOutput(t, p.Output(bg, nil))
		assert.Equal(t, "bisley.stalemate", result.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestBisleyWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.BisleyHint{FromCol: 2, ToZone: "tableau", ToIdx: 2}

	bg := new(interfaces.MockBisleyGame)
	setupBisleyWebMockDefaults(bg)
	bg.On("GetHint").Return(hint).Maybe()

	result := parseBisleyOutput(t, new(BisleyWebPresenter).Output(bg, nil))
	if result.Hint == nil {
		t.Fatal("Output must carry the hint -- the frontend reads state.hint")
	}
	assert.Equal(t, 2, result.Hint.FromCol)
}

func TestBisleyWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		setupBisleyWebMockDefaults(bg)
		bg.On("GetHint").Return(&domain.BisleyHint{FromCol: 0, ToZone: "king", ToIdx: 3})

		p := new(BisleyWebPresenter)
		result := parseBisleyOutput(t, p.HintOutput(bg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, 0, result.Hint.FromCol)
		assert.Equal(t, "king", result.Hint.ToZone)
		assert.Equal(t, 3, result.Hint.ToIdx)
		assert.Equal(t, "bisley.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		setupBisleyWebMockDefaults(bg)
		bg.On("GetHint").Return((*domain.BisleyHint)(nil))

		p := new(BisleyWebPresenter)
		result := parseBisleyOutput(t, p.HintOutput(bg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "bisley.noHint", result.MessageCode)
	})
}

func TestBisleyWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		bg.On("GetPhase").Return(domain.BisleyPhasePlaying)
		bg.On("GetGameEndFlag").Return(false)

		p := new(BisleyWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		bg.On("GetPhase").Return(domain.BisleyPhaseGameOver)
		bg.On("GetGameEndFlag").Return(true)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(BisleyWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "move")
	})
}

// #5553: ドメインのエラーが i18n キーを名乗るようになった。Error() をそのまま
// Message に入れると、画面に "bisley.errEmptyColumn" が出る。
func TestBisleyWebPresenter_Output_CodedErrorGoesOutAsACode(t *testing.T) {
	bg := new(interfaces.MockBisleyGame)
	setupBisleyOutputMock(bg)

	err := domain.NewDomainErrorCode(domain.ErrInvalidPlay, "bisley.errEmptyColumn", nil)
	var out controller.BisleyWebOutput
	require.NoError(t, json.Unmarshal([]byte(new(BisleyWebPresenter).Output(bg, err)), &out))

	assert.Equal(t, "bisley.errEmptyColumn", out.MessageCode)
	assert.Empty(t, out.Message, "生のキーを message に入れない")
}
