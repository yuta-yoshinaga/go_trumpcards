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

// omahaPlayerJSON is the JSON wire format for OmahaPlayer.
type omahaPlayerJSON struct {
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
func (op *OmahaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(omahaPlayerJSON{
		Player:              &op.Player,
		ChipHolder:          &op.ChipHolder,
		BettingPlayerBase:   &op.bettingPlayerBase,
		IsHuman:             op.isHuman,
		BestHand:            op.bestHand,
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
func (op *OmahaPlayer) EvalBestHand(communityCards []*Card) int {
	if len(op.cards) < 2 || len(communityCards) < 3 {
		op.handRank = PokerHandHighCard
		op.bestHand = nil
		return op.handRank
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
			rank := evalFiveCardHand(hand)
			if rank > bestRank || (rank == bestRank && compareHighCardsSlice(hand, bestCards) > 0) {
				bestRank = rank
				bestCards = make([]*Card, 5)
				copy(bestCards, hand)
			}
		}
	}

	op.handRank = bestRank
	op.bestHand = bestCards
	return op.handRank
}
