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

func setupFourteenOutWebMockDefaults(g *interfaces.MockFourteenOutGame) {
	g.On("GetPhase").Return(domain.FourteenOutPhasePlaying).Maybe()
	g.On("GetRemovedCount").Return(0).Maybe()
	g.On("CountRemovablePairs").Return(3).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("GetColumns").Return(foCuiColumns()).Maybe()
}

func parseFourteenOutOutput(t *testing.T, s string) *controller.FourteenOutWebOutput {
	t.Helper()
	var out controller.FourteenOutWebOutput
	err := json.Unmarshal([]byte(s), &out)
	assert.NoError(t, err)
	return &out
}

// setupFourteenOutOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので Hint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupFourteenOutOutputMock(g *interfaces.MockFourteenOutGame) {
	setupFourteenOutWebMockDefaults(g)
	g.On("Hint").Return(nil).Maybe()
}

func TestFourteenOutWebPresenter_Output_Playing(t *testing.T) {
	g := new(interfaces.MockFourteenOutGame)
	setupFourteenOutOutputMock(g)
	p := &FourteenOutWebPresenter{}
	out := parseFourteenOutOutput(t, p.Output(g, nil))

	assert.Equal(t, 0, out.Phase)
	assert.Equal(t, "fourteenout.playing", out.MessageCode)
	assert.Len(t, out.Columns, domain.FourteenOutColumnCnt)
	assert.Equal(t, 3, out.RemovablePairs)
}

func TestFourteenOutWebPresenter_Output_PlayingStalemate(t *testing.T) {
	g := new(interfaces.MockFourteenOutGame)
	g.On("GetPhase").Return(domain.FourteenOutPhasePlaying).Maybe()
	g.On("GetRemovedCount").Return(40).Maybe()
	g.On("CanUndo").Return(true).Maybe()
	g.On("IsStalemate").Return(true).Maybe()
	g.On("CountRemovablePairs").Return(3).Maybe()
	g.On("GetColumns").Return(foCuiColumns()).Maybe()
	g.On("Hint").Return(nil).Maybe()

	p := &FourteenOutWebPresenter{}
	out := parseFourteenOutOutput(t, p.Output(g, nil))
	assert.Equal(t, "fourteenout.stalemate", out.MessageCode)
	assert.True(t, out.IsStalemate)
}

func TestFourteenOutWebPresenter_Output_GameClear(t *testing.T) {
	g := new(interfaces.MockFourteenOutGame)
	g.On("GetPhase").Return(domain.FourteenOutPhaseGameClear).Maybe()
	g.On("GetRemovedCount").Return(52).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("CountRemovablePairs").Return(3).Maybe()
	g.On("GetColumns").Return(foCuiColumns()).Maybe()
	g.On("Hint").Return(nil).Maybe()

	p := &FourteenOutWebPresenter{}
	out := parseFourteenOutOutput(t, p.Output(g, nil))
	assert.Equal(t, "fourteenout.gameClear", out.MessageCode)
	assert.Equal(t, "52", out.MessageParams["removedCount"])
	// **補充回数は存在しない。**クローン元の Monte Carlo は山札から補充できる
	// ので dealCount を出していたが、Fourteen Out に補充は無い。
	assert.NotContains(t, out.MessageParams, "dealCount")
}

func TestFourteenOutWebPresenter_Output_GameOver(t *testing.T) {
	g := new(interfaces.MockFourteenOutGame)
	g.On("GetPhase").Return(domain.FourteenOutPhaseGameOver).Maybe()
	g.On("GetRemovedCount").Return(20).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("CountRemovablePairs").Return(3).Maybe()
	g.On("GetColumns").Return(foCuiColumns()).Maybe()
	g.On("Hint").Return(nil).Maybe()

	p := &FourteenOutWebPresenter{}
	out := parseFourteenOutOutput(t, p.Output(g, nil))
	assert.Equal(t, "fourteenout.gameOver", out.MessageCode)
}

func TestFourteenOutWebPresenter_Output_Error(t *testing.T) {
	g := new(interfaces.MockFourteenOutGame)
	setupFourteenOutOutputMock(g)
	p := &FourteenOutWebPresenter{}
	out := parseFourteenOutOutput(t, p.Output(g, errors.New("boom")))
	assert.Equal(t, "boom", out.Message)
}

