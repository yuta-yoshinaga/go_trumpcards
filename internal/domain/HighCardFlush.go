//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

// ハイカードフラッシュフェーズ定数
const (
	HighCardFlushPhaseBet    = 1 // ベットフェーズ
	HighCardFlushPhaseAction = 2 // アクションフェーズ（Raise/Fold選択）
	HighCardFlushPhaseEnd    = 3 // 終了フェーズ
)

// ハイカードフラッシュデフォルト値
const (
	HighCardFlushDefaultChips = 1000  // デフォルトチップ
	HighCardFlushMinBet       = 10    // 最低ベット額
	HighCardFlushMaxBet       = 10000 // 最大ベット額
	HighCardFlushHandSize     = 7     // 1人あたりの配布枚数

	// HighCardFlushDealerMinFlushLen is the minimum flush length required for the dealer to qualify.
	HighCardFlushDealerMinFlushLen = 3
	// HighCardFlushDealerMinHigh is the minimum high-card value required for the dealer to qualify (9 = Nine).
	HighCardFlushDealerMinHigh = 9
)

// Flush Bonus 配当倍率（プレイヤーの最長フラッシュ枚数で決まる独立サイドベット）
const (
	HighCardFlushBonusFour  = 1   // 4枚 1:1
	HighCardFlushBonusFive  = 10  // 5枚 10:1
	HighCardFlushBonusSix   = 50  // 6枚 50:1
	HighCardFlushBonusSeven = 100 // 7枚 100:1
)

// Straight Flush Bonus 配当倍率（プレイヤーの最長ストレートフラッシュ枚数で決まる独立サイドベット）
const (
	HighCardFlushSFBonusThree = 8    // 3枚 8:1
	HighCardFlushSFBonusFour  = 60   // 4枚 60:1
	HighCardFlushSFBonusFive  = 100  // 5枚 100:1
	HighCardFlushSFBonusSix   = 1000 // 6枚 1000:1
	HighCardFlushSFBonusSeven = 8000 // 7枚 8000:1
)

// HighCardFlushHand represents a player's best same-suit subset (the "flush").
type HighCardFlushHand struct {
	Length    int   // 同スート最大枚数
	Suit      int   // CardDesign 値
	HighCards []int // フラッシュ内のカード値（Aceは14）を降順に
}

// HighCardFlush ハイカードフラッシュクラス
type HighCardFlush struct {
	trumpCards        *TrumpCards // トランプカード
	playerHand        []*Card     // プレイヤーハンド (7枚)
	dealerHand        []*Card     // ディーラーハンド (7枚)
	chips             ChipHolder  // チップ
	anteBet           int         // アンテベット額
	flushBonusBet     int         // Flush Bonus サイドベット額
	straightFlushBet  int         // Straight Flush Bonus サイドベット額
	raiseBet          int         // レイズベット額（プレイ時にante×倍率で決定）
	phase             int         // 現在のフェーズ
	gameEndFlag       bool        // ゲーム終了フラグ
	result            GameResult  // ゲーム結果
	antePayout        int         // アンテ配当（元ベットを含む合計返却）
	raisePayout       int         // レイズ配当（元ベットを含む合計返却）
	flushBonusPayout  int         // Flush Bonus 配当（元ベットを含む合計返却）
	straightFlushPay  int         // Straight Flush Bonus 配当（元ベットを含む合計返却）
	dealerQualified   bool        // ディーラークオリファイフラグ
	playerFlushLen    int
	playerFlushSuit   int // 上の長さを数えたスート (CardDesign 値)         // プレイヤーのフラッシュ長
	dealerFlushLen    int // ディーラーのフラッシュ長
	playerStraightLen int // プレイヤーのストレートフラッシュ長（最大0なら無し）
	actionLogBase
}

// NewHighCardFlush コンストラクタ
func NewHighCardFlush(trumpCards *TrumpCards) *HighCardFlush {
	trumpCards.Shuffle()
	return &HighCardFlush{
		trumpCards: trumpCards,
		phase:      HighCardFlushPhaseBet,
	}
}

// NewDefaultHighCardFlush デフォルト設定のハイカードフラッシュを生成する
func NewDefaultHighCardFlush() *HighCardFlush {
	hcf := NewHighCardFlush(NewTrumpCards(0))
	hcf.chips.SetChips(HighCardFlushDefaultChips)
	return hcf
}

