//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// BlackJack Switch フェーズ定数
const (
	BJSwitchPhaseBet    = 1 // ベットフェーズ
	BJSwitchPhaseSwitch = 2 // スイッチフェーズ（2枚目を交換するか選ぶ）
	BJSwitchPhaseAction = 3 // ヒット/スタンド/ダブルダウンフェーズ
	BJSwitchPhaseEnd    = 4 // 終了フェーズ
)

// BlackJack Switch デフォルト値
const (
	BJSwitchDefaultChips = 1000  // デフォルトチップ
	BJSwitchMinBet       = 10    // 最低ベット額（1ハンドあたり）
	BJSwitchMaxBet       = 10000 // 最大ベット額（1ハンドあたり）
	BJSwitchHands        = 2     // 常に2ハンド
)

// BJSwitchDealerHitsSoft17 はディーラーがソフト17で必ずヒットする標準ルール。
// Blackjack Switch は通常 H17（dealer hits soft 17）で運用される。
const BJSwitchDealerHitsSoft17 = true

// BlackJackSwitch はブラックジャック・スイッチのゲーム本体。
// プレイヤーは常に2ハンドをディールされ、ディール直後に2ハンド目のカード
// を相互交換する「スイッチ」アクションを選べる。代償として:
//   - ディーラーが22になった場合は（ナチュラルBJを除き）バーストではなく
//     プッシュ扱い。
//   - プレイヤーのナチュラルBJ配当は 3:2 ではなく 1:1。
//
// 簡略化のため、スプリット / インシュランス / サレンダー / サイドベット
// / CPU はサポートしない（標準BJ実装に任せる）。
type BlackJackSwitch struct {
	trumpCards     *TrumpCards      // 山札
	player         *BlackJackPlayer // プレイヤー（チップ保持）
	dealer         *BlackJackPlayer // ディーラー
	hands          []*BlackJackHand // 常に2ハンド
	currentHandIdx int              // 現在操作中のハンドインデックス
	phase          int              // 現在のフェーズ
	gameEndFlag    bool             // ゲーム終了フラグ
	switched       bool             // 直近のラウンドでスイッチを実行したか
	handResults    []GameResult     // 各ハンドの勝敗結果
	handPayouts    []int            // 各ハンドの配当（ベット返却込み）
	dealerPushed22 bool             // ディーラー22プッシュが発生したか
	actionLogBase
}

// NewBlackJackSwitch コンストラクタ
func NewBlackJackSwitch(tc *TrumpCards, player, dealer *BlackJackPlayer) *BlackJackSwitch {
	tc.Shuffle()
	return &BlackJackSwitch{
		trumpCards: tc,
		player:     player,
		dealer:     dealer,
		hands:      newSwitchHands(),
		phase:      BJSwitchPhaseBet,
	}
}

// BJSwitchDeckCount は標準的なテーブルに合わせたシューのデッキ数。
const BJSwitchDeckCount = 6

// NewDefaultBlackJackSwitch デフォルト設定のブラックジャック・スイッチを生成。
func NewDefaultBlackJackSwitch() *BlackJackSwitch {
	bs := NewBlackJackSwitch(NewTrumpCardsWithDecks(BJSwitchDeckCount, 0), NewBlackJackPlayer(), NewBlackJackPlayer())
	bs.player.SetChips(BJSwitchDefaultChips)
	bs.dealer.SetChips(BJSwitchDefaultChips)
	return bs
}

// newSwitchHands は2つの空ハンドを生成する。
func newSwitchHands() []*BlackJackHand {
	out := make([]*BlackJackHand, BJSwitchHands)
	for i := range out {
		out[i] = NewBlackJackHand()
	}
	return out
}

