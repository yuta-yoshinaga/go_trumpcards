//go:build test

package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// teqCard は指定デザイン・値のカードを生成するテストヘルパー。
func teqCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// teqRow31Black は合計 31 の黒始まりの列を返す (先頭 ♠10)。
func teqRow31Black() []*domain.Card {
	return []*domain.Card{
		teqCard(domain.CardDesignSpade, 10),
		teqCard(domain.CardDesignClover, 10),
		teqCard(domain.CardDesignSpade, 10),
		teqCard(domain.CardDesignClover, 1),
	}
}

// teqRow32Black は合計 32 の黒始まりの列を返す。
func teqRow32Black() []*domain.Card {
	return []*domain.Card{
		teqCard(domain.CardDesignSpade, 10),
		teqCard(domain.CardDesignClover, 10),
		teqCard(domain.CardDesignSpade, 10),
		teqCard(domain.CardDesignClover, 2),
	}
}

func TestTrenteEtQuarante_ResetChips(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.SetChips(5)
	g.Reset()
	assert.Equal(t, domain.TrenteEtQuaranteDefaultChips, g.GetChips())
	assert.Equal(t, domain.TrenteEtQuarantePhaseBet, g.GetPhase())
	assert.False(t, g.GetGameEndFlag())
}

func TestTrenteEtQuarante_PlaceBet_RealDeal(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	require.NoError(t, g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 100))

	assert.Equal(t, domain.TrenteEtQuarantePhaseResult, g.GetPhase())
	assert.True(t, g.GetGameEndFlag())
	// Each row must stop in the 31–40 range.
	assert.GreaterOrEqual(t, g.GetNoirTotal(), domain.TrenteEtQuaranteTarget)
	assert.LessOrEqual(t, g.GetNoirTotal(), 40)
	assert.GreaterOrEqual(t, g.GetRougeTotal(), domain.TrenteEtQuaranteTarget)
	assert.LessOrEqual(t, g.GetRougeTotal(), 40)
	// Chips are consistent: 1000 - stake + payout.
	assert.Equal(t, 1000-100+g.GetPayout(), g.GetChips())
	assert.Equal(t, 1, g.GetRoundNumber())
}

func TestTrenteEtQuarante_PlaceBet_Errors(t *testing.T) {
	t.Run("invalid amount below min", func(t *testing.T) {
		g := domain.NewDefaultTrenteEtQuarante()
		g.Reset()
		err := g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 5)
		assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
	})
	t.Run("invalid amount not multiple", func(t *testing.T) {
		g := domain.NewDefaultTrenteEtQuarante()
		g.Reset()
		err := g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 15)
		assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
	})
	t.Run("invalid amount above max", func(t *testing.T) {
		g := domain.NewDefaultTrenteEtQuarante()
		g.Reset()
		err := g.PlaceBet(domain.TrenteEtQuaranteBetNoir, domain.TrenteEtQuaranteMaxBet+10)
		assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
	})
	t.Run("invalid bet type", func(t *testing.T) {
		g := domain.NewDefaultTrenteEtQuarante()
		g.Reset()
		err := g.PlaceBet(domain.TrenteEtQuaranteBet(99), 100)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
		assert.Equal(t, 1000, g.GetChips(), "chips untouched on invalid bet")
	})
	t.Run("insufficient chips", func(t *testing.T) {
		g := domain.NewDefaultTrenteEtQuarante()
		g.Reset()
		g.SetChips(50)
		err := g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 100)
		assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
	})
	t.Run("wrong phase", func(t *testing.T) {
		g := domain.NewDefaultTrenteEtQuarante()
		g.Reset()
		require.NoError(t, g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 100))
		err := g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 100)
		assert.True(t, errors.Is(err, domain.ErrWrongPhase))
	})
}

func TestTrenteEtQuarante_Resolve_NoirWins(t *testing.T) {
	// Noir 31 < Rouge 32 -> Noir (black) row wins, first card black.
	cases := []struct {
		bet    domain.TrenteEtQuaranteBet
		result domain.TrenteEtQuaranteResult
		payout int
	}{
		{domain.TrenteEtQuaranteBetNoir, domain.TrenteEtQuaranteResultWin, 200},
		{domain.TrenteEtQuaranteBetRouge, domain.TrenteEtQuaranteResultLose, 0},
		{domain.TrenteEtQuaranteBetCouleur, domain.TrenteEtQuaranteResultWin, 200}, // black first card, black wins -> match
		{domain.TrenteEtQuaranteBetInverse, domain.TrenteEtQuaranteResultLose, 0},
	}
	for _, tc := range cases {
		g := domain.NewDefaultTrenteEtQuarante()
		g.Reset()
		g.ResolveRowsForTest(tc.bet, 100, teqRow31Black(), teqRow32Black())
		assert.Equal(t, domain.TrenteEtQuaranteRowNoir, g.GetWinningRow())
		assert.False(t, g.GetFirstCardRed())
		assert.Equal(t, tc.result, g.GetResult(), "bet=%d", tc.bet)
		assert.Equal(t, tc.payout, g.GetPayout(), "bet=%d", tc.bet)
	}
}

