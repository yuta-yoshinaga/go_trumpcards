//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"testing"
)

func ltCard(suit, value int) *Card { return NewCard(suit, value, true) }

// ltFresh returns a dealt game with every seat's hand cleared, so a test can
// place exactly the cards it cares about.
func ltFresh(t *testing.T) *Literature {
	t.Helper()
	l := NewDefaultLiterature()
	l.Reset()
	for i := range LiteraturePlayerCnt {
		l.SetHandForTest(i, nil)
	}
	l.SetCurrentPlayerForTest(0)
	return l
}

// TestLiteratureDeal covers the deal, which the issue gets right.
func TestLiteratureDeal(t *testing.T) {
	if got := len(newLiteratureDeck()); got != LiteratureDeckSize {
		t.Fatalf("the deck holds %d cards, want %d", got, LiteratureDeckSize)
	}
	// 8 組 × 6 枚 = 48 枚を 6 人に 8 枚ずつ。
	if LiteratureHalfSuitCnt*LiteratureHalfSuitSize != LiteratureDeckSize {
		t.Fatal("the half-suits must account for the whole pack")
	}
	if LiteraturePlayerCnt*LiteratureHandSize != LiteratureDeckSize {
		t.Fatal("the deal must consume the whole pack")
	}

	l := NewDefaultLiterature()
	l.Reset()
	seen := map[[2]int]bool{}
	for i := range LiteraturePlayerCnt {
		p := l.GetPlayer(i)
		if got := p.GetCardsSize(); got != LiteratureHandSize {
			t.Errorf("seat %d holds %d, want %d", i, got, LiteratureHandSize)
		}
		for j := range p.GetCardsSize() {
			c := p.GetCard(j)
			// **8 は 1 枚も無い。**
			if c.GetValue() == 8 {
				t.Fatal("the eights are removed from this pack")
			}
			k := [2]int{c.GetDesign(), c.GetValue()}
			if seen[k] {
				t.Fatalf("card %v was dealt twice", k)
			}
			seen[k] = true
		}
	}
}

// **ハーフスートは低位 2-7 と高位 9-A。**8 が抜けているので境目に札が無い。
func TestLiteratureHalfSuits(t *testing.T) {
	for half := range LiteratureHalfSuitCnt {
		cards := LiteratureHalfSuitCards(half)
		if len(cards) != LiteratureHalfSuitSize {
			t.Fatalf("half-suit %d holds %d cards, want %d", half, len(cards), LiteratureHalfSuitSize)
		}
		for _, c := range cards {
			if got := LiteratureHalfSuitOf(c); got != half {
				t.Errorf("card %v maps to half-suit %d, want %d", c, got, half)
			}
		}
	}
	// 低位と高位は別の組。
	low := LiteratureHalfSuitOf(ltCard(CardDesignSpade, 7))
	high := LiteratureHalfSuitOf(ltCard(CardDesignSpade, 9))
	if low == high {
		t.Error("the seven and the nine are in different half-suits")
	}
	// **8 はどの組にも属さない。**
	if got := LiteratureHalfSuitOf(ltCard(CardDesignSpade, 8)); got != -1 {
		t.Errorf("the eight maps to %d, want -1 — it is not in this pack", got)
	}
	if LiteratureHalfSuitOf(nil) != -1 {
		t.Error("a nil card belongs to no half-suit")
	}
	if LiteratureHalfSuitCards(-1) != nil || LiteratureHalfSuitCards(LiteratureHalfSuitCnt) != nil {
		t.Error("an out-of-range half-suit has no cards")
	}
}

// **席は交互。**味方に要求できない規則があるので飾りではない。
func TestLiteratureSeatingAlternates(t *testing.T) {
	for seat := range LiteraturePlayerCnt {
		if LiteratureTeamOf(seat) != seat%2 {
			t.Errorf("seat %d is on team %d, want alternating seats", seat, LiteratureTeamOf(seat))
		}
	}
	// 隣は必ず相手。
	for seat := range LiteraturePlayerCnt {
		next := (seat + 1) % LiteraturePlayerCnt
		if LiteratureTeamOf(seat) == LiteratureTeamOf(next) {
			t.Errorf("seats %d and %d must be opponents", seat, next)
		}
	}
	// **範囲外は -1。**Go の剰余は -1 % 2 = -1 なので素通しは危険。
	if got := LiteratureTeamOf(-1); got != -1 {
		t.Errorf("LiteratureTeamOf(-1) = %d, want -1", got)
	}
	if got := LiteratureTeamOf(LiteraturePlayerCnt); got != -1 {
		t.Errorf("LiteratureTeamOf(%d) = %d, want -1", LiteraturePlayerCnt, got)
	}
}