// Reset ゲーム初期化（チップは保持。最低ベット未満ならデフォルトに戻す。）
func (bs *BlackJackSwitch) Reset() {
	bs.gameEndFlag = false
	bs.phase = BJSwitchPhaseBet
	bs.currentHandIdx = 0
	bs.switched = false
	bs.dealerPushed22 = false
	bs.hands = newSwitchHands()
	bs.handResults = nil
	bs.handPayouts = nil
	bs.actionLog = nil
	if bs.player.GetChips() < BJSwitchMinBet {
		bs.player.SetChips(BJSwitchDefaultChips)
	}
	if bs.dealer.GetChips() < BJSwitchMinBet {
		bs.dealer.SetChips(BJSwitchDefaultChips)
	}
	bs.player.Reset()
	bs.dealer.Reset()
	// シューを毎ラウンド再構築（簡略化）。標準BJのようなペネトレーション
	// 制御は変種の本質ではないため省略する。
	bs.trumpCards = NewTrumpCardsWithDecks(BJSwitchDeckCount, 0)
	bs.trumpCards.Shuffle()
}

// PlayerBet ベットしてカードを配り、スイッチフェーズへ進む。
// amount は 1 ハンドあたりの額。総コスト = amount * 2。
func (bs *BlackJackSwitch) PlayerBet(amount int) error {
	if bs.phase != BJSwitchPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if amount < BJSwitchMinBet || amount%BJSwitchMinBet != 0 || amount > BJSwitchMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid bet amount.")
	}
	totalCost := amount * BJSwitchHands
	if !bs.player.SubtractChips(totalCost) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	for _, h := range bs.hands {
		h.SetBet(amount)
	}
	// インターリーブ配り: hand0-c1, hand1-c1, dealer-c1, hand0-c2, hand1-c2, dealer-c2
	for round := 0; round < 2; round++ {
		for i := 0; i < BJSwitchHands; i++ {
			c := bs.trumpCards.DrawCard()
			if c == nil {
				return bs.refundOnDealFailure(totalCost)
			}
			bs.hands[i].AddCard(c)
		}
		c := bs.trumpCards.DrawCard()
		if c == nil {
			return bs.refundOnDealFailure(totalCost)
		}
		bs.dealer.AddCard(c)
	}
	bs.appendLog(0, "bet", fmt.Sprintf("bet %d on each hand", amount), nil)
	// ナチュラルBJを互いに持っていれば即終了。そうでなくとも、
	// スイッチの選択肢は残す（プレイヤーは現状維持を選べる）。
	if bs.dealerNaturalBJ() {
		bs.endGame()
		return nil
	}
	bs.phase = BJSwitchPhaseSwitch
	return nil
}

// refundOnDealFailure は山札枯渇時に状態を巻き戻す。
func (bs *BlackJackSwitch) refundOnDealFailure(totalCost int) error {
	bs.player.AddChips(totalCost)
	bs.player.Reset()
	bs.dealer.Reset()
	bs.hands = newSwitchHands()
	return ErrDeckExhausted
}

// PlayerSwitch ハンド0とハンド1の2枚目のカードを交換する。
func (bs *BlackJackSwitch) PlayerSwitch() error {
	if bs.phase != BJSwitchPhaseSwitch {
		return NewDomainError(ErrWrongPhase, "Switch is not allowed now.")
	}
	if bs.hands[0].GetCardsSize() < 2 || bs.hands[1].GetCardsSize() < 2 {
		return NewDomainError(ErrInvalidPlay, "Both hands need 2 cards before switching.")
	}
	c0 := bs.hands[0].GetCard(1)
	c1 := bs.hands[1].GetCard(1)
	bs.hands[0].SetCard(1, c1)
	bs.hands[1].SetCard(1, c0)
	bs.switched = true
	bs.appendLog(0, "switch", "switch second cards", []*Card{c0, c1})
	return bs.beginActionPhase()
}

