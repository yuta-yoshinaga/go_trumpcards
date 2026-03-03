package domain

// ポーカーハンドランク定数
const (
	PokerHandHighCard      = 0
	PokerHandOnePair       = 1
	PokerHandTwoPair       = 2
	PokerHandThreeOfAKind  = 3
	PokerHandStraight      = 4
	PokerHandFlush         = 5
	PokerHandFullHouse     = 6
	PokerHandFourOfAKind   = 7
	PokerHandStraightFlush = 8
	PokerHandRoyalFlush    = 9
	PokerHandFiveOfAKind   = 10
)

// PokerHandNames ポーカーハンド名
var PokerHandNames = []string{
	"High Card",
	"One Pair",
	"Two Pair",
	"Three of a Kind",
	"Straight",
	"Flush",
	"Full House",
	"Four of a Kind",
	"Straight Flush",
	"Royal Flush",
	"Five of a Kind",
}

// PokerPlayer ポーカープレイヤークラス
type PokerPlayer struct {
	Player                           // 親クラス
	ChipHolder                       // チップ管理
	bettingPlayerBase                // ベッティング共通状態
	isHuman           bool           // 人間フラグ
	playStyle         PokerPlayStyle // CPUプレイスタイル
	exchangeCount     int            // カード交換枚数
}

// NewPokerPlayer コンストラクタ
func NewPokerPlayer(isHuman bool, style PokerPlayStyle) *PokerPlayer {
	return &PokerPlayer{
		Player: Player{
			cards: make([]*Card, 0),
		},
		isHuman:   isHuman,
		playStyle: style,
	}
}

// ExchangeCard 指定インデックスのカードを交換
func (pp *PokerPlayer) ExchangeCard(idx int, card *Card) {
	if 0 <= idx && idx < len(pp.cards) {
		pp.cards[idx] = card
	}
}

// EvalHand ハンド評価
func (pp *PokerPlayer) EvalHand() int {
	pp.handRank = evalFiveCardHandWithJokers(pp.cards)
	return pp.handRank
}

// GetHandName ハンド名取得
func (pp *PokerPlayer) GetHandName() string {
	if 0 <= pp.handRank && pp.handRank < len(PokerHandNames) {
		return PokerHandNames[pp.handRank]
	}
	return "Unknown"
}

// GetIsHuman 人間フラグ取得
func (pp *PokerPlayer) GetIsHuman() bool { return pp.isHuman }

// GetPlayStyle プレイスタイル取得
func (pp *PokerPlayer) GetPlayStyle() PokerPlayStyle { return pp.playStyle }

// GetPlayStyleName プレイスタイル名取得
func (pp *PokerPlayer) GetPlayStyleName() string {
	if int(pp.playStyle) < len(PokerPlayStyleNames) {
		return PokerPlayStyleNames[pp.playStyle]
	}
	return "Unknown"
}

// GetExchangeCount 交換枚数取得
func (pp *PokerPlayer) GetExchangeCount() int { return pp.exchangeCount }

// SetExchangeCount 交換枚数設定
func (pp *PokerPlayer) SetExchangeCount(count int) { pp.exchangeCount = count }

// GetComparisonCards ハンド比較用カード取得 (BettingPlayerインターフェース)
func (pp *PokerPlayer) GetComparisonCards() []*Card {
	cards := make([]*Card, len(pp.cards))
	copy(cards, pp.cards)
	return cards
}

// evalFiveCardHandWithJokers ジョーカー対応の5枚ハンド評価
// ジョーカーがない場合は通常のevalFiveCardHandを呼ぶ
// ジョーカーがある場合は全52通りの代替を試してベストランクを返す
func evalFiveCardHandWithJokers(cards []*Card) int {
	if len(cards) != 5 {
		return PokerHandHighCard
	}

	// ジョーカーの位置を探す
	jokerIndices := make([]int, 0)
	for i, c := range cards {
		if c.GetDesign() == CardDesignJoker {
			jokerIndices = append(jokerIndices, i)
		}
	}

	if len(jokerIndices) == 0 {
		return evalFiveCardHand(cards)
	}

	// 代替カード候補 (4スート × 13値 = 52枚)
	bestRank := PokerHandHighCard
	substitutions := make([]*Card, 0, 52)
	for d := CardDesignSpade; d <= CardDesignDiamond; d++ {
		for v := 1; v <= CardValueMax; v++ {
			substitutions = append(substitutions, NewCard(d, v, false))
		}
	}

	if len(jokerIndices) == 1 {
		// ジョーカー1枚: 52通り試す
		idx := jokerIndices[0]
		original := cards[idx]
		for _, sub := range substitutions {
			cards[idx] = sub
			rank := evalFiveCardHand(cards)
			if rank > bestRank {
				bestRank = rank
			}
		}
		// 元のジョーカーポインタを復元
		cards[idx] = original
	} else {
		// ジョーカー2枚: 52×52通り試す
		idx0 := jokerIndices[0]
		idx1 := jokerIndices[1]
		original0 := cards[idx0]
		original1 := cards[idx1]
		for _, sub0 := range substitutions {
			cards[idx0] = sub0
			for _, sub1 := range substitutions {
				cards[idx1] = sub1
				rank := evalFiveCardHand(cards)
				if rank > bestRank {
					bestRank = rank
				}
			}
		}
		// 元のジョーカーポインタを復元
		cards[idx0] = original0
		cards[idx1] = original1
	}

	// FiveOfAKind判定: ジョーカーがある場合のみ可能
	// FourOfAKind + ジョーカー → FiveOfAKind
	if bestRank >= PokerHandFourOfAKind && bestRank < PokerHandFiveOfAKind {
		// ジョーカー以外のカードで4枚同数値をチェック
		nonJokerCards := make([]*Card, 0, 5)
		for i, c := range cards {
			isJoker := false
			for _, ji := range jokerIndices {
				if i == ji {
					isJoker = true
					break
				}
			}
			if !isJoker {
				nonJokerCards = append(nonJokerCards, c)
			}
		}
		if checkFiveOfAKind(nonJokerCards, len(jokerIndices)) {
			bestRank = PokerHandFiveOfAKind
		}
	}

	return bestRank
}

// checkFiveOfAKind ジョーカーを含めてFiveOfAKindが成立するか判定
func checkFiveOfAKind(nonJokerCards []*Card, jokerCount int) bool {
	valueCounts := make(map[int]int)
	for _, c := range nonJokerCards {
		valueCounts[c.GetValue()]++
	}
	for _, count := range valueCounts {
		if count+jokerCount >= 5 {
			return true
		}
	}
	return false
}
