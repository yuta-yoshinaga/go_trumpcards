//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// doubleAttackMaxSliceLen はデシリアライズ時のスライス長上限。
const doubleAttackMaxSliceLen = 1000

// DoubleAttackMaxHands はスプリットで増やせる手札の上限。
const DoubleAttackMaxHands = 4

// エラー値。
var (
	errDoubleAttackWrongPhase   = errors.New("doubleattack: action not allowed in this phase")
	errDoubleAttackAnteRange    = errors.New("doubleattack: ante is out of range")
	errDoubleAttackSideRange    = errors.New("doubleattack: bust it bet is out of range")
	errDoubleAttackAttackRange  = errors.New("doubleattack: the double attack bet may not exceed the ante")
	errDoubleAttackNotEnough    = errors.New("doubleattack: not enough chips for that wager")
	errDoubleAttackCannotSplit  = errors.New("doubleattack: this hand cannot be split")
	errDoubleAttackCannotDouble = errors.New("doubleattack: this hand cannot be doubled")
	errDoubleAttackFinished     = errors.New("doubleattack: the player is out of chips")
)

// DoubleAttackResult は 1 つの手札の決着。
type DoubleAttackResult int

// 決着の種類。
const (
	// DoubleAttackResultNone まだ決着していない
	DoubleAttackResultNone DoubleAttackResult = iota
	// DoubleAttackResultWin プレイヤーの勝ち
	DoubleAttackResultWin
	// DoubleAttackResultLose ディーラーの勝ち
	DoubleAttackResultLose
	// DoubleAttackResultPush 引き分け
	DoubleAttackResultPush
	// DoubleAttackResultBlackjack プレイヤーのブラックジャック (1:1)
	DoubleAttackResultBlackjack
)

// DoubleAttackResultMax は最大の決着値 (復元時の範囲検査に使う)。
const DoubleAttackResultMax = DoubleAttackResultBlackjack

// DoubleAttackResultName は決着の識別子を返す (i18n キーの一部に使う)。
func DoubleAttackResultName(r DoubleAttackResult) string {
	switch r {
	case DoubleAttackResultWin:
		return "win"
	case DoubleAttackResultLose:
		return "lose"
	case DoubleAttackResultPush:
		return "push"
	case DoubleAttackResultBlackjack:
		return "blackjack"
	default:
		return "none"
	}
}

// DoubleAttackHint は人間への助言。
type DoubleAttackHint struct {
	// Action は薦める操作 ("attack" / "hit" / "stand" / "double" / "split")。
	Action string
	// Reason は理由の識別子 (i18n キーの一部)。
	Reason string
}

// DoubleAttackBlackjack は追加ベット・ブラックジャックの卓。
//
// **アップカードを見てから賭け増しできる**のがこのゲームの本体で、名前もそこから来て
// いる。その有利さの対価として**プレイヤーのブラックジャックは 1:1** に抑えられ、
// デッキからは 10 が抜かれている (48 枚 × 8)。
type DoubleAttackBlackjack struct {
	shoe   *TrumpCards
	player *DoubleAttackBlackjackPlayer
	config DoubleAttackBlackjackConfig

	phase      DoubleAttackPhase
	hands      []*BlackJackHand
	activeHand int
	dealerHand *BlackJackHand
	// dealerHoleDealt は 2 枚目を配ったか。**追加ベットの後にしか配らない。**
	dealerHoleDealt bool
	anteBet         int
	attackBet       int
	bustItBet       int
	results         []DoubleAttackResult
	payout          int
	bustItPayout    int
	roundNumber     int
	gameEndFlag     bool
	actionLog       []*ActionLogEntry
	turnNumber      int
}

// NewDoubleAttackBlackjack は指定のシュー・プレイヤー・設定で卓を構築する。
func NewDoubleAttackBlackjack(shoe *TrumpCards, player *DoubleAttackBlackjackPlayer,
	config DoubleAttackBlackjackConfig,
) *DoubleAttackBlackjack {
	return &DoubleAttackBlackjack{
		shoe:        shoe,
		player:      player,
		config:      config,
		roundNumber: 1,
	}
}

// NewDefaultDoubleAttackBlackjack は既定の卓を構築する。
func NewDefaultDoubleAttackBlackjack() *DoubleAttackBlackjack {
	cfg := DefaultDoubleAttackBlackjackConfig()
	return NewDoubleAttackBlackjack(
		NewTrumpCardsSpanish21(DoubleAttackDeckCount),
		NewDoubleAttackBlackjackPlayer(cfg.InitialChips), cfg)
}