// SwitchPreviewScores は 2 枚目を入れ替えた場合の両ハンドの得点を返す。
// ok が false なら入れ替えられない (どちらかが 2 枚に満たない)。
//
// **打つまで得か損か分からなかった。**Web はホバーで先読みを出している (#5586)
// のに、CUI は `switch` を実行して結果を見るまで比べられなかった。得点は
// 実際に入れ替える PlayerSwitch と同じ CalculateBlackJackScore を通す ──
// 別に数え直すと、先読みと結果が食い違う。
func (bs *BlackJackSwitch) SwitchPreviewScores() (first, second int, ok bool) {
	if bs.hands[0].GetCardsSize() < 2 || bs.hands[1].GetCardsSize() < 2 {
		return 0, 0, false
	}
	swapped := func(h *BlackJackHand, other *BlackJackHand) []*Card {
		cards := make([]*Card, h.GetCardsSize())
		for i := range cards {
			cards[i] = h.GetCard(i)
		}
		cards[1] = other.GetCard(1)
		return cards
	}
	return CalculateBlackJackScore(swapped(bs.hands[0], bs.hands[1])),
		CalculateBlackJackScore(swapped(bs.hands[1], bs.hands[0])),
		true
}

// PlayerKeep スイッチを行わず現状のハンドでアクションフェーズへ進む。
func (bs *BlackJackSwitch) PlayerKeep() error {
	if bs.phase != BJSwitchPhaseSwitch {
		return NewDomainError(ErrWrongPhase, "Keep is not allowed now.")
	}
	bs.appendLog(0, "keep", "keep current hands", nil)
	return bs.beginActionPhase()
}

// beginActionPhase はナチュラルBJを自動スタンドし、アクションフェーズへ進む。
// 全ハンドBJなら即時に終局へ。
func (bs *BlackJackSwitch) beginActionPhase() error {
	bs.phase = BJSwitchPhaseAction
	for _, h := range bs.hands {
		if h.GetCardsSize() == 2 && h.GetScore() == 21 {
			h.SetStood(true)
		}
	}
	bs.advanceHand()
	return nil
}

// PlayerHit プレイヤーヒット
func (bs *BlackJackSwitch) PlayerHit() error {
	if bs.phase != BJSwitchPhaseAction {
		return NewDomainError(ErrWrongPhase, "Hit is not allowed now.")
	}
	hand := bs.currentHand()
	if hand == nil || hand.IsFinished() {
		return NewDomainError(ErrHandFinished, "This hand is already finished.")
	}
	c := bs.trumpCards.DrawCard()
	if c == nil {
		return ErrDeckExhausted
	}
	hand.AddCard(c)
	bs.appendLog(0, "hit", "hit", []*Card{c})
	if hand.GetScore() >= 22 {
		hand.SetBusted(true)
		bs.advanceHand()
	}
	return nil
}

// PlayerStand プレイヤースタンド
func (bs *BlackJackSwitch) PlayerStand() error {
	if bs.phase != BJSwitchPhaseAction {
		return NewDomainError(ErrWrongPhase, "Stand is not allowed now.")
	}
	hand := bs.currentHand()
	if hand == nil || hand.IsFinished() {
		return NewDomainError(ErrHandFinished, "This hand is already finished.")
	}
	hand.SetStood(true)
	bs.appendLog(0, "stand", "stand", nil)
	bs.advanceHand()
	return nil
}

// PlayerDoubleDown プレイヤーダブルダウン（2枚時のみ、1枚引いてスタンド）
func (bs *BlackJackSwitch) PlayerDoubleDown() error {
	if bs.phase != BJSwitchPhaseAction {
		return NewDomainError(ErrWrongPhase, "Double down is not allowed now.")
	}
	hand := bs.currentHand()
	if hand == nil || hand.IsFinished() {
		return NewDomainError(ErrHandFinished, "This hand is already finished.")
	}
	if hand.GetCardsSize() != 2 {
		return NewDomainError(ErrInvalidPlay, "Double down is only allowed with 2 cards.")
	}
	bet := hand.GetBet()
	if !bs.player.SubtractChips(bet) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for double down.")
	}
	hand.SetBet(bet * 2)
	hand.SetDoubled(true)
	c := bs.trumpCards.DrawCard()
	if c == nil {
		// 山札枯渇: ベット復元
		bs.player.AddChips(bet)
		hand.SetBet(bet)
		hand.SetDoubled(false)
		return ErrDeckExhausted
	}
	hand.AddCard(c)
	bs.appendLog(0, "doubledown", "double down", []*Card{c})
	if hand.GetScore() >= 22 {
		hand.SetBusted(true)
	} else {
		hand.SetStood(true)
	}
	bs.advanceHand()
	return nil
}

