//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

// SevenCardStudPlayer セブンカードスタッドプレイヤー
type SevenCardStudPlayer struct {
	Player                                     // 親クラス (cardsフィールドは使用しない)
	ChipHolder                                 // チップ管理
	bettingPlayerBase                          // ベッティング共通状態
	isHuman             bool                   // 人間フラグ
	holeCards           []*Card                // 伏せ札 (最大3枚: 1st, 2nd, 7th)
	doorCards           []*Card                // 表向き札 (最大4枚: 3rd-6th Street)
	bestHand            []*Card                // ベスト5枚
	lowQualifies        bool                   // 8-or-better のローが成立したか (Hi-Lo のみ)
	lowBestHand         []*Card                // ローのベスト5枚 (Hi-Lo のみ)
	playStyle           SevenCardStudPlayStyle // CPUプレイスタイル
	totalHands          int                    // 総ハンド数 (セッション通算)
	vpipCount           int                    // VPIP対象ハンド数
	pfrCount            int                    // PFR対象ハンド数
	threeBetOpportunity int                    // 3Bet機会数
	threeBetCount       int                    // 3Bet実行数
	postFlopBetRaise    int                    // ポストフロップ ベット+レイズ回数
	postFlopCall        int                    // ポストフロップ コール回数
}

// NewSevenCardStudPlayer コンストラクタ
func NewSevenCardStudPlayer(isHuman bool, style SevenCardStudPlayStyle) *SevenCardStudPlayer {
	return &SevenCardStudPlayer{
		Player:    Player{cards: make([]*Card, 0)},
		isHuman:   isHuman,
		holeCards: make([]*Card, 0, 3),
		doorCards: make([]*Card, 0, 4),
		playStyle: style,
	}
}

// GetIsHuman 人間フラグ取得
func (p *SevenCardStudPlayer) GetIsHuman() bool { return p.isHuman }

// GetHoleCards 伏せ札取得
func (p *SevenCardStudPlayer) GetHoleCards() []*Card { return p.holeCards }

// GetDoorCards 表向き札取得
func (p *SevenCardStudPlayer) GetDoorCards() []*Card { return p.doorCards }

// GetBestHand ベストハンド取得
func (p *SevenCardStudPlayer) GetBestHand() []*Card { return p.bestHand }

// GetPlayStyle プレイスタイル取得
func (p *SevenCardStudPlayer) GetPlayStyle() SevenCardStudPlayStyle { return p.playStyle }

// GetPlayStyleName プレイスタイル名取得
func (p *SevenCardStudPlayer) GetPlayStyleName() string {
	return playStyleName(int(p.playStyle), SevenCardStudPlayStyleNames)
}

// AddHoleCard 伏せ札を追加する
func (p *SevenCardStudPlayer) AddHoleCard(c *Card) {
	p.holeCards = append(p.holeCards, c)
}

// AddDoorCard 表向き札を追加する
func (p *SevenCardStudPlayer) AddDoorCard(c *Card) {
	p.doorCards = append(p.doorCards, c)
}

// GetAllCards 全カード取得 (伏せ札+表向き札)
func (p *SevenCardStudPlayer) GetAllCards() []*Card {
	all := make([]*Card, 0, len(p.holeCards)+len(p.doorCards))
	all = append(all, p.holeCards...)
	all = append(all, p.doorCards...)
	return all
}

// ClearCards 全カードをクリアする
func (p *SevenCardStudPlayer) ClearCards() {
	p.holeCards = p.holeCards[:0]
	p.doorCards = p.doorCards[:0]
	p.bestHand = nil
}

// GetTotalHands 総ハンド数取得
func (p *SevenCardStudPlayer) GetTotalHands() int { return p.totalHands }

// GetVPIPCount VPIP対象ハンド数取得
func (p *SevenCardStudPlayer) GetVPIPCount() int { return p.vpipCount }

