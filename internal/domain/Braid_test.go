//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

// newTestBraid returns a freshly reset Braid.
func newTestBraid() *Braid {
	b := NewDefaultBraid()
	b.Reset()
	return b
}

// clearBraidBoard empties every zone so a test can build the exact position it
// needs. Reset deals a random layout, so asserting on it directly would make the
// test depend on the shuffle.
func clearBraidBoard(b *Braid) {
	b.braid = nil
	b.stock = nil
	b.waste = nil
	for i := range BraidFieldCnt {
		b.fields[i] = nil
	}
	for i := range BraidHelperCnt {
		b.helpers[i] = nil
	}
	for i := range BraidFoundationCnt {
		b.foundation[i] = nil
	}
	b.history = nil
	b.actionLog = nil
	b.isStalemate = false
}

// braidCard builds a card for tests.
func braidCard(design, value int) *Card {
	return NewCard(design, value, true)
}

// fillBraidFields occupies all four braid fields with cards that cannot reach a
// foundation. refillFields fills *every* empty field, so a test that leaves
// artificial holes has them drain the braid before the interesting one does.
func fillBraidFields(b *Braid) {
	for i := range BraidFieldCnt {
		b.fields[i] = braidCard(CardDesignHeart, 2)
	}
}

func TestBraid_Reset(t *testing.T) {
	b := newTestBraid()

	if b.GetPhase() != BraidPhasePlaying {
		t.Errorf("phase = %v, want playing", b.GetPhase())
	}
	if got := len(b.GetBraid()); got != BraidSize {
		t.Errorf("braid = %d, want %d", got, BraidSize)
	}
	fields := b.GetFields()
	for i := range BraidFieldCnt {
		if fields[i] == nil {
			t.Errorf("braid field %d is empty after reset", i)
		}
	}
	helpers := b.GetHelpers()
	for i := range BraidHelperCnt {
		if helpers[i] == nil {
			t.Errorf("helper %d is empty after reset", i)
		}
	}
	// 開始札が 1 枚だけ基礎札 0 に置かれている。
	fd := b.GetFoundation()
	if len(fd[0]) != 1 {
		t.Errorf("foundation 0 = %d cards, want 1", len(fd[0]))
	}
	if b.GetBaseRank() != fd[0][0].GetValue() {
		t.Errorf("baseRank = %d, want %d", b.GetBaseRank(), fd[0][0].GetValue())
	}
	for i := 1; i < BraidFoundationCnt; i++ {
		if len(fd[i]) != 0 {
			t.Errorf("foundation %d = %d cards, want 0", i, len(fd[i]))
		}
	}

	wantStock := BraidTotalCards - BraidSize - BraidFieldCnt - BraidHelperCnt - 1
	if b.GetStockCount() != wantStock {
		t.Errorf("stock = %d, want %d", b.GetStockCount(), wantStock)
	}
	if len(b.GetWaste()) != 0 {
		t.Errorf("waste = %d, want 0", len(b.GetWaste()))
	}
	if b.GetMoveCount() != 0 {
		t.Errorf("moveCount = %d, want 0", b.GetMoveCount())
	}
	if b.GetPassesUsed() != 0 {
		t.Errorf("passesUsed = %d, want 0", b.GetPassesUsed())
	}
	if b.GetGameEndFlag() {
		t.Error("game should not be over after reset")
	}
	if !b.AllFaceUp() {
		t.Error("AllFaceUp should be true")
	}
}

func TestBraid_ResetDealsTheWholeDoubleDeck(t *testing.T) {
	b := newTestBraid()

	total := len(b.GetBraid()) + b.GetStockCount() + len(b.GetWaste())
	fields := b.GetFields()
	for i := range BraidFieldCnt {
		if fields[i] != nil {
			total++
		}
	}
	helpers := b.GetHelpers()
	for i := range BraidHelperCnt {
		if helpers[i] != nil {
			total++
		}
	}
	for _, pile := range b.GetFoundation() {
		total += len(pile)
	}
	if total != BraidTotalCards {
		t.Errorf("dealt %d cards, want %d", total, BraidTotalCards)
	}
}

func TestBraid_AwaitsDirection(t *testing.T) {
	b := newTestBraid()

	if !b.IsAwaitingDirection() {
		t.Fatal("a fresh game should await the direction")
	}
	if b.GetDirection() != BraidDirectionUnset {
		t.Errorf("direction = %v, want unset", b.GetDirection())
	}
	// 向きが決まるまでは基礎札に触れない。
	if err := b.MoveBraidToFoundation(); err == nil {
		t.Error("moving to a foundation should fail before the direction is chosen")
	}
	if err := b.AutoComplete(); err == nil {
		t.Error("auto-complete should fail before the direction is chosen")
	}
	// ただし山札はめくれる。向きを決める前に盤面を見るのは正当な手。
	if err := b.Draw(); err != nil {
		t.Errorf("Draw before choosing the direction: %v", err)
	}
}

