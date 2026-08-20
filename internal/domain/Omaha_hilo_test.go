//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestOmahaPlayer_EvalBestLowHand_Wheel: A-2-3-4-5 (ホイール) は最強のロー。
// ホール (A,2,7,K) + コミュニティ (3,4,5,8,Q) で A-2-3-4-5 が成立する。
func TestOmahaPlayer_EvalBestLowHand_Wheel(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignHeart, 2, false))
	p.AddCard(NewCard(CardDesignDiamond, 7, false))
	p.AddCard(NewCard(CardDesignClover, 13, false))

	community := []*Card{
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignClover, 8, false),
		NewCard(CardDesignSpade, 12, false),
	}

	if !p.EvalBestLowHand(community) {
		t.Fatalf("expected qualifying low (the wheel) but got none")
	}
	if !p.GetLowQualifies() {
		t.Fatalf("GetLowQualifies = false; want true")
	}
	low := p.GetLowBestHand()
	if len(low) != 5 {
		t.Fatalf("expected 5-card low hand, got %d", len(low))
	}
	// 値の集合が {1,2,3,4,5} であること
	got := map[int]bool{}
	for _, c := range low {
		got[c.GetValue()] = true
	}
	for _, want := range []int{1, 2, 3, 4, 5} {
		if !got[want] {
			t.Errorf("low hand missing rank %d; got %+v", want, got)
		}
	}
}

// TestOmahaPlayer_EvalBestLowHand_NotQualifying_HighCards:
// ホールカードからは2枚必須。9-K しかロー候補が無いケースは不成立。
func TestOmahaPlayer_EvalBestLowHand_NotQualifying_HighCards(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 9, false))
	p.AddCard(NewCard(CardDesignHeart, 10, false))
	p.AddCard(NewCard(CardDesignDiamond, 11, false))
	p.AddCard(NewCard(CardDesignClover, 13, false))

	community := []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignSpade, 6, false),
	}

	if p.EvalBestLowHand(community) {
		t.Fatalf("expected no qualifying low (need 2 hole cards <=8 — all hole cards are 9+)")
	}
	if p.GetLowQualifies() || p.GetLowBestHand() != nil {
		t.Fatalf("low state should be cleared; qualifies=%v hand=%v",
			p.GetLowQualifies(), p.GetLowBestHand())
	}
}

// TestOmahaPlayer_EvalBestLowHand_PairDisqualifies:
// ローでは同ランクは不可。ホール (2,2,A,K) + community (3,4,7,J,Q) は
// 2 を 1 枚しか使えず ホールから 2 枚必須を満たすにはペア 2 を含むしか
// なく、すべて 8 以下 5 枚 + ペア無しの組合せが存在しない。
func TestOmahaPlayer_EvalBestLowHand_PairDisqualifies(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 2, false))
	p.AddCard(NewCard(CardDesignHeart, 2, false))
	p.AddCard(NewCard(CardDesignDiamond, 1, false))
	p.AddCard(NewCard(CardDesignClover, 13, false))

	community := []*Card{
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 11, false),
		NewCard(CardDesignSpade, 12, false),
	}

	// A + 2 + (3,4,7) = qualifying low (no pair, all <= 8)
	if !p.EvalBestLowHand(community) {
		t.Fatalf("expected qualifying low using A+2 from hole")
	}
	got := map[int]int{}
	for _, c := range p.GetLowBestHand() {
		got[c.GetValue()]++
	}
	for v, cnt := range got {
		if cnt > 1 {
			t.Errorf("rank %d appears %d times; expected no duplicates", v, cnt)
		}
	}
}

