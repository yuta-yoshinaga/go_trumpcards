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

func setupMonteCarloWebMockDefaults(g *interfaces.MockMonteCarloGame) {
	g.On("GetPhase").Return(domain.MonteCarloPhasePlaying).Maybe()
	g.On("GetStockCount").Return(27).Maybe()
	g.On("GetRemovedCount").Return(0).Maybe()
	g.On("GetDealCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	var board [domain.MonteCarloGridSize][domain.MonteCarloGridSize]*domain.Card
	g.On("GetBoard").Return(board).Maybe()
}

func parseMonteCarloOutput(t *testing.T, s string) *controller.MonteCarloWebOutput {
	t.Helper()
	var out controller.MonteCarloWebOutput
	err := json.Unmarshal([]byte(s), &out)
	assert.NoError(t, err)
	return &out
}

// setupMonteCarloOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので Hint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupMonteCarloOutputMock(g *interfaces.MockMonteCarloGame) {
	setupMonteCarloWebMockDefaults(g)
	g.On("Hint").Return(nil).Maybe()
}

func TestMonteCarloWebPresenter_Output_Playing(t *testing.T) {
	g := new(interfaces.MockMonteCarloGame)
	setupMonteCarloOutputMock(g)
	p := &MonteCarloWebPresenter{}
	out := parseMonteCarloOutput(t, p.Output(g, nil))

	assert.Equal(t, 0, out.Phase)
	assert.Equal(t, "montecarlo.playing", out.MessageCode)
	assert.Len(t, out.Board, domain.MonteCarloGridSize)
	assert.Equal(t, 27, out.StockCount)
}

func TestMonteCarloWebPresenter_Output_PlayingStalemate(t *testing.T) {
	g := new(interfaces.MockMonteCarloGame)
	g.On("GetPhase").Return(domain.MonteCarloPhasePlaying).Maybe()
	g.On("GetStockCount").Return(0).Maybe()
	g.On("GetRemovedCount").Return(40).Maybe()
	g.On("GetDealCount").Return(3).Maybe()
	g.On("CanUndo").Return(true).Maybe()
	g.On("IsStalemate").Return(true).Maybe()
	var board [domain.MonteCarloGridSize][domain.MonteCarloGridSize]*domain.Card
	g.On("GetBoard").Return(board).Maybe()
	g.On("Hint").Return(nil).Maybe()

	p := &MonteCarloWebPresenter{}
	out := parseMonteCarloOutput(t, p.Output(g, nil))
	assert.Equal(t, "montecarlo.stalemate", out.MessageCode)
	assert.True(t, out.IsStalemate)
}

func TestMonteCarloWebPresenter_Output_GameClear(t *testing.T) {
	g := new(interfaces.MockMonteCarloGame)
	g.On("GetPhase").Return(domain.MonteCarloPhaseGameClear).Maybe()
	g.On("GetStockCount").Return(0).Maybe()
	g.On("GetRemovedCount").Return(52).Maybe()
	g.On("GetDealCount").Return(5).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	var board [domain.MonteCarloGridSize][domain.MonteCarloGridSize]*domain.Card
	g.On("GetBoard").Return(board).Maybe()
	g.On("Hint").Return(nil).Maybe()

	p := &MonteCarloWebPresenter{}
	out := parseMonteCarloOutput(t, p.Output(g, nil))
	assert.Equal(t, "montecarlo.gameClear", out.MessageCode)
	assert.Equal(t, "5", out.MessageParams["dealCount"])
	assert.Equal(t, "52", out.MessageParams["removedCount"])
}

func TestMonteCarloWebPresenter_Output_GameOver(t *testing.T) {
	g := new(interfaces.MockMonteCarloGame)
	g.On("GetPhase").Return(domain.MonteCarloPhaseGameOver).Maybe()
	g.On("GetStockCount").Return(0).Maybe()
	g.On("GetRemovedCount").Return(20).Maybe()
	g.On("GetDealCount").Return(2).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	var board [domain.MonteCarloGridSize][domain.MonteCarloGridSize]*domain.Card
	g.On("GetBoard").Return(board).Maybe()
	g.On("Hint").Return(nil).Maybe()

	p := &MonteCarloWebPresenter{}
	out := parseMonteCarloOutput(t, p.Output(g, nil))
	assert.Equal(t, "montecarlo.gameOver", out.MessageCode)
}

func TestMonteCarloWebPresenter_Output_Error(t *testing.T) {
	g := new(interfaces.MockMonteCarloGame)
	setupMonteCarloOutputMock(g)
	p := &MonteCarloWebPresenter{}
	out := parseMonteCarloOutput(t, p.Output(g, errors.New("boom")))
	assert.Equal(t, "boom", out.Message)
}