// GetPFRCount PFR対象ハンド数取得
func (p *SevenCardStudPlayer) GetPFRCount() int { return p.pfrCount }

// GetVPIP VPIP%取得 (0 if totalHands==0)
func (p *SevenCardStudPlayer) GetVPIP() int {
	if p.totalHands == 0 {
		return 0
	}
	return p.vpipCount * 100 / p.totalHands
}

// GetPFR PFR%取得 (0 if totalHands==0)
func (p *SevenCardStudPlayer) GetPFR() int {
	if p.totalHands == 0 {
		return 0
	}
	return p.pfrCount * 100 / p.totalHands
}

// IncrementTotalHands 総ハンド数をインクリメント
func (p *SevenCardStudPlayer) IncrementTotalHands() { p.totalHands++ }

// IncrementVPIP VPIP対象ハンド数をインクリメント
func (p *SevenCardStudPlayer) IncrementVPIP() { p.vpipCount++ }

// IncrementPFR PFR対象ハンド数をインクリメント
func (p *SevenCardStudPlayer) IncrementPFR() { p.pfrCount++ }

// GetThreeBetOpportunity 3Bet機会数取得
func (p *SevenCardStudPlayer) GetThreeBetOpportunity() int { return p.threeBetOpportunity }

// GetThreeBetCount 3Bet実行数取得
func (p *SevenCardStudPlayer) GetThreeBetCount() int { return p.threeBetCount }

// GetThreeBet 3Bet%取得 (0 if threeBetOpportunity==0)
func (p *SevenCardStudPlayer) GetThreeBet() int {
	if p.threeBetOpportunity == 0 {
		return 0
	}
	return p.threeBetCount * 100 / p.threeBetOpportunity
}

// IncrementThreeBetOpportunity 3Bet機会数をインクリメント
func (p *SevenCardStudPlayer) IncrementThreeBetOpportunity() { p.threeBetOpportunity++ }

// IncrementThreeBet 3Bet実行数をインクリメント
func (p *SevenCardStudPlayer) IncrementThreeBet() { p.threeBetCount++ }

// GetPostFlopBetRaise ポストフロップ ベット+レイズ回数取得
func (p *SevenCardStudPlayer) GetPostFlopBetRaise() int { return p.postFlopBetRaise }

// GetPostFlopCall ポストフロップ コール回数取得
func (p *SevenCardStudPlayer) GetPostFlopCall() int { return p.postFlopCall }

// IncrementPostFlopBetRaise ポストフロップ ベット+レイズ回数をインクリメント
func (p *SevenCardStudPlayer) IncrementPostFlopBetRaise() { p.postFlopBetRaise++ }

// IncrementPostFlopCall ポストフロップ コール回数をインクリメント
func (p *SevenCardStudPlayer) IncrementPostFlopCall() { p.postFlopCall++ }

// GetAFDisplay AF表示文字列取得 ("-"=アクションなし, "∞"=コールなし, "X.X"=通常)
func (p *SevenCardStudPlayer) GetAFDisplay() string {
	if p.postFlopBetRaise == 0 && p.postFlopCall == 0 {
		return "-"
	}
	if p.postFlopCall == 0 {
		return "∞"
	}
	return fmt.Sprintf("%.1f", float64(p.postFlopBetRaise)/float64(p.postFlopCall))
}

// GetComparisonCards ハンド比較用カード取得 (BettingPlayerインターフェース)
func (p *SevenCardStudPlayer) GetComparisonCards() []*Card {
	cards := make([]*Card, len(p.bestHand))
	copy(cards, p.bestHand)
	return cards
}