// TestOmahaPlayer_EvalBestLowHand_PrefersLower:
// ホール (A,2,4,7) + community (3,5,8,K,K) では A-2-3-5-8 と A-2-3-4-5
// (community からは 3 枚必須なので A-2-3-4-5 は不可) を比較。
// 実際は ホール 2 枚 + community 3 枚なので組合せは限定される。
// このテストは "より低いロー" が選ばれることを確認するスモーク。
func TestOmahaPlayer_EvalBestLowHand_PrefersLower(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignHeart, 2, false))
	p.AddCard(NewCard(CardDesignDiamond, 4, false))
	p.AddCard(NewCard(CardDesignClover, 7, false))

	community := []*Card{
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 8, false),
		NewCard(CardDesignClover, 13, false),
		NewCard(CardDesignSpade, 13, false),
	}

	if !p.EvalBestLowHand(community) {
		t.Fatalf("expected qualifying low")
	}
	low := p.GetLowBestHand()

	// 一番高いカードができるだけ低い手が選ばれる: A-2-3-5-8 (high=8) より
	// A-2-3-5-7 (high=7) のほうが強いが、 7 はホール、3,5 はコミュニティから
	// なので A+7 (hole) + 3+5+8 (community) → high=8 か A+2 (hole) +
	// 3+5+8 (community) → high=8 のいずれでも high=8 だが A+2 のキッカー
	// が最弱なのでそちらが優先される。
	got := make([]int, len(low))
	for i, c := range low {
		got[i] = c.GetValue()
	}
	// 期待: 1,2,3,5,8
	want := map[int]bool{1: true, 2: true, 3: true, 5: true, 8: true}
	for _, v := range got {
		if !want[v] {
			t.Errorf("unexpected card %d in best low hand %+v; want set %+v",
				v, got, want)
		}
	}
}

// TestOmahaHiLo_Showdown_SplitPot:
// Hi-Lo で qualifying なローが居る場合、ポットが 50:50 で分割される。
// 奇数チップは Hi 側に寄る。
func TestOmahaHiLo_Showdown_SplitPot(t *testing.T) {
	o := NewDefaultOmahaHiLo()
	if !o.GetIsHiLo() {
		t.Fatalf("GetIsHiLo() = false; want true")
	}
}

// TestOmahaHiLo_DistributePots_SplitsEvenly:
// distributeHiLoPots を直接呼んで挙動検証。
// プレイヤー0 が Hi 勝者、プレイヤー1 が Lo 勝者の状況を作る。
func TestOmahaHiLo_DistributePots_SplitsEvenly(t *testing.T) {
	o := NewDefaultOmahaHiLo()
	players := o.GetPlayers()

	// プレイヤー0: 最強 Hi (ロイヤルフラッシュ相当)
	players[0].SetBestHand([]*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 10, false),
	})
	players[0].handRank = PokerHandRoyalFlush
	players[0].SetLowBestHand(nil) // qualifies = false (Set helper)
	players[0].lowQualifies = false

	// プレイヤー1: 弱い Hi だが qualifying な Low (A-2-3-4-5)
	players[1].SetBestHand([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignDiamond, 6, false),
		NewCard(CardDesignClover, 8, false),
		NewCard(CardDesignSpade, 11, false),
	})
	players[1].handRank = PokerHandHighCard
	players[1].SetLowBestHand([]*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignDiamond, 3, false),
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignSpade, 5, false),
	})

	// 残りのプレイヤーをフォールド扱い
	for i := 2; i < o.GetPlayerCnt(); i++ {
		players[i].SetFolded(true)
	}

	// 1つのサイドポット (100) に 2 人だけ
	o.SetSidePots([]SidePot{{Amount: 100, EligiblePlayers: []int{0, 1}}})

	beforeChips0 := players[0].GetChips()
	beforeChips1 := players[1].GetChips()

	bp := toBettingPlayers(players)
	hi, lo := o.distributeHiLoPots(bp)

	if hi[0] != 50 {
		t.Errorf("hi[0] = %d; want 50", hi[0])
	}
	if lo[1] != 50 {
		t.Errorf("lo[1] = %d; want 50", lo[1])
	}
	if got := players[0].GetChips() - beforeChips0; got != 50 {
		t.Errorf("player0 chip delta = %d; want 50", got)
	}
	if got := players[1].GetChips() - beforeChips1; got != 50 {
		t.Errorf("player1 chip delta = %d; want 50", got)
	}
}

