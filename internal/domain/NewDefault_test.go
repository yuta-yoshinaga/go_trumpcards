//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// playerCounter is implemented by every game type that exposes player count.
// Used in table-driven NewDefault tests to verify the factory wires up the
// expected (1 human, N CPU) seating without depending on game-specific APIs.
type playerCounter interface {
	GetPlayerCnt() int
}

// assertHumanFirst checks that a multi-player game has the human seated at
// index 0 and CPUs at indices 1..N-1.
func assertHumanFirst(t *testing.T, g playerCounter, isHumanAt func(int) bool, expectedCnt int) {
	t.Helper()
	assert.Equal(t, expectedCnt, g.GetPlayerCnt(), "player count mismatch")
	assert.True(t, isHumanAt(0), "player 0 must be human")
	for i := 1; i < g.GetPlayerCnt(); i++ {
		assert.False(t, isHumanAt(i), "player %d must be CPU", i)
	}
}

func TestNewDefaultDaifugo(t *testing.T) {
	g := domain.NewDefaultDaifugo()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, domain.DaifugoPlayerCnt)
}

func TestNewDefaultSevens(t *testing.T) {
	g := domain.NewDefaultSevens()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, domain.SevensPlayerCnt)
}

func TestNewDefaultDoubt(t *testing.T) {
	g := domain.NewDefaultDoubt()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, 4)
}

func TestNewDefaultOldMaid(t *testing.T) {
	g := domain.NewDefaultOldMaid()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, 4)
}

func TestNewDefaultGoFish(t *testing.T) {
	g := domain.NewDefaultGoFish()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, 4)
}

func TestNewDefaultPigsTail(t *testing.T) {
	g := domain.NewDefaultPigsTail()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, 4)
}

func TestNewDefaultFiftyOne(t *testing.T) {
	g := domain.NewDefaultFiftyOne()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, 4)
}

func TestNewDefaultDurak(t *testing.T) {
	g := domain.NewDefaultDurak()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, 4)
}

func TestNewDefaultPageOne(t *testing.T) {
	g := domain.NewDefaultPageOne()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, 4)
}

func TestNewDefaultTwoTenJack(t *testing.T) {
	g := domain.NewDefaultTwoTenJack()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, 4)
}

func TestNewDefaultNapoleon(t *testing.T) {
	g := domain.NewDefaultNapoleon()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, 4)
}

func TestNewDefaultPoker(t *testing.T) {
	g := domain.NewDefaultPoker()
	assert.NotNil(t, g)
	players := g.GetPlayers()
	assert.Len(t, players, 4)
	assert.True(t, players[0].GetIsHuman(), "player 0 must be human")
	for i := 1; i < len(players); i++ {
		assert.False(t, players[i].GetIsHuman(), "player %d must be CPU", i)
	}
}

// Team-based 4-player games: human is team 0, alternating teams.

func TestNewDefaultBridge(t *testing.T) {
	g := domain.NewDefaultBridge()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, 4)
	assert.Equal(t, 0, g.GetPlayer(0).GetTeam())
	assert.Equal(t, 1, g.GetPlayer(1).GetTeam())
}

func TestNewDefaultEuchre(t *testing.T) {
	g := domain.NewDefaultEuchre()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, 4)
	assert.Equal(t, 0, g.GetPlayer(0).GetTeam())
	assert.Equal(t, 1, g.GetPlayer(1).GetTeam())
}

func TestNewDefaultPinochle(t *testing.T) {
	g := domain.NewDefaultPinochle()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, 4)
	assert.Equal(t, 0, g.GetPlayer(0).GetTeam())
	assert.Equal(t, 1, g.GetPlayer(1).GetTeam())
}

func TestNewDefaultWhist(t *testing.T) {
	g := domain.NewDefaultWhist()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, 4)
	assert.Equal(t, 0, g.GetPlayer(0).GetTeam())
	assert.Equal(t, 1, g.GetPlayer(1).GetTeam())
}

// 2-player games.

func TestNewDefaultGinRummy(t *testing.T) {
	g := domain.NewDefaultGinRummy()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, 2)
}