// Reset ゲーム初期化
func (hcf *HighCardFlush) Reset() {
	hcf.gameEndFlag = false
	hcf.phase = HighCardFlushPhaseBet
	hcf.playerHand = nil
	hcf.dealerHand = nil
	hcf.anteBet = 0
	hcf.flushBonusBet = 0
	hcf.straightFlushBet = 0
	hcf.raiseBet = 0
	hcf.result = 0
	hcf.antePayout = 0
	hcf.raisePayout = 0
	hcf.flushBonusPayout = 0
	hcf.straightFlushPay = 0
	hcf.dealerQualified = false
	hcf.playerFlushLen = 0
	hcf.dealerFlushLen = 0
	hcf.playerStraightLen = 0
	hcf.actionLog = nil
	if hcf.chips.GetChips() < HighCardFlushMinBet {
		hcf.chips.SetChips(HighCardFlushDefaultChips)
	}
	hcf.trumpCards = NewTrumpCards(0)
	for range 10 {
		hcf.trumpCards.Shuffle()
	}
}

// Bet アンテ＋オプションのサイドベットを置き、カードを配布する
func (hcf *HighCardFlush) Bet(ante, flushBonus, straightFlush int) error {
	if hcf.phase != HighCardFlushPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if err := validateHCFBet(ante, true); err != nil {
		return err
	}
	if err := validateHCFBet(flushBonus, false); err != nil {
		return err
	}
	if err := validateHCFBet(straightFlush, false); err != nil {
		return err
	}
	total := ante + flushBonus + straightFlush
	if !hcf.chips.SubtractChips(total) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	hcf.anteBet = ante
	hcf.flushBonusBet = flushBonus
	hcf.straightFlushBet = straightFlush
	hcf.appendLog(0, "bet", fmt.Sprintf("ante=%d flush_bonus=%d straight_flush=%d", ante, flushBonus, straightFlush), nil)

	hcf.deal()
	hcf.phase = HighCardFlushPhaseAction
	return nil
}

// validateHCFBet ベット額の妥当性を検査する。
// required=true ならゼロは不可。required=false なら 0 で「ベットしない」を表す。
func validateHCFBet(amount int, required bool) error {
	if amount < 0 {
		return NewDomainError(ErrInvalidAmount, "Bet must not be negative.")
	}
	if amount == 0 {
		if required {
			return NewDomainError(ErrInvalidAmount, "Ante bet must be positive.")
		}
		return nil
	}
	if amount < HighCardFlushMinBet || amount%HighCardFlushMinBet != 0 || amount > HighCardFlushMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid bet amount.")
	}
	return nil
}

// MaxRaiseMultiplier returns the maximum raise multiplier allowed for the player's current flush length.
// 2-4 cards: 1x, 5 cards: 2x, 6-7 cards: 3x.
func (hcf *HighCardFlush) MaxRaiseMultiplier() int {
	switch {
	case hcf.playerFlushLen >= 6:
		return 3
	case hcf.playerFlushLen == 5:
		return 2
	default:
		return 1
	}
}

// Raise レイズ（プレイ）する。multiplierは1〜MaxRaiseMultiplier()。
func (hcf *HighCardFlush) Raise(multiplier int) error {
	if hcf.phase != HighCardFlushPhaseAction {
		return NewDomainError(ErrWrongPhase, "Raise is only allowed during the action phase.")
	}
	max := hcf.MaxRaiseMultiplier()
	if multiplier < 1 || multiplier > max {
		return NewDomainError(ErrInvalidAmount, fmt.Sprintf("Raise multiplier must be between 1 and %d.", max))
	}
	raise := hcf.anteBet * multiplier
	if !hcf.chips.SubtractChips(raise) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for raise.")
	}
	hcf.raiseBet = raise
	hcf.appendLog(0, "raise", fmt.Sprintf("raise bet=%d (x%d)", raise, multiplier), nil)
	hcf.resolve()
	return nil
}

// Fold フォールド（アンテ没収。サイドベットは独立に評価される）
func (hcf *HighCardFlush) Fold() error {
	if hcf.phase != HighCardFlushPhaseAction {
		return NewDomainError(ErrWrongPhase, "Fold is only allowed during the action phase.")
	}
	hcf.appendLog(0, "fold", "player folds", nil)

	hcf.result = GameResultLose
	// サイドベットはフォールドしても評価される
	hcf.evaluateBonuses()
	if total := hcf.flushBonusPayout + hcf.straightFlushPay; total > 0 {
		hcf.chips.AddChips(total)
	}

	hcf.gameEndFlag = true
	hcf.phase = HighCardFlushPhaseEnd
	hcf.appendLog(-1, "result", "player folded", nil)
	return nil
}