// Reset はゲームを初期状態に戻す。
func (g *DoubleAttackBlackjack) Reset() {
	g.shoe = NewTrumpCardsSpanish21(DoubleAttackDeckCount)
	g.shoe.Shuffle()
	g.player.SetChips(g.config.InitialChips)
	g.roundNumber = 1
	g.gameEndFlag = false
	g.actionLog = nil
	g.turnNumber = 0
	g.appendLog("start", "extra bet blackjack begins", nil)
	g.startRound()
}

// startRound は 1 ラウンド分の状態を初期化する。
func (g *DoubleAttackBlackjack) startRound() {
	g.hands = nil
	g.activeHand = 0
	g.dealerHand = nil
	g.dealerHoleDealt = false
	g.anteBet = 0
	g.attackBet = 0
	g.bustItBet = 0
	g.results = nil
	g.payout = 0
	g.bustItPayout = 0
	g.phase = DoubleAttackPhaseBet
	g.ensureShoe()
}

// ensureShoe は 1 ラウンド分に足りない残り枚数ならシューを組み直す。
//
// スプリットとヒットで最大 20 枚ほど使うので、余裕をもって組み直す。
func (g *DoubleAttackBlackjack) ensureShoe() {
	if g.shoe.GetRemainingCount() < 30 {
		g.shoe = NewTrumpCardsSpanish21(DoubleAttackDeckCount)
		g.shoe.Shuffle()
	}
}

// NextRound は次のラウンドを始める。
func (g *DoubleAttackBlackjack) NextRound() error {
	if g.gameEndFlag {
		return errDoubleAttackFinished
	}
	if g.phase != DoubleAttackPhaseResult {
		return errDoubleAttackWrongPhase
	}
	g.roundNumber++
	g.startRound()
	if g.player.GetChips() < DoubleAttackAnteMin {
		g.gameEndFlag = true
		g.appendLog("gameEnd", "out of chips", nil)
	}
	return nil
}

// --- 賭け ---

// PlaceBet はアンティと任意の Bust It を置き、最初の 3 枚を配る。
//
// **ディーラーはアップカード 1 枚だけ**。2 枚目は追加ベットの後に配る ── そこが
// このゲームの情報の順序で、先に配ると賭け増しの意味が消える。
func (g *DoubleAttackBlackjack) PlaceBet(ante, bustIt int) error {
	if g.gameEndFlag {
		return errDoubleAttackFinished
	}
	if g.phase != DoubleAttackPhaseBet {
		return errDoubleAttackWrongPhase
	}
	if ante < DoubleAttackAnteMin || ante > DoubleAttackAnteMax {
		return errDoubleAttackAnteRange
	}
	if bustIt < 0 || bustIt > DoubleAttackBustItMax {
		return errDoubleAttackSideRange
	}
	if !g.player.SubtractChips(ante + bustIt) {
		return errDoubleAttackNotEnough
	}
	g.anteBet = ante
	g.bustItBet = bustIt
	g.appendLog("bet", fmt.Sprintf("ante %d, bustIt %d", ante, bustIt), nil)

	hand := NewBlackJackHand()
	hand.AddCard(g.shoe.DrawCard())
	hand.AddCard(g.shoe.DrawCard())
	hand.SetBet(ante)
	g.hands = []*BlackJackHand{hand}
	g.results = []DoubleAttackResult{DoubleAttackResultNone}

	g.dealerHand = NewBlackJackHand()
	g.dealerHand.AddCard(g.shoe.DrawCard()) // アップカードのみ

	g.appendLog("deal", "cards dealt", hand.GetCards())
	g.phase = DoubleAttackPhaseAttack
	return nil
}

// MaxAttackBet は追加ベットの上限を返す。**アンティと同額まで。**
func (g *DoubleAttackBlackjack) MaxAttackBet() int {
	if g.anteBet == 0 {
		return 0
	}
	if chips := g.player.GetChips(); chips < g.anteBet {
		return chips
	}
	return g.anteBet
}

