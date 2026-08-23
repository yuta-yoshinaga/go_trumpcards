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

// RemoveCardsByValue は指定した額面の札を最大 count 枚デッキから取り除き、
// 実際に取り除いた枚数を返す。
//
// **配る前に呼ぶこと。** 既に引かれた札は動かさず、未使用の札だけを外す。
// 席数でデッキが割り切れない卓 (32 枚を 3 人・5 人で分けるなど) で、
// 低い札を抜いて枚数を揃えるために使う。どの札が抜けたかが決まっているので、
// プレイヤーは残りを推測できる —— 無作為に捨てるとそれができない。
func (t *TrumpCards) RemoveCardsByValue(value, count int) int {
	if count <= 0 {
		return 0
	}
	kept := make([]*Card, 0, len(t.deck))
	removed := 0
	for _, c := range t.deck {
		if removed < count && c != nil && !c.GetDraw() && c.GetValue() == value {
			removed++
			continue
		}
		kept = append(kept, c)
	}
	if removed == 0 {
		return 0
	}
	t.deck = kept
	t.deckCnt = len(kept)
	return removed
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

// Replenish makes every card available to DrawCard again without changing
// the underlying deck order. Call this when re-dealing on the same deck
// (e.g. on a 2nd Reset) and randomization is handled separately by the
// caller (typically a seeded `rand.Rand` shuffling the drawn pile).
func (t *TrumpCards) Replenish() {
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

// NewTrumpCards32 German/Czech 32枚デッキコンストラクタ
// 7,8,9,10,J,Q,K,A (値: 1,7,8,9,10,11,12,13) × 4スート = 32枚
//
// **枚数を指定する NewTrumpCardsWithSuits では作れない。** あちらはスートごとに
// 値 1..13 を回して指定枚数で打ち切るので、32 を渡すと 13+13+6+0 になり
// ダイヤが 1 枚も入らない (実測)。使う値を並べて作るのはそのため。
func NewTrumpCards32() *TrumpCards {
	values := []int{1, 7, 8, 9, 10, 11, 12, 13} // A,7,8,9,10,J,Q,K
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	totalCards := len(values) * len(suits) // 32

	t := new(TrumpCards)
	t.deckCnt = totalCards
	t.deck = make([]*Card, 0, totalCards)
	for _, suit := range suits {
		for _, val := range values {
			t.deck = append(t.deck, NewCard(suit, val, false))
		}
	}
	t.deckInit()
	return t
}

// NewTrumpCardsBelote ベロート用32枚デッキコンストラクタ
// 7,8,9,10,J,Q,K,A (値: 1,7,8,9,10,11,12,13) × 4スート = 32枚
func NewTrumpCardsBelote() *TrumpCards {
	return NewTrumpCards32()
}

// NewTrumpCardsPrsi プルシー(チェコ版クレイジーエイト/Mau Mau)用32枚デッキコンストラクタ
// 7,8,9,10,J,Q,K,A (値: 1,7,8,9,10,11,12,13) × 4スート = 32枚
// ベロートと同一構成 (German/Czech 32-card pack)。
func NewTrumpCardsPrsi() *TrumpCards {
	return NewTrumpCards32()
}

// NewTrumpCardsHasenpfeffer ハーゼンプフェファー用25枚デッキコンストラクタ
// 9,10,J,Q,K,A (値: 1,9,10,11,12,13) × 4スート = 24枚 + ジョーカー1枚 = 25枚
//
// **25枚は 4人 × 6枚 + 伏せ札 1枚。** ジョーカーは Best Bower として全カード中
// 最強の切り札になるため、ユーカーの24枚に 1枚だけ足した構成になっている。
func NewTrumpCardsHasenpfeffer() *TrumpCards {
	values := []int{1, 9, 10, 11, 12, 13} // A,9,10,J,Q,K
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	totalCards := len(values)*len(suits) + 1 // 25

	t := new(TrumpCards)
	t.deckCnt = totalCards
	t.deck = make([]*Card, 0, totalCards)
	for _, suit := range suits {
		for _, val := range values {
			t.deck = append(t.deck, NewCard(suit, val, false))
		}
	}
	t.deck = append(t.deck, NewCard(CardDesignJoker, CardValueJoker, false))
	t.deckInit()
	return t
}

// NewTrumpCardsTeenDoPaanch 3-2-5用30枚デッキコンストラクタ
// 8,9,10,J,Q,K,A (値: 1,8,9,10,11,12,13) × 4スート = 28枚 + 7♠ + 7♥ = 30枚
//
// **28枚では 3人 × 10枚 に 2枚足りない。** 3-2-5 は 3+2+5 = 10 トリックを
// 打つので手札はちょうど 10枚必要で、そのために 7♠ と 7♥ の 2枚だけを足した
// 30枚デッキを使う（7 は 8 のすぐ下）。
func NewTrumpCardsTeenDoPaanch() *TrumpCards {
	values := []int{1, 8, 9, 10, 11, 12, 13} // A,8,9,10,J,Q,K
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	extraSevens := []int{CardDesignSpade, CardDesignHeart}
	totalCards := len(values)*len(suits) + len(extraSevens) // 30

	t := new(TrumpCards)
	t.deckCnt = totalCards
	t.deck = make([]*Card, 0, totalCards)
	for _, suit := range suits {
		for _, val := range values {
			t.deck = append(t.deck, NewCard(suit, val, false))
		}
	}
	for _, suit := range extraSevens {
		t.deck = append(t.deck, NewCard(suit, 7, false))
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

// NewTrumpCardsBriscola ブリスコラ用40枚デッキコンストラクタ
// A,2,3,4,5,6,7,J,Q,K (値: 1,2,3,4,5,6,7,11,12,13) × 4スート = 40枚
// 8,9,10 を除外する。
func NewTrumpCardsBriscola() *TrumpCards {
	briscolaValues := []int{1, 2, 3, 4, 5, 6, 7, 11, 12, 13}
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	totalCards := len(briscolaValues) * len(suits) // 40

	t := new(TrumpCards)
	t.deckCnt = totalCards
	t.deck = make([]*Card, 0, totalCards)
	for _, suit := range suits {
		for _, val := range briscolaValues {
			t.deck = append(t.deck, NewCard(suit, val, false))
		}
	}
	t.deckInit()
	return t
}

// NewTrumpCardsScopa スコパ用40枚デッキコンストラクタ
// A,2,3,4,5,6,7,J,Q,K (値: 1,2,3,4,5,6,7,11,12,13) × 4スート = 40枚
// 8,9,10 を除外する (ブリスコラと同一構成)。
func NewTrumpCardsScopa() *TrumpCards {
	scopaValues := []int{1, 2, 3, 4, 5, 6, 7, 11, 12, 13}
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	totalCards := len(scopaValues) * len(suits) // 40

	t := new(TrumpCards)
	t.deckCnt = totalCards
	t.deck = make([]*Card, 0, totalCards)
	for _, suit := range suits {
		for _, val := range scopaValues {
			t.deck = append(t.deck, NewCard(suit, val, false))
		}
	}
	t.deckInit()
	return t
}

// NewTrumpCardsSchnapsen シュナプセン/Sixty-Six 用20枚デッキコンストラクタ
// A,10,J,Q,K (値: 1,10,11,12,13) × 4スート = 20枚
// 2-9 を除外する。
func NewTrumpCardsSchnapsen() *TrumpCards {
	schnapsenValues := []int{1, 10, 11, 12, 13} // A,10,J,Q,K
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	totalCards := len(schnapsenValues) * len(suits) // 20

	t := new(TrumpCards)
	t.deckCnt = totalCards
	t.deck = make([]*Card, 0, totalCards)
	for _, suit := range suits {
		for _, val := range schnapsenValues {
			t.deck = append(t.deck, NewCard(suit, val, false))
		}
	}
	t.deckInit()
	return t
}

// NewTrumpCardsBezique ベジーク用64枚デッキコンストラクタ
// A,7,8,9,10,J,Q,K (値: 1,7,8,9,10,11,12,13) × 4スート × 2セット = 64枚
// 2-6 を除外し、32枚パックを2組重ねる。
func NewTrumpCardsBezique() *TrumpCards {
	beziqueValues := []int{1, 7, 8, 9, 10, 11, 12, 13} // A,7,8,9,10,J,Q,K
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	totalCards := len(beziqueValues) * len(suits) * 2 // 64

	t := new(TrumpCards)
	t.deckCnt = totalCards
	t.deck = make([]*Card, 0, totalCards)
	for range 2 {
		for _, suit := range suits {
			for _, val := range beziqueValues {
				t.deck = append(t.deck, NewCard(suit, val, false))
			}
		}
	}
	t.deckInit()
	return t
}

// ShortDeckValues はショートデック(6+)の札値 A,6-K。デッキ構築 (core) と
// 役判定 (casino) の両方が参照するため、untagged な core ファイルに置く (#2126)。
var ShortDeckValues = []int{1, 6, 7, 8, 9, 10, 11, 12, 13}

// NewTrumpCardsReversis レヴェルシ用48枚デッキコンストラクタ
// 標準52枚から**10を4枚抜いた**48枚。
// A,2,3,4,5,6,7,8,9,J,Q,K (値: 1..9,11,12,13) × 4スート = 48枚。
// ピノクルの48枚とは構成が違う（あちらは9〜Aの短いデッキを2組）ので流用できない。
// 4人に12枚ずつ配ると過不足なく0枚残る。
func NewTrumpCardsReversis() *TrumpCards {
	reversisValues := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12, 13} // 10 を除く
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	totalCards := len(reversisValues) * len(suits) // 48

	t := new(TrumpCards)
	t.deckCnt = totalCards
	t.deck = make([]*Card, 0, totalCards)
	for _, suit := range suits {
		for _, val := range reversisValues {
			t.deck = append(t.deck, NewCard(suit, val, false))
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

// TrappolaValues トラッポラの札位 (A,3,4,5,6,7,J,Q,K)。
//
// **36 枚だが ShortDeckValues とは別集合。** ショートデック / ナインティナインは
// A,6..K で 2..5 を抜くのに対し、トラッポラは 2 と 8,9,10 を抜く。
var TrappolaValues = []int{1, 3, 4, 5, 6, 7, 11, 12, 13}

// NewTrumpCardsTrappola トラッポラ用36枚デッキコンストラクタ
// A,3,4,5,6,7,J,Q,K (値: 1,3,4,5,6,7,11,12,13) × 4スート = 36枚
//
// **枚数を指定する NewTrumpCardsWithSuits では作れない。** あちらはスートごとに
// 値 1..13 を回して指定枚数で打ち切るので、36 を渡すと 13+13+10+0 になり
// ダイヤが 1 枚も入らない (実測)。NewTrumpCards32 と同じく値を並べて作る。
func NewTrumpCardsTrappola() *TrumpCards {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	totalCards := len(TrappolaValues) * len(suits) // 36

	t := new(TrumpCards)
	t.deckCnt = totalCards
	t.deck = make([]*Card, 0, totalCards)
	for _, suit := range suits {
		for _, val := range TrappolaValues {
			t.deck = append(t.deck, NewCard(suit, val, false))
		}
	}
	t.deckInit()
	return t
}

// NewTrumpCardsNinetyNine ナインティナイン(David Parlett)用36枚デッキコンストラクタ
// 正規のNinety-Nineパックは2,3,4,5を抜き、A,6,7,8,9,10,J,Q,K
// (値: 1,6,7,8,9,10,11,12,13) × 4スート = 36枚で構成される。
// これはショートデック(6+)と同一の36枚構成のため ShortDeckValues を共有する。
// 3人に12枚ずつ配ると過不足なく0枚残る。
func NewTrumpCardsNinetyNine() *TrumpCards {
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

// NewTrumpCardsFiveHundred 500 (Five Hundred) 用43枚デッキコンストラクタ
// ジョーカー1枚 + 赤スート(♥♦)の 4〜A(値: 4-13,1 = 11枚) + 黒スート(♠♣)の 5〜A(値: 5-13,1 = 10枚)
// 合計 11×2 + 10×2 + 1 = 43枚。赤の2〜3、黒の2〜4を抜いた標準的な4人用500デッキ。
func NewTrumpCardsFiveHundred() *TrumpCards {
	redValues := []int{1, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13} // A,4..K
	blackValues := []int{1, 5, 6, 7, 8, 9, 10, 11, 12, 13}  // A,5..K
	redSuits := []int{CardDesignHeart, CardDesignDiamond}
	blackSuits := []int{CardDesignSpade, CardDesignClover}

	t := new(TrumpCards)
	t.deck = make([]*Card, 0, 43)
	for _, suit := range redSuits {
		for _, val := range redValues {
			t.deck = append(t.deck, NewCard(suit, val, false))
		}
	}
	for _, suit := range blackSuits {
		for _, val := range blackValues {
			t.deck = append(t.deck, NewCard(suit, val, false))
		}
	}
	t.deck = append(t.deck, NewCard(CardDesignJoker, 1, false))
	t.deckCnt = len(t.deck) // 43
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
