//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"testing"
)

// desmocheCards は (suit, value) の並びから札を作るヘルパー。
func desmocheCards(pairs ...[2]int) []*Card {
	cards := make([]*Card, 0, len(pairs))
	for _, p := range pairs {
		cards = append(cards, NewCard(p[0], p[1], true))
	}
	return cards
}

// setDesmocheHand は player の手札を丸ごと入れ替える。
func setDesmocheHand(g *Desmoche, player int, cards []*Card) {
	p := g.GetPlayer(player)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// startDesmocheAct は player を Act フェーズの手番に据える。
func startDesmocheAct(g *Desmoche, player int) {
	g.SetCurrentPlayerForTest(player)
	g.SetPhaseForTest(DesmochePhaseAct)
}

func TestDesmocheValidateMeld(t *testing.T) {
	tests := []struct {
		name    string
		cards   []*Card
		want    DesmocheMeldKind
		wantErr bool
	}{
		{
			name:  "three of a kind is a set",
			cards: desmocheCards([2]int{CardDesignSpade, 7}, [2]int{CardDesignHeart, 7}, [2]int{CardDesignClover, 7}),
			want:  DesmocheMeldSet,
		},
		{
			name:  "four of a kind is a set",
			cards: desmocheCards([2]int{CardDesignSpade, 7}, [2]int{CardDesignHeart, 7}, [2]int{CardDesignClover, 7}, [2]int{CardDesignDiamond, 7}),
			want:  DesmocheMeldSet,
		},
		{
			name:  "same suit consecutive is a run",
			cards: desmocheCards([2]int{CardDesignHeart, 4}, [2]int{CardDesignHeart, 5}, [2]int{CardDesignHeart, 6}),
			want:  DesmocheMeldRun,
		},
		{
			name:  "unsorted run is still a run",
			cards: desmocheCards([2]int{CardDesignHeart, 6}, [2]int{CardDesignHeart, 4}, [2]int{CardDesignHeart, 5}),
			want:  DesmocheMeldRun,
		},
		{
			name:  "ace low run",
			cards: desmocheCards([2]int{CardDesignClover, 1}, [2]int{CardDesignClover, 2}, [2]int{CardDesignClover, 3}),
			want:  DesmocheMeldRun,
		},
		{
			name:  "ace high run",
			cards: desmocheCards([2]int{CardDesignClover, 12}, [2]int{CardDesignClover, 13}, [2]int{CardDesignClover, 1}),
			want:  DesmocheMeldRun,
		},
		{
			// **A は上か下の一方。**K-A-2 は繋がらない。
			name:    "run cannot wrap around the ace",
			cards:   desmocheCards([2]int{CardDesignClover, 13}, [2]int{CardDesignClover, 1}, [2]int{CardDesignClover, 2}),
			wantErr: true,
		},
		{
			name:    "two cards are not enough",
			cards:   desmocheCards([2]int{CardDesignSpade, 7}, [2]int{CardDesignHeart, 7}),
			wantErr: true,
		},
		{
			name:    "mixed suits do not form a run",
			cards:   desmocheCards([2]int{CardDesignHeart, 4}, [2]int{CardDesignSpade, 5}, [2]int{CardDesignHeart, 6}),
			wantErr: true,
		},
		{
			name:    "a gap breaks the run",
			cards:   desmocheCards([2]int{CardDesignHeart, 4}, [2]int{CardDesignHeart, 5}, [2]int{CardDesignHeart, 7}),
			wantErr: true,
		},
		{
			name:    "unrelated cards are neither",
			cards:   desmocheCards([2]int{CardDesignHeart, 4}, [2]int{CardDesignSpade, 9}, [2]int{CardDesignClover, 13}),
			wantErr: true,
		},
		{
			name:    "an empty card is rejected",
			cards:   []*Card{NewCard(CardDesignHeart, 4, true), nil, NewCard(CardDesignHeart, 6, true)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind, err := DesmocheValidateMeld(tt.cards)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got kind=%v", kind)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if kind != tt.want {
				t.Errorf("kind = %v, want %v", kind, tt.want)
			}
		})
	}
}

