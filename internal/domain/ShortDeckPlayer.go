package domain

import "fmt"

// ShortDeckPlayer ショートデックホールデムプレイヤークラス
type ShortDeckPlayer struct {
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

// NewShortDeckPlayer コンストラクタ
func NewShortDeckPlayer(isHuman bool, style HoldemPlayStyle) *ShortDeckPlayer {
	return &ShortDeckPlayer{
		Player:    Player{cards: make([]*Card, 0)},
		isHuman:   isHuman,
		playStyle: style,
	}
}

// GetIsHuman 人間フラグ取得
func (sp *ShortDeckPlayer) GetIsHuman() bool { return sp.isHuman }

// GetBestHand ベストハンド取得
func (sp *ShortDeckPlayer) GetBestHand() []*Card { return sp.bestHand }

// GetPlayStyle プレイスタイル取得
func (sp *ShortDeckPlayer) GetPlayStyle() HoldemPlayStyle { return sp.playStyle }

// GetPlayStyleName プレイスタイル名取得
func (sp *ShortDeckPlayer) GetPlayStyleName() string {
	return playStyleName(int(sp.playStyle), HoldemPlayStyleNames)
}

// GetTotalHands 総ハンド数取得
func (sp *ShortDeckPlayer) GetTotalHands() int { return sp.totalHands }

// GetVPIPCount VPIP対象ハンド数取得
func (sp *ShortDeckPlayer) GetVPIPCount() int { return sp.vpipCount }

// GetPFRCount PFR対象ハンド数取得
func (sp *ShortDeckPlayer) GetPFRCount() int { return sp.pfrCount }

// GetVPIP VPIP%取得 (0 if totalHands==0)
func (sp *ShortDeckPlayer) GetVPIP() int {
	if sp.totalHands == 0 {
		return 0
	}
	return sp.vpipCount * 100 / sp.totalHands
}

// GetPFR PFR%取得 (0 if totalHands==0)
func (sp *ShortDeckPlayer) GetPFR() int {
	if sp.totalHands == 0 {
		return 0
	}
	return sp.pfrCount * 100 / sp.totalHands
}

// IncrementTotalHands 総ハンド数をインクリメント
func (sp *ShortDeckPlayer) IncrementTotalHands() { sp.totalHands++ }

// IncrementVPIP VPIP対象ハンド数をインクリメント
func (sp *ShortDeckPlayer) IncrementVPIP() { sp.vpipCount++ }

// IncrementPFR PFR対象ハンド数をインクリメント
func (sp *ShortDeckPlayer) IncrementPFR() { sp.pfrCount++ }

// GetThreeBetOpportunity 3Bet機会数取得
func (sp *ShortDeckPlayer) GetThreeBetOpportunity() int { return sp.threeBetOpportunity }

// GetThreeBetCount 3Bet実行数取得
func (sp *ShortDeckPlayer) GetThreeBetCount() int { return sp.threeBetCount }

// GetThreeBet 3Bet%取得 (0 if threeBetOpportunity==0)
func (sp *ShortDeckPlayer) GetThreeBet() int {
	if sp.threeBetOpportunity == 0 {
		return 0
	}
	return sp.threeBetCount * 100 / sp.threeBetOpportunity
}

// IncrementThreeBetOpportunity 3Bet機会数をインクリメント
func (sp *ShortDeckPlayer) IncrementThreeBetOpportunity() { sp.threeBetOpportunity++ }

// IncrementThreeBet 3Bet実行数をインクリメント
func (sp *ShortDeckPlayer) IncrementThreeBet() { sp.threeBetCount++ }

// GetPostFlopBetRaise ポストフロップ ベット+レイズ回数取得
func (sp *ShortDeckPlayer) GetPostFlopBetRaise() int { return sp.postFlopBetRaise }

// GetPostFlopCall ポストフロップ コール回数取得
func (sp *ShortDeckPlayer) GetPostFlopCall() int { return sp.postFlopCall }

// IncrementPostFlopBetRaise ポストフロップ ベット+レイズ回数をインクリメント
func (sp *ShortDeckPlayer) IncrementPostFlopBetRaise() { sp.postFlopBetRaise++ }

// IncrementPostFlopCall ポストフロップ コール回数をインクリメント
func (sp *ShortDeckPlayer) IncrementPostFlopCall() { sp.postFlopCall++ }

// GetAFDisplay AF表示文字列取得 ("-"=アクションなし, "∞"=コールなし, "X.X"=通常)
func (sp *ShortDeckPlayer) GetAFDisplay() string {
	if sp.postFlopBetRaise == 0 && sp.postFlopCall == 0 {
		return "-"
	}
	if sp.postFlopCall == 0 {
		return "∞"
	}
	return fmt.Sprintf("%.1f", float64(sp.postFlopBetRaise)/float64(sp.postFlopCall))
}

// GetComparisonCards ハンド比較用カード取得 (BettingPlayerインターフェース)
func (sp *ShortDeckPlayer) GetComparisonCards() []*Card {
	cards := make([]*Card, len(sp.bestHand))
	copy(cards, sp.bestHand)
	return cards
}

// EvalBestHand コミュニティカードとホールカード(2枚)からベスト5枚を評価 (ショートデック用)
func (sp *ShortDeckPlayer) EvalBestHand(communityCards []*Card) int {
	all := make([]*Card, 0, len(sp.cards)+len(communityCards))
	all = append(all, sp.cards...)
	all = append(all, communityCards...)

	if len(all) < 5 {
		sp.handRank = ShortDeckHandHighCard
		sp.bestHand = nil
		return sp.handRank
	}

	combos := combinations(all, 5)
	bestRank := -1
	var bestCards []*Card

	for _, combo := range combos {
		rank := evalShortDeckFiveCardHand(combo)
		if rank > bestRank || (rank == bestRank && compareShortDeckHighCardsSlice(combo, bestCards) > 0) {
			bestRank = rank
			bestCards = make([]*Card, 5)
			copy(bestCards, combo)
		}
	}

	sp.handRank = bestRank
	sp.bestHand = bestCards
	return sp.handRank
}
