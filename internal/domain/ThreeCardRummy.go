//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// スリーカード・ラミーフェーズ定数
const (
	ThreeCardRummyPhaseBet    = 1 // ベットフェーズ
	ThreeCardRummyPhaseAction = 2 // アクションフェーズ（Play/Fold選択）
	ThreeCardRummyPhaseEnd    = 3 // 終了フェーズ
)

// スリーカード・ラミーデフォルト値
const (
	ThreeCardRummyDefaultChips = 1000  // デフォルトチップ
	ThreeCardRummyMinBet       = 10    // 最低ベット額
	ThreeCardRummyMaxBet       = 10000 // 最大ベット額
	ThreeCardRummyHandSize     = 3     // ハンドサイズ
)

// アンテボーナス配当倍率。**低い手ほど厚い。**
//
// Three Card Poker のボーナス表 (ストレートフラッシュ / スリーカード /
// ストレート) はここでは意味を持たない ── 役の強さではなく合計点の低さで
// 勝負するゲームなので、しきい値も点数で置く。
const (
	// ThreeCardRummyAnteBonusPerfect は 0 点 (同ランク3枚 / 同スート連番3枚) の配当。
	ThreeCardRummyAnteBonusPerfect = 9
	// ThreeCardRummyAnteBonusVeryLow は 1〜5 点の配当。
	ThreeCardRummyAnteBonusVeryLow = 3
	// ThreeCardRummyAnteBonusLow は 6〜10 点の配当。
	ThreeCardRummyAnteBonusLow = 1

	// ThreeCardRummyBonusVeryLowMax / LowMax はその区分の上限点。
	ThreeCardRummyBonusVeryLowMax = 5
	ThreeCardRummyBonusLowMax     = 10
)

// ローボーナス（サイドベット）の配当倍率。アンテボーナスより厚い ——
// ディーラーと無関係に自分の点だけで決まるぶん、当たりにくい。
const (
	// ThreeCardRummyLowBonusPerfect は 0 点の配当。
	ThreeCardRummyLowBonusPerfect = 100
	// ThreeCardRummyLowBonusVeryLow は 1〜5 点の配当。
	ThreeCardRummyLowBonusVeryLow = 20
	// ThreeCardRummyLowBonusLow は 6〜10 点の配当。
	ThreeCardRummyLowBonusLow = 4
)

// ThreeCardRummy スリーカード・ラミークラス
type ThreeCardRummy struct {
	trumpCards      *TrumpCards // トランプカード
	playerHand      []*Card     // プレイヤーハンド
	dealerHand      []*Card     // ディーラーハンド
	chips           ChipHolder  // チップ
	anteBet         int         // アンテベット額
	lowBonusBet     int         // ローボーナスベット額
	lastAnteBet     int         // 直前ラウンドのアンテ額 (Reset を跨いで残す)
	lastLowBonusBet int         // 直前ラウンドのローボーナス額 (Reset を跨いで残す)
	playBet         int         // プレイベット額
	phase           int         // 現在のフェーズ
	gameEndFlag     bool        // ゲーム終了フラグ
	result          GameResult  // ゲーム結果
	antePayout      int         // アンテ配当
	playPayout      int         // プレイ配当
	anteBonusPayout int         // アンテボーナス配当
	lowBonusPayout  int         // ローボーナス配当
	dealerQualified bool        // ディーラークオリファイフラグ
	playerScore     int         // プレイヤーの点数（低いほど強い）
	dealerScore     int         // ディーラーの点数（低いほど強い）
	actionLogBase
}

// NewThreeCardRummy コンストラクタ
func NewThreeCardRummy(trumpCards *TrumpCards) *ThreeCardRummy {
	trumpCards.Shuffle()
	return &ThreeCardRummy{
		trumpCards: trumpCards,
		phase:      ThreeCardRummyPhaseBet,
	}
}

// NewDefaultThreeCardRummy デフォルト設定のスリーカード・ラミーを生成するファクトリ関数
func NewDefaultThreeCardRummy() *ThreeCardRummy {
	tc := NewThreeCardRummy(NewTrumpCards(0))
	tc.chips.SetChips(ThreeCardRummyDefaultChips)
	return tc
}

