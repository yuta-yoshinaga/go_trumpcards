//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// freeBetMaxSliceLen はデシリアライズ時のスライス長上限。
const freeBetMaxSliceLen = 1000

// エラー値。
var (
	errFreeBetWrongPhase   = errors.New("freebet: action not allowed in this phase")
	errFreeBetAnteRange    = errors.New("freebet: ante is out of range")
	errFreeBetAnteUnit     = errors.New("freebet: ante must be a multiple of the betting unit")
	errFreeBetNotEnough    = errors.New("freebet: not enough chips for that wager")
	errFreeBetCannotDouble = errors.New("freebet: this hand cannot take a free double")
	errFreeBetCannotSplit  = errors.New("freebet: this hand cannot take a free split")
	errFreeBetFinished     = errors.New("freebet: the player is out of chips")
)

// FreeBetResult は 1 つの手札の決着。
type FreeBetResult int

// 決着の種類。
const (
	// FreeBetResultNone まだ決着していない
	FreeBetResultNone FreeBetResult = iota
	// FreeBetResultWin プレイヤーの勝ち
	FreeBetResultWin
	// FreeBetResultLose ディーラーの勝ち
	FreeBetResultLose
	// FreeBetResultPush 引き分け
	FreeBetResultPush
	// FreeBetResultBlackjack プレイヤーのブラックジャック (3:2)
	FreeBetResultBlackjack
	// FreeBetResultDealer22Push ディーラーが 22 でバストしたため引き分け
	FreeBetResultDealer22Push
)

// FreeBetResultMax は最大の決着値 (復元時の範囲検査に使う)。
const FreeBetResultMax = FreeBetResultDealer22Push

// FreeBetResultName は決着の識別子を返す (i18n キーの一部に使う)。
func FreeBetResultName(r FreeBetResult) string {
	switch r {
	case FreeBetResultWin:
		return "win"
	case FreeBetResultLose:
		return "lose"
	case FreeBetResultPush:
		return "push"
	case FreeBetResultBlackjack:
		return "blackjack"
	case FreeBetResultDealer22Push:
		return "dealer22Push"
	default:
		return "none"
	}
}

// FreeBetHint は人間への助言。
type FreeBetHint struct {
	// Action は薦める操作 ("hit" / "stand" / "freeDouble" / "freeSplit")。
	Action string
	// Reason は理由の識別子 (i18n キーの一部)。
	Reason string
}

// FreeBetBlackjack はフリーベット・ブラックジャックの卓。
//
// **ダブルとスプリットがハウス持ちになる代わりに、ディーラーの 22 が引き分けになる。**
// 両方セットで初めて成立する取引で、片方だけ実装すると期待値が大きく傾く。
//
// 賭け金は 2 本立てで管理する:
//   - `bet`     プレイヤー自身の金。負ければ失う。
//   - `freeBet` ハウスが出した金。勝てば同額が配当として付き、負けても失わない。
//
// **この 2 本を 1 本にまとめない。** まとめると「バストしても元のベット額しか
// 失わない」という規則を精算のあらゆる分岐で言い直す羽目になる。
type FreeBetBlackjack struct {
	shoe   *TrumpCards
	player *FreeBetBlackjackPlayer
	config FreeBetBlackjackConfig

	phase      FreeBetPhase
	hands      []*BlackJackHand
	freeBets   []int
	results    []FreeBetResult
	activeHand int
	dealerHand *BlackJackHand
	anteBet    int
	payout     int
	// dealerPushed22 はディーラーが 22 でバストしたか (表示用)。
	dealerPushed22 bool
	roundNumber    int
	gameEndFlag    bool
	actionLog      []*ActionLogEntry
	turnNumber     int
}

// NewFreeBetBlackjack は指定のシュー・プレイヤー・設定で卓を構築する。
func NewFreeBetBlackjack(shoe *TrumpCards, player *FreeBetBlackjackPlayer,
	config FreeBetBlackjackConfig,
) *FreeBetBlackjack {
	return &FreeBetBlackjack{shoe: shoe, player: player, config: config, roundNumber: 1}
}

// NewDefaultFreeBetBlackjack は既定の卓を構築する。
func NewDefaultFreeBetBlackjack() *FreeBetBlackjack {
	cfg := DefaultFreeBetBlackjackConfig()
	return NewFreeBetBlackjack(
		NewTrumpCardsWithDecks(FreeBetDeckCount, 0),
		NewFreeBetBlackjackPlayer(cfg.InitialChips), cfg)
}

