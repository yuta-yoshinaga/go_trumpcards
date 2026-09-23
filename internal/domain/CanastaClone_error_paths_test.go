//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func requireCode(t *testing.T, err error, code string) {
	t.Helper()
	require.Error(t, err)
	got, _ := ErrorMessageCode(err)
	require.Equal(t, code, got)
}
func testBoliviaPathGame() (*Bolivia, *BoliviaPlayer) {
	ps := make([]*BoliviaPlayer, BoliviaPlayerCnt)
	for i := range ps {
		ps[i] = NewBoliviaPlayer(i == 0, i%BoliviaTeamCnt)
	}
	g := NewBolivia(NewTrumpCardsWithDecks(3, 6), ps, DefaultBoliviaConfig())
	g.Reset()
	p := g.GetPlayer(0)
	p.Reset()
	return g, p
}
func testSambaPathGame() (*Samba, *SambaPlayer) {
	ps := make([]*SambaPlayer, SambaPlayerCnt)
	for i := range ps {
		ps[i] = NewSambaPlayer(i == 0, i%SambaTeamCnt)
	}
	g := NewSamba(NewTrumpCardsWithDecks(3, 6), ps, DefaultSambaConfig())
	g.Reset()
	p := g.GetPlayer(0)
	p.Reset()
	return g, p
}
func testHAFPathGame() (*HandAndFoot, *HandAndFootPlayer) {
	ps := []*HandAndFootPlayer{NewHandAndFootPlayer(true), NewHandAndFootPlayer(false), NewHandAndFootPlayer(false), NewHandAndFootPlayer(false)}
	g := NewHandAndFoot(NewTrumpCardsWithDecks(4, 8), ps, DefaultHandAndFootConfig())
	g.Reset()
	p := g.GetPlayer(0)
	p.Reset()
	return g, p
}
func cloneCard(d, v int) *Card { return NewCard(d, v, false) }

func TestBolivia_ErrorPathLines(t *testing.T) {
	g, p := testBoliviaPathGame()
	g.SetDiscardPile([]*Card{cloneCard(CardDesignSpade, 7)})
	p.SetHasInitMeld(true)
	p.AddCard(cloneCard(CardDesignHeart, 7))
	p.AddCard(cloneCard(CardDesignDiamond, 7))
	require.NoError(t, g.PlayerDrawFromDiscard([]int{0, 1}))
	requireCode(t, g.PlayerMeld(nil), "bolivia.errTopCardMustBeMelded")
	g, p = testBoliviaPathGame()
	g.phase, g.drewFromDiscard, g.drawnCard = BoliviaPhaseMeld, true, cloneCard(CardDesignSpade, 9)
	p.SetHasInitMeld(true)
	for i := 0; i < 3; i++ {
		p.AddCard(cloneCard(CardDesignSpade+i, 7))
	}
	requireCode(t, g.PlayerMeld([][]int{{0, 1, 2}}), "bolivia.errTopCardMustBeMelded")
	g, _ = testBoliviaPathGame()
	requireCode(t, func() error { _, err := g.resolveMeldGroup(0, []*Card{cloneCard(CardDesignJoker, 1)}); return err }(), "bolivia.errMeldNeedsAtLeastThreeCards")
	for _, tc := range []struct {
		cards []*Card
		code  string
	}{{[]*Card{cloneCard(CardDesignSpade, 7), cloneCard(CardDesignHeart, 7), cloneCard(CardDesignJoker, 1), cloneCard(CardDesignJoker, 2), cloneCard(CardDesignJoker, 3), cloneCard(CardDesignJoker, 4)}, "bolivia.errMeldAllowsAtMostThreeWildCards"}, {[]*Card{cloneCard(CardDesignSpade, 7), cloneCard(CardDesignHeart, 7), cloneCard(CardDesignJoker, 1), cloneCard(CardDesignJoker, 2), cloneCard(CardDesignJoker, 3)}, "bolivia.errWildCardsCannotExceedNaturalCards"}} {
		requireCode(t, g.validateNewSet(tc.cards), tc.code)
	}
	requireCode(t, g.validateNewSet([]*Card{cloneCard(CardDesignSpade, 7), cloneCard(CardDesignHeart, 8), cloneCard(CardDesignDiamond, 7)}), "bolivia.errSetMeldCardsMustHaveSameRank")
	for _, tc := range []struct {
		cards []*Card
		code  string
	}{{[]*Card{cloneCard(CardDesignSpade, 7), cloneCard(CardDesignHeart, 8)}, "bolivia.errMeldCardRankMismatch"}, {[]*Card{cloneCard(CardDesignClover, 3)}, "bolivia.errBlackThreeCannotMeld"}, {[]*Card{cloneCard(CardDesignJoker, 1)}, "bolivia.errMeldAllowsAtMostThreeWildCards"}} {
		requireCode(t, g.validateSetAddition(&BoliviaMeld{Cards: []*Card{cloneCard(CardDesignSpade, 7), cloneCard(CardDesignHeart, 7), cloneCard(CardDesignJoker, 1), cloneCard(CardDesignJoker, 2), cloneCard(CardDesignJoker, 3)}, Kind: BoliviaMeldSet}, tc.cards), tc.code)
	}
	completed := &BoliviaMeld{Cards: make([]*Card, 7), Kind: BoliviaMeldSet, IsNatural: true}
	escalera := &BoliviaMeld{Cards: make([]*Card, 7), Kind: BoliviaMeldEscalera, IsNatural: true}
	for _, many := range []bool{false, true} {
		g, p = testBoliviaPathGame()
		p.SetMelds([]*BoliviaMeld{completed, completed, escalera})
		g.phase = BoliviaPhaseDiscard
		if many {
			p.AddCard(cloneCard(CardDesignHeart, 4))
			p.AddCard(cloneCard(CardDesignHeart, 5))
		} else {
			p.AddCard(cloneCard(CardDesignHeart, 3))
		}
		code := "bolivia.errRedThreeCannotBeDiscarded"
		if many {
			code = "bolivia.errHandMustHaveAtMostOneCardToGoOut"
		}
		requireCode(t, g.PlayerGoOut(), code)
	}
}

