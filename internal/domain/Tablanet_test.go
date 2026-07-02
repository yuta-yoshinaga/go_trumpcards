//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// tablanetCard は Tablanet テスト用のカード生成ショートカット。
func tablanetCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// newTestTablanet は 4 人 (human=0) の Tablanet を Reset 済みで返す。
func newTestTablanet(t *testing.T, diff domain.TablanetCpuDifficulty) *domain.Tablanet {
	t.Helper()
	players := make([]*domain.TablanetPlayer, domain.TablanetPlayerCnt)
	players[0] = domain.NewTablanetPlayer(true)
	for i := 1; i < domain.TablanetPlayerCnt; i++ {
		players[i] = domain.NewTablanetPlayer(false)
	}
	cfg := domain.DefaultTablanetConfig()
	cfg.CpuDifficulty = diff
	g := domain.NewTablanet(domain.NewTrumpCards(0), players, cfg)
	g.Reset()
	return g
}

// setTablanetHand はプレイヤー idx の手札を指定カードで上書きする。
func setTablanetHand(g *domain.Tablanet, idx int, cards ...*domain.Card) {
	p := g.GetPlayer(idx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// tablanetTotalCards はすべての場所に散らばるカード総数を返す (常に 52 のはず)。
func tablanetTotalCards(g *domain.Tablanet) int {
	total := g.GetRemainingDeck() + len(g.GetTableCards())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		total += p.GetCardsSize() + p.CapturedCount()
	}
	return total
}

// tablanetDrive は人間手札[0]の最大捕獲でプレイしつつ終局まで進める。
func tablanetDrive(t *testing.T, g *domain.Tablanet) {
	t.Helper()
	for i := 0; i < 20000 && !g.GetGameEndFlag(); i++ {
		if g.GetPhase() != domain.TablanetPhasePlay {
			break
		}
		require.Equal(t, 52, tablanetTotalCards(g), "card conservation violated mid-game")
		if g.IsHumanTurn() {
			opts := g.GetCaptureOptions(0)
			require.NoError(t, g.PlayerPlay(0, opts[0]))
		} else {
			g.CpuPlay()
		}
	}
}

func TestTablanetResetInitialDeal(t *testing.T) {
	g := newTestTablanet(t, domain.TablanetCpuDifficultyNormal)
	assert.Equal(t, domain.TablanetPhasePlay, g.GetPhase())
	assert.Equal(t, 0, g.GetCurrentTurn())
	assert.Len(t, g.GetTableCards(), domain.TablanetInitialTableSize)
	for i := 0; i < domain.TablanetPlayerCnt; i++ {
		assert.Equal(t, domain.TablanetHandSize, g.GetPlayer(i).GetCardsSize())
	}
	// 52 - 16 (hands) - 4 (table) = 32.
	assert.Equal(t, 32, g.GetRemainingDeck())
	assert.Equal(t, 52, tablanetTotalCards(g))
}

func TestTablanetRankCaptureIsTablanet(t *testing.T) {
	g := newTestTablanet(t, domain.TablanetCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{tablanetCard(domain.CardDesignSpade, 5)})
	setTablanetHand(g, 0, tablanetCard(domain.CardDesignHeart, 5))

	require.NoError(t, g.PlayerPlay(0, []int{0}))
	assert.Empty(t, g.GetTableCards())
	assert.Equal(t, 2, g.GetPlayer(0).CapturedCount())
	assert.Equal(t, 1, g.GetPlayer(0).GetTablaCount(), "clearing the table is a Tablanet")
	assert.Equal(t, 0, g.GetLastCaptureIdx())
}

func TestTablanetSumCapture(t *testing.T) {
	g := newTestTablanet(t, domain.TablanetCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{
		tablanetCard(domain.CardDesignSpade, 2),
		tablanetCard(domain.CardDesignHeart, 3),
		tablanetCard(domain.CardDesignClover, 9),
	})
	setTablanetHand(g, 0, tablanetCard(domain.CardDesignDiamond, 5))

	require.NoError(t, g.PlayerPlay(0, []int{0, 1}))
	// 2+3 = 5 captured; the 9 remains, so no Tablanet.
	assert.Len(t, g.GetTableCards(), 1)
	assert.Equal(t, 3, g.GetPlayer(0).CapturedCount())
	assert.Equal(t, 0, g.GetPlayer(0).GetTablaCount())
}

func TestTablanetJackSweep(t *testing.T) {
	g := newTestTablanet(t, domain.TablanetCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{
		tablanetCard(domain.CardDesignSpade, 5),
		tablanetCard(domain.CardDesignHeart, 9),
		tablanetCard(domain.CardDesignClover, domain.TablanetJackValue),
	})
	setTablanetHand(g, 0, tablanetCard(domain.CardDesignDiamond, domain.TablanetJackValue))

	// Jack sweeps everything except the other Jack; table selection is ignored.
	require.NoError(t, g.PlayerPlay(0, nil))
	require.Len(t, g.GetTableCards(), 1)
	assert.Equal(t, domain.TablanetJackValue, g.GetTableCards()[0].GetValue())
	assert.Equal(t, 3, g.GetPlayer(0).CapturedCount()) // jack + 5 + 9
	assert.Equal(t, 0, g.GetPlayer(0).GetTablaCount(), "a Jack sweep is not a Tablanet")
}

func TestTablanetTrail(t *testing.T) {
	g := newTestTablanet(t, domain.TablanetCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{tablanetCard(domain.CardDesignSpade, 5)})
	setTablanetHand(g, 0, tablanetCard(domain.CardDesignHeart, 9))

	require.NoError(t, g.PlayerPlay(0, nil))
	assert.Len(t, g.GetTableCards(), 2)
	assert.Equal(t, 0, g.GetPlayer(0).CapturedCount())
}

func TestTablanetInvalidSelection(t *testing.T) {
	g := newTestTablanet(t, domain.TablanetCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{tablanetCard(domain.CardDesignSpade, 5)})
	setTablanetHand(g, 0, tablanetCard(domain.CardDesignHeart, 9))

	err := g.PlayerPlay(0, []int{0})
	require.Error(t, err, "9 cannot capture a lone 5")
	// Hand and table unchanged after a rejected play.
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
	assert.Len(t, g.GetTableCards(), 1)
}

func TestTablanetFaceCardRankOnly(t *testing.T) {
	g := newTestTablanet(t, domain.TablanetCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{
		tablanetCard(domain.CardDesignSpade, 12), // Q
		tablanetCard(domain.CardDesignHeart, 3),
	})
	setTablanetHand(g, 0, tablanetCard(domain.CardDesignDiamond, 12)) // Q

	// Cannot capture the 3 with a Queen.
	require.Error(t, g.PlayerPlay(0, []int{1}))
	// Rank-match capture is valid.
	require.NoError(t, g.PlayerPlay(0, []int{0}))
	assert.Len(t, g.GetTableCards(), 1)
	assert.Equal(t, 2, g.GetPlayer(0).CapturedCount())
}

func TestTablanetPlayGuards(t *testing.T) {
	g := newTestTablanet(t, domain.TablanetCpuDifficultyNormal)
	// Out-of-range index.
	g.SetCurrentTurn(0)
	require.Error(t, g.PlayerPlay(99, nil))
	// Not the human's turn.
	g.SetCurrentTurn(1)
	require.Error(t, g.PlayerPlay(0, nil))
	// Wrong phase.
	g.SetCurrentTurn(0)
	g.SetPhase(domain.TablanetPhaseGameEnd)
	require.Error(t, g.PlayerPlay(0, nil))
}

func TestTablanetFullDealEachDifficulty(t *testing.T) {
	for _, diff := range []domain.TablanetCpuDifficulty{
		domain.TablanetCpuDifficultyEasy,
		domain.TablanetCpuDifficultyNormal,
		domain.TablanetCpuDifficultyHard,
	} {
		g := newTestTablanet(t, diff)
		tablanetDrive(t, g)
		require.True(t, g.GetGameEndFlag(), "game should end for difficulty %d", diff)
		assert.Equal(t, domain.TablanetPhaseGameEnd, g.GetPhase())
		require.NotNil(t, g.GetLastDealDetail())
		assert.NotEmpty(t, g.GetWinners())
		// The whole deck should end up captured (table cleared to last capturer).
		captured := 0
		for i := 0; i < g.GetPlayerCnt(); i++ {
			captured += g.GetPlayer(i).CapturedCount()
		}
		assert.Equal(t, 52, captured, "all 52 cards captured at game end")
		assert.Empty(t, g.GetTableCards())
	}
}

func TestTablanetScoringBonuses(t *testing.T) {
	g := newTestTablanet(t, domain.TablanetCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{tablanetCard(domain.CardDesignSpade, 5)})
	// Empty every hand except the human.
	for i := 1; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).Reset()
	}
	// A single non-Jack card that clears the whole table scores a Tabla bonus.
	setTablanetHand(g, 0, tablanetCard(domain.CardDesignHeart, 5))
	require.NoError(t, g.PlayerPlay(0, []int{0})) // Tabla: captures the lone 5
	assert.GreaterOrEqual(t, g.GetPlayer(0).GetTablaCount(), 1)
}

