//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"testing"
)

func gdCard(suit, value int) *Card { return NewCard(suit, value, true) }

// gdFresh returns a dealt game with every hand cleared, so a test can place
// exactly the cards it cares about.
func gdFresh(t *testing.T, level int) *Guandan {
	t.Helper()
	g := NewDefaultGuandan()
	g.Reset()
	g.SetLevelForTest(level)
	for i := range GuandanPlayerCnt {
		g.SetHandForTest(i, nil)
	}
	g.SetPhaseForTest(GuandanPhasePlay)
	g.SetCurrentPlayerForTest(0)
	return g
}

// TestGuandanDeal covers the deck, which the issue gets right.
func TestGuandanDeal(t *testing.T) {
	deck := newGuandanDeck()
	if got := len(deck); got != GuandanDeckSize {
		t.Fatalf("the deck holds %d cards, want %d", got, GuandanDeckSize)
	}
	// **52 × 2 + ジョーカー 4。**
	jokers := 0
	for _, c := range deck {
		if GuandanIsJoker(c) {
			jokers++
		}
	}
	if jokers != 4 {
		t.Errorf("%d jokers, want 4", jokers)
	}
	if GuandanPlayerCnt*GuandanHandSize != GuandanDeckSize {
		t.Fatalf("%d seats x %d cards != %d", GuandanPlayerCnt, GuandanHandSize, GuandanDeckSize)
	}

	g := NewDefaultGuandan()
	g.Reset()
	for i := range GuandanPlayerCnt {
		if got := g.GetPlayer(i).GetCardsSize(); got != GuandanHandSize {
			t.Errorf("seat %d holds %d, want %d", i, got, GuandanHandSize)
		}
	}
	// 開始レベルは 2。
	if got := g.GetLevel(); got != GuandanMinLevel {
		t.Errorf("level = %d, want %d", got, GuandanMinLevel)
	}
}

// TestGuandanLevelCardsOutrankAces は issue が落としている序列を押さえる。
//
// **レベル札は A の上、黒ジョーカーの下**に割り込む。
func TestGuandanLevelCardsOutrankAces(t *testing.T) {
	const level = 5
	five := gdCard(CardDesignSpade, 5)
	ace := gdCard(CardDesignSpade, 1)
	blackJoker := gdCard(CardDesignJoker, 1)
	redJoker := gdCard(CardDesignJoker, 2)
	king := gdCard(CardDesignSpade, 13)

	if !GuandanIsLevelCard(five, level) {
		t.Fatal("the five is the level card at level 5")
	}
	// **A より強い。**
	if GuandanRank(five, level) <= GuandanRank(ace, level) {
		t.Error("a level card outranks an ace")
	}
	// **黒ジョーカーより弱い。**
	if GuandanRank(five, level) >= GuandanRank(blackJoker, level) {
		t.Error("a level card is below the black joker")
	}
	if GuandanRank(blackJoker, level) >= GuandanRank(redJoker, level) {
		t.Error("the red joker is the highest card")
	}
	// K は本来どおり A の下。
	if GuandanRank(king, level) >= GuandanRank(ace, level) {
		t.Error("the king stays below the ace")
	}

	// **序列は局ごとに変わる。**レベルが動けば同じ札の強さが変わる。
	if GuandanRank(five, 6) >= GuandanRank(ace, 6) {
		t.Error("at level 6 the five is an ordinary low card again")
	}
	if GuandanRank(gdCard(CardDesignSpade, 6), 6) <= GuandanRank(ace, 6) {
		t.Error("at level 6 the six becomes the level card")
	}
	if GuandanRank(nil, level) != 0 {
		t.Error("a nil card has no rank")
	}
	if GuandanIsLevelCard(nil, level) || GuandanIsLevelCard(blackJoker, level) {
		t.Error("neither nil nor a joker is a level card")
	}
	if GuandanIsJoker(nil) {
		t.Error("nil is not a joker")
	}
}

// TestGuandanOnlyHeartLevelCardsAreWild は issue の
// 「レベル札と同スートの該当ランクがワイルド」が不正確であることを押さえる。
//
// **ワイルドは ♥ のレベル札だけ。**2 デッキなのでちょうど 2 枚。
func TestGuandanOnlyHeartLevelCardsAreWild(t *testing.T) {
	const level = 5
	if !GuandanIsWild(gdCard(CardDesignHeart, 5), level) {
		t.Error("the heart level card is wild")
	}
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignDiamond} {
		if GuandanIsWild(gdCard(suit, 5), level) {
			t.Errorf("the level card in suit %d is NOT wild — only hearts are", suit)
		}
	}
	// レベルでない ♥ はワイルドではない。
	if GuandanIsWild(gdCard(CardDesignHeart, 6), level) {
		t.Error("a heart that is not the level card is not wild")
	}
	// ジョーカーはワイルドではない (逆にワイルドはジョーカーの代用にならない)。
	if GuandanIsWild(gdCard(CardDesignJoker, 2), level) {
		t.Error("a joker is not a wild card")
	}

	// **卓に 2 枚だけ。**
	wilds := 0
	for _, c := range newGuandanDeck() {
		if GuandanIsWild(c, level) {
			wilds++
		}
	}
	if wilds != 2 {
		t.Errorf("%d wild cards in the pack, want exactly 2", wilds)
	}
}

// **ワイルドは任意の札になるが、ジョーカーの代用にはならない。**
func TestGuandanWildFormsCombinationsButNotJokerBombs(t *testing.T) {
	const level = 5
	wild := gdCard(CardDesignHeart, 5)

	// ワイルド + K 1 枚でペアになる。
	c := GuandanEvaluate([]*Card{gdCard(CardDesignSpade, 13), wild}, level)
	if c == nil || c.Kind != GuandanComboPair {
		t.Fatalf("combo = %v, want a pair — the wild stands in for the second king", c)
	}
	if c.Rank != GuandanRank(gdCard(CardDesignSpade, 13), level) {
		t.Errorf("the pair ranks %d, want the king's rank", c.Rank)
	}

	// **ジョーカー 4 枚はワイルドでは作れない。**
	real4 := []*Card{
		gdCard(CardDesignJoker, 1), gdCard(CardDesignJoker, 1),
		gdCard(CardDesignJoker, 2), gdCard(CardDesignJoker, 2),
	}
	if got := GuandanEvaluate(real4, level); got == nil || got.Kind != GuandanComboJokerBomb {
		t.Fatalf("combo = %v, want the joker bomb", got)
	}
	faked := []*Card{
		gdCard(CardDesignJoker, 1), gdCard(CardDesignJoker, 2),
		gdCard(CardDesignJoker, 2), wild,
	}
	if got := GuandanEvaluate(faked, level); got != nil && got.Kind == GuandanComboJokerBomb {
		t.Error("a wild may not stand in for a joker")
	}
}