// Rebet は直前のラウンドと同じ額で賭け直す (#5513)。
//
// Web はラウンド終了時の額を React 側に覚えてワンクリック再ベットできるが、
// CLI/CUI には同等の手段が無く毎ラウンド手打ちしていた。**検査は Bet に任せる** --
// 別に書くと、チップ不足や上限の扱いが通常のベットとずれる。
func (tc *ThreeCardRummy) Rebet() error {
	if tc.lastAnteBet <= 0 {
		return NewDomainError(ErrInvalidPlay, "まだ賭けていないので再ベットできません")
	}
	return tc.Bet(tc.lastAnteBet, tc.lastLowBonusBet)
}

// Reset ゲーム初期化
func (tc *ThreeCardRummy) Reset() {
	tc.gameEndFlag = false
	tc.phase = ThreeCardRummyPhaseBet
	tc.playerHand = nil
	tc.dealerHand = nil
	tc.anteBet = 0
	tc.lowBonusBet = 0
	// lastAnteBet / lastLowBonusBet はここで消さない。次のラウンドで Rebet が
	// 参照する唯一の手掛かりで、Web が React 側に持つ lastBet に当たる。
	tc.playBet = 0
	tc.result = 0
	tc.antePayout = 0
	tc.playPayout = 0
	tc.anteBonusPayout = 0
	tc.lowBonusPayout = 0
	tc.dealerQualified = false
	tc.playerScore = 0
	tc.dealerScore = 0
	tc.actionLog = nil
	if tc.chips.GetChips() < ThreeCardRummyMinBet {
		tc.chips.SetChips(ThreeCardRummyDefaultChips)
	}
	tc.trumpCards = NewTrumpCards(0)
	for range 10 {
		tc.trumpCards.Shuffle()
	}
}

// Bet アンテベット＆カード配布
func (tc *ThreeCardRummy) Bet(ante, lowBonus int) error {
	if tc.phase != ThreeCardRummyPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if ante < ThreeCardRummyMinBet || ante%ThreeCardRummyMinBet != 0 || ante > ThreeCardRummyMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid ante amount.")
	}
	if lowBonus < 0 {
		return NewDomainError(ErrInvalidAmount, "Low Bonus bet must not be negative.")
	}
	if lowBonus > 0 && (lowBonus < ThreeCardRummyMinBet || lowBonus%ThreeCardRummyMinBet != 0 || lowBonus > ThreeCardRummyMaxBet) {
		return NewDomainError(ErrInvalidAmount, "Invalid Low Bonus bet amount.")
	}
	totalCost := ante + lowBonus
	if !tc.chips.SubtractChips(totalCost) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	tc.anteBet = ante
	tc.lastAnteBet = ante
	tc.lowBonusBet = lowBonus
	tc.lastLowBonusBet = lowBonus
	tc.appendLog(0, "bet", fmt.Sprintf("ante=%d lowbonus=%d", ante, lowBonus), nil)

	// ディール: 3枚ずつ配る
	tc.deal()
	tc.phase = ThreeCardRummyPhaseAction
	return nil
}

// Play プレイ（アンテと同額のプレイベットを置いて勝負）
func (tc *ThreeCardRummy) Play() error {
	if tc.phase != ThreeCardRummyPhaseAction {
		return NewDomainError(ErrWrongPhase, "Play is only allowed during the action phase.")
	}
	// プレイベットはアンテと同額
	if !tc.chips.SubtractChips(tc.anteBet) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for play bet.")
	}
	tc.playBet = tc.anteBet
	tc.appendLog(0, "play", fmt.Sprintf("play bet=%d", tc.playBet), nil)

	tc.resolve()
	return nil
}