func TestTablanetNextRoundResets(t *testing.T) {
	g := newTestTablanet(t, domain.TablanetCpuDifficultyNormal)
	tablanetDrive(t, g)
	require.True(t, g.GetGameEndFlag())
	g.NextRound()
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, domain.TablanetPhasePlay, g.GetPhase())
	assert.Len(t, g.GetTableCards(), domain.TablanetInitialTableSize)
	assert.Equal(t, 52, tablanetTotalCards(g))
}

func TestTablanetHint(t *testing.T) {
	g := newTestTablanet(t, domain.TablanetCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{tablanetCard(domain.CardDesignSpade, 5)})
	setTablanetHand(g, 0, tablanetCard(domain.CardDesignHeart, 5), tablanetCard(domain.CardDesignClover, 9))
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.NotEmpty(t, hint.CardIndices)
	assert.Equal(t, "tabla_sweep", hint.Reason)
}

func TestTablanetHintNilWhenNotHumanTurn(t *testing.T) {
	g := newTestTablanet(t, domain.TablanetCpuDifficultyNormal)
	g.SetCurrentTurn(1)
	assert.Nil(t, g.GetHint())
}

func TestTablanetConfigValidate(t *testing.T) {
	cfg := domain.DefaultTablanetConfig()
	require.NoError(t, cfg.Validate())
	cfg.CpuDifficulty = domain.TablanetCpuDifficulty(99)
	require.Error(t, cfg.Validate())
}

