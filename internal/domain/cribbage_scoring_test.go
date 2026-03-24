//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// helper: create card with design and value
func cCard(design, value int) *Card {
	return NewCard(design, value, false)
}

// ---- cribbageCardValue ----

func TestCribbageCardValue(t *testing.T) {
	tests := []struct {
		name  string
		value int
		want  int
	}{
		{"Ace", 1, 1},
		{"Two", 2, 2},
		{"Nine", 9, 9},
		{"Ten", 10, 10},
		{"Jack", 11, 10},
		{"Queen", 12, 10},
		{"King", 13, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cCard(CardDesignSpade, tt.value)
			assert.Equal(t, tt.want, cribbageCardValue(c))
		})
	}
}

// ---- CribbageScoreFifteens ----

func TestCribbageScoreFifteens(t *testing.T) {
	tests := []struct {
		name  string
		cards []*Card
		want  int
	}{
		{
			"5+10=15 → 2pts",
			[]*Card{cCard(1, 5), cCard(1, 10), cCard(2, 1), cCard(3, 1), cCard(4, 2)},
			2,
		},
		{
			"5+J+5+K=two fifteens → 4pts",
			[]*Card{cCard(1, 5), cCard(2, 11), cCard(3, 5), cCard(4, 13), cCard(1, 2)},
			// 5+J=15, 5+K=15, 5+J=15(2nd 5), 5+K=15(2nd 5) → but also 5+5+2+... hmm
			// Actually: 5♠+J=15, 5♠+K=15, 5♥+J=15, 5♥+K=15 → 8pts
			8,
		},
		{
			"no fifteens",
			[]*Card{cCard(1, 1), cCard(2, 1), cCard(3, 1), cCard(4, 1), cCard(1, 2)},
			0,
		},
		{
			"7+8=15 → 2pts",
			[]*Card{cCard(1, 7), cCard(2, 8), cCard(3, 1), cCard(4, 2), cCard(1, 3)},
			// 7+8=15, also 7+3+...? 7+2+3+...? 1+2+3+...?
			// 7+8=15 → 2pts; 7+5? no; 1+2+3+...? A+2+3+...=6,9...no
			2,
		},
		{
			"perfect 29: J♥+5♦+5♠+5♣+5♥ → 16 fifteens",
			[]*Card{cCard(3, 11), cCard(4, 5), cCard(1, 5), cCard(2, 5), cCard(3, 5)},
			// J(10)+5=15: 4 ways; 5+5+5=15: C(4,3)=4 ways; total 8 fifteens → 16pts
			16,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CribbageScoreFifteens(tt.cards))
		})
	}
}

// ---- CribbageScorePairs ----

func TestCribbageScorePairs(t *testing.T) {
	tests := []struct {
		name  string
		cards []*Card
		want  int
	}{
		{
			"one pair",
			[]*Card{cCard(1, 3), cCard(2, 3), cCard(3, 7), cCard(4, 8), cCard(1, 9)},
			2,
		},
		{
			"three of a kind",
			[]*Card{cCard(1, 3), cCard(2, 3), cCard(3, 3), cCard(4, 8), cCard(1, 9)},
			6,
		},
		{
			"four of a kind",
			[]*Card{cCard(1, 5), cCard(2, 5), cCard(3, 5), cCard(4, 5), cCard(1, 9)},
			12,
		},
		{
			"two pairs",
			[]*Card{cCard(1, 3), cCard(2, 3), cCard(3, 7), cCard(4, 7), cCard(1, 9)},
			4,
		},
		{
			"no pairs",
			[]*Card{cCard(1, 1), cCard(2, 3), cCard(3, 5), cCard(4, 7), cCard(1, 9)},
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CribbageScorePairs(tt.cards))
		})
	}
}

// ---- CribbageScoreRuns ----