// Fold フォールド（アンテ没収、ローボーナスは別途評価）
func (tc *ThreeCardRummy) Fold() error {
	if tc.phase != ThreeCardRummyPhaseAction {
		return NewDomainError(ErrWrongPhase, "Fold is only allowed during the action phase.")
	}
	tc.appendLog(0, "fold", "player folds", nil)

	tc.result = GameResultLose
	tc.playerScore = ThreeCardRummyScore(tc.playerHand)
	tc.dealerScore = ThreeCardRummyScore(tc.dealerHand)
	// **降りても資格の有無は計算する。** 配当には効かない (勝負していないので
	// calculatePayouts は走らない) が、結果画面はディーラーの手を開いて資格を
	// 書く。ここを飛ばすと 4 点の手に「クオリファイせず」と出る。
	tc.dealerQualified = tc.checkDealerQualifies()

	// ローボーナスはフォールドしても評価される
	tc.evaluateLowBonus()

	tc.gameEndFlag = true
	tc.phase = ThreeCardRummyPhaseEnd
	tc.appendLog(-1, "result", "player folded", nil)
	return nil
}

// deal 3枚ずつ配る
func (tc *ThreeCardRummy) deal() {
	tc.playerHand = make([]*Card, 0, ThreeCardRummyHandSize)
	tc.dealerHand = make([]*Card, 0, ThreeCardRummyHandSize)
	for range ThreeCardRummyHandSize {
		tc.playerHand = append(tc.playerHand, tc.trumpCards.DrawCard())
		tc.dealerHand = append(tc.dealerHand, tc.trumpCards.DrawCard())
	}
	// **自分の点数はこの時点で決まる。** 見えている 3 枚の合計なので隠す理由が
	// 無く、むしろ play/fold を決める唯一の材料。resolve まで 0 のままだと
	// CUI も Web も「0 点 = 最強」を表示してしまう。
	tc.playerScore = ThreeCardRummyScore(tc.playerHand)
	tc.appendLog(-1, "deal", "dealt 3 cards each", nil)
}

// resolve ゲーム解決（Play後の処理）
func (tc *ThreeCardRummy) resolve() {
	tc.playerScore = ThreeCardRummyScore(tc.playerHand)
	tc.dealerScore = ThreeCardRummyScore(tc.dealerHand)
	tc.dealerQualified = tc.checkDealerQualifies()

	switch {
	case !tc.dealerQualified:
		// **勝負が成立していないので点数は比べない。** アンテが 1:1 で付き、
		// プレイは返るので、手の善し悪しに関わらず取り分はアンテぶん増える
		// (staked 2×ante, returned 3×ante)。ここで素の大小を入れると、
		// 30 点対 21 点で **+10 儲けた局面に「ディーラーの勝ち」と赤字で
		// 出す**ことになり、CUI の文言・ページの勝敗演出・実際の残高が
		// 三者三様にずれる。
		tc.result = GameResultWin
	// **低いほうが勝つ。** 比較の向きがこのゲームの肝で、素直に大小を取ると
	// 全部逆になる。
	case tc.playerScore < tc.dealerScore:
		tc.result = GameResultWin
	case tc.playerScore > tc.dealerScore:
		tc.result = GameResultLose
	default:
		tc.result = GameResultDraw
	}

	// 配当計算
	tc.calculatePayouts()
	// ローボーナス評価
	tc.evaluateLowBonus()
	// アンテボーナス評価
	tc.evaluateAnteBonus()

	// チップ加算
	totalPayout := tc.antePayout + tc.playPayout + tc.anteBonusPayout + tc.lowBonusPayout
	if totalPayout > 0 {
		tc.chips.AddChips(totalPayout)
	}

	tc.gameEndFlag = true
	tc.phase = ThreeCardRummyPhaseEnd

	var resultStr string
	switch tc.result {
	case GameResultWin:
		resultStr = "player wins"
	case GameResultDraw:
		resultStr = "push"
	default:
		resultStr = "dealer wins"
	}
	tc.appendLog(-1, "result", resultStr, nil)
}

// checkDealerQualifies はディーラーが勝負に応じるかを返す。
//
// **合計 20 点以下でクオリファイ。** 低いほど強いゲームなので、条件も
// 「役が一定以上」ではなく「点が一定以下」になる。クオリファイしなければ
// アンテだけが払い戻され、プレイベットはプッシュ。
func (tc *ThreeCardRummy) checkDealerQualifies() bool {
	return tc.dealerScore <= ThreeCardRummyDealerQualifyMax
}

