//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// テキサスホールデムボーナスポーカーフェーズ定数
const (
	TexasHoldemBonusPhaseBet     = 1 // ベットフェーズ（アンテとボーナスサイドベット）
	TexasHoldemBonusPhasePreFlop = 2 // プリフロップ（2枚のホールカード後、フォップベット or フォールド）
	TexasHoldemBonusPhaseFlop    = 3 // フロップ後（チェック or レイズ 1×アンテ）
	TexasHoldemBonusPhaseTurn    = 4 // ターン後（チェック or レイズ 1×アンテ）
	TexasHoldemBonusPhaseEnd     = 5 // 終了フェーズ
)

// テキサスホールデムボーナスポーカーデフォルト値
const (
	TexasHoldemBonusDefaultChips = 1000  // デフォルトチップ
	TexasHoldemBonusMinBet       = 10    // 最低ベット額
	TexasHoldemBonusMaxBet       = 10000 // 最大ベット額
	TexasHoldemBonusHoleCards    = 2     // ホールカード枚数
	TexasHoldemBonusBoardCards   = 5     // コミュニティカード枚数
)

// アンテボーナス配当倍率（ストレート以上、ディーラーのハンドに関係なく支払い）
const (
	TexasHoldemBonusAntePayRoyalFlush    = 1000 // ロイヤルフラッシュ 1000:1
	TexasHoldemBonusAntePayStraightFlush = 200  // ストレートフラッシュ 200:1
	TexasHoldemBonusAntePayFourOfAKind   = 25   // フォーカード 25:1
	TexasHoldemBonusAntePayFullHouse     = 5    // フルハウス 5:1
	TexasHoldemBonusAntePayFlush         = 4    // フラッシュ 4:1
	TexasHoldemBonusAntePayStraight      = 1    // ストレート 1:1
)

// ボーナスサイドベット配当倍率（プレイヤーの2枚のホールカードに対して）
const (
	TexasHoldemBonusBonusPayAA         = 30 // ポケットエース 30:1
	TexasHoldemBonusBonusPayAKSuited   = 25 // AKスーテッド 25:1
	TexasHoldemBonusBonusPayAQAJSuited = 20 // AQ/AJスーテッド 20:1
	TexasHoldemBonusBonusPayAKOff      = 15 // AKオフスート 15:1
	TexasHoldemBonusBonusPayKKQQJJ     = 10 // KK/QQ/JJ 10:1
	TexasHoldemBonusBonusPayAQAJOff    = 5  // AQ/AJオフスート 5:1
	TexasHoldemBonusBonusPayMediumPair = 3  // 22〜TT 3:1
)

// TexasHoldemBonus テキサスホールデムボーナスポーカークラス
type TexasHoldemBonus struct {
	trumpCards     *TrumpCards       // トランプカード
	playerHand     []*Card           // プレイヤーホールカード
	dealerHand     []*Card           // ディーラーホールカード
	community      []*Card           // コミュニティカード
	chips          ChipHolder        // チップ
	anteBet        int               // アンテベット額
	bonusBet       int               // ボーナスサイドベット額
	flopBet        int               // フロップベット額（プリフロップで2×アンテ）
	turnBet        int               // ターンベット額（フロップ後で1×アンテ）
	riverBet       int               // リバーベット額（ターン後で1×アンテ）
	phase          int               // 現在のフェーズ
	gameEndFlag    bool              // ゲーム終了フラグ
	result         GameResult        // ゲーム結果
	antePayout     int               // アンテ＋アンテボーナス配当
	playPayout     int               // プレイベット配当合計（フロップ＋ターン＋リバー）
	bonusPayout    int               // ボーナスサイドベット配当
	playerHandRank int               // プレイヤー最良5枚ランク
	dealerHandRank int               // ディーラー最良5枚ランク
	playerBest     []*Card           // プレイヤー最良5枚
	dealerBest     []*Card           // ディーラー最良5枚
	actionLog      []*ActionLogEntry // 棋譜
}

// NewTexasHoldemBonus コンストラクタ
func NewTexasHoldemBonus(trumpCards *TrumpCards) *TexasHoldemBonus {
	trumpCards.Shuffle()
	return &TexasHoldemBonus{
		trumpCards: trumpCards,
		phase:      TexasHoldemBonusPhaseBet,
	}
}