// deal 7枚ずつ配り、プレイヤーのフラッシュ長を計算する
func (hcf *HighCardFlush) deal() {
	hcf.playerHand = make([]*Card, 0, HighCardFlushHandSize)
	hcf.dealerHand = make([]*Card, 0, HighCardFlushHandSize)
	for range HighCardFlushHandSize {
		hcf.playerHand = append(hcf.playerHand, hcf.trumpCards.DrawCard())
		hcf.dealerHand = append(hcf.dealerHand, hcf.trumpCards.DrawCard())
	}
	playerBest := evalHighCardFlushHand(hcf.playerHand)
	hcf.playerFlushLen = playerBest.Length
	hcf.playerFlushSuit = playerBest.Suit
	hcf.playerStraightLen = evalLongestStraightFlushLen(hcf.playerHand)
	hcf.appendLog(-1, "deal", fmt.Sprintf("dealt 7 cards each (player flush=%d)", hcf.playerFlushLen), nil)
}

// resolve Raise後の解決処理
func (hcf *HighCardFlush) resolve() {
	playerBest := evalHighCardFlushHand(hcf.playerHand)
	dealerBest := evalHighCardFlushHand(hcf.dealerHand)
	hcf.playerFlushLen = playerBest.Length
	hcf.playerFlushSuit = playerBest.Suit
	hcf.dealerFlushLen = dealerBest.Length
	hcf.dealerQualified = checkHCFDealerQualifies(dealerBest)

	cmp := compareHighCardFlushHands(playerBest, dealerBest)
	switch {
	case cmp > 0:
		hcf.result = GameResultWin
	case cmp < 0:
		hcf.result = GameResultLose
	default:
		hcf.result = GameResultDraw
	}

	hcf.calculatePayouts()
	hcf.evaluateBonuses()

	if total := hcf.antePayout + hcf.raisePayout + hcf.flushBonusPayout + hcf.straightFlushPay; total > 0 {
		hcf.chips.AddChips(total)
	}

	hcf.gameEndFlag = true
	hcf.phase = HighCardFlushPhaseEnd

	var resultStr string
	switch hcf.result {
	case GameResultWin:
		resultStr = "player wins"
	case GameResultDraw:
		resultStr = "push"
	default:
		resultStr = "dealer wins"
	}
	hcf.appendLog(-1, "result", resultStr, nil)
}

// calculatePayouts アンテ／レイズの配当（元ベット返却分を含む合計）
func (hcf *HighCardFlush) calculatePayouts() {
	if !hcf.dealerQualified {
		// ディーラー未クオリファイ: アンテ1:1、レイズはプッシュ
		hcf.antePayout = hcf.anteBet * 2
		hcf.raisePayout = hcf.raiseBet
		return
	}
	switch hcf.result {
	case GameResultWin:
		hcf.antePayout = hcf.anteBet * 2
		hcf.raisePayout = hcf.raiseBet * 2
	case GameResultDraw:
		hcf.antePayout = hcf.anteBet
		hcf.raisePayout = hcf.raiseBet
	case GameResultLose:
		hcf.antePayout = 0
		hcf.raisePayout = 0
	}
}

// evaluateBonuses Flush Bonus と Straight Flush Bonus のサイドベット評価。
// 戻り値は元ベット込みの合計返却額。当選しなかった場合は 0。
func (hcf *HighCardFlush) evaluateBonuses() {
	if hcf.flushBonusBet > 0 {
		if mult := flushBonusMultiplier(hcf.playerFlushLen); mult > 0 {
			hcf.flushBonusPayout = hcf.flushBonusBet + hcf.flushBonusBet*mult
		}
	}
	if hcf.straightFlushBet > 0 {
		if mult := straightFlushBonusMultiplier(hcf.playerStraightLen); mult > 0 {
			hcf.straightFlushPay = hcf.straightFlushBet + hcf.straightFlushBet*mult
		}
	}
}

// flushBonusMultiplier プレイヤーの最長フラッシュ長から Flush Bonus の倍率を返す。
func flushBonusMultiplier(length int) int {
	switch {
	case length >= 7:
		return HighCardFlushBonusSeven
	case length == 6:
		return HighCardFlushBonusSix
	case length == 5:
		return HighCardFlushBonusFive
	case length == 4:
		return HighCardFlushBonusFour
	default:
		return 0
	}
}