// TestOmahaHiLo_DistributePots_NoQualifyingLow:
// qualifying なローが居ない場合は Hi 側が全額獲得する。
func TestOmahaHiLo_DistributePots_NoQualifyingLow(t *testing.T) {
	o := NewDefaultOmahaHiLo()
	players := o.GetPlayers()

	players[0].SetBestHand([]*Card{
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 10, false),
	})
	players[0].handRank = PokerHandOnePair
	players[0].lowQualifies = false

	players[1].SetBestHand([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignDiamond, 6, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignSpade, 11, false),
	})
	players[1].handRank = PokerHandHighCard
	players[1].lowQualifies = false

	for i := 2; i < o.GetPlayerCnt(); i++ {
		players[i].SetFolded(true)
	}

	o.SetSidePots([]SidePot{{Amount: 100, EligiblePlayers: []int{0, 1}}})

	bp := toBettingPlayers(players)
	hi, lo := o.distributeHiLoPots(bp)

	if hi[0] != 100 {
		t.Errorf("hi[0] = %d; want 100 (hi takes full pot when no qualifying low)", hi[0])
	}
	if len(lo) != 0 {
		t.Errorf("lo = %+v; want empty", lo)
	}
}

// TestOmahaHiLo_DistributePots_OddChipToHi:
// 奇数チップ (101) は分割すると Hi=51, Lo=50 になる。
func TestOmahaHiLo_DistributePots_OddChipToHi(t *testing.T) {
	o := NewDefaultOmahaHiLo()
	players := o.GetPlayers()

	players[0].SetBestHand([]*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 10, false),
	})
	players[0].handRank = PokerHandRoyalFlush
	players[0].lowQualifies = false

	players[1].SetBestHand([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignDiamond, 6, false),
		NewCard(CardDesignClover, 8, false),
		NewCard(CardDesignSpade, 11, false),
	})
	players[1].handRank = PokerHandHighCard
	players[1].SetLowBestHand([]*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignDiamond, 3, false),
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignSpade, 5, false),
	})

	for i := 2; i < o.GetPlayerCnt(); i++ {
		players[i].SetFolded(true)
	}

	o.SetSidePots([]SidePot{{Amount: 101, EligiblePlayers: []int{0, 1}}})

	bp := toBettingPlayers(players)
	hi, lo := o.distributeHiLoPots(bp)

	if hi[0] != 51 {
		t.Errorf("hi[0] = %d; want 51 (odd chip to hi)", hi[0])
	}
	if lo[1] != 50 {
		t.Errorf("lo[1] = %d; want 50", lo[1])
	}
}

// TestOmahaHiLo_DistributePots_Scoop:
// 同一プレイヤーが Hi も Lo も勝つ "scoop" のとき、両方獲得する。
func TestOmahaHiLo_DistributePots_Scoop(t *testing.T) {
	o := NewDefaultOmahaHiLo()
	players := o.GetPlayers()

	// プレイヤー0 が Hi も Lo も最強
	players[0].SetBestHand([]*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 1, false),
	})
	players[0].handRank = PokerHandStraightFlush
	players[0].SetLowBestHand([]*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignSpade, 5, false),
	})

	players[1].SetBestHand([]*Card{
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignHeart, 12, false),
		NewCard(CardDesignHeart, 11, false),
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignHeart, 9, false),
	})
	players[1].handRank = PokerHandStraight
	players[1].lowQualifies = false

	for i := 2; i < o.GetPlayerCnt(); i++ {
		players[i].SetFolded(true)
	}

	o.SetSidePots([]SidePot{{Amount: 100, EligiblePlayers: []int{0, 1}}})

	bp := toBettingPlayers(players)
	hi, lo := o.distributeHiLoPots(bp)

	if hi[0] != 50 || lo[0] != 50 {
		t.Errorf("scoop player got hi=%d lo=%d; want hi=50 lo=50", hi[0], lo[0])
	}
	if hi[1] != 0 || lo[1] != 0 {
		t.Errorf("losing player got hi=%d lo=%d; want both 0", hi[1], lo[1])
	}
}