func TestDesmocheReset(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()

	for i, p := range g.GetPlayers() {
		if got := p.GetCardsSize(); got != DesmocheHandSize {
			t.Errorf("player %d has %d cards, want %d", i, got, DesmocheHandSize)
		}
	}
	// 52 - 9*4 - 1 (捨て札) = 15
	if got := g.GetStockCount(); got != 52-DesmocheHandSize*DesmochePlayerCnt-1 {
		t.Errorf("stock = %d, want %d", got, 52-DesmocheHandSize*DesmochePlayerCnt-1)
	}
	if g.GetDiscardTop() == nil {
		t.Error("the discard pile should be seeded with one card")
	}
	if got, want := g.GetPot(), DesmocheAnte*DesmochePlayerCnt; got != want {
		t.Errorf("pot = %d, want %d", got, want)
	}
	if g.GetPhase() != DesmochePhaseDraw {
		t.Errorf("phase = %v, want Draw", g.GetPhase())
	}
	if g.GetCurrentPlayerIdx() == 0 {
		t.Error("the dealer should not lead; the player to their left does")
	}
	for i := range DesmochePlayerCnt {
		if got := g.GetScore(i); got != -DesmocheAnte {
			t.Errorf("player %d score = %d, want %d", i, got, -DesmocheAnte)
		}
	}
}

func TestDesmocheDrawFromStock(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	cur := g.GetCurrentPlayerIdx()
	before := g.GetStockCount()

	if err := g.DrawFromStock(cur); err != nil {
		t.Fatalf("DrawFromStock: %v", err)
	}
	if got := g.GetStockCount(); got != before-1 {
		t.Errorf("stock = %d, want %d", got, before-1)
	}
	if got := g.GetPlayer(cur).GetCardsSize(); got != DesmocheHandSize+1 {
		t.Errorf("hand = %d, want %d", got, DesmocheHandSize+1)
	}
	if g.GetPhase() != DesmochePhaseAct {
		t.Errorf("phase = %v, want Act", g.GetPhase())
	}
}

func TestDesmocheDrawFromDiscard(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	cur := g.GetCurrentPlayerIdx()
	top := g.GetDiscardTop()

	if err := g.DrawFromDiscard(cur); err != nil {
		t.Fatalf("DrawFromDiscard: %v", err)
	}
	if g.GetDiscardTop() != nil {
		t.Error("the discard pile should be empty after taking its only card")
	}
	last := g.GetPlayer(cur).GetCard(DesmocheHandSize)
	if last != top {
		t.Errorf("the taken card is not in hand: got %v, want %v", last, top)
	}
}

func TestDesmocheDrawRejections(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	cur := g.GetCurrentPlayerIdx()

	if err := g.DrawFromStock((cur + 1) % DesmochePlayerCnt); err == nil {
		t.Error("expected an error when it is not that player's turn")
	}
	if err := g.DrawFromStock(cur); err != nil {
		t.Fatalf("DrawFromStock: %v", err)
	}
	if err := g.DrawFromStock(cur); err == nil {
		t.Error("expected an error when drawing twice in one turn")
	}
	if err := g.DrawFromDiscard(cur); err == nil {
		t.Error("expected an error when taking the discard outside the draw step")
	}

	g.SetPhaseForTest(DesmochePhaseDraw)
	g.SetDiscardForTest(nil)
	if err := g.DrawFromDiscard(cur); err == nil {
		t.Error("expected an error when the discard pile is empty")
	}
}

// TestDesmocheEmptyStockEndsRoundWithNoWinner は**山札切れで勝者なし**を確かめる。
// これは issue #4405 が触れていない規則で、ポットが持ち越されるのが要点。
func TestDesmocheEmptyStockEndsRoundWithNoWinner(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	potBefore := g.GetPot()
	g.SetStockForTest(nil)

	if err := g.DrawFromStock(g.GetCurrentPlayerIdx()); err != nil {
		t.Fatalf("DrawFromStock: %v", err)
	}
	if g.GetPhase() != DesmochePhaseRoundEnd {
		t.Fatalf("phase = %v, want RoundEnd", g.GetPhase())
	}
	if g.GetRoundWinner() != -1 {
		t.Errorf("roundWinner = %d, want -1", g.GetRoundWinner())
	}
	if !g.IsRoundExhausted() {
		t.Error("IsRoundExhausted should report the stock ran out")
	}
	if got := g.GetPot(); got != potBefore {
		t.Errorf("pot = %d, want it carried over as %d", got, potBefore)
	}

	// 次のラウンドでは持ち越し分に新しい掛け金が積まれる。
	if err := g.NextRound(); err != nil {
		t.Fatalf("NextRound: %v", err)
	}
	if got, want := g.GetPot(), potBefore+DesmocheAnte*DesmochePlayerCnt; got != want {
		t.Errorf("pot = %d, want %d", got, want)
	}
}