func TestSamba_ErrorPathLines(t *testing.T) {
	g, _ := testSambaPathGame()
	requireCode(t, g.validateNewSet([]*Card{cloneCard(CardDesignSpade, 7), cloneCard(CardDesignHeart, 8), cloneCard(CardDesignDiamond, 7)}), "samba.errSetMeldCardsMustHaveSameRank")
	for _, tc := range []struct {
		cards []*Card
		code  string
	}{{[]*Card{cloneCard(CardDesignSpade, 7), cloneCard(CardDesignHeart, 7), cloneCard(CardDesignJoker, 1), cloneCard(CardDesignJoker, 2), cloneCard(CardDesignJoker, 3), cloneCard(CardDesignJoker, 4)}, "samba.errMeldAllowsAtMostThreeWildCards"}, {[]*Card{cloneCard(CardDesignSpade, 7), cloneCard(CardDesignHeart, 7), cloneCard(CardDesignJoker, 1), cloneCard(CardDesignJoker, 2), cloneCard(CardDesignJoker, 3)}, "samba.errWildCardsCannotExceedNaturalCards"}} {
		requireCode(t, g.validateNewSet(tc.cards), tc.code)
	}
	for _, tc := range []struct {
		cards []*Card
		code  string
	}{{[]*Card{cloneCard(CardDesignSpade, 7), cloneCard(CardDesignHeart, 8)}, "samba.errMeldCardRankMismatch"}, {[]*Card{cloneCard(CardDesignClover, 3)}, "samba.errBlackThreeCannotMeld"}, {[]*Card{cloneCard(CardDesignJoker, 1)}, "samba.errMeldAllowsAtMostThreeWildCards"}} {
		requireCode(t, g.validateSetAddition(&SambaMeld{Cards: []*Card{cloneCard(CardDesignSpade, 7), cloneCard(CardDesignHeart, 7), cloneCard(CardDesignJoker, 1), cloneCard(CardDesignJoker, 2), cloneCard(CardDesignJoker, 3)}, Kind: SambaMeldSet}, tc.cards), tc.code)
	}
}

