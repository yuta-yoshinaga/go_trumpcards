//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupKlondikeWebMockDefaults(kg *interfaces.MockKlondikeGame) {
	kg.On("GetPhase").Return(domain.KlondikePhasePlaying).Maybe()
	kg.On("GetMoveCount").Return(0).Maybe()
	kg.On("GetStockCount").Return(24).Maybe()
	kg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()

	var tableau [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
	for i := 0; i < domain.KlondikeTableauCnt; i++ {
		tableau[i] = make([]*domain.KlondikeTableauCard, 0)
		for j := 0; j <= i; j++ {
			tableau[i] = append(tableau[i], &domain.KlondikeTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+1, false),
				FaceUp: j == i,
			})
		}
	}
	kg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.KlondikeFoundationCnt][]*domain.Card
	kg.On("GetFoundation").Return(foundation).Maybe()
}

func parseKlondikeOutput(t *testing.T, jsonStr string) *controller.KlondikeWebOutput {
	t.Helper()
	var out controller.KlondikeWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestKlondikeWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		kg := new(interfaces.MockKlondikeGame)
		setupKlondikeWebMockDefaults(kg)
		p := new(KlondikeWebPresenter)

		result := parseKlondikeOutput(t, p.Output(kg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Equal(t, 24, result.StockCount)
		assert.Empty(t, result.Waste)
		assert.Len(t, result.Tableau, domain.KlondikeTableauCnt)
		assert.Len(t, result.Foundation, domain.KlondikeFoundationCnt)
		assert.Equal(t, "klondike.playing", result.MessageCode)
	})

	t.Run("waste with cards", func(t *testing.T) {
		kg := new(interfaces.MockKlondikeGame)
		setupKlondikeWebMockDefaults(kg)
		kg.ExpectedCalls = filterCalls(kg.ExpectedCalls, "GetWaste")
		kg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})

		p := new(KlondikeWebPresenter)
		result := parseKlondikeOutput(t, p.Output(kg, nil))
		assert.Len(t, result.Waste, 1)
		assert.Equal(t, "HEART", result.Waste[0].Design)
		assert.Equal(t, 5, result.Waste[0].Value)
	})

	t.Run("face down card hides data", func(t *testing.T) {
		kg := new(interfaces.MockKlondikeGame)
		setupKlondikeWebMockDefaults(kg)
		p := new(KlondikeWebPresenter)

		result := parseKlondikeOutput(t, p.Output(kg, nil))
		// Column 1 has 2 cards: first face-down
		assert.Len(t, result.Tableau[1], 2)
		assert.False(t, result.Tableau[1][0].FaceUp)
		assert.Nil(t, result.Tableau[1][0].Card)
		assert.True(t, result.Tableau[1][1].FaceUp)
		assert.NotNil(t, result.Tableau[1][1].Card)
	})

	t.Run("foundation with cards", func(t *testing.T) {
		kg := new(interfaces.MockKlondikeGame)
		setupKlondikeWebMockDefaults(kg)
		kg.ExpectedCalls = filterCalls(kg.ExpectedCalls, "GetFoundation")
		var f [domain.KlondikeFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		kg.On("GetFoundation").Return(f)

		p := new(KlondikeWebPresenter)
		result := parseKlondikeOutput(t, p.Output(kg, nil))
		assert.Len(t, result.Foundation[0], 1)
		assert.Equal(t, "SPADE", result.Foundation[0][0].Design)
	})

	t.Run("with error", func(t *testing.T) {
		kg := new(interfaces.MockKlondikeGame)
		setupKlondikeWebMockDefaults(kg)
		p := new(KlondikeWebPresenter)

		result := parseKlondikeOutput(t, p.Output(kg, assert.AnError))
		assert.Equal(t, assert.AnError.Error(), result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		kg := new(interfaces.MockKlondikeGame)
		setupKlondikeWebMockDefaults(kg)
		kg.ExpectedCalls = filterCalls(kg.ExpectedCalls, "GetPhase")
		kg.On("GetPhase").Return(domain.KlondikePhaseGameClear)

		p := new(KlondikeWebPresenter)
		result := parseKlondikeOutput(t, p.Output(kg, nil))
		assert.Equal(t, "klondike.gameClear", result.MessageCode)
		assert.Contains(t, result.Message, "ゲームクリア")
		assert.NotEmpty(t, result.MessageParams)
	})

	t.Run("game over", func(t *testing.T) {
		kg := new(interfaces.MockKlondikeGame)
		setupKlondikeWebMockDefaults(kg)
		kg.ExpectedCalls = filterCalls(kg.ExpectedCalls, "GetPhase")
		kg.On("GetPhase").Return(domain.KlondikePhaseGameOver)

		p := new(KlondikeWebPresenter)
		result := parseKlondikeOutput(t, p.Output(kg, nil))
		assert.Equal(t, "klondike.gameOver", result.MessageCode)
		assert.Contains(t, result.Message, "ゲームオーバー")
	})
}

func TestKlondikeWebPresenter_HintOutput(t *testing.T) {
	t.Run("hint available", func(t *testing.T) {
		kg := new(interfaces.MockKlondikeGame)
		kg.On("GetHint").Return(&domain.KlondikeHint{
			FromZone:  "tableau",
			FromCol:   0,
			CardIndex: 2,
			ToZone:    "foundation",
			ToCol:     0,
		})
		kg.On("GetPhase").Return(domain.KlondikePhasePlaying)
		kg.On("GetMoveCount").Return(5)
		kg.On("GetStockCount").Return(20)

		p := new(KlondikeWebPresenter)
		result := parseKlondikeOutput(t, p.HintOutput(kg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "tableau", result.Hint.FromZone)
		assert.Equal(t, 0, result.Hint.FromCol)
		assert.Equal(t, 2, result.Hint.CardIndex)
		assert.Equal(t, "foundation", result.Hint.ToZone)
		assert.Equal(t, "klondike.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		kg := new(interfaces.MockKlondikeGame)
		kg.On("GetHint").Return((*domain.KlondikeHint)(nil))
		kg.On("GetPhase").Return(domain.KlondikePhasePlaying)
		kg.On("GetMoveCount").Return(0)
		kg.On("GetStockCount").Return(24)

		p := new(KlondikeWebPresenter)
		result := parseKlondikeOutput(t, p.HintOutput(kg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "klondike.noHint", result.MessageCode)
	})
}

func TestKlondikeWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("during game", func(t *testing.T) {
		kg := new(interfaces.MockKlondikeGame)
		kg.On("GetPhase").Return(domain.KlondikePhasePlaying)

		p := new(KlondikeWebPresenter)
		result := p.ActionLogOutput(kg)
		var out controller.ActionLogWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Empty(t, out.Entries)
	})

	t.Run("after game clear", func(t *testing.T) {
		kg := new(interfaces.MockKlondikeGame)
		kg.On("GetPhase").Return(domain.KlondikePhaseGameClear)
		kg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "draw", Detail: "test", Cards: nil},
		})

		p := new(KlondikeWebPresenter)
		result := p.ActionLogOutput(kg)
		var out controller.ActionLogWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Len(t, out.Entries, 1)
	})

	t.Run("after game over", func(t *testing.T) {
		kg := new(interfaces.MockKlondikeGame)
		kg.On("GetPhase").Return(domain.KlondikePhaseGameOver)
		kg.On("GetActionLog").Return([]*domain.ActionLogEntry{})

		p := new(KlondikeWebPresenter)
		result := p.ActionLogOutput(kg)
		var out controller.ActionLogWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Empty(t, out.Entries)
	})
}