// Attack はアップカードを見てから追加ベットを置く。0 なら見送り。
func (g *DoubleAttackBlackjack) Attack(amount int) error {
	if g.gameEndFlag {
		return errDoubleAttackFinished
	}
	if g.phase != DoubleAttackPhaseAttack {
		return errDoubleAttackWrongPhase
	}
	if amount < 0 || amount > g.MaxAttackBet() {
		return errDoubleAttackAttackRange
	}
	if amount > 0 && !g.player.SubtractChips(amount) {
		return errDoubleAttackNotEnough
	}
	g.attackBet = amount
	g.hands[0].SetBet(g.anteBet + amount)
	g.appendLog("attack", fmt.Sprintf("double attack %d", amount), nil)

	// **ここで初めてディーラーの 2 枚目を配る。**
	g.dealerHand.AddCard(g.shoe.DrawCard())
	g.dealerHoleDealt = true

	// どちらかがブラックジャックなら即決着。
	if g.hands[0].IsBlackJack() || g.dealerHand.IsBlackJack() {
		g.settle()
		return nil
	}
	g.phase = DoubleAttackPhasePlay
	return nil
}

// --- プレイ ---

// activeBlackJackHand はいま操作している手札を返す。
func (g *DoubleAttackBlackjack) activeBlackJackHand() *BlackJackHand {
	if g.activeHand < 0 || g.activeHand >= len(g.hands) {
		return nil
	}
	return g.hands[g.activeHand]
}

// Hit は 1 枚引く。
func (g *DoubleAttackBlackjack) Hit() error {
	h, err := g.playableHand()
	if err != nil {
		return err
	}
	h.AddCard(g.shoe.DrawCard())
	g.appendLog("hit", "hit", h.GetCards())
	if h.GetScore() > 21 {
		h.SetBusted(true)
		g.advanceHand()
		return nil
	}
	if h.GetScore() == 21 {
		h.SetStood(true)
		g.advanceHand()
	}
	return nil
}

// Stand はその手札を打ち止めにする。
func (g *DoubleAttackBlackjack) Stand() error {
	h, err := g.playableHand()
	if err != nil {
		return err
	}
	h.SetStood(true)
	g.appendLog("stand", "stand", nil)
	g.advanceHand()
	return nil
}

// CanDouble はいまの手札をダブルできるかを返す。**最初の 2 枚のときだけ。**
func (g *DoubleAttackBlackjack) CanDouble() bool {
	h := g.activeBlackJackHand()
	if h == nil || g.phase != DoubleAttackPhasePlay {
		return false
	}
	return h.GetCardsSize() == 2 && !h.IsDoubled() && g.player.GetChips() >= h.GetBet()
}

// Double は賭け金を倍にして 1 枚だけ引き、その手札を終える。
func (g *DoubleAttackBlackjack) Double() error {
	h, err := g.playableHand()
	if err != nil {
		return err
	}
	if !g.CanDouble() {
		return errDoubleAttackCannotDouble
	}
	if !g.player.SubtractChips(h.GetBet()) {
		return errDoubleAttackNotEnough
	}
	h.SetBet(h.GetBet() * 2)
	h.SetDoubled(true)
	h.AddCard(g.shoe.DrawCard())
	g.appendLog("double", fmt.Sprintf("double to %d", h.GetBet()), h.GetCards())
	if h.GetScore() > 21 {
		h.SetBusted(true)
	} else {
		h.SetStood(true)
	}
	g.advanceHand()
	return nil
}

// CanSplit はいまの手札をスプリットできるかを返す。
//
// **同じ数字 2 枚のときだけ。** 10 が抜けているので絵札どうしは J/Q/K の区別なく
// 「10 点」だが、ここは**数字の一致**を条件にする (絵札を混ぜて割るのは別ルール)。
func (g *DoubleAttackBlackjack) CanSplit() bool {
	h := g.activeBlackJackHand()
	if h == nil || g.phase != DoubleAttackPhasePlay {
		return false
	}
	if len(g.hands) >= DoubleAttackMaxHands || h.GetCardsSize() != 2 {
		return false
	}
	if h.GetCard(0).GetValue() != h.GetCard(1).GetValue() {
		return false
	}
	return g.player.GetChips() >= h.GetBet()
}

