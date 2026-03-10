package domain

import "fmt"

// バカラフェーズ定数
const (
	BaccaratPhaseBet = 1 // ベットフェーズ
	BaccaratPhaseEnd = 2 // 終了フェーズ
)

// バカラベットタイプ定数
const (
	BaccaratBetPlayer = 0 // プレイヤーベット
	BaccaratBetBanker = 1 // バンカーベット
	BaccaratBetTie    = 2 // タイベット
)

// バカラデフォルト値
const (
	BaccaratDefaultChips    = 1000  // デフォルトチップ
	BaccaratMinBet          = 10    // 最低ベット額
	BaccaratMaxBet          = 10000 // 最大ベット額
	BaccaratCommissionRate  = 5     // バンカーコミッション率(%)
	BaccaratTiePayoutRate   = 8     // タイ配当倍率
	BaccaratNaturalMinValue = 8     // ナチュラル最小値
	BaccaratDrawMaxValue    = 5     // プレイヤードロー閾値
	BaccaratStandMinValue   = 7     // スタンド最小値
)

// Baccarat バカラクラス
type Baccarat struct {
	trumpCards  *TrumpCards       // トランプカード
	playerHand  []*Card           // プレイヤーハンド
	bankerHand  []*Card           // バンカーハンド
	chips       ChipHolder        // チップ
	betAmount   int               // ベット額
	betType     int               // ベットタイプ
	phase       int               // 現在のフェーズ
	gameEndFlag bool              // ゲーム終了フラグ
	result      GameResult        // ゲーム結果
	payout      int               // 配当金額
	actionLog   []*ActionLogEntry // 棋譜
}

// NewBaccarat コンストラクタ
func NewBaccarat(trumpCards *TrumpCards) *Baccarat {
	trumpCards.Shuffle()
	b := &Baccarat{
		trumpCards: trumpCards,
		phase:      BaccaratPhaseBet,
	}
	return b
}

// NewDefaultBaccarat デフォルト設定のバカラを生成するファクトリ関数
func NewDefaultBaccarat() *Baccarat {
	b := NewBaccarat(NewTrumpCards(0))
	b.chips.SetChips(BaccaratDefaultChips)
	return b
}

// Reset ゲーム初期化
func (b *Baccarat) Reset() {
	b.gameEndFlag = false
	b.phase = BaccaratPhaseBet
	b.playerHand = nil
	b.bankerHand = nil
	b.betAmount = 0
	b.betType = 0
	b.result = 0
	b.payout = 0
	b.actionLog = nil
	if b.chips.GetChips() < BaccaratMinBet {
		b.chips.SetChips(BaccaratDefaultChips)
	}
	b.trumpCards = NewTrumpCards(0)
	for i := 0; i < 10; i++ {
		b.trumpCards.Shuffle()
	}
}

// Bet ベット＆ゲーム実行（バカラはベット後に全自動進行）
func (b *Baccarat) Bet(amount, betType int) error {
	if b.phase != BaccaratPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if betType < BaccaratBetPlayer || betType > BaccaratBetTie {
		return NewDomainError(ErrInvalidPlay, "Invalid bet type.")
	}
	if amount < BaccaratMinBet || amount%BaccaratMinBet != 0 || amount > BaccaratMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid bet amount.")
	}
	if !b.chips.SubtractChips(amount) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	b.betAmount = amount
	b.betType = betType
	b.appendLog(0, "bet", fmt.Sprintf("bet %d on %s", amount, betTypeName(betType)), nil)

	// ディール
	b.deal()
	// サードカードルール
	b.drawPhase()
	// 勝敗判定
	b.judge()
	return nil
}

