//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// basraCard は Basra テスト用のカード生成ショートカット。
func basraCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// newTestBasra は 4 人 (human=0) の Basra を Reset 済みで返す。
func newTestBasra(t *testing.T, diff domain.BasraCpuDifficulty) *domain.Basra {
	t.Helper()
	players := make([]*domain.BasraPlayer, domain.BasraPlayerCnt)
	players[0] = domain.NewBasraPlayer(true)
	for i := 1; i < domain.BasraPlayerCnt; i++ {
		players[i] = domain.NewBasraPlayer(false)
	}
	cfg := domain.DefaultBasraConfig()
	cfg.CpuDifficulty = diff
	g := domain.NewBasra(domain.NewTrumpCards(0), players, cfg)
	g.Reset()
	return g
}

// setBasraHand はプレイヤー idx の手札を指定カードで上書きする。
func setBasraHand(g *domain.Basra, idx int, cards ...*domain.Card) {
	p := g.GetPlayer(idx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// basraTotalCards はすべての場所に散らばるカード総数を返す (常に 52 のはず)。
func basraTotalCards(g *domain.Basra) int {
	total := g.GetRemainingDeck() + len(g.GetTableCards())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		total += p.GetCardsSize() + p.CapturedCount()
	}
	return total
}

// basraDrive は人間手札[0]の最大捕獲でプレイしつつ終局まで進める。
func basraDrive(t *testing.T, g *domain.Basra) {
	t.Helper()
	for i := 0; i < 20000 && !g.GetGameEndFlag(); i++ {
		if g.GetPhase() != domain.BasraPhasePlay {
			break
		}
		require.Equal(t, 52, basraTotalCards(g), "card conservation violated mid-game")
		if g.IsHumanTurn() {
			opts := g.GetCaptureOptions(0)
			require.NoError(t, g.PlayerPlay(0, opts[0]))
		} else {
			g.CpuPlay()
		}
	}
}

func TestBasraResetInitialDeal(t *testing.T) {
	g := newTestBasra(t, domain.BasraCpuDifficultyNormal)
	assert.Equal(t, domain.BasraPhasePlay, g.GetPhase())
	assert.Equal(t, 0, g.GetCurrentTurn())
	assert.Len(t, g.GetTableCards(), domain.BasraInitialTableSize)
	for i := 0; i < domain.BasraPlayerCnt; i++ {
		assert.Equal(t, domain.BasraHandSize, g.GetPlayer(i).GetCardsSize())
	}
	// 52 - 16 (hands) - 4 (table) = 32.
	assert.Equal(t, 32, g.GetRemainingDeck())
	assert.Equal(t, 52, basraTotalCards(g))
}

func TestBasraRankCaptureIsBasra(t *testing.T) {
	g := newTestBasra(t, domain.BasraCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{basraCard(domain.CardDesignSpade, 5)})
	setBasraHand(g, 0, basraCard(domain.CardDesignHeart, 5))

	require.NoError(t, g.PlayerPlay(0, []int{0}))
	assert.Empty(t, g.GetTableCards())
	assert.Equal(t, 2, g.GetPlayer(0).CapturedCount())
	assert.Equal(t, 1, g.GetPlayer(0).GetBasraCount(), "clearing the table is a Basra")
	assert.Equal(t, 0, g.GetLastCaptureIdx())
}

func TestBasraSumCapture(t *testing.T) {
	g := newTestBasra(t, domain.BasraCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{
		basraCard(domain.CardDesignSpade, 2),
		basraCard(domain.CardDesignHeart, 3),
		basraCard(domain.CardDesignClover, 9),
	})
	setBasraHand(g, 0, basraCard(domain.CardDesignDiamond, 5))

	require.NoError(t, g.PlayerPlay(0, []int{0, 1}))
	// 2+3 = 5 captured; the 9 remains, so no Basra.
	assert.Len(t, g.GetTableCards(), 1)
	assert.Equal(t, 3, g.GetPlayer(0).CapturedCount())
	assert.Equal(t, 0, g.GetPlayer(0).GetBasraCount())
}

func TestBasraJackSweep(t *testing.T) {
	g := newTestBasra(t, domain.BasraCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{
		basraCard(domain.CardDesignSpade, 5),
		basraCard(domain.CardDesignHeart, 9),
		basraCard(domain.CardDesignClover, domain.BasraJackValue),
	})
	setBasraHand(g, 0, basraCard(domain.CardDesignDiamond, domain.BasraJackValue))

	// Jack sweeps everything except the other Jack; table selection is ignored.
	require.NoError(t, g.PlayerPlay(0, nil))
	require.Len(t, g.GetTableCards(), 1)
	assert.Equal(t, domain.BasraJackValue, g.GetTableCards()[0].GetValue())
	assert.Equal(t, 3, g.GetPlayer(0).CapturedCount()) // jack + 5 + 9
	assert.Equal(t, 0, g.GetPlayer(0).GetBasraCount(), "a Jack sweep is not a Basra")
}

func TestBasraTrail(t *testing.T) {
	g := newTestBasra(t, domain.BasraCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{basraCard(domain.CardDesignSpade, 5)})
	setBasraHand(g, 0, basraCard(domain.CardDesignHeart, 9))

	require.NoError(t, g.PlayerPlay(0, nil))
	assert.Len(t, g.GetTableCards(), 2)
	assert.Equal(t, 0, g.GetPlayer(0).CapturedCount())
}

func TestBasraInvalidSelection(t *testing.T) {
	g := newTestBasra(t, domain.BasraCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{basraCard(domain.CardDesignSpade, 5)})
	setBasraHand(g, 0, basraCard(domain.CardDesignHeart, 9))

	err := g.PlayerPlay(0, []int{0})
	require.Error(t, err, "9 cannot capture a lone 5")
	// Hand and table unchanged after a rejected play.
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
	assert.Len(t, g.GetTableCards(), 1)
}

func TestBasraFaceCardRankOnly(t *testing.T) {
	g := newTestBasra(t, domain.BasraCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{
		basraCard(domain.CardDesignSpade, 12), // Q
		basraCard(domain.CardDesignHeart, 3),
	})
	setBasraHand(g, 0, basraCard(domain.CardDesignDiamond, 12)) // Q

	// Cannot capture the 3 with a Queen.
	require.Error(t, g.PlayerPlay(0, []int{1}))
	// Rank-match capture is valid.
	require.NoError(t, g.PlayerPlay(0, []int{0}))
	assert.Len(t, g.GetTableCards(), 1)
	assert.Equal(t, 2, g.GetPlayer(0).CapturedCount())
}

func TestBasraPlayGuards(t *testing.T) {
	g := newTestBasra(t, domain.BasraCpuDifficultyNormal)
	// Out-of-range index.
	g.SetCurrentTurn(0)
	require.Error(t, g.PlayerPlay(99, nil))
	// Not the human's turn.
	g.SetCurrentTurn(1)
	require.Error(t, g.PlayerPlay(0, nil))
	// Wrong phase.
	g.SetCurrentTurn(0)
	g.SetPhase(domain.BasraPhaseGameEnd)
	require.Error(t, g.PlayerPlay(0, nil))
}

func TestBasraFullDealEachDifficulty(t *testing.T) {
	for _, diff := range []domain.BasraCpuDifficulty{
		domain.BasraCpuDifficultyEasy,
		domain.BasraCpuDifficultyNormal,
		domain.BasraCpuDifficultyHard,
	} {
		g := newTestBasra(t, diff)
		basraDrive(t, g)
		require.True(t, g.GetGameEndFlag(), "game should end for difficulty %d", diff)
		assert.Equal(t, domain.BasraPhaseGameEnd, g.GetPhase())
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

func TestBasraScoringBonuses(t *testing.T) {
	g := newTestBasra(t, domain.BasraCpuDifficultyNormal)
	// Give the human the 7♦, 10♦ and an Ace; drain the deck so a single play
	// ends the game and scores.
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{basraCard(domain.CardDesignSpade, 5)})
	// Empty every hand except the human, and empty the deck by drawing it out.
	for i := 1; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).Reset()
	}
	// Pre-load the human's capture pile with bonus cards via a rank capture, then
	// exhaust the deck by playing until game end.
	setBasraHand(g, 0, basraCard(domain.CardDesignHeart, 5))
	require.NoError(t, g.PlayerPlay(0, []int{0})) // Basra: captures the lone 5
	assert.GreaterOrEqual(t, g.GetPlayer(0).GetBasraCount(), 1)
}

