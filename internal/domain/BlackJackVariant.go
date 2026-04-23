package domain

// BlackJackVariantName はブラックジャック系バリアントの識別子
type BlackJackVariantName string

// バリアント識別子定数
const (
	// BJVariantStandard 標準ブラックジャック (空文字列はデフォルト)
	BJVariantStandard BlackJackVariantName = ""
	// BJVariantSpanish21 スパニッシュ21
	BJVariantSpanish21 BlackJackVariantName = "spanish21"
)

// BJBonusPayout はバリアント固有のボーナス配当を表す
// (例: スパニッシュ21の 5-card 21 → MultiplierNum=3, MultiplierDen=2)
type BJBonusPayout struct {
	// NameKey はi18n表示キー (例: "spanish21.bonus.fivecard21")
	NameKey string
	// MultiplierNum は配当倍率の分子 (例: 3:2 の 3)
	MultiplierNum int
	// MultiplierDen は配当倍率の分母 (例: 3:2 の 2)
	MultiplierDen int
}

// BlackJackVariantConfig はブラックジャック系ゲームのバリアント固有設定
// nil の場合は標準ブラックジャック
type BlackJackVariantConfig struct {
	// Name はバリアント識別子
	Name BlackJackVariantName
	// DeckBuilder はデッキ生成関数 (nil = 標準52枚デッキ)
	DeckBuilder func(deckCount int) *TrumpCards
	// Player21AlwaysWins true の場合、プレイヤーが21なら常に勝利 (ディーラーも21でも)
	Player21AlwaysWins bool
	// PlayerBJBeatsDealerBJ true の場合、両者がナチュラルブラックジャックでもプレイヤー勝利
	PlayerBJBeatsDealerBJ bool
	// BonusEval はボーナス配当判定関数 (nil = ボーナス無し)
	// 引数: 完成したプレイヤーハンド、ディーラーアップカード
	// 戻り値: ボーナス配当 (nil = 通常配当)
	BonusEval func(hand *BlackJackHand, dealerUpcard *Card) *BJBonusPayout
}

// SpanishDeckValues はスパニッシュデッキ (10を除く) のカード値リスト
// (read-only — 内容を変更しないこと。Goにconst sliceがないためvarで宣言)
var SpanishDeckValues = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12, 13}

// SpanishDeckCardCount はスパニッシュデッキ1組の枚数 (12ランク × 4スート = 48)
const SpanishDeckCardCount = 48

// NewTrumpCardsSpanish21 はスパニッシュ21用のシュー (10を除いた48枚×deckCount組) を生成する
func NewTrumpCardsSpanish21(deckCount int) *TrumpCards {
	if deckCount <= 0 {
		deckCount = 1
	}
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	totalCards := SpanishDeckCardCount * deckCount

	t := new(TrumpCards)
	t.deckCnt = totalCards
	t.deck = make([]*Card, 0, totalCards)
	for range deckCount {
		for _, suit := range suits {
			for _, val := range SpanishDeckValues {
				t.deck = append(t.deck, NewCard(suit, val, false))
			}
		}
	}
	t.deckInit()
	return t
}

// Spanish21Variant はスパニッシュ21バリアントの設定を返す
func Spanish21Variant() *BlackJackVariantConfig {
	return &BlackJackVariantConfig{
		Name:                  BJVariantSpanish21,
		DeckBuilder:           NewTrumpCardsSpanish21,
		Player21AlwaysWins:    true,
		PlayerBJBeatsDealerBJ: true,
		BonusEval:             spanish21BonusEval,
	}
}

// ResolveBlackJackVariant はバリアント名から設定を復元する
// (UnmarshalJSON でバリアント情報を再構築するために使用)
func ResolveBlackJackVariant(name BlackJackVariantName) *BlackJackVariantConfig {
	switch name {
	case BJVariantSpanish21:
		return Spanish21Variant()
	default:
		return nil
	}
}

// spanish21BonusEval はスパニッシュ21のボーナス配当を判定する
// 21でない、またはボーナス対象外の場合は nil を返す
func spanish21BonusEval(hand *BlackJackHand, _ *Card) *BJBonusPayout {
	if hand == nil || hand.GetScore() != 21 {
		return nil
	}
	cards := hand.GetCards()
	n := len(cards)

	// 3枚構成のトリオボーナス (6-7-8、7-7-7) を優先判定
	if n == 3 {
		if isSixSevenEight(cards) {
			return spanishSuitBonus("678", cards)
		}
		if isTripleSevens(cards) {
			return spanishSuitBonus("777", cards)
		}
	}

	// 5枚以上の21ボーナス
	switch {
	case n == 5:
		return &BJBonusPayout{NameKey: "spanish21.bonus.fivecard21", MultiplierNum: 3, MultiplierDen: 2}
	case n == 6:
		return &BJBonusPayout{NameKey: "spanish21.bonus.sixcard21", MultiplierNum: 2, MultiplierDen: 1}
	case n >= 7:
		return &BJBonusPayout{NameKey: "spanish21.bonus.sevencard21", MultiplierNum: 3, MultiplierDen: 1}
	}

	return nil
}

// isSixSevenEight は手札が 6・7・8 の3枚構成か (順不同) 判定する
// 呼び出し元 (spanish21BonusEval) は BlackJackHand.GetCards() の結果を渡すため、
// cards に nil 要素は含まれない前提。
func isSixSevenEight(cards []*Card) bool {
	if len(cards) != 3 {
		return false
	}
	seen := make(map[int]bool, 3)
	for _, c := range cards {
		seen[c.GetValue()] = true
	}
	return seen[6] && seen[7] && seen[8]
}

// isTripleSevens は手札が 7-7-7 の3枚構成か判定する
// 呼び出し元の前提は isSixSevenEight と同じ (nil 要素なし)。
func isTripleSevens(cards []*Card) bool {
	if len(cards) != 3 {
		return false
	}
	for _, c := range cards {
		if c.GetValue() != 7 {
			return false
		}
	}
	return true
}

// spanishSuitBonus はスート構成からトリオボーナス (mixed/samesuit/spade) を決定する
//   - 全スペード     → 3:1
//   - 同一スート     → 2:1
//   - 異なるスート   → 3:2
func spanishSuitBonus(typeKey string, cards []*Card) *BJBonusPayout {
	suits := make(map[int]bool, 3)
	for _, c := range cards {
		suits[c.GetDesign()] = true
	}
	prefix := "spanish21.bonus." + typeKey
	if len(suits) == 1 {
		for s := range suits {
			if s == CardDesignSpade {
				return &BJBonusPayout{NameKey: prefix + ".spade", MultiplierNum: 3, MultiplierDen: 1}
			}
		}
		return &BJBonusPayout{NameKey: prefix + ".samesuit", MultiplierNum: 2, MultiplierDen: 1}
	}
	return &BJBonusPayout{NameKey: prefix + ".mixed", MultiplierNum: 3, MultiplierDen: 2}
}
