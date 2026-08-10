//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"sort"
)

// FiveCardStudPlayer ファイブカードスタッドプレイヤー
type FiveCardStudPlayer struct {
	Player                                    // 親クラス (cardsフィールドは使用しない)
	ChipHolder                                // チップ管理
	bettingPlayerBase                         // ベッティング共通状態
	isHuman             bool                  // 人間フラグ
	holeCards           []*Card               // 伏せ札 (最大1枚: 1st)
	doorCards           []*Card               // 表向き札 (最大4枚: 2nd-5th Street)
	bestHand            []*Card               // ベスト5枚
	playStyle           FiveCardStudPlayStyle // CPUプレイスタイル
	totalHands          int                   // 総ハンド数 (セッション通算)
	vpipCount           int                   // VPIP対象ハンド数
	pfrCount            int                   // PFR対象ハンド数
	threeBetOpportunity int                   // 3Bet機会数
	threeBetCount       int                   // 3Bet実行数
	postFlopBetRaise    int                   // ポストフロップ ベット+レイズ回数
	postFlopCall        int                   // ポストフロップ コール回数
	sokoMode            bool                  // Soko の役序列で評価する (Canadian Stud)
}

// SetSokoMode は Soko (Canadian Stud) の役序列で評価するかを設定する。
// ゲーム側 (FiveCardStud) が構築時とリセット時に伝播する。
func (p *FiveCardStudPlayer) SetSokoMode(v bool) { p.sokoMode = v }

// GetSokoMode は Soko モードかを返す。
func (p *FiveCardStudPlayer) GetSokoMode() bool { return p.sokoMode }

// NewFiveCardStudPlayer コンストラクタ
func NewFiveCardStudPlayer(isHuman bool, style FiveCardStudPlayStyle) *FiveCardStudPlayer {
	return &FiveCardStudPlayer{
		Player:    Player{cards: make([]*Card, 0)},
		isHuman:   isHuman,
		holeCards: make([]*Card, 0, 1),
		doorCards: make([]*Card, 0, 4),
		playStyle: style,
	}
}

// GetIsHuman 人間フラグ取得
func (p *FiveCardStudPlayer) GetIsHuman() bool { return p.isHuman }

// GetHoleCards 伏せ札取得
func (p *FiveCardStudPlayer) GetHoleCards() []*Card { return p.holeCards }

// GetDoorCards 表向き札取得
func (p *FiveCardStudPlayer) GetDoorCards() []*Card { return p.doorCards }

// GetBestHand ベストハンド取得
func (p *FiveCardStudPlayer) GetBestHand() []*Card { return p.bestHand }

// GetPlayStyle プレイスタイル取得
func (p *FiveCardStudPlayer) GetPlayStyle() FiveCardStudPlayStyle { return p.playStyle }

// GetPlayStyleName プレイスタイル名取得
func (p *FiveCardStudPlayer) GetPlayStyleName() string {
	return playStyleName(int(p.playStyle), FiveCardStudPlayStyleNames)
}

// AddHoleCard 伏せ札を追加する
func (p *FiveCardStudPlayer) AddHoleCard(c *Card) {
	p.holeCards = append(p.holeCards, c)
}

// AddDoorCard 表向き札を追加する
func (p *FiveCardStudPlayer) AddDoorCard(c *Card) {
	p.doorCards = append(p.doorCards, c)
}

// GetAllCards 全カード取得 (伏せ札+表向き札)
func (p *FiveCardStudPlayer) GetAllCards() []*Card {
	all := make([]*Card, 0, len(p.holeCards)+len(p.doorCards))
	all = append(all, p.holeCards...)
	all = append(all, p.doorCards...)
	return all
}

// ClearCards 全カードをクリアする
func (p *FiveCardStudPlayer) ClearCards() {
	p.holeCards = p.holeCards[:0]
	p.doorCards = p.doorCards[:0]
	p.bestHand = nil
}

// GetTotalHands 総ハンド数取得
func (p *FiveCardStudPlayer) GetTotalHands() int { return p.totalHands }

// GetVPIPCount VPIP対象ハンド数取得
func (p *FiveCardStudPlayer) GetVPIPCount() int { return p.vpipCount }

// GetPFRCount PFR対象ハンド数取得
func (p *FiveCardStudPlayer) GetPFRCount() int { return p.pfrCount }

// GetVPIP VPIP%取得 (0 if totalHands==0)
func (p *FiveCardStudPlayer) GetVPIP() int {
	return percentOf(p.vpipCount, p.totalHands)
}

// GetPFR PFR%取得 (0 if totalHands==0)
func (p *FiveCardStudPlayer) GetPFR() int {
	return percentOf(p.pfrCount, p.totalHands)
}

// IncrementTotalHands 総ハンド数をインクリメント
func (p *FiveCardStudPlayer) IncrementTotalHands() { p.totalHands++ }

// IncrementVPIP VPIP対象ハンド数をインクリメント
func (p *FiveCardStudPlayer) IncrementVPIP() { p.vpipCount++ }

// IncrementPFR PFR対象ハンド数をインクリメント
func (p *FiveCardStudPlayer) IncrementPFR() { p.pfrCount++ }

// GetThreeBetOpportunity 3Bet機会数取得
func (p *FiveCardStudPlayer) GetThreeBetOpportunity() int { return p.threeBetOpportunity }

// GetThreeBetCount 3Bet実行数取得
func (p *FiveCardStudPlayer) GetThreeBetCount() int { return p.threeBetCount }

// GetThreeBet 3Bet%取得 (0 if threeBetOpportunity==0)
func (p *FiveCardStudPlayer) GetThreeBet() int {
	return percentOf(p.threeBetCount, p.threeBetOpportunity)
}