func TestDesmocheMeldAndGoOut(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	potBefore := g.GetPot()
	scoreBefore := g.GetScore(0)

	// **ちょうど 10 枚**を出し切ると上がり。3 + 3 + 4。
	setDesmocheHand(g, 0, desmocheCards(
		[2]int{CardDesignSpade, 5}, [2]int{CardDesignHeart, 5}, [2]int{CardDesignClover, 5},
		[2]int{CardDesignDiamond, 9}, [2]int{CardDesignDiamond, 10}, [2]int{CardDesignDiamond, 11},
		[2]int{CardDesignHeart, 2}, [2]int{CardDesignHeart, 3}, [2]int{CardDesignHeart, 4}, [2]int{CardDesignHeart, 1},
	))
	startDesmocheAct(g, 0)

	if err := g.Meld(0, []int{0, 1, 2}); err != nil {
		t.Fatalf("Meld set: %v", err)
	}
	if got := g.MeldedCount(0); got != 3 {
		t.Fatalf("melded = %d, want 3", got)
	}
	if g.GetPhase() != DesmochePhaseAct {
		t.Fatalf("phase = %v, want Act (still mid-turn)", g.GetPhase())
	}
	if err := g.Meld(0, []int{0, 1, 2}); err != nil {
		t.Fatalf("Meld run: %v", err)
	}
	if err := g.Meld(0, []int{3, 0, 1, 2}); err != nil {
		t.Fatalf("Meld ace-low run: %v", err)
	}

	if got := g.MeldedCount(0); got != DesmocheGoOutSize {
		t.Fatalf("melded = %d, want %d", got, DesmocheGoOutSize)
	}
	if g.GetPhase() != DesmochePhaseRoundEnd {
		t.Fatalf("phase = %v, want RoundEnd", g.GetPhase())
	}
	if g.GetRoundWinner() != 0 {
		t.Errorf("roundWinner = %d, want 0", g.GetRoundWinner())
	}
	if got, want := g.GetScore(0), scoreBefore+potBefore; got != want {
		t.Errorf("score = %d, want %d", got, want)
	}
	if g.GetPot() != 0 {
		t.Errorf("pot = %d, want it emptied into the winner", g.GetPot())
	}
	if g.IsRoundExhausted() {
		t.Error("IsRoundExhausted should be false when someone went out")
	}
}

func TestDesmocheMeldRejections(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	setDesmocheHand(g, 0, desmocheCards(
		[2]int{CardDesignSpade, 5}, [2]int{CardDesignHeart, 5}, [2]int{CardDesignClover, 5},
		[2]int{CardDesignDiamond, 9},
	))

	g.SetCurrentPlayerForTest(0)
	g.SetPhaseForTest(DesmochePhaseDraw)
	if err := g.Meld(0, []int{0, 1, 2}); err == nil {
		t.Error("expected an error when melding before drawing")
	}

	startDesmocheAct(g, 0)
	if err := g.Meld(1, []int{0, 1, 2}); err == nil {
		t.Error("expected an error when it is not that player's turn")
	}
	if err := g.Meld(0, []int{0, 1, 9}); err == nil {
		t.Error("expected an error for an out-of-range index")
	}
	if err := g.Meld(0, []int{0, 1, 1}); err == nil {
		t.Error("expected an error for a repeated index")
	}
	if err := g.Meld(0, []int{0, 1, 3}); err == nil {
		t.Error("expected an error for cards that form no meld")
	}
	if got := g.GetPlayer(0).GetCardsSize(); got != 4 {
		t.Errorf("hand = %d, want it untouched at 4", got)
	}
}

