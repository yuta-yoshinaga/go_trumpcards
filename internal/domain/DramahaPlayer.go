//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
)

// DramahaPlayer ドラマハホールデムプレイヤークラス
type DramahaPlayer struct {
	Player                              // 親クラス
	ChipHolder                          // チップ管理
	bettingPlayerBase                   // ベッティング共通状態
	isHuman             bool            // 人間フラグ
	bestHand            []*Card         // ベスト5枚
	drawBestHand        []*Card         // ドロー側の5枚 (= ホールカードそのもの)
	drawRank            int             // ドロー側の役位
	playStyle           HoldemPlayStyle // CPUプレイスタイル (Holdemと共通)
	totalHands          int             // 総ハンド数 (セッション通算)
	vpipCount           int             // VPIP対象ハンド数
	pfrCount            int             // PFR対象ハンド数
	threeBetOpportunity int             // 3Bet機会数
	threeBetCount       int             // 3Bet実行数
	postFlopBetRaise    int             // ポストフロップ ベット+レイズ回数
	postFlopCall        int             // ポストフロップ コール回数
}

// NewDramahaPlayer コンストラクタ
func NewDramahaPlayer(isHuman bool, style HoldemPlayStyle) *DramahaPlayer {
	return &DramahaPlayer{
		Player:    Player{cards: make([]*Card, 0)},
		isHuman:   isHuman,
		playStyle: style,
	}
}

// GetIsHuman 人間フラグ取得
func (op *DramahaPlayer) GetIsHuman() bool { return op.isHuman }

// GetBestHand ベストハンド取得
func (op *DramahaPlayer) GetBestHand() []*Card { return op.bestHand }

// GetPlayStyle プレイスタイル取得
func (op *DramahaPlayer) GetPlayStyle() HoldemPlayStyle { return op.playStyle }

// GetPlayStyleName プレイスタイル名取得
func (op *DramahaPlayer) GetPlayStyleName() string {
	return playStyleName(int(op.playStyle), HoldemPlayStyleNames)
}

// GetTotalHands 総ハンド数取得
func (op *DramahaPlayer) GetTotalHands() int { return op.totalHands }

// GetVPIPCount VPIP対象ハンド数取得
func (op *DramahaPlayer) GetVPIPCount() int { return op.vpipCount }

// GetPFRCount PFR対象ハンド数取得
func (op *DramahaPlayer) GetPFRCount() int { return op.pfrCount }

// GetVPIP VPIP%取得 (0 if totalHands==0)
func (op *DramahaPlayer) GetVPIP() int {
	return percentOf(op.vpipCount, op.totalHands)
}

// GetPFR PFR%取得 (0 if totalHands==0)
func (op *DramahaPlayer) GetPFR() int {
	return percentOf(op.pfrCount, op.totalHands)
}

// IncrementTotalHands 総ハンド数をインクリメント
func (op *DramahaPlayer) IncrementTotalHands() { op.totalHands++ }

// IncrementVPIP VPIP対象ハンド数をインクリメント
func (op *DramahaPlayer) IncrementVPIP() { op.vpipCount++ }

// IncrementPFR PFR対象ハンド数をインクリメント
func (op *DramahaPlayer) IncrementPFR() { op.pfrCount++ }

// GetThreeBetOpportunity 3Bet機会数取得
func (op *DramahaPlayer) GetThreeBetOpportunity() int { return op.threeBetOpportunity }

// GetThreeBetCount 3Bet実行数取得
func (op *DramahaPlayer) GetThreeBetCount() int { return op.threeBetCount }

// GetThreeBet 3Bet%取得 (0 if threeBetOpportunity==0)
func (op *DramahaPlayer) GetThreeBet() int {
	return percentOf(op.threeBetCount, op.threeBetOpportunity)
}

// IncrementThreeBetOpportunity 3Bet機会数をインクリメント
func (op *DramahaPlayer) IncrementThreeBetOpportunity() { op.threeBetOpportunity++ }

// IncrementThreeBet 3Bet実行数をインクリメント
func (op *DramahaPlayer) IncrementThreeBet() { op.threeBetCount++ }

// GetPostFlopBetRaise ポストフロップ ベット+レイズ回数取得
func (op *DramahaPlayer) GetPostFlopBetRaise() int { return op.postFlopBetRaise }

// GetPostFlopCall ポストフロップ コール回数取得
func (op *DramahaPlayer) GetPostFlopCall() int { return op.postFlopCall }

// IncrementPostFlopBetRaise ポストフロップ ベット+レイズ回数をインクリメント
func (op *DramahaPlayer) IncrementPostFlopBetRaise() { op.postFlopBetRaise++ }

// IncrementPostFlopCall ポストフロップ コール回数をインクリメント
func (op *DramahaPlayer) IncrementPostFlopCall() { op.postFlopCall++ }

// GetAFDisplay AF表示文字列取得 ("-"=アクションなし, "∞"=コールなし, "X.X"=通常)
func (op *DramahaPlayer) GetAFDisplay() string {
	return afDisplay(op.postFlopBetRaise, op.postFlopCall)
}

// GetComparisonCards ハンド比較用カード取得 (BettingPlayerインターフェース)
func (op *DramahaPlayer) GetComparisonCards() []*Card {
	return copyOf(op.bestHand)
}