// straightFlushBonusMultiplier プレイヤーの最長ストレートフラッシュ長から SF Bonus の倍率を返す。
func straightFlushBonusMultiplier(length int) int {
	switch {
	case length >= 7:
		return HighCardFlushSFBonusSeven
	case length == 6:
		return HighCardFlushSFBonusSix
	case length == 5:
		return HighCardFlushSFBonusFive
	case length == 4:
		return HighCardFlushSFBonusFour
	case length == 3:
		return HighCardFlushSFBonusThree
	default:
		return 0
	}
}

// evalHighCardFlushHand finds the player's best flush (same-suit subset) from 7 cards.
// Aces count as 14.
func evalHighCardFlushHand(cards []*Card) HighCardFlushHand {
	suitVals := make(map[int][]int)
	for _, c := range cards {
		if c == nil {
			continue
		}
		v := c.GetValue()
		if v == 1 {
			v = 14
		}
		suitVals[c.GetDesign()] = append(suitVals[c.GetDesign()], v)
	}
	var best HighCardFlushHand
	suits := make([]int, 0, len(suitVals))
	for s := range suitVals {
		suits = append(suits, s)
	}
	sort.Ints(suits) // deterministic tie-break
	for _, s := range suits {
		vals := suitVals[s]
		sort.Sort(sort.Reverse(sort.IntSlice(vals)))
		candidate := HighCardFlushHand{Length: len(vals), Suit: s, HighCards: vals}
		if compareHighCardFlushHands(candidate, best) > 0 {
			best = candidate
		}
	}
	return best
}

// compareHighCardFlushHands compares two flushes. Returns >0 if a wins, <0 if b wins, 0 tie.
func compareHighCardFlushHands(a, b HighCardFlushHand) int {
	if a.Length != b.Length {
		return a.Length - b.Length
	}
	for i := 0; i < a.Length && i < b.Length; i++ {
		if a.HighCards[i] != b.HighCards[i] {
			return a.HighCards[i] - b.HighCards[i]
		}
	}
	return 0
}

// checkHCFDealerQualifies returns true when the dealer's best flush is 3+ cards with high ≥ 9.
// Length and HighCards are kept consistent by evalHighCardFlushHand, so a non-zero length
// always implies a non-empty HighCards slice.
func checkHCFDealerQualifies(best HighCardFlushHand) bool {
	if best.Length < HighCardFlushDealerMinFlushLen {
		return false
	}
	return best.HighCards[0] >= HighCardFlushDealerMinHigh
}

// evalLongestStraightFlushLen returns the longest run of consecutive ranks within a single suit.
// Aces count as both 1 (low) and 14 (high). Returns 0 if no run of length ≥ 3 exists.
func evalLongestStraightFlushLen(cards []*Card) int {
	suitSets := make(map[int]map[int]bool)
	for _, c := range cards {
		if c == nil {
			continue
		}
		set, ok := suitSets[c.GetDesign()]
		if !ok {
			set = make(map[int]bool)
			suitSets[c.GetDesign()] = set
		}
		v := c.GetValue()
		set[v] = true
		if v == 1 {
			set[14] = true
		}
	}
	best := 0
	for _, set := range suitSets {
		if l := longestConsecutiveRun(set); l > best {
			best = l
		}
	}
	if best < 3 {
		return 0
	}
	return best
}

// longestConsecutiveRun finds the longest streak of consecutive integers in the set.
func longestConsecutiveRun(set map[int]bool) int {
	best := 0
	for v := range set {
		if set[v-1] {
			continue // not a run start
		}
		cur := v
		for set[cur] {
			cur++
		}
		if cur-v > best {
			best = cur - v
		}
	}
	return best
}

// --- Getters ---

// GetPlayerHand プレイヤーハンドを取得する
func (hcf *HighCardFlush) GetPlayerHand() []*Card { return hcf.playerHand }

// GetDealerHand ディーラーハンドを取得する
func (hcf *HighCardFlush) GetDealerHand() []*Card { return hcf.dealerHand }

// GetPhase 現在のフェーズを取得する
func (hcf *HighCardFlush) GetPhase() int { return hcf.phase }

// GetGameEndFlag ゲーム終了フラグを取得する
func (hcf *HighCardFlush) GetGameEndFlag() bool { return hcf.gameEndFlag }

// GetAnteBet アンテベット額を取得する
func (hcf *HighCardFlush) GetAnteBet() int { return hcf.anteBet }

// GetFlushBonusBet Flush Bonusベット額を取得する
func (hcf *HighCardFlush) GetFlushBonusBet() int { return hcf.flushBonusBet }

