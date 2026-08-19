//go:build test

package domain_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// gutsCard は指定デザイン・値のカードを生成するテストヘルパー。
func gutsCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// gutsSetHand は player の手札を 2 枚に差し替える。
func gutsSetHand(p *domain.GutsPlayer, d1, v1, d2, v2 int) {
	p.ClearHand()
	p.AddCard(gutsCard(d1, v1))
	p.AddCard(gutsCard(d2, v2))
}

func TestGuts_ResetDealsRound(t *testing.T) {
	g := domain.NewDefaultGuts()
	assert.Equal(t, domain.GutsPhaseDeclare, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	// pot = ante * playerCount, human chips = starting - ante.
	cfg := g.GetConfig()
	assert.Equal(t, cfg.Ante*cfg.PlayerCount, g.GetPot())
	assert.Equal(t, cfg.StartingChips-cfg.Ante, g.GetChips())
	assert.Equal(t, domain.GutsHandSize, g.GetPlayer(0).GetCardsSize())
}

func TestGuts_Declare_Resolves(t *testing.T) {
	g := domain.NewDefaultGuts()
	require.NoError(t, g.Declare(true))
	assert.Equal(t, domain.GutsPhaseResult, g.GetPhase())
}

func TestGuts_Declare_Errors(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		g := domain.NewDefaultGuts()
		require.NoError(t, g.Declare(false))
		err := g.Declare(true)
		assert.True(t, errors.Is(err, domain.ErrWrongPhase))
	})
	t.Run("game ended", func(t *testing.T) {
		g := domain.NewDefaultGuts()
		cfg := g.GetConfig()
		cfg.TargetRounds = 1
		g.SetConfig(cfg)
		g.Reset()
		require.NoError(t, g.Declare(true))
		assert.True(t, g.GetGameEndFlag())
		err := g.Declare(true)
		assert.True(t, errors.Is(err, domain.ErrGameEnded))
	})
}

func TestGutsEval(t *testing.T) {
	// Pair beats high card.
	catPair, _ := domain.GutsEval([]*domain.Card{gutsCard(domain.CardDesignSpade, 8), gutsCard(domain.CardDesignHeart, 8)})
	catHigh, _ := domain.GutsEval([]*domain.Card{gutsCard(domain.CardDesignSpade, 13), gutsCard(domain.CardDesignHeart, 2)})
	assert.Equal(t, domain.GutsHandPair, catPair)
	assert.Equal(t, domain.GutsHandHighCard, catHigh)

	// Ace high: A-K beats K-Q.
	_, tbAK := domain.GutsEval([]*domain.Card{gutsCard(domain.CardDesignSpade, 1), gutsCard(domain.CardDesignHeart, 13)})
	_, tbKQ := domain.GutsEval([]*domain.Card{gutsCard(domain.CardDesignSpade, 13), gutsCard(domain.CardDesignHeart, 12)})
	assert.Equal(t, 1, domain.GutsCompare(domain.GutsHandHighCard, tbAK, domain.GutsHandHighCard, tbKQ))
	assert.Equal(t, -1, domain.GutsCompare(domain.GutsHandHighCard, tbKQ, domain.GutsHandHighCard, tbAK))
	assert.Equal(t, 0, domain.GutsCompare(domain.GutsHandHighCard, tbAK, domain.GutsHandHighCard, tbAK))

	// Invalid size.
	cat, tb := domain.GutsEval([]*domain.Card{gutsCard(domain.CardDesignSpade, 5)})
	assert.Equal(t, -1, cat)
	assert.Nil(t, tb)
}

func TestGuts_Settle_HumanWins(t *testing.T) {
	g := domain.NewDefaultGuts()
	gutsSetHand(g.GetPlayer(0), domain.CardDesignSpade, 10, domain.CardDesignClover, 10) // pair
	g.GetPlayer(0).SetIn(true)
	gutsSetHand(g.GetPlayer(1), domain.CardDesignSpade, 9, domain.CardDesignHeart, 5) // high card
	g.GetPlayer(1).SetIn(true)
	for i := 2; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetIn(false)
	}
	g.SetPot(100)
	g.SetPhase(domain.GutsPhaseDeclare)
	chipsBefore := g.GetPlayer(0).GetChips()
	matcherChipsBefore := g.GetPlayer(1).GetChips()

	g.SettleForTest()

	assert.Equal(t, 0, g.GetWinnerIdx())
	assert.Equal(t, domain.GutsResultWin, g.GetResult())
	assert.Equal(t, chipsBefore+100, g.GetPlayer(0).GetChips())
	assert.True(t, g.IsMatcher(1))
	assert.Equal(t, matcherChipsBefore-100, g.GetPlayer(1).GetChips())
	assert.Equal(t, 100, g.GetCarryPot())
	assert.Equal(t, domain.GutsPhaseResult, g.GetPhase())
}