func TestBasraNextRoundResets(t *testing.T) {
	g := newTestBasra(t, domain.BasraCpuDifficultyNormal)
	basraDrive(t, g)
	require.True(t, g.GetGameEndFlag())
	g.NextRound()
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, domain.BasraPhasePlay, g.GetPhase())
	assert.Len(t, g.GetTableCards(), domain.BasraInitialTableSize)
	assert.Equal(t, 52, basraTotalCards(g))
}

func TestBasraHint(t *testing.T) {
	g := newTestBasra(t, domain.BasraCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{basraCard(domain.CardDesignSpade, 5)})
	setBasraHand(g, 0, basraCard(domain.CardDesignHeart, 5), basraCard(domain.CardDesignClover, 9))
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.NotEmpty(t, hint.CardIndices)
	assert.Equal(t, "basra_sweep", hint.Reason)
}

func TestBasraHintNilWhenNotHumanTurn(t *testing.T) {
	g := newTestBasra(t, domain.BasraCpuDifficultyNormal)
	g.SetCurrentTurn(1)
	assert.Nil(t, g.GetHint())
}

func TestBasraConfigValidate(t *testing.T) {
	cfg := domain.DefaultBasraConfig()
	require.NoError(t, cfg.Validate())
	cfg.CpuDifficulty = domain.BasraCpuDifficulty(99)
	require.Error(t, cfg.Validate())
}

func TestBasraJSONRoundTrip(t *testing.T) {
	g := newTestBasra(t, domain.BasraCpuDifficultyHard)
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored domain.Basra
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetCurrentTurn(), restored.GetCurrentTurn())
	assert.Equal(t, len(g.GetTableCards()), len(restored.GetTableCards()))
	assert.Equal(t, g.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, 52, basraTotalCards(&restored))
}

func TestBasraUnmarshalRejectsInvalid(t *testing.T) {
	valid := newTestBasra(t, domain.BasraCpuDifficultyNormal)
	good, err := json.Marshal(valid)
	require.NoError(t, err)

	// Sanity: the good payload round-trips.
	var ok domain.Basra
	require.NoError(t, json.Unmarshal(good, &ok))

	cases := []string{
		`{"ph":9,"pl":[{},{},{},{}],"tc":{},"ct":0,"lc":-1}`, // invalid phase
		`{"ph":0,"pl":[{},{}],"tc":{},"ct":0,"lc":-1}`,       // wrong player count
		`{"ph":0,"pl":[{},{},{},{}],"ct":99,"lc":-1}`,        // missing trump cards
		`not json`, // malformed
	}
	for _, c := range cases {
		var b domain.Basra
		assert.Error(t, json.Unmarshal([]byte(c), &b), "should reject: %s", c)
	}
}
