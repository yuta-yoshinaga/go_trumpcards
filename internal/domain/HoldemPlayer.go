//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"sort"
)

// HoldemPlayer テキサスホールデムプレイヤークラス
type HoldemPlayer struct {
	Player                              // 親クラス
	ChipHolder                          // チップ管理
	bettingPlayerBase                   // ベッティング共通状態
	isHuman             bool            // 人間フラグ
	bestHand            []*Card         // ベスト5枚
	playStyle           HoldemPlayStyle // CPUプレイスタイル
	totalHands          int             // 総ハンド数 (セッション通算)
	vpipCount           int             // VPIP対象ハンド数
	pfrCount            int             // PFR対象ハンド数
	threeBetOpportunity int             // 3Bet機会数
	threeBetCount       int             // 3Bet実行数
	postFlopBetRaise    int             // ポストフロップ ベット+レイズ回数
	postFlopCall        int             // ポストフロップ コール回数
}

// NewHoldemPlayer コンストラクタ
func NewHoldemPlayer(isHuman bool, style HoldemPlayStyle) *HoldemPlayer {
	return &HoldemPlayer{
		Player:    Player{cards: make([]*Card, 0)},
		isHuman:   isHuman,
		playStyle: style,
	}
}

// GetIsHuman 人間フラグ取得
func (hp *HoldemPlayer) GetIsHuman() bool { return hp.isHuman }

// GetBestHand ベストハンド取得
func (hp *HoldemPlayer) GetBestHand() []*Card { return hp.bestHand }

// GetPlayStyle プレイスタイル取得
func (hp *HoldemPlayer) GetPlayStyle() HoldemPlayStyle { return hp.playStyle }

// GetPlayStyleName プレイスタイル名取得
func (hp *HoldemPlayer) GetPlayStyleName() string {
	return playStyleName(int(hp.playStyle), HoldemPlayStyleNames)
}

// GetTotalHands 総ハンド数取得
func (hp *HoldemPlayer) GetTotalHands() int { return hp.totalHands }

// GetVPIPCount VPIP対象ハンド数取得
func (hp *HoldemPlayer) GetVPIPCount() int { return hp.vpipCount }

// GetPFRCount PFR対象ハンド数取得
func (hp *HoldemPlayer) GetPFRCount() int { return hp.pfrCount }

// GetVPIP VPIP%取得 (0 if totalHands==0)
func (hp *HoldemPlayer) GetVPIP() int {
	return percentOf(hp.vpipCount, hp.totalHands)
}

// GetPFR PFR%取得 (0 if totalHands==0)
func (hp *HoldemPlayer) GetPFR() int {
	return percentOf(hp.pfrCount, hp.totalHands)
}

// IncrementTotalHands 総ハンド数をインクリメント
func (hp *HoldemPlayer) IncrementTotalHands() { hp.totalHands++ }

// IncrementVPIP VPIP対象ハンド数をインクリメント
func (hp *HoldemPlayer) IncrementVPIP() { hp.vpipCount++ }

// IncrementPFR PFR対象ハンド数をインクリメント
func (hp *HoldemPlayer) IncrementPFR() { hp.pfrCount++ }

// GetThreeBetOpportunity 3Bet機会数取得
func (hp *HoldemPlayer) GetThreeBetOpportunity() int { return hp.threeBetOpportunity }

// GetThreeBetCount 3Bet実行数取得
func (hp *HoldemPlayer) GetThreeBetCount() int { return hp.threeBetCount }

// GetThreeBet 3Bet%取得 (0 if threeBetOpportunity==0)
func (hp *HoldemPlayer) GetThreeBet() int {
	return percentOf(hp.threeBetCount, hp.threeBetOpportunity)
}

// IncrementThreeBetOpportunity 3Bet機会数をインクリメント
func (hp *HoldemPlayer) IncrementThreeBetOpportunity() { hp.threeBetOpportunity++ }

// IncrementThreeBet 3Bet実行数をインクリメント
func (hp *HoldemPlayer) IncrementThreeBet() { hp.threeBetCount++ }

// GetPostFlopBetRaise ポストフロップ ベット+レイズ回数取得
func (hp *HoldemPlayer) GetPostFlopBetRaise() int { return hp.postFlopBetRaise }

// GetPostFlopCall ポストフロップ コール回数取得
func (hp *HoldemPlayer) GetPostFlopCall() int { return hp.postFlopCall }

// IncrementPostFlopBetRaise ポストフロップ ベット+レイズ回数をインクリメント
func (hp *HoldemPlayer) IncrementPostFlopBetRaise() { hp.postFlopBetRaise++ }

// IncrementPostFlopCall ポストフロップ コール回数をインクリメント
func (hp *HoldemPlayer) IncrementPostFlopCall() { hp.postFlopCall++ }

// GetAFDisplay AF表示文字列取得 ("-"=アクションなし, "∞"=コールなし, "X.X"=通常)
func (hp *HoldemPlayer) GetAFDisplay() string {
	return afDisplay(hp.postFlopBetRaise, hp.postFlopCall)
}

// GetComparisonCards ハンド比較用カード取得 (BettingPlayerインターフェース)
func (hp *HoldemPlayer) GetComparisonCards() []*Card {
	return copyOf(hp.bestHand)
}

// holdemPlayerJSON is the JSON wire format for HoldemPlayer.
type holdemPlayerJSON struct {
	Player              *Player            `json:"p"`
	ChipHolder          *ChipHolder        `json:"ch"`
	BettingPlayerBase   *bettingPlayerBase `json:"bp"`
	IsHuman             bool               `json:"ih"`
	BestHand            []*Card            `json:"bh"`
	PlayStyle           HoldemPlayStyle    `json:"ps"`
	TotalHands          int                `json:"th"`
	VPIPCount           int                `json:"vc"`
	PFRCount            int                `json:"pc"`
	ThreeBetOpportunity int                `json:"to"`
	ThreeBetCount       int                `json:"tc"`
	PostFlopBetRaise    int                `json:"pb"`
	PostFlopCall        int                `json:"pf"`
}