func TestBraid_ChooseDirection(t *testing.T) {
	tests := []struct {
		name      string
		ascending bool
		want      BraidDirection
	}{
		{"ascending", true, BraidDirectionAscending},
		{"descending", false, BraidDirectionDescending},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newTestBraid()
			if err := b.ChooseDirection(tt.ascending); err != nil {
				t.Fatalf("ChooseDirection: %v", err)
			}
			if b.GetDirection() != tt.want {
				t.Errorf("direction = %v, want %v", b.GetDirection(), tt.want)
			}
			if b.IsAwaitingDirection() {
				t.Error("should no longer await the direction")
			}
			// 二度目は通らない。
			if err := b.ChooseDirection(tt.ascending); err == nil {
				t.Error("choosing the direction twice should fail")
			}
		})
	}
}

func TestBraid_ChooseDirectionRejectedWhenNotPlaying(t *testing.T) {
	b := newTestBraid()
	b.GiveUp()
	if err := b.ChooseDirection(true); err == nil {
		t.Error("ChooseDirection should fail after give-up")
	}
}

func TestBraid_FoundationBuildsUpInSuit(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 5
	b.direction = BraidDirectionAscending
	b.foundation[0] = []*Card{braidCard(CardDesignSpade, 5)}

	tests := []struct {
		name string
		card *Card
		want bool
	}{
		{"next rank same suit", braidCard(CardDesignSpade, 6), true},
		{"next rank other suit", braidCard(CardDesignHeart, 6), false},
		{"same rank", braidCard(CardDesignSpade, 5), false},
		{"two ranks up", braidCard(CardDesignSpade, 7), false},
		{"one rank down", braidCard(CardDesignSpade, 4), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := b.canPlaceOnFoundation(tt.card, 0); got != tt.want {
				t.Errorf("canPlaceOnFoundation = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBraid_FoundationDescendingWrapsAtAce(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 1
	b.direction = BraidDirectionDescending
	b.foundation[0] = []*Card{braidCard(CardDesignHeart, 1)}

	// 降順なので A の次は K。
	if !b.canPlaceOnFoundation(braidCard(CardDesignHeart, CardValueMax), 0) {
		t.Error("K should follow A when building down")
	}
	if b.canPlaceOnFoundation(braidCard(CardDesignHeart, 2), 0) {
		t.Error("2 must not follow A when building down")
	}
}

func TestBraid_FoundationAscendingWrapsAtKing(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = CardValueMax
	b.direction = BraidDirectionAscending
	b.foundation[0] = []*Card{braidCard(CardDesignClover, CardValueMax)}

	if !b.canPlaceOnFoundation(braidCard(CardDesignClover, 1), 0) {
		t.Error("A should follow K when building up")
	}
}

func TestBraid_EmptyFoundationOnlyTakesTheBaseRank(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 9
	b.direction = BraidDirectionAscending

	if !b.canPlaceOnFoundation(braidCard(CardDesignSpade, 9), 1) {
		t.Error("the base rank should open an empty foundation")
	}
	if b.canPlaceOnFoundation(braidCard(CardDesignSpade, 10), 1) {
		t.Error("a non-base rank must not open an empty foundation")
	}
}

func TestBraid_FullFoundationRejectsMore(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 1
	b.direction = BraidDirectionAscending
	for v := 1; v <= BraidFoundationTarget; v++ {
		b.foundation[0] = append(b.foundation[0], braidCard(CardDesignSpade, v))
	}
	if b.canPlaceOnFoundation(braidCard(CardDesignSpade, 1), 0) {
		t.Error("a complete foundation must not accept a 14th card")
	}
}

func TestBraid_CanPlaceRejectsBeforeDirection(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 5
	b.direction = BraidDirectionUnset
	if b.canPlaceOnFoundation(braidCard(CardDesignSpade, 5), 0) {
		t.Error("nothing may be placed before the direction is fixed")
	}
}

func TestBraid_MoveBraidToFoundation(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 5
	b.direction = BraidDirectionAscending
	b.foundation[0] = []*Card{braidCard(CardDesignSpade, 5)}
	b.braid = []*Card{braidCard(CardDesignHeart, 2), braidCard(CardDesignSpade, 6)}

	if err := b.MoveBraidToFoundation(); err != nil {
		t.Fatalf("MoveBraidToFoundation: %v", err)
	}
	if len(b.GetBraid()) != 1 {
		t.Errorf("braid = %d, want 1", len(b.GetBraid()))
	}
	if len(b.GetFoundation()[0]) != 2 {
		t.Errorf("foundation 0 = %d, want 2", len(b.GetFoundation()[0]))
	}
	// ブレイドが減っても、ブレイド札の枠は補充されない（枠が空いた時だけ）。
	fields := b.GetFields()
	for i := range BraidFieldCnt {
		if fields[i] != nil {
			t.Errorf("braid field %d filled itself on a braid move", i)
		}
	}

	// 末尾が置けない札なら失敗する。
	if err := b.MoveBraidToFoundation(); err == nil {
		t.Error("an unplayable braid tail should be rejected")
	}
	// ブレイドが空でも失敗する。
	b.braid = nil
	if err := b.MoveBraidToFoundation(); err == nil {
		t.Error("an empty braid should be rejected")
	}
}

func TestBraid_MoveFieldToFoundationRefillsFromBraid(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 5
	b.direction = BraidDirectionAscending
	b.foundation[0] = []*Card{braidCard(CardDesignSpade, 5)}
	fillBraidFields(b)
	b.fields[2] = braidCard(CardDesignSpade, 6)
	tail := braidCard(CardDesignHeart, 12)
	b.braid = []*Card{braidCard(CardDesignClover, 3), tail}

	if err := b.MoveFieldToFoundation(2); err != nil {
		t.Fatalf("MoveFieldToFoundation: %v", err)
	}
	// 空いた枠がブレイドの末尾で埋まる。これがブレイドを掘る唯一の経路。
	if got := b.GetFields()[2]; got != tail {
		t.Errorf("braid field 2 = %v, want the braid tail", got)
	}
	if len(b.GetBraid()) != 1 {
		t.Errorf("braid = %d, want 1", len(b.GetBraid()))
	}
}

func TestBraid_MoveFieldToFoundationLeavesTheHoleWhenBraidIsEmpty(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 5
	b.direction = BraidDirectionAscending
	b.foundation[0] = []*Card{braidCard(CardDesignSpade, 5)}
	b.fields[0] = braidCard(CardDesignSpade, 6)

	if err := b.MoveFieldToFoundation(0); err != nil {
		t.Fatalf("MoveFieldToFoundation: %v", err)
	}
	if b.GetFields()[0] != nil {
		t.Error("the field should stay empty once the braid is exhausted")
	}
}

func TestBraid_MoveFieldToFoundationErrors(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 5
	b.direction = BraidDirectionAscending
	b.foundation[0] = []*Card{braidCard(CardDesignSpade, 5)}

	if err := b.MoveFieldToFoundation(-1); err == nil {
		t.Error("a negative field index should be rejected")
	}
	if err := b.MoveFieldToFoundation(BraidFieldCnt); err == nil {
		t.Error("an out-of-range field index should be rejected")
	}
	if err := b.MoveFieldToFoundation(0); err == nil {
		t.Error("an empty field should be rejected")
	}
	b.fields[0] = braidCard(CardDesignHeart, 2)
	if err := b.MoveFieldToFoundation(0); err == nil {
		t.Error("an unplayable field card should be rejected")
	}
}

func TestBraid_MoveHelperToFoundation(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 5
	b.direction = BraidDirectionAscending
	b.foundation[0] = []*Card{braidCard(CardDesignSpade, 5)}
	b.helpers[3] = braidCard(CardDesignSpade, 6)
	b.braid = []*Card{braidCard(CardDesignClover, 3)}

	if err := b.MoveHelperToFoundation(3); err != nil {
		t.Fatalf("MoveHelperToFoundation: %v", err)
	}
	// ヘルパー枠はブレイドからは埋まらない。埋められるのは捨て札だけ。
	if b.GetHelpers()[3] != nil {
		t.Error("a helper must not refill itself from the braid")
	}
	if len(b.GetBraid()) != 1 {
		t.Errorf("braid = %d, want 1 (untouched)", len(b.GetBraid()))
	}
}

func TestBraid_MoveHelperToFoundationErrors(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 5
	b.direction = BraidDirectionAscending
	b.foundation[0] = []*Card{braidCard(CardDesignSpade, 5)}

	if err := b.MoveHelperToFoundation(-1); err == nil {
		t.Error("a negative helper index should be rejected")
	}
	if err := b.MoveHelperToFoundation(BraidHelperCnt); err == nil {
		t.Error("an out-of-range helper index should be rejected")
	}
	if err := b.MoveHelperToFoundation(0); err == nil {
		t.Error("an empty helper should be rejected")
	}
	b.helpers[0] = braidCard(CardDesignHeart, 2)
	if err := b.MoveHelperToFoundation(0); err == nil {
		t.Error("an unplayable helper card should be rejected")
	}
}

func TestBraid_MoveWasteToFoundation(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 5
	b.direction = BraidDirectionAscending
	b.foundation[0] = []*Card{braidCard(CardDesignSpade, 5)}

	if err := b.MoveWasteToFoundation(); err == nil {
		t.Error("an empty waste should be rejected")
	}
	b.waste = []*Card{braidCard(CardDesignHeart, 2)}
	if err := b.MoveWasteToFoundation(); err == nil {
		t.Error("an unplayable waste card should be rejected")
	}
	b.waste = append(b.waste, braidCard(CardDesignSpade, 6))
	if err := b.MoveWasteToFoundation(); err != nil {
		t.Fatalf("MoveWasteToFoundation: %v", err)
	}
	if len(b.GetWaste()) != 1 {
		t.Errorf("waste = %d, want 1", len(b.GetWaste()))
	}
}

func TestBraid_MoveWasteToHelper(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 5
	b.direction = BraidDirectionAscending
	card := braidCard(CardDesignHeart, 2)
	b.waste = []*Card{card}

	if err := b.MoveWasteToHelper(-1); err == nil {
		t.Error("a negative helper index should be rejected")
	}
	if err := b.MoveWasteToHelper(BraidHelperCnt); err == nil {
		t.Error("an out-of-range helper index should be rejected")
	}
	if err := b.MoveWasteToHelper(0); err != nil {
		t.Fatalf("MoveWasteToHelper: %v", err)
	}
	if b.GetHelpers()[0] != card {
		t.Error("the helper should hold the waste card")
	}
	if len(b.GetWaste()) != 0 {
		t.Errorf("waste = %d, want 0", len(b.GetWaste()))
	}
	// 埋まっている枠には置けない。
	b.waste = []*Card{braidCard(CardDesignClover, 4)}
	if err := b.MoveWasteToHelper(0); err == nil {
		t.Error("an occupied helper should be rejected")
	}
	// 捨て札が空でも置けない。
	b.waste = nil
	if err := b.MoveWasteToHelper(1); err == nil {
		t.Error("an empty waste should be rejected")
	}
}

func TestBraid_MoveWasteToHelperWorksBeforeTheDirectionIsChosen(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.direction = BraidDirectionUnset
	b.waste = []*Card{braidCard(CardDesignHeart, 2)}

	// ヘルパーへの退避は基礎札に触れないので、向きが未定でも打てる。
	if err := b.MoveWasteToHelper(0); err != nil {
		t.Errorf("MoveWasteToHelper before the direction is chosen: %v", err)
	}
}

func TestBraid_Draw(t *testing.T) {
	b := newTestBraid()
	before := b.GetStockCount()

	if err := b.Draw(); err != nil {
		t.Fatalf("Draw: %v", err)
	}
	if b.GetStockCount() != before-1 {
		t.Errorf("stock = %d, want %d", b.GetStockCount(), before-1)
	}
	if len(b.GetWaste()) != 1 {
		t.Errorf("waste = %d, want 1", len(b.GetWaste()))
	}
}

func TestBraid_RedealTwiceThenStops(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.direction = BraidDirectionAscending
	b.stock = nil
	b.waste = []*Card{braidCard(CardDesignSpade, 2), braidCard(CardDesignHeart, 3)}

	// 1 回目のめくり直し。
	if !b.CanRedeal() {
		t.Fatal("the first redeal should be allowed")
	}
	if err := b.Draw(); err != nil {
		t.Fatalf("first redeal: %v", err)
	}
	if b.GetPassesUsed() != 1 {
		t.Errorf("passesUsed = %d, want 1", b.GetPassesUsed())
	}
	if b.GetStockCount() != 2 || len(b.GetWaste()) != 0 {
		t.Errorf("after redeal: stock=%d waste=%d, want 2/0", b.GetStockCount(), len(b.GetWaste()))
	}

	// 2 回目のめくり直し。
	b.stock = nil
	b.waste = []*Card{braidCard(CardDesignSpade, 2)}
	if !b.CanRedeal() {
		t.Fatal("the second redeal should be allowed")
	}
	if err := b.Draw(); err != nil {
		t.Fatalf("second redeal: %v", err)
	}
	if b.GetPassesUsed() != BraidMaxPasses-1 {
		t.Errorf("passesUsed = %d, want %d", b.GetPassesUsed(), BraidMaxPasses-1)
	}

	// 3 回目は無い。
	b.stock = nil
	b.waste = []*Card{braidCard(CardDesignSpade, 2)}
	if b.CanRedeal() {
		t.Error("only two redeals are allowed")
	}
	if err := b.Draw(); err == nil {
		t.Error("a third redeal should be rejected")
	}
}

func TestBraid_CanRedealFalseWhenStockRemains(t *testing.T) {
	b := newTestBraid()
	if b.CanRedeal() {
		t.Error("CanRedeal should be false while the stock still has cards")
	}
	b.stock = nil
	b.waste = nil
	if b.CanRedeal() {
		t.Error("CanRedeal should be false with an empty waste")
	}
}

func TestBraid_DrawRejectedWhenNotPlaying(t *testing.T) {
	b := newTestBraid()
	b.GiveUp()
	if err := b.Draw(); err == nil {
		t.Error("Draw should fail after give-up")
	}
}

func TestBraid_GameClear(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 1
	b.direction = BraidDirectionAscending

	// 8 つのうち 7 つを完成させ、最後の 1 枚だけ残す。
	designs := []int{CardDesignSpade, CardDesignHeart, CardDesignClover, CardDesignDiamond}
	for i := range BraidFoundationCnt {
		d := designs[i%len(designs)]
		last := BraidFoundationTarget
		if i == BraidFoundationCnt-1 {
			last = BraidFoundationTarget - 1
		}
		for v := 1; v <= last; v++ {
			b.foundation[i] = append(b.foundation[i], braidCard(d, v))
		}
	}
	// 最後の山は K が欠けている。同スートの K をブレイドの末尾に置く。
	lastDesign := designs[(BraidFoundationCnt-1)%len(designs)]
	b.braid = []*Card{braidCard(lastDesign, CardValueMax)}

	if err := b.MoveBraidToFoundation(); err != nil {
		t.Fatalf("MoveBraidToFoundation: %v", err)
	}
	if b.GetPhase() != BraidPhaseGameClear {
		t.Errorf("phase = %v, want game clear", b.GetPhase())
	}
	if !b.GetGameEndFlag() {
		t.Error("GetGameEndFlag should be true after a clear")
	}
}

func TestBraid_GiveUp(t *testing.T) {
	b := newTestBraid()
	b.GiveUp()
	if b.GetPhase() != BraidPhaseGameOver {
		t.Errorf("phase = %v, want game over", b.GetPhase())
	}
	logLen := len(b.GetActionLog())
	// 二度目は何もしない。
	b.GiveUp()
	if len(b.GetActionLog()) != logLen {
		t.Error("a second give-up should not append another log entry")
	}
}

func TestBraid_HintPrefersTheDirectionFirst(t *testing.T) {
	b := newTestBraid()
	h := b.GetHint()
	if h == nil || h.FromZone != "direction" {
		t.Fatalf("hint = %+v, want the direction hint", h)
	}
}

func TestBraid_HintOrder(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 5
	b.direction = BraidDirectionAscending
	b.foundation[0] = []*Card{braidCard(CardDesignSpade, 5)}

	// ブレイド札が最優先。そこがブレイドを消化する唯一の経路だから。
	b.fields[1] = braidCard(CardDesignSpade, 6)
	b.braid = []*Card{braidCard(CardDesignSpade, 6)}
	b.helpers[0] = braidCard(CardDesignSpade, 6)
	b.waste = []*Card{braidCard(CardDesignSpade, 6)}
	if h := b.GetHint(); h == nil || h.FromZone != "field" || h.FromIdx != 1 {
		t.Errorf("hint = %+v, want field 1", h)
	}

	// 次はブレイドの末尾。
	b.fields[1] = nil
	if h := b.GetHint(); h == nil || h.FromZone != "braid" {
		t.Errorf("hint = %+v, want braid", h)
	}

	// 次はヘルパー。
	b.braid = nil
	if h := b.GetHint(); h == nil || h.FromZone != "helper" || h.FromIdx != 0 {
		t.Errorf("hint = %+v, want helper 0", h)
	}

	// 次は捨て札。
	b.helpers[0] = nil
	if h := b.GetHint(); h == nil || h.FromZone != "waste" || h.ToZone != "foundation" {
		t.Errorf("hint = %+v, want waste to foundation", h)
	}
}

func TestBraid_HintFallsBackToParkingAndDrawing(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 5
	b.direction = BraidDirectionAscending
	b.foundation[0] = []*Card{braidCard(CardDesignSpade, 5)}
	// 基礎札に行けない捨て札 → 空きヘルパーへ逃がす。
	b.waste = []*Card{braidCard(CardDesignHeart, 2)}
	if h := b.GetHint(); h == nil || h.ToZone != "helper" || h.ToIdx != 0 {
		t.Errorf("hint = %+v, want waste to helper 0", h)
	}

	// ヘルパーが埋まっていれば、残るのは山札をめくる手。
	for i := range BraidHelperCnt {
		b.helpers[i] = braidCard(CardDesignHeart, 2)
	}
	b.stock = []*Card{braidCard(CardDesignClover, 4)}
	if h := b.GetHint(); h == nil || h.FromZone != "stock" {
		t.Errorf("hint = %+v, want the stock hint", h)
	}
}

func TestBraid_HintNilWhenGameEnded(t *testing.T) {
	b := newTestBraid()
	b.GiveUp()
	if h := b.GetHint(); h != nil {
		t.Errorf("hint = %+v, want nil after give-up", h)
	}
}

func TestBraid_Stalemate(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 5
	b.direction = BraidDirectionAscending
	b.foundation[0] = []*Card{braidCard(CardDesignSpade, 5)}
	// 動かせる札が無く、山札も捨て札も空。
	b.braid = []*Card{braidCard(CardDesignHeart, 2)}
	for i := range BraidHelperCnt {
		b.helpers[i] = braidCard(CardDesignHeart, 2)
	}
	b.checkStalemate()

	if !b.IsStalemate() {
		t.Error("this position should be a stalemate")
	}
	if b.GetHint() != nil {
		t.Error("a stalemate should have no hint")
	}
}

func TestBraid_AutoComplete(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 5
	b.direction = BraidDirectionAscending
	b.foundation[0] = []*Card{braidCard(CardDesignSpade, 5)}
	b.fields[0] = braidCard(CardDesignSpade, 6)
	b.braid = []*Card{braidCard(CardDesignSpade, 7)}

	if err := b.AutoComplete(); err != nil {
		t.Fatalf("AutoComplete: %v", err)
	}
	// 6 を送る → 枠がブレイドの 7 で埋まる → その 7 も送られる。
	if got := len(b.GetFoundation()[0]); got != 3 {
		t.Errorf("foundation 0 = %d, want 3", got)
	}
	if len(b.GetBraid()) != 0 {
		t.Errorf("braid = %d, want 0", len(b.GetBraid()))
	}

	// もう送れる札は無い。
	if err := b.AutoComplete(); err == nil {
		t.Error("AutoComplete should fail when nothing can move")
	}
}

func TestBraid_AutoCompleteRejectedWhenNotPlaying(t *testing.T) {
	b := newTestBraid()
	b.GiveUp()
	if err := b.AutoComplete(); err == nil {
		t.Error("AutoComplete should fail after give-up")
	}
}

func TestBraid_Undo(t *testing.T) {
	b := newTestBraid()
	if b.CanUndo() {
		t.Error("a fresh game should have nothing to undo")
	}
	if err := b.Undo(); err == nil {
		t.Error("Undo on a fresh game should fail")
	}

	stockBefore := b.GetStockCount()
	if err := b.Draw(); err != nil {
		t.Fatalf("Draw: %v", err)
	}
	if !b.CanUndo() {
		t.Fatal("CanUndo should be true after a move")
	}
	if err := b.Undo(); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	if b.GetStockCount() != stockBefore {
		t.Errorf("stock = %d, want %d", b.GetStockCount(), stockBefore)
	}
	if len(b.GetWaste()) != 0 {
		t.Errorf("waste = %d, want 0", len(b.GetWaste()))
	}
	if b.GetMoveCount() != 0 {
		t.Errorf("moveCount = %d, want 0", b.GetMoveCount())
	}
}

func TestBraid_UndoRestoresTheDirection(t *testing.T) {
	b := newTestBraid()
	if err := b.ChooseDirection(true); err != nil {
		t.Fatalf("ChooseDirection: %v", err)
	}
	if err := b.Undo(); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	if !b.IsAwaitingDirection() {
		t.Error("undoing the direction choice should reopen it")
	}
}

func TestBraid_UndoN(t *testing.T) {
	b := newTestBraid()
	if err := b.UndoN(0); err == nil {
		t.Error("UndoN(0) should fail")
	}
	if err := b.UndoN(1); err == nil {
		t.Error("UndoN beyond the history should fail")
	}
	stockBefore := b.GetStockCount()
	for range 3 {
		if err := b.Draw(); err != nil {
			t.Fatalf("Draw: %v", err)
		}
	}
	if err := b.UndoN(3); err != nil {
		t.Fatalf("UndoN: %v", err)
	}
	if b.GetStockCount() != stockBefore {
		t.Errorf("stock = %d, want %d", b.GetStockCount(), stockBefore)
	}
}

func TestBraid_UndoToEscape(t *testing.T) {
	b := newTestBraid()
	if got := b.UndoToEscape(); got != 0 {
		t.Errorf("UndoToEscape = %d, want 0 when not stuck", got)
	}

	clearBraidBoard(b)
	b.baseRank = 5
	b.direction = BraidDirectionAscending
	b.foundation[0] = []*Card{braidCard(CardDesignSpade, 5)}
	b.stock = []*Card{braidCard(CardDesignHeart, 2)}
	for i := range BraidHelperCnt {
		b.helpers[i] = braidCard(CardDesignHeart, 2)
	}
	// めくり直しを使い切っていないと、山札が空でも「戻して配り直す」手が残る。
	b.passesUsed = BraidMaxPasses - 1
	// めくると打つ手が尽きる。
	if err := b.Draw(); err != nil {
		t.Fatalf("Draw: %v", err)
	}
	if !b.IsStalemate() {
		t.Fatal("the position should be stuck after the draw")
	}
	if got := b.UndoToEscape(); got != 1 {
		t.Errorf("UndoToEscape = %d, want 1", got)
	}

	// 履歴のどこまで戻っても詰んでいるなら -1。
	b.history = nil
	if got := b.UndoToEscape(); got != -1 {
		t.Errorf("UndoToEscape = %d, want -1", got)
	}
}

func TestBraid_ActionLog(t *testing.T) {
	b := newTestBraid()
	clearBraidBoard(b)
	b.baseRank = 5
	b.stock = []*Card{braidCard(CardDesignClover, 4)}
	b.foundation[0] = []*Card{braidCard(CardDesignSpade, 5)}
	b.fields[0] = braidCard(CardDesignSpade, 6)

	if err := b.ChooseDirection(true); err != nil {
		t.Fatalf("ChooseDirection: %v", err)
	}
	if err := b.MoveFieldToFoundation(0); err != nil {
		t.Fatalf("MoveFieldToFoundation: %v", err)
	}
	if err := b.Draw(); err != nil {
		t.Fatalf("Draw: %v", err)
	}

	log := b.GetActionLog()
	if len(log) != 3 {
		t.Fatalf("log = %d entries, want 3", len(log))
	}
	wantTypes := []string{"direction", "move", "draw"}
	for i, want := range wantTypes {
		if log[i].ActionType != want {
			t.Errorf("log[%d].ActionType = %q, want %q", i, log[i].ActionType, want)
		}
	}
	if len(log[1].Cards) != 1 {
		t.Errorf("the move entry should record the card, got %d", len(log[1].Cards))
	}
}

func TestBraid_JSONRoundTrip(t *testing.T) {
	b := newTestBraid()
	if err := b.ChooseDirection(false); err != nil {
		t.Fatalf("ChooseDirection: %v", err)
	}
	if err := b.Draw(); err != nil {
		t.Fatalf("Draw: %v", err)
	}

	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	restored := NewDefaultBraid()
	if err := json.Unmarshal(data, restored); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if restored.GetDirection() != b.GetDirection() {
		t.Errorf("direction = %v, want %v", restored.GetDirection(), b.GetDirection())
	}
	if restored.GetBaseRank() != b.GetBaseRank() {
		t.Errorf("baseRank = %d, want %d", restored.GetBaseRank(), b.GetBaseRank())
	}
	if restored.GetStockCount() != b.GetStockCount() {
		t.Errorf("stock = %d, want %d", restored.GetStockCount(), b.GetStockCount())
	}
	if len(restored.GetBraid()) != len(b.GetBraid()) {
		t.Errorf("braid = %d, want %d", len(restored.GetBraid()), len(b.GetBraid()))
	}
	if restored.GetMoveCount() != b.GetMoveCount() {
		t.Errorf("moveCount = %d, want %d", restored.GetMoveCount(), b.GetMoveCount())
	}
	if len(restored.GetActionLog()) != len(b.GetActionLog()) {
		t.Errorf("log = %d, want %d", len(restored.GetActionLog()), len(b.GetActionLog()))
	}
}

func TestBraid_UndoSurvivesAJSONRoundTrip(t *testing.T) {
	b := newTestBraid()
	if err := b.ChooseDirection(true); err != nil {
		t.Fatalf("ChooseDirection: %v", err)
	}
	stockBefore := b.GetStockCount()
	if err := b.Draw(); err != nil {
		t.Fatalf("Draw: %v", err)
	}

	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	restored := NewDefaultBraid()
	if err := json.Unmarshal(data, restored); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !restored.CanUndo() {
		t.Fatal("the undo stack must survive the wire (#4478)")
	}
	// 深さだけでなく中身も戻ること。空のスナップショットが並んでいると Undo は
	// 盤面を消してしまう。
	if err := restored.Undo(); err != nil {
		t.Fatalf("Undo after the round trip: %v", err)
	}
	if restored.GetStockCount() != stockBefore {
		t.Errorf("stock after undo = %d, want %d", restored.GetStockCount(), stockBefore)
	}
	if len(restored.GetWaste()) != 0 {
		t.Errorf("waste after undo = %d, want 0", len(restored.GetWaste()))
	}
	if len(restored.GetBraid()) != BraidSize {
		t.Errorf("braid after undo = %d, want %d", len(restored.GetBraid()), BraidSize)
	}
}

func TestBraid_UnmarshalJSONRejectsInvalidState(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"malformed", `{`},
		{"negative phase", `{"ps":-1}`},
		{"phase too large", `{"ps":99}`},
		{"negative move count", `{"mc":-1}`},
		{"negative base rank", `{"bk":-1}`},
		{"base rank too large", `{"bk":99}`},
		{"negative direction", `{"dr":-1}`},
		{"direction too large", `{"dr":9}`},
		{"negative passes", `{"pu":-1}`},
		{"too many passes", `{"pu":3}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewDefaultBraid()
			if err := json.Unmarshal([]byte(tt.data), b); err == nil {
				t.Errorf("Unmarshal(%s) should fail", tt.data)
			}
		})
	}
}

func TestBraid_UnmarshalJSONRejectsOversizedArrays(t *testing.T) {
	oversized := func(field string, n int) string {
		out := `{"` + field + `":[`
		for i := range n {
			if i > 0 {
				out += ","
			}
			out += `{"d":1,"v":1,"f":true}`
		}
		return out + `]}`
	}
	tests := []struct {
		name string
		data string
	}{
		{"braid", oversized("br", BraidTotalCards+1)},
		{"stock", oversized("st", BraidTotalCards+1)},
		{"waste", oversized("ws", BraidTotalCards+1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewDefaultBraid()
			if err := json.Unmarshal([]byte(tt.data), b); err == nil {
				t.Errorf("Unmarshal with an oversized %s should fail", tt.name)
			}
		})
	}

	// 基礎札は 13 枚が上限。
	pile := `{"fd":[[`
	for i := range BraidFoundationTarget + 1 {
		if i > 0 {
			pile += ","
		}
		pile += `{"d":1,"v":1,"f":true}`
	}
	pile += `]]}`
	b := NewDefaultBraid()
	if err := json.Unmarshal([]byte(pile), b); err == nil {
		t.Error("Unmarshal with an oversized foundation should fail")
	}
}

func TestBraid_SnapshotUnmarshalRespectsMaxSliceLen(t *testing.T) {
	// スナップショット側の上限も検査する。ここを通す JSON は KV から来るので、
	// 本体と同じだけ疑ってかかる必要がある。
	oversized := func(field string) string {
		out := `{"` + field + `":[`
		for i := range braidMaxSliceLen + 1 {
			if i > 0 {
				out += ","
			}
			out += `{"d":1,"v":1,"f":true}`
		}
		return out + `]}`
	}
	for _, field := range []string{"br", "st", "ws"} {
		t.Run(field, func(t *testing.T) {
			var s braidSnapshot
			if err := json.Unmarshal([]byte(oversized(field)), &s); err == nil {
				t.Errorf("snapshot with an oversized %s should fail", field)
			}
		})
	}

	t.Run("foundation", func(t *testing.T) {
		out := `{"fd":[[`
		for i := range braidMaxSliceLen + 1 {
			if i > 0 {
				out += ","
			}
			out += `{"d":1,"v":1,"f":true}`
		}
		out += `]]}`
		var s braidSnapshot
		if err := json.Unmarshal([]byte(out), &s); err == nil {
			t.Error("snapshot with an oversized foundation should fail")
		}
	})

	t.Run("malformed", func(t *testing.T) {
		var s braidSnapshot
		if err := json.Unmarshal([]byte(`{`), &s); err == nil {
			t.Error("a malformed snapshot should fail")
		}
	})
}

func TestBraid_UnmarshalJSONRejectsOversizedLogAndHistory(t *testing.T) {
	build := func(field, elem string) string {
		out := `{"` + field + `":[`
		for i := range braidMaxSliceLen + 1 {
			if i > 0 {
				out += ","
			}
			out += elem
		}
		return out + `]}`
	}
	tests := []struct {
		name string
		data string
	}{
		{"action log", build("al", `{}`)},
		{"history", build("hi", `{}`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewDefaultBraid()
			if err := json.Unmarshal([]byte(tt.data), b); err == nil {
				t.Errorf("Unmarshal with an oversized %s should fail", tt.name)
			}
		})
	}
}