// NewDefaultTexasHoldemBonus デフォルト設定でゲームを生成するファクトリ関数
func NewDefaultTexasHoldemBonus() *TexasHoldemBonus {
	t := NewTexasHoldemBonus(NewTrumpCards(0))
	t.chips.SetChips(TexasHoldemBonusDefaultChips)
	return t
}

// Reset ゲーム初期化
func (t *TexasHoldemBonus) Reset() {
	t.gameEndFlag = false
	t.phase = TexasHoldemBonusPhaseBet
	t.playerHand = nil
	t.dealerHand = nil
	t.community = nil
	t.anteBet = 0
	t.bonusBet = 0
	t.flopBet = 0
	t.turnBet = 0
	t.riverBet = 0
	t.result = 0
	t.antePayout = 0
	t.playPayout = 0
	t.bonusPayout = 0
	t.playerHandRank = 0
	t.dealerHandRank = 0
	t.playerBest = nil
	t.dealerBest = nil
	t.actionLog = nil
	if t.chips.GetChips() < TexasHoldemBonusMinBet {
		t.chips.SetChips(TexasHoldemBonusDefaultChips)
	}
	t.trumpCards = NewTrumpCards(0)
	for range 10 {
		t.trumpCards.Shuffle()
	}
}

// Bet アンテベット＋オプションのボーナスサイドベット。ホールカード（各2枚）を配る。
func (t *TexasHoldemBonus) Bet(ante, bonus int) error {
	if t.phase != TexasHoldemBonusPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if ante < TexasHoldemBonusMinBet || ante%TexasHoldemBonusMinBet != 0 || ante > TexasHoldemBonusMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid ante amount.")
	}
	if bonus < 0 {
		return NewDomainError(ErrInvalidAmount, "Bonus bet must not be negative.")
	}
	if bonus > 0 && (bonus < TexasHoldemBonusMinBet || bonus%TexasHoldemBonusMinBet != 0 || bonus > TexasHoldemBonusMaxBet) {
		return NewDomainError(ErrInvalidAmount, "Invalid bonus bet amount.")
	}
	totalCost := ante + bonus
	if !t.chips.SubtractChips(totalCost) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	t.anteBet = ante
	t.bonusBet = bonus
	t.appendLog(0, "bet", fmt.Sprintf("ante=%d bonus=%d", ante, bonus), nil)

	t.dealHole()
	t.phase = TexasHoldemBonusPhasePreFlop
	return nil
}

// Play プリフロップでフロップベット（2×アンテ）を置きフロップを公開する。
func (t *TexasHoldemBonus) Play() error {
	if t.phase != TexasHoldemBonusPhasePreFlop {
		return NewDomainError(ErrWrongPhase, "Play is only allowed during the pre-flop phase.")
	}
	bet := t.anteBet * 2
	if !t.chips.SubtractChips(bet) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for flop bet.")
	}
	t.flopBet = bet
	t.appendLog(0, "play", fmt.Sprintf("flop bet=%d", bet), nil)

	t.dealFlop()
	t.phase = TexasHoldemBonusPhaseFlop
	return nil
}

// Fold プリフロップでフォールド。アンテは没収、ボーナスは別途評価。
func (t *TexasHoldemBonus) Fold() error {
	if t.phase != TexasHoldemBonusPhasePreFlop {
		return NewDomainError(ErrWrongPhase, "Fold is only allowed during the pre-flop phase.")
	}
	t.appendLog(0, "fold", "player folds", nil)

	t.result = GameResultLose
	t.evaluateBonus()
	if t.bonusPayout > 0 {
		t.chips.AddChips(t.bonusPayout)
	}
	t.gameEndFlag = true
	t.phase = TexasHoldemBonusPhaseEnd
	t.appendLog(-1, "result", "player folded", nil)
	return nil
}

// Check フロップ後またはターン後にチェック（ベットせず次フェーズへ）。
func (t *TexasHoldemBonus) Check() error {
	switch t.phase {
	case TexasHoldemBonusPhaseFlop:
		t.appendLog(0, "check", "flop check", nil)
		t.dealTurn()
		t.phase = TexasHoldemBonusPhaseTurn
		return nil
	case TexasHoldemBonusPhaseTurn:
		t.appendLog(0, "check", "turn check", nil)
		t.dealRiver()
		t.resolve()
		return nil
	default:
		return NewDomainError(ErrWrongPhase, "Check is only allowed during the flop or turn phase.")
	}
}

