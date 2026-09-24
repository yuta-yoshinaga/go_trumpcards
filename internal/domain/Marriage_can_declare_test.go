//go:build test

package domain

import "testing"

func marriageCanDeclareTestPlayer(human bool, cards []*Card) *MarriagePlayer {
	p := NewMarriagePlayer(human)
	for _, c := range cards {
		p.AddCard(c)
	}
	return p
}

func marriageSevenSetsHand() []*Card {
	cards := make([]*Card, 0, MarriageHandSize)
	for _, value := range []int{2, 4, 6, 8, 10, 12, 13} {
		for _, design := range []int{CardDesignSpade, CardDesignHeart, CardDesignDiamond} {
			cards = append(cards, NewCard(design, value, false))
		}
	}
	return cards
}

func marriageValidDeclarationHand() []*Card {
	cards := []*Card{
		NewCard(CardDesignSpade, 3, false), NewCard(CardDesignSpade, 4, false), NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 6, false), NewCard(CardDesignHeart, 7, false), NewCard(CardDesignHeart, 8, false),
		NewCard(CardDesignClover, 10, false), NewCard(CardDesignClover, 11, false), NewCard(CardDesignClover, 12, false),
	}
	for _, value := range []int{2, 9, 13, 1} {
		for _, design := range []int{CardDesignSpade, CardDesignHeart, CardDesignDiamond} {
			cards = append(cards, NewCard(design, value, false))
		}
	}
	return cards
}

func TestMarriageCanDeclare(t *testing.T) {
	valid := append(marriageValidDeclarationHand(), NewCard(CardDesignClover, 7, false))
	invalidButZeroDeadwood := marriageSevenSetsHand()
	invalidButZeroDeadwood = append(invalidButZeroDeadwood, NewCard(CardDesignClover, 2, false))

	tests := []struct {
		name  string
		cards []*Card
		want  bool
	}{
		{name: "valid declaration", cards: valid, want: true},
		{name: "zero deadwood without three pure sequences", cards: invalidButZeroDeadwood, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Marriage{
				players:          []*MarriagePlayer{marriageCanDeclareTestPlayer(true, tt.cards)},
				currentPlayerIdx: 0,
				phase:            MarriagePhaseDiscard,
			}
			if got := g.CanDeclare(); got != tt.want {
				t.Fatalf("CanDeclare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarriageCanDeclareRejectsUnmeldableHandWithThreePureSequences(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 3, false), NewCard(CardDesignSpade, 4, false), NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 7, false), NewCard(CardDesignHeart, 8, false), NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignDiamond, 11, false), NewCard(CardDesignDiamond, 12, false), NewCard(CardDesignDiamond, 13, false),
	}
	for i := 0; i < 13; i++ {
		cards = append(cards, NewCard(CardDesignClover, 1, false))
	}

	if !MarriageHasPureSequences(cards, 0, 3) {
		t.Fatal("test hand must contain three mutually disjoint pure sequences")
	}
	g := &Marriage{
		players:          []*MarriagePlayer{marriageCanDeclareTestPlayer(true, cards)},
		currentPlayerIdx: 0,
		phase:            MarriagePhaseDiscard,
	}
	if g.CanDeclare() {
		t.Fatal("CanDeclare() = true for a hand that cannot meld all cards after any discard")
	}
}

func TestMarriageCanDeclareTurnGuards(t *testing.T) {
	cards := append(marriageValidDeclarationHand(), NewCard(CardDesignClover, 7, false))
	tests := []struct {
		name    string
		phase   MarriagePhase
		idx     int
		gameEnd bool
	}{
		{name: "draw phase", phase: MarriagePhaseDraw, idx: 0, gameEnd: false},
		{name: "cpu turn", phase: MarriagePhaseDiscard, idx: 1, gameEnd: false},
		{name: "game ended", phase: MarriagePhaseDiscard, idx: 0, gameEnd: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Marriage{
				players:          []*MarriagePlayer{marriageCanDeclareTestPlayer(true, cards), NewMarriagePlayer(false)},
				currentPlayerIdx: tt.idx,
				phase:            tt.phase,
				gameEndFlag:      tt.gameEnd,
			}
			if g.CanDeclare() {
				t.Fatal("CanDeclare() = true outside the human discard turn")
			}
		})
	}
}

func TestMarriageCpuFindDeclareCard(t *testing.T) {
	cards := append(marriageValidDeclarationHand(), NewCard(CardDesignClover, 7, false))
	g := &Marriage{wildRank: 0}
	idx, ok := g.cpuFindDeclareCard(marriageCanDeclareTestPlayer(false, cards))
	if !ok || idx != MarriageHandSize {
		t.Fatalf("cpuFindDeclareCard() = (%d, %v), want (%d, true)", idx, ok, MarriageHandSize)
	}
}

func TestMarriageFindDeclareCardKeepsFirstValidIndexWithDuplicateTypes(t *testing.T) {
	cards := append(marriageValidDeclarationHand(), NewCard(CardDesignClover, 7, false))
	// Duplicate a physical type so candidate pruning must preserve the lowest valid index.
	cards[21] = NewCard(cards[20].GetDesign(), cards[20].GetValue(), false)
	g := &Marriage{wildRank: 0}
	player := marriageCanDeclareTestPlayer(false, cards)
	got, gotOK := g.findDeclareCard(player)
	var want int
	wantOK := false
	for f := 0; f < len(cards); f++ {
		rem := append([]*Card(nil), cards[:f]...)
		rem = append(rem, cards[f+1:]...)
		if MarriageValidateDeclaration(rem, 0) {
			want, wantOK = f, true
			break
		}
	}
	if gotOK != wantOK || gotOK && got != want {
		t.Fatalf("findDeclareCard()=(%d,%v), old scan=(%d,%v)", got, gotOK, want, wantOK)
	}
}

func BenchmarkMarriageCanDeclare(b *testing.B) {
	cards := append(marriageValidDeclarationHand(), NewCard(CardDesignClover, 1, false))
	benchmarks := []struct {
		name  string
		phase MarriagePhase
		idx   int
	}{
		{name: "turn", phase: MarriagePhaseDiscard, idx: 0},
		{name: "out_of_turn", phase: MarriagePhaseDraw, idx: 0},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			g := &Marriage{
				players:          []*MarriagePlayer{marriageCanDeclareTestPlayer(true, cards)},
				currentPlayerIdx: bm.idx,
				phase:            bm.phase,
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = g.CanDeclare()
			}
		})
	}
}