// currentHand 現在操作中のハンドを返す（範囲外なら nil）。
func (bs *BlackJackSwitch) currentHand() *BlackJackHand {
	if bs.currentHandIdx < 0 || bs.currentHandIdx >= len(bs.hands) {
		return nil
	}
	return bs.hands[bs.currentHandIdx]
}

// advanceHand 次の未完了ハンドへ進む。全ハンド完了ならディーラープレイ→精算。
func (bs *BlackJackSwitch) advanceHand() {
	for i := 0; i < len(bs.hands); i++ {
		if !bs.hands[i].IsFinished() {
			bs.currentHandIdx = i
			return
		}
	}
	bs.dealerPlay()
}

// dealerPlay ディーラーターン。全ハンドがバーストの場合はホールカードのみ
// オープンしてゲーム終了する（標準BJと同じ）。
func (bs *BlackJackSwitch) dealerPlay() {
	if !bs.allHandsBusted() {
		bs.dealerPlayCards()
	}
	bs.endGame()
}

// dealerPlayCards ディーラーが標準ルール（H17）に従ってカードを引く。
func (bs *BlackJackSwitch) dealerPlayCards() {
	for {
		score := bs.dealer.GetScore()
		if score > 21 {
			break // バースト → そのまま終了（22プッシュは精算側で判定）
		}
		if score >= 18 {
			bs.appendLog(-1, "dealerstand", "dealer stand", nil)
			break
		}
		if score == 17 && (!BJSwitchDealerHitsSoft17 || !bs.dealer.IsSoft()) {
			bs.appendLog(-1, "dealerstand", "dealer stand", nil)
			break
		}
		c := bs.trumpCards.DrawCard()
		if c == nil {
			bs.appendLog(-1, "dealerstand", "dealer stand", nil)
			break
		}
		bs.dealer.AddCard(c)
		bs.appendLog(-1, "dealerhit", "dealer hit", []*Card{c})
	}
}

// allHandsBusted 全ハンドがバーストか。
func (bs *BlackJackSwitch) allHandsBusted() bool {
	for _, h := range bs.hands {
		if !h.IsBusted() {
			return false
		}
	}
	return true
}

// dealerNaturalBJ ディール直後のディーラーがナチュラルBJか。
func (bs *BlackJackSwitch) dealerNaturalBJ() bool {
	return bs.dealer.GetCardsSize() == 2 && bs.dealer.GetScore() == 21
}

// endGame ゲーム終了処理（精算 + 結果ログ）。
func (bs *BlackJackSwitch) endGame() {
	bs.resolvePayouts()
	bs.gameEndFlag = true
	bs.phase = BJSwitchPhaseEnd
	overall := bs.overallResult()
	var detail string
	switch overall {
	case GameResultWin:
		detail = "player wins"
	case GameResultDraw:
		detail = "draw"
	case GameResultLose:
		detail = "player loses"
	}
	bs.appendLog(-1, "result", detail, nil)
}

