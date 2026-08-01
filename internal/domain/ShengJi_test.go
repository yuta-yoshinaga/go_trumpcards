//go:build test

package domain

import "testing"

func sjiCard(suit, value int) *Card { return NewCard(suit, value, true) }

// sjiFresh は空の手札で始まる卓を返す。
func sjiFresh(t *testing.T, level, trump int) *ShengJi {
	t.Helper()
	s := NewDefaultShengJi()
	s.Reset()
	s.SetLevelForTest(level)
	s.SetTrumpForTest(trump)
	for i := range ShengJiPlayerCnt {
		s.SetHandForTest(i, nil)
	}
	s.SetKittyForTest(nil)
	s.SetPhaseForTest(ShengJiPhasePlay)
	s.SetCurrentPlayerForTest(0)
	return s
}

// **108 は 4 で割り切れる。**それでも 27 枚ずつ配ってはいけない。
func TestShengJiDealsTwentyFiveAndAnEightCardKitty(t *testing.T) {
	s := NewDefaultShengJi()
	s.Reset()

	for i := range ShengJiPlayerCnt {
		if got := s.GetPlayer(i).GetCardsSize(); got != ShengJiHandSize {
			t.Errorf("seat %d holds %d cards, want %d", i, got, ShengJiHandSize)
		}
	}
	if got := s.GetKittySize(); got != ShengJiKittySize {
		t.Errorf("the kitty holds %d cards, want %d", got, ShengJiKittySize)
	}
	if total := ShengJiHandSize*ShengJiPlayerCnt + ShengJiKittySize; total != ShengJiDeckSize {
		t.Errorf("%d cards accounted for, want %d", total, ShengJiDeckSize)
	}
}

// **底牌の中身は終局まで見えない。**見えると宣言側の埋め方が筒抜けになる。
func TestShengJiHidesTheKittyUntilTheHandIsSettled(t *testing.T) {
	s := NewDefaultShengJi()
	s.Reset()
	if s.GetKitty() != nil {
		t.Error("the kitty must stay hidden while the hand is live")
	}
	s.SetPhaseForTest(ShengJiPhaseHandEnd)
	if s.GetKitty() == nil {
		t.Error("the kitty must be visible once the hand is settled")
	}
}

// **切札は切札スートだけではない。**全スートのレベル札とジョーカーも切札。
func TestShengJiTrumpGroupIncludesEveryLevelCardAndJoker(t *testing.T) {
	const level, trump = 5, CardDesignSpade

	for _, tc := range []struct {
		name  string
		card  *Card
		trump bool
	}{
		{"trump suit", sjiCard(CardDesignSpade, 7), true},
		{"trump-suit level card", sjiCard(CardDesignSpade, level), true},
		{"off-suit level card", sjiCard(CardDesignHeart, level), true},
		{"black joker", sjiCard(CardDesignJoker, 1), true},
		{"red joker", sjiCard(CardDesignJoker, 2), true},
		{"a plain card", sjiCard(CardDesignHeart, 7), false},
		{"a plain ace", sjiCard(CardDesignDiamond, 1), false},
	} {
		if got := ShengJiIsTrump(tc.card, level, trump); got != tc.trump {
			t.Errorf("%s: trump=%v, want %v", tc.name, got, tc.trump)
		}
	}
}

// **無主でもレベル札とジョーカーは切札。**
func TestShengJiNoTrumpStillHasTrumps(t *testing.T) {
	const level = 5
	if !ShengJiIsTrump(sjiCard(CardDesignHeart, level), level, ShengJiNoTrump) {
		t.Error("a level card is a trump even with no trump suit")
	}
	if !ShengJiIsTrump(sjiCard(CardDesignJoker, 1), level, ShengJiNoTrump) {
		t.Error("a joker is a trump even with no trump suit")
	}
	if ShengJiIsTrump(sjiCard(CardDesignSpade, 7), level, ShengJiNoTrump) {
		t.Error("a plain card is not a trump with no trump suit")
	}
}

// 序列: 赤 > 黒 > 切札スートのレベル札 > 他スートのレベル札 > 切札スート > 平札。
func TestShengJiStrengthOrder(t *testing.T) {
	const level, trump = 5, CardDesignSpade
	order := []*Card{
		sjiCard(CardDesignHeart, 7),     // 平札
		sjiCard(CardDesignSpade, 7),     // 切札スート
		sjiCard(CardDesignSpade, 1),     // 切札スートの A
		sjiCard(CardDesignHeart, level), // 他スートのレベル札
		sjiCard(CardDesignSpade, level), // 切札スートのレベル札
		sjiCard(CardDesignJoker, 1),     // 黒
		sjiCard(CardDesignJoker, 2),     // 赤
	}
	for i := 1; i < len(order); i++ {
		lo := ShengJiStrength(order[i-1], level, trump)
		hi := ShengJiStrength(order[i], level, trump)
		if hi <= lo {
			t.Errorf("card %d (%d) must beat card %d (%d)", i, hi, i-1, lo)
		}
	}
	// **他スートのレベル札どうしは同格。**先に出したほうが勝つ。
	a := ShengJiStrength(sjiCard(CardDesignHeart, level), level, trump)
	b := ShengJiStrength(sjiCard(CardDesignDiamond, level), level, trump)
	if a != b {
		t.Errorf("off-suit level cards must rank equal, got %d and %d", a, b)
	}
}

// **得点札は 5 と 10 と K で、合計 200 点。**80 点はその 4 割。
func TestShengJiPointCardsTotalTwoHundred(t *testing.T) {
	total := 0
	for _, c := range newShengJiDeck() {
		total += ShengJiCardPoints(c)
	}
	if total != ShengJiTotalPoints {
		t.Errorf("the deck holds %d points, want %d", total, ShengJiTotalPoints)
	}
	if ShengJiCardPoints(sjiCard(CardDesignSpade, 5)) != 5 {
		t.Error("a five is worth five")
	}
	if ShengJiCardPoints(sjiCard(CardDesignSpade, 13)) != 10 {
		t.Error("a king is worth ten")
	}
	if ShengJiCardPoints(sjiCard(CardDesignSpade, 12)) != 0 {
		t.Error("a queen is worth nothing")
	}
	if ShengJiCardPoints(sjiCard(CardDesignJoker, 2)) != 0 {
		t.Error("a joker is worth nothing")
	}
}