// Split は同じ数字 2 枚を 2 つの手札に分ける。
func (g *DoubleAttackBlackjack) Split() error {
	h, err := g.playableHand()
	if err != nil {
		return err
	}
	if !g.CanSplit() {
		return errDoubleAttackCannotSplit
	}
	if !g.player.SubtractChips(h.GetBet()) {
		return errDoubleAttackNotEnough
	}
	moved := h.GetCard(1)
	isAces := h.GetCard(0).GetValue() == 1

	rebuilt := NewBlackJackHand()
	rebuilt.AddCard(h.GetCard(0))
	rebuilt.AddCard(g.shoe.DrawCard())
	rebuilt.SetBet(h.GetBet())

	other := NewBlackJackHand()
	other.AddCard(moved)
	other.AddCard(g.shoe.DrawCard())
	other.SetBet(h.GetBet())

	// **エースを割ったら 1 枚ずつで打ち止め。**
	if isAces {
		rebuilt.SetStood(true)
		other.SetStood(true)
	}

	g.hands[g.activeHand] = rebuilt
	g.hands = append(g.hands, nil)
	copy(g.hands[g.activeHand+2:], g.hands[g.activeHand+1:])
	g.hands[g.activeHand+1] = other
	g.results = append(g.results, DoubleAttackResultNone)

	g.appendLog("split", fmt.Sprintf("split into %d hands", len(g.hands)), nil)
	if isAces {
		g.advanceHand()
	}
	return nil
}

// playableHand は操作できる手札を返す。できなければエラー。
func (g *DoubleAttackBlackjack) playableHand() (*BlackJackHand, error) {
	if g.gameEndFlag {
		return nil, errDoubleAttackFinished
	}
	if g.phase != DoubleAttackPhasePlay {
		return nil, errDoubleAttackWrongPhase
	}
	h := g.activeBlackJackHand()
	if h == nil || h.IsStood() || h.IsBusted() {
		return nil, errDoubleAttackWrongPhase
	}
	return h, nil
}

// advanceHand は次の未決着の手札へ進み、無ければディーラーの番にする。
func (g *DoubleAttackBlackjack) advanceHand() {
	for g.activeHand < len(g.hands) {
		h := g.hands[g.activeHand]
		if !h.IsStood() && !h.IsBusted() {
			return
		}
		g.activeHand++
	}
	g.dealerPlay()
}

// --- ディーラーと精算 ---

// dealerPlay はディーラーの手を進める。**ソフト 17 でヒットする。**
func (g *DoubleAttackBlackjack) dealerPlay() {
	// 全部バストしていてもディーラーは引く ── **Bust It が生きているため。**
	for {
		score, soft := calcScore(g.dealerHand.GetCards())
		if score > 21 {
			g.dealerHand.SetBusted(true)
			break
		}
		// **ソフト 17 ではヒットする。** 17 で止まるのはハード 17 のときだけ。
		if score > 17 || (score == 17 && (!soft || !DoubleAttackDealerHitsSoft17)) {
			break
		}
		g.dealerHand.AddCard(g.shoe.DrawCard())
	}
	g.appendLog("dealer", "dealer plays", g.dealerHand.GetCards())
	g.settle()
}

// settle はすべての手札と Bust It を精算する。
func (g *DoubleAttackBlackjack) settle() {
	dealerScore := g.dealerHand.GetScore()
	dealerBusted := dealerScore > 21
	dealerBJ := g.dealerHand.IsBlackJack()

	total := 0
	for i, h := range g.hands {
		r, ret := g.settleHand(h, dealerScore, dealerBusted, dealerBJ)
		g.results[i] = r
		total += ret
	}
	g.bustItPayout = g.settleBustIt(dealerBusted)
	g.payout = total + g.bustItPayout
	g.player.AddChips(g.payout)
	g.appendLog("result", fmt.Sprintf("payout %d", g.payout), nil)
	g.phase = DoubleAttackPhaseResult
}

// settleHand は 1 つの手札を精算し、決着と払い戻し (賭け金の返却を含む) を返す。
func (g *DoubleAttackBlackjack) settleHand(h *BlackJackHand, dealerScore int,
	dealerBusted, dealerBJ bool,
) (DoubleAttackResult, int) {
	bet := h.GetBet()
	switch {
	case h.IsBusted():
		// **バストは引き分けにならない。** ディーラーがバストしても先に負けている。
		return DoubleAttackResultLose, 0
	case h.IsBlackJack() && !dealerBJ:
		// **ブラックジャックは 1:1。** 3:2 ではない。
		return DoubleAttackResultBlackjack,
			bet + bet*DoubleAttackBlackjackPayoutNum/DoubleAttackBlackjackPayoutDen
	case dealerBJ && !h.IsBlackJack():
		return DoubleAttackResultLose, 0
	case dealerBJ && h.IsBlackJack():
		return DoubleAttackResultPush, bet
	case dealerBusted:
		return DoubleAttackResultWin, bet * 2
	case h.GetScore() > dealerScore:
		return DoubleAttackResultWin, bet * 2
	case h.GetScore() < dealerScore:
		return DoubleAttackResultLose, 0
	default:
		return DoubleAttackResultPush, bet
	}
}