func TestDesmocheLayOff(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	setDesmocheHand(g, 0, desmocheCards(
		[2]int{CardDesignSpade, 5}, [2]int{CardDesignHeart, 5}, [2]int{CardDesignClover, 5},
		[2]int{CardDesignDiamond, 5}, [2]int{CardDesignDiamond, 9},
	))
	startDesmocheAct(g, 0)
	if err := g.Meld(0, []int{0, 1, 2}); err != nil {
		t.Fatalf("Meld: %v", err)
	}

	if err := g.LayOff(0, 1, 0); err == nil {
		t.Error("expected an error laying off a card that does not fit")
	}
	if err := g.LayOff(0, 0, 5); err == nil {
		t.Error("expected an error for an unknown meld")
	}
	if err := g.LayOff(0, 9, 0); err == nil {
		t.Error("expected an error for an out-of-range hand index")
	}
	if err := g.LayOff(0, 0, 0); err != nil {
		t.Fatalf("LayOff: %v", err)
	}
	if got := g.MeldedCount(0); got != 4 {
		t.Errorf("melded = %d, want 4", got)
	}
	if got := g.GetPlayer(0).GetCardsSize(); got != 1 {
		t.Errorf("hand = %d, want 1", got)
	}
}

// TestDesmocheDesmocheMove は **desmoche 本来の意味** — 自分の場のメルドから
// 札を抜いて別のメルドへ回す手 — を確かめる。issue #4405 はこれを「上がりの
// 宣言」と取り違えている。
func TestDesmocheDesmocheMove(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	setDesmocheHand(g, 0, desmocheCards(
		// ♥4-5-6-7 のラン (4 枚) と ♠7 ♣7 ♦7 のセット
		[2]int{CardDesignHeart, 4}, [2]int{CardDesignHeart, 5}, [2]int{CardDesignHeart, 6}, [2]int{CardDesignHeart, 7},
		[2]int{CardDesignSpade, 7}, [2]int{CardDesignClover, 7}, [2]int{CardDesignDiamond, 7},
		// 手札を空にすると checkGoneOut が手番を渡してしまうので 1 枚残す。
		[2]int{CardDesignSpade, 13},
	))
	startDesmocheAct(g, 0)
	if err := g.Meld(0, []int{0, 1, 2, 3}); err != nil {
		t.Fatalf("Meld run: %v", err)
	}
	if err := g.Meld(0, []int{0, 1, 2}); err != nil {
		t.Fatalf("Meld set: %v", err)
	}

	// ♥7 をランから抜いてセットへ。ランは ♥4-5-6 で有効なまま。
	if err := g.Desmoche(0, 0, 3, 1); err != nil {
		t.Fatalf("Desmoche: %v", err)
	}
	if got := len(g.GetMelds()[0].Cards); got != 3 {
		t.Errorf("run size = %d, want 3", got)
	}
	if got := len(g.GetMelds()[1].Cards); got != 4 {
		t.Errorf("set size = %d, want 4", got)
	}
	if got := g.MeldedCount(0); got != 7 {
		t.Errorf("melded total = %d, want it unchanged at 7", got)
	}
}

func TestDesmocheDesmocheMoveRejections(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	setDesmocheHand(g, 0, desmocheCards(
		[2]int{CardDesignHeart, 4}, [2]int{CardDesignHeart, 5}, [2]int{CardDesignHeart, 6},
		[2]int{CardDesignSpade, 7}, [2]int{CardDesignClover, 7}, [2]int{CardDesignDiamond, 7},
	))
	startDesmocheAct(g, 0)
	if err := g.Meld(0, []int{0, 1, 2}); err != nil {
		t.Fatalf("Meld run: %v", err)
	}
	if err := g.Meld(0, []int{0, 1, 2}); err != nil {
		t.Fatalf("Meld set: %v", err)
	}

	// **抜いた側が 3 枚を割るので不可。**
	if err := g.Desmoche(0, 0, 0, 1); err == nil {
		t.Error("expected an error when the source meld would drop below three cards")
	}
	if err := g.Desmoche(0, 0, 0, 0); err == nil {
		t.Error("expected an error when the source and target are the same meld")
	}
	if err := g.Desmoche(0, 5, 0, 1); err == nil {
		t.Error("expected an error for an unknown source meld")
	}
	if err := g.Desmoche(0, 0, 0, 5); err == nil {
		t.Error("expected an error for an unknown target meld")
	}
	if err := g.Desmoche(0, 0, 9, 1); err == nil {
		t.Error("expected an error for an out-of-range card index")
	}

	// 他人のメルドからは抜けない。
	g.GetMelds()[0].Owner = 1
	if err := g.Desmoche(0, 0, 0, 1); err == nil {
		t.Error("expected an error taking from another player's meld")
	}
}