// TestGuandanAdvanceIsOneTwoFour は issue の「1〜3ランク上昇」が誤りである
// ことを押さえる。
//
// **上位独占は +4。**「最大 3」では上位独占を狙う動機が消える。
func TestGuandanAdvanceIsOneTwoFour(t *testing.T) {
	if GuandanAdvanceFirstSecond != 4 {
		t.Fatalf("a 1-2 win advances %d, want 4 — the issue's max of 3 cannot express it",
			GuandanAdvanceFirstSecond)
	}

	for _, tc := range []struct {
		name  string
		order []int
		want  int
	}{
		// 席 0 と 2 が同じチーム。
		{"1-2 (the partners take the top two)", []int{0, 2, 1, 3}, GuandanAdvanceFirstSecond},
		{"1-3", []int{0, 1, 2, 3}, GuandanAdvanceFirstThird},
		{"1-4", []int{0, 1, 3, 2}, GuandanAdvanceFirstFourth},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := gdFresh(t, GuandanMinLevel)
			g.SetFinishedForTest(tc.order)
			g.FinishHandForTest()

			res := g.GetLastResult()
			if res == nil {
				t.Fatal("the hand must settle")
			}
			if res.Advance != tc.want {
				t.Errorf("advance = %d, want %d", res.Advance, tc.want)
			}
			if res.WinnerTeam != GuandanTeamOf(tc.order[0]) {
				t.Errorf("winner = %d, want the first-place team", res.WinnerTeam)
			}
			if got := g.GetTeamLevel(res.WinnerTeam); got != GuandanMinLevel+tc.want {
				t.Errorf("the team is at level %d, want %d", got, GuandanMinLevel+tc.want)
			}
		})
	}
}

// TestGuandanWinAtAceNeedsFirstSecondOrFirstThird は issue の
// 「上位2着を独占すれば勝利」が不足であることを押さえる。
//
// **1着-3着でも勝てる。**
func TestGuandanWinAtAceNeedsFirstSecondOrFirstThird(t *testing.T) {
	for _, tc := range []struct {
		name  string
		order []int
		win   bool
	}{
		{"1-2 at level A wins", []int{0, 2, 1, 3}, true},
		{"1-3 at level A also wins", []int{0, 1, 2, 3}, true},
		{"1-4 at level A does not", []int{0, 1, 3, 2}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := gdFresh(t, GuandanMinLevel)
			g.SetTeamLevelForTest(0, GuandanMaxLevel)
			g.SetFinishedForTest(tc.order)
			g.FinishHandForTest()

			if got := g.GetGameEndFlag(); got != tc.win {
				t.Fatalf("game end = %v, want %v", got, tc.win)
			}
			if tc.win {
				if got := g.GetWinnerTeam(); got != 0 {
					t.Errorf("winner = %d, want 0", got)
				}
				if got := g.GetPhase(); got != GuandanPhaseGameEnd {
					t.Errorf("phase = %v, want game end", got)
				}
			}
		})
	}

	// レベルが A に届いていなければ、上位独占でも終わらない。
	g := gdFresh(t, GuandanMinLevel)
	g.SetTeamLevelForTest(0, GuandanMaxLevel-1)
	g.SetFinishedForTest([]int{0, 2, 1, 3})
	g.FinishHandForTest()
	if g.GetGameEndFlag() {
		t.Error("a 1-2 below level A does not win the game")
	}
	// **レベルは A で頭打ち。**
	if got := g.GetTeamLevel(0); got != GuandanMaxLevel {
		t.Errorf("level = %d, want it capped at A (%d)", got, GuandanMaxLevel)
	}
}