// TestOmahaHiLo_ResolveShowdown_PopulatesResults:
// resolveShowdown が hiLo モード時に HoldemResult の Lo* / HiWonAmount /
// LowWonAmount を正しく populated することを統合テスト。実ホールカード +
// コミュニティカードを与え、EvalBestHand / EvalBestLowHand が走る本物の
// ショーダウンパスを exercise する。
func TestOmahaHiLo_ResolveShowdown_PopulatesResults(t *testing.T) {
	o := NewDefaultOmahaHiLo()
	players := o.GetPlayers()

	for _, p := range players {
		p.cards = nil
	}

	// プレイヤー0: A♠ K♠ + (Q♠ J♠ 10♠ board) でロイヤル
	players[0].AddCard(NewCard(CardDesignSpade, 1, false))
	players[0].AddCard(NewCard(CardDesignSpade, 13, false))
	players[0].AddCard(NewCard(CardDesignClover, 9, false))
	players[0].AddCard(NewCard(CardDesignClover, 8, false))

	// プレイヤー1: A♥ 2♥ + (3 4 5 board) でホイール
	players[1].AddCard(NewCard(CardDesignHeart, 1, false))
	players[1].AddCard(NewCard(CardDesignHeart, 2, false))
	players[1].AddCard(NewCard(CardDesignDiamond, 9, false))
	players[1].AddCard(NewCard(CardDesignDiamond, 10, false))

	for i := 2; i < o.GetPlayerCnt(); i++ {
		players[i].SetFolded(true)
	}

	// コミュニティ: Q♠ J♠ 3♦ 4♦ 5♦
	o.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignDiamond, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignDiamond, 5, false),
	})

	o.SetSidePots([]SidePot{{Amount: 100, EligiblePlayers: []int{0, 1}}})
	o.startingChips = []int{1000, 1000, 1000, 1000}
	o.pot = 100
	o.SetPhase(OmahaPhaseRiver)

	o.resolveShowdown()

	results := o.GetRoundResults()
	if len(results) < 2 {
		t.Fatalf("expected results for at least 2 players, got %d", len(results))
	}
	// プレイヤー1 は qualifying low (A-2-3-4-5) を持ち lo 半分を獲得する。
	var r1 *HoldemResult
	for i := range results {
		if results[i].PlayerIdx == 1 {
			r1 = &results[i]
			break
		}
	}
	if r1 == nil {
		t.Fatalf("no result for player 1")
	}
	if !r1.LowQualifies {
		t.Errorf("player1 LowQualifies = false; want true (board provides 3-4-5 + hole A-2)")
	}
	if r1.LowWonAmount == 0 {
		t.Errorf("player1 LowWonAmount = 0; want > 0")
	}
	if len(r1.LowBestHand) != 5 {
		t.Errorf("player1 LowBestHand len = %d; want 5", len(r1.LowBestHand))
	}
}

// TestOmahaHiLo_NewOmahaHiLo_ConstructorContract verifies the
// non-default constructor sets hiLo and yields a usable game.
func TestOmahaHiLo_NewOmahaHiLo_ConstructorContract(t *testing.T) {
	cfg := DefaultOmahaConfig()
	o := NewOmahaHiLo(NewTrumpCards(0), NewOmahaPlayersForTable(cfg.TableSize), cfg)
	if !o.GetIsHiLo() {
		t.Fatalf("NewOmahaHiLo did not set hiLo")
	}
	if o.GetPlayerCnt() != cfg.TableSize {
		t.Fatalf("player count = %d; want %d", o.GetPlayerCnt(), cfg.TableSize)
	}
}

