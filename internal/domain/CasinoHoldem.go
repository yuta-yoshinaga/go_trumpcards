//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// カジノホールデムフェーズ定数
const (
	CasinoHoldemPhaseBet  = 1 // ベットフェーズ（アンテと AA ボーナス）
	CasinoHoldemPhaseFlop = 2 // フロップ後（コール 2×アンテ or フォールド）
	CasinoHoldemPhaseEnd  = 3 // 終了フェーズ（リバーまで公開・配当確定）
)

// カジノホールデムデフォルト値
const (
	CasinoHoldemDefaultChips = 1000  // デフォルトチップ
	CasinoHoldemMinBet       = 10    // 最低ベット額
	CasinoHoldemMaxBet       = 10000 // 最大ベット額
	CasinoHoldemHoleCards    = 2     // ホールカード枚数
	CasinoHoldemBoardCards   = 5     // コミュニティカード枚数
)

// CasinoHoldemDealerQualifyMinPair はディーラーがクオリファイするのに
// 必要な最小ペア値（4のペア以上）。Ace は 14 として扱う。
const CasinoHoldemDealerQualifyMinPair = 4

// アンテ配当倍率（プレイヤー勝利時に支払う、ハンドランクに応じて変動）
const (
	CasinoHoldemAntePayRoyalFlush    = 100 // ロイヤルフラッシュ 100:1
	CasinoHoldemAntePayStraightFlush = 20  // ストレートフラッシュ 20:1
	CasinoHoldemAntePayFourOfAKind   = 10  // フォーカード 10:1
	CasinoHoldemAntePayFullHouse     = 3   // フルハウス 3:1
	CasinoHoldemAntePayFlush         = 2   // フラッシュ 2:1
	CasinoHoldemAntePayOther         = 1   // ストレート以下 1:1
)

// AA ボーナスサイドベット配当倍率（プレイヤーの 2 枚 + フロップ 3 枚）
// Pair of Aces 以上で支払われる。Pair of Aces 未満（Pair of Kings 以下や
// High Card）はベットを失う。
const (
	CasinoHoldemBonusPayRoyalFlush    = 100 // ロイヤルフラッシュ 100:1
	CasinoHoldemBonusPayStraightFlush = 50  // ストレートフラッシュ 50:1
	CasinoHoldemBonusPayFourOfAKind   = 40  // フォーカード 40:1
	CasinoHoldemBonusPayFullHouse     = 30  // フルハウス 30:1
	CasinoHoldemBonusPayFlush         = 20  // フラッシュ 20:1
	CasinoHoldemBonusPayStraight      = 10  // ストレート 10:1
	CasinoHoldemBonusPayThreeOfAKind  = 8   // スリーカード 8:1
	CasinoHoldemBonusPayTwoPair       = 7   // ツーペア 7:1
	CasinoHoldemBonusPayPairOfAces    = 7   // エースペア 7:1
)

// CasinoHoldem カジノホールデムクラス
type CasinoHoldem struct {
	trumpCards     *TrumpCards // トランプカード
	playerHand     []*Card     // プレイヤーホールカード
	dealerHand     []*Card     // ディーラーホールカード
	community      []*Card     // コミュニティカード
	chips          ChipHolder  // チップ
	anteBet        int         // アンテベット額
	bonusBet       int         // AA ボーナスベット額
	callBet        int         // コールベット額（2×アンテ）
	phase          int         // 現在のフェーズ
	gameEndFlag    bool        // ゲーム終了フラグ
	result         GameResult  // ゲーム結果
	dealerQualify  bool        // ディーラークオリファイフラグ
	antePayout     int         // アンテ配当
	callPayout     int         // コール配当
	bonusPayout    int         // AA ボーナス配当
	playerHandRank int         // プレイヤー最良5枚ランク
	dealerHandRank int         // ディーラー最良5枚ランク
	playerBest     []*Card     // プレイヤー最良5枚
	dealerBest     []*Card     // ディーラー最良5枚
	actionLogBase
}

// NewCasinoHoldem コンストラクタ
func NewCasinoHoldem(trumpCards *TrumpCards) *CasinoHoldem {
	trumpCards.Shuffle()
	return &CasinoHoldem{
		trumpCards: trumpCards,
		phase:      CasinoHoldemPhaseBet,
	}
}

