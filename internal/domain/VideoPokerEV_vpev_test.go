//go:build test && vpev

package domain

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"testing"
)

// vpFastMultiplier computes the payout from rank/suit counts. Wild ranks are
// evaluated by category feasibility rather than by constructing substitutions.
func vpFastMultiplier(cfg *VideoPokerVariantConfig, hand []*Card) int {
	if len(hand) != 5 {
		return 0
	}
	var counts [14]int
	var suits [5]int
	wild, nonwild := 0, 0
	deuces := 0
	suit, sameSuit := 0, true
	for _, c := range hand {
		if c.GetValue() == 2 {
			deuces++
		}
		if cfg.IsWild != nil && cfg.IsWild(c) {
			wild++
			continue
		}
		nonwild++
		v := c.GetValue()
		if v >= 1 && v <= 13 {
			counts[v]++
		}
		s := c.GetDesign()
		if suit == 0 {
			suit = s
		} else if suit != s {
			sameSuit = false
		}
		if s >= 0 && s < len(suits) {
			suits[s]++
		}
	}
	if cfg.Name == "deuceswild" && deuces == 4 {
		return 200
	}
	straight, straightFlush, royalFlush := false, false, false
	for start := 1; start <= 10; start++ {
		need := 0
		for v := start; v < start+5; v++ {
			x := v
			if x == 14 {
				x = 1
			}
			if counts[x] == 0 {
				need++
			}
		}
		if need <= wild {
			straight = true
			if start == 10 {
			}
			if sameSuit {
				straightFlush = true
				if start == 10 {
					royalFlush = true
				}
			}
		}
	}
	// Ace-low straight.
	need := 0
	for _, v := range []int{1, 2, 3, 4, 5} {
		if counts[v] == 0 {
			need++
		}
	}
	if need <= wild {
		straight = true
		if sameSuit {
			straightFlush = true
		}
	}
	flush := sameSuit || nonwild == 0
	maxc, pairs, trips, quads := 0, 0, 0, 0
	for v := 1; v <= 13; v++ {
		n := counts[v]
		if n > maxc {
			maxc = n
		}
		if n == 2 {
			pairs++
		}
		if n == 3 {
			trips++
		}
		if n == 4 {
			quads++
		}
	}
	five := maxc+wild >= 5
	four := maxc+wild >= 4
	full := false
	if wild == 0 {
		full = trips > 0 && pairs > 0
	} else {
		for a := 1; a <= 13 && !full; a++ {
			for b := 1; b <= 13; b++ {
				if a == b {
					continue
				}
				if max(0, 3-counts[a])+max(0, 2-counts[b]) <= wild {
					full = true
					break
				}
			}
		}
	}
	three := maxc+wild >= 3
	wildRoyal := royalFlush && wild > 0
	if cfg.Name == "jacksorbetter" {
		rank := evalFiveCardHand(hand)
		if rank == PokerHandRoyalFlush {
			return 800
		}
		switch rank {
		case PokerHandStraightFlush:
			return 50
		case PokerHandFourOfAKind:
			return 25
		case PokerHandFullHouse:
			return 9
		case PokerHandFlush:
			return 6
		case PokerHandStraight:
			return 4
		case PokerHandThreeOfAKind:
			return 3
		case PokerHandTwoPair:
			return 2
		case PokerHandOnePair:
			if isJacksOrBetterHand(hand) {
				return 1
			}
		}
		return 0
	}
	if cfg.Name == "deuceswild" {
		if royalFlush && wild == 0 {
			return 800
		}
		if wildRoyal {
			return 25
		}
		if five {
			return 15
		}
		if straightFlush {
			return 9
		}
		if four {
			return 5
		}
		if full {
			return 3
		}
		if flush {
			return 2
		}
		if straight {
			return 2
		}
		if three {
			return 1
		}
		return 0
	}
	if royalFlush && wild == 0 {
		return 800
	}
	if wildRoyal {
		return 100
	}
	if five {
		return 200
	}
	if straightFlush {
		return 50
	}
	if four {
		return 20
	}
	if full {
		return 7
	}
	if flush {
		return 5
	}
	if straight {
		return 3
	}
	if three {
		return 2
	}
	if pairs >= 2 {
		return 1
	}
	for v := 1; v <= 13; v++ {
		if (counts[v] >= 2 || wild > 0 && counts[v] >= 1) && (v == 1 || v == 13) {
			return 1
		}
	}
	return 0
}

func vpDeck(cfg *VideoPokerVariantConfig) []*Card {
	d := make([]*Card, 0, 52+cfg.JokerCount)
	for s := CardDesignSpade; s <= CardDesignDiamond; s++ {
		for v := 1; v <= 13; v++ {
			d = append(d, NewCard(s, v, false))
		}
	}
	for i := 1; i <= cfg.JokerCount; i++ {
		d = append(d, NewCard(CardDesignJoker, i, false))
	}
	return d
}

func vpHandString(hand []*Card) string {
	labels := make([]string, len(hand))
	for i, c := range hand {
		suit := map[int]string{CardDesignSpade: "S", CardDesignClover: "C", CardDesignHeart: "H", CardDesignDiamond: "D", CardDesignJoker: "J"}[c.GetDesign()]
		labels[i] = fmt.Sprintf("%s%d", suit, c.GetValue())
	}
	return fmt.Sprint(labels)
}