// TestDesmocheDesmocheMoveRejectsBadFit は、抜いた側は有効なままでも移す先に
// 合わない札は動かせないことを確かめる。
func TestDesmocheDesmocheMoveRejectsBadFit(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	setDesmocheHand(g, 0, desmocheCards(
		[2]int{CardDesignHeart, 4}, [2]int{CardDesignHeart, 5}, [2]int{CardDesignHeart, 6}, [2]int{CardDesignHeart, 7},
		[2]int{CardDesignSpade, 2}, [2]int{CardDesignClover, 2}, [2]int{CardDesignDiamond, 2},
	))
	startDesmocheAct(g, 0)
	if err := g.Meld(0, []int{0, 1, 2, 3}); err != nil {
		t.Fatalf("Meld run: %v", err)
	}
	if err := g.Meld(0, []int{0, 1, 2}); err != nil {
		t.Fatalf("Meld set: %v", err)
	}
	// ♥7 を抜くとランは有効だが、2 のセットには入らない。
	if err := g.Desmoche(0, 0, 3, 1); err == nil {
		t.Error("expected an error when the card does not fit the target meld")
	}
	if got := len(g.GetMelds()[0].Cards); got != 4 {
		t.Errorf("run size = %d, want it left at 4 after the rejection", got)
	}
}

func TestDesmocheDiscardPassesTheTurn(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	cur := g.GetCurrentPlayerIdx()
	if err := g.DrawFromStock(cur); err != nil {
		t.Fatalf("DrawFromStock: %v", err)
	}
	card := g.GetPlayer(cur).GetCard(0)

	if err := g.Discard(cur, 99); err == nil {
		t.Error("expected an error for an out-of-range discard index")
	}
	if err := g.Discard(cur, 0); err != nil {
		t.Fatalf("Discard: %v", err)
	}
	if g.GetDiscardTop() != card {
		t.Error("the discarded card should be on top of the pile")
	}
	if got, want := g.GetCurrentPlayerIdx(), (cur+1)%DesmochePlayerCnt; got != want {
		t.Errorf("current = %d, want %d", got, want)
	}
	if g.GetPhase() != DesmochePhaseDraw {
		t.Errorf("phase = %v, want Draw", g.GetPhase())
	}
}

// TestDesmocheEmptyHandBelowGoOutPassesTheTurn は、3+3+3 で手札が尽きても
// 10 枚に届かない場合に**捨てられないまま詰まらない**ことを確かめる。
func TestDesmocheEmptyHandBelowGoOutPassesTheTurn(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	setDesmocheHand(g, 0, desmocheCards(
		[2]int{CardDesignSpade, 5}, [2]int{CardDesignHeart, 5}, [2]int{CardDesignClover, 5},
	))
	startDesmocheAct(g, 0)
	if err := g.Meld(0, []int{0, 1, 2}); err != nil {
		t.Fatalf("Meld: %v", err)
	}
	if g.GetPhase() != DesmochePhaseDraw {
		t.Errorf("phase = %v, want Draw (turn passed)", g.GetPhase())
	}
	if got := g.GetCurrentPlayerIdx(); got != 1 {
		t.Errorf("current = %d, want 1", got)
	}
	if g.GetRoundWinner() != -1 {
		t.Error("nine melded cards must not count as going out")
	}
}

func TestDesmocheNextRound(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	if err := g.NextRound(); err == nil {
		t.Error("expected an error while the round is still live")
	}

	g.SetStockForTest(nil)
	if err := g.DrawFromStock(g.GetCurrentPlayerIdx()); err != nil {
		t.Fatalf("DrawFromStock: %v", err)
	}
	if err := g.NextRound(); err != nil {
		t.Fatalf("NextRound: %v", err)
	}
	if got := g.GetPhase(); got != DesmochePhaseDraw {
		t.Errorf("phase = %v, want Draw", got)
	}
	for i, p := range g.GetPlayers() {
		if got := p.GetCardsSize(); got != DesmocheHandSize {
			t.Errorf("player %d has %d cards, want %d", i, got, DesmocheHandSize)
		}
	}
	if len(g.GetMelds()) != 0 {
		t.Error("melds should be cleared between rounds")
	}
	// ディーラーが 1 つ進むので、先手も 1 つ進む。
	if got := g.GetCurrentPlayerIdx(); got != 2 {
		t.Errorf("current = %d, want 2", got)
	}
}

