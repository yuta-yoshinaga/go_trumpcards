//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupSpiteAndMaliceWebMockDefaults(g *interfaces.MockSpiteAndMaliceGame) {
	g.On("GetPhase").Return(domain.SpiteAndMalicePhasePlaying).Maybe()
	g.On("GetCurrent").Return(0).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("GetWinner").Return(-1).Maybe()
	g.On("GetStockSize").Return(40).Maybe()
	g.On("GetCompletedSize").Return(0).Maybe()
	g.On("GetConfig").Return(domain.DefaultSpiteAndMaliceConfig()).Maybe()
	g.On("CanAutoComplete").Return(false).Maybe()
	var foundations [domain.SpiteAndMaliceFoundationCnt][]*domain.Card
	foundations[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	g.On("GetFoundations").Return(foundations).Maybe()
	for i := range domain.SpiteAndMaliceFoundationCnt {
		g.On("GetFoundationTopValue", i).Return(0).Maybe()
	}
	human := domain.NewSpiteAndMalicePlayer(false)
	human.AddToHand(domain.NewCard(domain.CardDesignSpade, 5, false))
	human.AddToGoal(domain.NewCard(domain.CardDesignHeart, 9, false))
	human.PushSide(0, domain.NewCard(domain.CardDesignClover, 3, false))
	cpu := domain.NewSpiteAndMalicePlayer(true)
	cpu.AddToHand(domain.NewCard(domain.CardDesignDiamond, 7, false))
	cpu.AddToGoal(domain.NewCard(domain.CardDesignSpade, 4, false))
	g.On("GetPlayer", 0).Return(human).Maybe()
	g.On("GetPlayer", 1).Return(cpu).Maybe()
}

func decodeSpiteAndMaliceWebOutput(t *testing.T, raw string) *controller.SpiteAndMaliceWebOutput {
	t.Helper()
	var out controller.SpiteAndMaliceWebOutput
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return &out
}

// setupSpiteAndMaliceOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupSpiteAndMaliceOutputMock(g *interfaces.MockSpiteAndMaliceGame) {
	setupSpiteAndMaliceWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestSpiteAndMaliceWebPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		setupSpiteAndMaliceOutputMock(g)
		raw := new(SpiteAndMaliceWebPresenter).Output(g, nil)
		out := decodeSpiteAndMaliceWebOutput(t, raw)
		assert.Equal(t, "spiteandmalice.playing", out.MessageCode)
		assert.Equal(t, domain.SpiteAndMaliceGoalSizeDefault, out.GoalSize)
		assert.Len(t, out.Players[0].Hand, 1)
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		setupSpiteAndMaliceOutputMock(g)
		raw := new(SpiteAndMaliceWebPresenter).Output(g, assert.AnError)
		out := decodeSpiteAndMaliceWebOutput(t, raw)
		assert.Equal(t, assert.AnError.Error(), out.Message)
	})

	t.Run("human win", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		g.On("GetPhase").Return(domain.SpiteAndMalicePhaseGameOver).Maybe()
		g.On("GetCurrent").Return(0).Maybe()
		g.On("GetMoveCount").Return(42).Maybe()
		g.On("GetWinner").Return(domain.SpiteAndMaliceHumanIdx).Maybe()
		g.On("GetStockSize").Return(0).Maybe()
		g.On("GetCompletedSize").Return(0).Maybe()
		g.On("GetConfig").Return(domain.DefaultSpiteAndMaliceConfig()).Maybe()
		g.On("CanAutoComplete").Return(false).Maybe()
		var foundations [domain.SpiteAndMaliceFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		for i := range domain.SpiteAndMaliceFoundationCnt {
			g.On("GetFoundationTopValue", i).Return(0).Maybe()
		}
		g.On("GetPlayer", 0).Return(domain.NewSpiteAndMalicePlayer(false)).Maybe()
		g.On("GetPlayer", 1).Return(domain.NewSpiteAndMalicePlayer(true)).Maybe()
		g.On("GetHint").Return(nil).Maybe()
		raw := new(SpiteAndMaliceWebPresenter).Output(g, nil)
		out := decodeSpiteAndMaliceWebOutput(t, raw)
		assert.Equal(t, "spiteandmalice.win", out.MessageCode)
		assert.Equal(t, "42", out.MessageParams["moveCount"])
	})

	t.Run("cpu win", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		g.On("GetPhase").Return(domain.SpiteAndMalicePhaseGameOver).Maybe()
		g.On("GetCurrent").Return(1).Maybe()
		g.On("GetMoveCount").Return(42).Maybe()
		g.On("GetWinner").Return(domain.SpiteAndMaliceCpuIdx).Maybe()
		g.On("GetStockSize").Return(0).Maybe()
		g.On("GetCompletedSize").Return(0).Maybe()
		g.On("GetConfig").Return(domain.DefaultSpiteAndMaliceConfig()).Maybe()
		g.On("CanAutoComplete").Return(false).Maybe()
		var foundations [domain.SpiteAndMaliceFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		for i := range domain.SpiteAndMaliceFoundationCnt {
			g.On("GetFoundationTopValue", i).Return(0).Maybe()
		}
		g.On("GetPlayer", 0).Return(domain.NewSpiteAndMalicePlayer(false)).Maybe()
		g.On("GetPlayer", 1).Return(domain.NewSpiteAndMalicePlayer(true)).Maybe()
		g.On("GetHint").Return(nil).Maybe()
		raw := new(SpiteAndMaliceWebPresenter).Output(g, nil)
		out := decodeSpiteAndMaliceWebOutput(t, raw)
		assert.Equal(t, "spiteandmalice.lose", out.MessageCode)
	})

	// Regression test for the visibility bug from PR #1503 review:
	// the CPU's hand must NEVER be revealed in the JSON output, regardless of
	// whose turn it currently is.
	t.Run("cpu hand stays hidden on both turns", func(t *testing.T) {
		for _, current := range []int{domain.SpiteAndMaliceHumanIdx, domain.SpiteAndMaliceCpuIdx} {
			g := new(interfaces.MockSpiteAndMaliceGame)
			g.On("GetPhase").Return(domain.SpiteAndMalicePhasePlaying).Maybe()
			g.On("GetCurrent").Return(current).Maybe()
			g.On("GetMoveCount").Return(0).Maybe()
			g.On("GetWinner").Return(-1).Maybe()
			g.On("GetStockSize").Return(0).Maybe()
			g.On("GetCompletedSize").Return(0).Maybe()
			g.On("GetConfig").Return(domain.DefaultSpiteAndMaliceConfig()).Maybe()
			g.On("CanAutoComplete").Return(false).Maybe()
			var foundations [domain.SpiteAndMaliceFoundationCnt][]*domain.Card
			g.On("GetFoundations").Return(foundations).Maybe()
			for i := range domain.SpiteAndMaliceFoundationCnt {
				g.On("GetFoundationTopValue", i).Return(0).Maybe()
			}
			human := domain.NewSpiteAndMalicePlayer(false)
			human.AddToHand(domain.NewCard(domain.CardDesignSpade, 5, false))
			cpu := domain.NewSpiteAndMalicePlayer(true)
			cpu.AddToHand(domain.NewCard(domain.CardDesignDiamond, 3, false))
			cpu.AddToHand(domain.NewCard(domain.CardDesignHeart, 9, false))
			g.On("GetPlayer", 0).Return(human).Maybe()
			g.On("GetPlayer", 1).Return(cpu).Maybe()
			g.On("GetHint").Return(nil).Maybe()
			raw := new(SpiteAndMaliceWebPresenter).Output(g, nil)
			out := decodeSpiteAndMaliceWebOutput(t, raw)
			// 人間の手札は常に公開
			require.Len(t, out.Players[0].Hand, 1)
			require.NotNil(t, out.Players[0].Hand[0])
			assert.Equal(t, 5, out.Players[0].Hand[0].Value)
			// CPU の手札は枚数のみ (要素は nil) — どちらのターンでも
			require.Len(t, out.Players[1].Hand, 2)
			for _, c := range out.Players[1].Hand {
				assert.Nil(t, c, "current=%d: cpu hand element must be nil", current)
			}
		}
	})

	t.Run("nil player", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		g.On("GetPhase").Return(domain.SpiteAndMalicePhasePlaying).Maybe()
		g.On("GetCurrent").Return(0).Maybe()
		g.On("GetMoveCount").Return(0).Maybe()
		g.On("GetWinner").Return(-1).Maybe()
		g.On("GetStockSize").Return(0).Maybe()
		g.On("GetCompletedSize").Return(0).Maybe()
		g.On("GetConfig").Return(domain.DefaultSpiteAndMaliceConfig()).Maybe()
		g.On("CanAutoComplete").Return(false).Maybe()
		var foundations [domain.SpiteAndMaliceFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		for i := range domain.SpiteAndMaliceFoundationCnt {
			g.On("GetFoundationTopValue", i).Return(0).Maybe()
		}
		g.On("GetPlayer", 0).Return((*domain.SpiteAndMalicePlayer)(nil)).Maybe()
		g.On("GetPlayer", 1).Return((*domain.SpiteAndMalicePlayer)(nil)).Maybe()
		g.On("GetHint").Return(nil).Maybe()
		raw := new(SpiteAndMaliceWebPresenter).Output(g, nil)
		out := decodeSpiteAndMaliceWebOutput(t, raw)
		assert.Nil(t, out.Players[0].Hand)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestSpiteAndMaliceWebPresenterOutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		smg := new(interfaces.MockSpiteAndMaliceGame)
		setupSpiteAndMaliceWebMockDefaults(smg)
		smg.On("GetHint").Return(&domain.SpiteAndMaliceHint{Source: domain.SpiteAndMaliceSourceHand, Index: 0, FoundationIdx: 1, Discard: false}).Maybe()

		result := new(SpiteAndMaliceWebPresenter).Output(smg, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	t.Run("not once the game is over", func(t *testing.T) {
		smg := new(interfaces.MockSpiteAndMaliceGame)
		setupSpiteAndMaliceWebMockDefaults(smg)
		smg.ExpectedCalls = filterCalls(smg.ExpectedCalls, "GetPhase")
		smg.On("GetPhase").Return(domain.SpiteAndMalicePhaseGameOver)
		smg.On("GetHint").Return(&domain.SpiteAndMaliceHint{Source: domain.SpiteAndMaliceSourceHand, Index: 0, FoundationIdx: 1, Discard: false}).Maybe()

		result := new(SpiteAndMaliceWebPresenter).Output(smg, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestSpiteAndMaliceWebPresenter_HintOutput(t *testing.T) {
	tests := []struct {
		name   string
		hint   *domain.SpiteAndMaliceHint
		source string
		code   string
	}{
		{"goal", &domain.SpiteAndMaliceHint{Source: domain.SpiteAndMaliceSourceGoal, FoundationIdx: 1}, "goal", "spiteandmalice.hintAvailable"},
		{"hand", &domain.SpiteAndMaliceHint{Source: domain.SpiteAndMaliceSourceHand, Index: 2, FoundationIdx: 0}, "hand", "spiteandmalice.hintAvailable"},
		{"side", &domain.SpiteAndMaliceHint{Source: domain.SpiteAndMaliceSourceSide, Index: 1, FoundationIdx: 2}, "side", "spiteandmalice.hintAvailable"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := new(interfaces.MockSpiteAndMaliceGame)
			setupSpiteAndMaliceWebMockDefaults(g)
			g.On("GetHint").Return(tt.hint)
			raw := new(SpiteAndMaliceWebPresenter).HintOutput(g)
			out := decodeSpiteAndMaliceWebOutput(t, raw)
			assert.Equal(t, tt.code, out.MessageCode)
			require.NotNil(t, out.Hint)
			assert.Equal(t, tt.source, out.Hint.Source)
		})
	}

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		setupSpiteAndMaliceWebMockDefaults(g)
		g.On("GetHint").Return((*domain.SpiteAndMaliceHint)(nil))
		raw := new(SpiteAndMaliceWebPresenter).HintOutput(g)
		out := decodeSpiteAndMaliceWebOutput(t, raw)
		assert.Equal(t, "spiteandmalice.noHint", out.MessageCode)
		assert.Nil(t, out.Hint)
	})

	t.Run("unknown source string", func(t *testing.T) {
		assert.Equal(t, "", sourceToString(domain.SpiteAndMaliceSource(99)))
	})
}

func TestSpiteAndMaliceWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		g.On("GetPhase").Return(domain.SpiteAndMalicePhasePlaying)
		g.On("GetGameEndFlag").Return(false)
		assert.NotEmpty(t, new(SpiteAndMaliceWebPresenter).ActionLogOutput(g))
	})
	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		g.On("GetPhase").Return(domain.SpiteAndMalicePhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "playHand"}})
		assert.NotEmpty(t, new(SpiteAndMaliceWebPresenter).ActionLogOutput(g))
	})
}