func TestTrenteEtQuarante_Resolve_RougeWins(t *testing.T) {
	// Noir 32 > Rouge 31 -> Rouge (red) row wins, first card black.
	cases := []struct {
		bet    domain.TrenteEtQuaranteBet
		result domain.TrenteEtQuaranteResult
	}{
		{domain.TrenteEtQuaranteBetNoir, domain.TrenteEtQuaranteResultLose},
		{domain.TrenteEtQuaranteBetRouge, domain.TrenteEtQuaranteResultWin},
		{domain.TrenteEtQuaranteBetCouleur, domain.TrenteEtQuaranteResultLose}, // black first card, red wins -> mismatch
		{domain.TrenteEtQuaranteBetInverse, domain.TrenteEtQuaranteResultWin},
	}
	for _, tc := range cases {
		g := domain.NewDefaultTrenteEtQuarante()
		g.Reset()
		g.ResolveRowsForTest(tc.bet, 100, teqRow32Black(), teqRow31Black())
		assert.Equal(t, domain.TrenteEtQuaranteRowRouge, g.GetWinningRow())
		assert.Equal(t, tc.result, g.GetResult(), "bet=%d", tc.bet)
	}
}

func TestTrenteEtQuarante_Resolve_Push(t *testing.T) {
	// Both 32 (not 31) -> push, stake returned.
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	g.ResolveRowsForTest(domain.TrenteEtQuaranteBetNoir, 100, teqRow32Black(), teqRow32Black())
	assert.Equal(t, domain.TrenteEtQuaranteRowNone, g.GetWinningRow())
	assert.False(t, g.GetRefait())
	assert.Equal(t, domain.TrenteEtQuaranteResultDraw, g.GetResult())
	assert.Equal(t, 100, g.GetPayout())
}

func TestTrenteEtQuarante_Resolve_Refait(t *testing.T) {
	// Both 31 -> Refait, half stake to house.
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	g.ResolveRowsForTest(domain.TrenteEtQuaranteBetRouge, 100, teqRow31Black(), teqRow31Black())
	assert.Equal(t, domain.TrenteEtQuaranteRowNone, g.GetWinningRow())
	assert.True(t, g.GetRefait())
	assert.Equal(t, domain.TrenteEtQuaranteResultLose, g.GetResult())
	assert.Equal(t, 50, g.GetPayout())
}

func TestTrenteEtQuarante_Resolve_CouleurRedFirstCard(t *testing.T) {
	// First card red (♥10) and Rouge row wins -> Couleur wins.
	noir := []*domain.Card{
		teqCard(domain.CardDesignHeart, 10),
		teqCard(domain.CardDesignClover, 10),
		teqCard(domain.CardDesignSpade, 10),
		teqCard(domain.CardDesignClover, 2),
	} // total 32
	rouge := teqRow31Black() // total 31 -> Rouge wins
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	g.ResolveRowsForTest(domain.TrenteEtQuaranteBetCouleur, 100, noir, rouge)
	assert.True(t, g.GetFirstCardRed())
	assert.Equal(t, domain.TrenteEtQuaranteRowRouge, g.GetWinningRow())
	assert.Equal(t, domain.TrenteEtQuaranteResultWin, g.GetResult())
}

func TestTrenteEtQuarante_NextRound(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	require.NoError(t, g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 100))
	chips := g.GetChips()
	g.NextRound()
	assert.Equal(t, domain.TrenteEtQuarantePhaseBet, g.GetPhase())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, chips, g.GetChips(), "chips persist across rounds")
	assert.Equal(t, 1, g.GetRoundNumber())
}

func TestTrenteEtQuarante_MultiRoundReshuffle(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	for i := 0; i < 80; i++ {
		if g.GetChips() < domain.TrenteEtQuaranteMinBet {
			g.SetChips(domain.TrenteEtQuaranteDefaultChips)
		}
		require.NoError(t, g.PlaceBet(domain.TrenteEtQuaranteBetNoir, domain.TrenteEtQuaranteMinBet))
		assert.GreaterOrEqual(t, g.GetNoirTotal(), domain.TrenteEtQuaranteTarget)
		g.NextRound()
	}
	assert.GreaterOrEqual(t, g.GetRemainingDeck(), 0)
}

func TestTrenteEtQuarante_Hint(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "even_odds", hint.Reason)

	require.NoError(t, g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 100))
	assert.Nil(t, g.GetHint(), "no hint in result phase")
}