func TestTablanetJSONRoundTrip(t *testing.T) {
	g := newTestTablanet(t, domain.TablanetCpuDifficultyHard)
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored domain.Tablanet
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetCurrentTurn(), restored.GetCurrentTurn())
	assert.Equal(t, len(g.GetTableCards()), len(restored.GetTableCards()))
	assert.Equal(t, g.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, 52, tablanetTotalCards(&restored))
}

func TestTablanetUnmarshalRejectsInvalid(t *testing.T) {
	valid := newTestTablanet(t, domain.TablanetCpuDifficultyNormal)
	good, err := json.Marshal(valid)
	require.NoError(t, err)

	// Sanity: the good payload round-trips.
	var ok domain.Tablanet
	require.NoError(t, json.Unmarshal(good, &ok))

	cases := []string{
		`{"ph":9,"pl":[{},{},{},{}],"tc":{},"ct":0,"lc":-1}`, // invalid phase
		`{"ph":0,"pl":[{},{}],"tc":{},"ct":0,"lc":-1}`,       // wrong player count
		`{"ph":0,"pl":[{},{},{},{}],"ct":99,"lc":-1}`,        // missing trump cards
		`not json`, // malformed
	}
	for _, c := range cases {
		var b domain.Tablanet
		assert.Error(t, json.Unmarshal([]byte(c), &b), "should reject: %s", c)
	}
}