// TestLiteratureAskRules covers the four conditions the issue only partly states.
func TestLiteratureAskRules(t *testing.T) {
	setup := func(t *testing.T) *Literature {
		t.Helper()
		l := ltFresh(t)
		// 席 0 は ♠ 低位を 1 枚持つ。
		l.SetHandForTest(0, []*Card{ltCard(CardDesignSpade, 2)})
		l.SetHandForTest(1, []*Card{ltCard(CardDesignSpade, 3)})
		l.SetHandForTest(2, []*Card{ltCard(CardDesignSpade, 4)})
		return l
	}

	// **味方には要求できない。**席 0 と 2 は同じチーム。
	t.Run("you may only ask an opponent", func(t *testing.T) {
		l := setup(t)
		if err := l.LiteratureCanAsk(0, 2, ltCard(CardDesignSpade, 4)); err == nil {
			t.Error("asking a teammate must be refused")
		}
		if err := l.LiteratureCanAsk(0, 1, ltCard(CardDesignSpade, 3)); err != nil {
			t.Errorf("asking an opponent must be allowed: %v", err)
		}
	})

	// **自分がそのハーフスートを持っていなければ訊けない。**
	t.Run("you must hold a card of that half-suit", func(t *testing.T) {
		l := setup(t)
		// ♥ 高位は 1 枚も持っていない。
		if err := l.LiteratureCanAsk(0, 1, ltCard(CardDesignHeart, 1)); err == nil {
			t.Error("asking for a half-suit you do not hold must be refused")
		}
	})

	// **自分が持っている札は訊けない。**
	t.Run("you may not ask for a card you hold", func(t *testing.T) {
		l := setup(t)
		if err := l.LiteratureCanAsk(0, 1, ltCard(CardDesignSpade, 2)); err == nil {
			t.Error("asking for a card in your own hand must be refused")
		}
	})

	// **手札の尽きた相手には訊けない。**
	t.Run("you may not ask a player with no cards", func(t *testing.T) {
		l := setup(t)
		l.SetHandForTest(1, nil)
		if err := l.LiteratureCanAsk(0, 1, ltCard(CardDesignSpade, 3)); err == nil {
			t.Error("asking an empty-handed player must be refused")
		}
	})

	t.Run("other guards", func(t *testing.T) {
		l := setup(t)
		if err := l.LiteratureCanAsk(1, 0, ltCard(CardDesignSpade, 2)); err == nil {
			t.Error("asking out of turn must be refused")
		}
		if err := l.LiteratureCanAsk(0, 99, ltCard(CardDesignSpade, 3)); err == nil {
			t.Error("an unknown seat must be refused")
		}
		// **8 はこの卓に無い。**
		if err := l.LiteratureCanAsk(0, 1, ltCard(CardDesignSpade, 8)); err == nil {
			t.Error("a card outside the pack must be refused")
		}
		// 決着済みの組は訊けない。
		l.SetHalfSuitForTest(LiteratureHalfSuitOf(ltCard(CardDesignSpade, 3)), LiteratureHalfTeam0)
		if err := l.LiteratureCanAsk(0, 1, ltCard(CardDesignSpade, 3)); err == nil {
			t.Error("a settled half-suit must be refused")
		}
	})
}