// MarshalJSON implements json.Marshaler.
func (hp *HoldemPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(holdemPlayerJSON{
		Player:              &hp.Player,
		ChipHolder:          &hp.ChipHolder,
		BettingPlayerBase:   &hp.bettingPlayerBase,
		IsHuman:             hp.isHuman,
		BestHand:            hp.bestHand,
		PlayStyle:           hp.playStyle,
		TotalHands:          hp.totalHands,
		VPIPCount:           hp.vpipCount,
		PFRCount:            hp.pfrCount,
		ThreeBetOpportunity: hp.threeBetOpportunity,
		ThreeBetCount:       hp.threeBetCount,
		PostFlopBetRaise:    hp.postFlopBetRaise,
		PostFlopCall:        hp.postFlopCall,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (hp *HoldemPlayer) UnmarshalJSON(data []byte) error {
	var j holdemPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Player != nil {
		hp.Player = *j.Player
	}
	if j.ChipHolder != nil {
		hp.ChipHolder = *j.ChipHolder
	}
	if j.BettingPlayerBase != nil {
		hp.bettingPlayerBase = *j.BettingPlayerBase
	}
	hp.isHuman = j.IsHuman
	hp.bestHand = j.BestHand
	hp.playStyle = j.PlayStyle
	hp.totalHands = j.TotalHands
	hp.vpipCount = j.VPIPCount
	hp.pfrCount = j.PFRCount
	hp.threeBetOpportunity = j.ThreeBetOpportunity
	hp.threeBetCount = j.ThreeBetCount
	hp.postFlopBetRaise = j.PostFlopBetRaise
	hp.postFlopCall = j.PostFlopCall
	return nil
}

// EvalBestHand コミュニティカードとホールカードからベスト5枚を評価
func (hp *HoldemPlayer) EvalBestHand(communityCards []*Card) int {
	all := make([]*Card, 0, len(hp.cards)+len(communityCards))
	all = append(all, hp.cards...)
	all = append(all, communityCards...)

	if len(all) < 5 {
		hp.handRank = PokerHandHighCard
		hp.bestHand = nil
		return hp.handRank
	}

	combos := combinations(all, 5)
	bestRank := -1
	var bestCards []*Card

	for _, combo := range combos {
		rank := evalFiveCardHand(combo)
		if rank > bestRank || (rank == bestRank && compareHighCardsSlice(combo, bestCards) > 0) {
			bestRank = rank
			bestCards = make([]*Card, 5)
			copy(bestCards, combo)
		}
	}

	hp.handRank = bestRank
	hp.bestHand = bestCards
	return hp.handRank
}

// combinations n枚からk枚を選ぶ全組み合わせを返す
func combinations(cards []*Card, k int) [][]*Card {
	var result [][]*Card
	n := len(cards)
	if k > n {
		return result
	}
	combo := make([]int, k)
	var generate func(start, idx int)
	generate = func(start, idx int) {
		if idx == k {
			hand := make([]*Card, k)
			for i, ci := range combo {
				hand[i] = cards[ci]
			}
			result = append(result, hand)
			return
		}
		for i := start; i <= n-(k-idx); i++ {
			combo[idx] = i
			generate(i+1, idx+1)
		}
	}
	generate(0, 0)
	return result
}

// isWheelHand ホイール (A-2-3-4-5) かどうか判定
func isWheelHand(cards []*Card) bool {
	if len(cards) != 5 {
		return false
	}
	vals := make([]int, 5)
	for i, c := range cards {
		vals[i] = c.GetValue()
	}
	sort.Ints(vals)
	return vals[0] == 1 && vals[1] == 2 && vals[2] == 3 && vals[3] == 4 && vals[4] == 5
}

// tieBreakValues カード値リストを (出現回数 DESC, 値 DESC) でソートしたユニーク値リストを返す
// ペア系ハンドの正しいタイブレーク順序を保証する
func tieBreakValues(vals []int) []int {
	freq := make(map[int]int)
	for _, v := range vals {
		freq[v]++
	}
	unique := make([]int, 0, len(freq))
	for v := range freq {
		unique = append(unique, v)
	}
	sort.Slice(unique, func(i, j int) bool {
		if freq[unique[i]] != freq[unique[j]] {
			return freq[unique[i]] > freq[unique[j]]
		}
		return unique[i] > unique[j]
	})
	return unique
}

// compareHighCardsSlice 2つの5枚ハンドのハイカード比較 (a > b: 1, a < b: -1, a == b: 0)
func compareHighCardsSlice(a, b []*Card) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	aWheel := isWheelHand(a)
	bWheel := isWheelHand(b)
	aVals := make([]int, len(a))
	bVals := make([]int, len(b))
	for i, c := range a {
		v := c.GetValue()
		if v == 1 && !aWheel {
			v = 14
		}
		aVals[i] = v
	}
	for i, c := range b {
		v := c.GetValue()
		if v == 1 && !bWheel {
			v = 14
		}
		bVals[i] = v
	}
	aTB := tieBreakValues(aVals)
	bTB := tieBreakValues(bVals)
	for i := 0; i < len(aTB) && i < len(bTB); i++ {
		if aTB[i] > bTB[i] {
			return 1
		}
		if aTB[i] < bTB[i] {
			return -1
		}
	}
	return 0
}