func vpMaskCards(hand []*Card, mask uint8) string {
	kept := make([]*Card, 0, 5)
	for i, c := range hand {
		if mask&(1<<i) != 0 {
			kept = append(kept, c)
		}
	}
	return vpHandString(kept)
}

func vpExactHoldEVs(cfg *VideoPokerVariantConfig, hand []*Card) [32]float64 {
	var out [32]float64
	deck := vpDeck(cfg)
	used := make(map[*Card]bool)
	for _, c := range hand {
		for _, d := range deck {
			if d.GetDesign() == c.GetDesign() && d.GetValue() == c.GetValue() {
				used[d] = true
				break
			}
		}
	}
	remaining := make([]*Card, 0, len(deck)-5)
	for _, d := range deck {
		if !used[d] {
			remaining = append(remaining, d)
		}
	}
	for mask := 0; mask < 32; mask++ {
		base := make([]*Card, 0, 5)
		for i, c := range hand {
			if mask&(1<<i) != 0 {
				base = append(base, c)
			}
		}
		need := 5 - len(base)
		if need == 0 {
			out[mask] = float64(vpFastMultiplier(cfg, base))
			continue
		}
		total := 0
		n := 0
		var walk func(int, int)
		walk = func(at, left int) {
			if left == 0 {
				total += vpFastMultiplier(cfg, base)
				n++
				return
			}
			for i := at; i <= len(remaining)-left; i++ {
				base = append(base, remaining[i])
				walk(i+1, left-1)
				base = base[:len(base)-1]
			}
		}
		walk(0, need)
		if n > 0 {
			out[mask] = float64(total) / float64(n)
		}
	}
	return out
}
func vpOptimalHold(cfg *VideoPokerVariantConfig, hand []*Card) (uint8, float64) {
	evs := vpExactHoldEVs(cfg, hand)
	best := 0
	for i := 1; i < 32; i++ {
		if evs[i] > evs[best] {
			best = i
		}
	}
	return uint8(best), evs[best]
}

func TestVPEVFastEvaluatorMatchesDomain(t *testing.T) {
	card := func(s, v int) *Card { return NewCard(s, v, false) }
	boundary := [][]*Card{
		{card(1, 1), card(1, 13), card(1, 12), card(1, 11), card(1, 10)},
		{card(1, 9), card(1, 10), card(1, 11), card(1, 12), card(1, 13)},
		{card(1, 2), card(1, 3), card(1, 4), card(1, 5), card(1, 6)},
		{card(1, 1), card(2, 1), card(3, 1), card(4, 1), card(1, 9)},
		{card(1, 13), card(2, 13), card(3, 13), card(4, 13), card(1, 2)},
		{card(1, 3), card(2, 3), card(3, 3), card(1, 8), card(2, 8)},
		{card(1, 5), card(2, 5), card(3, 5), card(4, 5), card(1, 8)},
		{card(1, 7), card(2, 7), card(1, 11), card(2, 11), card(3, 2)},
		{card(1, 11), card(2, 11), card(3, 4), card(4, 8), card(1, 2)},
		{card(1, 1), card(1, 3), card(1, 5), card(1, 7), card(1, 9)},
		{card(1, 2), card(2, 2), card(3, 2), card(4, 2), card(1, 9)},
		{card(1, 1), card(1, 13), card(1, 12), card(1, 11), card(0, 1)},
		{card(2, 1), card(2, 13), card(2, 12), card(2, 11), card(0, 1)},
		{card(1, 13), card(2, 13), card(3, 13), card(4, 13), card(0, 1)},
		{card(1, 12), card(2, 12), card(3, 12), card(1, 8), card(0, 1)},
		{card(1, 10), card(1, 11), card(1, 12), card(1, 13), card(0, 1)},
		{card(1, 1), card(2, 2), card(3, 3), card(4, 4), card(0, 1)},
		{card(1, 2), card(2, 2), card(3, 2), card(1, 9), card(0, 1)},
		{card(1, 13), card(2, 13), card(1, 8), card(2, 8), card(0, 1)},
		{card(1, 2), card(2, 2), card(3, 2), card(4, 2), card(0, 1)},
	}
	for _, cfg := range []*VideoPokerVariantConfig{JacksOrBetterConfig(), DeucesWildConfig(), JokerPokerConfig()} {
		for _, h := range boundary {
			hasJoker := false
			for _, c := range h {
				if c.GetDesign() == CardDesignJoker {
					hasJoker = true
				}
			}
			if hasJoker && cfg.Name != "jokerpoker" {
				continue
			}
			_, want, _ := cfg.GetResult(h, VideoPokerMaxBet)
			got := vpFastMultiplier(cfg, h)
			if got != want {
				t.Fatalf("%s boundary %s got %d want %d", cfg.Name, vpHandString(h), got, want)
			}
		}
		r := rand.New(rand.NewSource(10969))
		deck := vpDeck(cfg)
		for n := 0; n < 300000; n++ {
			r.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
			h := deck[:5]
			_, want, _ := cfg.GetResult(h, VideoPokerMaxBet)
			got := vpFastMultiplier(cfg, h)
			if got != want {
				t.Fatalf("%s %s got %d want %d", cfg.Name, vpHandString(h), got, want)
			}
		}
	}
}