// TestGuandanTribute covers the system the issue omits entirely.
func TestGuandanTribute(t *testing.T) {
	// **上位独占の次は敗者 2 人が払う。**
	t.Run("after a 1-2 win both losers pay", func(t *testing.T) {
		g := gdFresh(t, GuandanMinLevel)
		// 席 1 と 3 が敗者。ワイルドを除く最強は ♠A。
		g.SetHandForTest(1, []*Card{gdCard(CardDesignSpade, 1), gdCard(CardDesignSpade, 3)})
		g.SetHandForTest(3, []*Card{gdCard(CardDesignClover, 13), gdCard(CardDesignClover, 4)})
		g.SetHandForTest(0, []*Card{gdCard(CardDesignHeart, 7)})
		g.SetHandForTest(2, []*Card{gdCard(CardDesignHeart, 8)})

		prev := &GuandanHandResult{Order: [GuandanPlayerCnt]int{0, 2, 1, 3}, FirstSecond: true}
		g.PrepareTributeForTest(prev)

		if got := len(g.GetTributes()); got != 2 {
			t.Fatalf("%d tributes, want both losers to pay", got)
		}
		// 払ったのは最強の 1 枚。
		if c := g.GetTributes()[0].Card; c == nil || c.GetValue() != 1 {
			t.Errorf("seat 2 paid %v, want its ace", c)
		}
	})

	// **それ以外は最下位だけが払う。**
	t.Run("after a 1-3 or 1-4 win only the last player pays", func(t *testing.T) {
		g := gdFresh(t, GuandanMinLevel)
		g.SetHandForTest(3, []*Card{gdCard(CardDesignSpade, 1), gdCard(CardDesignSpade, 3)})
		g.SetHandForTest(0, []*Card{gdCard(CardDesignHeart, 7)})

		prev := &GuandanHandResult{Order: [GuandanPlayerCnt]int{0, 1, 2, 3}, FirstSecond: false}
		g.PrepareTributeForTest(prev)

		if got := len(g.GetTributes()); got != 1 {
			t.Fatalf("%d tributes, want only the last player to pay", got)
		}
		if got := g.GetTributes()[0].From; got != 3 {
			t.Errorf("seat %d paid, want the last-place seat", got)
		}
	})

	// **貢を払える札が 1 枚も無ければ、還貢のフェーズには入らない。**
	// 入ってしまうと、返す相手のいない席が延々と還貢を試みて局が止まる。
	t.Run("a payer with nothing but wilds skips the tribute phase", func(t *testing.T) {
		const level = 5
		g := gdFresh(t, level)
		// ♥5 はワイルドなので貢に出せない。この席は払える札を持たない。
		g.SetHandForTest(3, []*Card{gdCard(CardDesignHeart, level), gdCard(CardDesignHeart, level)})
		g.SetHandForTest(0, []*Card{gdCard(CardDesignHeart, 7)})
		g.SetPhaseForTest(GuandanPhaseTribute)

		prev := &GuandanHandResult{Order: [GuandanPlayerCnt]int{0, 1, 2, 3}, FirstSecond: false}
		g.PrepareTributeForTest(prev)

		if got := len(g.GetTributes()); got != 0 {
			t.Fatalf("%d tributes, want none", got)
		}
		if got := g.GetPhase(); got != GuandanPhasePlay {
			t.Errorf("phase %v, want play -- nobody owes a return", got)
		}
	})

	// **払うのはワイルドを除く最強。**ワイルドは手元に残る。
	t.Run("the wild card is never paid as tribute", func(t *testing.T) {
		const level = 5
		g := gdFresh(t, level)
		// ♥5 はワイルド。序列上は A より強いが、貢には出せない。
		g.SetHandForTest(3, []*Card{gdCard(CardDesignHeart, 5), gdCard(CardDesignSpade, 1)})
		g.SetHandForTest(0, []*Card{gdCard(CardDesignHeart, 7)})

		prev := &GuandanHandResult{Order: [GuandanPlayerCnt]int{0, 1, 2, 3}, FirstSecond: false}
		g.PrepareTributeForTest(prev)

		if got := len(g.GetTributes()); got != 1 {
			t.Fatalf("%d tributes, want 1", got)
		}
		paid := g.GetTributes()[0].Card
		if GuandanIsWild(paid, level) {
			t.Fatal("the wild card must never be paid as tribute")
		}
		if paid.GetValue() != 1 {
			t.Errorf("paid %v, want the ace — the highest NON-WILD card", paid)
		}
	})

	// **赤ジョーカーで取り消される。**
	t.Run("red jokers cancel the tribute", func(t *testing.T) {
		// 上位独占の次: 敗者 2 人が 1 枚ずつ。
		g := gdFresh(t, GuandanMinLevel)
		g.SetHandForTest(1, []*Card{gdCard(CardDesignJoker, 2), gdCard(CardDesignSpade, 3)})
		g.SetHandForTest(3, []*Card{gdCard(CardDesignJoker, 2), gdCard(CardDesignSpade, 4)})
		prev := &GuandanHandResult{Order: [GuandanPlayerCnt]int{0, 2, 1, 3}, FirstSecond: true}
		if !g.GuandanTributeCancelled(prev) {
			t.Error("one red joker each cancels the tribute after a 1-2 win")
		}

		// 片方が 2 枚でも取り消し。
		g2 := gdFresh(t, GuandanMinLevel)
		g2.SetHandForTest(1, []*Card{gdCard(CardDesignJoker, 2), gdCard(CardDesignJoker, 2)})
		g2.SetHandForTest(3, []*Card{gdCard(CardDesignSpade, 4)})
		if !g2.GuandanTributeCancelled(prev) {
			t.Error("two red jokers in one hand cancels it too")
		}

		// **それ以外は最下位が 2 枚持っているときだけ。**
		other := &GuandanHandResult{Order: [GuandanPlayerCnt]int{0, 1, 2, 3}, FirstSecond: false}
		g3 := gdFresh(t, GuandanMinLevel)
		g3.SetHandForTest(3, []*Card{gdCard(CardDesignJoker, 2), gdCard(CardDesignSpade, 4)})
		if g3.GuandanTributeCancelled(other) {
			t.Error("one red joker is not enough after a 1-3 or 1-4 win")
		}
		g3.SetHandForTest(3, []*Card{gdCard(CardDesignJoker, 2), gdCard(CardDesignJoker, 2)})
		if !g3.GuandanTributeCancelled(other) {
			t.Error("two red jokers in the last player's hand cancels it")
		}

		if g3.GuandanTributeCancelled(nil) {
			t.Error("there is nothing to cancel before the first hand")
		}
	})

	// 還貢で札が戻る。
	t.Run("the winner returns an unwanted card", func(t *testing.T) {
		g := gdFresh(t, GuandanMinLevel)
		g.SetHandForTest(3, []*Card{gdCard(CardDesignSpade, 1), gdCard(CardDesignSpade, 3)})
		g.SetHandForTest(0, []*Card{gdCard(CardDesignHeart, 7)})
		prev := &GuandanHandResult{Order: [GuandanPlayerCnt]int{0, 1, 2, 3}, FirstSecond: false}
		g.SetPhaseForTest(GuandanPhaseTribute)
		g.PrepareTributeForTest(prev)

		before := g.GetPlayer(3).GetCardsSize()
		if err := g.ReturnTribute(0, 0); err != nil {
			t.Fatalf("ReturnTribute: %v", err)
		}
		if got := g.GetPlayer(3).GetCardsSize(); got != before+1 {
			t.Errorf("the payer holds %d, want one card back", got)
		}
		if g.GetTributes()[0].Returned == nil {
			t.Error("the return must be recorded")
		}
		// 全部返し終えたらプレイへ。
		if got := g.GetPhase(); got != GuandanPhasePlay {
			t.Errorf("phase = %v, want play", got)
		}
	})

	// 貢を払う札が無い席は飛ばす (手札が空でも落ちない)。
	t.Run("a payer with nothing to give is skipped", func(t *testing.T) {
		g := gdFresh(t, GuandanMinLevel)
		g.SetHandForTest(3, nil)
		prev := &GuandanHandResult{Order: [GuandanPlayerCnt]int{0, 1, 2, 3}, FirstSecond: false}
		g.PrepareTributeForTest(prev)
		if got := len(g.GetTributes()); got != 0 {
			t.Errorf("%d tributes, want none — the payer had no card", got)
		}
	})

	// **取り消されたらそのままプレイへ。**貢のフェーズで止まらない。
	t.Run("a cancelled tribute goes straight to play", func(t *testing.T) {
		g := gdFresh(t, GuandanMinLevel)
		g.SetHandForTest(3, []*Card{gdCard(CardDesignJoker, 2), gdCard(CardDesignJoker, 2)})
		g.SetPhaseForTest(GuandanPhaseTribute)
		prev := &GuandanHandResult{Order: [GuandanPlayerCnt]int{0, 1, 2, 3}, FirstSecond: false}
		g.PrepareTributeForTest(prev)

		if !g.IsTributeCancelled() {
			t.Fatal("two red jokers cancels it")
		}
		if got := g.GetPhase(); got != GuandanPhasePlay {
			t.Errorf("phase = %v, want play", got)
		}
		if got := len(g.GetTributes()); got != 0 {
			t.Errorf("%d tributes, want none", got)
		}
	})

	t.Run("return guards", func(t *testing.T) {
		g := gdFresh(t, GuandanMinLevel)
		if err := g.ReturnTribute(0, 0); err == nil {
			t.Error("returning outside the tribute phase must be refused")
		}
		g.SetPhaseForTest(GuandanPhaseTribute)
		if err := g.ReturnTribute(0, 0); err == nil {
			t.Error("returning when you owe nothing must be refused")
		}
	})
}