// GetStraightFlushBet Straight Flush Bonusベット額を取得する
func (hcf *HighCardFlush) GetStraightFlushBet() int { return hcf.straightFlushBet }

// GetRaiseBet レイズベット額を取得する
func (hcf *HighCardFlush) GetRaiseBet() int { return hcf.raiseBet }

// GetResult ゲーム結果を取得する
func (hcf *HighCardFlush) GetResult() GameResult { return hcf.result }

// GetAntePayout アンテ配当（元ベット込みの返却額）を取得する
func (hcf *HighCardFlush) GetAntePayout() int { return hcf.antePayout }

// GetRaisePayout レイズ配当を取得する
func (hcf *HighCardFlush) GetRaisePayout() int { return hcf.raisePayout }

// GetFlushBonusPayout Flush Bonus 配当を取得する
func (hcf *HighCardFlush) GetFlushBonusPayout() int { return hcf.flushBonusPayout }

// GetStraightFlushPayout Straight Flush Bonus 配当を取得する
func (hcf *HighCardFlush) GetStraightFlushPayout() int { return hcf.straightFlushPay }

// GetTotalPayout 合計配当
func (hcf *HighCardFlush) GetTotalPayout() int {
	return hcf.antePayout + hcf.raisePayout + hcf.flushBonusPayout + hcf.straightFlushPay
}

// GetDealerQualified ディーラークオリファイ
func (hcf *HighCardFlush) GetDealerQualified() bool { return hcf.dealerQualified }

// GetPlayerFlushLen プレイヤーの最長フラッシュ長
func (hcf *HighCardFlush) GetPlayerFlushLen() int { return hcf.playerFlushLen }

// GetPlayerFlushSuit は GetPlayerFlushLen が数えたスートを返す。
//
// **長さだけでは、7 枚のうちどれがその何枚なのか分からない** (#5607)。同着は
// evalHighCardFlushHand が高い札 → スート順で決めており、その決着をそのまま返す
// ので、画面の印と長さの行が別のスートを指すことはない。
func (hcf *HighCardFlush) GetPlayerFlushSuit() int { return hcf.playerFlushSuit }

// GetDealerFlushLen ディーラーの最長フラッシュ長
func (hcf *HighCardFlush) GetDealerFlushLen() int { return hcf.dealerFlushLen }

// GetPlayerStraightFlushLen プレイヤーの最長ストレートフラッシュ長（0なら無し）
func (hcf *HighCardFlush) GetPlayerStraightFlushLen() int { return hcf.playerStraightLen }

// GetChips チップ
func (hcf *HighCardFlush) GetChips() int { return hcf.chips.GetChips() }

// --- Test helpers ---

// SetPhase フェーズ設定（テスト用）
func (hcf *HighCardFlush) SetPhase(phase int) { hcf.phase = phase }

// SetPlayerHand プレイヤーハンド設定（テスト用）
func (hcf *HighCardFlush) SetPlayerHand(cards []*Card) {
	hcf.playerHand = cards
	best := evalHighCardFlushHand(cards)
	hcf.playerFlushLen = best.Length
	hcf.playerFlushSuit = best.Suit
	hcf.playerStraightLen = evalLongestStraightFlushLen(cards)
}

// SetDealerHand ディーラーハンド設定（テスト用）
func (hcf *HighCardFlush) SetDealerHand(cards []*Card) {
	hcf.dealerHand = cards
	hcf.dealerFlushLen = evalHighCardFlushHand(cards).Length
}

// SetAnteBet アンテベット額設定（テスト用）
func (hcf *HighCardFlush) SetAnteBet(amount int) { hcf.anteBet = amount }

// SetFlushBonusBet Flush Bonus ベット額設定（テスト用）
func (hcf *HighCardFlush) SetFlushBonusBet(amount int) { hcf.flushBonusBet = amount }

// SetStraightFlushBet Straight Flush Bonus ベット額設定（テスト用）
func (hcf *HighCardFlush) SetStraightFlushBet(amount int) { hcf.straightFlushBet = amount }

// SetRaiseBet レイズベット額設定（テスト用）
func (hcf *HighCardFlush) SetRaiseBet(amount int) { hcf.raiseBet = amount }

// SetResult ゲーム結果設定（テスト用）
func (hcf *HighCardFlush) SetResult(result GameResult) { hcf.result = result }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (hcf *HighCardFlush) SetGameEndFlag(flag bool) { hcf.gameEndFlag = flag }