// NewDefaultCasinoHoldem デフォルト設定でゲームを生成するファクトリ関数
func NewDefaultCasinoHoldem() *CasinoHoldem {
	c := NewCasinoHoldem(NewTrumpCards(0))
	c.chips.SetChips(CasinoHoldemDefaultChips)
	return c
}

// Reset ゲーム初期化
func (c *CasinoHoldem) Reset() {
	c.gameEndFlag = false
	c.phase = CasinoHoldemPhaseBet
	c.playerHand = nil
	c.dealerHand = nil
	c.community = nil
	c.anteBet = 0
	c.bonusBet = 0
	c.callBet = 0
	c.result = 0
	c.dealerQualify = false
	c.antePayout = 0
	c.callPayout = 0
	c.bonusPayout = 0
	c.playerHandRank = 0
	c.dealerHandRank = 0
	c.playerBest = nil
	c.dealerBest = nil
	c.actionLog = nil
	if c.chips.GetChips() < CasinoHoldemMinBet {
		c.chips.SetChips(CasinoHoldemDefaultChips)
	}
	c.trumpCards = NewTrumpCards(0)
	for range 10 {
		c.trumpCards.Shuffle()
	}
}

// Bet アンテベット＋オプションの AA ボーナス。ホールカード（各2枚）と
// フロップ（3枚）を配り、フロップフェーズへ遷移する。
func (c *CasinoHoldem) Bet(ante, bonus int) error {
	if c.phase != CasinoHoldemPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if ante < CasinoHoldemMinBet || ante%CasinoHoldemMinBet != 0 || ante > CasinoHoldemMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid ante amount.")
	}
	if bonus < 0 {
		return NewDomainError(ErrInvalidAmount, "Bonus bet must not be negative.")
	}
	if bonus > 0 && (bonus < CasinoHoldemMinBet || bonus%CasinoHoldemMinBet != 0 || bonus > CasinoHoldemMaxBet) {
		return NewDomainError(ErrInvalidAmount, "Invalid bonus bet amount.")
	}
	totalCost := ante + bonus
	if !c.chips.SubtractChips(totalCost) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	c.anteBet = ante
	c.bonusBet = bonus
	c.appendLog(0, "bet", fmt.Sprintf("ante=%d bonus=%d", ante, bonus), nil)

	c.dealHole()
	c.dealFlop()
	c.phase = CasinoHoldemPhaseFlop
	return nil
}

// Call フロップ後にコールベット（2×アンテ）を置きターンとリバーを公開して勝負を確定する。
func (c *CasinoHoldem) Call() error {
	if c.phase != CasinoHoldemPhaseFlop {
		return NewDomainError(ErrWrongPhase, "Call is only allowed during the flop phase.")
	}
	bet := c.anteBet * 2
	if !c.chips.SubtractChips(bet) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for call bet.")
	}
	c.callBet = bet
	c.appendLog(0, "call", fmt.Sprintf("call bet=%d", bet), nil)

	c.dealTurnAndRiver()
	c.resolve()
	return nil
}

// Fold フロップ後にフォールド。アンテとコールベットは没収、
// AA ボーナスは現在の 5 枚（ホール+フロップ）で評価する。
func (c *CasinoHoldem) Fold() error {
	if c.phase != CasinoHoldemPhaseFlop {
		return NewDomainError(ErrWrongPhase, "Fold is only allowed during the flop phase.")
	}
	c.appendLog(0, "fold", "player folds", nil)

	c.result = GameResultLose
	c.evaluateBonus()
	if c.bonusPayout > 0 {
		c.chips.AddChips(c.bonusPayout)
	}
	c.gameEndFlag = true
	c.phase = CasinoHoldemPhaseEnd
	c.appendLog(-1, "result", "player folded", nil)
	return nil
}

// dealHole 各2枚のホールカードを配る。
func (c *CasinoHoldem) dealHole() {
	c.playerHand = make([]*Card, 0, CasinoHoldemHoleCards)
	c.dealerHand = make([]*Card, 0, CasinoHoldemHoleCards)
	for range CasinoHoldemHoleCards {
		c.playerHand = append(c.playerHand, c.trumpCards.DrawCard())
		c.dealerHand = append(c.dealerHand, c.trumpCards.DrawCard())
	}
	c.appendLog(-1, "deal", "dealt 2 hole cards each", nil)
}