// TestGuandanCombinations covers the combination vocabulary.
func TestGuandanCombinations(t *testing.T) {
	const level = 5
	for _, tc := range []struct {
		name  string
		cards []*Card
		want  GuandanComboKind
	}{
		{"single", []*Card{gdCard(CardDesignSpade, 9)}, GuandanComboSingle},
		{"pair", []*Card{gdCard(CardDesignSpade, 9), gdCard(CardDesignHeart, 9)}, GuandanComboPair},
		{"triple", []*Card{
			gdCard(CardDesignSpade, 9), gdCard(CardDesignHeart, 9), gdCard(CardDesignClover, 9),
		}, GuandanComboTriple},
		{"bomb of four", []*Card{
			gdCard(CardDesignSpade, 9), gdCard(CardDesignHeart, 9),
			gdCard(CardDesignClover, 9), gdCard(CardDesignDiamond, 9),
		}, GuandanComboBomb},
		{"full house", []*Card{
			gdCard(CardDesignSpade, 9), gdCard(CardDesignHeart, 9), gdCard(CardDesignClover, 9),
			gdCard(CardDesignSpade, 4), gdCard(CardDesignHeart, 4),
		}, GuandanComboFullHouse},
		{"straight", []*Card{
			gdCard(CardDesignSpade, 6), gdCard(CardDesignHeart, 7), gdCard(CardDesignClover, 8),
			gdCard(CardDesignDiamond, 9), gdCard(CardDesignSpade, 10),
		}, GuandanComboStraight},
		{"straight flush", []*Card{
			gdCard(CardDesignSpade, 6), gdCard(CardDesignSpade, 7), gdCard(CardDesignSpade, 8),
			gdCard(CardDesignSpade, 9), gdCard(CardDesignSpade, 10),
		}, GuandanComboStraightFlush},
		{"joker bomb", []*Card{
			gdCard(CardDesignJoker, 1), gdCard(CardDesignJoker, 1),
			gdCard(CardDesignJoker, 2), gdCard(CardDesignJoker, 2),
		}, GuandanComboJokerBomb},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := GuandanEvaluate(tc.cards, level)
			if got == nil || got.Kind != tc.want {
				t.Errorf("combo = %v, want %v", got, tc.want)
			}
		})
	}

	// **飛行機 (連続トリプル 2 組) と木板 (連続ペア 3 組)。**
	t.Run("plate: two consecutive triples", func(t *testing.T) {
		got := GuandanEvaluate([]*Card{
			gdCard(CardDesignSpade, 7), gdCard(CardDesignHeart, 7), gdCard(CardDesignClover, 7),
			gdCard(CardDesignSpade, 8), gdCard(CardDesignHeart, 8), gdCard(CardDesignClover, 8),
		}, level)
		if got == nil || got.Kind != GuandanComboPlate {
			t.Errorf("combo = %v, want a plate", got)
		}
	})
	t.Run("tube: three consecutive pairs", func(t *testing.T) {
		got := GuandanEvaluate([]*Card{
			gdCard(CardDesignSpade, 7), gdCard(CardDesignHeart, 7),
			gdCard(CardDesignSpade, 8), gdCard(CardDesignHeart, 8),
			gdCard(CardDesignSpade, 9), gdCard(CardDesignHeart, 9),
		}, level)
		if got == nil || got.Kind != GuandanComboTube {
			t.Errorf("combo = %v, want a tube", got)
		}
	})
	// 6 枚でも連続していなければ役にならない。
	t.Run("six unrelated cards are nothing", func(t *testing.T) {
		got := GuandanEvaluate([]*Card{
			gdCard(CardDesignSpade, 2), gdCard(CardDesignHeart, 4),
			gdCard(CardDesignSpade, 6), gdCard(CardDesignHeart, 8),
			gdCard(CardDesignSpade, 10), gdCard(CardDesignHeart, 12),
		}, level)
		if got != nil {
			t.Errorf("combo = %v, want nothing", got)
		}
	})
	// **A は上にも下にも使える。**
	t.Run("the ace runs both high and low", func(t *testing.T) {
		high := GuandanEvaluate([]*Card{
			gdCard(CardDesignSpade, 10), gdCard(CardDesignHeart, 11), gdCard(CardDesignClover, 12),
			gdCard(CardDesignDiamond, 13), gdCard(CardDesignSpade, 1),
		}, level)
		if high == nil || high.Kind != GuandanComboStraight {
			t.Errorf("10-J-Q-K-A = %v, want a straight", high)
		}
		low := GuandanEvaluate([]*Card{
			gdCard(CardDesignSpade, 1), gdCard(CardDesignHeart, 2), gdCard(CardDesignClover, 3),
			gdCard(CardDesignDiamond, 4), gdCard(CardDesignSpade, 5),
		}, level)
		if low == nil || low.Kind != GuandanComboStraight {
			t.Errorf("A-2-3-4-5 = %v, want a straight", low)
		}
	})
	// 同ランクが混ざった 5 枚は連続にならない。
	t.Run("a repeated rank is not a straight", func(t *testing.T) {
		got := GuandanEvaluate([]*Card{
			gdCard(CardDesignSpade, 6), gdCard(CardDesignHeart, 6), gdCard(CardDesignClover, 7),
			gdCard(CardDesignDiamond, 8), gdCard(CardDesignSpade, 9),
		}, level)
		if got != nil && got.Kind == GuandanComboStraight {
			t.Error("6-6-7-8-9 is not a straight")
		}
	})
	// ジョーカーは連続に混ぜられない。
	t.Run("a joker cannot join a straight", func(t *testing.T) {
		got := GuandanEvaluate([]*Card{
			gdCard(CardDesignSpade, 6), gdCard(CardDesignHeart, 7), gdCard(CardDesignClover, 8),
			gdCard(CardDesignDiamond, 9), gdCard(CardDesignJoker, 2),
		}, level)
		if got != nil && got.Kind == GuandanComboStraight {
			t.Error("a joker is not a straight card")
		}
	})

	if GuandanEvaluate(nil, level) != nil {
		t.Error("no cards is no combination")
	}
	if GuandanEvaluate([]*Card{nil}, level) != nil {
		t.Error("a nil card is no combination")
	}
	// 役にならない組み合わせは弾く。
	if got := GuandanEvaluate([]*Card{
		gdCard(CardDesignSpade, 2), gdCard(CardDesignHeart, 7), gdCard(CardDesignClover, 13),
	}, level); got != nil {
		t.Errorf("combo = %v, want nothing", got)
	}
}

