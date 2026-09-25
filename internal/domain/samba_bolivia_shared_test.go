//go:build test

package domain

import (
	"testing"
)

// --- helper: カードを素早く作る ---

// newCard は Design=d, Value=v のカードをテスト用に返す。
func newCard(d, v int) *Card {
	return &Card{design: d, value: v}
}

// newJoker はジョーカーを返す。
func newJoker() *Card {
	return &Card{design: CardDesignJoker, value: 0}
}

// --- sbIsWildCard ---

func TestSbIsWildCard(t *testing.T) {
	cases := []struct {
		name string
		card *Card
		want bool
	}{
		{"joker", newJoker(), true},
		{"2 of hearts", newCard(CardDesignHeart, 2), true},
		{"2 of spades", newCard(CardDesignSpade, 2), true},
		{"ace of spades", newCard(CardDesignSpade, 1), false},
		{"king of hearts", newCard(CardDesignHeart, 13), false},
		{"3 of clubs", newCard(CardDesignClover, 3), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sbIsWildCard(tc.card)
			if got != tc.want {
				t.Errorf("sbIsWildCard(%v) = %v, want %v", tc.card, got, tc.want)
			}
		})
	}
}

// --- sbSequenceCardValue ---

func TestSbSequenceCardValue(t *testing.T) {
	cases := []struct {
		value int
		want  int
	}{
		{1, 14}, // Ace → 14
		{2, 2},
		{3, 3},
		{10, 10},
		{13, 13}, // King → 13
	}
	for _, tc := range cases {
		card := newCard(CardDesignSpade, tc.value)
		got := sbSequenceCardValue(card)
		if got != tc.want {
			t.Errorf("sbSequenceCardValue(value=%d) = %d, want %d", tc.value, got, tc.want)
		}
	}
}

// --- sbSequenceCardValues ---

func TestSbSequenceCardValues_skipsWild(t *testing.T) {
	cards := []*Card{
		newCard(CardDesignSpade, 4),
		newJoker(), // wild → skipped
		newCard(CardDesignSpade, 5),
	}
	got := sbSequenceCardValues(cards)
	if len(got) != 2 || got[0] != 4 || got[1] != 5 {
		t.Errorf("sbSequenceCardValues = %v, want [4 5]", got)
	}
}

func TestSbSequenceCardValues_ace(t *testing.T) {
	cards := []*Card{
		newCard(CardDesignHeart, 1), // Ace → 14
	}
	got := sbSequenceCardValues(cards)
	if len(got) != 1 || got[0] != 14 {
		t.Errorf("sbSequenceCardValues with ace = %v, want [14]", got)
	}
}

func TestSbSequenceCardValues_empty(t *testing.T) {
	got := sbSequenceCardValues([]*Card{})
	if len(got) != 0 {
		t.Errorf("sbSequenceCardValues empty = %v, want []", got)
	}
}

// --- sbSequenceSuit ---

func TestSbSequenceSuit(t *testing.T) {
	t.Run("returns first natural suit", func(t *testing.T) {
		cards := []*Card{
			newJoker(), // wild → skipped
			newCard(CardDesignHeart, 5),
			newCard(CardDesignHeart, 6),
		}
		if got := sbSequenceSuit(cards); got != CardDesignHeart {
			t.Errorf("sbSequenceSuit = %d, want %d", got, CardDesignHeart)
		}
	})
	t.Run("all wild returns -1", func(t *testing.T) {
		cards := []*Card{newJoker(), newCard(CardDesignSpade, 2)}
		if got := sbSequenceSuit(cards); got != -1 {
			t.Errorf("sbSequenceSuit all-wild = %d, want -1", got)
		}
	})
	t.Run("empty returns -1", func(t *testing.T) {
		if got := sbSequenceSuit([]*Card{}); got != -1 {
			t.Errorf("sbSequenceSuit empty = %d, want -1", got)
		}
	})
}

// --- sbGroupIsSetShaped ---

func TestSbGroupIsSetShaped(t *testing.T) {
	t.Run("all same rank", func(t *testing.T) {
		cards := []*Card{
			newCard(CardDesignSpade, 7),
			newCard(CardDesignHeart, 7),
			newCard(CardDesignClover, 7),
		}
		if !sbGroupIsSetShaped(cards) {
			t.Error("expected true for same-rank cards")
		}
	})
	t.Run("different rank", func(t *testing.T) {
		cards := []*Card{
			newCard(CardDesignSpade, 7),
			newCard(CardDesignHeart, 8),
		}
		if sbGroupIsSetShaped(cards) {
			t.Error("expected false for different-rank cards")
		}
	})
	t.Run("wilds ignored", func(t *testing.T) {
		cards := []*Card{
			newCard(CardDesignSpade, 9),
			newJoker(),
			newCard(CardDesignHeart, 9),
		}
		if !sbGroupIsSetShaped(cards) {
			t.Error("expected true when wilds ignored")
		}
	})
	t.Run("all wilds", func(t *testing.T) {
		// 全ワイルドはナチュラルがなく rank==0 のままなので true
		cards := []*Card{newJoker(), newCard(CardDesignSpade, 2)}
		if !sbGroupIsSetShaped(cards) {
			t.Error("expected true for all-wild group")
		}
	})
}

// --- sbNaturalRank ---

