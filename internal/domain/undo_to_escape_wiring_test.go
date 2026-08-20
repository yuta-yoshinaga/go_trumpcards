//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Each of the 40 solitaires delegates UndoToEscape to the shared undoToEscape
// helper, passing a closure that reads its own snapshot type's isStalemate
// field. The closure only runs when the game is stalemated AND history is
// non-empty, so the pre-existing per-game tests (which cover the "not
// stalemated" and "empty history" paths) never execute it. This table drives
// every game through the closure so a snapshot wired to the wrong field is
// caught.
func TestUndoToEscape_ClosureReadsEachGamesOwnSnapshot(t *testing.T) {
	games := []struct {
		name string
		run  func(stale []bool) int
	}{
		{"Accordion", func(stale []bool) int {
			g := &Accordion{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &accordionSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"AcesUp", func(stale []bool) int {
			g := &AcesUp{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &acesUpSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"AmericanToad", func(stale []bool) int {
			g := &AmericanToad{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &americanToadSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"AuldLangSyne", func(stale []bool) int {
			g := &AuldLangSyne{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &auldLangSyneSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"BakersDozen", func(stale []bool) int {
			g := &BakersDozen{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &bakersDozenSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"BeleagueredCastle", func(stale []bool) int {
			g := &BeleagueredCastle{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &beleagueredCastleSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Bristol", func(stale []bool) int {
			// Bristol は手詰まりをフィールドに持たず毎回計算する (#5993 のレビュー)。
			// 実際に詰んだ盤面を組んでから履歴を積む。
			g := &Bristol{phase: BristolPhasePlaying}
			suits := []int{CardDesignSpade, CardDesignHeart, CardDesignClover, CardDesignDiamond}
			for i := range g.tableau {
				g.tableau[i] = []*Card{NewCard(suits[i%len(suits)], 5, true)}
			}
			for _, st := range stale {
				g.history = append(g.history, &bristolSnapshot{isStalemate: st})
			}
			return g.UndoToEscape()
		}},
		{"Bisley", func(stale []bool) int {
			g := &Bisley{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &bisleySnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Braid", func(stale []bool) int {
			g := &Braid{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &braidSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Calculation", func(stale []bool) int {
			g := &Calculation{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &calculationSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Congress", func(stale []bool) int {
			g := &Congress{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &congressSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Crescent", func(stale []bool) int {
			g := &Crescent{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &crescentSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Cruel", func(stale []bool) int {
			g := &Cruel{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &cruelSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Duchess", func(stale []bool) int {
			g := &Duchess{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &duchessSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Easthaven", func(stale []bool) int {
			g := &Easthaven{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &easthavenSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"EightOff", func(stale []bool) int {
			g := &EightOff{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &eightOffSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"FlowerGarden", func(stale []bool) int {
			g := &FlowerGarden{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &flowerGardenSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"FortyAndEight", func(stale []bool) int {
			g := &FortyAndEight{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &fortyAndEightSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"FortyThieves", func(stale []bool) int {
			g := &FortyThieves{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &fortyThievesSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"FreeCell", func(stale []bool) int {
			g := &FreeCell{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &freeCellSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Gaps", func(stale []bool) int {
			g := &Gaps{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &gapsSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Golf", func(stale []bool) int {
			g := &Golf{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &golfSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"GrandfathersClock", func(stale []bool) int {
			g := &GrandfathersClock{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &grandfathersClockSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"KingAlbert", func(stale []bool) int {
			g := &KingAlbert{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &kingAlbertSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Klondike", func(stale []bool) int {
			g := &Klondike{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &klondikeSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"MissMilligan", func(stale []bool) int {
			g := &MissMilligan{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &missMilliganSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"NapoleonsSquare", func(stale []bool) int {
			g := &NapoleonsSquare{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &napoleonsSquareSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Penguin", func(stale []bool) int {
			g := &Penguin{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &penguinSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Pyramid", func(stale []bool) int {
			g := &Pyramid{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &pyramidSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"RussianSolitaire", func(stale []bool) int {
			g := &RussianSolitaire{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &russianSolitaireSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Scorpion", func(stale []bool) int {
			g := &Scorpion{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &scorpionSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"SeahavenTowers", func(stale []bool) int {
			g := &SeahavenTowers{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &seahavenTowersSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"SirTommy", func(stale []bool) int {
			g := &SirTommy{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &sirTommySnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Spider", func(stale []bool) int {
			g := &Spider{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &spiderSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Spiderette", func(stale []bool) int {
			g := &Spiderette{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &spideretteSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"StreetsAndAlleys", func(stale []bool) int {
			g := &StreetsAndAlleys{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &streetsAndAlleysSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Sultan", func(stale []bool) int {
			g := &Sultan{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &sultanSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Terrace", func(stale []bool) int {
			g := &Terrace{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &terraceSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"TriPeaks", func(stale []bool) int {
			g := &TriPeaks{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &triPeaksSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Wasp", func(stale []bool) int {
			g := &Wasp{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &waspSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Windmill", func(stale []bool) int {
			g := &Windmill{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &windmillSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
		{"Yukon", func(stale []bool) int {
			g := &Yukon{isStalemate: true}
			for _, s := range stale {
				g.history = append(g.history, &yukonSnapshot{isStalemate: s})
			}
			return g.UndoToEscape()
		}},
	}

	// Distance is counted from the end of history, so the flags must be read
	// per element: a closure hardwired to false would return 1 for both cases,
	// and one hardwired to true would return -1.
	scenarios := []struct {
		name  string
		stale []bool
		want  int
	}{
		{"escape is the most recent snapshot", []bool{true, false}, 1},
		{"must walk past a stalemated tail", []bool{false, true}, 2},
		{"no escape anywhere", []bool{true, true}, -1},
	}

	assert.Len(t, games, 42, "every game delegating to undoToEscape must be listed")

	for _, g := range games {
		for _, sc := range scenarios {
			assert.Equal(t, sc.want, g.run(sc.stale), "%s: %s", g.name, sc.name)
		}
	}
}