// **爆弾は通常役をすべて上回る。**爆弾どうしは種類 → 枚数 → ランク。
// **ワイルドの読みは常に最強で確定する。**弱いほうで確定させると、本来勝てる
// 手が場に負ける。
func TestGuandanWildTakesTheStrongestReading(t *testing.T) {
	const level = 5

	t.Run("a straight extends upward, not downward", func(t *testing.T) {
		// 6-7-8-9 + ♥5(ワイルド) は 5-6-7-8-9 とも 6-7-8-9-10 とも読める。
		cards := []*Card{
			gdCard(CardDesignSpade, 6), gdCard(CardDesignSpade, 7),
			gdCard(CardDesignSpade, 8), gdCard(CardDesignSpade, 9),
			gdCard(CardDesignHeart, level),
		}
		c := GuandanEvaluate(cards, level)
		if c == nil || c.Kind != GuandanComboStraight {
			t.Fatalf("got %v, want a straight", c)
		}
		if c.Rank != 10 {
			t.Errorf("rank %d, want 10 -- the wild should extend the top", c.Rank)
		}
	})

	// **6 枚の連続役は読みが一意に決まる。**ワイルドが 2 枚しかないので、
	// 複数の開始位置が成り立つ形は必ずプレートとしても読めてしまう。
	t.Run("two pairs plus two wilds read as a plate", func(t *testing.T) {
		cards := []*Card{
			gdCard(CardDesignSpade, 8), gdCard(CardDesignClover, 8),
			gdCard(CardDesignSpade, 9), gdCard(CardDesignClover, 9),
			gdCard(CardDesignHeart, level), gdCard(CardDesignHeart, level),
		}
		c := GuandanEvaluate(cards, level)
		if c == nil || c.Kind != GuandanComboPlate {
			t.Fatalf("got %v, want a plate", c)
		}
		if c.Rank != 9 {
			t.Errorf("rank %d, want 9", c.Rank)
		}
	})

	// ワイルド 1 枚で作る連続ペアは開始位置が 1 つしかない。
	t.Run("a tube with one wild is unambiguous", func(t *testing.T) {
		cards := []*Card{
			gdCard(CardDesignSpade, 8), gdCard(CardDesignClover, 8),
			gdCard(CardDesignSpade, 9), gdCard(CardDesignClover, 9),
			gdCard(CardDesignSpade, 10), gdCard(CardDesignHeart, level),
		}
		c := GuandanEvaluate(cards, level)
		if c == nil || c.Kind != GuandanComboTube {
			t.Fatalf("got %v, want a tube", c)
		}
		if c.Rank != 10 {
			t.Errorf("rank %d, want 10", c.Rank)
		}
	})

	// 読みが 1 通りしかない場合まで持ち上げてはいけない。
	t.Run("a plain straight keeps its own top", func(t *testing.T) {
		cards := []*Card{
			gdCard(CardDesignSpade, 6), gdCard(CardDesignSpade, 7), gdCard(CardDesignSpade, 8),
			gdCard(CardDesignSpade, 9), gdCard(CardDesignSpade, 10),
		}
		c := GuandanEvaluate(cards, level)
		if c == nil || c.Kind != GuandanComboStraightFlush {
			t.Fatalf("got %v, want a straight flush", c)
		}
		if c.Rank != 10 {
			t.Errorf("rank %d, want 10", c.Rank)
		}
	})
}

