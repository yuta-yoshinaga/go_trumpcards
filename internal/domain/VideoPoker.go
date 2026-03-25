package domain

import "fmt"

// ビデオポーカーフェーズ定数
const (
	VideoPokerPhaseBet    = 1 // ベットフェーズ
	VideoPokerPhaseDraw   = 2 // ドローフェーズ（ホールド選択）
	VideoPokerPhaseResult = 3 // 結果フェーズ
)

// ビデオポーカーデフォルト値
const (
	VideoPokerDefaultChips = 1000 // デフォルトチップ
	VideoPokerMinBet       = 1    // 最低ベット（コイン）
	VideoPokerMaxBet       = 5    // 最大ベット（コイン）
	VideoPokerHandSize     = 5    // ハンドサイズ
)

// VideoPoker ビデオポーカークラス
type VideoPoker struct {
	trumpCards  *TrumpCards
	hand        []*Card
	chips       ChipHolder
	betAmount   int
	heldIndices [VideoPokerHandSize]bool
	phase       int
	gameEndFlag bool
	result      GameResult
	payout      int
	handRank    int
	handName    string
	actionLog   []*ActionLogEntry
	config      *VideoPokerVariantConfig
}

// NewVideoPoker コンストラクタ
func NewVideoPoker(trumpCards *TrumpCards, config *VideoPokerVariantConfig) *VideoPoker {
	trumpCards.Shuffle()
	return &VideoPoker{
		trumpCards: trumpCards,
		phase:      VideoPokerPhaseBet,
		config:     config,
	}
}

// NewDefaultVideoPoker デフォルト設定のビデオポーカー（Jacks or Better）を生成するファクトリ関数
func NewDefaultVideoPoker() *VideoPoker {
	vp := NewVideoPoker(NewTrumpCards(0), JacksOrBetterConfig())
	vp.chips.SetChips(VideoPokerDefaultChips)
	return vp
}

// NewDeucesWildVideoPoker Deuces Wildバリアントを生成するファクトリ関数
func NewDeucesWildVideoPoker() *VideoPoker {
	vp := NewVideoPoker(NewTrumpCards(0), DeucesWildConfig())
	vp.chips.SetChips(VideoPokerDefaultChips)
	return vp
}

// NewJokerPokerVideoPoker Joker Pokerバリアントを生成するファクトリ関数
func NewJokerPokerVideoPoker() *VideoPoker {
	vp := NewVideoPoker(NewTrumpCards(1), JokerPokerConfig())
	vp.chips.SetChips(VideoPokerDefaultChips)
	return vp
}

// Reset ゲーム初期化
func (vp *VideoPoker) Reset() {
	vp.gameEndFlag = false
	vp.phase = VideoPokerPhaseBet
	vp.hand = nil
	vp.betAmount = 0
	vp.heldIndices = [VideoPokerHandSize]bool{}
	vp.result = 0
	vp.payout = 0
	vp.handRank = 0
	vp.handName = ""
	vp.actionLog = nil
	if vp.chips.GetChips() < VideoPokerMinBet {
		vp.chips.SetChips(VideoPokerDefaultChips)
	}
	vp.trumpCards = NewTrumpCards(vp.config.JokerCount)
	vp.trumpCards.Shuffle()
}

// Bet ベット＆ディール（1〜5コイン）
func (vp *VideoPoker) Bet(amount int) error {
	if vp.phase != VideoPokerPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if amount < VideoPokerMinBet || amount > VideoPokerMaxBet {
		return NewDomainError(ErrInvalidAmount, "Bet must be between 1 and 5 coins.")
	}
	if !vp.chips.SubtractChips(amount) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	vp.betAmount = amount
	vp.appendLog(0, "bet", fmt.Sprintf("bet %d coin(s)", amount), nil)

	// ディール: 5枚配る
	vp.hand = make([]*Card, VideoPokerHandSize)
	for i := range VideoPokerHandSize {
		vp.hand[i] = vp.trumpCards.DrawCard()
	}
	vp.appendLog(0, "deal", "dealt 5 cards", vp.hand)

	vp.phase = VideoPokerPhaseDraw
	return nil
}