func TestShengJiEvaluate(t *testing.T) {
	const level, trump = 5, CardDesignSpade

	t.Run("single", func(t *testing.T) {
		c := ShengJiEvaluate([]*Card{sjiCard(CardDesignHeart, 7)}, level, trump)
		if c == nil || c.Kind != ShengJiComboSingle {
			t.Fatalf("got %v, want a single", c)
		}
	})

	// **対子は同ランクではなく同一の札 2 枚。**2 デッキなので ♠K が 2 枚ある。
	t.Run("a pair needs two identical cards", func(t *testing.T) {
		same := ShengJiEvaluate([]*Card{sjiCard(CardDesignHeart, 7), sjiCard(CardDesignHeart, 7)}, level, trump)
		if same == nil || same.Kind != ShengJiComboPair {
			t.Fatalf("got %v, want a pair", same)
		}
		mixed := ShengJiEvaluate([]*Card{sjiCard(CardDesignHeart, 7), sjiCard(CardDesignDiamond, 7)}, level, trump)
		if mixed != nil {
			t.Errorf("got %v, want nothing -- two suits are not a pair here", mixed)
		}
	})

	t.Run("a tractor is consecutive pairs of one suit", func(t *testing.T) {
		c := ShengJiEvaluate([]*Card{
			sjiCard(CardDesignHeart, 7), sjiCard(CardDesignHeart, 7),
			sjiCard(CardDesignHeart, 8), sjiCard(CardDesignHeart, 8),
		}, level, trump)
		if c == nil || c.Kind != ShengJiComboTractor {
			t.Fatalf("got %v, want a tractor", c)
		}
		if c.Size != 4 {
			t.Errorf("size %d, want 4", c.Size)
		}
	})

	// **レベル札はそのスートの平札から抜ける。**レベルが 5 なら 4 と 6 は隣。
	t.Run("a tractor closes over the level rank", func(t *testing.T) {
		c := ShengJiEvaluate([]*Card{
			sjiCard(CardDesignHeart, 4), sjiCard(CardDesignHeart, 4),
			sjiCard(CardDesignHeart, 6), sjiCard(CardDesignHeart, 6),
		}, level, trump)
		if c == nil || c.Kind != ShengJiComboTractor {
			t.Fatalf("got %v, want a tractor -- 4 and 6 are adjacent when 5 is the level", c)
		}
	})

	t.Run("non-consecutive pairs are not a tractor", func(t *testing.T) {
		c := ShengJiEvaluate([]*Card{
			sjiCard(CardDesignHeart, 7), sjiCard(CardDesignHeart, 7),
			sjiCard(CardDesignHeart, 10), sjiCard(CardDesignHeart, 10),
		}, level, trump)
		if c != nil {
			t.Errorf("got %v, want nothing", c)
		}
	})

	// **1 手はひとつのスートから。**切札群はひとつのスートとして扱う。
	t.Run("mixed suits are not a combination", func(t *testing.T) {
		c := ShengJiEvaluate([]*Card{sjiCard(CardDesignHeart, 7), sjiCard(CardDesignDiamond, 8)}, level, trump)
		if c != nil {
			t.Errorf("got %v, want nothing", c)
		}
	})

	t.Run("odd counts above one are not a combination", func(t *testing.T) {
		c := ShengJiEvaluate([]*Card{
			sjiCard(CardDesignHeart, 7), sjiCard(CardDesignHeart, 7), sjiCard(CardDesignHeart, 8),
		}, level, trump)
		if c != nil {
			t.Errorf("got %v, want nothing", c)
		}
	})

	t.Run("nothing at all", func(t *testing.T) {
		if c := ShengJiEvaluate(nil, level, trump); c != nil {
			t.Errorf("got %v, want nothing", c)
		}
		if c := ShengJiEvaluate([]*Card{nil}, level, trump); c != nil {
			t.Errorf("got %v, want nothing", c)
		}
	})
}

func TestShengJiBeats(t *testing.T) {
	const level, trump = 5, CardDesignSpade
	single := func(c *Card) *ShengJiCombo { return ShengJiEvaluate([]*Card{c}, level, trump) }

	lead := single(sjiCard(CardDesignHeart, 7))
	ledSuit := lead.Suit

	t.Run("a higher card of the led suit wins", func(t *testing.T) {
		if !ShengJiBeats(single(sjiCard(CardDesignHeart, 9)), lead, ledSuit) {
			t.Error("a nine must beat a seven in the led suit")
		}
	})

	t.Run("a lower card of the led suit loses", func(t *testing.T) {
		if ShengJiBeats(single(sjiCard(CardDesignHeart, 3)), lead, ledSuit) {
			t.Error("a three must not beat a seven")
		}
	})

	// **切札だけが割り込める。**
	t.Run("a trump beats a plain card", func(t *testing.T) {
		if !ShengJiBeats(single(sjiCard(CardDesignSpade, 2)), lead, ledSuit) {
			t.Error("the lowest trump must beat a plain seven")
		}
	})

	// **別スートの平札は勝てない。**切らずに捨てただけ。
	t.Run("a plain card of another suit never wins", func(t *testing.T) {
		if ShengJiBeats(single(sjiCard(CardDesignDiamond, 1)), lead, ledSuit) {
			t.Error("an off-suit ace must not win")
		}
	})

	t.Run("the shape and the count must match", func(t *testing.T) {
		pair := ShengJiEvaluate([]*Card{sjiCard(CardDesignHeart, 9), sjiCard(CardDesignHeart, 9)}, level, trump)
		if ShengJiBeats(pair, lead, ledSuit) {
			t.Error("a pair must not beat a single")
		}
	})

	t.Run("nil never wins", func(t *testing.T) {
		if ShengJiBeats(nil, lead, ledSuit) || ShengJiBeats(lead, nil, ledSuit) {
			t.Error("nil must not win")
		}
	})
}