func TestGuandanBombsBeatEverything(t *testing.T) {
	straight := &GuandanCombo{Kind: GuandanComboStraight, Rank: 14, Size: 5}
	bomb4 := &GuandanCombo{Kind: GuandanComboBomb, Rank: 3, Size: 4}
	bomb5 := &GuandanCombo{Kind: GuandanComboBomb, Rank: 3, Size: 5}
	sf := &GuandanCombo{Kind: GuandanComboStraightFlush, Rank: 7, Size: 5}
	jokerBomb := &GuandanCombo{Kind: GuandanComboJokerBomb, Rank: 100, Size: 4}

	if !GuandanBeats(bomb4, straight) {
		t.Error("a bomb beats a straight even with a lower rank")
	}
	if GuandanBeats(straight, bomb4) {
		t.Error("a straight never beats a bomb")
	}
	if !GuandanBeats(bomb5, bomb4) {
		t.Error("a bigger bomb wins")
	}
	if !GuandanBeats(sf, bomb5) {
		t.Error("a straight flush outranks a plain bomb")
	}
	if !GuandanBeats(jokerBomb, sf) {
		t.Error("the joker bomb is the highest")
	}
	// 通常役は同じ種類・同じ枚数でしか比べられない。
	pair := &GuandanCombo{Kind: GuandanComboPair, Rank: 9, Size: 2}
	triple := &GuandanCombo{Kind: GuandanComboTriple, Rank: 3, Size: 3}
	if GuandanBeats(triple, pair) {
		t.Error("a triple does not beat a pair — they are different combinations")
	}
	if !GuandanBeats(&GuandanCombo{Kind: GuandanComboPair, Rank: 10, Size: 2}, pair) {
		t.Error("a higher pair beats a lower pair")
	}
	// 場が空なら何でも出せる。
	if !GuandanBeats(pair, nil) {
		t.Error("anything leads an empty table")
	}
	if GuandanBeats(nil, pair) {
		t.Error("nothing beats something")
	}
}

// **レベル札の割り込みは単札の比較にも効く。**
func TestGuandanLevelCardBeatsAceAsASingle(t *testing.T) {
	const level = 5
	g := gdFresh(t, level)
	g.SetHandForTest(0, []*Card{gdCard(CardDesignSpade, 1)})
	g.SetHandForTest(1, []*Card{gdCard(CardDesignSpade, 5)})

	if err := g.PlayCards(0, []int{0}); err != nil {
		t.Fatalf("lead: %v", err)
	}
	// ♠5 はこの局のレベル札なので A に勝つ。
	if err := g.PlayCards(1, []int{0}); err != nil {
		t.Errorf("the level card must beat the ace: %v", err)
	}
}

func TestGuandanPlayGuards(t *testing.T) {
	g := gdFresh(t, GuandanMinLevel)
	g.SetHandForTest(0, []*Card{gdCard(CardDesignSpade, 9), gdCard(CardDesignHeart, 3)})

	if err := g.PlayCards(1, []int{0}); err == nil {
		t.Error("playing out of turn must be refused")
	}
	if err := g.PlayCards(0, nil); err == nil {
		t.Error("playing nothing must be refused")
	}
	if err := g.PlayCards(0, []int{99}); err == nil {
		t.Error("an out-of-range index must be refused")
	}
	// **同じ札を 2 回数えられない。**
	if err := g.PlayCards(0, []int{0, 0}); err == nil {
		t.Error("naming the same card twice must be refused")
	}
	if err := g.PlayCards(0, []int{0, 1}); err == nil {
		t.Error("a non-combination must be refused")
	}
	// リード時にパスはできない。
	if err := g.Pass(0); err == nil {
		t.Error("you must lead rather than pass")
	}
	g.SetPhaseForTest(GuandanPhaseHandEnd)
	if err := g.PlayCards(0, []int{0}); err == nil {
		t.Error("playing outside the play phase must be refused")
	}
	if err := g.Pass(0); err == nil {
		t.Error("passing outside the play phase must be refused")
	}
}

// 弱い役は前の役を越えられない。
func TestGuandanMustBeatTheLastPlay(t *testing.T) {
	g := gdFresh(t, GuandanMinLevel)
	g.SetHandForTest(0, []*Card{gdCard(CardDesignSpade, 10)})
	g.SetHandForTest(1, []*Card{gdCard(CardDesignSpade, 9), gdCard(CardDesignHeart, 11)})

	if err := g.PlayCards(0, []int{0}); err != nil {
		t.Fatalf("lead: %v", err)
	}
	if err := g.PlayCards(1, []int{0}); err == nil {
		t.Error("a lower single must be refused")
	}
	if err := g.PlayCards(1, []int{1}); err != nil {
		t.Errorf("a higher single must be accepted: %v", err)
	}
}

// **一周してリードした本人に戻ったら場が流れる。**
func TestGuandanTableClearsAfterAFullLoop(t *testing.T) {
	g := gdFresh(t, GuandanMinLevel)
	for i := range GuandanPlayerCnt {
		g.SetHandForTest(i, []*Card{gdCard(CardDesignSpade, 3+i), gdCard(CardDesignClover, 3+i)})
	}
	if err := g.PlayCards(0, []int{0}); err != nil {
		t.Fatalf("lead: %v", err)
	}
	for _, seat := range []int{1, 2, 3} {
		if err := g.Pass(seat); err != nil {
			t.Fatalf("Pass(%d): %v", seat, err)
		}
	}
	if g.GetLastCombo() != nil {
		t.Error("the table must clear once it comes back round to the leader")
	}
	if got := g.GetCurrentPlayerIdx(); got != 0 {
		t.Errorf("turn = %d, want the leader", got)
	}
}

// 3 人が上がった時点で局は決まる。
func TestGuandanHandEndsWhenThreeAreOut(t *testing.T) {
	g := gdFresh(t, GuandanMinLevel)
	g.SetFinishedForTest([]int{0, 1, 2})
	g.FinishHandForTest()

	res := g.GetLastResult()
	if res == nil {
		t.Fatal("the hand must settle")
	}
	// 残り 1 人が最下位に入る。
	if res.Order[3] != 3 {
		t.Errorf("last place = %d, want the seat that never went out", res.Order[3])
	}
	if got := g.GetPhase(); got != GuandanPhaseHandEnd {
		t.Errorf("phase = %v, want hand end", got)
	}
}

