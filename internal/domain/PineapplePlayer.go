//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
)

// PineapplePlayer パイナップルポーカープレイヤークラス
type PineapplePlayer struct {
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

// NewPineapplePlayer コンストラクタ
func NewPineapplePlayer(isHuman bool, style HoldemPlayStyle) *PineapplePlayer {
	return &PineapplePlayer{
		Player:    Player{cards: make([]*Card, 0)},
		isHuman:   isHuman,
		playStyle: style,
	}
}

// GetIsHuman 人間フラグ取得
func (pp *PineapplePlayer) GetIsHuman() bool { return pp.isHuman }

// GetBestHand ベストハンド取得
func (pp *PineapplePlayer) GetBestHand() []*Card { return pp.bestHand }

// GetPlayStyle プレイスタイル取得
func (pp *PineapplePlayer) GetPlayStyle() HoldemPlayStyle { return pp.playStyle }

// GetPlayStyleName プレイスタイル名取得
func (pp *PineapplePlayer) GetPlayStyleName() string {
	return playStyleName(int(pp.playStyle), HoldemPlayStyleNames)
}

// GetTotalHands 総ハンド数取得
func (pp *PineapplePlayer) GetTotalHands() int { return pp.totalHands }

// GetVPIPCount VPIP対象ハンド数取得
func (pp *PineapplePlayer) GetVPIPCount() int { return pp.vpipCount }

// GetPFRCount PFR対象ハンド数取得
func (pp *PineapplePlayer) GetPFRCount() int { return pp.pfrCount }

// GetVPIP VPIP%取得 (0 if totalHands==0)
func (pp *PineapplePlayer) GetVPIP() int {
	return percentOf(pp.vpipCount, pp.totalHands)
}

// GetPFR PFR%取得 (0 if totalHands==0)
func (pp *PineapplePlayer) GetPFR() int {
	return percentOf(pp.pfrCount, pp.totalHands)
}

// IncrementTotalHands 総ハンド数をインクリメント
func (pp *PineapplePlayer) IncrementTotalHands() { pp.totalHands++ }

// IncrementVPIP VPIP対象ハンド数をインクリメント
func (pp *PineapplePlayer) IncrementVPIP() { pp.vpipCount++ }

// IncrementPFR PFR対象ハンド数をインクリメント
func (pp *PineapplePlayer) IncrementPFR() { pp.pfrCount++ }

// GetThreeBetOpportunity 3Bet機会数取得
func (pp *PineapplePlayer) GetThreeBetOpportunity() int { return pp.threeBetOpportunity }

// GetThreeBetCount 3Bet実行数取得
func (pp *PineapplePlayer) GetThreeBetCount() int { return pp.threeBetCount }

// GetThreeBet 3Bet%取得 (0 if threeBetOpportunity==0)
func (pp *PineapplePlayer) GetThreeBet() int {
	return percentOf(pp.threeBetCount, pp.threeBetOpportunity)
}

// IncrementThreeBetOpportunity 3Bet機会数をインクリメント
func (pp *PineapplePlayer) IncrementThreeBetOpportunity() { pp.threeBetOpportunity++ }

// IncrementThreeBet 3Bet実行数をインクリメント
func (pp *PineapplePlayer) IncrementThreeBet() { pp.threeBetCount++ }

// GetPostFlopBetRaise ポストフロップ ベット+レイズ回数取得
func (pp *PineapplePlayer) GetPostFlopBetRaise() int { return pp.postFlopBetRaise }

// GetPostFlopCall ポストフロップ コール回数取得
func (pp *PineapplePlayer) GetPostFlopCall() int { return pp.postFlopCall }

// IncrementPostFlopBetRaise ポストフロップ ベット+レイズ回数をインクリメント
func (pp *PineapplePlayer) IncrementPostFlopBetRaise() { pp.postFlopBetRaise++ }

// IncrementPostFlopCall ポストフロップ コール回数をインクリメント
func (pp *PineapplePlayer) IncrementPostFlopCall() { pp.postFlopCall++ }

// GetAFDisplay AF表示文字列取得 ("-"=アクションなし, "∞"=コールなし, "X.X"=通常)
func (pp *PineapplePlayer) GetAFDisplay() string {
	return afDisplay(pp.postFlopBetRaise, pp.postFlopCall)
}

// GetComparisonCards ハンド比較用カード取得 (BettingPlayerインターフェース)
func (pp *PineapplePlayer) GetComparisonCards() []*Card {
	return copyOf(pp.bestHand)
}

// EvalBestHand コミュニティカードとホールカードからベスト5枚を評価
// ディスカード後は2枚のホールカード + コミュニティカードからC(N,5)で評価 (Holdemと同じ)
// ディスカード前は3枚のホールカード + コミュニティカードからC(N,5)で評価
func (pp *PineapplePlayer) EvalBestHand(communityCards []*Card) int {
	all := make([]*Card, 0, len(pp.cards)+len(communityCards))
	all = append(all, pp.cards...)
	all = append(all, communityCards...)

	if len(all) < 5 {
		pp.handRank = PokerHandHighCard
		pp.bestHand = nil
		return pp.handRank
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

	pp.handRank = bestRank
	pp.bestHand = bestCards
	return pp.handRank
}

// pineapplePlayerJSON is the JSON wire format for PineapplePlayer.
type pineapplePlayerJSON struct {
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
func (pp *PineapplePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(pineapplePlayerJSON{
		Player:              &pp.Player,
		ChipHolder:          &pp.ChipHolder,
		BettingPlayerBase:   &pp.bettingPlayerBase,
		IsHuman:             pp.isHuman,
		BestHand:            pp.bestHand,
		PlayStyle:           pp.playStyle,
		TotalHands:          pp.totalHands,
		VPIPCount:           pp.vpipCount,
		PFRCount:            pp.pfrCount,
		ThreeBetOpportunity: pp.threeBetOpportunity,
		ThreeBetCount:       pp.threeBetCount,
		PostFlopBetRaise:    pp.postFlopBetRaise,
		PostFlopCall:        pp.postFlopCall,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (pp *PineapplePlayer) UnmarshalJSON(data []byte) error {
	var j pineapplePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Player != nil {
		pp.Player = *j.Player
	}
	if j.ChipHolder != nil {
		pp.ChipHolder = *j.ChipHolder
	}
	if j.BettingPlayerBase != nil {
		pp.bettingPlayerBase = *j.BettingPlayerBase
	}
	pp.isHuman = j.IsHuman
	pp.bestHand = j.BestHand
	pp.playStyle = j.PlayStyle
	pp.totalHands = j.TotalHands
	pp.vpipCount = j.VPIPCount
	pp.pfrCount = j.PFRCount
	pp.threeBetOpportunity = j.ThreeBetOpportunity
	pp.threeBetCount = j.ThreeBetCount
	pp.postFlopBetRaise = j.PostFlopBetRaise
	pp.postFlopCall = j.PostFlopCall
	return nil
}