// **リードされたスートを持っているなら、そこから出さなければならない。**
func TestShengJiMustFollowTheLedSuit(t *testing.T) {
	const level, trump = 5, CardDesignSpade
	s := sjiFresh(t, level, trump)
	s.SetHandForTest(0, []*Card{sjiCard(CardDesignHeart, 7)})
	s.SetHandForTest(1, []*Card{sjiCard(CardDesignHeart, 9), sjiCard(CardDesignDiamond, 3)})

	if err := s.Play(0, []int{0}); err != nil {
		t.Fatalf("the lead was rejected: %v", err)
	}
	if err := s.Play(1, []int{1}); err == nil {
		t.Error("discarding while holding the led suit must be rejected")
	}
	if err := s.Play(1, []int{0}); err != nil {
		t.Errorf("following the led suit was rejected: %v", err)
	}
}

// **対子がリードされたら、そのスートの対子を先に出さなければならない。**
// 枚数だけ合わせて対子を温存できると、拖拉機を出す意味が無くなる。
func TestShengJiMustPlayPairsWhenAPairIsLed(t *testing.T) {
	const level, trump = 5, CardDesignSpade
	s := sjiFresh(t, level, trump)
	s.SetHandForTest(0, []*Card{sjiCard(CardDesignHeart, 7), sjiCard(CardDesignHeart, 7)})
	// 席 1 は ♥ の対子と、ばらの ♥ を 2 枚持つ。
	s.SetHandForTest(1, []*Card{
		sjiCard(CardDesignHeart, 9), sjiCard(CardDesignHeart, 9),
		sjiCard(CardDesignHeart, 3), sjiCard(CardDesignHeart, 4),
	})

	if err := s.Play(0, []int{0, 1}); err != nil {
		t.Fatalf("the lead was rejected: %v", err)
	}
	// ばら 2 枚で枚数だけ合わせるのは不可。
	if err := s.Play(1, []int{2, 3}); err == nil {
		t.Error("holding a pair of the led suit, two odd cards must be rejected")
	}
	if err := s.Play(1, []int{0, 1}); err != nil {
		t.Errorf("playing the pair was rejected: %v", err)
	}
}

// 対子を持っていなければ、そのスートのばら札で構わない。
func TestShengJiOddCardsAreFineWithoutAPair(t *testing.T) {
	const level, trump = 5, CardDesignSpade
	s := sjiFresh(t, level, trump)
	s.SetHandForTest(0, []*Card{sjiCard(CardDesignHeart, 7), sjiCard(CardDesignHeart, 7)})
	s.SetHandForTest(1, []*Card{sjiCard(CardDesignHeart, 3), sjiCard(CardDesignHeart, 4)})

	if err := s.Play(0, []int{0, 1}); err != nil {
		t.Fatalf("the lead was rejected: %v", err)
	}
	if err := s.Play(1, []int{0, 1}); err != nil {
		t.Errorf("without a pair, odd cards of the led suit must be allowed: %v", err)
	}
}

// スートを持っていなければ何を出してもよい。
func TestShengJiVoidPlayerMayDiscard(t *testing.T) {
	const level, trump = 5, CardDesignSpade
	s := sjiFresh(t, level, trump)
	s.SetHandForTest(0, []*Card{sjiCard(CardDesignHeart, 7)})
	s.SetHandForTest(1, []*Card{sjiCard(CardDesignDiamond, 3)})

	if err := s.Play(0, []int{0}); err != nil {
		t.Fatalf("the lead was rejected: %v", err)
	}
	if err := s.Play(1, []int{0}); err != nil {
		t.Errorf("a void player must be free to discard: %v", err)
	}
}

func TestShengJiPlayGuards(t *testing.T) {
	const level, trump = 5, CardDesignSpade
	s := sjiFresh(t, level, trump)
	s.SetHandForTest(0, []*Card{sjiCard(CardDesignHeart, 7), sjiCard(CardDesignDiamond, 8)})

	if err := s.Play(1, []int{0}); err == nil {
		t.Error("playing out of turn must be rejected")
	}
	if err := s.Play(0, nil); err == nil {
		t.Error("playing nothing must be rejected")
	}
	if err := s.Play(0, []int{9}); err == nil {
		t.Error("an out-of-range index must be rejected")
	}
	// **ひとつのスートから出さないと形にならない。**
	if err := s.Play(0, []int{0, 1}); err == nil {
		t.Error("a mixed-suit lead must be rejected")
	}
}

// **点を集めるのは守備側。**宣言側が取った点はどこにも積まれない。
func TestShengJiOnlyDefendersCollectPoints(t *testing.T) {
	const level, trump = 5, CardDesignSpade
	s := sjiFresh(t, level, trump)
	declarers := s.GetDeclarerTeam()
	defenders := 1 - declarers

	// 席 1 (守備側) が K を含むトリックを取る。
	s.SetHandForTest(0, []*Card{sjiCard(CardDesignHeart, 7)})
	s.SetHandForTest(1, []*Card{sjiCard(CardDesignHeart, 13)})
	s.SetHandForTest(2, []*Card{sjiCard(CardDesignHeart, 3)})
	s.SetHandForTest(3, []*Card{sjiCard(CardDesignHeart, 4)})
	for seat := range ShengJiPlayerCnt {
		if err := s.Play(seat, []int{0}); err != nil {
			t.Fatalf("seat %d was rejected: %v", seat, err)
		}
	}
	if got := s.GetTeamPoints(defenders); got != 10 {
		t.Errorf("the defenders hold %d points, want 10", got)
	}
	if got := s.GetTeamPoints(declarers); got != 0 {
		t.Errorf("the declarers hold %d points, want 0 -- they do not collect", got)
	}
}

// **宣言側が守りきったときの昇級は 0 点で 3、40 点未満で 2、80 点未満で 1。**
func TestShengJiDeclarerAdvanceTable(t *testing.T) {
	for _, tc := range []struct {
		points, want int
	}{
		{0, 3},
		{5, 2},
		{35, 2},
		{40, 1},
		{75, 1},
	} {
		if got := shengJiDeclarerAdvance(tc.points); got != tc.want {
			t.Errorf("%d points: advance %d, want %d", tc.points, got, tc.want)
		}
	}
}

