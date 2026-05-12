package domain

import (
	"encoding/json"
	"math/rand"
)

// TrumpCards トランプカードクラス
type TrumpCards struct {
	deck        []*Card // 山札
	deckDrawCnt int     // 山札配った枚数
	deckCnt     int     // 山札枚数
}

// NewTrumpCards コンストラクタ
func NewTrumpCards(jokerCnt int) *TrumpCards {
	return NewTrumpCardsWithDecks(1, jokerCnt)
}

// NewTrumpCardsWithDecks マルチデッキ対応コンストラクタ
func NewTrumpCardsWithDecks(deckCount, jokerCnt int) *TrumpCards {
	t := new(TrumpCards)
	t.deckCnt = CardCnt*deckCount + jokerCnt
	t.cardsInit(deckCount, jokerCnt)
	t.deckInit()
	return t
}

// cardsInit カード初期化
func (t *TrumpCards) cardsInit(deckCount, jokerCnt int) {
	t.deck = make([]*Card, 0, t.deckCnt)

	// デザインのリスト
	designs := []int{
		CardDesignSpade,
		CardDesignClover,
		CardDesignHeart,
		CardDesignDiamond,
	}

	// 通常カード (deckCount デッキ分)
	for range deckCount {
		for _, design := range designs {
			for val := 1; val <= CardValueMax; val++ {
				card := NewCard(design, val, false)
				t.deck = append(t.deck, card)
			}
		}
	}

	// ジョーカー
	for i := 1; i <= jokerCnt; i++ {
		card := NewCard(CardDesignJoker, i, false)
		t.deck = append(t.deck, card)
	}
}

// deckInit 山札初期化
func (t *TrumpCards) deckInit() {
	t.deckDrawFlagInit()
	t.deckDrawCnt = 0
}

// deckDrawFlagInit 山札ドローフラグ初期化
func (t *TrumpCards) deckDrawFlagInit() {
	for _, v := range t.deck {
		v.SetDraw(false)
	}
}

// Shuffle 山札シャッフル
func (t *TrumpCards) Shuffle() {
	n := len(t.deck)
	for i := n - 1; i >= 0; i-- {
		j := rand.Intn(i + 1)
		t.deck[i], t.deck[j] = t.deck[j], t.deck[i]
	}
	t.deckInit()
}

// GetRemainingCount 山札の残り枚数
func (t *TrumpCards) GetRemainingCount() int {
	return t.deckCnt - t.deckDrawCnt
}

// GetTotalCount 山札の総枚数
func (t *TrumpCards) GetTotalCount() int {
	return t.deckCnt
}

// NewTrumpCardsWithSuits スート指定マルチデッキコンストラクタ
// suits で指定されたスートのみを使い、合計 totalCards 枚のデッキを作成する。
// 各スートの13枚をラウンドロビンで繰り返し追加し、totalCards 枚に達したら停止する。
func NewTrumpCardsWithSuits(totalCards int, suits []int) *TrumpCards {
	t := new(TrumpCards)
	t.deckCnt = totalCards
	t.deck = make([]*Card, 0, totalCards)
	for len(t.deck) < totalCards {
		for _, suit := range suits {
			for val := 1; val <= CardValueMax; val++ {
				if len(t.deck) >= totalCards {
					break
				}
				t.deck = append(t.deck, NewCard(suit, val, false))
			}
			if len(t.deck) >= totalCards {
				break
			}
		}
	}
	t.deckInit()
	return t
}

// NewTrumpCardsEuchre ユーカー用24枚デッキコンストラクタ
// 9,10,J,Q,K,A (値: 1,9,10,11,12,13) × 4スート = 24枚
func NewTrumpCardsEuchre() *TrumpCards {
	euchreValues := []int{1, 9, 10, 11, 12, 13} // A,9,10,J,Q,K
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	totalCards := len(euchreValues) * len(suits) // 24

	t := new(TrumpCards)
	t.deckCnt = totalCards
	t.deck = make([]*Card, 0, totalCards)
	for _, suit := range suits {
		for _, val := range euchreValues {
			t.deck = append(t.deck, NewCard(suit, val, false))
		}
	}
	t.deckInit()
	return t
}

// NewTrumpCardsBelote ベロート用32枚デッキコンストラクタ
// 7,8,9,10,J,Q,K,A (値: 1,7,8,9,10,11,12,13) × 4スート = 32枚
func NewTrumpCardsBelote() *TrumpCards {
	beloteValues := []int{1, 7, 8, 9, 10, 11, 12, 13} // A,7,8,9,10,J,Q,K
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	totalCards := len(beloteValues) * len(suits) // 32

	t := new(TrumpCards)
	t.deckCnt = totalCards
	t.deck = make([]*Card, 0, totalCards)
	for _, suit := range suits {
		for _, val := range beloteValues {
			t.deck = append(t.deck, NewCard(suit, val, false))
		}
	}
	t.deckInit()
	return t
}

// NewTrumpCardsPinochle ピノクル用48枚デッキコンストラクタ
// 9,10,J,Q,K,A (値: 1,9,10,11,12,13) × 4スート × 2セット = 48枚
func NewTrumpCardsPinochle() *TrumpCards {
	pinochleValues := []int{1, 9, 10, 11, 12, 13} // A,9,10,J,Q,K
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	totalCards := len(pinochleValues) * len(suits) * 2 // 48

	t := new(TrumpCards)
	t.deckCnt = totalCards
	t.deck = make([]*Card, 0, totalCards)
	for range 2 {
		for _, suit := range suits {
			for _, val := range pinochleValues {
				t.deck = append(t.deck, NewCard(suit, val, false))
			}
		}
	}
	t.deckInit()
	return t
}

// NewTrumpCardsShortDeck ショートデック(6+)用36枚デッキコンストラクタ
// A,6,7,8,9,10,J,Q,K (値: 1,6,7,8,9,10,11,12,13) × 4スート = 36枚
func NewTrumpCardsShortDeck() *TrumpCards {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	totalCards := len(ShortDeckValues) * len(suits) // 36

	t := new(TrumpCards)
	t.deckCnt = totalCards
	t.deck = make([]*Card, 0, totalCards)
	for _, suit := range suits {
		for _, val := range ShortDeckValues {
			t.deck = append(t.deck, NewCard(suit, val, false))
		}
	}
	t.deckInit()
	return t
}

// DrawCard 山札配る
func (t *TrumpCards) DrawCard() *Card {
	var res *Card = nil
	if t.deckDrawCnt < t.deckCnt {
		t.deck[t.deckDrawCnt].SetDraw(true)
		res = t.deck[t.deckDrawCnt]
		t.deckDrawCnt++
	}
	return res
}

// trumpCardsJSON is the JSON wire format for TrumpCards.
type trumpCardsJSON struct {
	Deck        []*Card `json:"dk"`
	DeckDrawCnt int     `json:"dc"`
	DeckCnt     int     `json:"dn"`
}

// MarshalJSON implements json.Marshaler.
func (t *TrumpCards) MarshalJSON() ([]byte, error) {
	return json.Marshal(trumpCardsJSON{
		Deck:        t.deck,
		DeckDrawCnt: t.deckDrawCnt,
		DeckCnt:     t.deckCnt,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *TrumpCards) UnmarshalJSON(data []byte) error {
	var j trumpCardsJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	t.deck = j.Deck
	if t.deck == nil {
		t.deck = make([]*Card, 0)
	}
	t.deckDrawCnt = j.DeckDrawCnt
	t.deckCnt = j.DeckCnt
	return nil
}
