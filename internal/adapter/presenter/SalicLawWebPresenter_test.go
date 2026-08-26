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

func setupSalicLawWebMockDefaults(g *interfaces.MockSalicLawGame) {
	g.On("GetPhase").Return(domain.SalicLawPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(96).Maybe()
	g.On("GetOpenPiles").Return(domain.SalicLawTableauCnt).Maybe()
	queens := make([]*domain.Card, domain.SalicLawQueenCnt)
	for i := range queens {
		queens[i] = domain.NewCard(i%4+1, domain.SalicLawQueenValue, true)
	}
	g.On("GetQueens").Return(queens).Maybe()

	var tableau [domain.SalicLawTableauCnt][]*domain.Card
	for i := range domain.SalicLawTableauCnt {
		tableau[i] = []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, domain.CardValueMax, true),
			domain.NewCard(domain.CardDesignSpade, i+2, true),
		}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.SalicLawFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func parseSalicLawOutput(t *testing.T, jsonStr string) *controller.SalicLawWebOutput {
	t.Helper()
	var out controller.SalicLawWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

// setupSalicLawOutputMock は Output 用の既定を組む。
//
// **Output() も受動ヒントを埋めるようになった** (#4483) ので、GetHint を
// 呼べるようにしておく必要がある。共有ヘルパー側に置くと、先に登録された
// この期待が HintOutput テストの「ヒントあり」を食ってしまう。
func setupSalicLawOutputMock(g *interfaces.MockSalicLawGame) {
	setupSalicLawWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestSalicLawWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockSalicLawGame)
		setupSalicLawOutputMock(g)

		result := parseSalicLawOutput(t, new(SalicLawWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 96, result.StockCount)
		assert.Len(t, result.Tableau, domain.SalicLawTableauCnt)
		assert.Len(t, result.Foundation, domain.SalicLawFoundationCnt)
		// **退場した Q は API にも載せる。**ページが「なぜ 8 枚無いのか」を
		// 説明できるのはこの一覧だけ。
		assert.Len(t, result.Queens, domain.SalicLawQueenCnt)
		assert.Equal(t, domain.SalicLawTableauCnt, result.OpenPiles)
		assert.Equal(t, "saliclaw.playing", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		g := new(interfaces.MockSalicLawGame)
		setupSalicLawOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)

		result := parseSalicLawOutput(t, new(SalicLawWebPresenter).Output(g, nil))
		assert.Equal(t, "saliclaw.stalemate", result.MessageCode)
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockSalicLawGame)
		setupSalicLawOutputMock(g)

		result := parseSalicLawOutput(t, new(SalicLawWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	for _, tc := range []struct {
		name string
		val  domain.SalicLawPhase
		code string
	}{
		{"game clear", domain.SalicLawPhaseGameClear, "saliclaw.gameClear"},
		{"game over", domain.SalicLawPhaseGameOver, "saliclaw.gameOver"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockSalicLawGame)
			setupSalicLawOutputMock(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			result := parseSalicLawOutput(t, new(SalicLawWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない。ここが埋まっていないと
// フロントの `state.hint` は常に undefined で、それを読む分岐が全部死ぬ (#4483)。
func TestSalicLawWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.SalicLawHint{FromZone: "tableau", FromIdx: 2, ToZone: "foundation", ToIdx: 1}

	t.Run("while the game is playable", func(t *testing.T) {
		g := new(interfaces.MockSalicLawGame)
		setupSalicLawWebMockDefaults(g)
		g.On("GetHint").Return(hint).Maybe()

		result := parseSalicLawOutput(t, new(SalicLawWebPresenter).Output(g, nil))
		if result.Hint == nil {
			t.Fatal("Output must carry the hint -- the frontend reads state.hint")
		}
		assert.Equal(t, "tableau", result.Hint.FromZone)
		assert.Equal(t, 2, result.Hint.FromIdx)
	})
}

func TestSalicLawWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockSalicLawGame)
		setupSalicLawWebMockDefaults(g)
		g.On("GetHint").Return(&domain.SalicLawHint{
			FromZone: "stock", FromIdx: -1, ToZone: "tableau", ToIdx: 3,
		})

		result := parseSalicLawOutput(t, new(SalicLawWebPresenter).HintOutput(g))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "stock", result.Hint.FromZone)
		assert.Equal(t, "tableau", result.Hint.ToZone)
		assert.Equal(t, 3, result.Hint.ToIdx)
		assert.Equal(t, "saliclaw.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockSalicLawGame)
		setupSalicLawWebMockDefaults(g)
		g.On("GetHint").Return((*domain.SalicLawHint)(nil))

		result := parseSalicLawOutput(t, new(SalicLawWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "saliclaw.noHint", result.MessageCode)
	})
}

func TestSalicLawWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockSalicLawGame)
		g.On("GetPhase").Return(domain.SalicLawPhasePlaying)
		g.On("GetGameEndFlag").Return(false)

		assert.Contains(t, new(SalicLawWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockSalicLawGame)
		g.On("GetPhase").Return(domain.SalicLawPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(SalicLawWebPresenter).ActionLogOutput(g), "move")
	})
}

// Web も同じエラーを受け取る。コードを Message に入れるとキー文字列が画面に
// 出るので、MessageCode に振り分けること (#5562)。
func TestSalicLawWebPresenter_ErrorCodesGoToMessageCode(t *testing.T) {
	g := new(interfaces.MockSalicLawGame)
	setupSalicLawOutputMock(g)
	err := domain.NewDomainErrorCode(domain.ErrInvalidPlay, "saliclaw.errKingIsTheBase", nil)

	var res controller.SalicLawWebOutput
	require.NoError(t, json.Unmarshal([]byte(new(SalicLawWebPresenter).Output(g, err)), &res))
	assert.Equal(t, "saliclaw.errKingIsTheBase", res.MessageCode)
	assert.Empty(t, res.Message)
}

// パラメータ付きのコードは値も渡すこと。落とすと "{{pile}}" が画面に出る。
func TestSalicLawWebPresenter_ErrorParamsSurvive(t *testing.T) {
	g := new(interfaces.MockSalicLawGame)
	setupSalicLawOutputMock(g)
	err := domain.NewDomainErrorCode(domain.ErrInvalidPlay, "saliclaw.errPileEmpty", map[string]string{"pile": "5"})

	var res controller.SalicLawWebOutput
	require.NoError(t, json.Unmarshal([]byte(new(SalicLawWebPresenter).Output(g, err)), &res))
	assert.Equal(t, map[string]string{"pile": "5"}, res.MessageParams)
}

// コードを持たないエラーは今までどおり Message に入ること。
func TestSalicLawWebPresenter_UncodedErrorsStayInMessage(t *testing.T) {
	g := new(interfaces.MockSalicLawGame)
	setupSalicLawOutputMock(g)

	var res controller.SalicLawWebOutput
	require.NoError(t, json.Unmarshal([]byte(new(SalicLawWebPresenter).Output(g, errors.New("boom"))), &res))
	assert.Equal(t, "boom", res.Message)
	assert.Empty(t, res.MessageCode)
}