// **的中なら手番継続、外れれば手番は要求先へ移る。**
func TestLiteratureAskOutcome(t *testing.T) {
	t.Run("a hit keeps the turn", func(t *testing.T) {
		l := ltFresh(t)
		l.SetHandForTest(0, []*Card{ltCard(CardDesignSpade, 2)})
		l.SetHandForTest(1, []*Card{ltCard(CardDesignSpade, 3), ltCard(CardDesignHeart, 2)})

		if err := l.Ask(0, 1, ltCard(CardDesignSpade, 3)); err != nil {
			t.Fatalf("Ask: %v", err)
		}
		if got := l.GetCurrentPlayerIdx(); got != 0 {
			t.Errorf("turn = %d, want the asker to keep it", got)
		}
		if got := l.GetPlayer(0).GetCardsSize(); got != 2 {
			t.Errorf("the asker holds %d, want 2", got)
		}
		if got := l.GetPlayer(1).GetCardsSize(); got != 1 {
			t.Errorf("the target holds %d, want 1", got)
		}
		if a := l.GetLastAsk(); a == nil || !a.Success {
			t.Error("the ask must be recorded as a hit")
		}
	})

	t.Run("a miss passes the turn to the target", func(t *testing.T) {
		l := ltFresh(t)
		l.SetHandForTest(0, []*Card{ltCard(CardDesignSpade, 2)})
		l.SetHandForTest(1, []*Card{ltCard(CardDesignHeart, 2)})

		if err := l.Ask(0, 1, ltCard(CardDesignSpade, 3)); err != nil {
			t.Fatalf("Ask: %v", err)
		}
		if got := l.GetCurrentPlayerIdx(); got != 1 {
			t.Errorf("turn = %d, want it to pass to the target", got)
		}
		if a := l.GetLastAsk(); a == nil || a.Success {
			t.Error("the ask must be recorded as a miss")
		}
	})

	// **要求の履歴は公開情報。**推理の材料になる。
	t.Run("asks are recorded for everyone to see", func(t *testing.T) {
		l := ltFresh(t)
		l.SetHandForTest(0, []*Card{ltCard(CardDesignSpade, 2)})
		l.SetHandForTest(1, []*Card{ltCard(CardDesignSpade, 3)})
		_ = l.Ask(0, 1, ltCard(CardDesignSpade, 3))
		if got := len(l.GetAsks()); got != 1 {
			t.Errorf("%d asks recorded, want 1", got)
		}
		if a := l.GetAsks()[0]; a.From != 0 || a.To != 1 {
			t.Errorf("the ask records %d->%d, want 0->1", a.From, a.To)
		}
	})
}

// TestLiteratureClaimHasThreeOutcomes は issue の
// 「1枚でも誤れば相手チームに渡る」が誤りであることを押さえる。
//
// **自チーム内で所在を言い間違えただけなら相手には渡らず、無効になる。**
func TestLiteratureClaimHasThreeOutcomes(t *testing.T) {
	const half = 0 // ♠ 低位
	cards := LiteratureHalfSuitCards(half)

	// 席 0 と 2 (同じチーム) に 6 枚を分ける。
	placeWithOwnTeam := func(l *Literature) {
		l.SetHandForTest(0, cards[:3])
		l.SetHandForTest(2, cards[3:])
		l.SetHandForTest(4, nil)
	}

	t.Run("all six with your team and the placement right -> you win it", func(t *testing.T) {
		l := ltFresh(t)
		placeWithOwnTeam(l)
		holders := []int{0, 0, 0, 2, 2, 2}
		if err := l.Claim(0, half, holders); err != nil {
			t.Fatalf("Claim: %v", err)
		}
		res := l.GetLastClaim()
		if res.Outcome != LiteratureClaimWon {
			t.Errorf("outcome = %v, want a win", res.Outcome)
		}
		if got := l.LiteratureTeamHalfSuits(0); got != 1 {
			t.Errorf("team 0 has %d half-suits, want 1", got)
		}
	})

	// **ここが issue の誤り。**相手には渡らない。
	t.Run("all six with your team but the placement wrong -> CANCELLED, not lost", func(t *testing.T) {
		l := ltFresh(t)
		placeWithOwnTeam(l)
		// 所在を入れ替えて申告する。
		holders := []int{2, 2, 2, 0, 0, 0}
		if err := l.Claim(0, half, holders); err != nil {
			t.Fatalf("Claim: %v", err)
		}
		res := l.GetLastClaim()
		if res.Outcome != LiteratureClaimCancelled {
			t.Errorf("outcome = %v, want it CANCELLED — misplacing within your own team does not hand it over", res.Outcome)
		}
		if res.AwardedTeam != -1 {
			t.Errorf("awarded to team %d, want nobody", res.AwardedTeam)
		}
		// **どちらのチームも取らない。**
		if got := l.LiteratureTeamHalfSuits(0); got != 0 {
			t.Errorf("team 0 has %d, want 0", got)
		}
		if got := l.LiteratureTeamHalfSuits(1); got != 0 {
			t.Errorf("team 1 has %d, want 0 — the issue says it goes to them, which is wrong", got)
		}
		if got := l.LiteratureCancelledCount(); got != 1 {
			t.Errorf("%d cancelled, want 1", got)
		}
	})

	t.Run("an opponent holds one -> they win it", func(t *testing.T) {
		l := ltFresh(t)
		l.SetHandForTest(0, cards[:3])
		l.SetHandForTest(2, cards[3:5])
		// 席 1 は相手チーム。
		l.SetHandForTest(1, cards[5:])
		holders := []int{0, 0, 0, 2, 2, 2}
		if err := l.Claim(0, half, holders); err != nil {
			t.Fatalf("Claim: %v", err)
		}
		res := l.GetLastClaim()
		if res.Outcome != LiteratureClaimLost {
			t.Errorf("outcome = %v, want the opponents to take it", res.Outcome)
		}
		if got := l.LiteratureTeamHalfSuits(1); got != 1 {
			t.Errorf("team 1 has %d, want 1", got)
		}
	})

	// 決着した組の札は場から消える。
	t.Run("a settled half-suit leaves every hand", func(t *testing.T) {
		l := ltFresh(t)
		placeWithOwnTeam(l)
		l.SetHandForTest(1, []*Card{ltCard(CardDesignHeart, 2)})
		_ = l.Claim(0, half, []int{0, 0, 0, 2, 2, 2})
		for i := range LiteraturePlayerCnt {
			p := l.GetPlayer(i)
			for j := range p.GetCardsSize() {
				if LiteratureHalfSuitOf(p.GetCard(j)) == half {
					t.Fatalf("seat %d still holds a card of the settled half-suit", i)
				}
			}
		}
	})
}

