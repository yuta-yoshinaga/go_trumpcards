//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupAmericanToadWebMockDefaults(g *interfaces.MockAmericanToadGame) {
	g.On("LegalTargetsFrom", mock.Anything, mock.Anything).Return([]int{}, []int{}).Maybe()
	g.On("ReserveLegalTargets").Return([]int{}, []int{}).Maybe()
	g.On("WasteLegalTargets").Return([]int{}, []int{}).Maybe()
	g.On("GetPhase").Return(domain.AmericanToadPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(75).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()
	g.On("GetReserve").Return([]*domain.Card{domain.NewCard(domain.CardDesignClover, 3, true)}).Maybe()
	g.On("GetBaseRank").Return(5).Maybe()
	g.On("GetPassesUsed").Return(0).Maybe()
	g.On("CanRedeal").Return(false).Maybe()

	var tableau [domain.AmericanToadTableauCnt][]*domain.AmericanToadTableauCard
	for i := range domain.AmericanToadTableauCnt {
		tableau[i] = []*domain.AmericanToadTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, i+2, false), FaceUp: true},
		}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.AmericanToadFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func parseAmericanToadOutput(t *testing.T, jsonStr string) *controller.AmericanToadWebOutput {
	t.Helper()
	var out controller.AmericanToadWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

// setupAmericanToadOutputMock は Output 用の既定を組む。
//
// **Output() も受動ヒントを埋めるようになった** (#4483) ので、GetHint を
// 呼べるようにしておく必要がある。共有ヘルパー側に置くと、先に登録された
// この期待が HintOutput テストの「ヒントあり」を食ってしまう。
func setupAmericanToadOutputMock(g *interfaces.MockAmericanToadGame) {
	setupAmericanToadWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestAmericanToadWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadOutputMock(g)

		result := parseAmericanToadOutput(t, new(AmericanToadWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 75, result.StockCount)
		assert.Len(t, result.Reserve, 1)
		assert.Len(t, result.Tableau, domain.AmericanToadTableauCnt)
		assert.Len(t, result.Foundation, domain.AmericanToadFoundationCnt)
		assert.Len(t, result.Waste, 1)
		assert.Equal(t, 5, result.BaseRank)
		assert.Equal(t, 0, result.PassesUsed)
		assert.False(t, result.CanRedeal)
		assert.Equal(t, "americantoad.playing", result.MessageCode)
	})

	// There is only one redeal, so the client must be told when it is on offer
	// rather than inferring it from an empty stock and a non-empty waste.
	t.Run("an available redeal is surfaced", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "CanRedeal")
		g.On("CanRedeal").Return(true)

		result := parseAmericanToadOutput(t, new(AmericanToadWebPresenter).Output(g, nil))
		assert.True(t, result.CanRedeal)
		assert.Equal(t, "americantoad.redealAvailable", result.MessageCode)
	})

	t.Run("stalemate outranks the redeal notice", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "CanRedeal")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("CanRedeal").Return(true)
		g.On("IsStalemate").Return(true)

		result := parseAmericanToadOutput(t, new(AmericanToadWebPresenter).Output(g, nil))
		assert.Equal(t, "americantoad.stalemate", result.MessageCode)
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadOutputMock(g)

		result := parseAmericanToadOutput(t, new(AmericanToadWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	for _, tc := range []struct {
		name string
		val  domain.AmericanToadPhase
		code string
	}{
		{"game clear", domain.AmericanToadPhaseGameClear, "americantoad.gameClear"},
		{"game over", domain.AmericanToadPhaseGameOver, "americantoad.gameOver"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockAmericanToadGame)
			setupAmericanToadOutputMock(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			result := parseAmericanToadOutput(t, new(AmericanToadWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない。ここが埋まっていないと
// フロントの `state.hint` は常に undefined で、それを読む分岐が全部死ぬ (#4483)。
func TestAmericanToadWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.AmericanToadHint{FromZone: "tableau", FromIdx: 2, CardIndex: 3, ToZone: "foundation", ToIdx: 1}

	g := new(interfaces.MockAmericanToadGame)
	setupAmericanToadWebMockDefaults(g)
	g.On("GetHint").Return(hint).Maybe()

	result := parseAmericanToadOutput(t, new(AmericanToadWebPresenter).Output(g, nil))
	if result.Hint == nil {
		t.Fatal("Output must carry the hint -- the frontend reads state.hint")
	}
	assert.Equal(t, "tableau", result.Hint.FromZone)
	assert.Equal(t, 2, result.Hint.FromIdx)
	assert.Equal(t, 3, result.Hint.CardIndex)

	// **手詰まりでは探索を走らせない。**ゲートを消したときに CI が捕まえるよう、
	// 手元の負のコントロールではなくテストとして固定する。
	t.Run("not while stalemate", func(t *testing.T) {
		g2 := new(interfaces.MockAmericanToadGame)
		setupAmericanToadWebMockDefaults(g2)
		g2.ExpectedCalls = filterCalls(g2.ExpectedCalls, "IsStalemate")
		g2.On("IsStalemate").Return(true).Maybe()
		g2.On("GetHint").Return(hint).Maybe()

		result := parseAmericanToadOutput(t, new(AmericanToadWebPresenter).Output(g2, nil))
		assert.Nil(t, result.Hint)
		g2.AssertNotCalled(t, "GetHint")
	})
}

func TestAmericanToadWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadWebMockDefaults(g)
		g.On("GetHint").Return(&domain.AmericanToadHint{
			FromZone: "tableau", FromIdx: 2, CardIndex: 1, ToZone: "foundation", ToIdx: 3,
		})

		result := parseAmericanToadOutput(t, new(AmericanToadWebPresenter).HintOutput(g))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "tableau", result.Hint.FromZone)
		assert.Equal(t, 2, result.Hint.FromIdx)
		assert.Equal(t, 1, result.Hint.CardIndex)
		assert.Equal(t, 3, result.Hint.ToIdx)
		assert.Equal(t, "americantoad.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadWebMockDefaults(g)
		g.On("GetHint").Return((*domain.AmericanToadHint)(nil))

		result := parseAmericanToadOutput(t, new(AmericanToadWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "americantoad.noHint", result.MessageCode)
	})
}

func TestAmericanToadWebPresenter_LegalTargets(t *testing.T) {
	setup := func() *interfaces.MockAmericanToadGame {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "LegalTargetsFrom")
		g.On("LegalTargetsFrom", 0, 0).Return([]int{2}, []int(nil)).Maybe()
		// Other non-empty tableau columns have no legal targets in this fixture.
		g.On("LegalTargetsFrom", mock.Anything, mock.Anything).Return([]int(nil), []int(nil)).Maybe()
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "ReserveLegalTargets")
		g.On("ReserveLegalTargets").Return([]int(nil), []int{1, 5}).Maybe()
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "WasteLegalTargets")
		g.On("WasteLegalTargets").Return([]int{3}, []int(nil)).Maybe()
		g.On("GetHint").Return((*domain.AmericanToadHint)(nil)).Maybe()
		return g
	}
	p := new(AmericanToadWebPresenter)
	for _, output := range []string{p.Output(setup(), nil), p.HintOutput(setup())} {
		got := parseAmericanToadOutput(t, output)
		assert.Equal(t, []int{2}, got.LegalTargets["tableau-0-0"].Tableau)
		assert.Empty(t, got.LegalTargets["tableau-0-0"].Foundation)
		assert.Equal(t, []int{1, 5}, got.LegalTargets["reserve"].Foundation)
		assert.Empty(t, got.LegalTargets["reserve"].Tableau)
		assert.Equal(t, []int{3}, got.LegalTargets["waste"].Tableau)
		assert.Empty(t, got.LegalTargets["waste"].Foundation)
	}

	// Exercise nil-array normalization for the alternate reserve and waste shapes.
	alternateNilTargets := new(interfaces.MockAmericanToadGame)
	setupAmericanToadWebMockDefaults(alternateNilTargets)
	alternateNilTargets.ExpectedCalls = filterCalls(alternateNilTargets.ExpectedCalls, "ReserveLegalTargets")
	alternateNilTargets.On("ReserveLegalTargets").Return([]int{4}, []int(nil)).Maybe()
	alternateNilTargets.ExpectedCalls = filterCalls(alternateNilTargets.ExpectedCalls, "WasteLegalTargets")
	alternateNilTargets.On("WasteLegalTargets").Return([]int(nil), []int{6}).Maybe()
	alternateNilTargets.On("GetHint").Return((*domain.AmericanToadHint)(nil)).Maybe()
	alternate := parseAmericanToadOutput(t, p.Output(alternateNilTargets, nil))
	assert.Equal(t, []int{4}, alternate.LegalTargets["reserve"].Tableau)
	assert.Empty(t, alternate.LegalTargets["reserve"].Foundation)
	assert.Empty(t, alternate.LegalTargets["waste"].Tableau)
	assert.Equal(t, []int{6}, alternate.LegalTargets["waste"].Foundation)

	nilTableauTargets := setup()
	nilTableauTargets.ExpectedCalls = filterCalls(nilTableauTargets.ExpectedCalls, "LegalTargetsFrom")
	nilTableauTargets.On("LegalTargetsFrom", 0, 0).Return([]int(nil), []int{7}).Maybe()
	nilTableauTargets.On("LegalTargetsFrom", mock.Anything, mock.Anything).Return([]int(nil), []int(nil)).Maybe()
	nilTableau := parseAmericanToadOutput(t, p.Output(nilTableauTargets, nil))
	assert.Empty(t, nilTableau.LegalTargets["tableau-0-0"].Tableau)
	assert.Equal(t, []int{7}, nilTableau.LegalTargets["tableau-0-0"].Foundation)

	empty := new(interfaces.MockAmericanToadGame)
	setupAmericanToadWebMockDefaults(empty)
	empty.ExpectedCalls = filterCalls(empty.ExpectedCalls, "LegalTargetsFrom")
	empty.ExpectedCalls = filterCalls(empty.ExpectedCalls, "ReserveLegalTargets")
	empty.ExpectedCalls = filterCalls(empty.ExpectedCalls, "WasteLegalTargets")
	empty.On("LegalTargetsFrom", mock.Anything, mock.Anything).Return([]int(nil), []int(nil)).Maybe()
	empty.On("ReserveLegalTargets").Return([]int(nil), []int(nil)).Maybe()
	empty.On("WasteLegalTargets").Return([]int(nil), []int(nil)).Maybe()
	empty.On("GetHint").Return((*domain.AmericanToadHint)(nil)).Maybe()
	assert.Empty(t, parseAmericanToadOutput(t, p.Output(empty, nil)).LegalTargets)
}

func TestAmericanToadWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		g.On("GetPhase").Return(domain.AmericanToadPhasePlaying)
		g.On("GetGameEndFlag").Return(false)

		assert.Contains(t, new(AmericanToadWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		g.On("GetPhase").Return(domain.AmericanToadPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", DetailCode: "americantoad.log.move", DetailParams: map[string]string{
				"value1": "0", "value2": "0", "value3": "1",
			}},
		})

		assert.Contains(t, new(AmericanToadWebPresenter).ActionLogOutput(g), "move")
	})
}
