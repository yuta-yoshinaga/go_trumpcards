//go:build test

package presenter

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupPerseveranceCuiMockDefaults(bg *interfaces.MockPerseveranceGame) {
	bg.On("GetPhase").Return(domain.PerseverancePhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
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

func TestPerseveranceCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		setupPerseveranceCuiMockDefaults(bg)
		p := new(PerseveranceCuiPresenter)

		result := p.Output(bg, nil)
		assert.Contains(t, result, "Perseverance")
		assert.Contains(t, result, "Foundation")
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "手数: 0")
		// Playing phase surfaces the empty-column caveat; no column is at 1 card.
		assert.Contains(t, result, "空列は再利用できません")
		assert.NotContains(t, result, "残り1枚")
	})

	t.Run("single-card column is flagged with a warning marker", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		setupPerseveranceCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetTableau")
		var tableau [domain.PerseveranceTableauCnt][]*domain.PerseveranceTableauCard
		// Column 0 has a single card (at-risk); the rest keep four.
		tableau[0] = []*domain.PerseveranceTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, 5, false), FaceUp: true},
		}
		for i := 1; i < domain.PerseveranceTableauCnt; i++ {
			tableau[i] = make([]*domain.PerseveranceTableauCard, 4)
			for j := range 4 {
				tableau[i][j] = &domain.PerseveranceTableauCard{
					Card:   domain.NewCard(domain.CardDesignHeart, j+1, false),
					FaceUp: true,
				}
			}
		}
		bg.On("GetTableau").Return(tableau)
		p := new(PerseveranceCuiPresenter)

		result := p.Output(bg, nil)
		assert.Contains(t, result, "残り1枚")
	})

	t.Run("with error", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		setupPerseveranceCuiMockDefaults(bg)
		p := new(PerseveranceCuiPresenter)

		result := p.Output(bg, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		setupPerseveranceCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.PerseverancePhaseGameClear)

		p := new(PerseveranceCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		setupPerseveranceCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.PerseverancePhaseGameOver)

		p := new(PerseveranceCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})

	t.Run("stalemate", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		setupPerseveranceCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)
		bg.On("UndoToEscape").Return(0).Maybe()

		p := new(PerseveranceCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "手詰まり")
	})

	t.Run("empty column", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		setupPerseveranceCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetTableau")
		var emptyTableau [domain.PerseveranceTableauCnt][]*domain.PerseveranceTableauCard
		bg.On("GetTableau").Return(emptyTableau)

		p := new(PerseveranceCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("foundation with cards", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		setupPerseveranceCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetFoundation")
		var foundation [domain.PerseveranceFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		bg.On("GetFoundation").Return(foundation)

		p := new(PerseveranceCuiPresenter)
		result := p.Output(bg, nil)
		assert.NotEmpty(t, result)
	})
}

func TestPerseveranceCuiPresenter_HintOutput(t *testing.T) {
	t.Run("foundation hint", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		bg.On("GetHint").Return(&domain.PerseveranceHint{
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(PerseveranceCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "タブロー列0")
		assert.Contains(t, result, "ファンデーション")
	})

	t.Run("tableau hint", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		bg.On("GetHint").Return(&domain.PerseveranceHint{
			FromCol:   1,
			CardIndex: 0,
			ToZone:    "tableau",
			ToCol:     3,
		})

		p := new(PerseveranceCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "タブロー列1")
		assert.Contains(t, result, "タブロー列3")
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		bg.On("GetHint").Return((*domain.PerseveranceHint)(nil))

		p := new(PerseveranceCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestPerseveranceCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		bg.On("GetGameEndFlag").Return(false)

		p := new(PerseveranceCuiPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		bg := new(interfaces.MockPerseveranceGame)
		bg.On("GetGameEndFlag").Return(true)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(PerseveranceCuiPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "move")
	})
}

// #5581: 置ける先が無いときに黙ると、コマンドが効いていないのか置けないのかが
// 区別できない。
func TestPerseveranceCuiPresenter_TargetsOutput(t *testing.T) {
	i18n.SetLang("ja")

	t.Run("lists the columns and foundations", func(t *testing.T) {
		g := new(interfaces.MockPerseveranceGame)
		g.On("LegalTargets", 3).Return([]int{1, 7}, []int{2})

		out := new(PerseveranceCuiPresenter).TargetsOutput(g, 3)
		assert.Contains(t, out, i18n.Tf("perseverance.targetTableau", "col", "1"))
		assert.Contains(t, out, i18n.Tf("perseverance.targetTableau", "col", "7"))
		assert.Contains(t, out, i18n.Tf("perseverance.targetFoundation", "idx", "2"))
	})

	t.Run("says so when there is nowhere to go", func(t *testing.T) {
		g := new(interfaces.MockPerseveranceGame)
		g.On("LegalTargets", 3).Return([]int(nil), []int(nil))

		out := new(PerseveranceCuiPresenter).TargetsOutput(g, 3)
		assert.Contains(t, out, i18n.Tf("perseverance.targetsNone", "col", "3"))
		// 空行で済ませないこと。
		assert.NotEmpty(t, strings.TrimSpace(out))
	})

	// **範囲外はドメインに訊く前に断る。**訊いても nil が返るだけだが、
	// 「置ける先がありません」と答えると、存在しない列があるように読める。
	t.Run("rejects a column that does not exist", func(t *testing.T) {
		g := new(interfaces.MockPerseveranceGame)
		out := new(PerseveranceCuiPresenter).TargetsOutput(g, domain.PerseveranceTableauCnt)
		assert.Contains(t, out, i18n.Tf("invalidColumn", "val", strconv.Itoa(domain.PerseveranceTableauCnt)))
		assert.NotContains(t, out, i18n.Tf("perseverance.targetsNone", "col", strconv.Itoa(domain.PerseveranceTableauCnt)))
		g.AssertNotCalled(t, "LegalTargets", mock.Anything)
	})
}