// Hold ホールド選択＆ドロー（indicesはキープするカードの0-basedインデックス）
func (vp *VideoPoker) Hold(indices []int) error {
	if vp.phase != VideoPokerPhaseDraw {
		return NewDomainError(ErrWrongPhase, "Hold is only allowed during the draw phase.")
	}
	// インデックスバリデーション
	seen := [VideoPokerHandSize]bool{}
	for _, idx := range indices {
		if idx < 0 || idx >= VideoPokerHandSize {
			return NewDomainError(ErrInvalidCard, "Card index must be between 0 and 4.")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidCard, "Duplicate card index.")
		}
		seen[idx] = true
	}

	// ホールドフラグを設定
	vp.heldIndices = [VideoPokerHandSize]bool{}
	for _, idx := range indices {
		vp.heldIndices[idx] = true
	}

	// ホールドされていないカードを交換
	replacedCount := 0
	for i := range VideoPokerHandSize {
		if !vp.heldIndices[i] {
			vp.hand[i] = vp.trumpCards.DrawCard()
			replacedCount++
		}
	}
	vp.appendLog(0, "draw", fmt.Sprintf("held %d, drew %d", len(indices), replacedCount), vp.hand)

	// 役判定＆配当計算
	vp.evaluate()
	vp.phase = VideoPokerPhaseResult
	vp.gameEndFlag = true
	return nil
}

// evaluate ハンド評価＆配当計算
func (vp *VideoPoker) evaluate() {
	rank, multiplier, handName := vp.config.GetResult(vp.hand, vp.betAmount)
	vp.handRank = rank
	vp.payout = vp.betAmount * multiplier
	vp.chips.AddChips(vp.payout)

	if vp.payout > 0 {
		vp.result = GameResultWin
		vp.handName = handName
	} else {
		vp.result = GameResultLose
		vp.handName = ""
	}
	displayName := handName
	if displayName == "" {
		if rank >= 0 && rank < len(PokerHandNames) {
			displayName = PokerHandNames[rank]
		}
	}
	vp.appendLog(0, "result", fmt.Sprintf("%s payout=%d", displayName, vp.payout), vp.hand)
}

// appendLog 棋譜にエントリを追加する
func (vp *VideoPoker) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	vp.actionLog = append(vp.actionLog, &ActionLogEntry{
		TurnNumber: len(vp.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- Getters ---

// GetHand ハンド取得
func (vp *VideoPoker) GetHand() []*Card { return vp.hand }

// GetPhase 現在のフェーズ
func (vp *VideoPoker) GetPhase() int { return vp.phase }

// GetGameEndFlag ゲーム終了フラグ
func (vp *VideoPoker) GetGameEndFlag() bool { return vp.gameEndFlag }

// GetBetAmount ベット額
func (vp *VideoPoker) GetBetAmount() int { return vp.betAmount }

// GetChips チップ
func (vp *VideoPoker) GetChips() int { return vp.chips.GetChips() }

// GetResult ゲーム結果
func (vp *VideoPoker) GetResult() GameResult { return vp.result }

// GetPayout 配当金額
func (vp *VideoPoker) GetPayout() int { return vp.payout }

// GetHandRank ハンドランク
func (vp *VideoPoker) GetHandRank() int { return vp.handRank }

// GetHandName ハンド名
func (vp *VideoPoker) GetHandName() string { return vp.handName }

// GetHeldIndices ホールドインデックス
func (vp *VideoPoker) GetHeldIndices() [VideoPokerHandSize]bool { return vp.heldIndices }

// GetActionLog 棋譜を取得する
func (vp *VideoPoker) GetActionLog() []*ActionLogEntry { return vp.actionLog }

// GetVariantName バリアント名を取得する
func (vp *VideoPoker) GetVariantName() string { return vp.config.Name }

// --- Test helpers ---

// SetPhase フェーズ設定（テスト用）
func (vp *VideoPoker) SetPhase(phase int) { vp.phase = phase }

// SetHand ハンド設定（テスト用）
func (vp *VideoPoker) SetHand(cards []*Card) { vp.hand = cards }

// SetBetAmount ベット額設定（テスト用）
func (vp *VideoPoker) SetBetAmount(amount int) { vp.betAmount = amount }

// SetChips チップ設定（テスト用）
func (vp *VideoPoker) SetChips(chips int) { vp.chips.SetChips(chips) }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (vp *VideoPoker) SetGameEndFlag(flag bool) { vp.gameEndFlag = flag }

// SetResult ゲーム結果設定（テスト用）
func (vp *VideoPoker) SetResult(result GameResult) { vp.result = result }

// SetHeldIndices ホールドインデックス設定（テスト用）
func (vp *VideoPoker) SetHeldIndices(held [VideoPokerHandSize]bool) { vp.heldIndices = held }

// SetPayout 配当金額設定（テスト用）
func (vp *VideoPoker) SetPayout(payout int) { vp.payout = payout }

// SetHandRank ハンドランク設定（テスト用）
func (vp *VideoPoker) SetHandRank(rank int) { vp.handRank = rank }

// SetHandName ハンド名設定（テスト用）
func (vp *VideoPoker) SetHandName(name string) { vp.handName = name }