// IncrementThreeBetOpportunity 3Bet機会数をインクリメント
func (p *FiveCardStudPlayer) IncrementThreeBetOpportunity() { p.threeBetOpportunity++ }

// IncrementThreeBet 3Bet実行数をインクリメント
func (p *FiveCardStudPlayer) IncrementThreeBet() { p.threeBetCount++ }

// GetPostFlopBetRaise ポストフロップ ベット+レイズ回数取得
func (p *FiveCardStudPlayer) GetPostFlopBetRaise() int { return p.postFlopBetRaise }

// GetPostFlopCall ポストフロップ コール回数取得
func (p *FiveCardStudPlayer) GetPostFlopCall() int { return p.postFlopCall }

// IncrementPostFlopBetRaise ポストフロップ ベット+レイズ回数をインクリメント
func (p *FiveCardStudPlayer) IncrementPostFlopBetRaise() { p.postFlopBetRaise++ }

// IncrementPostFlopCall ポストフロップ コール回数をインクリメント
func (p *FiveCardStudPlayer) IncrementPostFlopCall() { p.postFlopCall++ }

// GetAFDisplay AF表示文字列取得 ("-"=アクションなし, "∞"=コールなし, "X.X"=通常)
func (p *FiveCardStudPlayer) GetAFDisplay() string {
	return afDisplay(p.postFlopBetRaise, p.postFlopCall)
}

// GetComparisonCards ハンド比較用カード取得 (BettingPlayerインターフェース)
func (p *FiveCardStudPlayer) GetComparisonCards() []*Card {
	return copyOf(p.bestHand)
}

// EvalBestHand 全カード (最大5枚) からベスト5枚を評価
func (p *FiveCardStudPlayer) EvalBestHand() int {
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
		rank := p.evalHand(combo)
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

// evalHand は5枚の役を評価する。Soko は4枚ストレート/4枚フラッシュを含む独自の
// 序列を使うため、標準の評価器とは戻り値のスケールが違う（soko_hand_eval.go 参照）。
// 両者を混ぜて比較してはいけないので、1ハンド内では必ず同じ側を使う。
func (p *FiveCardStudPlayer) evalHand(cards []*Card) int {
	if p.sokoMode {
		return evalSokoHand(cards)
	}
	return evalFiveCardHand(cards)
}

// EvalVisibleHand 表向き札のみからハンドを評価 (ベッティング順序決定用)
// 5枚未満の場合はハイカード比較用にソートした値を返す
func (p *FiveCardStudPlayer) EvalVisibleHand() int {
	if len(p.doorCards) == 0 {
		return PokerHandHighCard
	}
	if len(p.doorCards) >= 5 {
		return evalFiveCardHand(p.doorCards[:5])
	}
	// 5枚未満: 簡易評価 (ペア検出)
	return evalPartialHandFcs(p.doorCards)
}

// evalPartialHandFcs 5枚未満のカードを簡易評価する
func evalPartialHandFcs(cards []*Card) int {
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

// compareVisibleHandsFcs 2人のプレイヤーの表向き札を比較する (a > b: 1, a < b: -1, a == b: 0)
func compareVisibleHandsFcs(a, b *FiveCardStudPlayer) int {
	rankA := a.EvalVisibleHand()
	rankB := b.EvalVisibleHand()
	if rankA > rankB {
		return 1
	}
	if rankA < rankB {
		return -1
	}
	// 同ランク: ハイカード比較
	aVals := sortedDoorCardValuesFcs(a.doorCards)
	bVals := sortedDoorCardValuesFcs(b.doorCards)
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

// sortedDoorCardValuesFcs 表向き札の値をソートして返す (Ace=14)
func sortedDoorCardValuesFcs(cards []*Card) []int {
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

// fiveCardStudPlayerJSON is the JSON wire format for FiveCardStudPlayer.
type fiveCardStudPlayerJSON struct {
	Player              *Player               `json:"p"`
	ChipHolder          *ChipHolder           `json:"ch"`
	BettingPlayerBase   *bettingPlayerBase    `json:"bp"`
	IsHuman             bool                  `json:"ih"`
	HoleCards           []*Card               `json:"hc"`
	DoorCards           []*Card               `json:"dc"`
	BestHand            []*Card               `json:"bh"`
	PlayStyle           FiveCardStudPlayStyle `json:"ps"`
	TotalHands          int                   `json:"th"`
	VPIPCount           int                   `json:"vc"`
	PFRCount            int                   `json:"pc"`
	ThreeBetOpportunity int                   `json:"to"`
	ThreeBetCount       int                   `json:"tc"`
	PostFlopBetRaise    int                   `json:"pb"`
	PostFlopCall        int                   `json:"pf"`
}

// MarshalJSON implements json.Marshaler.
func (p *FiveCardStudPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(fiveCardStudPlayerJSON{
		Player:              &p.Player,
		ChipHolder:          &p.ChipHolder,
		BettingPlayerBase:   &p.bettingPlayerBase,
		IsHuman:             p.isHuman,
		HoleCards:           p.holeCards,
		DoorCards:           p.doorCards,
		BestHand:            p.bestHand,
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
func (p *FiveCardStudPlayer) UnmarshalJSON(data []byte) error {
	var j fiveCardStudPlayerJSON
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
		p.holeCards = make([]*Card, 0, 1)
	}
	p.doorCards = j.DoorCards
	if p.doorCards == nil {
		p.doorCards = make([]*Card, 0, 4)
	}
	p.bestHand = j.BestHand
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