func TestSbNaturalRank(t *testing.T) {
	t.Run("returns rank of first natural", func(t *testing.T) {
		cards := []*Card{newJoker(), newCard(CardDesignHeart, 8)}
		if got := sbNaturalRank(cards); got != 8 {
			t.Errorf("sbNaturalRank = %d, want 8", got)
		}
	})
	t.Run("all wild returns 0", func(t *testing.T) {
		cards := []*Card{newJoker(), newCard(CardDesignSpade, 2)}
		if got := sbNaturalRank(cards); got != 0 {
			t.Errorf("sbNaturalRank all-wild = %d, want 0", got)
		}
	})
	t.Run("empty returns 0", func(t *testing.T) {
		if got := sbNaturalRank([]*Card{}); got != 0 {
			t.Errorf("sbNaturalRank empty = %d, want 0", got)
		}
	})
}

// --- sbFilterUnused ---

func TestSbFilterUnused(t *testing.T) {
	c1 := newCard(CardDesignSpade, 4)
	c2 := newCard(CardDesignSpade, 5)
	c3 := newCard(CardDesignSpade, 6)
	used := map[*Card]bool{c2: true}
	got := sbFilterUnused([]*Card{c1, c2, c3}, used)
	if len(got) != 2 || got[0] != c1 || got[1] != c3 {
		t.Errorf("sbFilterUnused = %v, want [c1 c3]", got)
	}
}

func TestSbFilterUnused_empty(t *testing.T) {
	got := sbFilterUnused([]*Card{}, map[*Card]bool{})
	if len(got) != 0 {
		t.Error("sbFilterUnused empty input should return empty")
	}
}

// --- sbValidateSequenceCards ---

func TestSbValidateSequenceCards_valid(t *testing.T) {
	// ♠ 4-5-6 は有効なシーケンス
	cards := []*Card{
		newCard(CardDesignSpade, 4),
		newCard(CardDesignSpade, 5),
		newCard(CardDesignSpade, 6),
	}
	if err := sbValidateSequenceCards(cards, "test"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestSbValidateSequenceCards_validWithAce(t *testing.T) {
	// ♥ Q-K-A は有効 (A=14)
	cards := []*Card{
		newCard(CardDesignHeart, 12),
		newCard(CardDesignHeart, 13),
		newCard(CardDesignHeart, 1),
	}
	if err := sbValidateSequenceCards(cards, "test"); err != nil {
		t.Errorf("expected nil error for Q-K-A, got %v", err)
	}
}

func TestSbValidateSequenceCards_tooFew(t *testing.T) {
	cards := []*Card{
		newCard(CardDesignSpade, 5),
		newCard(CardDesignSpade, 6),
	}
	err := sbValidateSequenceCards(cards, "test")
	if err == nil {
		t.Error("expected error for < 3 cards")
	}
}

func TestSbValidateSequenceCards_wildCardRejected(t *testing.T) {
	cards := []*Card{
		newCard(CardDesignSpade, 4),
		newJoker(), // wild → rejected
		newCard(CardDesignSpade, 6),
	}
	err := sbValidateSequenceCards(cards, "test")
	if err == nil {
		t.Error("expected error when wild card present")
	}
}

func TestSbValidateSequenceCards_threeRejected(t *testing.T) {
	// 3 はシーケンスに使えない
	cards := []*Card{
		newCard(CardDesignSpade, 3),
		newCard(CardDesignSpade, 4),
		newCard(CardDesignSpade, 5),
	}
	err := sbValidateSequenceCards(cards, "test")
	if err == nil {
		t.Error("expected error when 3 is present")
	}
}

func TestSbValidateSequenceCards_mixedSuit(t *testing.T) {
	cards := []*Card{
		newCard(CardDesignSpade, 4),
		newCard(CardDesignHeart, 5), // 別スート
		newCard(CardDesignSpade, 6),
	}
	err := sbValidateSequenceCards(cards, "test")
	if err == nil {
		t.Error("expected error for mixed suits")
	}
}

func TestSbValidateSequenceCards_duplicate(t *testing.T) {
	cards := []*Card{
		newCard(CardDesignSpade, 5),
		newCard(CardDesignSpade, 5), // 重複
		newCard(CardDesignSpade, 6),
	}
	err := sbValidateSequenceCards(cards, "test")
	if err == nil {
		t.Error("expected error for duplicate rank")
	}
}

func TestSbValidateSequenceCards_nonConsecutive(t *testing.T) {
	cards := []*Card{
		newCard(CardDesignSpade, 4),
		newCard(CardDesignSpade, 5),
		newCard(CardDesignSpade, 7), // 6 が抜けている
	}
	err := sbValidateSequenceCards(cards, "test")
	if err == nil {
		t.Error("expected error for non-consecutive ranks")
	}
}

func TestSbValidateSequenceCards_errPrefixPropagated(t *testing.T) {
	// エラーコードの接頭辞が正しく伝わることを確認
	cards := []*Card{
		newCard(CardDesignSpade, 4),
		newCard(CardDesignSpade, 5),
	}
	errSamba := sbValidateSequenceCards(cards, "samba")
	errBolivia := sbValidateSequenceCards(cards, "bolivia")
	if errSamba == nil || errBolivia == nil {
		t.Fatal("expected errors")
	}
	de1, ok1 := errSamba.(*DomainError)
	de2, ok2 := errBolivia.(*DomainError)
	if !ok1 || !ok2 {
		t.Fatal("expected *DomainError")
	}
	if de1.Code == de2.Code {
		t.Errorf("expected different error codes for different prefixes, got both %q", de1.Code)
	}
}