func TestDesmocheGameEndsAfterFiveRounds(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	g.SetRoundNumberForTest(DesmocheRounds - 1)
	g.SetStockForTest(nil)

	if err := g.DrawFromStock(g.GetCurrentPlayerIdx()); err != nil {
		t.Fatalf("DrawFromStock: %v", err)
	}
	if !g.GetGameEndFlag() {
		t.Fatal("the game should be over after the last round")
	}
	if g.GetPhase() != DesmochePhaseGameEnd {
		t.Errorf("phase = %v, want GameEnd", g.GetPhase())
	}
	if idx := g.GetWinnerIdx(); idx < 0 || idx >= DesmochePlayerCnt {
		t.Errorf("winnerIdx = %d, want a real seat", idx)
	}
	if err := g.NextRound(); err == nil {
		t.Error("expected an error dealing after the game is over")
	}
	if err := g.DrawFromStock(0); err == nil {
		t.Error("expected an error drawing after the game is over")
	}
	if err := g.Meld(0, []int{0, 1, 2}); err == nil {
		t.Error("expected an error melding after the game is over")
	}
}

// TestDesmocheWinnerIsTheBiggestPot は、最終集計で**収支が最も多い席**が勝つ
// ことを確かめる。ポーカーの役ランキングは一切使わない (issue #4405 の誤り)。
func TestDesmocheGameWinnerIsTheRichestSeat(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	g.SetRoundNumberForTest(DesmocheRounds - 1)

	// 席 2 に上がらせて、その席が最終的な勝者になることを確かめる。
	setDesmocheHand(g, 2, desmocheCards(
		[2]int{CardDesignSpade, 5}, [2]int{CardDesignHeart, 5}, [2]int{CardDesignClover, 5},
		[2]int{CardDesignDiamond, 9}, [2]int{CardDesignDiamond, 10}, [2]int{CardDesignDiamond, 11},
		[2]int{CardDesignHeart, 2}, [2]int{CardDesignHeart, 3}, [2]int{CardDesignHeart, 4}, [2]int{CardDesignHeart, 1},
	))
	startDesmocheAct(g, 2)
	if err := g.Meld(2, []int{0, 1, 2}); err != nil {
		t.Fatalf("Meld: %v", err)
	}
	if err := g.Meld(2, []int{0, 1, 2}); err != nil {
		t.Fatalf("Meld: %v", err)
	}
	if err := g.Meld(2, []int{3, 0, 1, 2}); err != nil {
		t.Fatalf("Meld: %v", err)
	}
	if !g.GetGameEndFlag() {
		t.Fatal("the game should be over")
	}
	if got := g.GetWinnerIdx(); got != 2 {
		t.Errorf("winnerIdx = %d, want 2", got)
	}
}

func TestDesmocheCpuDecide(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()

	g.SetCurrentPlayerForTest(1)
	g.SetPhaseForTest(DesmochePhaseDraw)
	if act := g.DesmocheCpuDecide(1); act.MeldIdxs != nil || act.DiscardIdx != -1 {
		t.Errorf("during the draw step the CPU should do nothing, got %+v", act)
	}

	setDesmocheHand(g, 1, desmocheCards(
		[2]int{CardDesignSpade, 5}, [2]int{CardDesignHeart, 5}, [2]int{CardDesignClover, 5},
		[2]int{CardDesignDiamond, 9},
	))
	g.SetPhaseForTest(DesmochePhaseAct)
	act := g.DesmocheCpuDecide(1)
	if len(act.MeldIdxs) != 3 {
		t.Fatalf("the CPU should have found the set of 5s, got %+v", act)
	}
	if err := g.Meld(1, act.MeldIdxs); err != nil {
		t.Fatalf("the CPU's meld was rejected: %v", err)
	}

	// メルドが尽きたら捨て札を選ぶ。
	setDesmocheHand(g, 1, desmocheCards(
		[2]int{CardDesignSpade, 5}, [2]int{CardDesignSpade, 6}, [2]int{CardDesignDiamond, 13},
	))
	g.SetPhaseForTest(DesmochePhaseAct)
	act = g.DesmocheCpuDecide(1)
	if act.MeldIdxs != nil {
		t.Fatalf("no meld is available, got %+v", act)
	}
	// **♠5-♠6 は伸びる見込みがあるので、孤立した ♦K を捨てる。**
	if act.DiscardIdx != 2 {
		t.Errorf("DiscardIdx = %d, want 2 (the unconnected king)", act.DiscardIdx)
	}
}