// Raise フロップ後またはターン後に1×アンテを置いて次フェーズへ進む。
func (t *TexasHoldemBonus) Raise() error {
	bet := t.anteBet
	switch t.phase {
	case TexasHoldemBonusPhaseFlop:
		if !t.chips.SubtractChips(bet) {
			return NewDomainError(ErrInsufficientChips, "Insufficient chips for turn bet.")
		}
		t.turnBet = bet
		t.appendLog(0, "raise", fmt.Sprintf("turn bet=%d", bet), nil)
		t.dealTurn()
		t.phase = TexasHoldemBonusPhaseTurn
		return nil
	case TexasHoldemBonusPhaseTurn:
		if !t.chips.SubtractChips(bet) {
			return NewDomainError(ErrInsufficientChips, "Insufficient chips for river bet.")
		}
		t.riverBet = bet
		t.appendLog(0, "raise", fmt.Sprintf("river bet=%d", bet), nil)
		t.dealRiver()
		t.resolve()
		return nil
	default:
		return NewDomainError(ErrWrongPhase, "Raise is only allowed during the flop or turn phase.")
	}
}

// dealHole 各2枚のホールカードを配る。
func (t *TexasHoldemBonus) dealHole() {
	t.playerHand = make([]*Card, 0, TexasHoldemBonusHoleCards)
	t.dealerHand = make([]*Card, 0, TexasHoldemBonusHoleCards)
	for range TexasHoldemBonusHoleCards {
		t.playerHand = append(t.playerHand, t.trumpCards.DrawCard())
		t.dealerHand = append(t.dealerHand, t.trumpCards.DrawCard())
	}
	t.appendLog(-1, "deal", "dealt 2 hole cards each", nil)
}

// dealFlop 3枚のフロップを配る。
func (t *TexasHoldemBonus) dealFlop() {
	t.community = make([]*Card, 0, TexasHoldemBonusBoardCards)
	for range 3 {
		t.community = append(t.community, t.trumpCards.DrawCard())
	}
	t.updatePlayerCurrentRank()
	t.appendLog(-1, "flop", "flop dealt", nil)
}

// dealTurn ターンを配る。
func (t *TexasHoldemBonus) dealTurn() {
	t.community = append(t.community, t.trumpCards.DrawCard())
	t.updatePlayerCurrentRank()
	t.appendLog(-1, "turn", "turn dealt", nil)
}

// updatePlayerCurrentRank プレイヤーの現在の最良ハンドランクを更新する。
// フロップ／ターン時点でプレイヤーのホール＋コミュニティから最良5枚を評価する。
// ヒント生成（フロントエンド）が中盤の手の強さを参照できるようにするため。
// resolve() がリバー後に同じフィールドを上書きする。
func (t *TexasHoldemBonus) updatePlayerCurrentRank() {
	all := append([]*Card{}, t.playerHand...)
	all = append(all, t.community...)
	t.playerHandRank, _ = evalBestFromSeven(all)
}

// dealRiver リバーを配る。
func (t *TexasHoldemBonus) dealRiver() {
	t.community = append(t.community, t.trumpCards.DrawCard())
	t.appendLog(-1, "river", "river dealt", nil)
}

// resolve ショーダウン処理（リバー後）
func (t *TexasHoldemBonus) resolve() {
	playerAll := append([]*Card{}, t.playerHand...)
	playerAll = append(playerAll, t.community...)
	dealerAll := append([]*Card{}, t.dealerHand...)
	dealerAll = append(dealerAll, t.community...)

	t.playerHandRank, t.playerBest = evalBestFromSeven(playerAll)
	t.dealerHandRank, t.dealerBest = evalBestFromSeven(dealerAll)

	cmp := t.compareBest()
	switch {
	case cmp > 0:
		t.result = GameResultWin
	case cmp < 0:
		t.result = GameResultLose
	default:
		t.result = GameResultDraw
	}

	t.calculatePayouts()
	t.evaluateBonus()

	totalPayout := t.antePayout + t.playPayout + t.bonusPayout
	if totalPayout > 0 {
		t.chips.AddChips(totalPayout)
	}

	t.gameEndFlag = true
	t.phase = TexasHoldemBonusPhaseEnd

	var resultStr string
	switch t.result {
	case GameResultWin:
		resultStr = "player wins"
	case GameResultDraw:
		resultStr = "push"
	default:
		resultStr = "dealer wins"
	}
	t.appendLog(-1, "result", resultStr, nil)
}