func TestLiteratureClaimGuards(t *testing.T) {
	l := ltFresh(t)
	cards := LiteratureHalfSuitCards(0)
	l.SetHandForTest(0, cards)

	if err := l.Claim(1, 0, []int{0, 0, 0, 0, 0, 0}); err == nil {
		t.Error("claiming out of turn must be refused")
	}
	if err := l.Claim(0, -1, []int{0, 0, 0, 0, 0, 0}); err == nil {
		t.Error("an out-of-range half-suit must be refused")
	}
	if err := l.Claim(0, LiteratureHalfSuitCnt, []int{0, 0, 0, 0, 0, 0}); err == nil {
		t.Error("an out-of-range half-suit must be refused")
	}
	if err := l.Claim(0, 0, []int{0, 0}); err == nil {
		t.Error("a claim must place all six cards")
	}
	// **味方以外の席には置けない。**
	if err := l.Claim(0, 0, []int{0, 0, 0, 0, 0, 1}); err == nil {
		t.Error("placing a card with an opponent must be refused")
	}
	// 決着済みの組は宣言できない。
	l.SetHalfSuitForTest(0, LiteratureHalfTeam1)
	if err := l.Claim(0, 0, []int{0, 0, 0, 0, 0, 0}); err == nil {
		t.Error("a settled half-suit must be refused")
	}
}

// TestLiteratureWinNeedsFive は issue の「4組（半分）を先に獲得」が誤りである
// ことを押さえる。
//
// **8 組の過半数なので 5 組。**4 組では相手も 4 組になり得て決着しない。
func TestLiteratureWinNeedsFive(t *testing.T) {
	if LiteratureWinThreshold != 5 {
		t.Fatalf("LiteratureWinThreshold = %d, want 5 — the issue's 4 is exactly half and decides nothing",
			LiteratureWinThreshold)
	}
	// 4 対 4 が起こり得ることを示す。
	if LiteratureHalfSuitCnt%2 != 0 {
		t.Fatal("with an even number of half-suits a half-and-half split is possible")
	}

	// 4 組では決着しない。
	l := ltFresh(t)
	for i := range 4 {
		l.SetHalfSuitForTest(i, LiteratureHalfTeam0)
	}
	l.CheckGameEndForTest()
	if l.GetGameEndFlag() {
		t.Error("four half-suits must not end the game — the opponents can still reach four")
	}

	// 5 組で決着する。
	l.SetHalfSuitForTest(4, LiteratureHalfTeam0)
	l.CheckGameEndForTest()
	if !l.GetGameEndFlag() {
		t.Fatal("five half-suits takes the game")
	}
	if got := l.GetWinnerTeam(); got != 0 {
		t.Errorf("winner = %d, want 0", got)
	}
}

