//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// OmahaPlayer オマハホールデムプレイヤークラス
type OmahaPlayer struct {
	Player                              // 親クラス
	ChipHolder                          // チップ管理
	bettingPlayerBase                   // ベッティング共通状態
	isHuman             bool            // 人間フラグ
	bestHand            []*Card         // ベスト5枚
	lowBestHand         []*Card         // Hi-Lo用ローベスト5枚 (qualified の場合のみ非nil)
	lowQualifies        bool            // 8 or better のローが成立したかどうか
	playStyle           HoldemPlayStyle // CPUプレイスタイル (Holdemと共通)
	totalHands          int             // 総ハンド数 (セッション通算)
	vpipCount           int             // VPIP対象ハンド数
	pfrCount            int             // PFR対象ハンド数
	threeBetOpportunity int             // 3Bet機会数
	threeBetCount       int             // 3Bet実行数
	postFlopBetRaise    int             // ポストフロップ ベット+レイズ回数
	postFlopCall        int             // ポストフロップ コール回数
}

// NewOmahaPlayer コンストラクタ
func NewOmahaPlayer(isHuman bool, style HoldemPlayStyle) *OmahaPlayer {
	return &OmahaPlayer{
		Player:    Player{cards: make([]*Card, 0)},
		isHuman:   isHuman,
		playStyle: style,
	}
}

// GetIsHuman 人間フラグ取得
func (op *OmahaPlayer) GetIsHuman() bool { return op.isHuman }

// GetBestHand ベストハンド取得
func (op *OmahaPlayer) GetBestHand() []*Card { return op.bestHand }

// GetPlayStyle プレイスタイル取得
func (op *OmahaPlayer) GetPlayStyle() HoldemPlayStyle { return op.playStyle }

// GetPlayStyleName プレイスタイル名取得
func (op *OmahaPlayer) GetPlayStyleName() string {
	return playStyleName(int(op.playStyle), HoldemPlayStyleNames)
}

// GetTotalHands 総ハンド数取得
func (op *OmahaPlayer) GetTotalHands() int { return op.totalHands }

// GetVPIPCount VPIP対象ハンド数取得
func (op *OmahaPlayer) GetVPIPCount() int { return op.vpipCount }

// GetPFRCount PFR対象ハンド数取得
func (op *OmahaPlayer) GetPFRCount() int { return op.pfrCount }

// GetVPIP VPIP%取得 (0 if totalHands==0)
func (op *OmahaPlayer) GetVPIP() int {
	if op.totalHands == 0 {
		return 0
	}
	return op.vpipCount * 100 / op.totalHands
}

// GetPFR PFR%取得 (0 if totalHands==0)
func (op *OmahaPlayer) GetPFR() int {
	if op.totalHands == 0 {
		return 0
	}
	return op.pfrCount * 100 / op.totalHands
}

// IncrementTotalHands 総ハンド数をインクリメント
func (op *OmahaPlayer) IncrementTotalHands() { op.totalHands++ }

// IncrementVPIP VPIP対象ハンド数をインクリメント
func (op *OmahaPlayer) IncrementVPIP() { op.vpipCount++ }

// IncrementPFR PFR対象ハンド数をインクリメント
func (op *OmahaPlayer) IncrementPFR() { op.pfrCount++ }

// GetThreeBetOpportunity 3Bet機会数取得
func (op *OmahaPlayer) GetThreeBetOpportunity() int { return op.threeBetOpportunity }

// GetThreeBetCount 3Bet実行数取得
func (op *OmahaPlayer) GetThreeBetCount() int { return op.threeBetCount }

// GetThreeBet 3Bet%取得 (0 if threeBetOpportunity==0)
func (op *OmahaPlayer) GetThreeBet() int {
	if op.threeBetOpportunity == 0 {
		return 0
	}
	return op.threeBetCount * 100 / op.threeBetOpportunity
}

// IncrementThreeBetOpportunity 3Bet機会数をインクリメント
func (op *OmahaPlayer) IncrementThreeBetOpportunity() { op.threeBetOpportunity++ }

// IncrementThreeBet 3Bet実行数をインクリメント
func (op *OmahaPlayer) IncrementThreeBet() { op.threeBetCount++ }

