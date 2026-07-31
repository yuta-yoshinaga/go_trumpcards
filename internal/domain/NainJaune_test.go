//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"testing"
)

func njCard(design, value int) *Card { return NewCard(design, value, true) }

// njReady puts a fresh game into the play phase for the given seat.
func njReady(t *testing.T, seat int) *NainJaune {
	t.Helper()
	n := NewDefaultNainJaune()
	n.Reset()
	n.SetPhaseForTest(NainJaunePhasePlay)
	n.SetCurrentPlayerForTest(seat)
	return n
}

// setNjHand replaces a seat's hand outright.
func setNjHand(n *NainJaune, seat int, cards []*Card) {
	p := n.GetPlayer(seat)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// njAntedBoard returns a board with one player's ante on it.
func njAntedBoard() NainJauneBoard {
	var b NainJauneBoard
	b.Ante(1)
	return b
}

// TestNainJauneBoxesAreTheRightCards pins the correction that matters most:
// the issue names the wrong suits and misses the king entirely.
func TestNainJauneBoxesAreTheRightCards(t *testing.T) {
	if NainJauneBoxCount != 5 {
		t.Fatalf("NainJauneBoxCount = %d, want 5", NainJauneBoxCount)
	}
	want := []struct {
		box    NainJauneBox
		name   string
		design int
		value  int
	}{
		{NainJauneBoxTen, "ten", CardDesignDiamond, 10},
		{NainJauneBoxJack, "jack", CardDesignClover, 11},
		{NainJauneBoxQueen, "queen", CardDesignSpade, 12},
		{NainJauneBoxKing, "king", CardDesignHeart, 13},
		{NainJauneBoxDwarf, "dwarf", CardDesignDiamond, 7},
	}
	for _, w := range want {
		if got := w.box.String(); got != w.name {
			t.Errorf("box %d = %q, want %q", int(w.box), got, w.name)
		}
		got, ok := NainJauneBoxForCard(njCard(w.design, w.value))
		if !ok || got != w.box {
			t.Errorf("%s: card gave %v/%v, want %v/true", w.name, got, ok, w.box)
		}
	}

	// **スート違いは取れない。**issue の「♠10」「♥Q」はここで落ちる。
	for _, c := range []*Card{
		njCard(CardDesignSpade, 10),
		njCard(CardDesignHeart, 12),
		njCard(CardDesignSpade, 7),
		njCard(CardDesignDiamond, 11),
	} {
		if _, ok := NainJauneBoxForCard(c); ok {
			t.Errorf("a card of the wrong suit must not claim a box: %v %d", c.GetDesign(), c.GetValue())
		}
	}
	if _, ok := NainJauneBoxForCard(nil); ok {
		t.Error("an empty card claims nothing")
	}
	if got := NainJauneBox(99).String(); got != "unknown" {
		t.Errorf("out-of-range box = %q, want %q", got, "unknown")
	}
}

// TestNainJauneAnteIsGraduated pins the per-box stake the issue leaves vague.
func TestNainJauneAnteIsGraduated(t *testing.T) {
	var b NainJauneBoard
	if got := b.Ante(1); got != NainJauneAnteTotal {
		t.Fatalf("Ante paid %d, want %d", got, NainJauneAnteTotal)
	}
	if NainJauneAnteTotal != 15 {
		t.Fatalf("NainJauneAnteTotal = %d, want 15", NainJauneAnteTotal)
	}
	for box, want := range map[NainJauneBox]int{
		NainJauneBoxTen: 1, NainJauneBoxJack: 2, NainJauneBoxQueen: 3,
		NainJauneBoxKing: 4, NainJauneBoxDwarf: 5,
	} {
		if got := b.Get(box); got != want {
			t.Errorf("%s = %d, want %d", box, got, want)
		}
	}

	// 人数ぶん積まれる。
	var multi NainJauneBoard
	multi.Ante(NainJaunePlayerCnt)
	if got := multi.Get(NainJauneBoxDwarf); got != 5*NainJaunePlayerCnt {
		t.Errorf("dwarf = %d, want %d", got, 5*NainJaunePlayerCnt)
	}

	// **取られなかった区画は持ち越す。**
	b.Ante(1)
	if got := b.Get(NainJauneBoxDwarf); got != 10 {
		t.Errorf("dwarf = %d, want it carried over to 10", got)
	}
	if got := b.Take(NainJauneBoxDwarf); got != 10 {
		t.Errorf("Take returned %d, want 10", got)
	}
	if got := b.Get(NainJauneBoxDwarf); got != 0 {
		t.Errorf("dwarf = %d, want it emptied", got)
	}

	b.Add(NainJauneBox(-1), 5)
	b.Add(NainJauneBoxTen, -5)
	if got := b.Take(NainJauneBox(99)); got != 0 {
		t.Errorf("Take of an unknown box = %d, want 0", got)
	}
	if got := b.Get(NainJauneBox(99)); got != 0 {
		t.Errorf("Get of an unknown box = %d, want 0", got)
	}
}

// TestNainJaunePoints pins the settlement unit: points, not cards.
func TestNainJaunePoints(t *testing.T) {
	for _, tc := range []struct {
		card *Card
		want int
	}{
		{njCard(CardDesignSpade, 1), 1},
		{njCard(CardDesignSpade, 5), 5},
		{njCard(CardDesignSpade, 10), 10},
		{njCard(CardDesignSpade, 11), 10},
		{njCard(CardDesignSpade, 12), 10},
		{njCard(CardDesignSpade, 13), 10},
		{nil, 0},
	} {
		if got := NainJaunePoints(tc.card); got != tc.want {
			t.Errorf("points = %d, want %d", got, tc.want)
		}
	}
}

func TestNainJauneDealLeavesATalon(t *testing.T) {
	n := NewDefaultNainJaune()
	n.Reset()

	for i, p := range n.GetPlayers() {
		if got := p.GetCardsSize(); got != NainJauneHandSize {
			t.Errorf("seat %d holds %d, want %d", i, got, NainJauneHandSize)
		}
	}
	// **talon がある。**52 - 12*4 = 4。
	if got := n.GetTalonCount(); got != 52-NainJauneHandSize*NainJaunePlayerCnt {
		t.Errorf("talon = %d, want %d", got, 52-NainJauneHandSize*NainJaunePlayerCnt)
	}
	if got := n.GetCurrentPlayerIdx(); got != 1 {
		t.Errorf("current = %d, want 1", got)
	}
	// 全員が 15 払っている。
	for i, p := range n.GetPlayers() {
		if got := p.GetChips(); got > 0 {
			t.Errorf("seat %d chips = %d, want it to have paid the ante", i, got)
		}
	}
}

// TestNainJauneRunIgnoresSuit is the decisive difference from Pope Joan.
func TestNainJauneRunIgnoresSuit(t *testing.T) {
	n := njReady(t, 0)
	n.SetRunRankForTest(0)
	setNjHand(n, 0, []*Card{njCard(CardDesignSpade, 5), njCard(CardDesignHeart, 9)})
	// 席 1 は**別スート**の 6 を持つ。これで続けられる。
	setNjHand(n, 1, []*Card{njCard(CardDesignClover, 6)})
	for i := 2; i < NainJaunePlayerCnt; i++ {
		setNjHand(n, i, []*Card{njCard(CardDesignDiamond, 2)})
	}

	if err := n.Play(0, 0); err != nil {
		t.Fatalf("Play ♠5: %v", err)
	}
	if got := n.GetRunRank(); got != 5 {
		t.Errorf("run rank = %d, want 5", got)
	}
	if got := n.GetCurrentPlayerIdx(); got != 1 {
		t.Fatalf("current = %d, want 1 (holds a six of another suit)", got)
	}
	if err := n.Play(1, 0); err != nil {
		t.Fatalf("Play ♣6 on ♠5: %v", err)
	}
	if got := n.GetRunRank(); got != 6 {
		t.Errorf("run rank = %d, want 6", got)
	}
}

func TestNainJauneRunRejectsAWrongRank(t *testing.T) {
	n := njReady(t, 0)
	n.SetRunRankForTest(5)
	setNjHand(n, 0, []*Card{njCard(CardDesignSpade, 7), njCard(CardDesignHeart, 6)})
	// ランクが 1 つ上でなければ拒否される。
	if err := n.Play(0, 0); err == nil {
		t.Error("expected an error playing a seven on a five")
	}
	if err := n.Play(0, 1); err != nil {
		t.Fatalf("Play a six on a five: %v", err)
	}
}

// TestNainJauneKingEndsTheRun covers what the issue assigns to a nonexistent
// "sequence" box: a king stops the run and its player leads again.
func TestNainJauneKingEndsTheRun(t *testing.T) {
	n := njReady(t, 0)
	n.SetRunRankForTest(12)
	setNjHand(n, 0, []*Card{njCard(CardDesignSpade, 13), njCard(CardDesignHeart, 2)})
	for i := 1; i < NainJaunePlayerCnt; i++ {
		setNjHand(n, i, []*Card{njCard(CardDesignClover, 4)})
	}

	if err := n.Play(0, 0); err != nil {
		t.Fatalf("Play ♠K: %v", err)
	}
	if got := n.GetRunRank(); got != 0 {
		t.Errorf("run rank = %d, want 0: a king ends the run", got)
	}
	// **出した本人**が好きな札で次を始める。
	if got := n.GetCurrentPlayerIdx(); got != 0 {
		t.Errorf("current = %d, want 0, who played the king", got)
	}
}

func TestNainJauneStopWhenNobodyCanContinue(t *testing.T) {
	n := njReady(t, 0)
	n.SetRunRankForTest(0)
	setNjHand(n, 0, []*Card{njCard(CardDesignSpade, 5), njCard(CardDesignHeart, 9)})
	// 誰も 6 を持っていない。
	for i := 1; i < NainJaunePlayerCnt; i++ {
		setNjHand(n, i, []*Card{njCard(CardDesignClover, 2)})
	}

	if err := n.Play(0, 0); err != nil {
		t.Fatalf("Play: %v", err)
	}
	if got := n.GetRunRank(); got != 0 {
		t.Errorf("run rank = %d, want 0: nobody can continue", got)
	}
	if got := n.GetCurrentPlayerIdx(); got != 0 {
		t.Errorf("current = %d, want 0, who played last", got)
	}
}

// TestNainJauneBoxesPayOnTheExactCard は、スートまで一致した札だけが区画を
// 取ることを実際の手番で確かめる。
func TestNainJauneBoxesPayOnTheExactCard(t *testing.T) {
	// ♦7 は 5 枚積まれている。
	n := njReady(t, 0)
	n.SetBoardForTest(njAntedBoard())
	n.SetRunRankForTest(6)
	setNjHand(n, 0, []*Card{njCard(CardDesignDiamond, 7), njCard(CardDesignHeart, 2)})
	for i := 1; i < NainJaunePlayerCnt; i++ {
		setNjHand(n, i, []*Card{njCard(CardDesignClover, 4)})
	}
	before := n.GetPlayer(0).GetChips()
	if err := n.Play(0, 0); err != nil {
		t.Fatalf("Play ♦7: %v", err)
	}
	if got := n.GetPlayer(0).GetChips() - before; got != 5 {
		t.Errorf("the dwarf paid %d, want 5", got)
	}
	if got := n.GetBoard().Get(NainJauneBoxDwarf); got != 0 {
		t.Errorf("dwarf = %d, want it emptied", got)
	}

	// ♠7 は同じランクでもスートが違うので払わない。
	other := njReady(t, 0)
	other.SetBoardForTest(njAntedBoard())
	other.SetRunRankForTest(6)
	setNjHand(other, 0, []*Card{njCard(CardDesignSpade, 7), njCard(CardDesignHeart, 2)})
	for i := 1; i < NainJaunePlayerCnt; i++ {
		setNjHand(other, i, []*Card{njCard(CardDesignClover, 4)})
	}
	beforeOther := other.GetPlayer(0).GetChips()
	if err := other.Play(0, 0); err != nil {
		t.Fatalf("Play ♠7: %v", err)
	}
	if got := other.GetPlayer(0).GetChips(); got != beforeOther {
		t.Errorf("an off-suit seven paid %d; it should pay nothing", got-beforeOther)
	}
	if got := other.GetBoard().Get(NainJauneBoxDwarf); got != 5 {
		t.Errorf("dwarf = %d, want it left standing", got)
	}
}

// TestNainJauneGoingOutCollectsPointsNotCards is the settlement correction.
func TestNainJauneGoingOutCollectsPointsNotCards(t *testing.T) {
	n := njReady(t, 0)
	n.SetRunRankForTest(0)
	setNjHand(n, 0, []*Card{njCard(CardDesignSpade, 2)})
	// 席 1 は 2 枚だが **K + 5 = 15 点**。枚数なら 2 のはず。
	setNjHand(n, 1, []*Card{njCard(CardDesignHeart, 13), njCard(CardDesignHeart, 5)})
	setNjHand(n, 2, []*Card{njCard(CardDesignClover, 1)})
	setNjHand(n, 3, nil)

	before1 := n.GetPlayer(1).GetChips()
	before2 := n.GetPlayer(2).GetChips()
	beforeWin := n.GetPlayer(0).GetChips()

	if err := n.Play(0, 0); err != nil {
		t.Fatalf("Play: %v", err)
	}
	if n.GetPhase() != NainJaunePhaseDealEnd {
		t.Fatalf("phase = %v, want DealEnd", n.GetPhase())
	}
	if got := n.GetPlayer(1).GetChips(); got != before1-15 {
		t.Errorf("seat 1 chips = %d, want %d (king 10 + five 5)", got, before1-15)
	}
	if got := n.GetPlayer(2).GetChips(); got != before2-1 {
		t.Errorf("seat 2 chips = %d, want %d (an ace is one)", got, before2-1)
	}
	if got := n.GetPlayer(0).GetChips(); got != beforeWin+16 {
		t.Errorf("winner chips = %d, want %d", got, beforeWin+16)
	}
	if got := n.GetDealWinner(); got != 0 {
		t.Errorf("deal winner = %d, want 0", got)
	}
}

func TestNainJaunePlayRejections(t *testing.T) {
	n := njReady(t, 0)
	setNjHand(n, 0, []*Card{njCard(CardDesignSpade, 2)})
	if err := n.Play(1, 0); err == nil {
		t.Error("expected an error playing out of turn")
	}
	if err := n.Play(0, 9); err == nil {
		t.Error("expected an error for an out-of-range hand index")
	}
	n.SetPhaseForTest(NainJaunePhaseDealEnd)
	if err := n.Play(0, 0); err == nil {
		t.Error("expected an error playing outside the play phase")
	}
}

func TestNainJauneNextDealAndGameEnd(t *testing.T) {
	n := NewDefaultNainJaune()
	n.Reset()
	if err := n.NextDeal(); err == nil {
		t.Error("expected an error while the deal is still live")
	}

	n.SetPhaseForTest(NainJaunePhaseDealEnd)
	if err := n.NextDeal(); err != nil {
		t.Fatalf("NextDeal: %v", err)
	}
	if n.GetPhase() != NainJaunePhasePlay {
		t.Errorf("phase = %v, want Play", n.GetPhase())
	}
	if got := n.GetCurrentPlayerIdx(); got != 2 {
		t.Errorf("current = %d, want 2", got)
	}

	n.SetDealNumberForTest(n.GetConfig().TargetDeals - 1)
	n.SetRunRankForTest(0)
	for i := range NainJaunePlayerCnt {
		n.GetPlayer(i).Reset()
	}
	setNjHand(n, 0, []*Card{njCard(CardDesignSpade, 2)})
	n.SetCurrentPlayerForTest(0)
	if err := n.Play(0, 0); err != nil {
		t.Fatalf("Play: %v", err)
	}
	if !n.GetGameEndFlag() {
		t.Fatal("the game should be over after the last deal")
	}
	if idx := n.GetWinnerIdx(); idx < 0 || idx >= NainJaunePlayerCnt {
		t.Errorf("winnerIdx = %d, want a real seat", idx)
	}
	if err := n.NextDeal(); err == nil {
		t.Error("expected an error dealing after the game is over")
	}
	if err := n.Play(0, 0); err == nil {
		t.Error("expected an error playing after the game is over")
	}
}

func TestNainJauneCpuDecide(t *testing.T) {
	n := njReady(t, 1)
	n.SetBoardForTest(njAntedBoard())
	n.SetRunRankForTest(6)
	// ♦7 は区画つき。ほかに ♥7 もあるが、区画のあるほうを選ぶ。
	setNjHand(n, 1, []*Card{njCard(CardDesignHeart, 7), njCard(CardDesignDiamond, 7)})
	idx := n.NainJauneCpuDecide(1)
	if idx != 1 {
		t.Fatalf("the CPU should prefer the card that claims a box, got %d", idx)
	}

	// 区画のある札が無ければ最も低い札。
	n.SetRunRankForTest(0)
	setNjHand(n, 1, []*Card{njCard(CardDesignHeart, 9), njCard(CardDesignClover, 3)})
	idx = n.NainJauneCpuDecide(1)
	if got := n.GetPlayer(1).GetCard(idx).GetValue(); got != 3 {
		t.Errorf("the CPU led a %d, want its lowest (3)", got)
	}

	if n.NainJauneCpuDecide(9) != -1 {
		t.Error("an unknown seat must not produce a move")
	}
}

// TestNainJauneCpuDrivesDealsToAnEnd checks that four CPUs always finish a deal.
func TestNainJauneCpuDrivesDealsToAnEnd(t *testing.T) {
	for trial := range 30 {
		n := NewDefaultNainJaune()
		n.Reset()
		for range 2000 {
			if n.GetPhase() != NainJaunePhasePlay {
				break
			}
			cur := n.GetCurrentPlayerIdx()
			idx := n.NainJauneCpuDecide(cur)
			if idx < 0 {
				t.Fatalf("trial %d: seat %d had nothing playable", trial, cur)
			}
			if err := n.Play(cur, idx); err != nil {
				t.Fatalf("trial %d: %v", trial, err)
			}
		}
		if ph := n.GetPhase(); ph != NainJaunePhaseDealEnd && ph != NainJaunePhaseGameEnd {
			t.Fatalf("trial %d: the deal never ended (phase = %v)", trial, ph)
		}
		if n.GetDealWinner() < 0 {
			t.Fatalf("trial %d: nobody went out", trial)
		}
	}
}

func TestNainJauneActionLog(t *testing.T) {
	n := NewDefaultNainJaune()
	n.Reset()
	log := n.GetActionLog()
	if len(log) == 0 {
		t.Fatal("the deal should be logged")
	}
	if log[0].ActionType != "deal" {
		t.Errorf("first entry = %q, want a deal", log[0].ActionType)
	}
	last := log[len(log)-1]
	if last.TurnNumber != len(log) {
		t.Errorf("TurnNumber = %d, want %d", last.TurnNumber, len(log))
	}
}

func TestNainJauneConfigValidate(t *testing.T) {
	if err := DefaultNainJauneConfig().Validate(); err != nil {
		t.Fatalf("the default config should validate: %v", err)
	}
	if err := (NainJauneConfig{CpuDifficulty: 5, TargetDeals: 5}).Validate(); err == nil {
		t.Error("expected an error for an unknown difficulty")
	}
	if err := (NainJauneConfig{TargetDeals: 0}).Validate(); err == nil {
		t.Error("expected an error for zero deals")
	}
	n := NewDefaultNainJaune()
	n.SetConfig(NainJauneConfig{TargetDeals: 3})
	if n.GetConfig().TargetDeals != 3 {
		t.Error("SetConfig/GetConfig disagree")
	}
}

func TestNainJauneAccessorBounds(t *testing.T) {
	n := NewDefaultNainJaune()
	n.Reset()
	if n.GetPlayer(-1) != nil || n.GetPlayer(NainJaunePlayerCnt) != nil {
		t.Error("GetPlayer should return nil outside the table")
	}
	if got := n.GetDealNumber(); got != 0 {
		t.Errorf("GetDealNumber = %d, want 0", got)
	}
}

func TestNainJauneJSONRoundTrip(t *testing.T) {
	n := NewDefaultNainJaune()
	n.Reset()
	data, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got NainJaune
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	// **盤のチップは持ち越しそのもの。**
	for i := range NainJauneBoxCount {
		b := NainJauneBox(i)
		if got.GetBoard().Get(b) != n.GetBoard().Get(b) {
			t.Errorf("%s = %d, want %d", b, got.GetBoard().Get(b), n.GetBoard().Get(b))
		}
	}
	if got.GetTalonCount() != n.GetTalonCount() {
		t.Errorf("talon = %d, want %d", got.GetTalonCount(), n.GetTalonCount())
	}
	if len(got.GetActionLog()) != len(n.GetActionLog()) {
		t.Errorf("action log = %d entries, want %d", len(got.GetActionLog()), len(n.GetActionLog()))
	}
}

func TestNainJauneUnmarshalRejectsGarbage(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"not json", `{`},
		{"wrong player count", `{"pl":[null,null],"cfg":{"cd":0,"td":5},"ph":0}`},
		{"bad config", `{"pl":[{},{},{},{}],"cfg":{"cd":9,"td":5},"ph":0}`},
		{"unknown phase", `{"pl":[{},{},{},{}],"cfg":{"cd":0,"td":5},"ph":9}`},
		// **runRank が範囲外だと出せる札が無くなって固まる。**
		{"run rank too high", `{"pl":[{},{},{},{}],"cfg":{"cd":0,"td":5},"ph":0,"rr":99}`},
		{"negative run rank", `{"pl":[{},{},{},{}],"cfg":{"cd":0,"td":5},"ph":0,"rr":-1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var n NainJaune
			if err := json.Unmarshal([]byte(tt.data), &n); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestNainJauneUnmarshalClampsIndices(t *testing.T) {
	var n NainJaune
	data := `{"pl":[{},{},{},{}],"cfg":{"cd":0,"td":5},"ph":0,"cur":99,"dl":-3,` +
		`"dw":42,"wi":42,"rr":0,` +
		`"aw":[null,{"Box":99,"Player":0,"Chips":1},{"Box":0,"Player":99,"Chips":1},` +
		`{"Box":4,"Player":1,"Chips":5}]}`
	if err := json.Unmarshal([]byte(data), &n); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := n.GetCurrentPlayerIdx(); got != 0 {
		t.Errorf("current = %d, want it clamped to 0", got)
	}
	if got := n.GetWinnerIdx(); got != -1 {
		t.Errorf("winnerIdx = %d, want -1", got)
	}
	if got := n.GetDealWinner(); got != -1 {
		t.Errorf("dealWinner = %d, want -1", got)
	}
	// **0 は「好きな札で始められる」という意味を持つ。**
	if got := n.GetRunRank(); got != 0 {
		t.Errorf("runRank = %d, want 0 preserved", got)
	}
	if got := len(n.GetAwards()); got != 1 {
		t.Errorf("awards = %d, want the malformed ones dropped", got)
	}
}

func TestNainJauneUnmarshalKeepsALiveRun(t *testing.T) {
	var live NainJaune
	data := `{"pl":[{},{},{},{}],"cfg":{"cd":0,"td":5},"ph":0,"rr":8}`
	if err := json.Unmarshal([]byte(data), &live); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := live.GetRunRank(); got != 8 {
		t.Errorf("runRank = %d, want 8 preserved", got)
	}
}
