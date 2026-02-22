package entities

import "math/rand"

// TrumpCards トランプカードクラス
type TrumpCards struct {
	deck        []*Card // 山札
	deckDrawCnt int     // 山札配った枚数
	deckCnt     int     // 山札枚数
}

// NewTrumpCards コンストラクタ
func NewTrumpCards(jokerCnt int) *TrumpCards {
	t := new(TrumpCards)
	t.deckCnt = CardCnt + jokerCnt
	t.cardsInit()
	t.deckInit()
	return t
}

// cardsInit カード初期化
func (t *TrumpCards) cardsInit() {
	t.deck = make([]*Card, 0, t.deckCnt)

	// デザインのリスト
	designs := []int{
		CardDesignSpade,
		CardDesignClover,
		CardDesignHeart,
		CardDesignDiamond,
	}

	// 通常カード (各スート 1-13)
	for _, design := range designs {
		for val := 1; val <= CardValueMax; val++ {
			card := NewCard(design, val, false)
			t.deck = append(t.deck, card)
		}
	}

	// ジョーカー (残り枚数分)
	jokerCount := t.deckCnt - CardCnt
	for i := 1; i <= jokerCount; i++ {
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