// EvalBestHand 全7枚からベスト5枚を評価
func (p *SevenCardStudPlayer) EvalBestHand() int {
	all := p.GetAllCards()

	if len(all) < 5 {
		p.handRank = PokerHandHighCard
		p.bestHand = nil
		return p.handRank
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

	p.handRank = bestRank
	p.bestHand = bestCards
	return p.handRank
}

// EvalBestHandRazz 全7枚からA-5ローボール(ラズ)用ベスト5枚を評価
// ストレート・フラッシュは無視し、最も低いハンドを選択する
func (p *SevenCardStudPlayer) EvalBestHandRazz() int {
	all := p.GetAllCards()

	if len(all) < 5 {
		p.handRank = PokerHandHighCard
		p.bestHand = nil
		return p.handRank
	}

	combos := combinations(all, 5)
	bestRank := -1
	var bestCards []*Card

	for _, combo := range combos {
		rank := evalRazzHand(combo)
		if bestRank == -1 || rank < bestRank || (rank == bestRank && compareRazzCards(combo, bestCards) < 0) {
			bestRank = rank
			bestCards = make([]*Card, 5)
			copy(bestCards, combo)
		}
	}

	p.handRank = bestRank
	p.bestHand = bestCards
	return p.handRank
}

// SevenCardStudRazzBestLow returns the strongest 5-card Razz low from cards and
// whether it is complete. With fewer than 5 cards the low cannot be made yet, so
// the cards are returned sorted ascending (Ace low) for a progress display and
// complete is false. This is a pure, non-mutating read used by the CUI. It lives
// here (not hand_eval.go) because it depends on combinations, which only ships in
// the casino worker build.
func SevenCardStudRazzBestLow(cards []*Card) (best []*Card, complete bool) {
	if len(cards) < 5 {
		sorted := make([]*Card, len(cards))
		copy(sorted, cards)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].GetValue() < sorted[j].GetValue()
		})
		return sorted, false
	}
	bestRank := -1
	for _, combo := range combinations(cards, 5) {
		rank := evalRazzHand(combo)
		if bestRank == -1 || rank < bestRank || (rank == bestRank && compareRazzCards(combo, best) < 0) {
			bestRank = rank
			best = append([]*Card(nil), combo...)
		}
	}
	return best, true
}

// EvalVisibleHand 表向き札のみからハンドを評価 (ベッティング順序決定用)
// 5枚未満の場合はハイカード比較用にソートした値を返す
func (p *SevenCardStudPlayer) EvalVisibleHand() int {
	if len(p.doorCards) == 0 {
		return PokerHandHighCard
	}
	if len(p.doorCards) >= 5 {
		return evalFiveCardHand(p.doorCards[:5])
	}
	// 5枚未満: 簡易評価 (ペア検出)
	return evalPartialHand(p.doorCards)
}

// evalPartialHand 5枚未満のカードを簡易評価する
func evalPartialHand(cards []*Card) int {
	if len(cards) == 0 {
		return PokerHandHighCard
	}
	freq := make(map[int]int)
	for _, c := range cards {
		freq[c.GetValue()]++
	}
	maxFreq := 0
	for _, f := range freq {
		if f > maxFreq {
			maxFreq = f
		}
	}
	switch {
	case maxFreq >= 4:
		return PokerHandFourOfAKind
	case maxFreq == 3:
		// スリーカードまたはフルハウス
		pairCount := 0
		for _, f := range freq {
			if f == 2 {
				pairCount++
			}
		}
		if pairCount > 0 {
			return PokerHandFullHouse
		}
		return PokerHandThreeOfAKind
	case maxFreq == 2:
		pairCount := 0
		for _, f := range freq {
			if f == 2 {
				pairCount++
			}
		}
		if pairCount >= 2 {
			return PokerHandTwoPair
		}
		return PokerHandOnePair
	default:
		return PokerHandHighCard
	}
}