// GetPostFlopBetRaise ポストフロップ ベット+レイズ回数取得
func (op *OmahaPlayer) GetPostFlopBetRaise() int { return op.postFlopBetRaise }

// GetPostFlopCall ポストフロップ コール回数取得
func (op *OmahaPlayer) GetPostFlopCall() int { return op.postFlopCall }

// IncrementPostFlopBetRaise ポストフロップ ベット+レイズ回数をインクリメント
func (op *OmahaPlayer) IncrementPostFlopBetRaise() { op.postFlopBetRaise++ }

// IncrementPostFlopCall ポストフロップ コール回数をインクリメント
func (op *OmahaPlayer) IncrementPostFlopCall() { op.postFlopCall++ }

// GetAFDisplay AF表示文字列取得 ("-"=アクションなし, "∞"=コールなし, "X.X"=通常)
func (op *OmahaPlayer) GetAFDisplay() string {
	if op.postFlopBetRaise == 0 && op.postFlopCall == 0 {
		return "-"
	}
	if op.postFlopCall == 0 {
		return "∞"
	}
	return fmt.Sprintf("%.1f", float64(op.postFlopBetRaise)/float64(op.postFlopCall))
}

// GetComparisonCards ハンド比較用カード取得 (BettingPlayerインターフェース)
func (op *OmahaPlayer) GetComparisonCards() []*Card {
	cards := make([]*Card, len(op.bestHand))
	copy(cards, op.bestHand)
	return cards
}

// GetLowBestHand Hi-Lo用ローベスト5枚取得 (未評価/不成立時はnil)
func (op *OmahaPlayer) GetLowBestHand() []*Card { return op.lowBestHand }

// GetLowQualifies 8-or-better のローが成立したか
func (op *OmahaPlayer) GetLowQualifies() bool { return op.lowQualifies }

// GetLowComparisonCards Hi-Loロー比較用カード取得 (compareRazzCards 用)
func (op *OmahaPlayer) GetLowComparisonCards() []*Card {
	cards := make([]*Card, len(op.lowBestHand))
	copy(cards, op.lowBestHand)
	return cards
}