// resolvePayouts 各ハンドを精算する。
//
// Blackjack Switch のペイアウトルール:
//   - プレイヤーバースト: 没収。
//   - ディーラーがナチュラルBJ（2枚で21）: プレイヤーがナチュラル21でなければ負け。
//   - ディーラーが22で終了 (= dealerPushed22): プレイヤーが21を含めバーストしていなければ全ハンドプッシュ。
//   - ナチュラル21（プレイヤーが2枚で21）の配当は通常勝ちと同じ 1:1（3:2 ではない）。
func (bs *BlackJackSwitch) resolvePayouts() {
	dealerScore := bs.dealer.GetScore()
	dealerNaturalBJ := bs.dealerNaturalBJ()
	bs.dealerPushed22 = dealerScore == 22 && !dealerNaturalBJ

	bs.handResults = make([]GameResult, len(bs.hands))
	bs.handPayouts = make([]int, len(bs.hands))

	for i, h := range bs.hands {
		result := bs.judgeHand(h, dealerScore, dealerNaturalBJ)
		bs.handResults[i] = result
		bet := h.GetBet()
		switch result {
		case GameResultWin:
			bs.handPayouts[i] = bet * 2 // 1:1（BJも含めて 1:1）
			bs.player.AddChips(bs.handPayouts[i])
		case GameResultDraw:
			bs.handPayouts[i] = bet
			bs.player.AddChips(bet)
		case GameResultLose:
			bs.handPayouts[i] = 0
		}
	}
}

// judgeHand 単一ハンドの勝敗を判定する。
func (bs *BlackJackSwitch) judgeHand(h *BlackJackHand, dealerScore int, dealerNaturalBJ bool) GameResult {
	playerScore := h.GetScore()
	if playerScore > 21 {
		return GameResultLose
	}
	if dealerNaturalBJ {
		if playerScore == 21 && h.GetCardsSize() == 2 {
			return GameResultDraw
		}
		return GameResultLose
	}
	// ディーラー22プッシュ（ナチュラルBJを除く）
	if dealerScore == 22 {
		if playerScore == 21 && h.GetCardsSize() == 2 {
			return GameResultWin
		}
		return GameResultDraw
	}
	if dealerScore > 22 {
		return GameResultWin
	}
	// 通常比較
	switch {
	case playerScore > dealerScore:
		return GameResultWin
	case playerScore < dealerScore:
		return GameResultLose
	default:
		return GameResultDraw
	}
}

// overallResult 全ハンドを集計した「総合的な勝敗」を返す（CLI/Webの一行サマリ用）。
// 1勝1敗 や 1勝1分 などはドロー扱い、勝ち越し/負け越しはそれぞれ Win / Lose。
func (bs *BlackJackSwitch) overallResult() GameResult {
	wins, losses := 0, 0
	for _, r := range bs.handResults {
		switch r {
		case GameResultWin:
			wins++
		case GameResultLose:
			losses++
		}
	}
	switch {
	case wins > losses:
		return GameResultWin
	case losses > wins:
		return GameResultLose
	default:
		return GameResultDraw
	}
}

// --- Getters ---

// GetPlayer プレイヤー
func (bs *BlackJackSwitch) GetPlayer() *BlackJackPlayer { return bs.player }

// GetDealer ディーラー
func (bs *BlackJackSwitch) GetDealer() *BlackJackPlayer { return bs.dealer }

// GetHands プレイヤーハンド一覧（常に2件）
func (bs *BlackJackSwitch) GetHands() []*BlackJackHand { return bs.hands }

// GetCurrentHandIdx 現在操作中のハンドインデックス
func (bs *BlackJackSwitch) GetCurrentHandIdx() int { return bs.currentHandIdx }

// GetPhase 現在のフェーズ
func (bs *BlackJackSwitch) GetPhase() int { return bs.phase }

// GetGameEndFlag ゲーム終了フラグ
func (bs *BlackJackSwitch) GetGameEndFlag() bool { return bs.gameEndFlag }

// IsSwitched 直近のラウンドでスイッチを実行したか
func (bs *BlackJackSwitch) IsSwitched() bool { return bs.switched }

// IsDealerPushed22 ディーラー22プッシュが発生したか
func (bs *BlackJackSwitch) IsDealerPushed22() bool { return bs.dealerPushed22 }

// GetHandResults 各ハンドの勝敗結果（終了後）
func (bs *BlackJackSwitch) GetHandResults() []GameResult { return bs.handResults }