func TestShengJiSettlement(t *testing.T) {
	const level, trump = 5, CardDesignSpade

	t.Run("the declarers hold and climb", func(t *testing.T) {
		s := sjiFresh(t, level, trump)
		declarers := s.GetDeclarerTeam()
		before := s.GetTeamLevel(declarers)
		s.FinishHandForTest()

		r := s.GetLastResult()
		if r == nil || !r.DeclarerHeld {
			t.Fatalf("result %v, want the declarers to have held", r)
		}
		// 守備側 0 点なので 3 段階。
		if r.Advance != 3 {
			t.Errorf("advance %d, want 3", r.Advance)
		}
		if got := s.GetTeamLevel(declarers); got != before+3 {
			t.Errorf("level %d, want %d", got, before+3)
		}
		if s.GetDeclarerTeam() != declarers {
			t.Error("the declarers keep the deal when they hold")
		}
	})

	// **80 点で宣言側が交代する。**そこは昇級 0。
	t.Run("eighty points takes the deal without a climb", func(t *testing.T) {
		s := sjiFresh(t, level, trump)
		declarers := s.GetDeclarerTeam()
		defenders := 1 - declarers
		beforeDef := s.GetTeamLevel(defenders)
		s.teamPoints[defenders] = ShengJiDefenderTarget
		s.FinishHandForTest()

		r := s.GetLastResult()
		if r == nil || r.DeclarerHeld {
			t.Fatalf("result %v, want the declarers to have been beaten", r)
		}
		if r.Advance != 0 {
			t.Errorf("advance %d, want 0 at exactly eighty", r.Advance)
		}
		if got := s.GetTeamLevel(defenders); got != beforeDef {
			t.Errorf("level %d, want %d -- eighty only takes the deal", got, beforeDef)
		}
		if s.GetDeclarerTeam() != defenders {
			t.Error("the defenders must take over the deal")
		}
	})

	// 80 を超えてから **40 点ごとに 1 段階。**
	t.Run("every forty above eighty is one level", func(t *testing.T) {
		s := sjiFresh(t, level, trump)
		defenders := 1 - s.GetDeclarerTeam()
		before := s.GetTeamLevel(defenders)
		s.teamPoints[defenders] = ShengJiDefenderTarget + 2*ShengJiAdvanceStep
		s.FinishHandForTest()

		if got := s.GetLastResult().Advance; got != 2 {
			t.Errorf("advance %d, want 2", got)
		}
		if got := s.GetTeamLevel(defenders); got != before+2 {
			t.Errorf("level %d, want %d", got, before+2)
		}
	})
}

// **A は飛び越えられない。**K から 3 段階でも A で止まり、A の局を守りきって
// 初めて勝ちになる (打A)。
func TestShengJiCannotSkipPastTheAce(t *testing.T) {
	const level, trump = 5, CardDesignSpade

	t.Run("a big climb stops at the ace", func(t *testing.T) {
		s := sjiFresh(t, level, trump)
		declarers := s.GetDeclarerTeam()
		s.SetTeamLevelForTest(declarers, 13) // K
		s.FinishHandForTest()

		if got := s.GetTeamLevel(declarers); got != ShengJiMaxLevel {
			t.Errorf("level %d, want the ace (%d)", got, ShengJiMaxLevel)
		}
		if s.GetGameEndFlag() {
			t.Error("reaching the ace is not winning -- the ace hand still has to be held")
		}
	})

	t.Run("holding the deal at the ace wins", func(t *testing.T) {
		s := sjiFresh(t, level, trump)
		declarers := s.GetDeclarerTeam()
		s.SetTeamLevelForTest(declarers, ShengJiMaxLevel)
		s.FinishHandForTest()

		if !s.GetGameEndFlag() {
			t.Fatal("holding the deal at the ace must win the game")
		}
		if s.GetWinnerTeam() != declarers {
			t.Errorf("winner %d, want %d", s.GetWinnerTeam(), declarers)
		}
		if s.GetPhase() != ShengJiPhaseGameEnd {
			t.Errorf("phase %v, want the game-end phase", s.GetPhase())
		}
	})

	// 交代しただけ (80 点ちょうど) では誰も上がらないので終わらない。
	t.Run("taking the deal at the ace does not win", func(t *testing.T) {
		s := sjiFresh(t, level, trump)
		defenders := 1 - s.GetDeclarerTeam()
		s.SetTeamLevelForTest(defenders, ShengJiMaxLevel)
		s.teamPoints[defenders] = ShengJiDefenderTarget
		s.FinishHandForTest()

		if s.GetGameEndFlag() {
			t.Error("taking the deal with no climb must not end the game")
		}
	})
}

// **守備側が最終トリックを取ると、底牌が倍率つきで守備側に入る。**
func TestShengJiKittyGoesToTheDefendersWithAMultiplier(t *testing.T) {
	const level, trump = 5, CardDesignSpade
	s := sjiFresh(t, level, trump)
	declarers := s.GetDeclarerTeam()
	defenders := 1 - declarers

	// 底牌に K が 1 枚 = 10 点。
	s.SetKittyForTest([]*Card{sjiCard(CardDesignDiamond, 13)})

	// 席 1 (守備側) が最後のトリックを単張で取る。
	s.SetHandForTest(0, []*Card{sjiCard(CardDesignHeart, 7)})
	s.SetHandForTest(1, []*Card{sjiCard(CardDesignHeart, 9)})
	s.SetHandForTest(2, []*Card{sjiCard(CardDesignHeart, 3)})
	s.SetHandForTest(3, []*Card{sjiCard(CardDesignHeart, 4)})
	for seat := range ShengJiPlayerCnt {
		if err := s.Play(seat, []int{0}); err != nil {
			t.Fatalf("seat %d was rejected: %v", seat, err)
		}
	}

	r := s.GetLastResult()
	if r == nil {
		t.Fatal("the hand did not settle")
	}
	if ShengJiTeamOf(s.GetLastTrickWinner()) != defenders {
		t.Fatalf("seat %d took the last trick, want a defender", s.GetLastTrickWinner())
	}
	// 単張なので倍率は 2。10 点 × 2 = 20 点。
	if r.KittyMultiplier != 2 {
		t.Errorf("multiplier %d, want 2 for a single", r.KittyMultiplier)
	}
	if r.KittyPoints != 20 {
		t.Errorf("kitty points %d, want 20", r.KittyPoints)
	}
	if r.DefenderPoints != 20 {
		t.Errorf("defender points %d, want 20", r.DefenderPoints)
	}
}