func TestCribbageScoreRuns(t *testing.T) {
	tests := []struct {
		name  string
		cards []*Card
		want  int
	}{
		{
			"run of 3",
			[]*Card{cCard(1, 3), cCard(2, 4), cCard(3, 5), cCard(4, 9), cCard(1, 13)},
			3,
		},
		{
			"run of 4",
			[]*Card{cCard(1, 3), cCard(2, 4), cCard(3, 5), cCard(4, 6), cCard(1, 13)},
			4,
		},
		{
			"run of 5",
			[]*Card{cCard(1, 3), cCard(2, 4), cCard(3, 5), cCard(4, 6), cCard(1, 7)},
			5,
		},
		{
			"double run of 3 (3-4-4-5)",
			[]*Card{cCard(1, 3), cCard(2, 4), cCard(3, 4), cCard(4, 5), cCard(1, 9)},
			6, // 3 × 2 = 6
		},
		{
			"double-double run of 3 (3-3-4-5-5)",
			[]*Card{cCard(1, 3), cCard(2, 3), cCard(3, 4), cCard(4, 5), cCard(1, 5)},
			12, // 3 × 2 × 2 = 12
		},
		{
			"triple run of 3 (3-3-3-4-5)",
			[]*Card{cCard(1, 3), cCard(2, 3), cCard(3, 3), cCard(4, 4), cCard(1, 5)},
			9, // 3 × 3 = 9
		},
		{
			"double run of 4 (3-4-5-5-6)",
			[]*Card{cCard(1, 3), cCard(2, 4), cCard(3, 5), cCard(4, 5), cCard(1, 6)},
			8, // 4 × 2 = 8
		},
		{
			"A-2-3 is a valid run",
			[]*Card{cCard(1, 1), cCard(2, 2), cCard(3, 3), cCard(4, 9), cCard(1, 13)},
			3,
		},
		{
			"Q-K-A is NOT a run (no wrap)",
			[]*Card{cCard(1, 12), cCard(2, 13), cCard(3, 1), cCard(4, 7), cCard(1, 9)},
			0,
		},
		{
			"no runs",
			[]*Card{cCard(1, 1), cCard(2, 3), cCard(3, 5), cCard(4, 7), cCard(1, 9)},
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CribbageScoreRuns(tt.cards))
		})
	}
}

// ---- CribbageScoreFlush ----

func TestCribbageScoreFlush(t *testing.T) {
	tests := []struct {
		name    string
		hand    []*Card
		starter *Card
		isCrib  bool
		want    int
	}{
		{
			"4-card flush (hand only)",
			[]*Card{cCard(1, 3), cCard(1, 5), cCard(1, 7), cCard(1, 9)},
			cCard(2, 10),
			false,
			4,
		},
		{
			"5-card flush (hand + starter)",
			[]*Card{cCard(1, 3), cCard(1, 5), cCard(1, 7), cCard(1, 9)},
			cCard(1, 10),
			false,
			5,
		},
		{
			"no flush",
			[]*Card{cCard(1, 3), cCard(2, 5), cCard(1, 7), cCard(1, 9)},
			cCard(1, 10),
			false,
			0,
		},
		{
			"crib: 4 same suit + different starter = 0",
			[]*Card{cCard(1, 3), cCard(1, 5), cCard(1, 7), cCard(1, 9)},
			cCard(2, 10),
			true,
			0,
		},
		{
			"crib: 5 same suit = 5",
			[]*Card{cCard(1, 3), cCard(1, 5), cCard(1, 7), cCard(1, 9)},
			cCard(1, 10),
			true,
			5,
		},
		{
			"nil starter, 4 same suit = 4",
			[]*Card{cCard(1, 3), cCard(1, 5), cCard(1, 7), cCard(1, 9)},
			nil,
			false,
			4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CribbageScoreFlush(tt.hand, tt.starter, tt.isCrib))
		})
	}
}

// ---- CribbageScoreNobs ----

func TestCribbageScoreNobs(t *testing.T) {
	tests := []struct {
		name    string
		hand    []*Card
		starter *Card
		want    int
	}{
		{
			"J matching starter suit",
			[]*Card{cCard(1, 11), cCard(2, 5), cCard(3, 7), cCard(4, 9)},
			cCard(1, 3),
			1,
		},
		{
			"J not matching starter suit",
			[]*Card{cCard(1, 11), cCard(2, 5), cCard(3, 7), cCard(4, 9)},
			cCard(2, 3),
			0,
		},
		{
			"no J in hand",
			[]*Card{cCard(1, 3), cCard(2, 5), cCard(3, 7), cCard(4, 9)},
			cCard(1, 10),
			0,
		},
		{
			"nil starter",
			[]*Card{cCard(1, 11), cCard(2, 5), cCard(3, 7), cCard(4, 9)},
			nil,
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CribbageScoreNobs(tt.hand, tt.starter))
		})
	}
}