func TestVPEVKnownValues(t *testing.T) {
	cfg := JacksOrBetterConfig()
	h := []*Card{NewCard(1, 1, false), NewCard(1, 13, false), NewCard(1, 12, false), NewCard(1, 11, false), NewCard(3, 2, false)}
	ev := vpExactHoldEVs(cfg, h)
	if d := ev[15] - 872.0/47; d > 1e-9 || d < -1e-9 {
		t.Fatalf("royal draw EV %.12f", ev[15])
	}
	royal := []*Card{NewCard(1, 1, false), NewCard(1, 13, false), NewCard(1, 12, false), NewCard(1, 11, false), NewCard(1, 10, false)}
	if got := vpExactHoldEVs(cfg, royal)[31]; got != 800 {
		t.Fatalf("made royal EV %v", got)
	}
}

type vpLoss struct {
	hand, selectedCards, optimalCards string
	selected, optimal                 uint8
	selectedEV, optimalEV             float64
}

func vpMeasureStrategy(t *testing.T, cfg *VideoPokerVariantConfig, name string, strategy func([]*Card) uint8, deals int, seed int64) {
	t.Helper()
	r := rand.New(rand.NewSource(seed))
	deck := vpDeck(cfg)
	hands := make([][]*Card, deals)
	for i := range hands {
		r.Shuffle(len(deck), func(a, b int) { deck[a], deck[b] = deck[b], deck[a] })
		hands[i] = append([]*Card(nil), deck[:5]...)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	losses := []vpLoss{}
	var bestSum, stratSum float64
	wrong := 0
	next := make(chan int)
	workers := runtime.NumCPU()
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range next {
				h := hands[i]
				evs := vpExactHoldEVs(cfg, h)
				best := 0
				for m := 1; m < 32; m++ {
					if evs[m] > evs[best] {
						best = m
					}
				}
				chosen := strategy(h)
				l := vpLoss{vpHandString(h), vpMaskCards(h, chosen), vpMaskCards(h, uint8(best)), uint8(chosen), uint8(best), evs[chosen], evs[best]}
				mu.Lock()
				bestSum += evs[best]
				stratSum += evs[chosen]
				if int(chosen) != best {
					wrong++
				}
				if evs[best] > evs[chosen] {
					losses = append(losses, l)
				}
				mu.Unlock()
			}
		}()
	}
	for i := range hands {
		next <- i
	}
	close(next)
	wg.Wait()
	sort.Slice(losses, func(i, j int) bool {
		return losses[i].optimalEV-losses[i].selectedEV > losses[j].optimalEV-losses[j].selectedEV
	})
	if len(losses) > 20 {
		losses = losses[:20]
	}
	t.Logf("%s %s optimal %.6f strategy %.6f diff %.6f errors %d", cfg.Name, name, bestSum/float64(deals), stratSum/float64(deals), (bestSum-stratSum)/float64(deals), wrong)
	for _, l := range losses {
		t.Logf("%s selected=%s optimal=%s EV %.6f/%.6f", l.hand, l.selectedCards, l.optimalCards, l.selectedEV, l.optimalEV)
	}
}
func TestVPEVMeasureDiscardAll(t *testing.T) {
	deals := 200
	if s := os.Getenv("VPEV_DEALS"); s != "" {
		if n, e := strconv.Atoi(s); e == nil && n > 0 {
			deals = n
		}
	}
	for _, c := range []*VideoPokerVariantConfig{JacksOrBetterConfig(), DeucesWildConfig(), JokerPokerConfig()} {
		vpMeasureStrategy(t, c, "discard all", func([]*Card) uint8 { return 0 }, deals, 10969)
	}
}

// TestVPEVMeasureStrategyTable measures RecommendVideoPokerHold against the
// exact optimum. VPEV_VARIANT limits it to one variant name.
func TestVPEVMeasureStrategyTable(t *testing.T) {
	deals := 200
	if s := os.Getenv("VPEV_DEALS"); s != "" {
		if n, e := strconv.Atoi(s); e == nil && n > 0 {
			deals = n
		}
	}
	only := os.Getenv("VPEV_VARIANT")
	seed := int64(10969)
	if v, e := strconv.ParseInt(os.Getenv("VPEV_SEED"), 10, 64); e == nil {
		seed = v
	}
	for _, c := range []*VideoPokerVariantConfig{JacksOrBetterConfig(), DeucesWildConfig(), JokerPokerConfig()} {
		if only != "" && only != c.Name {
			continue
		}
		vpMeasureStrategy(t, c, "strategy table", func(hand []*Card) uint8 {
			advice := RecommendVideoPokerHold(c.Name, hand)
			var mask uint8
			for i, held := range advice.Hold {
				if held {
					mask |= 1 << i
				}
			}
			return mask
		}, deals, seed)
	}
}