func TestGuts_Settle_HumanLosesMatches(t *testing.T) {
	g := domain.NewDefaultGuts()
	gutsSetHand(g.GetPlayer(0), domain.CardDesignSpade, 4, domain.CardDesignClover, 7) // weak high card
	g.GetPlayer(0).SetIn(true)
	gutsSetHand(g.GetPlayer(1), domain.CardDesignSpade, 11, domain.CardDesignHeart, 11) // pair
	g.GetPlayer(1).SetIn(true)
	for i := 2; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetIn(false)
	}
	g.SetPot(50)
	g.SettleForTest()

	assert.Equal(t, 1, g.GetWinnerIdx())
	assert.Equal(t, domain.GutsResultLose, g.GetResult())
	assert.True(t, g.IsMatcher(0))
	assert.Equal(t, []int{0}, g.GetMatchers())
}

func TestGuts_Settle_SoleWinnerCleanWin(t *testing.T) {
	g := domain.NewDefaultGuts()
	g.GetPlayer(0).SetIn(true)
	for i := 1; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetIn(false)
	}
	g.SetPot(80)
	g.SettleForTest()

	assert.Equal(t, 0, g.GetWinnerIdx())
	assert.Empty(t, g.GetMatchers())
	assert.Equal(t, 0, g.GetCarryPot(), "clean win seeds no carry pot")
}

func TestGuts_Settle_NobodyIn(t *testing.T) {
	g := domain.NewDefaultGuts()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetIn(false)
	}
	g.SetPot(120)
	g.SettleForTest()

	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, domain.GutsResultNone, g.GetResult())
	assert.Equal(t, 120, g.GetCarryPot(), "pot carries over entirely")
}

func TestGuts_NextRound(t *testing.T) {
	g := domain.NewDefaultGuts()
	require.NoError(t, g.Declare(true))
	require.Equal(t, domain.GutsPhaseResult, g.GetPhase())
	if g.GetGameEndFlag() {
		t.Skip("game ended in round 1")
	}
	g.NextRound()
	assert.Equal(t, domain.GutsPhaseDeclare, g.GetPhase())
	assert.Equal(t, 2, g.GetRoundNumber())
	// NextRound is a no-op while in declare phase.
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
}

func TestGuts_GameEndByRounds(t *testing.T) {
	g := domain.NewDefaultGuts()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	require.NoError(t, g.Declare(false))
	assert.True(t, g.GetGameEndFlag())
	assert.GreaterOrEqual(t, g.GetMatchWinnerIdx(), 0)
	// NextRound after game end is a no-op.
	g.NextRound()
	assert.True(t, g.GetGameEndFlag())
}

func TestGuts_Hint(t *testing.T) {
	g := domain.NewDefaultGuts()
	gutsSetHand(g.GetPlayer(0), domain.CardDesignSpade, 7, domain.CardDesignHeart, 7) // pair
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, domain.GutsDeclarationIn, hint.Declaration)
	assert.Equal(t, "strong_hand", hint.Reason)

	gutsSetHand(g.GetPlayer(0), domain.CardDesignSpade, 3, domain.CardDesignHeart, 5) // weak
	hint = g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, domain.GutsDeclarationOut, hint.Declaration)

	require.NoError(t, g.Declare(false))
	assert.Nil(t, g.GetHint(), "no hint in result phase")
}

func TestGuts_ActionLog(t *testing.T) {
	g := domain.NewDefaultGuts()
	require.NoError(t, g.Declare(true))
	assert.NotEmpty(t, g.GetActionLog())
}

func TestGutsConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultGutsConfig().Validate())
	assert.Error(t, domain.GutsConfig{PlayerCount: 1, Ante: 10, StartingChips: 200, TargetRounds: 10}.Validate())
	assert.Error(t, domain.GutsConfig{PlayerCount: 8, Ante: 10, StartingChips: 200, TargetRounds: 10}.Validate())
	assert.Error(t, domain.GutsConfig{PlayerCount: 4, Ante: 0, StartingChips: 200, TargetRounds: 10}.Validate())
	assert.Error(t, domain.GutsConfig{PlayerCount: 4, Ante: 10, StartingChips: 1, TargetRounds: 10}.Validate())
	// Starting chips passes the range check but is below the ante → immediate elimination.
	assert.Error(t, domain.GutsConfig{PlayerCount: 4, Ante: 50, StartingChips: 10, TargetRounds: 10}.Validate())
	assert.Error(t, domain.GutsConfig{PlayerCount: 4, Ante: 10, StartingChips: 200, TargetRounds: 0}.Validate())
}

func TestGutsDeclarationValid(t *testing.T) {
	assert.True(t, domain.GutsDeclarationValid(domain.GutsDeclarationIn))
	assert.True(t, domain.GutsDeclarationValid(domain.GutsDeclarationOut))
	assert.False(t, domain.GutsDeclarationValid(domain.GutsDeclaration(9)))
}

func TestGutsPlayer_JSON(t *testing.T) {
	p := domain.NewGutsPlayer(true, 300)
	p.SetIn(true)
	p.AddRoundBet(20)
	data, err := json.Marshal(p)
	require.NoError(t, err)
	var got domain.GutsPlayer
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, 300, got.GetChips())
	assert.True(t, got.GetIn())
	assert.Equal(t, 20, got.GetRoundBet())
	assert.True(t, got.GetIsHuman())
}

