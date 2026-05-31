package domain

import (
	"testing"
)

func TestScopaCardValue(t *testing.T) {
	cases := []struct {
		val  int
		want int
	}{
		{1, 1}, {2, 2}, {7, 7}, {11, 8}, {12, 9}, {13, 10},
	}
	for _, c := range cases {
		card := NewCard(CardDesignHeart, c.val, false)
		if got := ScopaCardValue(card); got != c.want {
			t.Errorf("ScopaCardValue(val=%d) = %d, want %d", c.val, got, c.want)
		}
	}
	if got := ScopaCardValue(nil); got != 0 {
		t.Errorf("ScopaCardValue(nil) = %d, want 0", got)
	}
}

func TestScopaCardPredicates(t *testing.T) {
	settebello := NewCard(CardDesignDiamond, 7, false)
	if !ScopaIsSetteBello(settebello) {
		t.Error("7♦ should be settebello")
	}
	if !ScopaIsDiamond(settebello) || !ScopaIsSeven(settebello) {
		t.Error("7♦ should be a diamond and a seven")
	}
	sevenHearts := NewCard(CardDesignHeart, 7, false)
	if ScopaIsSetteBello(sevenHearts) {
		t.Error("7♥ should not be settebello")
	}
	if ScopaIsDiamond(sevenHearts) {
		t.Error("7♥ should not be a diamond")
	}
	if ScopaIsDiamond(nil) || ScopaIsSeven(nil) || ScopaIsSetteBello(nil) {
		t.Error("nil card should be false for all predicates")
	}
}

func TestHasSingleValueMatch(t *testing.T) {
	table := []*Card{
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 11, false), // value 8
	}
	if !hasSingleValueMatch(table, 3) {
		t.Error("expected single match for value 3")
	}
	if !hasSingleValueMatch(table, 8) {
		t.Error("expected single match for value 8 (J)")
	}
	if hasSingleValueMatch(table, 5) {
		t.Error("did not expect a single match for value 5")
	}
}

func TestEnumerateScopaCaptures_SingleMatchForced(t *testing.T) {
	// 場に 5 と 2+3 がある。手札 5 を出すと単独 5 のみが合法。
	table := []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 3, false),
	}
	played := NewCard(CardDesignDiamond, 5, false)
	caps := EnumerateScopaCaptures(played, table)
	if len(caps) != 1 || len(caps[0]) != 1 || caps[0][0] != 0 {
		t.Fatalf("expected forced single capture [[0]], got %v", caps)
	}
}

func TestEnumerateScopaCaptures_Combinations(t *testing.T) {
	// 場に 2,3,5。手札 5 を出す。単独 5 が存在するので 2+3 の組合せは禁止。
	table := []*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignSpade, 5, false),
	}
	played := NewCard(CardDesignDiamond, 5, false)
	caps := EnumerateScopaCaptures(played, table)
	if len(caps) != 1 {
		t.Fatalf("expected only the forced single 5, got %v", caps)
	}

	// 場に 2,3 (単独 5 なし)。手札 5 で 2+3 の組合せが合法。
	table2 := []*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 3, false),
	}
	caps2 := EnumerateScopaCaptures(played, table2)
	if len(caps2) != 1 || len(caps2[0]) != 2 {
		t.Fatalf("expected one 2-card combination, got %v", caps2)
	}
	if EnumerateScopaCaptures(nil, table2) != nil {
		t.Error("nil played card should yield nil captures")
	}
}

func TestIsValidScopaCapture(t *testing.T) {
	table := []*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignSpade, 5, false),
	}
	five := NewCard(CardDesignDiamond, 5, false)

	// 単独一致があるので組合せは不可。
	if isValidScopaCapture(five, table, []int{0, 1}) {
		t.Error("combination should be invalid when a single match exists")
	}
	// 単独 5 は OK。
	if !isValidScopaCapture(five, table, []int{2}) {
		t.Error("single matching card should be a valid capture")
	}

	// 単独一致なしの場の組合せ。
	table2 := []*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 3, false),
	}
	if !isValidScopaCapture(five, table2, []int{0, 1}) {
		t.Error("2+3 should capture a 5 when no single 5 exists")
	}
	if isValidScopaCapture(five, table2, []int{0}) {
		t.Error("a single 2 cannot capture a 5")
	}

	// 範囲外 / 重複 / 空 / nil。
	if isValidScopaCapture(five, table2, []int{5}) {
		t.Error("out-of-range index should be invalid")
	}
	if isValidScopaCapture(five, table2, []int{0, 0}) {
		t.Error("duplicate index should be invalid")
	}
	if isValidScopaCapture(five, table2, nil) {
		t.Error("empty selection should be invalid")
	}
	if isValidScopaCapture(nil, table2, []int{0}) {
		t.Error("nil played card should be invalid")
	}
}