// GetLowComparisonCards Hi-Loロー比較用カード取得 (compareRazzCards 用)
func (op *DramahaPlayer) GetLowComparisonCards() []*Card {
	cards := make([]*Card, len(op.drawBestHand))
	copy(cards, op.drawBestHand)
	return cards
}

// dramahaPlayerJSON is the JSON wire format for DramahaPlayer.
type dramahaPlayerJSON struct {
	Player              *Player            `json:"p"`
	ChipHolder          *ChipHolder        `json:"ch"`
	BettingPlayerBase   *bettingPlayerBase `json:"bp"`
	IsHuman             bool               `json:"ih"`
	BestHand            []*Card            `json:"bh"`
	DrawBestHand        []*Card            `json:"dbh,omitempty"`
	DrawRank            int                `json:"dr,omitempty"`
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
func (op *DramahaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(dramahaPlayerJSON{
		Player:              &op.Player,
		ChipHolder:          &op.ChipHolder,
		BettingPlayerBase:   &op.bettingPlayerBase,
		IsHuman:             op.isHuman,
		BestHand:            op.bestHand,
		DrawBestHand:        op.drawBestHand,
		DrawRank:            op.drawRank,
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
func (op *DramahaPlayer) UnmarshalJSON(data []byte) error {
	var j dramahaPlayerJSON
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
	op.drawBestHand = j.DrawBestHand
	op.drawRank = j.DrawRank
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
// ドラマハルール: ホールカードから必ず2枚、コミュニティカードから必ず3枚を使う
// PeekBestHand は現在の手札とボードから最善の 5 枚役を求めて返す。**状態を
// 変えない。**
//
// 表示だけのために EvalBestHand を呼ぶと、描画のたびに handRank / bestHand を
// 書き換えてしまう。CUI の途中経過表示はこちらを使う (#4680)。手札 2 枚・
// ボード 3 枚に満たないときはハイカード扱いで、確定した組は返さない。
func (op *DramahaPlayer) PeekBestHand(communityCards []*Card) (rank int, best []*Card) {
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

func (op *DramahaPlayer) EvalBestHand(communityCards []*Card) int {
	// **判定は PeekBestHand が唯一の出どころ。**同じ探索を2つ持つと、片方だけ
	// 直したときに「表示とショーダウンで役が違う」ずれになる。
	op.handRank, op.bestHand = op.PeekBestHand(communityCards)
	return op.handRank
}

// EvalDrawHand は 5 枚のホールカードを**そのまま**ドローポーカーの手として
// 評価する。ポットのもう半分はこちらが取る。
//
// **ボードは一切見ない。** Omaha 側 (EvalBestHand) は「ホール 2 枚 + ボード
// 3 枚」を総当たりするが、ドロー側は配られた 5 枚がそのまま役になる。同じ
// 手札を二重に使うのがこのゲームの発想で、探索も評価軸も別物。
//
// **ローと違って必ず成立する。** 8 or Better のローは条件を満たさなければ
// 「不成立」になり得たが、5 枚あればどんな手でも役として順位が付く。だから
// 「ドロー側の勝者が居ないのでハイ側が総取り」は起こらない。
//
// 戻り値: 評価した役位。5 枚に満たなければ PokerHandHighCard を返して状態を
// クリアする。
func (op *DramahaPlayer) EvalDrawHand() int {
	op.drawRank = PokerHandHighCard
	op.drawBestHand = nil

	if len(op.cards) != DramahaHoleCards {
		return op.drawRank
	}
	hand := make([]*Card, DramahaHoleCards)
	copy(hand, op.cards)
	op.drawRank = evalFiveCardHand(hand)
	op.drawBestHand = hand
	return op.drawRank
}

// PeekDrawHand は現在のホールカードのドロー側役位を返す。**状態を変えない。**
// 途中経過の表示に使う (EvalDrawHand を表示のために呼ぶと、描画のたびに
// drawRank を書き換えてしまう)。
func (op *DramahaPlayer) PeekDrawHand() (rank int, best []*Card) {
	if len(op.cards) != DramahaHoleCards {
		return PokerHandHighCard, nil
	}
	hand := make([]*Card, DramahaHoleCards)
	copy(hand, op.cards)
	return evalFiveCardHand(hand), hand
}

// GetDrawRank はドロー側の役位を返す。
func (op *DramahaPlayer) GetDrawRank() int { return op.drawRank }

// ReplaceCard は指定位置のホールカードを差し替える。範囲外なら false。
//
// **スライスごと外に出さない。** 手札を返すと呼び出し側がいくらでも書き換えられ、
// 「誰がいつ札を変えたか」がドメインの外に散る。差し替えは 1 枚ずつここを通す。
func (op *DramahaPlayer) ReplaceCard(idx int, c *Card) bool {
	if idx < 0 || idx >= len(op.cards) || c == nil {
		return false
	}
	op.cards[idx] = c
	return true
}

// HoleCardsCopy はホールカードの複製を返す（評価・表示用）。
func (op *DramahaPlayer) HoleCardsCopy() []*Card {
	cards := make([]*Card, len(op.cards))
	copy(cards, op.cards)
	return cards
}

// GetDrawBestHand はドロー側の 5 枚を返す。
func (op *DramahaPlayer) GetDrawBestHand() []*Card { return op.drawBestHand }

// SetDrawHand はドロー側の評価結果を設定する（復元・テスト用）。
func (op *DramahaPlayer) SetDrawHand(rank int, best []*Card) {
	op.drawRank, op.drawBestHand = rank, best
}