// calculatePayouts アンテ/プレイの配当計算
func (tc *ThreeCardRummy) calculatePayouts() {
	if !tc.dealerQualified {
		// ディーラー未クオリファイ: アンテ1:1返却、プレイはプッシュ（返却）
		tc.antePayout = tc.anteBet * 2 // 元のベット + 1:1
		tc.playPayout = tc.playBet     // プッシュ（返却のみ）
		return
	}

	switch tc.result {
	case GameResultWin:
		tc.antePayout = tc.anteBet * 2 // 元のベット + 1:1
		tc.playPayout = tc.playBet * 2 // 元のベット + 1:1
	case GameResultDraw:
		tc.antePayout = tc.anteBet // プッシュ（返却のみ）
		tc.playPayout = tc.playBet // プッシュ（返却のみ）
	case GameResultLose:
		tc.antePayout = 0 // 没収
		tc.playPayout = 0 // 没収
	}
}

// evaluateAnteBonus はアンテボーナスを評価する（ディーラーの結果に関係なく）。
// ボーナスのみで、元ベットの返却は含まない。
//
// **しきい値は点数。** 0 点が最も厚く、点が上がるほど薄くなる。
func (tc *ThreeCardRummy) evaluateAnteBonus() {
	switch score := tc.playerScore; {
	case score == ThreeCardRummyPerfectScore:
		tc.anteBonusPayout = tc.anteBet * ThreeCardRummyAnteBonusPerfect
	case score <= ThreeCardRummyBonusVeryLowMax:
		tc.anteBonusPayout = tc.anteBet * ThreeCardRummyAnteBonusVeryLow
	case score <= ThreeCardRummyBonusLowMax:
		tc.anteBonusPayout = tc.anteBet * ThreeCardRummyAnteBonusLow
	}
}

// evaluateLowBonus はローボーナス（独立したサイドベット）を評価する。
//
// **ディーラーの手とは無関係**に、自分の点の低さだけで払う。だからフォールド
// しても評価される —— 降りたのは勝負であって、この賭けではない。
//
// クローン元の Three Card Poker は「ペアプラス」で、ペア以上の**役**に払っていた。
// このゲームに役は無いので、しきい値は点数に置き換わる。
func (tc *ThreeCardRummy) evaluateLowBonus() {
	if tc.lowBonusBet <= 0 {
		return
	}
	mult := 0
	switch score := tc.playerScore; {
	case score == ThreeCardRummyPerfectScore:
		mult = ThreeCardRummyLowBonusPerfect
	case score <= ThreeCardRummyBonusVeryLowMax:
		mult = ThreeCardRummyLowBonusVeryLow
	case score <= ThreeCardRummyBonusLowMax:
		mult = ThreeCardRummyLowBonusLow
	default:
		return // 払わない（賭け金は没収）
	}
	tc.lowBonusPayout = tc.lowBonusBet + tc.lowBonusBet*mult
}

// --- Getters ---

// GetPlayerHand プレイヤーハンド取得
func (tc *ThreeCardRummy) GetPlayerHand() []*Card { return tc.playerHand }

// GetDealerHand ディーラーハンド取得
func (tc *ThreeCardRummy) GetDealerHand() []*Card { return tc.dealerHand }

// GetPhase 現在のフェーズ
func (tc *ThreeCardRummy) GetPhase() int { return tc.phase }

// GetGameEndFlag ゲーム終了フラグ
func (tc *ThreeCardRummy) GetGameEndFlag() bool { return tc.gameEndFlag }

// GetAnteBet アンテベット額
func (tc *ThreeCardRummy) GetAnteBet() int { return tc.anteBet }

// GetLowBonusBet ローボーナスベット額
func (tc *ThreeCardRummy) GetLowBonusBet() int { return tc.lowBonusBet }

// GetPlayBet プレイベット額
func (tc *ThreeCardRummy) GetPlayBet() int { return tc.playBet }