// settleBustIt は Bust It を精算する。
//
// **プレイヤーの手とは無関係。** 自分が全部バストしていても、ディーラーがバストすれば
// 払われる。だから全滅していてもディーラーは引き切る。
func (g *DoubleAttackBlackjack) settleBustIt(dealerBusted bool) int {
	if g.bustItBet == 0 || !dealerBusted {
		return 0
	}
	pay := DoubleAttackBustItPayout(g.dealerHand.GetCardsSize())
	return g.bustItBet + g.bustItBet*pay
}

// --- ヒント ---

// GetHint は人間への助言を返す。判断どころでなければ nil。
func (g *DoubleAttackBlackjack) GetHint() *DoubleAttackHint {
	if g.gameEndFlag {
		return nil
	}
	switch g.phase {
	case DoubleAttackPhaseAttack:
		// **アップカードが弱いときだけ乗せる。** 10 が抜けた 48 枚デッキでは
		// ディーラーのバスト率が下がるので、強いアップカードに乗せるのは損。
		up := g.dealerUpValue()
		if up >= 2 && up <= 6 {
			return &DoubleAttackHint{Action: "attack", Reason: "weakUpCard"}
		}
		return &DoubleAttackHint{Action: "stand", Reason: "strongUpCard"}
	case DoubleAttackPhasePlay:
		h := g.activeBlackJackHand()
		if h == nil {
			return nil
		}
		if g.CanSplit() && h.GetCard(0).GetValue() == 1 {
			return &DoubleAttackHint{Action: "split", Reason: "splitAces"}
		}
		if h.GetScore() <= 11 {
			return &DoubleAttackHint{Action: "hit", Reason: "cannotBust"}
		}
		if h.GetScore() >= 17 {
			return &DoubleAttackHint{Action: "stand", Reason: "standPat"}
		}
		if up := g.dealerUpValue(); up >= 2 && up <= 6 {
			return &DoubleAttackHint{Action: "stand", Reason: "dealerMayBust"}
		}
		return &DoubleAttackHint{Action: "hit", Reason: "chaseDealer"}
	default:
		return nil
	}
}

// dealerUpValue はアップカードのブラックジャック上の値 (A は 11) を返す。
func (g *DoubleAttackBlackjack) dealerUpValue() int {
	if g.dealerHand == nil || g.dealerHand.GetCardsSize() == 0 {
		return 0
	}
	v := g.dealerHand.GetCard(0).GetValue()
	switch {
	case v == 1:
		return 11
	case v > 10:
		return 10
	default:
		return v
	}
}

// --- ログ ---

// appendLog は行動ログを 1 行足す。
func (g *DoubleAttackBlackjack) appendLog(actionType, detail string, cards []*Card) {
	g.turnNumber++
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.turnNumber,
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
	if len(g.actionLog) > doubleAttackMaxSliceLen {
		g.actionLog = g.actionLog[len(g.actionLog)-doubleAttackMaxSliceLen:]
	}
}

// --- アクセサ ---

// GetPhase は現在のフェーズを返す。
func (g *DoubleAttackBlackjack) GetPhase() DoubleAttackPhase { return g.phase }

// GetHands はプレイヤーの手札 (スプリットで増える) を返す。
func (g *DoubleAttackBlackjack) GetHands() []*BlackJackHand { return g.hands }

// GetHandCount は手札の数を返す。
func (g *DoubleAttackBlackjack) GetHandCount() int { return len(g.hands) }

// GetActiveHandIdx はいま操作している手札の位置を返す。
func (g *DoubleAttackBlackjack) GetActiveHandIdx() int { return g.activeHand }

// GetDealerCards はディーラーの札を返す。**2 枚目は追加ベットの後にしか無い。**
func (g *DoubleAttackBlackjack) GetDealerCards() []*Card {
	if g.dealerHand == nil {
		return nil
	}
	return g.dealerHand.GetCards()
}

// GetDealerScore はディーラーの点数を返す。
func (g *DoubleAttackBlackjack) GetDealerScore() int {
	if g.dealerHand == nil {
		return 0
	}
	return g.dealerHand.GetScore()
}