func TestMonteCarloWebPresenter_Output_BoardWithCard(t *testing.T) {
	g := new(interfaces.MockMonteCarloGame)
	g.On("GetPhase").Return(domain.MonteCarloPhasePlaying).Maybe()
	g.On("GetStockCount").Return(27).Maybe()
	g.On("GetRemovedCount").Return(0).Maybe()
	g.On("GetDealCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	var board [domain.MonteCarloGridSize][domain.MonteCarloGridSize]*domain.Card
	board[0][0] = domain.NewCard(domain.CardDesignSpade, 7, false)
	g.On("GetBoard").Return(board).Maybe()
	g.On("Hint").Return(nil).Maybe()

	p := &MonteCarloWebPresenter{}
	out := parseMonteCarloOutput(t, p.Output(g, nil))
	assert.NotNil(t, out.Board[0][0].Card)
	assert.Nil(t, out.Board[0][1].Card)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestMonteCarloWebPresenterOutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		mcg := new(interfaces.MockMonteCarloGame)
		setupMonteCarloWebMockDefaults(mcg)
		mcg.On("Hint").Return(&domain.MonteCarloHint{Action: "remove", FromR: 0, FromC: 1, ToR: 1, ToC: 2}).Maybe()

		result := new(MonteCarloWebPresenter).Output(mcg, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	t.Run("not while stalemate", func(t *testing.T) {
		mcg := new(interfaces.MockMonteCarloGame)
		setupMonteCarloWebMockDefaults(mcg)
		mcg.ExpectedCalls = filterCalls(mcg.ExpectedCalls, "IsStalemate")
		mcg.On("IsStalemate").Return(true)
		mcg.On("Hint").Return(&domain.MonteCarloHint{Action: "remove", FromR: 0, FromC: 1, ToR: 1, ToC: 2}).Maybe()

		result := new(MonteCarloWebPresenter).Output(mcg, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestMonteCarloWebPresenter_HintOutput_Remove(t *testing.T) {
	g := new(interfaces.MockMonteCarloGame)
	setupMonteCarloWebMockDefaults(g)
	g.On("Hint").Return(&domain.MonteCarloHint{
		Action: domain.MonteCarloHintActionRemove,
		FromR:  0, FromC: 1,
		ToR: 1, ToC: 2,
	}).Maybe()

	p := &MonteCarloWebPresenter{}
	out := parseMonteCarloOutput(t, p.HintOutput(g))
	assert.Equal(t, "montecarlo.hintAvailable", out.MessageCode)
	assert.NotNil(t, out.Hint)
	assert.Equal(t, domain.MonteCarloHintActionRemove, out.Hint.Action)
	assert.Equal(t, 1, out.Hint.FromC)
	assert.Equal(t, 2, out.Hint.ToC)
}

func TestMonteCarloWebPresenter_HintOutput_Deal(t *testing.T) {
	g := new(interfaces.MockMonteCarloGame)
	setupMonteCarloWebMockDefaults(g)
	g.On("Hint").Return(&domain.MonteCarloHint{Action: domain.MonteCarloHintActionDeal}).Maybe()

	p := &MonteCarloWebPresenter{}
	out := parseMonteCarloOutput(t, p.HintOutput(g))
	assert.Equal(t, "montecarlo.hintAvailable", out.MessageCode)
	assert.NotNil(t, out.Hint)
	assert.Equal(t, domain.MonteCarloHintActionDeal, out.Hint.Action)
}

func TestMonteCarloWebPresenter_HintOutput_None(t *testing.T) {
	g := new(interfaces.MockMonteCarloGame)
	setupMonteCarloWebMockDefaults(g)
	g.On("Hint").Return((*domain.MonteCarloHint)(nil)).Maybe()

	p := &MonteCarloWebPresenter{}
	out := parseMonteCarloOutput(t, p.HintOutput(g))
	assert.Equal(t, "montecarlo.noHint", out.MessageCode)
	assert.Nil(t, out.Hint)
}

func TestMonteCarloWebPresenter_ActionLog_Playing(t *testing.T) {
	g := new(interfaces.MockMonteCarloGame)
	g.On("GetPhase").Return(domain.MonteCarloPhasePlaying)
	g.On("GetGameEndFlag").Return(false)
	p := &MonteCarloWebPresenter{}
	result := p.ActionLogOutput(g)
	assert.Contains(t, result, "entries")
}

func TestMonteCarloWebPresenter_ActionLog_GameOver(t *testing.T) {
	g := new(interfaces.MockMonteCarloGame)
	g.On("GetPhase").Return(domain.MonteCarloPhaseGameOver)
	g.On("GetGameEndFlag").Return(true)
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, ActionType: "remove", Detail: "test"},
	})
	p := &MonteCarloWebPresenter{}
	result := p.ActionLogOutput(g)
	assert.Contains(t, result, "remove")
}