// CompareVisibleHands 2人のプレイヤーの表向き札を比較する (a > b: 1, a < b: -1, a == b: 0)
func CompareVisibleHands(a, b *SevenCardStudPlayer) int {
	rankA := a.EvalVisibleHand()
	rankB := b.EvalVisibleHand()
	if rankA > rankB {
		return 1
	}
	if rankA < rankB {
		return -1
	}
	// 同ランク: ハイカード比較
	aVals := sortedDoorCardValues(a.doorCards)
	bVals := sortedDoorCardValues(b.doorCards)
	for i := 0; i < len(aVals) && i < len(bVals); i++ {
		if aVals[i] > bVals[i] {
			return 1
		}
		if aVals[i] < bVals[i] {
			return -1
		}
	}
	return 0
}

// CompareVisibleHandsLow 2人のプレイヤーの表向き札をローボール比較する (低い方が強い)
// a が強い(低い): 1, b が強い: -1, 同じ: 0
func CompareVisibleHandsLow(a, b *SevenCardStudPlayer) int {
	rankA := a.EvalVisibleHand()
	rankB := b.EvalVisibleHand()
	// ローボール: 低いランクが強い
	if rankA < rankB {
		return 1
	}
	if rankA > rankB {
		return -1
	}
	// 同ランク: カード値が低い方が強い (Ace=1)
	aVals := sortedDoorCardValuesLow(a.doorCards)
	bVals := sortedDoorCardValuesLow(b.doorCards)
	for i := 0; i < len(aVals) && i < len(bVals); i++ {
		if aVals[i] < bVals[i] {
			return 1
		}
		if aVals[i] > bVals[i] {
			return -1
		}
	}
	return 0
}

// sortedDoorCardValuesLow 表向き札の値を降順ソートして返す (Ace=1, ローボール用)
func sortedDoorCardValuesLow(cards []*Card) []int {
	vals := make([]int, len(cards))
	for i, c := range cards {
		vals[i] = c.GetValue() // Ace stays as 1
	}
	sort.Sort(sort.Reverse(sort.IntSlice(vals)))
	return vals
}

// sortedDoorCardValues 表向き札の値をソートして返す (Ace=14)
func sortedDoorCardValues(cards []*Card) []int {
	vals := make([]int, len(cards))
	for i, c := range cards {
		v := c.GetValue()
		if v == 1 {
			v = 14
		}
		vals[i] = v
	}
	sort.Sort(sort.Reverse(sort.IntSlice(vals)))
	return vals
}

// SuitRank ブリングイン決定用のスートランキング (低い方がブリングイン)
// スペード=4(最高), ハート=3, ダイヤ=2, クラブ=1(最低)
func SuitRank(design int) int {
	switch design {
	case CardDesignSpade:
		return 4
	case CardDesignHeart:
		return 3
	case CardDesignDiamond:
		return 2
	case CardDesignClover:
		return 1
	default:
		return 0
	}
}

// EvalBestLowHandEightOrBetter は Hi-Lo (8 or Better) 用のローベスト5枚を
// 評価する。7 枚から C(7,5)=21 通りを見て、**5 枚すべて 8 以下・ランク重複なし**
// (Ace=1) を満たす中で最も低いものを採る。
//
// スタッドにコミュニティカードは無いので、オマハのような「手札から2枚・場から
// 3枚」の制約は付かない。素直に 7 枚から 5 枚を選ぶ。
//
// 判定そのものは isQualifyingOmahaLow を使う。名前はオマハだが中身は
// 「8 以下・ペア無し」という汎用の 8-or-better 判定で、ここで別実装を書くと
// 同じ規則が 2 箇所に散る。
//
// 戻り値: qualifying なローが見つかったかどうか。
func (p *SevenCardStudPlayer) EvalBestLowHandEightOrBetter() bool {
	p.lowQualifies = false
	p.lowBestHand = nil

	all := p.GetAllCards()
	if len(all) < 5 {
		return false
	}

	var bestCards []*Card
	for _, combo := range combinations(all, 5) {
		if !isQualifyingOmahaLow(combo) {
			continue
		}
		if bestCards == nil || compareRazzCards(combo, bestCards) < 0 {
			bestCards = make([]*Card, 5)
			copy(bestCards, combo)
		}
	}
	if bestCards == nil {
		return false
	}
	p.lowQualifies = true
	p.lowBestHand = bestCards
	return true
}