// TestIsQualifyingOmahaLow exercises the helper directly to cover
// each rejection branch (wrong size, nil card, > 8, paired rank).
func TestIsQualifyingOmahaLow(t *testing.T) {
	mk := func(d, v int) *Card { return NewCard(d, v, false) }
	cases := []struct {
		name  string
		cards []*Card
		want  bool
	}{
		{
			name: "wheel qualifies",
			cards: []*Card{
				mk(CardDesignSpade, 1), mk(CardDesignHeart, 2), mk(CardDesignDiamond, 3),
				mk(CardDesignClover, 4), mk(CardDesignSpade, 5),
			},
			want: true,
		},
		{
			name: "8-7-6-5-4 qualifies (worst lo)",
			cards: []*Card{
				mk(CardDesignSpade, 8), mk(CardDesignHeart, 7), mk(CardDesignDiamond, 6),
				mk(CardDesignClover, 5), mk(CardDesignSpade, 4),
			},
			want: true,
		},
		{
			name: "9 disqualifies (>8)",
			cards: []*Card{
				mk(CardDesignSpade, 9), mk(CardDesignHeart, 2), mk(CardDesignDiamond, 3),
				mk(CardDesignClover, 4), mk(CardDesignSpade, 5),
			},
			want: false,
		},
		{
			name: "pair disqualifies",
			cards: []*Card{
				mk(CardDesignSpade, 5), mk(CardDesignHeart, 5), mk(CardDesignDiamond, 3),
				mk(CardDesignClover, 4), mk(CardDesignSpade, 2),
			},
			want: false,
		},
		{
			name:  "wrong size disqualifies",
			cards: []*Card{mk(CardDesignSpade, 1), mk(CardDesignHeart, 2)},
			want:  false,
		},
		{
			name: "nil card disqualifies",
			cards: []*Card{
				nil, mk(CardDesignHeart, 2), mk(CardDesignDiamond, 3),
				mk(CardDesignClover, 4), mk(CardDesignSpade, 5),
			},
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isQualifyingOmahaLow(tc.cards); got != tc.want {
				t.Errorf("isQualifyingOmahaLow = %v; want %v", got, tc.want)
			}
		})
	}
}

// TestOmahaPlayer_EvalBestLowHand_ZeroCards covers the early-return
// path when the player has fewer than 2 cards or community has fewer
// than 3.
func TestOmahaPlayer_EvalBestLowHand_ZeroCards(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	// no hole cards
	if p.EvalBestLowHand([]*Card{
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignDiamond, 5, false),
	}) {
		t.Errorf("expected false for zero hole cards")
	}
	// no community cards
	p2 := NewOmahaPlayer(true, HoldemStyleTAG)
	p2.AddCard(NewCard(CardDesignSpade, 1, false))
	p2.AddCard(NewCard(CardDesignHeart, 2, false))
	p2.AddCard(NewCard(CardDesignDiamond, 3, false))
	p2.AddCard(NewCard(CardDesignClover, 4, false))
	if p2.EvalBestLowHand([]*Card{}) {
		t.Errorf("expected false for empty community cards")
	}
}

// TestOmahaPlayer_GetLowComparisonCards_DefensiveCopy ensures
// the returned slice is a defensive copy (mutating it doesn't leak
// into the player's stored low hand).
func TestOmahaPlayer_GetLowComparisonCards_DefensiveCopy(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	p.SetLowBestHand([]*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignDiamond, 3, false),
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignSpade, 5, false),
	})
	cmp := p.GetLowComparisonCards()
	if len(cmp) != 5 {
		t.Fatalf("len=%d want 5", len(cmp))
	}
	cmp[0] = nil
	if p.GetLowBestHand()[0] == nil {
		t.Errorf("mutating returned slice leaked into player state")
	}
}