// GetResult ゲーム結果
func (tc *ThreeCardRummy) GetResult() GameResult { return tc.result }

// GetAntePayout アンテ配当
func (tc *ThreeCardRummy) GetAntePayout() int { return tc.antePayout }

// GetPlayPayout プレイ配当
func (tc *ThreeCardRummy) GetPlayPayout() int { return tc.playPayout }

// GetAnteBonusPayout アンテボーナス配当
func (tc *ThreeCardRummy) GetAnteBonusPayout() int { return tc.anteBonusPayout }

// GetLowBonusPayout ローボーナス配当
func (tc *ThreeCardRummy) GetLowBonusPayout() int { return tc.lowBonusPayout }

// GetTotalPayout 合計配当
func (tc *ThreeCardRummy) GetTotalPayout() int {
	return tc.antePayout + tc.playPayout + tc.anteBonusPayout + tc.lowBonusPayout
}

// GetDealerQualified ディーラークオリファイ
func (tc *ThreeCardRummy) GetDealerQualified() bool { return tc.dealerQualified }

// GetPlayerScore プレイヤーの点数（低いほど強い）
func (tc *ThreeCardRummy) GetPlayerScore() int { return tc.playerScore }

// GetDealerScore ディーラーの点数（低いほど強い）
func (tc *ThreeCardRummy) GetDealerScore() int { return tc.dealerScore }

// GetChips チップ
func (tc *ThreeCardRummy) GetChips() int { return tc.chips.GetChips() }

// --- Test helpers ---

// SetPhase フェーズ設定（テスト用）
func (tc *ThreeCardRummy) SetPhase(phase int) { tc.phase = phase }

// SetPlayerHand プレイヤーハンド設定（テスト用）
func (tc *ThreeCardRummy) SetPlayerHand(cards []*Card) { tc.playerHand = cards }

// SetDealerHand ディーラーハンド設定（テスト用）
func (tc *ThreeCardRummy) SetDealerHand(cards []*Card) { tc.dealerHand = cards }

// SetAnteBet アンテベット額設定（テスト用）
func (tc *ThreeCardRummy) SetAnteBet(amount int) { tc.anteBet = amount }

// SetLowBonusBet ローボーナスベット額設定（テスト用）
func (tc *ThreeCardRummy) SetLowBonusBet(amount int) { tc.lowBonusBet = amount }

// SetPlayBet プレイベット額設定（テスト用）
func (tc *ThreeCardRummy) SetPlayBet(amount int) { tc.playBet = amount }

// SetResult ゲーム結果設定（テスト用）
func (tc *ThreeCardRummy) SetResult(result GameResult) { tc.result = result }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (tc *ThreeCardRummy) SetGameEndFlag(flag bool) { tc.gameEndFlag = flag }

// SetChips チップ設定（テスト用）
func (tc *ThreeCardRummy) SetChips(chips int) { tc.chips.SetChips(chips) }

// SetDealerQualified ディーラークオリファイ設定（テスト用）
func (tc *ThreeCardRummy) SetDealerQualified(qualified bool) { tc.dealerQualified = qualified }

// SetPlayerScore プレイヤーの点数（低いほど強い）設定（テスト用）
func (tc *ThreeCardRummy) SetPlayerScore(score int) { tc.playerScore = score }

// SetDealerScore ディーラーの点数（低いほど強い）設定（テスト用）
func (tc *ThreeCardRummy) SetDealerScore(score int) { tc.dealerScore = score }

// SetAntePayout アンテ配当設定（テスト用）
func (tc *ThreeCardRummy) SetAntePayout(payout int) { tc.antePayout = payout }

// SetPlayPayout プレイ配当設定（テスト用）
func (tc *ThreeCardRummy) SetPlayPayout(payout int) { tc.playPayout = payout }

// SetAnteBonusPayout アンテボーナス配当設定（テスト用）
func (tc *ThreeCardRummy) SetAnteBonusPayout(payout int) { tc.anteBonusPayout = payout }

// SetLowBonusPayout ローボーナス配当設定（テスト用）
func (tc *ThreeCardRummy) SetLowBonusPayout(payout int) { tc.lowBonusPayout = payout }