// GetLowQualifies は 8-or-better のローが成立したかを返す。
func (p *SevenCardStudPlayer) GetLowQualifies() bool { return p.lowQualifies }

// GetLowBestHand はローのベスト5枚を返す。成立していなければ nil。
func (p *SevenCardStudPlayer) GetLowBestHand() []*Card { return p.lowBestHand }

// sevenCardStudPlayerJSON is the JSON wire format for SevenCardStudPlayer.
type sevenCardStudPlayerJSON struct {
	Player              *Player                `json:"p"`
	ChipHolder          *ChipHolder            `json:"ch"`
	BettingPlayerBase   *bettingPlayerBase     `json:"bp"`
	IsHuman             bool                   `json:"ih"`
	HoleCards           []*Card                `json:"hc"`
	DoorCards           []*Card                `json:"dc"`
	BestHand            []*Card                `json:"bh"`
	LowQualifies        bool                   `json:"lq,omitempty"`
	LowBestHand         []*Card                `json:"lb,omitempty"`
	PlayStyle           SevenCardStudPlayStyle `json:"ps"`
	TotalHands          int                    `json:"th"`
	VPIPCount           int                    `json:"vc"`
	PFRCount            int                    `json:"pc"`
	ThreeBetOpportunity int                    `json:"to"`
	ThreeBetCount       int                    `json:"tc"`
	PostFlopBetRaise    int                    `json:"pb"`
	PostFlopCall        int                    `json:"pf"`
}

// MarshalJSON implements json.Marshaler.
func (p *SevenCardStudPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(sevenCardStudPlayerJSON{
		Player:              &p.Player,
		ChipHolder:          &p.ChipHolder,
		BettingPlayerBase:   &p.bettingPlayerBase,
		IsHuman:             p.isHuman,
		HoleCards:           p.holeCards,
		DoorCards:           p.doorCards,
		BestHand:            p.bestHand,
		LowQualifies:        p.lowQualifies,
		LowBestHand:         p.lowBestHand,
		PlayStyle:           p.playStyle,
		TotalHands:          p.totalHands,
		VPIPCount:           p.vpipCount,
		PFRCount:            p.pfrCount,
		ThreeBetOpportunity: p.threeBetOpportunity,
		ThreeBetCount:       p.threeBetCount,
		PostFlopBetRaise:    p.postFlopBetRaise,
		PostFlopCall:        p.postFlopCall,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SevenCardStudPlayer) UnmarshalJSON(data []byte) error {
	var j sevenCardStudPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Player != nil {
		p.Player = *j.Player
	}
	if j.ChipHolder != nil {
		p.ChipHolder = *j.ChipHolder
	}
	if j.BettingPlayerBase != nil {
		p.bettingPlayerBase = *j.BettingPlayerBase
	}
	p.isHuman = j.IsHuman
	p.holeCards = j.HoleCards
	if p.holeCards == nil {
		p.holeCards = make([]*Card, 0, 3)
	}
	p.doorCards = j.DoorCards
	if p.doorCards == nil {
		p.doorCards = make([]*Card, 0, 4)
	}
	p.bestHand = j.BestHand
	p.lowQualifies = j.LowQualifies
	p.lowBestHand = j.LowBestHand
	p.playStyle = j.PlayStyle
	p.totalHands = j.TotalHands
	p.vpipCount = j.VPIPCount
	p.pfrCount = j.PFRCount
	p.threeBetOpportunity = j.ThreeBetOpportunity
	p.threeBetCount = j.ThreeBetCount
	p.postFlopBetRaise = j.PostFlopBetRaise
	p.postFlopCall = j.PostFlopCall
	return nil
}