// **無効があるので合計は 8 に満たないことがある。**そのときは多いほうの勝ち。
func TestLiteratureCancelledHalfSuitsDecideByMajority(t *testing.T) {
	l := ltFresh(t)
	// 4-3 で 1 組が無効。どちらも 5 に届かないが、全組が決着している。
	for i := range 4 {
		l.SetHalfSuitForTest(i, LiteratureHalfTeam0)
	}
	for i := 4; i < 7; i++ {
		l.SetHalfSuitForTest(i, LiteratureHalfTeam1)
	}
	l.SetHalfSuitForTest(7, LiteratureHalfCancelled)

	if got := l.LiteratureOpenCount(); got != 0 {
		t.Fatalf("%d half-suits are still open, want none", got)
	}
	if got := l.LiteratureCancelledCount(); got != 1 {
		t.Fatalf("%d cancelled, want 1", got)
	}
	l.CheckGameEndForTest()
	if !l.GetGameEndFlag() {
		t.Fatal("the game ends once every half-suit is settled")
	}
	if got := l.GetWinnerTeam(); got != 0 {
		t.Errorf("winner = %d, want 0 — four beats three even though neither reached five", got)
	}
}

// 無効が絡んで同数になったら勝者なし。
func TestLiteratureDrawWhenCancelledEvensItOut(t *testing.T) {
	l := ltFresh(t)
	for i := range 3 {
		l.SetHalfSuitForTest(i, LiteratureHalfTeam0)
	}
	for i := 3; i < 6; i++ {
		l.SetHalfSuitForTest(i, LiteratureHalfTeam1)
	}
	l.SetHalfSuitForTest(6, LiteratureHalfCancelled)
	l.SetHalfSuitForTest(7, LiteratureHalfCancelled)

	l.CheckGameEndForTest()
	if !l.GetGameEndFlag() {
		t.Fatal("the game ends once every half-suit is settled")
	}
	if got := l.GetWinnerTeam(); got != -1 {
		t.Errorf("winner = %d, want none — three against three with two cancelled", got)
	}
}

// **手札が尽きたら、まだ手札のある味方へ手番を渡す。**
func TestLiteratureEmptyHandPassesToATeammate(t *testing.T) {
	l := ltFresh(t)
	// 席 0 は最後の 1 枚を渡してしまう。
	l.SetHandForTest(0, []*Card{ltCard(CardDesignSpade, 2)})
	l.SetHandForTest(1, []*Card{ltCard(CardDesignSpade, 3)})
	// 味方 (席 2) には手札がある。
	l.SetHandForTest(2, []*Card{ltCard(CardDesignHeart, 2)})

	if err := l.Ask(0, 1, ltCard(CardDesignSpade, 3)); err != nil {
		t.Fatalf("Ask: %v", err)
	}
	// 席 0 は 2 枚になったので手番は残る。
	if got := l.GetCurrentPlayerIdx(); got != 0 {
		t.Fatalf("turn = %d, want 0", got)
	}

	// 手札を空にして、決着で手番が味方へ移ることを見る。
	// **味方 (席 4) には別の組の札を残しておく。**渡す先が無いと検証にならない。
	l.SetHandForTest(0, LiteratureHalfSuitCards(2)[:3])
	l.SetHandForTest(2, LiteratureHalfSuitCards(2)[3:])
	l.SetHandForTest(4, []*Card{ltCard(CardDesignDiamond, 2)})
	l.SetCurrentPlayerForTest(0)
	if err := l.Claim(0, 2, []int{0, 0, 0, 2, 2, 2}); err != nil {
		t.Fatalf("Claim: %v", err)
	}
	// 席 0 は空になったので、手札のある味方へ渡る。
	if got := l.GetCurrentPlayerIdx(); LiteratureTeamOf(got) != LiteratureTeamOf(0) {
		t.Errorf("turn went to seat %d, want a teammate", got)
	}
	if p := l.GetPlayer(l.GetCurrentPlayerIdx()); p.GetCardsSize() == 0 {
		t.Error("the turn must go to someone who still has cards")
	}
}