func TestHandAndFoot_ErrorPathLines(t *testing.T) {
	g, p := testHAFPathGame()
	card7 := cloneCard(CardDesignSpade, 7)
	g.SetDiscardPile([]*Card{card7})
	p.AddCard(cloneCard(CardDesignJoker, 1))
	p.AddCard(cloneCard(CardDesignHeart, 7))
	requireCode(t, g.PlayerDrawFromDiscard([]int{0, 1}), "handandfoot.errPairMustBeNatural")
	g, p = testHAFPathGame()
	g.SetDiscardPile([]*Card{card7})
	p.AddCard(cloneCard(CardDesignHeart, 7))
	p.AddCard(cloneCard(CardDesignDiamond, 7))
	require.NoError(t, g.PlayerDrawFromDiscard([]int{0, 1}))
	requireCode(t, g.PlayerMeld(nil), "handandfoot.errTopCardMustBeMelded")
	g, p = testHAFPathGame()
	g.phase = HandAndFootPhaseMeld
	g.drewFromDiscard = true
	g.drawnCard = card7
	for i := 0; i < 3; i++ {
		p.AddCard(cloneCard(CardDesignSpade+i%3, 7))
	}
	requireCode(t, g.PlayerMeld([][]int{{0, 1, 2}}), "handandfoot.errTopCardMustBeMelded")
	for _, tc := range []struct {
		cards []*Card
		code  string
	}{{[]*Card{cloneCard(CardDesignClover, 3), cloneCard(CardDesignSpade, 3), cloneCard(CardDesignHeart, 3)}, "handandfoot.errBlackThreeCannotMeld"}, {[]*Card{cloneCard(CardDesignSpade, 7), cloneCard(CardDesignHeart, 8), cloneCard(CardDesignDiamond, 7)}, "handandfoot.errMeldCardsMustHaveSameRank"}, {[]*Card{cloneCard(CardDesignSpade, 7), cloneCard(CardDesignHeart, 7), cloneCard(CardDesignJoker, 1), cloneCard(CardDesignJoker, 2), cloneCard(CardDesignJoker, 3), cloneCard(CardDesignJoker, 4)}, "handandfoot.errMeldAllowsAtMostThreeWildCards"}, {[]*Card{cloneCard(CardDesignSpade, 7), cloneCard(CardDesignHeart, 7), cloneCard(CardDesignJoker, 1), cloneCard(CardDesignJoker, 2), cloneCard(CardDesignJoker, 3)}, "handandfoot.errWildCardsCannotExceedNaturalCards"}} {
		requireCode(t, g.validateNewMeld(tc.cards), tc.code)
	}
	requireCode(t, g.validateMeldAddition(&CanastaMeld{Cards: []*Card{cloneCard(CardDesignSpade, 7), cloneCard(CardDesignHeart, 7)}}, []*Card{cloneCard(CardDesignSpade, 8)}), "handandfoot.errMeldCardRankMismatch")
	requireCode(t, g.validateMeldAddition(&CanastaMeld{Cards: []*Card{cloneCard(CardDesignSpade, 3), cloneCard(CardDesignHeart, 3)}}, []*Card{cloneCard(CardDesignClover, 3)}), "handandfoot.errBlackThreeCannotMeld")
	requireCode(t, g.validateMeldAddition(&CanastaMeld{Cards: []*Card{cloneCard(CardDesignSpade, 7), cloneCard(CardDesignHeart, 7), cloneCard(CardDesignJoker, 1), cloneCard(CardDesignJoker, 2), cloneCard(CardDesignJoker, 3)}}, []*Card{cloneCard(CardDesignJoker, 4)}), "handandfoot.errMeldAllowsAtMostThreeWildCards")
	requireCode(t, g.validateMeldAddition(&CanastaMeld{Cards: []*Card{cloneCard(CardDesignSpade, 7), cloneCard(CardDesignHeart, 7), cloneCard(CardDesignJoker, 1), cloneCard(CardDesignJoker, 2)}}, []*Card{cloneCard(CardDesignJoker, 3)}), "handandfoot.errWildCardsCannotExceedNaturalCards")
	cfg := g.GetConfig()
	cfg.RedCanastasToGoOut, cfg.BlackCanastasToGoOut = 0, 0
	g.SetConfig(cfg)
	p.Reset()
	p.SetInFoot(true)
	g.SetTeamMelds(0, []*CanastaMeld{{Cards: make([]*Card, 7), IsNatural: true}, {Cards: []*Card{cloneCard(CardDesignSpade, 7), cloneCard(CardDesignHeart, 7), cloneCard(CardDesignJoker, 1), cloneCard(CardDesignJoker, 2), cloneCard(CardDesignJoker, 3), cloneCard(CardDesignJoker, 4), cloneCard(CardDesignJoker, 5)}, IsNatural: false}})
	g.phase = HandAndFootPhaseDiscard
	p.AddCard(cloneCard(CardDesignHeart, 3))
	requireCode(t, g.PlayerGoOut(), "handandfoot.errRedThreeCannotBeDiscarded")
	g, p = testHAFPathGame()
	cfg = g.GetConfig()
	cfg.RedCanastasToGoOut, cfg.BlackCanastasToGoOut = 0, 0
	g.SetConfig(cfg)
	p.SetInFoot(true)
	g.SetTeamMelds(0, []*CanastaMeld{{Cards: make([]*Card, 7), IsNatural: true}, {Cards: []*Card{cloneCard(CardDesignSpade, 7), cloneCard(CardDesignHeart, 7), cloneCard(CardDesignJoker, 1), cloneCard(CardDesignJoker, 2), cloneCard(CardDesignJoker, 3), cloneCard(CardDesignJoker, 4), cloneCard(CardDesignJoker, 5)}, IsNatural: false}})
	g.phase = HandAndFootPhaseDiscard
	p.AddCard(cloneCard(CardDesignHeart, 4))
	p.AddCard(cloneCard(CardDesignHeart, 5))
	requireCode(t, g.PlayerGoOut(), "handandfoot.errHandMustHaveAtMostOneCardToGoOut")
}