func TestDesmocheCpuDecideEmptyHand(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	setDesmocheHand(g, 1, nil)
	startDesmocheAct(g, 1)
	if act := g.DesmocheCpuDecide(1); act.DiscardIdx != -1 {
		t.Errorf("DiscardIdx = %d, want -1 for an empty hand", act.DiscardIdx)
	}
}

// TestDesmocheCpuDrivesRoundsToAnEnd は、CPU だけで回してもラウンドが必ず
// 終わることを確かめる。
func TestDesmocheCpuDrivesRoundsToAnEnd(t *testing.T) {
	for trial := range 50 {
		g := NewDefaultDesmoche()
		g.Reset()
		for range 5000 {
			if g.GetPhase() == DesmochePhaseRoundEnd || g.GetPhase() == DesmochePhaseGameEnd {
				break
			}
			cur := g.GetCurrentPlayerIdx()
			if g.GetPhase() == DesmochePhaseDraw {
				if err := g.DrawFromStock(cur); err != nil {
					t.Fatalf("trial %d: DrawFromStock: %v", trial, err)
				}
				continue
			}
			act := g.DesmocheCpuDecide(cur)
			if act.MeldIdxs != nil {
				if err := g.Meld(cur, act.MeldIdxs); err != nil {
					t.Fatalf("trial %d: Meld: %v", trial, err)
				}
				continue
			}
			if act.DiscardIdx < 0 {
				t.Fatalf("trial %d: the CPU had nothing to do in the act step", trial)
			}
			if err := g.Discard(cur, act.DiscardIdx); err != nil {
				t.Fatalf("trial %d: Discard: %v", trial, err)
			}
		}
		if p := g.GetPhase(); p != DesmochePhaseRoundEnd && p != DesmochePhaseGameEnd {
			t.Fatalf("trial %d: the round never ended (phase = %v)", trial, p)
		}
	}
}

func TestDesmocheActionLog(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	if len(g.GetActionLog()) == 0 {
		t.Fatal("the deal should be logged")
	}
	cur := g.GetCurrentPlayerIdx()
	if err := g.DrawFromStock(cur); err != nil {
		t.Fatalf("DrawFromStock: %v", err)
	}
	log := g.GetActionLog()
	last := log[len(log)-1]
	if last.ActionType != "draw" {
		t.Errorf("ActionType = %q, want %q", last.ActionType, "draw")
	}
	if last.PlayerIdx != cur {
		t.Errorf("PlayerIdx = %d, want %d", last.PlayerIdx, cur)
	}
	if last.TurnNumber != len(log) {
		t.Errorf("TurnNumber = %d, want %d", last.TurnNumber, len(log))
	}
}

func TestDesmocheConfigValidate(t *testing.T) {
	if err := DefaultDesmocheConfig().Validate(); err != nil {
		t.Fatalf("the default config should validate: %v", err)
	}
	if err := (DesmocheConfig{CpuDifficulty: 5}).Validate(); err == nil {
		t.Error("expected an error for an unknown difficulty")
	}
	if err := (DesmocheConfig{CpuDifficulty: -1}).Validate(); err == nil {
		t.Error("expected an error for a negative difficulty")
	}
	g := NewDefaultDesmoche()
	g.SetConfig(DesmocheConfig{CpuDifficulty: DesmocheCpuDifficultyNormal})
	if g.GetConfig().CpuDifficulty != DesmocheCpuDifficultyNormal {
		t.Error("SetConfig/GetConfig disagree")
	}
}

func TestDesmocheAccessorBounds(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	if g.GetPlayer(-1) != nil || g.GetPlayer(DesmochePlayerCnt) != nil {
		t.Error("GetPlayer should return nil outside the table")
	}
	if g.GetScore(-1) != 0 || g.GetScore(DesmochePlayerCnt) != 0 {
		t.Error("GetScore should return 0 outside the table")
	}
	if got := g.GetRoundNumber(); got != 0 {
		t.Errorf("GetRoundNumber = %d, want 0", got)
	}
}