// Reset はゲームを初期状態に戻す。
func (g *FreeBetBlackjack) Reset() {
	g.shoe = NewTrumpCardsWithDecks(FreeBetDeckCount, 0)
	g.shoe.Shuffle()
	g.player.SetChips(g.config.InitialChips)
	g.roundNumber = 1
	g.gameEndFlag = false
	g.actionLog = nil
	g.turnNumber = 0
	g.appendLog("start", "free bet blackjack begins", nil)
	g.startRound()
}

// startRound は 1 ラウンド分の状態を初期化する。
func (g *FreeBetBlackjack) startRound() {
	g.hands = nil
	g.freeBets = nil
	g.results = nil
	g.activeHand = 0
	g.dealerHand = nil
	g.anteBet = 0
	g.payout = 0
	g.dealerPushed22 = false
	g.phase = FreeBetPhaseBet
	g.ensureShoe()
}

// ensureShoe は残りが乏しければシューを組み直す。
func (g *FreeBetBlackjack) ensureShoe() {
	if g.shoe.GetRemainingCount() < 30 {
		g.shoe = NewTrumpCardsWithDecks(FreeBetDeckCount, 0)
		g.shoe.Shuffle()
	}
}

// NextRound は次のラウンドを始める。
func (g *FreeBetBlackjack) NextRound() error {
	if g.gameEndFlag {
		return errFreeBetFinished
	}
	if g.phase != FreeBetPhaseResult {
		return errFreeBetWrongPhase
	}
	g.roundNumber++
	g.startRound()
	if g.player.GetChips() < FreeBetAnteMin {
		g.gameEndFlag = true
		g.appendLog("gameEnd", "out of chips", nil)
	}
	return nil
}

// --- 賭けと配り ---

// PlaceBet はアンティを置いて配る。
func (g *FreeBetBlackjack) PlaceBet(ante int) error {
	if g.gameEndFlag {
		return errFreeBetFinished
	}
	if g.phase != FreeBetPhaseBet {
		return errFreeBetWrongPhase
	}
	if ante < FreeBetAnteMin || ante > FreeBetAnteMax {
		return errFreeBetAnteRange
	}
	// **3:2 の配当が割り切れるように刻みを固定する。**
	if ante%FreeBetAnteUnit != 0 {
		return errFreeBetAnteUnit
	}
	if !g.player.SubtractChips(ante) {
		return errFreeBetNotEnough
	}
	g.anteBet = ante

	hand := NewBlackJackHand()
	hand.AddCard(g.shoe.DrawCard())
	hand.AddCard(g.shoe.DrawCard())
	hand.SetBet(ante)
	g.hands = []*BlackJackHand{hand}
	g.freeBets = []int{0}
	g.results = []FreeBetResult{FreeBetResultNone}

	g.dealerHand = NewBlackJackHand()
	g.dealerHand.AddCard(g.shoe.DrawCard())
	g.dealerHand.AddCard(g.shoe.DrawCard())

	g.appendLog("deal", fmt.Sprintf("ante %d", ante), hand.GetCards())

	if hand.IsBlackJack() || g.dealerHand.IsBlackJack() {
		g.settle()
		return nil
	}
	g.phase = FreeBetPhasePlay
	return nil
}

// --- プレイ ---

// activeBlackJackHand はいま操作している手札を返す。
func (g *FreeBetBlackjack) activeBlackJackHand() *BlackJackHand {
	if g.activeHand < 0 || g.activeHand >= len(g.hands) {
		return nil
	}
	return g.hands[g.activeHand]
}

// playableHand は操作できる手札を返す。できなければエラー。
func (g *FreeBetBlackjack) playableHand() (*BlackJackHand, error) {
	if g.gameEndFlag {
		return nil, errFreeBetFinished
	}
	if g.phase != FreeBetPhasePlay {
		return nil, errFreeBetWrongPhase
	}
	h := g.activeBlackJackHand()
	if h == nil || h.IsStood() || h.IsBusted() {
		return nil, errFreeBetWrongPhase
	}
	return h, nil
}

