package domain

// バカラサイドベット種類定数
const (
	BacSideBetPlayerPair = 1
	BacSideBetBankerPair = 2
)

// バカラペア結果定数
const (
	BacPairNone  = 0
	BacPairMatch = 1
)

// バカラペア配当倍率
const BacPairPayoutRate = 11

// BacSideBetResult バカラサイドベット結果
type BacSideBetResult struct {
	BetType    int
	ResultType int
	ResultName string
	BetAmount  int
	Payout     int
}

// BetTypeName ベット種別名を返す
func (r *BacSideBetResult) BetTypeName() string {
	switch r.BetType {
	case BacSideBetPlayerPair:
		return "Player Pair"
	case BacSideBetBankerPair:
		return "Banker Pair"
	default:
		return "Unknown"
	}
}

// EvaluateBaccaratPair 2枚のカードがペアかを判定する
func EvaluateBaccaratPair(card1, card2 *Card) (int, string) {
	if isSameValue(card1, card2) {
		return BacPairMatch, "Pair"
	}
	return BacPairNone, ""
}

// BacPairPayout ペア結果に応じた配当倍率を返す
func BacPairPayout(resultType int) int {
	if resultType == BacPairMatch {
		return BacPairPayoutRate
	}
	return 0
}
