//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"sort"
)

// FollowTheQueenPlayer フォロー・ザ・クイーンプレイヤー
type FollowTheQueenPlayer struct {
	Player                                    // 親クラス (cardsフィールドは使用しない)
	ChipHolder                                // チップ管理
	bettingPlayerBase                         // ベッティング共通状態
	isHuman           bool                    // 人間フラグ
	holeCards         []*Card                 // 伏せ札 (最大3枚: 1st, 2nd, 7th)
	doorCards         []*Card                 // 表向き札 (最大4枚: 3rd-6th Street)
	bestHand          []*Card                 // ベスト5枚
	playStyle         FollowTheQueenPlayStyle // CPUプレイスタイル
	// wildRank は現在の第2ワイルドのランク (0 は未設定)。**関数ではなく値で
	// 持つ。**Worker はリクエストごとに KV から組み直すので、述語を関数で
	// 持たせると往復のたびに失われ、復元した盤だけワイルドを見なくなる。
	wildRank            int
	totalHands          int // 総ハンド数 (セッション通算)
	vpipCount           int // VPIP対象ハンド数
	pfrCount            int // PFR対象ハンド数
	threeBetOpportunity int // 3Bet機会数
	threeBetCount       int // 3Bet実行数
	postFlopBetRaise    int // ポストフロップ ベット+レイズ回数
	postFlopCall        int // ポストフロップ コール回数
}

// NewFollowTheQueenPlayer コンストラクタ
func NewFollowTheQueenPlayer(isHuman bool, style FollowTheQueenPlayStyle) *FollowTheQueenPlayer {
	return &FollowTheQueenPlayer{
		Player:    Player{cards: make([]*Card, 0)},
		isHuman:   isHuman,
		holeCards: make([]*Card, 0, 3),
		doorCards: make([]*Card, 0, 4),
		playStyle: style,
	}
}

// GetIsHuman 人間フラグ取得
func (p *FollowTheQueenPlayer) GetIsHuman() bool { return p.isHuman }

// GetHoleCards 伏せ札取得
func (p *FollowTheQueenPlayer) GetHoleCards() []*Card { return p.holeCards }

// GetDoorCards 表向き札取得
func (p *FollowTheQueenPlayer) GetDoorCards() []*Card { return p.doorCards }

// GetBestHand ベストハンド取得
func (p *FollowTheQueenPlayer) GetBestHand() []*Card { return p.bestHand }

// GetPlayStyle プレイスタイル取得
func (p *FollowTheQueenPlayer) GetPlayStyle() FollowTheQueenPlayStyle { return p.playStyle }

// GetPlayStyleName プレイスタイル名取得
func (p *FollowTheQueenPlayer) GetPlayStyleName() string {
	return playStyleName(int(p.playStyle), FollowTheQueenPlayStyleNames)
}

// AddHoleCard 伏せ札を追加する
func (p *FollowTheQueenPlayer) AddHoleCard(c *Card) {
	p.holeCards = append(p.holeCards, c)
}

// AddDoorCard 表向き札を追加する
func (p *FollowTheQueenPlayer) AddDoorCard(c *Card) {
	p.doorCards = append(p.doorCards, c)
}

// GetAllCards 全カード取得 (伏せ札+表向き札)
func (p *FollowTheQueenPlayer) GetAllCards() []*Card {
	all := make([]*Card, 0, len(p.holeCards)+len(p.doorCards))
	all = append(all, p.holeCards...)
	all = append(all, p.doorCards...)
	return all
}

// ClearCards 全カードをクリアする
func (p *FollowTheQueenPlayer) ClearCards() {
	p.holeCards = p.holeCards[:0]
	p.doorCards = p.doorCards[:0]
	p.bestHand = nil
}

// GetTotalHands 総ハンド数取得
func (p *FollowTheQueenPlayer) GetTotalHands() int { return p.totalHands }

// GetVPIPCount VPIP対象ハンド数取得
func (p *FollowTheQueenPlayer) GetVPIPCount() int { return p.vpipCount }

// GetPFRCount PFR対象ハンド数取得
func (p *FollowTheQueenPlayer) GetPFRCount() int { return p.pfrCount }

// GetVPIP VPIP%取得 (0 if totalHands==0)
func (p *FollowTheQueenPlayer) GetVPIP() int {
	return percentOf(p.vpipCount, p.totalHands)
}

// GetPFR PFR%取得 (0 if totalHands==0)
func (p *FollowTheQueenPlayer) GetPFR() int {
	return percentOf(p.pfrCount, p.totalHands)
}