// dealFlop 3枚のフロップを配る。
func (c *CasinoHoldem) dealFlop() {
	c.community = make([]*Card, 0, CasinoHoldemBoardCards)
	for range 3 {
		c.community = append(c.community, c.trumpCards.DrawCard())
	}
	c.updatePlayerCurrentRank()
	c.appendLog(-1, "flop", "flop dealt", nil)
}

// dealTurnAndRiver 残り 2 枚（ターン＋リバー）を一気に公開する。
func (c *CasinoHoldem) dealTurnAndRiver() {
	c.community = append(c.community, c.trumpCards.DrawCard())
	c.community = append(c.community, c.trumpCards.DrawCard())
	c.appendLog(-1, "turn_river", "turn and river dealt", nil)
}

// updatePlayerCurrentRank プレイヤーの現在の最良ハンドランクを更新する。
// フロップ時点でホール+フロップの 5 枚から評価する。
// ヒント生成（フロントエンド）が現在の手の強さを参照できるようにするため。
// resolve() がリバー後に同じフィールドを上書きする。
func (c *CasinoHoldem) updatePlayerCurrentRank() {
	all := append([]*Card{}, c.playerHand...)
	all = append(all, c.community...)
	c.playerHandRank, _ = evalBestFromSeven(all)
}

// resolve ショーダウン処理（リバー後）。ディーラークオリファイ判定と配当計算を行う。
func (c *CasinoHoldem) resolve() {
	playerAll := append([]*Card{}, c.playerHand...)
	playerAll = append(playerAll, c.community...)
	dealerAll := append([]*Card{}, c.dealerHand...)
	dealerAll = append(dealerAll, c.community...)

	c.playerHandRank, c.playerBest = evalBestFromSeven(playerAll)
	c.dealerHandRank, c.dealerBest = evalBestFromSeven(dealerAll)
	c.dealerQualify = casinoHoldemDealerQualifies(c.dealerHandRank, c.dealerBest)

	cmp := c.compareBest()
	switch {
	case cmp > 0:
		c.result = GameResultWin
	case cmp < 0:
		c.result = GameResultLose
	default:
		c.result = GameResultDraw
	}

	c.calculatePayouts()
	c.evaluateBonus()

	totalPayout := c.antePayout + c.callPayout + c.bonusPayout
	if totalPayout > 0 {
		c.chips.AddChips(totalPayout)
	}

	c.gameEndFlag = true
	c.phase = CasinoHoldemPhaseEnd

	var resultStr string
	switch c.result {
	case GameResultWin:
		resultStr = "player wins"
	case GameResultDraw:
		resultStr = "push"
	default:
		resultStr = "dealer wins"
	}
	c.appendLog(-1, "result", resultStr, nil)
}

// compareBest プレイヤーとディーラーの最良5枚を比較する
func (c *CasinoHoldem) compareBest() int {
	if c.playerHandRank > c.dealerHandRank {
		return 1
	}
	if c.playerHandRank < c.dealerHandRank {
		return -1
	}
	return compareHighCardsSlice(c.playerBest, c.dealerBest)
}

// calculatePayouts 勝敗・ディーラークオリファイに基づくアンテ／コール配当
//
// ルール：
//   - ディーラーがクオリファイしない場合：アンテはハンドランクに応じた配当を支払う
//     （勝敗無関係）。コールベットはプッシュ（元金返却のみ）。
//   - ディーラーがクオリファイしてプレイヤー勝利：アンテはハンドランクに応じた配当、
//     コールベットは 1:1。
//   - ディーラーがクオリファイして引き分け：アンテとコールベットはプッシュ。
//   - ディーラーがクオリファイしてプレイヤー敗北：アンテとコールベットを失う。
//
// ペイアウト額は「元金 + 配当」を表す。負けた場合のみ 0。
func (c *CasinoHoldem) calculatePayouts() {
	mult := c.anteMultiplier()
	if !c.dealerQualify {
		// アンテはハンドランクに応じて支払い、コールはプッシュ
		c.antePayout = c.anteBet + c.anteBet*mult
		c.callPayout = c.callBet
		return
	}
	switch c.result {
	case GameResultWin:
		c.antePayout = c.anteBet + c.anteBet*mult
		c.callPayout = c.callBet * 2
	case GameResultDraw:
		c.antePayout = c.anteBet
		c.callPayout = c.callBet
	case GameResultLose:
		c.antePayout = 0
		c.callPayout = 0
	}
}