// compareBest プレイヤーとディーラーの最良5枚を比較する
func (t *TexasHoldemBonus) compareBest() int {
	if t.playerHandRank > t.dealerHandRank {
		return 1
	}
	if t.playerHandRank < t.dealerHandRank {
		return -1
	}
	return compareHighCardsSlice(t.playerBest, t.dealerBest)
}

// totalPlayBet フロップ＋ターン＋リバーの合計ベット額
func (t *TexasHoldemBonus) totalPlayBet() int {
	return t.flopBet + t.turnBet + t.riverBet
}

// calculatePayouts アンテ／プレイベット／アンテボーナスの配当計算（ショーダウン時）
func (t *TexasHoldemBonus) calculatePayouts() {
	playBet := t.totalPlayBet()
	switch t.result {
	case GameResultWin:
		t.antePayout = t.anteBet * 2
		t.playPayout = playBet * 2
	case GameResultDraw:
		t.antePayout = t.anteBet
		t.playPayout = playBet
	case GameResultLose:
		t.antePayout = 0
		t.playPayout = 0
	}
	if mult := t.anteBonusMultiplier(); mult > 0 {
		t.antePayout += t.anteBet * mult
	}
}

// anteBonusMultiplier プレイヤーが Straight 以上を作ったときのアンテボーナス倍率を返す。
// 勝敗に関係なく支払う。
func (t *TexasHoldemBonus) anteBonusMultiplier() int {
	switch t.playerHandRank {
	case PokerHandRoyalFlush:
		return TexasHoldemBonusAntePayRoyalFlush
	case PokerHandStraightFlush:
		return TexasHoldemBonusAntePayStraightFlush
	case PokerHandFourOfAKind:
		return TexasHoldemBonusAntePayFourOfAKind
	case PokerHandFullHouse:
		return TexasHoldemBonusAntePayFullHouse
	case PokerHandFlush:
		return TexasHoldemBonusAntePayFlush
	case PokerHandStraight:
		return TexasHoldemBonusAntePayStraight
	default:
		return 0
	}
}

// evaluateBonus サイドベット評価（プレイヤーの2枚のホールカード）
func (t *TexasHoldemBonus) evaluateBonus() {
	if t.bonusBet <= 0 || len(t.playerHand) < 2 {
		return
	}
	mult := bonusMultiplier(t.playerHand[0], t.playerHand[1])
	if mult > 0 {
		t.bonusPayout = t.bonusBet + t.bonusBet*mult
	}
}

// bonusMultiplier 2枚のホールカードからボーナス倍率を計算する。
// Card values are 1=Ace (treated as high), 2..10, 11=J, 12=Q, 13=K.
func bonusMultiplier(a, b *Card) int {
	va, vb := a.GetValue(), b.GetValue()
	suited := a.GetDesign() == b.GetDesign()

	aceCount := 0
	if va == 1 {
		aceCount++
	}
	if vb == 1 {
		aceCount++
	}

	// ペア
	if va == vb {
		switch va {
		case 1:
			return TexasHoldemBonusBonusPayAA
		case 13, 12, 11:
			return TexasHoldemBonusBonusPayKKQQJJ
		default:
			if va >= 2 && va <= 10 {
				return TexasHoldemBonusBonusPayMediumPair
			}
		}
		return 0
	}

	// 非ペア。Ace を片方含む場合のみペイアウト対象。
	if aceCount != 1 {
		return 0
	}
	other := va
	if va == 1 {
		other = vb
	}
	switch other {
	case 13:
		if suited {
			return TexasHoldemBonusBonusPayAKSuited
		}
		return TexasHoldemBonusBonusPayAKOff
	case 12, 11:
		if suited {
			return TexasHoldemBonusBonusPayAQAJSuited
		}
		return TexasHoldemBonusBonusPayAQAJOff
	}
	return 0
}