// 宣言側が最終トリックを取れば底牌は動かない。
func TestShengJiKittyStaysWhenTheDeclarersTakeTheLastTrick(t *testing.T) {
	const level, trump = 5, CardDesignSpade
	s := sjiFresh(t, level, trump)

	s.SetKittyForTest([]*Card{sjiCard(CardDesignDiamond, 13)})
	s.SetHandForTest(0, []*Card{sjiCard(CardDesignHeart, 9)})
	s.SetHandForTest(1, []*Card{sjiCard(CardDesignHeart, 7)})
	s.SetHandForTest(2, []*Card{sjiCard(CardDesignHeart, 3)})
	s.SetHandForTest(3, []*Card{sjiCard(CardDesignHeart, 4)})
	for seat := range ShengJiPlayerCnt {
		if err := s.Play(seat, []int{0}); err != nil {
			t.Fatalf("seat %d was rejected: %v", seat, err)
		}
	}

	r := s.GetLastResult()
	if r == nil {
		t.Fatal("the hand did not settle")
	}
	if r.KittyPoints != 0 || r.KittyMultiplier != 0 {
		t.Errorf("kitty %d x%d, want nothing", r.KittyPoints, r.KittyMultiplier)
	}
}

func TestShengJiDeclare(t *testing.T) {
	const level = 5
	newTable := func() *ShengJi {
		s := NewDefaultShengJi()
		s.Reset()
		s.SetLevelForTest(level)
		for i := range ShengJiPlayerCnt {
			s.SetHandForTest(i, nil)
		}
		s.SetPhaseForTest(ShengJiPhaseDeclare)
		s.SetCurrentPlayerForTest(0)
		return s
	}

	// **亮牌はレベル札を見せて行う。**持っていなければ宣言できない。
	t.Run("you must hold a level card of that suit", func(t *testing.T) {
		s := newTable()
		s.SetHandForTest(0, []*Card{sjiCard(CardDesignHeart, 7)})
		if err := s.Declare(0, CardDesignHeart); err == nil {
			t.Error("declaring without a level card must be rejected")
		}
	})

	t.Run("a single level card declares", func(t *testing.T) {
		s := newTable()
		s.SetHandForTest(0, []*Card{sjiCard(CardDesignHeart, level)})
		if err := s.Declare(0, CardDesignHeart); err != nil {
			t.Fatalf("declaring was rejected: %v", err)
		}
		if s.GetTrumpSuit() != CardDesignHeart {
			t.Errorf("trump %d, want hearts", s.GetTrumpSuit())
		}
		if d := s.GetDeclaration(); d == nil || d.Strength != 1 {
			t.Errorf("declaration %v, want strength 1", d)
		}
	})

	// **強い宣言だけが上書きできる。**同じ強さでは覆せない。
	t.Run("only a stronger declaration overrides", func(t *testing.T) {
		s := newTable()
		s.SetHandForTest(0, []*Card{sjiCard(CardDesignHeart, level)})
		s.SetHandForTest(1, []*Card{sjiCard(CardDesignDiamond, level)})
		s.SetHandForTest(2, []*Card{sjiCard(CardDesignClover, level), sjiCard(CardDesignClover, level)})

		if err := s.Declare(0, CardDesignHeart); err != nil {
			t.Fatalf("the first declaration was rejected: %v", err)
		}
		if err := s.Declare(1, CardDesignDiamond); err == nil {
			t.Error("an equal declaration must not override")
		}
		_ = s.Declare(1, ShengJiNoTrump)
		if err := s.Declare(2, CardDesignClover); err != nil {
			t.Fatalf("a pair must override a single: %v", err)
		}
		if s.GetTrumpSuit() != CardDesignClover {
			t.Errorf("trump %d, want clovers", s.GetTrumpSuit())
		}
	})

	// **誰も亮牌しなければ無主。**
	t.Run("nobody declaring leaves no trump suit", func(t *testing.T) {
		s := newTable()
		for seat := range ShengJiPlayerCnt {
			if err := s.Declare(seat, ShengJiNoTrump); err != nil {
				t.Fatalf("passing was rejected: %v", err)
			}
		}
		if s.GetTrumpSuit() != ShengJiNoTrump {
			t.Errorf("trump %d, want no trump", s.GetTrumpSuit())
		}
		if s.GetPhase() != ShengJiPhaseKitty {
			t.Errorf("phase %v, want the kitty phase", s.GetPhase())
		}
	})

	t.Run("declaring out of phase or out of turn is rejected", func(t *testing.T) {
		s := newTable()
		s.SetHandForTest(0, []*Card{sjiCard(CardDesignHeart, level)})
		if err := s.Declare(1, CardDesignHeart); err == nil {
			t.Error("declaring out of turn must be rejected")
		}
		s.SetPhaseForTest(ShengJiPhasePlay)
		if err := s.Declare(0, CardDesignHeart); err == nil {
			t.Error("declaring outside the declaring phase must be rejected")
		}
	})
}