// anteMultiplier プレイヤーの最良 5 枚に対するアンテ配当倍率を返す。
// ストレート以下は 1:1（[CasinoHoldemAntePayOther]）。
func (c *CasinoHoldem) anteMultiplier() int {
	switch c.playerHandRank {
	case PokerHandRoyalFlush:
		return CasinoHoldemAntePayRoyalFlush
	case PokerHandStraightFlush:
		return CasinoHoldemAntePayStraightFlush
	case PokerHandFourOfAKind:
		return CasinoHoldemAntePayFourOfAKind
	case PokerHandFullHouse:
		return CasinoHoldemAntePayFullHouse
	case PokerHandFlush:
		return CasinoHoldemAntePayFlush
	default:
		return CasinoHoldemAntePayOther
	}
}

// evaluateBonus AA ボーナスサイドベット評価
// プレイヤーの 2 枚のホールカード + フロップ 3 枚 = 5 枚で評価する。
// Pair of Aces 以上で支払われる。それ以外（含む弱いペア）は元金没収。
func (c *CasinoHoldem) evaluateBonus() {
	if c.bonusBet <= 0 || len(c.playerHand) < 2 || len(c.community) < 3 {
		return
	}
	five := append([]*Card{}, c.playerHand...)
	five = append(five, c.community[:3]...)
	rank, best := evalBestFromSeven(five)
	mult := casinoHoldemBonusMultiplier(rank, best)
	if mult > 0 {
		c.bonusPayout = c.bonusBet + c.bonusBet*mult
	}
}

// casinoHoldemBonusMultiplier は AA ボーナスのハンドランクに対する倍率を返す。
// One Pair の場合は Ace ペアでなければ 0。
func casinoHoldemBonusMultiplier(rank int, best []*Card) int {
	switch rank {
	case PokerHandRoyalFlush:
		return CasinoHoldemBonusPayRoyalFlush
	case PokerHandStraightFlush:
		return CasinoHoldemBonusPayStraightFlush
	case PokerHandFourOfAKind:
		return CasinoHoldemBonusPayFourOfAKind
	case PokerHandFullHouse:
		return CasinoHoldemBonusPayFullHouse
	case PokerHandFlush:
		return CasinoHoldemBonusPayFlush
	case PokerHandStraight:
		return CasinoHoldemBonusPayStraight
	case PokerHandThreeOfAKind:
		return CasinoHoldemBonusPayThreeOfAKind
	case PokerHandTwoPair:
		return CasinoHoldemBonusPayTwoPair
	case PokerHandOnePair:
		if hasAcePair(best) {
			return CasinoHoldemBonusPayPairOfAces
		}
		return 0
	default:
		return 0
	}
}

// hasAcePair は best 5 枚の中に A のペア（少なくとも 2 枚の A）が
// 含まれているかを返す。Card の値は 1=A。
func hasAcePair(best []*Card) bool {
	count := 0
	for _, card := range best {
		if card != nil && card.GetValue() == 1 {
			count++
		}
	}
	return count >= 2
}

// casinoHoldemDealerQualifies はディーラーが Pair of Fours 以上で
// クオリファイするかを判定する。
//
// ルール：
//   - ハンドランクが Two Pair 以上なら自動的にクオリファイ。
//   - One Pair の場合は、ペアの値が 4 以上ならクオリファイ。Ace は 14 として扱う。
//   - High Card の場合はクオリファイしない。
func casinoHoldemDealerQualifies(rank int, best []*Card) bool {
	if rank > PokerHandOnePair {
		return true
	}
	if rank < PokerHandOnePair {
		return false
	}
	// ペアの値を特定する
	counts := make(map[int]int, 5)
	for _, card := range best {
		if card == nil {
			continue
		}
		counts[card.GetValue()]++
	}
	for value, n := range counts {
		if n < 2 {
			continue
		}
		// Ace は常にクオリファイ（14 扱い）
		if value == 1 {
			return true
		}
		if value >= CasinoHoldemDealerQualifyMinPair {
			return true
		}
	}
	return false
}