// GetHandPayouts 各ハンドの配当（ベット返却込み、終了後）
func (bs *BlackJackSwitch) GetHandPayouts() []int { return bs.handPayouts }

// GetTotalPayout 全ハンドの合計配当（ベット返却込み）
func (bs *BlackJackSwitch) GetTotalPayout() int {
	total := 0
	for _, p := range bs.handPayouts {
		total += p
	}
	return total
}

// GetOverallResult 総合勝敗（一行サマリ用）
func (bs *BlackJackSwitch) GetOverallResult() GameResult { return bs.overallResult() }

// --- Test helpers ---

// SetPhase テスト用
func (bs *BlackJackSwitch) SetPhase(phase int) { bs.phase = phase }

// SetCurrentHandIdx テスト用
func (bs *BlackJackSwitch) SetCurrentHandIdx(idx int) { bs.currentHandIdx = idx }

// SetHands テスト用
func (bs *BlackJackSwitch) SetHands(hands []*BlackJackHand) { bs.hands = hands }

// --- JSON ---

// blackJackSwitchJSON は BlackJackSwitch の JSON ワイヤーフォーマット。
type blackJackSwitchJSON struct {
	TrumpCards     *TrumpCards       `json:"tc"`
	Player         *BlackJackPlayer  `json:"pl"`
	Dealer         *BlackJackPlayer  `json:"dl"`
	Hands          []*BlackJackHand  `json:"hd"`
	CurrentHandIdx int               `json:"ci"`
	Phase          int               `json:"ps"`
	GameEndFlag    bool              `json:"ge"`
	Switched       bool              `json:"sw"`
	HandResults    []GameResult      `json:"hr"`
	HandPayouts    []int             `json:"hp"`
	DealerPushed22 bool              `json:"dp"`
	ActionLog      []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (bs *BlackJackSwitch) MarshalJSON() ([]byte, error) {
	return json.Marshal(blackJackSwitchJSON{
		TrumpCards:     bs.trumpCards,
		Player:         bs.player,
		Dealer:         bs.dealer,
		Hands:          bs.hands,
		CurrentHandIdx: bs.currentHandIdx,
		Phase:          bs.phase,
		GameEndFlag:    bs.gameEndFlag,
		Switched:       bs.switched,
		HandResults:    bs.handResults,
		HandPayouts:    bs.handPayouts,
		DealerPushed22: bs.dealerPushed22,
		ActionLog:      bs.actionLog,
	})
}

const blackJackSwitchMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (bs *BlackJackSwitch) UnmarshalJSON(data []byte) error {
	var j blackJackSwitchJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Hands) > blackJackSwitchMaxSliceLen ||
		len(j.HandResults) > blackJackSwitchMaxSliceLen ||
		len(j.HandPayouts) > blackJackSwitchMaxSliceLen ||
		len(j.ActionLog) > blackJackSwitchMaxSliceLen {
		return fmt.Errorf("blackjackswitch: input array exceeds maximum allowed size")
	}
	bs.trumpCards = j.TrumpCards
	if bs.trumpCards == nil {
		bs.trumpCards = NewTrumpCards(0)
	}
	bs.player = j.Player
	if bs.player == nil {
		bs.player = NewBlackJackPlayer()
	}
	bs.dealer = j.Dealer
	if bs.dealer == nil {
		bs.dealer = NewBlackJackPlayer()
	}
	bs.hands = j.Hands
	if len(bs.hands) != BJSwitchHands {
		bs.hands = newSwitchHands()
	}
	bs.currentHandIdx = j.CurrentHandIdx
	bs.phase = j.Phase
	if bs.phase == 0 {
		bs.phase = BJSwitchPhaseBet
	}
	bs.gameEndFlag = j.GameEndFlag
	bs.switched = j.Switched
	bs.handResults = j.HandResults
	bs.handPayouts = j.HandPayouts
	bs.dealerPushed22 = j.DealerPushed22
	bs.actionLog = j.ActionLog
	if bs.actionLog == nil {
		bs.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