// deal 2枚ずつ配る
func (b *Baccarat) deal() {
	b.playerHand = make([]*Card, 0, 3)
	b.bankerHand = make([]*Card, 0, 3)
	// プレイヤー1枚目、バンカー1枚目、プレイヤー2枚目、バンカー2枚目
	b.playerHand = append(b.playerHand, b.trumpCards.DrawCard())
	b.bankerHand = append(b.bankerHand, b.trumpCards.DrawCard())
	b.playerHand = append(b.playerHand, b.trumpCards.DrawCard())
	b.bankerHand = append(b.bankerHand, b.trumpCards.DrawCard())
	b.appendLog(-1, "deal", fmt.Sprintf("player=%d banker=%d",
		b.CalculateHandValue(b.playerHand), b.CalculateHandValue(b.bankerHand)), nil)
}

// drawPhase サードカードルール適用
func (b *Baccarat) drawPhase() {
	playerTotal := b.CalculateHandValue(b.playerHand)
	bankerTotal := b.CalculateHandValue(b.bankerHand)

	// ナチュラル判定（どちらかが8か9ならドロー無し）
	if playerTotal >= BaccaratNaturalMinValue || bankerTotal >= BaccaratNaturalMinValue {
		b.appendLog(-1, "natural", fmt.Sprintf("player=%d banker=%d", playerTotal, bankerTotal), nil)
		return
	}

	// プレイヤーのサードカード
	playerDrew := false
	var playerThirdCardValue int
	if b.shouldPlayerDraw(playerTotal) {
		card := b.trumpCards.DrawCard()
		b.playerHand = append(b.playerHand, card)
		playerDrew = true
		playerThirdCardValue = b.cardPointValue(card)
		b.appendLog(-1, "draw", fmt.Sprintf("player draws %d", playerThirdCardValue), []*Card{card})
	}

	// バンカーのサードカード
	if b.shouldBankerDraw(bankerTotal, playerThirdCardValue, playerDrew) {
		card := b.trumpCards.DrawCard()
		b.bankerHand = append(b.bankerHand, card)
		b.appendLog(-1, "draw", fmt.Sprintf("banker draws %d", b.cardPointValue(card)), []*Card{card})
	}
}

// judge 勝敗判定＆配当計算
func (b *Baccarat) judge() {
	playerTotal := b.CalculateHandValue(b.playerHand)
	bankerTotal := b.CalculateHandValue(b.bankerHand)

	if playerTotal > bankerTotal {
		b.result = GameResultWin
		b.appendLog(-1, "result", "player wins", nil)
	} else if bankerTotal > playerTotal {
		b.result = GameResultLose
		b.appendLog(-1, "result", "banker wins", nil)
	} else {
		b.result = GameResultDraw
		b.appendLog(-1, "result", "tie", nil)
	}

	// 配当計算
	b.payout = b.calculatePayout()
	b.chips.AddChips(b.payout)

	b.gameEndFlag = true
	b.phase = BaccaratPhaseEnd
}

// calculatePayout 配当計算
func (b *Baccarat) calculatePayout() int {
	switch b.betType {
	case BaccaratBetPlayer:
		if b.result == GameResultWin {
			return b.betAmount * 2
		}
		if b.result == GameResultDraw {
			return b.betAmount // 引き分けは返却
		}
		return 0
	case BaccaratBetBanker:
		if b.result == GameResultLose { // バンカー勝利
			// 5%コミッション
			winnings := b.betAmount - b.betAmount*BaccaratCommissionRate/100
			return b.betAmount + winnings
		}
		if b.result == GameResultDraw {
			return b.betAmount
		}
		return 0
	case BaccaratBetTie:
		if b.result == GameResultDraw {
			return b.betAmount + b.betAmount*BaccaratTiePayoutRate
		}
		return 0
	}
	return 0
}

// CalculateHandValue ハンドの合計値を計算（mod 10）
func (b *Baccarat) CalculateHandValue(cards []*Card) int {
	total := 0
	for _, card := range cards {
		total += b.cardPointValue(card)
	}
	return total % 10
}

// cardPointValue カードのポイント値（A=1, 2-9=額面, 10/J/Q/K=0）
func (b *Baccarat) cardPointValue(card *Card) int {
	v := card.GetValue()
	if v >= 10 {
		return 0
	}
	return v
}