// --- Getters ---

// GetPlayerHand プレイヤーホールカード取得
func (c *CasinoHoldem) GetPlayerHand() []*Card { return c.playerHand }

// GetDealerHand ディーラーホールカード取得
func (c *CasinoHoldem) GetDealerHand() []*Card { return c.dealerHand }

// GetCommunity コミュニティカード取得
func (c *CasinoHoldem) GetCommunity() []*Card { return c.community }

// GetPhase 現在のフェーズ
func (c *CasinoHoldem) GetPhase() int { return c.phase }

// GetGameEndFlag ゲーム終了フラグ
func (c *CasinoHoldem) GetGameEndFlag() bool { return c.gameEndFlag }

// GetAnteBet アンテベット額
func (c *CasinoHoldem) GetAnteBet() int { return c.anteBet }

// GetBonusBet AA ボーナスベット額
func (c *CasinoHoldem) GetBonusBet() int { return c.bonusBet }

// GetCallBet コールベット額（2×アンテ）
func (c *CasinoHoldem) GetCallBet() int { return c.callBet }

// RecommendCall はフロップ後にコール (2×アンテ) を推奨するかを返す。基本戦略として
// ワンペア以上、または A/K を含むときコール、それ以外はフォールドを推奨する。
// (プレイヤーハンドランクはショーダウンまで設定されないため、その場で評価する。)
func (c *CasinoHoldem) RecommendCall() bool {
	cards := make([]*Card, 0, len(c.playerHand)+len(c.community))
	cards = append(cards, c.playerHand...)
	cards = append(cards, c.community...)
	if len(cards) < 5 {
		return true
	}
	if evalFiveCardHand(cards[:5]) >= PokerHandOnePair {
		return true
	}
	for _, cc := range cards {
		if v := cc.GetValue(); v == 1 || v == 13 { // Ace or King
			return true
		}
	}
	return false
}

// GetResult ゲーム結果
func (c *CasinoHoldem) GetResult() GameResult { return c.result }

// GetDealerQualify ディーラークオリファイフラグ
func (c *CasinoHoldem) GetDealerQualify() bool { return c.dealerQualify }

// GetAntePayout アンテ配当
func (c *CasinoHoldem) GetAntePayout() int { return c.antePayout }

// GetCallPayout コール配当
func (c *CasinoHoldem) GetCallPayout() int { return c.callPayout }

// GetBonusPayout AA ボーナスサイドベット配当
func (c *CasinoHoldem) GetBonusPayout() int { return c.bonusPayout }

// GetTotalPayout 合計配当
func (c *CasinoHoldem) GetTotalPayout() int {
	return c.antePayout + c.callPayout + c.bonusPayout
}

// GetPlayerHandRank プレイヤーハンドランク
func (c *CasinoHoldem) GetPlayerHandRank() int { return c.playerHandRank }

// GetDealerHandRank ディーラーハンドランク
func (c *CasinoHoldem) GetDealerHandRank() int { return c.dealerHandRank }

// GetPlayerBest プレイヤーの最良5枚
func (c *CasinoHoldem) GetPlayerBest() []*Card { return c.playerBest }

// GetDealerBest ディーラーの最良5枚
func (c *CasinoHoldem) GetDealerBest() []*Card { return c.dealerBest }

// GetChips チップ
func (c *CasinoHoldem) GetChips() int { return c.chips.GetChips() }

// --- Test helpers ---

// SetPhase フェーズ設定（テスト用）
func (c *CasinoHoldem) SetPhase(phase int) { c.phase = phase }

// SetPlayerHand プレイヤーホールカード設定（テスト用）
func (c *CasinoHoldem) SetPlayerHand(cards []*Card) { c.playerHand = cards }

// SetDealerHand ディーラーホールカード設定（テスト用）
func (c *CasinoHoldem) SetDealerHand(cards []*Card) { c.dealerHand = cards }

// SetCommunity コミュニティカード設定（テスト用）
func (c *CasinoHoldem) SetCommunity(cards []*Card) { c.community = cards }

// SetAnteBet アンテベット額設定（テスト用）
func (c *CasinoHoldem) SetAnteBet(amount int) { c.anteBet = amount }

// SetBonusBet AA ボーナスベット額設定（テスト用）
func (c *CasinoHoldem) SetBonusBet(amount int) { c.bonusBet = amount }