func TestLiteratureIsHumanTurnAndCpuPlay(t *testing.T) {
	l := NewDefaultLiterature()
	l.Reset()
	if !l.IsHumanTurn() {
		t.Error("seat 0 is the human and starts")
	}

	over := NewDefaultLiterature()
	over.Reset()
	over.SetPhaseForTest(LiteraturePhaseGameEnd)
	over.gameEndFlag = true
	if over.IsHumanTurn() {
		t.Error("a finished game is nobody's turn")
	}
	over.CpuPlay()
}

// **CPU だけでゲームを終わらせられること。**途中で止まると詰む。
//
// 全席を CPU にして `CpuPlay` そのものを回す。ここを手で組み直すと、
// 本番の経路（訊けないときは宣言する）を通らないまま「動いた」ことにできてしまう。
func TestLiteratureCpuDrivesAFullGame(t *testing.T) {
	for attempt := range 20 {
		players := make([]*LiteraturePlayer, 0, LiteraturePlayerCnt)
		for range LiteraturePlayerCnt {
			players = append(players, NewLiteraturePlayer(false))
		}
		l := NewLiterature(players, DefaultLiteratureConfig())
		l.Reset()
		for step := 0; step < 2000; step++ {
			if l.GetGameEndFlag() {
				break
			}
			l.CpuPlay()
		}
		if !l.GetGameEndFlag() {
			t.Fatalf("attempt %d: the game never finished", attempt)
		}
		a, b := l.LiteratureTeamHalfSuits(0), l.LiteratureTeamHalfSuits(1)
		// **終わり方は 2 通り。**過半数に届いた時点で打ち切るか、全組が決着するか。
		// 前者では組が余るので「合計 8」を常に要求してはいけない。
		switch {
		case a >= LiteratureWinThreshold || b >= LiteratureWinThreshold:
			// 過半数に届いた側が勝者。残りの組は結果を動かせない。
			want := 0
			if b >= LiteratureWinThreshold {
				want = 1
			}
			if got := l.GetWinnerTeam(); got != want {
				t.Fatalf("attempt %d: winner = %d, want %d", attempt, got, want)
			}
		default:
			// 過半数に届かずに終わったなら、全組が決着していなければならない。
			if got := l.LiteratureOpenCount(); got != 0 {
				t.Fatalf("attempt %d: the game ended with %d half-suits still open and nobody past %d",
					attempt, got, LiteratureWinThreshold)
			}
			total := a + b + l.LiteratureCancelledCount()
			if total != LiteratureHalfSuitCnt {
				t.Fatalf("attempt %d: %d half-suits accounted for, want %d", attempt, total, LiteratureHalfSuitCnt)
			}
		}
		// **無効があるので、獲得数の合計が 8 になるとは限らない。**
		if a+b+l.LiteratureCancelledCount()+l.LiteratureOpenCount() != LiteratureHalfSuitCnt {
			t.Fatalf("attempt %d: the half-suit states do not add up to %d", attempt, LiteratureHalfSuitCnt)
		}
	}
}

// **CPU は相手の手札を見ない。**確信できる組だけ宣言する。
func TestLiteratureCpuClaimsOnlyWhatItKnows(t *testing.T) {
	l := ltFresh(t)
	cards := LiteratureHalfSuitCards(0)
	// 席 0 が 6 枚とも持っていれば、他の情報なしで宣言できる。
	l.SetHandForTest(0, cards)
	half, holders := l.LiteratureCpuClaim(0)
	if half != 0 {
		t.Fatalf("half = %d, want 0 — the whole set is in hand", half)
	}
	for _, seat := range holders {
		if seat != 0 {
			t.Errorf("holder = %d, want 0", seat)
		}
	}

	// 3 枚しか持たず、味方の所在も分からなければ宣言しない。
	l2 := ltFresh(t)
	l2.SetHandForTest(0, cards[:3])
	l2.SetHandForTest(2, cards[3:])
	if half, _ := l2.LiteratureCpuClaim(0); half >= 0 {
		t.Error("the CPU must not claim a half-suit whose locations it cannot know")
	}

	if half, _ := l2.LiteratureCpuClaim(99); half >= 0 {
		t.Error("an unknown seat claims nothing")
	}
	if to, c := l2.LiteratureCpuAsk(99); to >= 0 || c != nil {
		t.Error("an unknown seat asks nothing")
	}
	// 手札が空なら訊けない。
	l2.SetHandForTest(0, nil)
	if _, c := l2.LiteratureCpuAsk(0); c != nil {
		t.Error("an empty hand cannot ask")
	}
}