// IncrementTotalHands 総ハンド数をインクリメント
func (p *FollowTheQueenPlayer) IncrementTotalHands() { p.totalHands++ }

// IncrementVPIP VPIP対象ハンド数をインクリメント
func (p *FollowTheQueenPlayer) IncrementVPIP() { p.vpipCount++ }

// IncrementPFR PFR対象ハンド数をインクリメント
func (p *FollowTheQueenPlayer) IncrementPFR() { p.pfrCount++ }

// GetThreeBetOpportunity 3Bet機会数取得
func (p *FollowTheQueenPlayer) GetThreeBetOpportunity() int { return p.threeBetOpportunity }

// GetThreeBetCount 3Bet実行数取得
func (p *FollowTheQueenPlayer) GetThreeBetCount() int { return p.threeBetCount }

// GetThreeBet 3Bet%取得 (0 if threeBetOpportunity==0)
func (p *FollowTheQueenPlayer) GetThreeBet() int {
	return percentOf(p.threeBetCount, p.threeBetOpportunity)
}

// IncrementThreeBetOpportunity 3Bet機会数をインクリメント
func (p *FollowTheQueenPlayer) IncrementThreeBetOpportunity() { p.threeBetOpportunity++ }

// IncrementThreeBet 3Bet実行数をインクリメント
func (p *FollowTheQueenPlayer) IncrementThreeBet() { p.threeBetCount++ }

// GetPostFlopBetRaise ポストフロップ ベット+レイズ回数取得
func (p *FollowTheQueenPlayer) GetPostFlopBetRaise() int { return p.postFlopBetRaise }

// GetPostFlopCall ポストフロップ コール回数取得
func (p *FollowTheQueenPlayer) GetPostFlopCall() int { return p.postFlopCall }

// IncrementPostFlopBetRaise ポストフロップ ベット+レイズ回数をインクリメント
func (p *FollowTheQueenPlayer) IncrementPostFlopBetRaise() { p.postFlopBetRaise++ }

// IncrementPostFlopCall ポストフロップ コール回数をインクリメント
func (p *FollowTheQueenPlayer) IncrementPostFlopCall() { p.postFlopCall++ }

// GetAFDisplay AF表示文字列取得 ("-"=アクションなし, "∞"=コールなし, "X.X"=通常)
func (p *FollowTheQueenPlayer) GetAFDisplay() string {
	return afDisplay(p.postFlopBetRaise, p.postFlopCall)
}

// GetComparisonCards ハンド比較用カード取得 (BettingPlayerインターフェース)
func (p *FollowTheQueenPlayer) GetComparisonCards() []*Card {
	return copyOf(p.bestHand)
}

// PeekBestHand は現在の持ち札から最善の 5 枚役を求めて返す。**状態を変えない。**
//
// 表示だけのために EvalBestHand を呼ぶと、描画のたびに handRank / bestHand を
// 書き換えてしまう。CUI の途中経過表示はこちらを使う (#4695)。5 枚未満のときは
// ハイカード扱いで、確定した組は返さない。
func (p *FollowTheQueenPlayer) PeekBestHand() (rank int, best []*Card) {
	all := p.GetAllCards()
	if len(all) < 5 {
		return PokerHandHighCard, nil
	}

	bestRank := -1
	var bestCards []*Card
	for _, combo := range combinations(all, 5) {
		// **置換後の 5 枚を持ち回る。** 元の combo を bestHand にすると、
		// 同位の比較がワイルドの印刷された額面を読み、ワイルドで作った
		// フォーカードが本物のフォーカードに勝ってしまう。
		r, hand := evalFiveCardHandWithWilds(combo, p.IsWild)
		if r > bestRank || (r == bestRank && compareHighCardsSlice(hand, bestCards) > 0) {
			bestRank = r
			bestCards = hand
		}
	}
	return bestRank, bestCards
}

// SetWildRank は第2ワイルドのランクを設定する（0 で解除）。ゲーム側が表向きの
// 札を見て決め、全プレイヤーに配る。
func (p *FollowTheQueenPlayer) SetWildRank(rank int) { p.wildRank = rank }

// GetWildRank は現在の第2ワイルドのランクを返す。
func (p *FollowTheQueenPlayer) GetWildRank() int { return p.wildRank }