// threeCardRummyJSON is the JSON wire format for ThreeCardRummy.
type threeCardRummyJSON struct {
	TrumpCards      *TrumpCards       `json:"tc"`
	PlayerHand      []*Card           `json:"ph"`
	DealerHand      []*Card           `json:"dh"`
	Chips           *ChipHolder       `json:"ch"`
	AnteBet         int               `json:"ab"`
	LowBonusBet     int               `json:"lb"`
	LastAnteBet     int               `json:"lab"`
	LastLowBonusBet int               `json:"llb"`
	PlayBet         int               `json:"pb"`
	Phase           int               `json:"ps"`
	GameEndFlag     bool              `json:"ge"`
	Result          GameResult        `json:"rs"`
	AntePayout      int               `json:"ap"`
	PlayPayout      int               `json:"plp"`
	AnteBonusPayout int               `json:"abp"`
	LowBonusPayout  int               `json:"lbp"`
	DealerQualified bool              `json:"dq"`
	PlayerScore     int               `json:"ps2"`
	DealerScore     int               `json:"ds"`
	ActionLog       []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (tc *ThreeCardRummy) MarshalJSON() ([]byte, error) {
	return json.Marshal(threeCardRummyJSON{
		TrumpCards:      tc.trumpCards,
		PlayerHand:      tc.playerHand,
		DealerHand:      tc.dealerHand,
		Chips:           &tc.chips,
		AnteBet:         tc.anteBet,
		LowBonusBet:     tc.lowBonusBet,
		LastAnteBet:     tc.lastAnteBet,
		LastLowBonusBet: tc.lastLowBonusBet,
		PlayBet:         tc.playBet,
		Phase:           tc.phase,
		GameEndFlag:     tc.gameEndFlag,
		Result:          tc.result,
		AntePayout:      tc.antePayout,
		PlayPayout:      tc.playPayout,
		AnteBonusPayout: tc.anteBonusPayout,
		LowBonusPayout:  tc.lowBonusPayout,
		DealerQualified: tc.dealerQualified,
		PlayerScore:     tc.playerScore,
		DealerScore:     tc.dealerScore,
		ActionLog:       tc.actionLog,
	})
}

// threeCardRummyMaxSliceLen caps slice sizes during deserialisation.
const threeCardRummyMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (tc *ThreeCardRummy) UnmarshalJSON(data []byte) error {
	var j threeCardRummyJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.PlayerHand) > threeCardRummyMaxSliceLen || len(j.DealerHand) > threeCardRummyMaxSliceLen ||
		len(j.ActionLog) > threeCardRummyMaxSliceLen {
		return fmt.Errorf("threecardrummy: input array exceeds maximum allowed size")
	}

	tc.trumpCards = j.TrumpCards
	if tc.trumpCards == nil {
		tc.trumpCards = NewTrumpCards(0)
	}
	tc.playerHand = j.PlayerHand
	if tc.playerHand == nil {
		tc.playerHand = make([]*Card, 0)
	}
	tc.dealerHand = j.DealerHand
	if tc.dealerHand == nil {
		tc.dealerHand = make([]*Card, 0)
	}
	if j.Chips != nil {
		tc.chips = *j.Chips
	}
	tc.anteBet = j.AnteBet
	tc.lowBonusBet = j.LowBonusBet
	tc.lastAnteBet = j.LastAnteBet
	tc.lastLowBonusBet = j.LastLowBonusBet
	tc.playBet = j.PlayBet
	tc.phase = j.Phase
	tc.gameEndFlag = j.GameEndFlag
	tc.result = j.Result
	tc.antePayout = j.AntePayout
	tc.playPayout = j.PlayPayout
	tc.anteBonusPayout = j.AnteBonusPayout
	tc.lowBonusPayout = j.LowBonusPayout
	tc.dealerQualified = j.DealerQualified
	tc.playerScore = j.PlayerScore
	tc.dealerScore = j.DealerScore
	tc.actionLog = j.ActionLog
	if tc.actionLog == nil {
		tc.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
