//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

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
	handKey     string
	actionLogBase
	config *VideoPokerVariantConfig
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
	vp.handKey = ""
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
		vp.handKey = videoPokerHandKey(handName)
	} else {
		vp.result = GameResultLose
		vp.handName = ""
		vp.handKey = ""
	}
	displayName := handName
	if displayName == "" {
		if rank >= 0 && rank < len(PokerHandNames) {
			displayName = PokerHandNames[rank]
		}
	}
	vp.appendLog(0, "result", fmt.Sprintf("%s payout=%d", displayName, vp.payout), vp.hand)
}

// videoPokerHandKey maps a variant hand name (the English string returned by a
// VideoPokerVariantConfig.GetResult) to a stable, locale-independent key. The
// key matches the frontend payout-table row keys
// (frontend/src/utils/videoPokerPayout.ts) and the "<variant>.hand.<key>" CUI
// translation keys, so both clients can localize the hand without reverse-looking
// up the English string. It returns "" for an unknown or empty (losing) name.
func videoPokerHandKey(handName string) string {
	switch handName {
	case "Royal Flush":
		return "royalFlush"
	case "Natural Royal Flush":
		return "naturalRoyalFlush"
	case "Wild Royal Flush":
		return "wildRoyalFlush"
	case "Four Deuces":
		return "fourDeuces"
	case "Five of a Kind":
		return "fiveOfAKind"
	case "Straight Flush":
		return "straightFlush"
	case "Four of a Kind":
		return "fourOfAKind"
	case "Full House":
		return "fullHouse"
	case "Flush":
		return "flush"
	case "Straight":
		return "straight"
	case "Three of a Kind":
		return "threeOfAKind"
	case "Two Pair":
		return "twoPair"
	case "Jacks or Better":
		return "jacksOrBetter"
	case "Kings or Better":
		return "kingsOrBetter"
	default:
		return ""
	}
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

// GetCurrentHandKey はいま手元にある5枚が配当対象の役かを、ロケール非依存の
// 安定キーで返す (配当が付かないなら空文字)。
//
// ドロー中に「現在の役」を出すためのもの。Web は evaluateJokerPokerMadeHand で
// 同じことをしているのに、CUI は手札とホールド推奨しか出していなかった (#5508)。
//
// **評価はバリアント自身の GetResult を通す。**別に書くと、ワイルドの扱いや配当の
// 下限がバリアントごとにずれる。ベット額は royal のジャックポット段だけに効き、
// キーには影響しない。
//
// **状態を一切変えない。**チップも結果も動かさないので、描画のたびに呼んでよい。
func (vp *VideoPoker) GetCurrentHandKey() string {
	if len(vp.hand) != 5 {
		return ""
	}
	// 配当の付かない手は GetResult が名前を返さないので、キーも空になる
	// (videoPokerHandKey の default)。**その規約はテストで固定してある** --
	// 破れたら「払われない役名」が現在の役として出てしまう。
	_, _, handName := vp.config.GetResult(vp.hand, vp.betAmount)
	return videoPokerHandKey(handName)
}

// GetHandKey は役の安定キー（ロケール非依存）を返す。役なし時は空文字。
func (vp *VideoPoker) GetHandKey() string { return vp.handKey }

// GetHeldIndices ホールドインデックス
func (vp *VideoPoker) GetHeldIndices() [VideoPokerHandSize]bool { return vp.heldIndices }

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

// SetHandKey 役キー設定（テスト用）
func (vp *VideoPoker) SetHandKey(key string) { vp.handKey = key }

// videoPokerJSON is the JSON wire format for VideoPoker.
type videoPokerJSON struct {
	TrumpCards  *TrumpCards              `json:"tc"`
	Hand        []*Card                  `json:"hd"`
	Chips       *ChipHolder              `json:"ch"`
	BetAmount   int                      `json:"ba"`
	HeldIndices [VideoPokerHandSize]bool `json:"hi"`
	Phase       int                      `json:"ps"`
	GameEndFlag bool                     `json:"ge"`
	Result      GameResult               `json:"rs"`
	Payout      int                      `json:"po"`
	HandRank    int                      `json:"hr"`
	HandName    string                   `json:"hn"`
	HandKey     string                   `json:"hk"`
	ActionLog   []*ActionLogEntry        `json:"al"`
	ConfigName  string                   `json:"cn"`
	JokerCount  int                      `json:"jc"`
}

// MarshalJSON implements json.Marshaler.
func (vp *VideoPoker) MarshalJSON() ([]byte, error) {
	j := videoPokerJSON{
		TrumpCards:  vp.trumpCards,
		Hand:        vp.hand,
		Chips:       &vp.chips,
		BetAmount:   vp.betAmount,
		HeldIndices: vp.heldIndices,
		Phase:       vp.phase,
		GameEndFlag: vp.gameEndFlag,
		Result:      vp.result,
		Payout:      vp.payout,
		HandRank:    vp.handRank,
		HandName:    vp.handName,
		HandKey:     vp.handKey,
		ActionLog:   vp.actionLog,
	}
	if vp.config != nil {
		j.ConfigName = vp.config.Name
		j.JokerCount = vp.config.JokerCount
	}
	return json.Marshal(j)
}

// videoPokerMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const videoPokerMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (vp *VideoPoker) UnmarshalJSON(data []byte) error {
	var j videoPokerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Hand) > videoPokerMaxSliceLen || len(j.ActionLog) > videoPokerMaxSliceLen {
		return fmt.Errorf("videopoker: input array exceeds maximum allowed size")
	}

	vp.trumpCards = j.TrumpCards
	if vp.trumpCards == nil {
		vp.trumpCards = NewTrumpCards(j.JokerCount)
	}
	vp.hand = j.Hand
	if vp.hand == nil {
		vp.hand = make([]*Card, 0)
	}
	if j.Chips != nil {
		vp.chips = *j.Chips
	}
	vp.betAmount = j.BetAmount
	vp.heldIndices = j.HeldIndices
	vp.phase = j.Phase
	vp.gameEndFlag = j.GameEndFlag
	vp.result = j.Result
	vp.payout = j.Payout
	vp.handRank = j.HandRank
	vp.handName = j.HandName
	vp.handKey = j.HandKey
	vp.actionLog = j.ActionLog
	if vp.actionLog == nil {
		vp.actionLog = make([]*ActionLogEntry, 0)
	}
	vp.config = resolveVideoPokerConfig(j.ConfigName)
	return nil
}

// resolveVideoPokerConfig restores a VideoPokerVariantConfig by name.
func resolveVideoPokerConfig(name string) *VideoPokerVariantConfig {
	switch name {
	case "deuceswild":
		return DeucesWildConfig()
	case "jokerpoker":
		return JokerPokerConfig()
	default:
		return JacksOrBetterConfig()
	}
}