// IsWild はその札がワイルドかを返す。**Q は常に、加えて現在の wildRank が。**
//
// 評価はここを通る 1 本だけ。ワイルドを見ない評価器が 1 つでも残ると、
// 「ワイルドのある手が普通に負ける」という、規則が飾りになった状態になる。
func (p *FollowTheQueenPlayer) IsWild(card *Card) bool {
	if card == nil {
		return false
	}
	v := card.GetValue()
	return v == FollowTheQueenQueenValue || (p.wildRank != 0 && v == p.wildRank)
}

// EvalBestHand 全7枚からベスト5枚を評価し、結果をプレイヤーに記録する。
func (p *FollowTheQueenPlayer) EvalBestHand() int {
	// **判定は PeekBestHand が唯一の出どころ。**同じ探索を2つ持つと、片方だけ
	// 直したときに「表示とショーダウンで役が違う」ずれになる。
	p.handRank, p.bestHand = p.PeekBestHand()
	return p.handRank
}

// EvalVisibleHand 表向き札のみからハンドを評価 (ベッティング順序決定用)
// 5枚未満の場合はハイカード比較用にソートした値を返す
func (p *FollowTheQueenPlayer) EvalVisibleHand() int {
	if len(p.doorCards) == 0 {
		return PokerHandHighCard
	}
	if len(p.doorCards) >= 5 {
		return evalFiveCardHand(p.doorCards[:5])
	}
	// 5枚未満: 簡易評価 (ペア検出)
	return followTheQueenEvalPartialHand(p.doorCards)
}

// followthequeenPlayerJSON is the JSON wire format for FollowTheQueenPlayer.
type followTheQueenPlayerJSON struct {
	Player            *Player                 `json:"p"`
	ChipHolder        *ChipHolder             `json:"ch"`
	BettingPlayerBase *bettingPlayerBase      `json:"bp"`
	IsHuman           bool                    `json:"ih"`
	HoleCards         []*Card                 `json:"hc"`
	DoorCards         []*Card                 `json:"dc"`
	BestHand          []*Card                 `json:"bh"`
	PlayStyle         FollowTheQueenPlayStyle `json:"ps"`
	// WildRank も往復させる。載せ忘れると、復元した盤でワイルドが消える。
	WildRank            int `json:"wr,omitempty"`
	TotalHands          int `json:"th"`
	VPIPCount           int `json:"vc"`
	PFRCount            int `json:"pc"`
	ThreeBetOpportunity int `json:"to"`
	ThreeBetCount       int `json:"tc"`
	PostFlopBetRaise    int `json:"pb"`
	PostFlopCall        int `json:"pf"`
}

// MarshalJSON implements json.Marshaler.
func (p *FollowTheQueenPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(followTheQueenPlayerJSON{
		Player:              &p.Player,
		ChipHolder:          &p.ChipHolder,
		BettingPlayerBase:   &p.bettingPlayerBase,
		IsHuman:             p.isHuman,
		HoleCards:           p.holeCards,
		DoorCards:           p.doorCards,
		BestHand:            p.bestHand,
		PlayStyle:           p.playStyle,
		WildRank:            p.wildRank,
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
func (p *FollowTheQueenPlayer) UnmarshalJSON(data []byte) error {
	var j followTheQueenPlayerJSON
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
	p.playStyle = j.PlayStyle
	p.wildRank = j.WildRank
	p.totalHands = j.TotalHands
	p.vpipCount = j.VPIPCount
	p.pfrCount = j.PFRCount
	p.threeBetOpportunity = j.ThreeBetOpportunity
	p.threeBetCount = j.ThreeBetCount
	p.postFlopBetRaise = j.PostFlopBetRaise
	p.postFlopCall = j.PostFlopCall
	return nil
}

// followTheQueenEvalPartialHand 5枚未満のカードを簡易評価する
func followTheQueenEvalPartialHand(cards []*Card) int {
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

// followTheQueenCompareVisibleHands 2人のプレイヤーの表向き札を比較する (a > b: 1, a < b: -1, a == b: 0)
func followTheQueenCompareVisibleHands(a, b *FollowTheQueenPlayer) int {
	rankA := a.EvalVisibleHand()
	rankB := b.EvalVisibleHand()
	if rankA > rankB {
		return 1
	}
	if rankA < rankB {
		return -1
	}
	// 同ランク: ハイカード比較
	aVals := followTheQueenSortedDoorCardValues(a.doorCards)
	bVals := followTheQueenSortedDoorCardValues(b.doorCards)
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

// followTheQueenSortedDoorCardValues 表向き札の値をソートして返す (Ace=14)
func followTheQueenSortedDoorCardValues(cards []*Card) []int {
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