func TestLiteratureAccessors(t *testing.T) {
	l := NewDefaultLiterature()
	l.Reset()
	if got := l.GetWinnerTeam(); got != -1 {
		t.Errorf("winner = %d, want -1", got)
	}
	if got := len(l.GetPlayers()); got != LiteraturePlayerCnt {
		t.Errorf("%d seats, want %d", got, LiteraturePlayerCnt)
	}
	if l.GetPlayer(-1) != nil || l.GetPlayer(99) != nil {
		t.Error("an out-of-range seat must be nil")
	}
	if got := len(l.GetAsks()); got != 0 {
		t.Errorf("the ask history starts empty, got %d", got)
	}
	if got := len(l.GetClaims()); got != 0 {
		t.Errorf("the claim history starts empty, got %d", got)
	}
	if l.GetLastAsk() != nil || l.GetLastClaim() != nil {
		t.Error("nothing has happened yet")
	}
	if len(l.GetActionLog()) == 0 {
		t.Error("dealing writes to the action log")
	}
	if got := l.LiteratureOpenCount(); got != LiteratureHalfSuitCnt {
		t.Errorf("%d half-suits open, want all %d", got, LiteratureHalfSuitCnt)
	}
	if got := l.GetHalfSuitState(-1); got != LiteratureHalfOpen {
		t.Errorf("an out-of-range half-suit reads %v, want open", got)
	}
	if got := l.LiteratureTeamHalfSuits(99); got != 0 {
		t.Errorf("an out-of-range team has %d", got)
	}
	if l.LiteratureHoldsHalfSuit(99, 0) {
		t.Error("an unknown seat holds nothing")
	}
	l.SetHandForTest(99, nil)
	l.SetHalfSuitForTest(99, LiteratureHalfTeam0)

	cfg := l.GetConfig()
	l.SetConfig(cfg)
	if l.GetConfig() != cfg {
		t.Error("SetConfig must take effect")
	}
	// プレイヤーのチーム判定。
	p := l.GetPlayer(0)
	if p.GetTeam(0) != p.GetTeam(2) || p.GetTeam(0) == p.GetTeam(1) {
		t.Error("seats 0/2 are teammates and 0/1 are opponents")
	}
	if literatureCardName(nil) != "-" {
		t.Error("a nil card has no name")
	}
}

// **誰の手札にも無い札は所在不明として扱う。**決着済みの組を参照したときに
// 落ちてはいけない。
func TestLiteratureOwnerOfAnAbsentCard(t *testing.T) {
	l := ltFresh(t)
	if got := l.literatureOwnerOf(ltCard(CardDesignSpade, 2)); got != -1 {
		t.Errorf("owner = %d, want -1 — nobody holds it", got)
	}
	l.SetHandForTest(3, []*Card{ltCard(CardDesignSpade, 2)})
	if got := l.literatureOwnerOf(ltCard(CardDesignSpade, 2)); got != 3 {
		t.Errorf("owner = %d, want 3", got)
	}
}

// **誰も動けなくなったら決着させる。**手番を渡す先が無いまま止まってはいけない。
func TestLiteraturePassTurnSettlesWhenNobodyHasCards(t *testing.T) {
	l := ltFresh(t)
	// 手札は全員空。決着していない組が残っていても終わらせる。
	l.passTurn()
	if !l.GetGameEndFlag() {
		t.Fatal("the game must end once nobody can act")
	}

	// 手札のある席があればそこへ渡る。
	l2 := ltFresh(t)
	l2.SetHandForTest(3, []*Card{ltCard(CardDesignSpade, 2)})
	l2.SetCurrentPlayerForTest(0)
	l2.passTurn()
	if got := l2.GetCurrentPlayerIdx(); got != 3 {
		t.Errorf("turn = %d, want the only seat with cards", got)
	}
	if l2.GetGameEndFlag() {
		t.Error("the game must not end while someone can still act")
	}
}