func TestGuandanNextHandGuards(t *testing.T) {
	g := gdFresh(t, GuandanMinLevel)
	if err := g.NextHand(); err == nil {
		t.Error("dealing again mid-hand must be refused")
	}

	g.SetFinishedForTest([]int{0, 1, 2, 3})
	g.FinishHandForTest()
	if err := g.NextHand(); err != nil {
		t.Fatalf("NextHand: %v", err)
	}
	// **次局は貢から始まる。**
	if got := g.GetPhase(); got != GuandanPhaseTribute && got != GuandanPhasePlay {
		t.Errorf("phase = %v, want tribute (or play if it was cancelled)", got)
	}
	for i := range GuandanPlayerCnt {
		if g.GetPlayer(i).GetCardsSize() == 0 {
			t.Errorf("seat %d was not dealt", i)
		}
	}
	// **基準レベルは 1 着チームのレベルになる。**
	if got := g.GetLevel(); got != g.GetTeamLevel(g.GetDeclarerTeam()) {
		t.Errorf("level = %d, want the declaring team's level", got)
	}
}

// **CPU だけで 1 局を回し切れること。**途中で止まると詰む。
func TestGuandanCpuDrivesAFullHand(t *testing.T) {
	for attempt := range 20 {
		players := make([]*GuandanPlayer, 0, GuandanPlayerCnt)
		for range GuandanPlayerCnt {
			players = append(players, NewGuandanPlayer(false))
		}
		g := NewGuandan(players, DefaultGuandanConfig())
		g.Reset()
		for step := 0; step < 5000; step++ {
			if g.GetPhase() == GuandanPhaseHandEnd || g.GetGameEndFlag() {
				break
			}
			g.CpuPlay()
		}
		if g.GetPhase() != GuandanPhaseHandEnd && !g.GetGameEndFlag() {
			t.Fatalf("attempt %d: the hand never finished (phase %v)", attempt, g.GetPhase())
		}
		res := g.GetLastResult()
		if res == nil {
			t.Fatalf("attempt %d: no settlement", attempt)
		}
		// **全席が順位に現れる。**
		seen := map[int]bool{}
		for _, seat := range res.Order {
			if seat < 0 || seat >= GuandanPlayerCnt || seen[seat] {
				t.Fatalf("attempt %d: bad finishing order %v", attempt, res.Order)
			}
			seen[seat] = true
		}
		// **上昇量は 1 / 2 / 4 のいずれか。**
		switch res.Advance {
		case GuandanAdvanceFirstFourth, GuandanAdvanceFirstThird, GuandanAdvanceFirstSecond:
		default:
			t.Fatalf("attempt %d: advance = %d, want 1, 2 or 4", attempt, res.Advance)
		}
	}
}

// **CPU は次局の貢も自分で処理する。**貢のフェーズで止まらない。
func TestGuandanCpuHandlesTheTributePhase(t *testing.T) {
	players := make([]*GuandanPlayer, 0, GuandanPlayerCnt)
	for range GuandanPlayerCnt {
		players = append(players, NewGuandanPlayer(false))
	}
	g := NewGuandan(players, DefaultGuandanConfig())
	g.Reset()
	for step := 0; step < 5000 && g.GetPhase() != GuandanPhaseHandEnd && !g.GetGameEndFlag(); step++ {
		g.CpuPlay()
	}
	if g.GetGameEndFlag() {
		t.Skip("the game ended in one hand")
	}
	if err := g.NextHand(); err != nil {
		t.Fatalf("NextHand: %v", err)
	}
	for step := 0; step < 200 && g.GetPhase() == GuandanPhaseTribute; step++ {
		g.CpuPlay()
	}
	if got := g.GetPhase(); got != GuandanPhasePlay {
		t.Errorf("phase = %v, want the tribute to resolve into play", got)
	}
	// 貢のあとも全員の手札が 27 枚のまま (交換なので枚数は動かない)。
	for i := range GuandanPlayerCnt {
		if got := g.GetPlayer(i).GetCardsSize(); got != GuandanHandSize {
			t.Errorf("seat %d holds %d after the tribute, want %d — it is an exchange", i, got, GuandanHandSize)
		}
	}
}

func TestGuandanCpuEdges(t *testing.T) {
	g := gdFresh(t, GuandanMinLevel)
	if got := g.GuandanCpuPlay(99); got != nil {
		t.Errorf("an unknown seat plays %v, want nothing", got)
	}
	g.SetHandForTest(0, nil)
	if got := g.GuandanCpuPlay(0); got != nil {
		t.Errorf("an empty hand plays %v, want nothing", got)
	}

	// **ワイルドは温存する。**リードでは非ワイルドを先に出す。
	const level = 5
	g2 := gdFresh(t, level)
	g2.SetHandForTest(0, []*Card{gdCard(CardDesignHeart, 5), gdCard(CardDesignSpade, 3)})
	idxs := g2.GuandanCpuPlay(0)
	if len(idxs) != 1 || GuandanIsWild(g2.GetPlayer(0).GetCard(idxs[0]), level) {
		t.Error("the CPU must keep the wild card back when leading")
	}

	over := NewDefaultGuandan()
	over.Reset()
	over.SetPhaseForTest(GuandanPhaseGameEnd)
	over.gameEndFlag = true
	if over.IsHumanTurn() {
		t.Error("a finished game is nobody's turn")
	}
	over.CpuPlay()

	settled := gdFresh(t, GuandanMinLevel)
	settled.SetPhaseForTest(GuandanPhaseHandEnd)
	if settled.IsHumanTurn() {
		t.Error("the settlement is nobody's turn")
	}
	settled.CpuPlay()
}

// **27 枚を添字で指定して役を組む**ので、配った時点で並んでいないと実用に耐えない。
func TestGuandanDealsASortedHand(t *testing.T) {
	g := NewDefaultGuandan()
	g.Reset()

	for seat := range GuandanPlayerCnt {
		p := g.GetPlayer(seat)
		if p.GetCardsSize() != GuandanHandSize {
			t.Fatalf("seat %d holds %d cards, want %d", seat, p.GetCardsSize(), GuandanHandSize)
		}
		for i := 1; i < p.GetCardsSize(); i++ {
			prev, cur := p.GetCard(i-1), p.GetCard(i)
			pr, cr := GuandanRank(prev, g.GetLevel()), GuandanRank(cur, g.GetLevel())
			if pr > cr {
				t.Fatalf("seat %d is out of order at %d: rank %d then %d", seat, i, pr, cr)
			}
			// 同ランクはスート順。ペアを添字で拾うにはここまで揃っている必要がある。
			if pr == cr && prev.GetDesign() > cur.GetDesign() {
				t.Fatalf("seat %d is out of suit order at %d", seat, i)
			}
		}
	}

	// **レベル札は末尾に固まる。**A より強いという序列が並びに出ていること。
	p := g.GetPlayer(0)
	last := p.GetCard(p.GetCardsSize() - 1)
	if !GuandanIsLevelCard(last, g.GetLevel()) && !GuandanIsJoker(last) {
		t.Errorf("the hand ends with %v, want a level card or a joker", last)
	}
}