// ---- CribbageScoreHand (integration) ----

func TestCribbageScoreHand(t *testing.T) {
	tests := []struct {
		name    string
		hand    []*Card
		starter *Card
		isCrib  bool
		want    int
	}{
		{
			"perfect 29: J♥+5♦+5♠+5♣, starter=5♥",
			[]*Card{cCard(3, 11), cCard(4, 5), cCard(1, 5), cCard(2, 5)},
			cCard(3, 5),
			false,
			29,
		},
		{
			"zero hand: A♠+3♦+7♣+9♥, starter=12♠ (no fifteens/pairs/runs/flush)",
			[]*Card{cCard(1, 1), cCard(4, 3), cCard(2, 7), cCard(3, 9)},
			cCard(1, 12),
			false,
			0,
		},
		{
			"double run: 3♠-4♠-4♥-5♠, starter=K♦",
			// pairs: (4,4)=2, runs: 3×2=6, fifteens: 5+K=15 → 2
			[]*Card{cCard(1, 3), cCard(1, 4), cCard(3, 4), cCard(1, 5)},
			cCard(4, 13),
			false,
			10,
		},
		{
			"flush + run: 3♠-4♠-5♠-7♠, starter=9♦",
			// flush 4, run 3, fifteens: {3,5,7}=15 → 2, nobs: 0
			[]*Card{cCard(1, 3), cCard(1, 4), cCard(1, 5), cCard(1, 7)},
			cCard(4, 9),
			false,
			9,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detail := CribbageScoreHand(tt.hand, tt.starter, tt.isCrib)
			assert.Equal(t, tt.want, detail.Total)
		})
	}
}

// ---- CribbageScorePegging ----

func TestCribbageScorePegging(t *testing.T) {
	tests := []struct {
		name     string
		played   []*Card
		pegCount int
		want     int
	}{
		{
			"pair",
			[]*Card{cCard(1, 8), cCard(2, 8)},
			16,
			2,
		},
		{
			"three of a kind",
			[]*Card{cCard(1, 8), cCard(2, 8), cCard(3, 8)},
			24,
			6,
		},
		{
			"four of a kind",
			[]*Card{cCard(1, 3), cCard(2, 3), cCard(3, 3), cCard(4, 3)},
			12,
			12,
		},
		{
			"fifteen",
			[]*Card{cCard(1, 5), cCard(2, 10)},
			15,
			2, // 15=2pts, no pair, no run
		},
		{
			"31",
			[]*Card{cCard(1, 10), cCard(2, 10), cCard(3, 10), cCard(4, 1)},
			31,
			2, // 31=2pts, no pair (10,10,10 → trips=6 but also... wait 10,10,10,1 last is 1, not matching 10)
			// last card is A(1), prev cards are 10s. pairCount=0. No run. 31=2pts → 2
		},
		{
			"run of 3 in play order (5-3-4)",
			[]*Card{cCard(1, 5), cCard(2, 3), cCard(3, 4)},
			12,
			3,
		},
		{
			"run of 4 (2-3-5-4)",
			[]*Card{cCard(1, 2), cCard(2, 3), cCard(3, 5), cCard(4, 4)},
			14,
			4,
		},
		{
			"no score",
			[]*Card{cCard(1, 1), cCard(2, 7)},
			8,
			0,
		},
		{
			"last card (single card, count=10)",
			[]*Card{cCard(1, 10)},
			10,
			0,
		},
		{
			"fifteen + pair (7-8-7 is not a run, but if 7+8=15... no, count=22)",
			[]*Card{cCard(1, 7), cCard(2, 8), cCard(3, 7)},
			22,
			0, // no pair (7,8,7 → last is 7, prev is 8 → no pair), no run (7,8,7 sorted=7,7,8 not consecutive)
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CribbageScorePegging(tt.played, tt.pegCount))
		})
	}
}

// ---- sortInts ----

func TestSortInts(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"sorted", []int{1, 2, 3}, []int{1, 2, 3}},
		{"reverse", []int{3, 2, 1}, []int{1, 2, 3}},
		{"mixed", []int{5, 3, 4, 1, 2}, []int{1, 2, 3, 4, 5}},
		{"empty", []int{}, []int{}},
		{"single", []int{1}, []int{1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortInts(tt.in)
			assert.Equal(t, tt.want, tt.in)
		})
	}
}