// appendLog 棋譜にエントリを追加する
func (t *TexasHoldemBonus) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	t.actionLog = append(t.actionLog, &ActionLogEntry{
		TurnNumber: len(t.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- Getters ---

// GetPlayerHand プレイヤーホールカード取得
func (t *TexasHoldemBonus) GetPlayerHand() []*Card { return t.playerHand }

// GetDealerHand ディーラーホールカード取得
func (t *TexasHoldemBonus) GetDealerHand() []*Card { return t.dealerHand }

// GetCommunity コミュニティカード取得
func (t *TexasHoldemBonus) GetCommunity() []*Card { return t.community }

// GetPhase 現在のフェーズ
func (t *TexasHoldemBonus) GetPhase() int { return t.phase }

// GetGameEndFlag ゲーム終了フラグ
func (t *TexasHoldemBonus) GetGameEndFlag() bool { return t.gameEndFlag }

// GetAnteBet アンテベット額
func (t *TexasHoldemBonus) GetAnteBet() int { return t.anteBet }

// GetBonusBet ボーナスサイドベット額
func (t *TexasHoldemBonus) GetBonusBet() int { return t.bonusBet }

// GetFlopBet フロップベット額（2×アンテ）
func (t *TexasHoldemBonus) GetFlopBet() int { return t.flopBet }

// GetTurnBet ターンベット額（1×アンテ）
func (t *TexasHoldemBonus) GetTurnBet() int { return t.turnBet }

// GetRiverBet リバーベット額（1×アンテ）
func (t *TexasHoldemBonus) GetRiverBet() int { return t.riverBet }

// GetTotalPlayBet フロップ＋ターン＋リバーの合計ベット額
func (t *TexasHoldemBonus) GetTotalPlayBet() int { return t.totalPlayBet() }

// GetResult ゲーム結果
func (t *TexasHoldemBonus) GetResult() GameResult { return t.result }

// GetAntePayout アンテ配当（アンテボーナス込み）
func (t *TexasHoldemBonus) GetAntePayout() int { return t.antePayout }

// GetPlayPayout プレイベット配当合計
func (t *TexasHoldemBonus) GetPlayPayout() int { return t.playPayout }

// GetBonusPayout ボーナスサイドベット配当
func (t *TexasHoldemBonus) GetBonusPayout() int { return t.bonusPayout }

// GetTotalPayout 合計配当
func (t *TexasHoldemBonus) GetTotalPayout() int {
	return t.antePayout + t.playPayout + t.bonusPayout
}

// GetPlayerHandRank プレイヤーハンドランク
func (t *TexasHoldemBonus) GetPlayerHandRank() int { return t.playerHandRank }

// GetDealerHandRank ディーラーハンドランク
func (t *TexasHoldemBonus) GetDealerHandRank() int { return t.dealerHandRank }

// GetPlayerBest プレイヤーの最良5枚
func (t *TexasHoldemBonus) GetPlayerBest() []*Card { return t.playerBest }

// GetDealerBest ディーラーの最良5枚
func (t *TexasHoldemBonus) GetDealerBest() []*Card { return t.dealerBest }

// GetChips チップ
func (t *TexasHoldemBonus) GetChips() int { return t.chips.GetChips() }

// GetActionLog 棋譜を取得する
func (t *TexasHoldemBonus) GetActionLog() []*ActionLogEntry { return t.actionLog }

// --- Test helpers ---

// SetPhase フェーズ設定（テスト用）
func (t *TexasHoldemBonus) SetPhase(phase int) { t.phase = phase }

// SetPlayerHand プレイヤーホールカード設定（テスト用）
func (t *TexasHoldemBonus) SetPlayerHand(cards []*Card) { t.playerHand = cards }

// SetDealerHand ディーラーホールカード設定（テスト用）
func (t *TexasHoldemBonus) SetDealerHand(cards []*Card) { t.dealerHand = cards }

// SetCommunity コミュニティカード設定（テスト用）
func (t *TexasHoldemBonus) SetCommunity(cards []*Card) { t.community = cards }

// SetAnteBet アンテベット額設定（テスト用）
func (t *TexasHoldemBonus) SetAnteBet(amount int) { t.anteBet = amount }

// SetBonusBet ボーナスベット額設定（テスト用）
func (t *TexasHoldemBonus) SetBonusBet(amount int) { t.bonusBet = amount }

// SetFlopBet フロップベット額設定（テスト用）
func (t *TexasHoldemBonus) SetFlopBet(amount int) { t.flopBet = amount }

// SetTurnBet ターンベット額設定（テスト用）
func (t *TexasHoldemBonus) SetTurnBet(amount int) { t.turnBet = amount }

// SetRiverBet リバーベット額設定（テスト用）
func (t *TexasHoldemBonus) SetRiverBet(amount int) { t.riverBet = amount }

// SetChips チップ設定（テスト用）
func (t *TexasHoldemBonus) SetChips(chips int) { t.chips.SetChips(chips) }

// texasHoldemBonusJSON は TexasHoldemBonus の JSON ワイヤーフォーマット
type texasHoldemBonusJSON struct {
	TrumpCards     *TrumpCards       `json:"tc"`
	PlayerHand     []*Card           `json:"ph"`
	DealerHand     []*Card           `json:"dh"`
	Community      []*Card           `json:"cm"`
	Chips          *ChipHolder       `json:"ch"`
	AnteBet        int               `json:"ab"`
	BonusBet       int               `json:"bb"`
	FlopBet        int               `json:"fb"`
	TurnBet        int               `json:"tb"`
	RiverBet       int               `json:"rb"`
	Phase          int               `json:"ps"`
	GameEndFlag    bool              `json:"ge"`
	Result         GameResult        `json:"rs"`
	AntePayout     int               `json:"ap"`
	PlayPayout     int               `json:"pp"`
	BonusPayout    int               `json:"bp"`
	PlayerHandRank int               `json:"pr"`
	DealerHandRank int               `json:"dr"`
	PlayerBest     []*Card           `json:"pb"`
	DealerBest     []*Card           `json:"db"`
	ActionLog      []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (t *TexasHoldemBonus) MarshalJSON() ([]byte, error) {
	return json.Marshal(texasHoldemBonusJSON{
		TrumpCards:     t.trumpCards,
		PlayerHand:     t.playerHand,
		DealerHand:     t.dealerHand,
		Community:      t.community,
		Chips:          &t.chips,
		AnteBet:        t.anteBet,
		BonusBet:       t.bonusBet,
		FlopBet:        t.flopBet,
		TurnBet:        t.turnBet,
		RiverBet:       t.riverBet,
		Phase:          t.phase,
		GameEndFlag:    t.gameEndFlag,
		Result:         t.result,
		AntePayout:     t.antePayout,
		PlayPayout:     t.playPayout,
		BonusPayout:    t.bonusPayout,
		PlayerHandRank: t.playerHandRank,
		DealerHandRank: t.dealerHandRank,
		PlayerBest:     t.playerBest,
		DealerBest:     t.dealerBest,
		ActionLog:      t.actionLog,
	})
}

// texasHoldemBonusMaxSliceLen caps slice sizes during deserialisation.
const texasHoldemBonusMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (t *TexasHoldemBonus) UnmarshalJSON(data []byte) error {
	var j texasHoldemBonusJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.PlayerHand) > texasHoldemBonusMaxSliceLen ||
		len(j.DealerHand) > texasHoldemBonusMaxSliceLen ||
		len(j.Community) > texasHoldemBonusMaxSliceLen ||
		len(j.PlayerBest) > texasHoldemBonusMaxSliceLen ||
		len(j.DealerBest) > texasHoldemBonusMaxSliceLen ||
		len(j.ActionLog) > texasHoldemBonusMaxSliceLen {
		return fmt.Errorf("texasholdembonus: input array exceeds maximum allowed size")
	}

	t.trumpCards = j.TrumpCards
	if t.trumpCards == nil {
		t.trumpCards = NewTrumpCards(0)
	}
	t.playerHand = sliceOrEmpty(j.PlayerHand)
	t.dealerHand = sliceOrEmpty(j.DealerHand)
	t.community = sliceOrEmpty(j.Community)
	if j.Chips != nil {
		t.chips = *j.Chips
	}
	t.anteBet = j.AnteBet
	t.bonusBet = j.BonusBet
	t.flopBet = j.FlopBet
	t.turnBet = j.TurnBet
	t.riverBet = j.RiverBet
	t.phase = j.Phase
	t.gameEndFlag = j.GameEndFlag
	t.result = j.Result
	t.antePayout = j.AntePayout
	t.playPayout = j.PlayPayout
	t.bonusPayout = j.BonusPayout
	t.playerHandRank = j.PlayerHandRank
	t.dealerHandRank = j.DealerHandRank
	t.playerBest = sliceOrEmpty(j.PlayerBest)
	t.dealerBest = sliceOrEmpty(j.DealerBest)
	t.actionLog = j.ActionLog
	if t.actionLog == nil {
		t.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