func TestGuandanAccessors(t *testing.T) {
	g := NewDefaultGuandan()
	g.Reset()
	if got := g.GetHandNumber(); got != 1 {
		t.Errorf("hand number = %d, want 1", got)
	}
	if got := g.GetWinnerTeam(); got != -1 {
		t.Errorf("winner = %d, want -1", got)
	}
	if got := g.GetLastPlayerIdx(); got != -1 {
		t.Errorf("last player = %d, want -1", got)
	}
	if g.GetLastCombo() != nil || g.GetLastResult() != nil {
		t.Error("nothing has happened yet")
	}
	if got := len(g.GetPlayers()); got != GuandanPlayerCnt {
		t.Errorf("%d seats, want %d", got, GuandanPlayerCnt)
	}
	if g.GetPlayer(-1) != nil || g.GetPlayer(99) != nil {
		t.Error("an out-of-range seat must be nil")
	}
	if len(g.GetActionLog()) == 0 {
		t.Error("dealing writes to the action log")
	}
	if got := len(g.GetFinished()); got != 0 {
		t.Errorf("%d finishers, want none", got)
	}
	if got := len(g.GetTributes()); got != 0 {
		t.Errorf("%d tributes, want none in the first hand", got)
	}
	if g.IsTributeCancelled() {
		t.Error("nothing was cancelled in the first hand")
	}
	for _, team := range []int{-1, GuandanTeamCnt} {
		if got := g.GetTeamLevel(team); got != GuandanMinLevel {
			t.Errorf("team %d reads level %d", team, got)
		}
	}
	// **パートナーは向かい合わせ。範囲外は -1。**
	if GuandanTeamOf(0) != GuandanTeamOf(2) || GuandanTeamOf(1) != GuandanTeamOf(3) {
		t.Error("seats 0/2 and 1/3 are partners")
	}
	if GuandanTeamOf(-1) != -1 || GuandanTeamOf(GuandanPlayerCnt) != -1 {
		t.Error("an out-of-range seat has no team")
	}
	g.SetHandForTest(99, nil)
	g.SetTeamLevelForTest(99, 5)

	cfg := g.GetConfig()
	g.SetConfig(cfg)
	if g.GetConfig() != cfg {
		t.Error("SetConfig must take effect")
	}
	p := g.GetPlayer(0)
	if p.GetTeam(0) != p.GetTeam(2) || p.GetTeam(0) == p.GetTeam(1) {
		t.Error("seats 0/2 are partners and 0/1 are opponents")
	}
}

func TestGuandanConfigValidate(t *testing.T) {
	if err := DefaultGuandanConfig().Validate(); err != nil {
		t.Errorf("the default config must validate: %v", err)
	}
	if err := (GuandanConfig{CpuDifficulty: 9}).Validate(); err == nil {
		t.Error("a bad difficulty must not validate")
	}
}

func TestGuandanRoundTripsThroughJSON(t *testing.T) {
	g := NewDefaultGuandan()
	g.Reset()
	g.SetTeamLevelForTest(1, 7)

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored Guandan
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// **レベルが往復しないと序列そのものが変わる。**
	if got := restored.GetTeamLevel(1); got != 7 {
		t.Errorf("team 1 is at level %d, want 7", got)
	}
	if got := restored.GetLevel(); got != g.GetLevel() {
		t.Errorf("level = %d, want %d", got, g.GetLevel())
	}
	if got := restored.GetPlayer(0).GetCardsSize(); got != GuandanHandSize {
		t.Errorf("the restored hand holds %d, want %d", got, GuandanHandSize)
	}
}

// **壊れた状態を弾く。**KV から戻る値なので、範囲外のまま受け入れると詰む。
func TestGuandanRejectsBadJSON(t *testing.T) {
	base := `"pl":[{},{},{},{}],"cf":{"cd":0},"ph":0,"ci":0,"lp":-1,"wt":-1,"dt":0,"lc":2,"lv":[2,2]`
	for _, tc := range []struct {
		name string
		body string
	}{
		{"not json", `{`},
		{"wrong player count", `{"pl":[],"cf":{"cd":0},"ph":0}`},
		{"bad phase", `{` + base + `,"ph":99}`},
		{"bad current seat", `{` + base + `,"ci":9}`},
		{"bad last player", `{` + base + `,"lp":9}`},
		{"bad winner team", `{` + base + `,"wt":9}`},
		{"bad declarer team", `{` + base + `,"dt":9}`},
		{"bad level", `{` + base + `,"lc":99}`},
		{"bad team level", `{"pl":[{},{},{},{}],"cf":{"cd":0},"ph":0,"ci":0,"lp":-1,"wt":-1,"dt":0,"lc":2,"lv":[99,2]}`},
		{"too many finishers", `{` + base + `,"fi":[0,1,2,3,0]}`},
		{"bad finisher seat", `{` + base + `,"fi":[9]}`},
		{"bad tribute seat", `{` + base + `,"tb":[{"From":9,"To":0}]}`},
		{"unknown combination", `{` + base + `,"lb":{"Kind":99,"Rank":3,"Size":1}}`},
		{"bad config", `{"pl":[{},{},{},{}],"cf":{"cd":99},"ph":0,"ci":0,"lp":-1,"wt":-1,"dt":0,"lc":2,"lv":[2,2]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var g Guandan
			if err := json.Unmarshal([]byte(tc.body), &g); err == nil {
				t.Error("must be rejected")
			}
		})
	}
}