// SetChips チップ設定（テスト用）
func (hcf *HighCardFlush) SetChips(chips int) { hcf.chips.SetChips(chips) }

// SetDealerQualified ディーラークオリファイ設定（テスト用）
func (hcf *HighCardFlush) SetDealerQualified(qualified bool) { hcf.dealerQualified = qualified }

// highCardFlushJSON is the JSON wire format for HighCardFlush.
type highCardFlushJSON struct {
	TrumpCards        *TrumpCards       `json:"tc"`
	PlayerHand        []*Card           `json:"ph"`
	DealerHand        []*Card           `json:"dh"`
	Chips             *ChipHolder       `json:"ch"`
	AnteBet           int               `json:"ab"`
	FlushBonusBet     int               `json:"fb"`
	StraightFlushBet  int               `json:"sb"`
	RaiseBet          int               `json:"rb"`
	Phase             int               `json:"ps"`
	GameEndFlag       bool              `json:"ge"`
	Result            GameResult        `json:"rs"`
	AntePayout        int               `json:"ap"`
	RaisePayout       int               `json:"rp"`
	FlushBonusPayout  int               `json:"fbp"`
	StraightFlushPay  int               `json:"sbp"`
	DealerQualified   bool              `json:"dq"`
	PlayerFlushLen    int               `json:"pfl"`
	DealerFlushLen    int               `json:"dfl"`
	PlayerStraightLen int               `json:"psl"`
	ActionLog         []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (hcf *HighCardFlush) MarshalJSON() ([]byte, error) {
	return json.Marshal(highCardFlushJSON{
		TrumpCards:        hcf.trumpCards,
		PlayerHand:        hcf.playerHand,
		DealerHand:        hcf.dealerHand,
		Chips:             &hcf.chips,
		AnteBet:           hcf.anteBet,
		FlushBonusBet:     hcf.flushBonusBet,
		StraightFlushBet:  hcf.straightFlushBet,
		RaiseBet:          hcf.raiseBet,
		Phase:             hcf.phase,
		GameEndFlag:       hcf.gameEndFlag,
		Result:            hcf.result,
		AntePayout:        hcf.antePayout,
		RaisePayout:       hcf.raisePayout,
		FlushBonusPayout:  hcf.flushBonusPayout,
		StraightFlushPay:  hcf.straightFlushPay,
		DealerQualified:   hcf.dealerQualified,
		PlayerFlushLen:    hcf.playerFlushLen,
		DealerFlushLen:    hcf.dealerFlushLen,
		PlayerStraightLen: hcf.playerStraightLen,
		ActionLog:         hcf.actionLog,
	})
}

// highCardFlushMaxSliceLen caps slice sizes during deserialisation.
const highCardFlushMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (hcf *HighCardFlush) UnmarshalJSON(data []byte) error {
	var j highCardFlushJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.PlayerHand) > highCardFlushMaxSliceLen || len(j.DealerHand) > highCardFlushMaxSliceLen ||
		len(j.ActionLog) > highCardFlushMaxSliceLen {
		return fmt.Errorf("highcardflush: input array exceeds maximum allowed size")
	}

	hcf.trumpCards = j.TrumpCards
	if hcf.trumpCards == nil {
		hcf.trumpCards = NewTrumpCards(0)
	}
	hcf.playerHand = j.PlayerHand
	if hcf.playerHand == nil {
		hcf.playerHand = make([]*Card, 0)
	}
	hcf.dealerHand = j.DealerHand
	if hcf.dealerHand == nil {
		hcf.dealerHand = make([]*Card, 0)
	}
	if j.Chips != nil {
		hcf.chips = *j.Chips
	}
	hcf.anteBet = j.AnteBet
	hcf.flushBonusBet = j.FlushBonusBet
	hcf.straightFlushBet = j.StraightFlushBet
	hcf.raiseBet = j.RaiseBet
	hcf.phase = j.Phase
	hcf.gameEndFlag = j.GameEndFlag
	hcf.result = j.Result
	hcf.antePayout = j.AntePayout
	hcf.raisePayout = j.RaisePayout
	hcf.flushBonusPayout = j.FlushBonusPayout
	hcf.straightFlushPay = j.StraightFlushPay
	hcf.dealerQualified = j.DealerQualified
	hcf.playerFlushLen = j.PlayerFlushLen
	hcf.dealerFlushLen = j.DealerFlushLen
	hcf.playerStraightLen = j.PlayerStraightLen
	hcf.actionLog = j.ActionLog
	if hcf.actionLog == nil {
		hcf.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
