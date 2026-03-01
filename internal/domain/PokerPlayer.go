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
	Player                       // 親クラス
	handRank      int            // ハンドランク
	chips         int            // チップ
	isHuman       bool           // 人間フラグ
	folded        bool           // フォールド済
	allIn         bool           // オールイン済
	currentBet    int            // 現ラウンドベット額
	playStyle     PokerPlayStyle // CPUプレイスタイル
	exchangeCount int            // カード交換枚数
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

// AddCard カード追加
func (pp *PokerPlayer) AddCard(card *Card) {
	pp.cards = append(pp.cards, card)
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

// GetHandRank ハンドランク取得
func (pp *PokerPlayer) GetHandRank() int {
	return pp.handRank
}

// GetHandName ハンド名取得
func (pp *PokerPlayer) GetHandName() string {
	if 0 <= pp.handRank && pp.handRank < len(PokerHandNames) {
		return PokerHandNames[pp.handRank]
	}
	return "Unknown"
}

// GetChips チップ取得
func (pp *PokerPlayer) GetChips() int {
	return pp.chips
}

// SetChips チップ設定
func (pp *PokerPlayer) SetChips(chips int) {
	pp.chips = chips
}

// AddChips チップ追加
func (pp *PokerPlayer) AddChips(amount int) {
	pp.chips += amount
}

// SubtractChips チップ減算 (不足時はfalseを返す)
func (pp *PokerPlayer) SubtractChips(amount int) bool {
	if pp.chips < amount {
		return false
	}
	pp.chips -= amount
	return true
}

// GetIsHuman 人間フラグ取得
func (pp *PokerPlayer) GetIsHuman() bool { return pp.isHuman }

// GetFolded フォールド状態取得
func (pp *PokerPlayer) GetFolded() bool { return pp.folded }

// SetFolded フォールド状態設定
func (pp *PokerPlayer) SetFolded(folded bool) { pp.folded = folded }

// GetAllIn オールイン状態取得
func (pp *PokerPlayer) GetAllIn() bool { return pp.allIn }

// SetAllIn オールイン状態設定
func (pp *PokerPlayer) SetAllIn(allIn bool) { pp.allIn = allIn }

// GetCurrentBet 現ラウンドベット取得
func (pp *PokerPlayer) GetCurrentBet() int { return pp.currentBet }

// SetCurrentBet 現ラウンドベット設定
func (pp *PokerPlayer) SetCurrentBet(bet int) { pp.currentBet = bet }

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

// SetHandRank ハンドランク設定（テスト用）
func (pp *PokerPlayer) SetHandRank(rank int) { pp.handRank = rank }

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