// IsDealerHoleDealt はディーラーの 2 枚目が配られたかを返す。
func (g *DoubleAttackBlackjack) IsDealerHoleDealt() bool { return g.dealerHoleDealt }

// GetAnteBet はアンティ額を返す。
func (g *DoubleAttackBlackjack) GetAnteBet() int { return g.anteBet }

// GetAttackBet は追加ベット額を返す。
func (g *DoubleAttackBlackjack) GetAttackBet() int { return g.attackBet }

// GetBustItBet は Bust It 額を返す。
func (g *DoubleAttackBlackjack) GetBustItBet() int { return g.bustItBet }

// GetResults は手札ごとの決着を返す。
func (g *DoubleAttackBlackjack) GetResults() []DoubleAttackResult { return g.results }

// GetPayout はこのラウンドで戻ってきた総額 (賭け金の返却を含む) を返す。
func (g *DoubleAttackBlackjack) GetPayout() int { return g.payout }

// GetBustItPayout は Bust It からの払い戻しを返す。
func (g *DoubleAttackBlackjack) GetBustItPayout() int { return g.bustItPayout }

// GetChips は保有チップ数を返す。
func (g *DoubleAttackBlackjack) GetChips() int { return g.player.GetChips() }

// SetChips は保有チップ数を設定する。
func (g *DoubleAttackBlackjack) SetChips(chips int) { g.player.SetChips(chips) }

// GetRoundNumber は現在のラウンド番号を返す。
func (g *DoubleAttackBlackjack) GetRoundNumber() int { return g.roundNumber }

// GetGameEndFlag はゲームが終了したかを返す。
func (g *DoubleAttackBlackjack) GetGameEndFlag() bool { return g.gameEndFlag }

// GetPlayer はプレイヤーを返す。
func (g *DoubleAttackBlackjack) GetPlayer() *DoubleAttackBlackjackPlayer { return g.player }

// GetConfig は設定を返す。
func (g *DoubleAttackBlackjack) GetConfig() DoubleAttackBlackjackConfig { return g.config }

// SetConfig は設定を差し替える。
func (g *DoubleAttackBlackjack) SetConfig(c DoubleAttackBlackjackConfig) { g.config = c }

// GetActionLog は行動ログを返す。
func (g *DoubleAttackBlackjack) GetActionLog() []*ActionLogEntry { return g.actionLog }

// GetRemainingCards はシューの残り枚数を返す。
func (g *DoubleAttackBlackjack) GetRemainingCards() int { return g.shoe.GetRemainingCount() }

// --- 永続化 ---

// doubleAttackJSON is the JSON wire format for DoubleAttackBlackjack.
type doubleAttackJSON struct {
	Shoe            *TrumpCards                  `json:"sh"`
	Player          *DoubleAttackBlackjackPlayer `json:"pl"`
	Config          DoubleAttackBlackjackConfig  `json:"cf"`
	Phase           int                          `json:"ph"`
	Hands           []*BlackJackHand             `json:"hd"`
	ActiveHand      int                          `json:"ah"`
	DealerHand      *BlackJackHand               `json:"dh"`
	DealerHoleDealt bool                         `json:"dd"`
	AnteBet         int                          `json:"an"`
	AttackBet       int                          `json:"at"`
	BustItBet       int                          `json:"bi"`
	Results         []int                        `json:"rs"`
	Payout          int                          `json:"po"`
	BustItPayout    int                          `json:"bp"`
	RoundNumber     int                          `json:"rn"`
	GameEndFlag     bool                         `json:"ge"`
	ActionLog       []*ActionLogEntry            `json:"al"`
	TurnNumber      int                          `json:"tn"`
}