// omahaPlayerJSON is the JSON wire format for OmahaPlayer.
type omahaPlayerJSON struct {
	Player              *Player            `json:"p"`
	ChipHolder          *ChipHolder        `json:"ch"`
	BettingPlayerBase   *bettingPlayerBase `json:"bp"`
	IsHuman             bool               `json:"ih"`
	BestHand            []*Card            `json:"bh"`
	LowBestHand         []*Card            `json:"lh,omitempty"`
	LowQualifies        bool               `json:"lq,omitempty"`
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
func (op *OmahaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(omahaPlayerJSON{
		Player:              &op.Player,
		ChipHolder:          &op.ChipHolder,
		BettingPlayerBase:   &op.bettingPlayerBase,
		IsHuman:             op.isHuman,
		BestHand:            op.bestHand,
		LowBestHand:         op.lowBestHand,
		LowQualifies:        op.lowQualifies,
		PlayStyle:           op.playStyle,
		TotalHands:          op.totalHands,
		VPIPCount:           op.vpipCount,
		PFRCount:            op.pfrCount,
		ThreeBetOpportunity: op.threeBetOpportunity,
		ThreeBetCount:       op.threeBetCount,
		PostFlopBetRaise:    op.postFlopBetRaise,
		PostFlopCall:        op.postFlopCall,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (op *OmahaPlayer) UnmarshalJSON(data []byte) error {
	var j omahaPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Player != nil {
		op.Player = *j.Player
	}
	if j.ChipHolder != nil {
		op.ChipHolder = *j.ChipHolder
	}
	if j.BettingPlayerBase != nil {
		op.bettingPlayerBase = *j.BettingPlayerBase
	}
	op.isHuman = j.IsHuman
	op.bestHand = j.BestHand
	op.lowBestHand = j.LowBestHand
	op.lowQualifies = j.LowQualifies
	op.playStyle = j.PlayStyle
	op.totalHands = j.TotalHands
	op.vpipCount = j.VPIPCount
	op.pfrCount = j.PFRCount
	op.threeBetOpportunity = j.ThreeBetOpportunity
	op.threeBetCount = j.ThreeBetCount
	op.postFlopBetRaise = j.PostFlopBetRaise
	op.postFlopCall = j.PostFlopCall
	return nil
}

// EvalBestHand コミュニティカードとホールカード(4枚)からベスト5枚を評価
// オマハルール: ホールカードから必ず2枚、コミュニティカードから必ず3枚を使う
// PeekBestHand は現在の手札とボードから最善の 5 枚役を求めて返す。**状態を
// 変えない。**
//
// 表示だけのために EvalBestHand を呼ぶと、描画のたびに handRank / bestHand を
// 書き換えてしまう。CUI の途中経過表示はこちらを使う (#4680)。手札 2 枚・
// ボード 3 枚に満たないときはハイカード扱いで、確定した組は返さない。
func (op *OmahaPlayer) PeekBestHand(communityCards []*Card) (rank int, best []*Card) {
	if len(op.cards) < 2 || len(communityCards) < 3 {
		return PokerHandHighCard, nil
	}

	holePairs := combinations(op.cards, 2)         // C(4,2) = 6
	commTriples := combinations(communityCards, 3) // C(5,3) = 10

	bestRank := -1
	var bestCards []*Card
	for _, pair := range holePairs {
		for _, triple := range commTriples {
			hand := make([]*Card, 0, 5)
			hand = append(hand, pair...)
			hand = append(hand, triple...)
			r := evalFiveCardHand(hand)
			if r > bestRank || (r == bestRank && compareHighCardsSlice(hand, bestCards) > 0) {
				bestRank = r
				bestCards = make([]*Card, 5)
				copy(bestCards, hand)
			}
		}
	}
	return bestRank, bestCards
}

func (op *OmahaPlayer) EvalBestHand(communityCards []*Card) int {
	// **判定は PeekBestHand が唯一の出どころ。**同じ探索を2つ持つと、片方だけ
	// 直したときに「表示とショーダウンで役が違う」ずれになる。
	op.handRank, op.bestHand = op.PeekBestHand(communityCards)
	return op.handRank
}

// EvalBestLowHand Hi-Lo (8 or Better) 用のローベスト5枚を評価する。
// オマハルール: ホールカードから必ず2枚、コミュニティカードから必ず3枚。
// 5枚すべてが8以下 (Ace=1)、かつランクの重複が無い場合のみ qualified。
// qualified なローが存在しない場合は (false, nil) で状態をクリアする。
//
// 戻り値: qualifying な低手が見つかったかどうか。
func (op *OmahaPlayer) EvalBestLowHand(communityCards []*Card) bool {
	op.lowQualifies = false
	op.lowBestHand = nil

	if len(op.cards) < 2 || len(communityCards) < 3 {
		return false
	}

	holePairs := combinations(op.cards, 2)         // C(4,2) = 6
	commTriples := combinations(communityCards, 3) // C(5,3) = 10

	var bestCards []*Card
	var hand [5]*Card
	for _, pair := range holePairs {
		hand[0], hand[1] = pair[0], pair[1]
		for _, triple := range commTriples {
			hand[2], hand[3], hand[4] = triple[0], triple[1], triple[2]
			if !isQualifyingOmahaLow(hand[:]) {
				continue
			}
			if bestCards == nil || compareRazzCards(hand[:], bestCards) < 0 {
				bestCards = make([]*Card, 5)
				copy(bestCards, hand[:])
			}
		}
	}

	if bestCards == nil {
		return false
	}
	op.lowQualifies = true
	op.lowBestHand = bestCards
	return true
}

// isQualifyingOmahaLow は5枚カードがオマハ Hi-Lo の有効なロー
// (8 or Better、ペア無し、Ace=1) を満たすか判定する。
// ホットパス: showdown ごとにプレイヤー1人あたり最大60回呼ばれるため、
// map ではなく uint16 のビットマスクで重複検出を行いアロケーションを避ける。
func isQualifyingOmahaLow(cards []*Card) bool {
	if len(cards) != 5 {
		return false
	}
	var seen uint16
	for _, c := range cards {
		if c == nil {
			return false
		}
		v := c.GetValue() // Ace == 1
		if v < 1 || v > 8 {
			return false
		}
		mask := uint16(1) << v
		if seen&mask != 0 {
			return false
		}
		seen |= mask
	}
	return true
}