func TestGutsPlayer_UnmarshalErrors(t *testing.T) {
	var p domain.GutsPlayer
	assert.Error(t, json.Unmarshal([]byte(`{"ch":-5,"rb":0}`), &p))
	assert.Error(t, json.Unmarshal([]byte(`{"ch":10,"rb":-1}`), &p))
	assert.Error(t, json.Unmarshal([]byte(`not json`), &p))
}

func TestGuts_JSONRoundTrip(t *testing.T) {
	g := domain.NewDefaultGuts()
	require.NoError(t, g.Declare(true))

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var got domain.Guts
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, g.GetPhase(), got.GetPhase())
	assert.Equal(t, g.GetChips(), got.GetChips())
	assert.Equal(t, g.GetRoundNumber(), got.GetRoundNumber())
	assert.Equal(t, g.GetWinnerIdx(), got.GetWinnerIdx())
	assert.Equal(t, g.GetResult(), got.GetResult())
	assert.Equal(t, g.GetPlayerCnt(), got.GetPlayerCnt())
}

func TestGuts_UnmarshalValidation(t *testing.T) {
	cases := map[string]string{
		"not json":       `not json`,
		"invalid config": `{"cf":{"pc":9,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1}]}`,
		"player mismatch": `{"cf":{"pc":4,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1}],` +
			`"ph":0,"rn":1}`,
		"too few players": `{"cf":{"pc":2,"an":10,"sc":200,"tr":10},"ps":[{"ch":1}],"ph":0,"rn":1}`,
		"invalid phase":   `{"cf":{"pc":2,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1}],"ph":9,"rn":1}`,
		"round zero":      `{"cf":{"pc":2,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1}],"ph":0,"rn":0}`,
		"negative pot":    `{"cf":{"pc":2,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1}],"ph":0,"rn":1,"pt":-1}`,
		"winner range":    `{"cf":{"pc":2,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1}],"ph":0,"rn":1,"wi":5}`,
		"result range":    `{"cf":{"pc":2,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1}],"ph":0,"rn":1,"re":9}`,
		"matcher range":   `{"cf":{"pc":2,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1}],"ph":0,"rn":1,"ma":[7]}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			var g domain.Guts
			assert.Error(t, json.Unmarshal([]byte(body), &g))
		})
	}
}

func TestGuts_UnmarshalDefaults(t *testing.T) {
	var g domain.Guts
	require.NoError(t, json.Unmarshal([]byte(`{"cf":{"pc":2,"an":10,"sc":200,"tr":10},"ps":[{"ch":200},{"ch":200}],"ph":0,"rn":1}`), &g))
	assert.Equal(t, 2, g.GetPlayerCnt())
	assert.NotNil(t, g.GetMatchers())
	assert.NotNil(t, g.GetActionLog())
}

// gutsGolden mirrors frontend/src/utils/__fixtures__/gutsGuide.golden.json.
type gutsGolden struct {
	Cases []struct {
		Name  string `json:"name"`
		Cards []struct{ Suit, Value int }
		Pair  bool   `json:"pair"`
		Tier  string `json:"tier"`
	} `json:"cases"`
}

// #5697: 宣言ガイドは Go (GutsEvaluateGuide) と TypeScript (evaluateGutsGuide) の
// 2 か所にある。片方だけ直せば CUI と Web が違う診断を出すので、同じ golden vector を
// 両方から検証して防ぐ。
func TestGutsEvaluateGuide_GoldenVectors(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(
		"..", "..", "frontend", "src", "utils", "__fixtures__", "gutsGuide.golden.json"))
	require.NoError(t, err)
	var golden gutsGolden
	require.NoError(t, json.Unmarshal(raw, &golden))
	require.NotEmpty(t, golden.Cases)

	tiers := map[string]bool{}
	for _, c := range golden.Cases {
		t.Run(c.Name, func(t *testing.T) {
			cards := make([]*domain.Card, 0, len(c.Cards))
			for _, cd := range c.Cards {
				cards = append(cards, domain.NewCard(cd.Suit, cd.Value, false))
			}

			guide := domain.GutsEvaluateGuide(cards)

			require.NotNil(t, guide)
			assert.Equal(t, c.Pair, guide.Pair)
			assert.Equal(t, c.Tier, guide.Tier)
		})
		tiers[c.Tier] = true
	}
	// 負のコントロール: 3 区分すべてを踏んでいない golden は境界を守れない。
	assert.Equal(t, map[string]bool{
		domain.GutsGuideTierHigh: true, domain.GutsGuideTierMedium: true, domain.GutsGuideTierLow: true,
	}, tiers)
}

// 手札が揃っていないときは診断できない (GutsEval が -1 を返す)。
func TestGutsEvaluateGuide_IncompleteHand(t *testing.T) {
	assert.Nil(t, domain.GutsEvaluateGuide(nil))
	assert.Nil(t, domain.GutsEvaluateGuide([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)}))
}