// **底牌には必ず 8 枚戻す。**
func TestShengJiBuryKitty(t *testing.T) {
	s := NewDefaultShengJi()
	s.Reset()
	for seat := range ShengJiPlayerCnt {
		if err := s.Declare(seat, ShengJiNoTrump); err != nil {
			t.Fatalf("passing was rejected: %v", err)
		}
	}
	if s.GetPhase() != ShengJiPhaseKitty {
		t.Fatalf("phase %v, want the kitty phase", s.GetPhase())
	}
	taker := s.GetCurrentPlayerIdx()
	if got := s.GetPlayer(taker).GetCardsSize(); got != ShengJiHandSize+ShengJiKittySize {
		t.Fatalf("the taker holds %d cards, want %d", got, ShengJiHandSize+ShengJiKittySize)
	}

	if err := s.BuryKitty(taker, []int{0, 1}); err == nil {
		t.Error("burying the wrong number of cards must be rejected")
	}
	if err := s.BuryKitty(taker, []int{0, 0, 1, 2, 3, 4, 5, 6}); err == nil {
		t.Error("burying the same card twice must be rejected")
	}

	if err := s.BuryKitty(taker, []int{0, 1, 2, 3, 4, 5, 6, 7}); err != nil {
		t.Fatalf("burying was rejected: %v", err)
	}
	if got := s.GetPlayer(taker).GetCardsSize(); got != ShengJiHandSize {
		t.Errorf("the taker holds %d cards, want %d", got, ShengJiHandSize)
	}
	if got := s.GetKittySize(); got != ShengJiKittySize {
		t.Errorf("the kitty holds %d cards, want %d", got, ShengJiKittySize)
	}
	if s.GetPhase() != ShengJiPhasePlay {
		t.Errorf("phase %v, want the playing phase", s.GetPhase())
	}
	// **リードするのは底牌を取った側。**
	if s.GetCurrentPlayerIdx() != taker {
		t.Errorf("seat %d leads, want %d", s.GetCurrentPlayerIdx(), taker)
	}
}

// **CPU だけで 1 局を最後まで回せること。**手が止まると局が終わらない。
func TestShengJiCpuDrivesAFullHand(t *testing.T) {
	s := NewDefaultShengJi()
	for i := range ShengJiPlayerCnt {
		s.players[i] = NewShengJiPlayer(false)
	}
	s.Reset()

	for step := 0; step < 5000; step++ {
		if s.GetGameEndFlag() || s.GetPhase() == ShengJiPhaseHandEnd {
			break
		}
		s.CpuPlay()
	}
	if s.GetPhase() != ShengJiPhaseHandEnd && !s.GetGameEndFlag() {
		t.Fatalf("the hand stalled in phase %v after 5000 steps", s.GetPhase())
	}
	if s.GetLastResult() == nil {
		t.Error("the hand settled without a result")
	}
	for i := range ShengJiPlayerCnt {
		if got := s.GetPlayer(i).GetCardsSize(); got != 0 {
			t.Errorf("seat %d still holds %d cards", i, got)
		}
	}
}

// **CPU が対子リードに応じられないと局が永久に止まる。**checkFollow は対子の
// 温存を禁じるので、CPU が弱い単札から拾うと自分の対子を割って弾かれる。
// 単札 1 枚のフォールバックも枚数が合わずに弾かれ、手番が進まなくなる。
func TestShengJiCpuFollowsAPairWithoutStalling(t *testing.T) {
	const level, trump = 5, CardDesignSpade
	s := sjiFresh(t, level, trump)

	// 人間が ♥9 の対子をリードする。
	s.SetHandForTest(0, []*Card{sjiCard(CardDesignHeart, 9), sjiCard(CardDesignHeart, 9)})
	// 席 1 は「弱い単札 + 強い対子」。素朴に弱い順で 2 枚拾うと対子が割れる。
	s.SetHandForTest(1, []*Card{
		sjiCard(CardDesignHeart, 3), sjiCard(CardDesignHeart, 6), sjiCard(CardDesignHeart, 6),
	})
	s.SetHandForTest(2, []*Card{sjiCard(CardDesignHeart, 4), sjiCard(CardDesignHeart, 7)})
	s.SetHandForTest(3, []*Card{sjiCard(CardDesignHeart, 8), sjiCard(CardDesignHeart, 10)})

	if err := s.Play(0, []int{0, 1}); err != nil {
		t.Fatalf("the lead was rejected: %v", err)
	}
	if s.GetCurrentPlayerIdx() != 1 {
		t.Fatalf("seat %d is on turn, want seat 1", s.GetCurrentPlayerIdx())
	}

	before := s.GetPlayer(1).GetCardsSize()
	s.CpuPlay()
	if got := s.GetPlayer(1).GetCardsSize(); got == before {
		t.Fatal("the CPU played nothing -- the seat is stuck and the hand can never finish")
	}
	if s.GetCurrentPlayerIdx() == 1 {
		t.Fatal("the turn did not advance past seat 1")
	}
}

// 対子リードに対しては、CPU が持っている対子を割らずに出すこと。
func TestShengJiCpuKeepsItsPairIntact(t *testing.T) {
	const level, trump = 5, CardDesignSpade
	s := sjiFresh(t, level, trump)
	s.SetHandForTest(0, []*Card{sjiCard(CardDesignHeart, 9), sjiCard(CardDesignHeart, 9)})
	s.SetHandForTest(1, []*Card{
		sjiCard(CardDesignHeart, 3), sjiCard(CardDesignHeart, 6), sjiCard(CardDesignHeart, 6),
	})
	if err := s.Play(0, []int{0, 1}); err != nil {
		t.Fatalf("the lead was rejected: %v", err)
	}

	idxs := s.ShengJiCpuPlay(1)
	if len(idxs) != 2 {
		t.Fatalf("the CPU offered %d cards, want 2", len(idxs))
	}
	cards := make([]*Card, 0, len(idxs))
	for _, i := range idxs {
		cards = append(cards, s.GetPlayer(1).GetCard(i))
	}
	if shengJiPairCount(cards) != 1 {
		t.Errorf("the CPU split its pair: %v", cards)
	}
}

func TestShengJiNextHandGuards(t *testing.T) {
	s := NewDefaultShengJi()
	s.Reset()
	if err := s.NextHand(); err == nil {
		t.Error("dealing the next hand mid-play must be rejected")
	}
	s.SetPhaseForTest(ShengJiPhaseHandEnd)
	before := s.GetHandNumber()
	if err := s.NextHand(); err != nil {
		t.Fatalf("the next hand was rejected: %v", err)
	}
	if got := s.GetHandNumber(); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
	if s.GetPhase() != ShengJiPhaseDeclare {
		t.Errorf("phase %v, want the declaring phase", s.GetPhase())
	}
}