func TestNewDefaultCribbage(t *testing.T) {
	g := domain.NewDefaultCribbage()
	assert.NotNil(t, g)
	assert.True(t, g.GetPlayer(0).GetIsHuman(), "player 0 must be human")
	for i := 1; i < domain.CribbagePlayerCnt; i++ {
		assert.False(t, g.GetPlayer(i).GetIsHuman(), "player %d must be CPU", i)
	}
}

func TestNewDefaultSpeed(t *testing.T) {
	g := domain.NewDefaultSpeed()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, domain.SpeedPlayerCnt)
}

func TestNewDefaultWar(t *testing.T) {
	g := domain.NewDefaultWar()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, domain.WarPlayerCnt)
}

func TestNewDefaultCanasta(t *testing.T) {
	g := domain.NewDefaultCanasta()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, 2)
}

// Solitaire and single-deck card games (no player count to assert).

func TestNewDefaultKlondike(t *testing.T) {
	assert.NotNil(t, domain.NewDefaultKlondike())
}

func TestNewDefaultFreeCell(t *testing.T) {
	assert.NotNil(t, domain.NewDefaultFreeCell())
}

func TestNewDefaultSpider(t *testing.T) {
	assert.NotNil(t, domain.NewDefaultSpider())
}

func TestNewDefaultPyramid(t *testing.T) {
	assert.NotNil(t, domain.NewDefaultPyramid())
}

func TestNewDefaultTriPeaks(t *testing.T) {
	assert.NotNil(t, domain.NewDefaultTriPeaks())
}

func TestNewDefaultGolf(t *testing.T) {
	assert.NotNil(t, domain.NewDefaultGolf())
}

func TestNewDefaultClockSolitaire(t *testing.T) {
	assert.NotNil(t, domain.NewDefaultClockSolitaire())
}

func TestNewDefaultCanfield(t *testing.T) {
	assert.NotNil(t, domain.NewDefaultCanfield())
}

func TestNewDefaultFortyThieves(t *testing.T) {
	assert.NotNil(t, domain.NewDefaultFortyThieves())
}

func TestNewDefaultYukon(t *testing.T) {
	assert.NotNil(t, domain.NewDefaultYukon())
}

func TestNewDefaultScorpion(t *testing.T) {
	assert.NotNil(t, domain.NewDefaultScorpion())
}

func TestNewDefaultPokerSquares(t *testing.T) {
	assert.NotNil(t, domain.NewDefaultPokerSquares())
}

// Table-size based community-card games. We verify human is at seat 0
// against the configured table size for each.

func TestNewDefaultHoldem(t *testing.T) {
	cfg := domain.DefaultHoldemConfig()
	g := domain.NewDefaultHoldem()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, cfg.TableSize)
}

func TestNewDefaultOmaha(t *testing.T) {
	cfg := domain.DefaultOmahaConfig()
	g := domain.NewDefaultOmaha()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, cfg.TableSize)
}

func TestNewDefaultShortDeck(t *testing.T) {
	cfg := domain.DefaultShortDeckConfig()
	g := domain.NewDefaultShortDeck()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, cfg.TableSize)
}

func TestNewDefaultPineapple(t *testing.T) {
	cfg := domain.DefaultPineappleConfig()
	g := domain.NewDefaultPineapple()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, cfg.TableSize)
}

func TestNewDefaultSevenCardStud(t *testing.T) {
	cfg := domain.DefaultSevenCardStudConfig()
	g := domain.NewDefaultSevenCardStud()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, cfg.TableSize)
	assert.False(t, g.GetIsLowball(), "stud should not be lowball")
}

func TestNewDefaultRazz(t *testing.T) {
	cfg := domain.DefaultRazzConfig()
	g := domain.NewDefaultRazz()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, cfg.TableSize)
	assert.True(t, g.GetIsLowball(), "razz must be lowball")
}

func TestNewDefaultIndianPoker(t *testing.T) {
	g := domain.NewDefaultIndianPoker()
	assert.NotNil(t, g)
	assertHumanFirst(t, g, func(i int) bool { return g.GetPlayer(i).GetIsHuman() }, g.GetPlayerCnt())
}