// Hit は 1 枚引く。
func (g *FreeBetBlackjack) Hit() error {
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
func (g *FreeBetBlackjack) Stand() error {
	h, err := g.playableHand()
	if err != nil {
		return err
	}
	h.SetStood(true)
	g.appendLog("stand", "stand", nil)
	g.advanceHand()
	return nil
}

// CanFreeDouble はいまの手札を無料ダブルできるかを返す。
//
// **ハードの 9-11 だけ。** ソフトは含まない。
func (g *FreeBetBlackjack) CanFreeDouble() bool {
	h := g.activeBlackJackHand()
	if h == nil || g.phase != FreeBetPhasePlay || h.IsStood() || h.IsBusted() {
		return false
	}
	return FreeBetCanFreeDouble(h.GetScore(), h.IsSoft(), h.GetCardsSize())
}

// FreeDouble はハウス持ちで賭け金を倍にし、1 枚だけ引いてその手札を終える。
//
// **プレイヤーはチップを出さない。** 勝てば倍額の配当が付き、負けても元の賭け金
// しか失わない ── これが 22 プッシュの対価。
func (g *FreeBetBlackjack) FreeDouble() error {
	h, err := g.playableHand()
	if err != nil {
		return err
	}
	if !g.CanFreeDouble() {
		return errFreeBetCannotDouble
	}
	g.freeBets[g.activeHand] += h.GetBet()
	h.SetDoubled(true)
	h.AddCard(g.shoe.DrawCard())
	g.appendLog("freeDouble", fmt.Sprintf("free double to %d", h.GetBet()+g.freeBets[g.activeHand]),
		h.GetCards())
	if h.GetScore() > 21 {
		h.SetBusted(true)
	} else {
		h.SetStood(true)
	}
	g.advanceHand()
	return nil
}

// CanFreeSplit はいまの手札を無料スプリットできるかを返す。
//
// **10 点札のペアは割れない。**
func (g *FreeBetBlackjack) CanFreeSplit() bool {
	h := g.activeBlackJackHand()
	if h == nil || g.phase != FreeBetPhasePlay || h.IsStood() || h.IsBusted() {
		return false
	}
	if len(g.hands) >= FreeBetMaxHands || h.GetCardsSize() != 2 {
		return false
	}
	return FreeBetCanFreeSplit(h.GetCard(0).GetValue(), h.GetCard(1).GetValue())
}

// FreeSplit はハウス持ちで手札を 2 つに分ける。
//
// **2 つ目の手札はまるごとハウスの金。** プレイヤーはチップを出さない。
func (g *FreeBetBlackjack) FreeSplit() error {
	h, err := g.playableHand()
	if err != nil {
		return err
	}
	if !g.CanFreeSplit() {
		return errFreeBetCannotSplit
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
	// **2 つ目はプレイヤーの金が 0。** 賭け金はすべて freeBets 側で持つ。
	other.SetBet(0)

	if isAces {
		rebuilt.SetStood(true)
		other.SetStood(true)
	}

	freeHere := g.freeBets[g.activeHand]
	g.hands[g.activeHand] = rebuilt
	g.insertAt(g.activeHand+1, other, h.GetBet(), FreeBetResultNone)
	g.freeBets[g.activeHand] = freeHere

	g.appendLog("freeSplit", fmt.Sprintf("free split into %d hands", len(g.hands)), nil)
	if isAces {
		g.advanceHand()
	}
	return nil
}

// insertAt は位置 idx に手札を差し込む (3 本のスライスを揃えて動かす)。
func (g *FreeBetBlackjack) insertAt(idx int, h *BlackJackHand, free int, r FreeBetResult) {
	g.hands = append(g.hands, nil)
	copy(g.hands[idx+1:], g.hands[idx:])
	g.hands[idx] = h

	g.freeBets = append(g.freeBets, 0)
	copy(g.freeBets[idx+1:], g.freeBets[idx:])
	g.freeBets[idx] = free

	g.results = append(g.results, FreeBetResultNone)
	copy(g.results[idx+1:], g.results[idx:])
	g.results[idx] = r
}

// advanceHand は次の未決着の手札へ進み、無ければディーラーの番にする。
func (g *FreeBetBlackjack) advanceHand() {
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
func (g *FreeBetBlackjack) dealerPlay() {
	for {
		score, soft := calcScore(g.dealerHand.GetCards())
		if score > 21 {
			g.dealerHand.SetBusted(true)
			break
		}
		if score > 17 || (score == 17 && (!soft || !FreeBetDealerHitsSoft17)) {
			break
		}
		g.dealerHand.AddCard(g.shoe.DrawCard())
	}
	g.appendLog("dealer", "dealer plays", g.dealerHand.GetCards())
	g.settle()
}

// settle はすべての手札を精算する。
func (g *FreeBetBlackjack) settle() {
	dealerScore := g.dealerHand.GetScore()
	dealerBJ := g.dealerHand.IsBlackJack()
	// **22 でのバストだけが引き分けになる。** 23 以上は普通のバスト。
	g.dealerPushed22 = dealerScore == FreeBetDealerPushTotal

	total := 0
	for i, h := range g.hands {
		r, ret := g.settleHand(h, g.freeBets[i], dealerScore, dealerBJ)
		g.results[i] = r
		total += ret
	}
	g.payout = total
	g.player.AddChips(g.payout)
	g.appendLog("result", fmt.Sprintf("payout %d", g.payout), nil)
	g.phase = FreeBetPhaseResult
}

// settleHand は 1 つの手札を精算し、決着と払い戻し (賭け金の返却を含む) を返す。
//
// **払い戻しはプレイヤーの金 (bet) とハウスの金 (freeBet) で扱いが違う。**
// 勝てば free のぶんも配当として付くが、負けても free は失わない (もともと
// プレイヤーの金ではない)。
func (g *FreeBetBlackjack) settleHand(h *BlackJackHand, free, dealerScore int,
	dealerBJ bool,
) (FreeBetResult, int) {
	bet := h.GetBet()
	switch {
	case h.IsBusted():
		// **バストは先に負けている。** ディーラーの 22 は関係しない。
		return FreeBetResultLose, 0
	case h.IsBlackJack() && !dealerBJ:
		return FreeBetResultBlackjack, bet + bet*FreeBetBlackjackPayoutNum/FreeBetBlackjackPayoutDen
	case dealerBJ && !h.IsBlackJack():
		return FreeBetResultLose, 0
	case dealerBJ && h.IsBlackJack():
		return FreeBetResultPush, bet
	case g.dealerPushed22:
		// **これが無料ダブル / 無料スプリットの対価。** 勝てたはずの手が引き分けになる。
		return FreeBetResultDealer22Push, bet
	case dealerScore > 21:
		return FreeBetResultWin, bet*2 + free
	case h.GetScore() > dealerScore:
		return FreeBetResultWin, bet*2 + free
	case h.GetScore() < dealerScore:
		return FreeBetResultLose, 0
	default:
		return FreeBetResultPush, bet
	}
}

// --- ヒント ---

// GetHint は人間への助言を返す。判断どころでなければ nil。
func (g *FreeBetBlackjack) GetHint() *FreeBetHint {
	if g.gameEndFlag || g.phase != FreeBetPhasePlay {
		return nil
	}
	h := g.activeBlackJackHand()
	if h == nil {
		return nil
	}
	// **タダなら使う。** 負けても元の賭け金しか失わないので、条件を満たすなら常に得。
	if g.CanFreeSplit() {
		return &FreeBetHint{Action: "freeSplit", Reason: "freeIsFree"}
	}
	if g.CanFreeDouble() {
		return &FreeBetHint{Action: "freeDouble", Reason: "freeIsFree"}
	}
	if h.GetScore() <= 11 {
		return &FreeBetHint{Action: "hit", Reason: "cannotBust"}
	}
	if h.GetScore() >= 17 {
		return &FreeBetHint{Action: "stand", Reason: "standPat"}
	}
	if up := g.dealerUpValue(); up >= 2 && up <= 6 {
		return &FreeBetHint{Action: "stand", Reason: "dealerMayBust"}
	}
	return &FreeBetHint{Action: "hit", Reason: "chaseDealer"}
}

// dealerUpValue はアップカードの値 (A は 11) を返す。
func (g *FreeBetBlackjack) dealerUpValue() int {
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
func (g *FreeBetBlackjack) appendLog(actionType, detail string, cards []*Card) {
	g.turnNumber++
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.turnNumber,
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
	if len(g.actionLog) > freeBetMaxSliceLen {
		g.actionLog = g.actionLog[len(g.actionLog)-freeBetMaxSliceLen:]
	}
}

// --- アクセサ ---

// GetPhase は現在のフェーズを返す。
func (g *FreeBetBlackjack) GetPhase() FreeBetPhase { return g.phase }

// GetHands はプレイヤーの手札を返す。
func (g *FreeBetBlackjack) GetHands() []*BlackJackHand { return g.hands }

// GetHandCount は手札の数を返す。
func (g *FreeBetBlackjack) GetHandCount() int { return len(g.hands) }

// GetFreeBets は手札ごとの**ハウスが出した額**を返す。
func (g *FreeBetBlackjack) GetFreeBets() []int { return g.freeBets }

// GetFreeBet は手札 idx のハウス出資額を返す。
func (g *FreeBetBlackjack) GetFreeBet(idx int) int {
	if idx < 0 || idx >= len(g.freeBets) {
		return 0
	}
	return g.freeBets[idx]
}

// GetActiveHandIdx はいま操作している手札の位置を返す。
func (g *FreeBetBlackjack) GetActiveHandIdx() int { return g.activeHand }

// GetDealerCards はディーラーの札を返す。
func (g *FreeBetBlackjack) GetDealerCards() []*Card {
	if g.dealerHand == nil {
		return nil
	}
	return g.dealerHand.GetCards()
}

// GetDealerScore はディーラーの点数を返す。
func (g *FreeBetBlackjack) GetDealerScore() int {
	if g.dealerHand == nil {
		return 0
	}
	return g.dealerHand.GetScore()
}

// IsDealerPushed22 はディーラーが 22 でバストしたかを返す。
func (g *FreeBetBlackjack) IsDealerPushed22() bool { return g.dealerPushed22 }

// GetAnteBet はアンティ額を返す。
func (g *FreeBetBlackjack) GetAnteBet() int { return g.anteBet }

// GetResults は手札ごとの決着を返す。
func (g *FreeBetBlackjack) GetResults() []FreeBetResult { return g.results }

// GetPayout はこのラウンドで戻ってきた総額を返す。
func (g *FreeBetBlackjack) GetPayout() int { return g.payout }

// GetChips は保有チップ数を返す。
func (g *FreeBetBlackjack) GetChips() int { return g.player.GetChips() }

// SetChips は保有チップ数を設定する。
func (g *FreeBetBlackjack) SetChips(chips int) { g.player.SetChips(chips) }

// GetRoundNumber は現在のラウンド番号を返す。
func (g *FreeBetBlackjack) GetRoundNumber() int { return g.roundNumber }

// GetGameEndFlag はゲームが終了したかを返す。
func (g *FreeBetBlackjack) GetGameEndFlag() bool { return g.gameEndFlag }

// GetPlayer はプレイヤーを返す。
func (g *FreeBetBlackjack) GetPlayer() *FreeBetBlackjackPlayer { return g.player }

// GetConfig は設定を返す。
func (g *FreeBetBlackjack) GetConfig() FreeBetBlackjackConfig { return g.config }

// SetConfig は設定を差し替える。
func (g *FreeBetBlackjack) SetConfig(c FreeBetBlackjackConfig) { g.config = c }

// GetActionLog は行動ログを返す。
func (g *FreeBetBlackjack) GetActionLog() []*ActionLogEntry { return g.actionLog }

// GetRemainingCards はシューの残り枚数を返す。
func (g *FreeBetBlackjack) GetRemainingCards() int { return g.shoe.GetRemainingCount() }

// --- 永続化 ---

// freeBetJSON is the JSON wire format for FreeBetBlackjack.
type freeBetJSON struct {
	Shoe           *TrumpCards             `json:"sh"`
	Player         *FreeBetBlackjackPlayer `json:"pl"`
	Config         FreeBetBlackjackConfig  `json:"cf"`
	Phase          int                     `json:"ph"`
	Hands          []*BlackJackHand        `json:"hd"`
	FreeBets       []int                   `json:"fb"`
	Results        []int                   `json:"rs"`
	ActiveHand     int                     `json:"ah"`
	DealerHand     *BlackJackHand          `json:"dh"`
	AnteBet        int                     `json:"an"`
	Payout         int                     `json:"po"`
	DealerPushed22 bool                    `json:"d22"`
	RoundNumber    int                     `json:"rn"`
	GameEndFlag    bool                    `json:"ge"`
	ActionLog      []*ActionLogEntry       `json:"al"`
	TurnNumber     int                     `json:"tn"`
}

// MarshalJSON implements json.Marshaler.
func (g *FreeBetBlackjack) MarshalJSON() ([]byte, error) {
	results := make([]int, 0, len(g.results))
	for _, r := range g.results {
		results = append(results, int(r))
	}
	return json.Marshal(freeBetJSON{
		Shoe: g.shoe, Player: g.player, Config: g.config,
		Phase: int(g.phase), Hands: g.hands, FreeBets: g.freeBets, Results: results,
		ActiveHand: g.activeHand, DealerHand: g.dealerHand, AnteBet: g.anteBet,
		Payout: g.payout, DealerPushed22: g.dealerPushed22,
		RoundNumber: g.roundNumber, GameEndFlag: g.gameEndFlag,
		ActionLog: g.actionLog, TurnNumber: g.turnNumber,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **3 本のスライスが揃っていることが不変条件。** 手札 / ハウス出資 / 決着は必ず同数で、
// ここがずれると精算がインデックス違いで別の手札の金を動かす。範囲チェックだけでは
// 捕まらないので、長さの一致まで見る。
func (g *FreeBetBlackjack) UnmarshalJSON(data []byte) error {
	var j freeBetJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Player == nil {
		return fmt.Errorf("freebet: the player is missing")
	}
	if err := freeBetValidate(&j); err != nil {
		return err
	}

	g.shoe = j.Shoe
	if g.shoe == nil {
		g.shoe = NewTrumpCardsWithDecks(FreeBetDeckCount, 0)
	}
	g.player = j.Player
	g.config = j.Config
	g.phase = FreeBetPhase(j.Phase)
	g.hands = j.Hands
	g.freeBets = j.FreeBets
	g.results = make([]FreeBetResult, 0, len(j.Results))
	for _, r := range j.Results {
		g.results = append(g.results, FreeBetResult(r))
	}
	g.activeHand = j.ActiveHand
	g.dealerHand = j.DealerHand
	g.anteBet = j.AnteBet
	g.payout = j.Payout
	g.dealerPushed22 = j.DealerPushed22
	g.roundNumber = j.RoundNumber
	g.gameEndFlag = j.GameEndFlag
	g.actionLog = j.ActionLog
	g.turnNumber = j.TurnNumber
	return nil
}

// freeBetValidate は保存データの範囲と整合を検証する。
func freeBetValidate(j *freeBetJSON) error {
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < int(FreeBetPhaseBet) || j.Phase > int(FreeBetPhaseMax) {
		return fmt.Errorf("freebet: phase out of range: %d", j.Phase)
	}
	if j.AnteBet < 0 {
		return fmt.Errorf("freebet: ante must not be negative: %d", j.AnteBet)
	}
	if j.Payout < 0 {
		return fmt.Errorf("freebet: payout must not be negative: %d", j.Payout)
	}
	if j.RoundNumber < 1 {
		return fmt.Errorf("freebet: round number out of range: %d", j.RoundNumber)
	}
	if len(j.Hands) > FreeBetMaxHands {
		return fmt.Errorf("freebet: %d hands exceeds the split limit of %d",
			len(j.Hands), FreeBetMaxHands)
	}
	// **3 本のスライスは必ず同数。**
	if len(j.FreeBets) != len(j.Hands) {
		return fmt.Errorf("freebet: %d free bets for %d hands", len(j.FreeBets), len(j.Hands))
	}
	if len(j.Results) != len(j.Hands) {
		return fmt.Errorf("freebet: %d results for %d hands", len(j.Results), len(j.Hands))
	}
	for _, f := range j.FreeBets {
		if f < 0 {
			return fmt.Errorf("freebet: a free bet must not be negative: %d", f)
		}
	}
	for _, r := range j.Results {
		if r < int(FreeBetResultNone) || r > int(FreeBetResultMax) {
			return fmt.Errorf("freebet: result out of range: %d", r)
		}
	}
	if j.ActiveHand < 0 || (len(j.Hands) > 0 && j.ActiveHand > len(j.Hands)) {
		return fmt.Errorf("freebet: active hand out of range: %d", j.ActiveHand)
	}
	if j.Phase == int(FreeBetPhaseBet) && len(j.Hands) > 0 {
		return fmt.Errorf("freebet: cards are dealt before the ante is placed")
	}
	if j.AnteBet == 0 && len(j.Hands) > 0 {
		return fmt.Errorf("freebet: cards are dealt without an ante")
	}
	if len(j.ActionLog) > freeBetMaxSliceLen {
		return fmt.Errorf("freebet: action log too long: %d", len(j.ActionLog))
	}
	return nil
}