func TestTrenteEtQuarante_ActionLog(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	require.NoError(t, g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 100))
	assert.NotEmpty(t, g.GetActionLog())
}

func TestTrenteEtQuaranteConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultTrenteEtQuaranteConfig().Validate())
	assert.Error(t, domain.TrenteEtQuaranteConfig{DefaultBet: 99}.Validate())
	assert.Error(t, domain.TrenteEtQuaranteConfig{DefaultBet: -1}.Validate())
}

func TestTrenteEtQuaranteConfig_JSON(t *testing.T) {
	cfg := domain.TrenteEtQuaranteConfig{DefaultBet: domain.TrenteEtQuaranteBetCouleur}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	var got domain.TrenteEtQuaranteConfig
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, cfg, got)
}

func TestTrenteEtQuarantePlayer(t *testing.T) {
	p := domain.NewTrenteEtQuarantePlayer(500)
	assert.Equal(t, 500, p.GetChips())
	p.AddChips(100)
	assert.Equal(t, 600, p.GetChips())
	assert.True(t, p.SubtractChips(200))
	assert.False(t, p.SubtractChips(10000))
	p.RecordRound(true)
	p.RecordRound(false)
	assert.Equal(t, 2, p.GetRoundsPlayed())
	assert.Equal(t, 1, p.GetRoundsWon())
	p.ResetStats()
	assert.Equal(t, 0, p.GetRoundsPlayed())

	data, err := json.Marshal(p)
	require.NoError(t, err)
	var got domain.TrenteEtQuarantePlayer
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, 400, got.GetChips())
}

func TestTrenteEtQuarantePlayer_UnmarshalErrors(t *testing.T) {
	var p domain.TrenteEtQuarantePlayer
	assert.Error(t, json.Unmarshal([]byte(`{"ch":-5,"rp":0,"rw":0}`), &p))
	assert.Error(t, json.Unmarshal([]byte(`{"ch":100,"rp":1,"rw":5}`), &p))
	assert.Error(t, json.Unmarshal([]byte(`{"ch":100,"rp":-1,"rw":0}`), &p))
}

func TestTrenteEtQuarante_JSONRoundTrip(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	require.NoError(t, g.PlaceBet(domain.TrenteEtQuaranteBetRouge, 100))

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var got domain.TrenteEtQuarante
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, g.GetPhase(), got.GetPhase())
	assert.Equal(t, g.GetChips(), got.GetChips())
	assert.Equal(t, g.GetCurrentBet(), got.GetCurrentBet())
	assert.Equal(t, g.GetNoirTotal(), got.GetNoirTotal())
	assert.Equal(t, g.GetWinningRow(), got.GetWinningRow())
	assert.Equal(t, g.GetResult(), got.GetResult())
}

func TestTrenteEtQuarante_UnmarshalValidation(t *testing.T) {
	cases := map[string]string{
		"not json":          `not json`,
		"invalid config":    `{"cf":{"db":9},"ph":1,"cb":0}`,
		"invalid phase":     `{"cf":{"db":0},"ph":9,"cb":0}`,
		"invalid bet":       `{"cf":{"db":0},"ph":1,"cb":9}`,
		"negative stake":    `{"cf":{"db":0},"ph":1,"cb":0,"st":-5}`,
		"winning row range": `{"cf":{"db":0},"ph":1,"cb":0,"wr":5}`,
		"result range":      `{"cf":{"db":0},"ph":1,"cb":0,"re":9}`,
		"row too long":      `{"cf":{"db":0},"ph":1,"cb":0,"nr":` + teqLongRowJSON() + `}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			var g domain.TrenteEtQuarante
			assert.Error(t, json.Unmarshal([]byte(body), &g))
		})
	}
}

func TestTrenteEtQuarante_UnmarshalDefaults(t *testing.T) {
	// Minimal valid payload rebuilds empty slices / default player and shoe.
	var g domain.TrenteEtQuarante
	require.NoError(t, json.Unmarshal([]byte(`{"cf":{"db":0},"ph":1,"cb":0}`), &g))
	assert.NotNil(t, g.GetNoirRow())
	assert.NotNil(t, g.GetRougeRow())
	assert.Equal(t, domain.TrenteEtQuaranteDefaultChips, g.GetChips())
	assert.GreaterOrEqual(t, g.GetRemainingDeck(), 0)
}

// teqLongRowJSON は上限超過の列 JSON を生成する。
func teqLongRowJSON() string {
	cards := make([]*domain.Card, domain.TrenteEtQuaranteMaxRowLen+1)
	for i := range cards {
		cards[i] = teqCard(domain.CardDesignSpade, 5)
	}
	b, _ := json.Marshal(cards)
	return string(b)
}