// フェーズと帰属のアクセサ。
func TestLiteraturePhaseAndHalfSuitState(t *testing.T) {
	l := ltFresh(t)
	if got := l.GetPhase(); got != LiteraturePhasePlay {
		t.Errorf("phase = %v, want play", got)
	}
	for half := range LiteratureHalfSuitCnt {
		if got := l.GetHalfSuitState(half); got != LiteratureHalfOpen {
			t.Errorf("half-suit %d reads %v, want open", half, got)
		}
	}
	l.SetHalfSuitForTest(0, LiteratureHalfCancelled)
	if got := l.GetHalfSuitState(0); got != LiteratureHalfCancelled {
		t.Errorf("half-suit 0 reads %v, want cancelled", got)
	}
	if got := l.GetHalfSuitState(LiteratureHalfSuitCnt); got != LiteratureHalfOpen {
		t.Errorf("an out-of-range half-suit reads %v, want open", got)
	}

	l.SetPhaseForTest(LiteraturePhaseGameEnd)
	if got := l.GetPhase(); got != LiteraturePhaseGameEnd {
		t.Errorf("phase = %v, want game end", got)
	}
}

func TestLiteratureConfigValidate(t *testing.T) {
	if err := DefaultLiteratureConfig().Validate(); err != nil {
		t.Errorf("the default config must validate: %v", err)
	}
	if err := (LiteratureConfig{CpuDifficulty: 9}).Validate(); err == nil {
		t.Error("a bad difficulty must not validate")
	}
}

func TestLiteratureRoundTripsThroughJSON(t *testing.T) {
	l := ltFresh(t)
	l.SetHandForTest(0, []*Card{ltCard(CardDesignSpade, 2)})
	l.SetHandForTest(1, []*Card{ltCard(CardDesignSpade, 3)})
	_ = l.Ask(0, 1, ltCard(CardDesignSpade, 3))

	data, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored Literature
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// **要求の履歴も往復する。**公開情報なので落とすと推理が成立しない。
	if got := len(restored.GetAsks()); got != 1 {
		t.Errorf("%d asks survived, want 1", got)
	}
	if got := restored.GetCurrentPlayerIdx(); got != l.GetCurrentPlayerIdx() {
		t.Errorf("turn = %d, want %d", got, l.GetCurrentPlayerIdx())
	}
}

// **壊れた状態を弾く。**KV から戻る値なので、範囲外のまま受け入れると詰む。
func TestLiteratureRejectsBadJSON(t *testing.T) {
	base := `"pl":[{},{},{},{},{},{}],"cf":{"cd":0},"ph":0,"ci":0,"wt":-1`
	for _, tc := range []struct {
		name string
		body string
	}{
		{"not json", `{`},
		{"wrong player count", `{"pl":[],"cf":{"cd":0},"ph":0}`},
		{"bad phase", `{` + base + `,"ph":99}`},
		{"bad current seat", `{` + base + `,"ci":9}`},
		{"bad winner team", `{` + base + `,"wt":9}`},
		{"bad half-suit state", `{` + base + `,"hs":[9,0,0,0,0,0,0,0]}`},
		{"bad claimed half-suit", `{` + base + `,"cl":[{"Player":0,"HalfSuit":99,"Outcome":0,"AwardedTeam":-1}]}`},
		{"bad claim outcome", `{` + base + `,"cl":[{"Player":0,"HalfSuit":0,"Outcome":99,"AwardedTeam":-1}]}`},
		{"bad ask seat", `{` + base + `,"as":[{"From":9,"To":0}]}`},
		{"bad config", `{"pl":[{},{},{},{},{},{}],"cf":{"cd":99},"ph":0,"ci":0,"wt":-1}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var l Literature
			if err := json.Unmarshal([]byte(tc.body), &l); err == nil {
				t.Error("must be rejected")
			}
		})
	}
}