// MarshalJSON implements json.Marshaler.
func (g *DoubleAttackBlackjack) MarshalJSON() ([]byte, error) {
	results := make([]int, 0, len(g.results))
	for _, r := range g.results {
		results = append(results, int(r))
	}
	return json.Marshal(doubleAttackJSON{
		Shoe: g.shoe, Player: g.player, Config: g.config,
		Phase: int(g.phase), Hands: g.hands, ActiveHand: g.activeHand,
		DealerHand: g.dealerHand, DealerHoleDealt: g.dealerHoleDealt,
		AnteBet: g.anteBet, AttackBet: g.attackBet, BustItBet: g.bustItBet,
		Results: results, Payout: g.payout, BustItPayout: g.bustItPayout,
		RoundNumber: g.roundNumber, GameEndFlag: g.gameEndFlag,
		ActionLog: g.actionLog, TurnNumber: g.turnNumber,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **範囲チェックだけでは足りない。** 追加ベットがアンティを超えている / 追加ベットを
// 置く前にディーラーの 2 枚目がある / 手札の数と決着の数が合わない、といった組み合わせは
// 個々の値としては range 内でも到達できない。壊れた保存データが静かに配当を変えるのを
// 防ぐため、**値どうしの整合**まで見る。
func (g *DoubleAttackBlackjack) UnmarshalJSON(data []byte) error {
	var j doubleAttackJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Player == nil {
		return fmt.Errorf("doubleattack: the player is missing")
	}
	if err := doubleAttackValidate(&j); err != nil {
		return err
	}

	g.shoe = j.Shoe
	if g.shoe == nil {
		g.shoe = NewTrumpCardsSpanish21(DoubleAttackDeckCount)
	}
	g.player = j.Player
	g.config = j.Config
	g.phase = DoubleAttackPhase(j.Phase)
	g.hands = j.Hands
	g.activeHand = j.ActiveHand
	g.dealerHand = j.DealerHand
	g.dealerHoleDealt = j.DealerHoleDealt
	g.anteBet = j.AnteBet
	g.attackBet = j.AttackBet
	g.bustItBet = j.BustItBet
	g.results = make([]DoubleAttackResult, 0, len(j.Results))
	for _, r := range j.Results {
		g.results = append(g.results, DoubleAttackResult(r))
	}
	g.payout = j.Payout
	g.bustItPayout = j.BustItPayout
	g.roundNumber = j.RoundNumber
	g.gameEndFlag = j.GameEndFlag
	g.actionLog = j.ActionLog
	g.turnNumber = j.TurnNumber
	return nil
}

// doubleAttackValidate は保存データの範囲と整合を検証する。
func doubleAttackValidate(j *doubleAttackJSON) error {
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < int(DoubleAttackPhaseBet) || j.Phase > int(DoubleAttackPhaseMax) {
		return fmt.Errorf("doubleattack: phase out of range: %d", j.Phase)
	}
	for name, v := range map[string]int{
		"ante": j.AnteBet, "double attack": j.AttackBet,
		"bust it": j.BustItBet, "payout": j.Payout,
	} {
		if v < 0 {
			return fmt.Errorf("doubleattack: %s must not be negative: %d", name, v)
		}
	}
	if j.RoundNumber < 1 {
		return fmt.Errorf("doubleattack: round number out of range: %d", j.RoundNumber)
	}
	if len(j.Hands) > DoubleAttackMaxHands {
		return fmt.Errorf("doubleattack: %d hands exceeds the split limit of %d",
			len(j.Hands), DoubleAttackMaxHands)
	}
	if len(j.Results) != len(j.Hands) {
		return fmt.Errorf("doubleattack: %d results for %d hands", len(j.Results), len(j.Hands))
	}
	for _, r := range j.Results {
		if r < int(DoubleAttackResultNone) || r > int(DoubleAttackResultMax) {
			return fmt.Errorf("doubleattack: result out of range: %d", r)
		}
	}
	if j.ActiveHand < 0 || (len(j.Hands) > 0 && j.ActiveHand > len(j.Hands)) {
		return fmt.Errorf("doubleattack: active hand out of range: %d", j.ActiveHand)
	}
	// **追加ベットはアンティを超えない。** ここを素通しすると、賭けていない額に
	// 配当が付く保存データが作れる。
	if j.AttackBet > j.AnteBet {
		return fmt.Errorf("doubleattack: the double attack bet %d exceeds the ante %d",
			j.AttackBet, j.AnteBet)
	}
	// **アップカードだけの間に 2 枚目があってはいけない。**
	if !j.DealerHoleDealt && j.DealerHand != nil && j.DealerHand.GetCardsSize() > 1 {
		return fmt.Errorf("doubleattack: the hole card is dealt before the double attack")
	}
	if j.Phase == int(DoubleAttackPhaseBet) && len(j.Hands) > 0 {
		return fmt.Errorf("doubleattack: cards are dealt before the ante is placed")
	}
	if len(j.ActionLog) > doubleAttackMaxSliceLen {
		return fmt.Errorf("doubleattack: action log too long: %d", len(j.ActionLog))
	}
	return nil
}