// SetCallBet コールベット額設定（テスト用）
func (c *CasinoHoldem) SetCallBet(amount int) { c.callBet = amount }

// SetChips チップ設定（テスト用）
func (c *CasinoHoldem) SetChips(chips int) { c.chips.SetChips(chips) }

// casinoHoldemJSON は CasinoHoldem の JSON ワイヤーフォーマット
type casinoHoldemJSON struct {
	TrumpCards     *TrumpCards       `json:"tc"`
	PlayerHand     []*Card           `json:"ph"`
	DealerHand     []*Card           `json:"dh"`
	Community      []*Card           `json:"cm"`
	Chips          *ChipHolder       `json:"ch"`
	AnteBet        int               `json:"ab"`
	BonusBet       int               `json:"bb"`
	CallBet        int               `json:"cb"`
	Phase          int               `json:"ps"`
	GameEndFlag    bool              `json:"ge"`
	Result         GameResult        `json:"rs"`
	DealerQualify  bool              `json:"dq"`
	AntePayout     int               `json:"ap"`
	CallPayout     int               `json:"cp"`
	BonusPayout    int               `json:"bp"`
	PlayerHandRank int               `json:"pr"`
	DealerHandRank int               `json:"dr"`
	PlayerBest     []*Card           `json:"pb"`
	DealerBest     []*Card           `json:"db"`
	ActionLog      []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (c *CasinoHoldem) MarshalJSON() ([]byte, error) {
	return json.Marshal(casinoHoldemJSON{
		TrumpCards:     c.trumpCards,
		PlayerHand:     c.playerHand,
		DealerHand:     c.dealerHand,
		Community:      c.community,
		Chips:          &c.chips,
		AnteBet:        c.anteBet,
		BonusBet:       c.bonusBet,
		CallBet:        c.callBet,
		Phase:          c.phase,
		GameEndFlag:    c.gameEndFlag,
		Result:         c.result,
		DealerQualify:  c.dealerQualify,
		AntePayout:     c.antePayout,
		CallPayout:     c.callPayout,
		BonusPayout:    c.bonusPayout,
		PlayerHandRank: c.playerHandRank,
		DealerHandRank: c.dealerHandRank,
		PlayerBest:     c.playerBest,
		DealerBest:     c.dealerBest,
		ActionLog:      c.actionLog,
	})
}

// casinoHoldemMaxSliceLen caps slice sizes during deserialisation.
const casinoHoldemMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (c *CasinoHoldem) UnmarshalJSON(data []byte) error {
	var j casinoHoldemJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.PlayerHand) > casinoHoldemMaxSliceLen ||
		len(j.DealerHand) > casinoHoldemMaxSliceLen ||
		len(j.Community) > casinoHoldemMaxSliceLen ||
		len(j.PlayerBest) > casinoHoldemMaxSliceLen ||
		len(j.DealerBest) > casinoHoldemMaxSliceLen ||
		len(j.ActionLog) > casinoHoldemMaxSliceLen {
		return fmt.Errorf("casinoholdem: input array exceeds maximum allowed size")
	}

	c.trumpCards = j.TrumpCards
	if c.trumpCards == nil {
		c.trumpCards = NewTrumpCards(0)
	}
	c.playerHand = sliceOrEmpty(j.PlayerHand)
	c.dealerHand = sliceOrEmpty(j.DealerHand)
	c.community = sliceOrEmpty(j.Community)
	if j.Chips != nil {
		c.chips = *j.Chips
	}
	c.anteBet = j.AnteBet
	c.bonusBet = j.BonusBet
	c.callBet = j.CallBet
	c.phase = j.Phase
	c.gameEndFlag = j.GameEndFlag
	c.result = j.Result
	c.dealerQualify = j.DealerQualify
	c.antePayout = j.AntePayout
	c.callPayout = j.CallPayout
	c.bonusPayout = j.BonusPayout
	c.playerHandRank = j.PlayerHandRank
	c.dealerHandRank = j.DealerHandRank
	c.playerBest = sliceOrEmpty(j.PlayerBest)
	c.dealerBest = sliceOrEmpty(j.DealerBest)
	c.actionLog = j.ActionLog
	if c.actionLog == nil {
		c.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