func TestShengJiAccessors(t *testing.T) {
	s := NewDefaultShengJi()
	s.Reset()

	if s.GetPlayer(-1) != nil || s.GetPlayer(ShengJiPlayerCnt) != nil {
		t.Error("out-of-range seats must return nil")
	}
	if len(s.GetPlayers()) != ShengJiPlayerCnt {
		t.Errorf("%d seats, want %d", len(s.GetPlayers()), ShengJiPlayerCnt)
	}
	if s.GetTeamLevel(-1) != 0 || s.GetTeamLevel(ShengJiTeamCnt) != 0 {
		t.Error("out-of-range teams must return zero")
	}
	if s.GetTeamPoints(-1) != 0 || s.GetTeamPoints(ShengJiTeamCnt) != 0 {
		t.Error("out-of-range teams must return zero points")
	}
	if ShengJiTeamOf(-1) != -1 || ShengJiTeamOf(ShengJiPlayerCnt) != -1 {
		t.Error("out-of-range seats have no team")
	}
	if ShengJiTeamOf(0) != ShengJiTeamOf(2) || ShengJiTeamOf(0) == ShengJiTeamOf(1) {
		t.Error("partners must sit opposite")
	}
	if s.GetHandNumber() != 1 || s.GetWinnerTeam() != -1 || s.GetGameEndFlag() {
		t.Error("a fresh game is misreported")
	}
	if s.GetTrickCount() != 0 || s.GetLastTrickWinner() != -1 || s.GetLeadCombo() != nil {
		t.Error("a fresh hand is misreported")
	}
	if len(s.GetTrick()) != 0 {
		t.Error("a fresh trick must be empty")
	}
	if s.GetTrickLeader() < 0 || s.GetTrickLeader() >= ShengJiPlayerCnt {
		t.Errorf("trick leader %d is out of range", s.GetTrickLeader())
	}
	if len(s.GetActionLog()) == 0 {
		t.Error("dealing must be logged")
	}

	cfg := s.GetConfig()
	s.SetConfig(cfg)
	if s.GetConfig() != cfg {
		t.Error("the config did not round-trip")
	}
	if !s.IsHumanTurn() {
		t.Error("seat 0 is the human and opens the declaring phase")
	}
}

// nil や範囲外を渡しても落ちないこと。**GetCard は範囲外で nil を返す。**
func TestShengJiHandlesNilAndOutOfRange(t *testing.T) {
	const level, trump = 5, CardDesignSpade

	if shengJiNaturalRank(nil) != 0 {
		t.Error("a nil card has no rank")
	}
	if ShengJiIsJoker(nil) || ShengJiIsLevelCard(nil, level) || ShengJiIsTrump(nil, level, trump) {
		t.Error("a nil card is nothing")
	}
	if ShengJiStrength(nil, level, trump) != 0 {
		t.Error("a nil card has no strength")
	}
	if ShengJiCardPoints(nil) != 0 {
		t.Error("a nil card is worth nothing")
	}
	// **ジョーカーはレベル札ではない。**値がレベルと一致しても別枠。
	if ShengJiIsLevelCard(sjiCard(CardDesignJoker, 1), 1) {
		t.Error("a joker is never a level card")
	}
	if shengJiNaturalRank(sjiCard(CardDesignJoker, 1)) != 0 {
		t.Error("a joker has no natural rank")
	}
	// A は 14 として数える。
	if shengJiNaturalRank(sjiCard(CardDesignSpade, 1)) != 14 {
		t.Error("an ace ranks 14")
	}

	s := sjiFresh(t, level, trump)
	s.SetTeamLevelForTest(-1, 9)
	s.SetTeamLevelForTest(ShengJiTeamCnt, 9)
	for i := range ShengJiTeamCnt {
		if s.GetTeamLevel(i) == 9 {
			t.Error("an out-of-range team must be ignored")
		}
	}
	s.SetHandForTest(-1, []*Card{sjiCard(CardDesignSpade, 2)})
	if p := s.GetPlayer(0); p.GetCardsSize() != 0 {
		t.Error("an out-of-range seat must be ignored")
	}
	if p := s.GetPlayer(0); p.GetTeam(2) != ShengJiTeamOf(2) {
		t.Error("GetTeam must agree with ShengJiTeamOf")
	}
}

// **局が終わったら手番は人間に戻らない。**戻ると精算画面から抜けられる。
func TestShengJiIsHumanTurnStopsAtTheEnd(t *testing.T) {
	s := sjiFresh(t, 5, CardDesignSpade)
	if !s.IsHumanTurn() {
		t.Fatal("seat 0 is the human and holds the turn")
	}
	for _, ph := range []ShengJiPhase{ShengJiPhaseHandEnd, ShengJiPhaseGameEnd} {
		s.SetPhaseForTest(ph)
		if s.IsHumanTurn() {
			t.Errorf("phase %v must not be the human's turn", ph)
		}
	}
	s.SetPhaseForTest(ShengJiPhasePlay)
	s.SetCurrentPlayerForTest(1)
	if s.IsHumanTurn() {
		t.Error("seat 1 is a CPU")
	}
	// CPU も終局後は動かない。
	s.gameEndFlag = true
	before := s.GetPlayer(1).GetCardsSize()
	s.CpuPlay()
	if s.GetPlayer(1).GetCardsSize() != before {
		t.Error("the CPU must not act once the game is over")
	}
}

// **切札で切れば非切札のトリックを取れる。**
func TestShengJiTrumpTakesThePlainTrick(t *testing.T) {
	const level, trump = 5, CardDesignSpade
	s := sjiFresh(t, level, trump)
	s.SetHandForTest(0, []*Card{sjiCard(CardDesignHeart, 1)})
	s.SetHandForTest(1, []*Card{sjiCard(CardDesignSpade, 2)})
	s.SetHandForTest(2, []*Card{sjiCard(CardDesignHeart, 3)})
	s.SetHandForTest(3, []*Card{sjiCard(CardDesignHeart, 4)})
	for seat := range ShengJiPlayerCnt {
		if err := s.Play(seat, []int{0}); err != nil {
			t.Fatalf("seat %d was rejected: %v", seat, err)
		}
	}
	if got := s.GetLastTrickWinner(); got != 1 {
		t.Errorf("seat %d took the trick, want seat 1 -- the trump cut wins", got)
	}
}

