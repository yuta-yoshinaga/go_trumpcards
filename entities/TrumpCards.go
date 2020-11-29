package entities

import "math/rand"

// TrumpCards トランプカードクラス
type TrumpCards struct {
	deck        []*Card // 山札
	deckDrowCnt int     // 山札配った枚数
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
	t.deck = make([]*Card, 0)
	for i := 0; i < t.deckCnt; i++ {
		var design, value int
		if 0 <= i && i <= 12 {
			// スペード
			design = CardDesignSpade
			value = i + 1
		} else if 13 <= i && i <= 25 {
			// クローバー
			design = CardDesignClover
			value = (i - 13) + 1
		} else if 26 <= i && i <= 38 {
			// ハート
			design = CardDesignHeart
			value = (i - 26) + 1
		} else if 39 <= i && i <= 51 {
			// ダイアモンド
			design = CardDesignDiamond
			value = (i - 39) + 1
		} else {
			// ジョーカー
			design = CardValueJoker
			value = (i - 52) + 1
		}
		card := NewCard(design, value, false)
		t.deck = append(t.deck, card)
	}
}

// deckInit 山札初期化
func (t *TrumpCards) deckInit() {
	t.deckDrowFlagInit()
	t.deckDrowCnt = 0
}

// deckDrowFlagInit 山札ドローフラグ初期化
func (t *TrumpCards) deckDrowFlagInit() {
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
	if t.deckDrowCnt < t.deckCnt {
		t.deck[t.deckDrowCnt].SetDraw(true)
		res = t.deck[t.deckDrowCnt]
		t.deckDrowCnt++
	}
	return res
}