// shouldPlayerDraw プレイヤーがサードカードを引くか（0-5なら引く）
func (b *Baccarat) shouldPlayerDraw(playerTotal int) bool {
	return playerTotal <= BaccaratDrawMaxValue
}

// shouldBankerDraw バンカーがサードカードを引くか
func (b *Baccarat) shouldBankerDraw(bankerTotal, playerThirdCardValue int, playerDrew bool) bool {
	// プレイヤーがドローしなかった場合、バンカーは0-5で引く
	if !playerDrew {
		return bankerTotal <= BaccaratDrawMaxValue
	}
	// プレイヤーがドローした場合のバンカールール
	switch bankerTotal {
	case 0, 1, 2:
		return true
	case 3:
		return playerThirdCardValue != 8
	case 4:
		return playerThirdCardValue >= 2 && playerThirdCardValue <= 7
	case 5:
		return playerThirdCardValue >= 4 && playerThirdCardValue <= 7
	case 6:
		return playerThirdCardValue == 6 || playerThirdCardValue == 7
	default: // 7+
		return false
	}
}

// betTypeName ベットタイプ名
func betTypeName(betType int) string {
	switch betType {
	case BaccaratBetPlayer:
		return "player"
	case BaccaratBetBanker:
		return "banker"
	case BaccaratBetTie:
		return "tie"
	default:
		return "unknown"
	}
}

// appendLog 棋譜にエントリを追加する
func (b *Baccarat) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	b.actionLog = append(b.actionLog, &ActionLogEntry{
		TurnNumber: len(b.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- Getters ---

// GetPlayerHand プレイヤーハンド取得
func (b *Baccarat) GetPlayerHand() []*Card { return b.playerHand }

// GetBankerHand バンカーハンド取得
func (b *Baccarat) GetBankerHand() []*Card { return b.bankerHand }

// GetPhase 現在のフェーズ
func (b *Baccarat) GetPhase() int { return b.phase }

// GetGameEndFlag ゲーム終了フラグ
func (b *Baccarat) GetGameEndFlag() bool { return b.gameEndFlag }

// GetBetAmount ベット額
func (b *Baccarat) GetBetAmount() int { return b.betAmount }

// GetBetType ベットタイプ
func (b *Baccarat) GetBetType() int { return b.betType }

// GetResult ゲーム結果
func (b *Baccarat) GetResult() GameResult { return b.result }

// GetPayout 配当金額
func (b *Baccarat) GetPayout() int { return b.payout }

// GetChips チップ
func (b *Baccarat) GetChips() int { return b.chips.GetChips() }

// GetPlayerHandValue プレイヤーハンド合計値
func (b *Baccarat) GetPlayerHandValue() int { return b.CalculateHandValue(b.playerHand) }

// GetBankerHandValue バンカーハンド合計値
func (b *Baccarat) GetBankerHandValue() int { return b.CalculateHandValue(b.bankerHand) }

// GetActionLog 棋譜を取得する
func (b *Baccarat) GetActionLog() []*ActionLogEntry { return b.actionLog }

// --- Test helpers ---

// SetPhase フェーズ設定（テスト用）
func (b *Baccarat) SetPhase(phase int) { b.phase = phase }

// SetPlayerHand プレイヤーハンド設定（テスト用）
func (b *Baccarat) SetPlayerHand(cards []*Card) { b.playerHand = cards }

// SetBankerHand バンカーハンド設定（テスト用）
func (b *Baccarat) SetBankerHand(cards []*Card) { b.bankerHand = cards }

// SetBetAmount ベット額設定（テスト用）
func (b *Baccarat) SetBetAmount(amount int) { b.betAmount = amount }

// SetBetType ベットタイプ設定（テスト用）
func (b *Baccarat) SetBetType(betType int) { b.betType = betType }

// SetResult ゲーム結果設定（テスト用）
func (b *Baccarat) SetResult(result GameResult) { b.result = result }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (b *Baccarat) SetGameEndFlag(flag bool) { b.gameEndFlag = flag }

// SetChips チップ設定（テスト用）
func (b *Baccarat) SetChips(chips int) { b.chips.SetChips(chips) }