func TestDesmocheJSONRoundTrip(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	setDesmocheHand(g, 0, desmocheCards(
		[2]int{CardDesignSpade, 5}, [2]int{CardDesignHeart, 5}, [2]int{CardDesignClover, 5},
		[2]int{CardDesignDiamond, 9},
	))
	startDesmocheAct(g, 0)
	if err := g.Meld(0, []int{0, 1, 2}); err != nil {
		t.Fatalf("Meld: %v", err)
	}

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got Desmoche
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.GetPot() != g.GetPot() {
		t.Errorf("pot = %d, want %d (the carry-over must survive a round trip)", got.GetPot(), g.GetPot())
	}
	if got.GetStockCount() != g.GetStockCount() {
		t.Errorf("stock = %d, want %d", got.GetStockCount(), g.GetStockCount())
	}
	if len(got.GetMelds()) != len(g.GetMelds()) {
		t.Errorf("melds = %d, want %d", len(got.GetMelds()), len(g.GetMelds()))
	}
	if got.GetCurrentPlayerIdx() != g.GetCurrentPlayerIdx() {
		t.Errorf("current = %d, want %d", got.GetCurrentPlayerIdx(), g.GetCurrentPlayerIdx())
	}
	if len(got.GetActionLog()) != len(g.GetActionLog()) {
		t.Errorf("action log = %d entries, want %d", len(got.GetActionLog()), len(g.GetActionLog()))
	}
}

// TestDesmocheUnmarshalRejectsGarbage は KV から戻る生バイト列を信用しないこと
// を確かめる。
func TestDesmocheUnmarshalRejectsGarbage(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"not json", `{`},
		{"wrong player count", `{"pl":[null,null],"cfg":{"cd":0},"ph":0}`},
		{"bad config", `{"pl":[{},{},{},{}],"cfg":{"cd":9},"ph":0}`},
		{"unknown phase", `{"pl":[{},{},{},{}],"cfg":{"cd":0},"ph":9}`},
		{"negative phase", `{"pl":[{},{},{},{}],"cfg":{"cd":0},"ph":-1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var g Desmoche
			if err := json.Unmarshal([]byte(tt.data), &g); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestDesmocheUnmarshalClampsIndices(t *testing.T) {
	var g Desmoche
	data := `{"pl":[{},{},{},{}],"cfg":{"cd":0},"ph":0,"cur":99,"dl":-3,` +
		`"rw":42,"wi":42,"sc":[1],` +
		`"me":[null,{"Owner":0,"Kind":0,"Cards":[]},{"Owner":99,"Kind":0,"Cards":[{},{},{}]}]}`
	if err := json.Unmarshal([]byte(data), &g); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := g.GetCurrentPlayerIdx(); got != 0 {
		t.Errorf("current = %d, want it clamped to 0", got)
	}
	if got := g.GetRoundWinner(); got != -1 {
		t.Errorf("roundWinner = %d, want -1", got)
	}
	if got := g.GetWinnerIdx(); got != -1 {
		t.Errorf("winnerIdx = %d, want -1", got)
	}
	if got := len(g.GetMelds()); got != 0 {
		t.Errorf("melds = %d, want the malformed ones dropped", got)
	}
	// scores は席数ぶん確保し直す。
	if got := g.GetScore(3); got != 0 {
		t.Errorf("GetScore(3) = %d, want 0", got)
	}
	if got := g.GetScore(0); got != 1 {
		t.Errorf("GetScore(0) = %d, want the value that was sent", got)
	}
}

// **注記が主張している規則そのもの。** 他家のメルドへ付けた札は、そのメルドの
// 持ち主のものになり、自分の上がり枚数 (10) には数えない (#5720)。
func TestDesmoche_MeldedCountIgnoresOtherOwners(t *testing.T) {
	g := NewDefaultDesmoche()
	g.Reset()
	run := func(owner int) *DesmocheMeld {
		return &DesmocheMeld{Owner: owner, Kind: DesmocheMeldRun, Cards: []*Card{
			NewCard(CardDesignSpade, 5, true),
			NewCard(CardDesignSpade, 6, true),
			NewCard(CardDesignSpade, 7, true),
		}}
	}
	g.melds = []*DesmocheMeld{run(1), run(2)}

	if got := g.MeldedCount(0); got != 0 {
		t.Errorf("MeldedCount(0) = %d, want 0 — 他家のメルドは自分の枚数に入らない", got)
	}
	// 自分のメルドはもちろん数える (負のコントロール: 常に 0 を返す実装を弾く)。
	g.melds = []*DesmocheMeld{run(0), run(1)}
	if got := g.MeldedCount(0); got != 3 {
		t.Errorf("MeldedCount(0) = %d, want 3", got)
	}
}