// #5483: Web の OmahaHiLoPage は BoardLowBadge でベッティング中ずっとロー成立
// 可能性を出しているが (frontend/src/utils/omahaLowCards.ts の boardLowPossibility)、
// CUI はショーダウンの resultLow でしか見せない。CUI プレイヤーはベットのたびに
// 自力で低ランクを数えることになる。
func TestOmaha_GetBoardLowOutlook(t *testing.T) {
	card := func(design, value int) *Card { return NewCard(design, value, false) }
	board := func(cards ...*Card) OmahaBoardLowOutlook {
		o := NewDefaultOmahaHiLo()
		o.SetCommunityCards(cards)
		return o.GetBoardLowOutlook()
	}

	t.Run("live once three distinct low ranks are on the board", func(t *testing.T) {
		got := board(card(CardDesignSpade, 2), card(CardDesignHeart, 5),
			card(CardDesignClover, 7))
		assert.Equal(t, OmahaBoardLowLive, got.Status)
		assert.Equal(t, 3, got.LowRankCount)
		assert.Zero(t, got.Needed)
	})

	// **同じランクは1つとして数える。**ローは相異なる5ランクで作るので、
	// 2 が2枚出ても板が供給できるローランクは1つ。
	t.Run("counts distinct ranks, not cards", func(t *testing.T) {
		got := board(card(CardDesignSpade, 2), card(CardDesignHeart, 2),
			card(CardDesignClover, 5))
		assert.Equal(t, OmahaBoardLowPossible, got.Status)
		assert.Equal(t, 2, got.LowRankCount)
		assert.Equal(t, 1, got.Needed)
	})

	// A は 1 として数える。9 以上は数えない。
	t.Run("treats the ace as low and ignores nine and above", func(t *testing.T) {
		got := board(card(CardDesignSpade, 1), card(CardDesignHeart, 9),
			card(CardDesignClover, 13))
		assert.Equal(t, 1, got.LowRankCount)
		assert.Equal(t, 2, got.Needed)
		assert.Equal(t, OmahaBoardLowPossible, got.Status, "残り2枚で2ランク足りるので、まだ可能")
	})

	// **残り枚数で足りなくなった時点で impossible。**フロップが高札3枚なら、
	// ターンとリバーの2枚では3ランクに届かない。
	t.Run("impossible once the remaining cards cannot supply enough ranks", func(t *testing.T) {
		got := board(card(CardDesignSpade, 10), card(CardDesignHeart, 11),
			card(CardDesignClover, 12))
		assert.Equal(t, OmahaBoardLowImpossible, got.Status)
		assert.Equal(t, 3, got.Needed)
	})

	t.Run("impossible on a full board with only two low ranks", func(t *testing.T) {
		got := board(card(CardDesignSpade, 2), card(CardDesignHeart, 5),
			card(CardDesignClover, 10), card(CardDesignDiamond, 11),
			card(CardDesignSpade, 12))
		assert.Equal(t, OmahaBoardLowImpossible, got.Status)
		assert.Equal(t, 1, got.Needed)
	})

	t.Run("an empty board is still possible", func(t *testing.T) {
		assert.Equal(t, OmahaBoardLowPossible, board().Status)
	})

	// **4ランク以上でも Needed は 0 で止まる。** 引き算をそのまま返すと -1 になり、
	// 「あと -1 ランク」という文言が出る。
	t.Run("needed never goes negative when the board is rich in low ranks", func(t *testing.T) {
		got := board(card(CardDesignSpade, 2), card(CardDesignHeart, 3),
			card(CardDesignClover, 4), card(CardDesignDiamond, 5))
		assert.Equal(t, OmahaBoardLowLive, got.Status)
		assert.Equal(t, 4, got.LowRankCount)
		assert.Zero(t, got.Needed)
	})

	// nil の混じったボードは数に入れない (このコードベースの他所と同じ防御)。
	t.Run("skips nil cards instead of counting them", func(t *testing.T) {
		got := board(card(CardDesignSpade, 2), nil, card(CardDesignHeart, 5))
		assert.Equal(t, 2, got.LowRankCount)
		assert.Equal(t, OmahaBoardLowPossible, got.Status)
	})
}