// 内部ヘルパーの縮退ケース。**壊れた状態でも落ちずに何かを返すこと。**
func TestShengJiHelperDegenerateCases(t *testing.T) {
	const level, trump = 5, CardDesignSpade
	s := sjiFresh(t, level, trump)

	if got := s.trickWinnerOffset(); got != 0 {
		t.Errorf("an empty trick resolves to %d, want 0", got)
	}
	// 形として成立しない手が混ざっても、リードが勝ったままになること。
	s.trick = [][]*Card{
		{sjiCard(CardDesignHeart, 7)},
		{sjiCard(CardDesignHeart, 3), sjiCard(CardDesignDiamond, 9)},
	}
	if got := s.trickWinnerOffset(); got != 0 {
		t.Errorf("a malformed follow won at %d, want the lead to hold", got)
	}
	// リード自体が形にならないときも 0 に落ちること。
	s.trick = [][]*Card{{sjiCard(CardDesignHeart, 7), sjiCard(CardDesignDiamond, 9)}}
	if got := s.trickWinnerOffset(); got != 0 {
		t.Errorf("a malformed lead resolves to %d, want 0", got)
	}

	if got := s.weakestIdx(-1, -1); got != 0 {
		t.Errorf("an out-of-range seat resolves to %d, want 0", got)
	}
	// そのスートを 1 枚も持っていなければ 0 に落ちる。
	s.SetHandForTest(0, []*Card{sjiCard(CardDesignHeart, 7)})
	if got := s.weakestIdx(0, CardDesignDiamond); got != 0 {
		t.Errorf("a void seat resolves to %d, want 0", got)
	}
	// スートを指定すれば、そのスートのいちばん弱い札を返す。
	s.SetHandForTest(0, []*Card{
		sjiCard(CardDesignDiamond, 9), sjiCard(CardDesignHeart, 3), sjiCard(CardDesignDiamond, 4),
	})
	if got := s.weakestIdx(0, CardDesignDiamond); got != 2 {
		t.Errorf("the weakest diamond is at %d, want 2", got)
	}
}

func TestShengJiConfigValidate(t *testing.T) {
	if err := DefaultShengJiConfig().Validate(); err != nil {
		t.Errorf("the default config is invalid: %v", err)
	}
	if err := (ShengJiConfig{CpuDifficulty: 9}).Validate(); err == nil {
		t.Error("an out-of-range difficulty must be rejected")
	}
}

func TestShengJiRoundTripsThroughJSON(t *testing.T) {
	s := NewDefaultShengJi()
	s.Reset()
	for seat := range ShengJiPlayerCnt {
		_ = s.Declare(seat, ShengJiNoTrump)
	}

	data, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("marshalling failed: %v", err)
	}
	restored := NewDefaultShengJi()
	if err := restored.UnmarshalJSON(data); err != nil {
		t.Fatalf("unmarshalling failed: %v", err)
	}
	if restored.GetPhase() != s.GetPhase() {
		t.Errorf("phase %v, want %v", restored.GetPhase(), s.GetPhase())
	}
	if restored.GetKittySize() != s.GetKittySize() {
		t.Errorf("kitty %d, want %d", restored.GetKittySize(), s.GetKittySize())
	}
	if restored.GetLevel() != s.GetLevel() {
		t.Errorf("level %d, want %d", restored.GetLevel(), s.GetLevel())
	}
}

// **KV から戻る値なので範囲を検査する。**
func TestShengJiRejectsBadJSON(t *testing.T) {
	for _, tc := range []struct {
		name, body string
	}{
		{"not json", `{`},
		{"too few seats", `{"pl":[],"ph":0,"ci":0,"le":2,"lv":[2,2],"ts":0}`},
		{"unknown phase", `{"pl":[{},{},{},{}],"ph":9,"ci":0,"le":2,"lv":[2,2],"ts":0}`},
		{"bad seat", `{"pl":[{},{},{},{}],"ph":0,"ci":9,"le":2,"lv":[2,2],"ts":0}`},
		{"bad trick leader", `{"pl":[{},{},{},{}],"ph":0,"ci":0,"tl":9,"le":2,"lv":[2,2],"ts":0}`},
		{"bad declarer team", `{"pl":[{},{},{},{}],"ph":0,"ci":0,"dt":9,"le":2,"lv":[2,2],"ts":0}`},
		{"bad winner team", `{"pl":[{},{},{},{}],"ph":0,"ci":0,"wt":9,"le":2,"lv":[2,2],"ts":0}`},
		{"bad trump suit", `{"pl":[{},{},{},{}],"ph":0,"ci":0,"le":2,"lv":[2,2],"ts":9}`},
		{"bad level", `{"pl":[{},{},{},{}],"ph":0,"ci":0,"le":99,"lv":[2,2],"ts":0}`},
		{"bad team level", `{"pl":[{},{},{},{}],"ph":0,"ci":0,"le":2,"lv":[99,2],"ts":0}`},
		{"bad last trick winner", `{"pl":[{},{},{},{}],"ph":0,"ci":0,"le":2,"lv":[2,2],"ts":0,"lw":9}`},
		{"bad declaring seat", `{"pl":[{},{},{},{}],"ph":0,"ci":0,"le":2,"lv":[2,2],"ts":0,"dc":{"Seat":9,"Suit":1,"Strength":1}}`},
		{"bad declared suit", `{"pl":[{},{},{},{}],"ph":0,"ci":0,"le":2,"lv":[2,2],"ts":0,"dc":{"Seat":0,"Suit":9,"Strength":1}}`},
	} {
		s := NewDefaultShengJi()
		if err := s.UnmarshalJSON([]byte(tc.body)); err == nil {
			t.Errorf("%s: accepted, want rejected", tc.name)
		}
	}
}