// **列は長さがまちまちで、空列もある。**固定グリッドではないので、出力も
// 列ごとに実際の枚数を持たなければならない。
func TestFourteenOutWebPresenter_Output_SerialisesRaggedColumns(t *testing.T) {
	g := new(interfaces.MockFourteenOutGame)
	g.On("GetPhase").Return(domain.FourteenOutPhasePlaying).Maybe()
	g.On("GetRemovedCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("CountRemovablePairs").Return(3).Maybe()
	g.On("GetColumns").Return(foCuiColumns(
		[]*domain.Card{foCuiCard(domain.CardDesignSpade, 3), foCuiCard(domain.CardDesignSpade, 9)},
		[]*domain.Card{foCuiCard(domain.CardDesignHeart, 5)},
	)).Maybe()
	g.On("Hint").Return(nil).Maybe()

	p := &FourteenOutWebPresenter{}
	out := parseFourteenOutOutput(t, p.Output(g, nil))
	require.Len(t, out.Columns, domain.FourteenOutColumnCnt)
	require.Len(t, out.Columns[0], 2, "column 0 keeps both cards")
	require.Len(t, out.Columns[1], 1)
	assert.Empty(t, out.Columns[2], "an emptied column serialises as empty, not as a padded row")
	assert.NotNil(t, out.Columns[0][1].Card)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestFourteenOutWebPresenterOutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		mcg := new(interfaces.MockFourteenOutGame)
		setupFourteenOutWebMockDefaults(mcg)
		mcg.On("Hint").Return(&domain.FourteenOutHint{Action: "remove", FromCol: 1, ToCol: 2}).Maybe()

		result := new(FourteenOutWebPresenter).Output(mcg, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	t.Run("not while stalemate", func(t *testing.T) {
		mcg := new(interfaces.MockFourteenOutGame)
		setupFourteenOutWebMockDefaults(mcg)
		mcg.ExpectedCalls = filterCalls(mcg.ExpectedCalls, "IsStalemate")
		mcg.On("IsStalemate").Return(true)
		mcg.On("Hint").Return(&domain.FourteenOutHint{Action: "remove", FromCol: 1, ToCol: 2}).Maybe()

		result := new(FourteenOutWebPresenter).Output(mcg, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestFourteenOutWebPresenter_HintOutput_Remove(t *testing.T) {
	g := new(interfaces.MockFourteenOutGame)
	setupFourteenOutWebMockDefaults(g)
	g.On("Hint").Return(&domain.FourteenOutHint{
		Action:  domain.FourteenOutHintActionRemove,
		FromCol: 1, ToCol: 2,
	}).Maybe()

	p := &FourteenOutWebPresenter{}
	out := parseFourteenOutOutput(t, p.HintOutput(g))
	assert.Equal(t, "fourteenout.hintAvailable", out.MessageCode)
	assert.NotNil(t, out.Hint)
	assert.Equal(t, domain.FourteenOutHintActionRemove, out.Hint.Action)
	assert.Equal(t, 1, out.Hint.FromCol)
	assert.Equal(t, 2, out.Hint.ToCol)
}

func TestFourteenOutWebPresenter_HintOutput_None(t *testing.T) {
	g := new(interfaces.MockFourteenOutGame)
	setupFourteenOutWebMockDefaults(g)
	g.On("Hint").Return((*domain.FourteenOutHint)(nil)).Maybe()

	p := &FourteenOutWebPresenter{}
	out := parseFourteenOutOutput(t, p.HintOutput(g))
	assert.Equal(t, "fourteenout.noHint", out.MessageCode)
	assert.Nil(t, out.Hint)
}

func TestFourteenOutWebPresenter_ActionLog_Playing(t *testing.T) {
	g := new(interfaces.MockFourteenOutGame)
	g.On("GetPhase").Return(domain.FourteenOutPhasePlaying)
	g.On("GetGameEndFlag").Return(false)
	p := &FourteenOutWebPresenter{}
	result := p.ActionLogOutput(g)
	assert.Contains(t, result, "entries")
}

func TestFourteenOutWebPresenter_ActionLog_GameOver(t *testing.T) {
	g := new(interfaces.MockFourteenOutGame)
	g.On("GetPhase").Return(domain.FourteenOutPhaseGameOver)
	g.On("GetGameEndFlag").Return(true)
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, ActionType: "remove", Detail: "test"},
	})
	p := &FourteenOutWebPresenter{}
	result := p.ActionLogOutput(g)
	assert.Contains(t, result, "remove")
}
