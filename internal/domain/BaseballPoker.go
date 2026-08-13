//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
)

// BaseballPhase は進行の段階。
type BaseballPhase int

const (
	// BaseballPhaseBetting はベットラウンド。
	BaseballPhaseBetting BaseballPhase = iota
	// BaseballPhaseBuyIn は表の 3 を配られた席が買い増しか降りかを選ぶ場面。
	BaseballPhaseBuyIn
	// BaseballPhaseShowdown はショーダウン。
	BaseballPhaseShowdown
	// BaseballPhaseGameEnd は終局。
	BaseballPhaseGameEnd
)

// BaseballPhaseMax は最大のフェーズ値 (復元時の範囲検査に使う)。
const BaseballPhaseMax = BaseballPhaseGameEnd

// 操作。**共有のベット定数をそのまま使う。**
const (
	BaseballActionFold  = bettingActionFold
	BaseballActionCheck = bettingActionCheck
	BaseballActionCall  = bettingActionCall
	BaseballActionBet   = bettingActionBet
	BaseballActionRaise = bettingActionRaise
)

// 買い増しの返事。**「払う」と「降りる」しかない。**
const (
	// BaseballBuyPay はポットを買い増して続ける。
	BaseballBuyPay = 0
	// BaseballBuyFold は買わずに降りる。
	BaseballBuyFold = 1
)

// baseballMaxSliceLen は復元時に受け付けるスライスの上限。
const baseballMaxSliceLen = 512

// baseballMaxCpuSteps は CPU を進める 1 回あたりの上限。
const baseballMaxCpuSteps = 128

// baseballMaxRaisesPerRound は 1 ラウンドのレイズ上限。
const baseballMaxRaisesPerRound = 3

// baseballStreets は表札を配るストリートの数 (3rd〜6th)。
const baseballStreets = BaseballUpCards

// エラー値。
var (
	errBaseballNotBetting     = errors.New("baseballpoker: not a betting round")
	errBaseballNotYourTurn    = errors.New("baseballpoker: it is not your turn")
	errBaseballUnknownAction  = errors.New("baseballpoker: unknown action")
	errBaseballCannotCheck    = errors.New("baseballpoker: there is a bet to call")
	errBaseballCannotBet      = errors.New("baseballpoker: someone has already bet")
	errBaseballRaiseCapped    = errors.New("baseballpoker: the raise cap for this round is reached")
	errBaseballBetRange       = errors.New("baseballpoker: bet out of range")
	errBaseballNotBuying      = errors.New("baseballpoker: nobody has to buy the pot right now")
	errBaseballBadBuyAnswer   = errors.New("baseballpoker: answer pay or fold")
	errBaseballHandInProgress = errors.New("baseballpoker: the hand is still in progress")
	errBaseballGameOver       = errors.New("baseballpoker: the game is over")
	errBaseballTurnRange      = errors.New("baseballpoker: turn seat out of range")
	errBaseballPhaseRange     = errors.New("baseballpoker: phase out of range")
	errBaseballStreetRange    = errors.New("baseballpoker: street out of range")
	errBaseballSliceTooLong   = errors.New("baseballpoker: slice too long")
	errBaseballSeatCount      = errors.New("baseballpoker: seat count does not match the config")
	errBaseballNegativePot    = errors.New("baseballpoker: pot must not be negative")
	errBaseballBuyerRange     = errors.New("baseballpoker: buyer seat out of range")
	errBaseballBuyerPhase     = errors.New("baseballpoker: a buyer is set outside the buy-in phase")
)

// BaseballResult は 1 席のハンド結果。
type BaseballResult struct {
	PlayerIdx int
	HandRank  int
	// UsedWild は 3 や 9 を役に使ったか。
	UsedWild  bool
	WonAmount int
}

// BaseballPoker はベースボールポーカーの卓。
//
// **セブンカードスタッドに 2 つの名物ルールが乗ったゲーム。**
//
//  1. **3 と 9 は常にワイルド。** 山に 8 枚あるので、役の相場がまるごと
//     上がる ── ワイルド無しの感覚でフォールドすると降りすぎる。
//  2. **表向きに配られた札でイベントが起きる。** 4 なら伏せ札を 1 枚もらえ、
//     3 なら「そのときのポットを払って続けるか、降りるか」を迫られる。
//
// 2 つ目が効いていて、**同じ 3 が「ワイルドで嬉しい」と「表で出ると高くつく」
// の両方**になっている。伏せて配られた 3 はただのワイルドで、表で出た 3 だけが
// 請求書になる ── ワイルドの判定を配られた向きと混ぜると、この対比が消える。
type BaseballPoker struct {
	deck    *TrumpCards
	players []*BaseballPokerPlayer
	config  BaseballPokerConfig

	phase BaseballPhase
	// street は配り終えた表札の数 (0..4)。
	street int
	pot    int
	// currentBet はこのラウンドで合わせるべき額。
	currentBet int
	raiseCount int
	turn       int
	// actedFlags はこのラウンドで既に打ったか。
	actedFlags []bool
	// buyer は買い増しを迫られている席 (-1 なら誰もいない)。
	buyer int
	// buyCost は buyer が払うべき額。**迫られた時点のポットで固定する。**
	buyCost int

	handNumber  int
	results     []BaseballResult
	gameEndFlag bool

	actionLog  []*ActionLogEntry
	turnNumber int
}

// NewBaseballPoker は BaseballPoker を構築する。
func NewBaseballPoker(deck *TrumpCards, players []*BaseballPokerPlayer, config BaseballPokerConfig) *BaseballPoker {
	return &BaseballPoker{
		deck:       deck,
		players:    players,
		config:     config,
		buyer:      -1,
		actedFlags: make([]bool, len(players)),
		results:    make([]BaseballResult, 0, len(players)),
		actionLog:  make([]*ActionLogEntry, 0),
	}
}

// NewDefaultBaseballPoker は既定設定の卓を返す。
func NewDefaultBaseballPoker() *BaseballPoker {
	cfg := DefaultBaseballPokerConfig()
	return NewBaseballPoker(NewTrumpCards(0), NewBaseballPokerPlayersForTable(cfg.Seats, cfg.InitialChips), cfg)
}

// Reset はゲームを最初から始める。
func (g *BaseballPoker) Reset() {
	if err := g.config.Validate(); err != nil {
		g.config = DefaultBaseballPokerConfig()
	}
	g.players = NewBaseballPokerPlayersForTable(g.config.Seats, g.config.InitialChips)
	g.actedFlags = make([]bool, len(g.players))
	g.handNumber = 0
	g.gameEndFlag = false
	g.actionLog = g.actionLog[:0]
	g.turnNumber = 0
	g.startHand()
}

// startHand は 1 ハンドを配る。
func (g *BaseballPoker) startHand() {
	g.deck.Replenish()
	g.deck.Shuffle()
	g.pot = 0
	g.currentBet = 0
	g.raiseCount = 0
	g.street = 0
	g.buyer = -1
	g.buyCost = 0
	g.results = g.results[:0]
	g.handNumber++
	g.phase = BaseballPhaseBetting

	for _, p := range g.players {
		p.ResetForHand()
	}

	// アンティ。チップの足りない席はあるだけ出す。
	for _, p := range g.players {
		ante := min(g.config.Ante, p.GetChips())
		p.SubtractChips(ante)
		g.pot += ante
		if p.GetChips() == 0 {
			p.SetAllIn(true)
		}
	}

	// **2 伏せ + 1 表。** 3rd ストリートの形。
	for range BaseballDownCards {
		for _, p := range g.players {
			p.AddDealtCard(g.draw(), false)
		}
	}
	for i, p := range g.players {
		p.AddDealtCard(g.draw(), true)
		g.appendLog(i, "deal", "up", p.FaceUpCards())
	}
	g.street = 1

	for i := range g.results {
		g.results[i] = BaseballResult{}
	}
	g.resetRound()
	// 配った表札のイベントを解決してから最初の手番を決める。
	g.resolveUpCardEvents(0)
}

// draw は山から 1 枚引く。
func (g *BaseballPoker) draw() *Card {
	return g.deck.DrawCard()
}

// resetRound は次のベットラウンドの状態に戻す。
func (g *BaseballPoker) resetRound() {
	g.currentBet = 0
	g.raiseCount = 0
	for i := range g.actedFlags {
		g.actedFlags[i] = false
	}
	for _, p := range g.players {
		p.SetCurrentBet(0)
	}
	g.turn = g.firstActiveSeat()
}

// resolveUpCardEvents はそのストリートで表向きに配られた札のイベントを解決する。
//
// **4 は自動で処理し、3 は人に聞く。** ボーナス札は選択の余地がないので
// その場で配り、買い増しは払うか降りるかを選ばせる必要があるので、席を
// `buyer` に立てて手番を止める。`from` はこのストリートで配り始めた席。
func (g *BaseballPoker) resolveUpCardEvents(from int) {
	for i := from; i < len(g.players); i++ {
		p := g.players[i]
		if p.GetFolded() {
			continue
		}
		last := g.lastFaceUpCard(p)
		if last == nil {
			continue
		}
		if last.GetValue() == BaseballBonusFour {
			// **ボーナスは伏せて配る。** 表で配るとイベントが連鎖する。
			p.AddBonusCard(g.draw())
			g.appendLog(i, "bonus", "four", nil)
			continue
		}
		if last.GetValue() == BaseballWildThree {
			// **払う額はこの時点のポットで固定する。** 後続の席の買い増しで
			// ポットが膨らむため、解決時のポットで請求すると同じ 3 でも
			// 席順によって額が変わり、席順が有利不利になる。
			g.buyer = i
			g.buyCost = min(g.pot, p.GetChips())
			g.phase = BaseballPhaseBuyIn
			g.appendLog(i, "buyin", "asked", nil)
			return
		}
	}
	g.buyer = -1
	g.buyCost = 0
	g.phase = BaseballPhaseBetting
	g.turn = g.firstActiveSeat()
	if g.activePlayers() <= 1 {
		g.finishHand()
	}
}

// lastFaceUpCard はその席に最後に表向きで配られた札を返す。
func (g *BaseballPoker) lastFaceUpCard(p *BaseballPokerPlayer) *Card {
	cards, faceUp := p.GetCards(), p.GetFaceUp()
	for i := len(cards) - 1; i >= 0; i-- {
		if i < len(faceUp) && faceUp[i] {
			return cards[i]
		}
	}
	return nil
}

// AnswerBuyIn は買い増しの返事を処理する。
//
// **払えない席にも「降りる」がある。** チップが尽きている席に払わせると
// マイナスになるので、額はチップで頭打ちにしてある (実質オールイン)。
func (g *BaseballPoker) AnswerBuyIn(answer int) error {
	if g.gameEndFlag {
		return errBaseballGameOver
	}
	if g.phase != BaseballPhaseBuyIn || g.buyer < 0 {
		return errBaseballNotBuying
	}
	if answer != BaseballBuyPay && answer != BaseballBuyFold {
		return errBaseballBadBuyAnswer
	}
	i := g.buyer
	p := g.players[i]
	if answer == BaseballBuyFold {
		p.SetFolded(true)
		g.appendLog(i, "buyin", "fold", nil)
	} else {
		cost := min(g.buyCost, p.GetChips())
		p.SubtractChips(cost)
		g.pot += cost
		if p.GetChips() == 0 {
			p.SetAllIn(true)
		}
		g.appendLog(i, "buyin", "pay", nil)
	}
	// **同じストリートの残りの席も見る。** 表の 3 が 2 席に出ることがある。
	g.resolveUpCardEvents(i + 1)
	return nil
}

// PlayerAction は人間の手を処理する。
func (g *BaseballPoker) PlayerAction(action, amount int) error {
	if g.gameEndFlag {
		return errBaseballGameOver
	}
	if g.phase != BaseballPhaseBetting {
		return errBaseballNotBetting
	}
	if !g.IsHumanTurn() {
		return errBaseballNotYourTurn
	}
	if err := g.applyAction(g.turn, action, amount); err != nil {
		return err
	}
	g.advanceAfterAction()
	return nil
}

// applyAction は 1 手を席に適用する。
func (g *BaseballPoker) applyAction(i, action, amount int) error {
	p := g.players[i]
	toCall := g.currentBet - p.GetCurrentBet()

	switch action {
	case BaseballActionFold:
		p.SetFolded(true)
		g.appendLog(i, "fold", "", nil)
	case BaseballActionCheck:
		if toCall > 0 {
			return errBaseballCannotCheck
		}
		g.appendLog(i, "check", "", nil)
	case BaseballActionCall:
		g.moveToPot(p, min(toCall, p.GetChips()))
		g.appendLog(i, "call", "", nil)
	case BaseballActionBet:
		if g.currentBet > 0 {
			return errBaseballCannotBet
		}
		if amount <= 0 || amount > p.GetChips() {
			return errBaseballBetRange
		}
		g.moveToPot(p, amount)
		g.currentBet = p.GetCurrentBet()
		g.appendLog(i, "bet", "", nil)
	case BaseballActionRaise:
		if g.raiseCount >= baseballMaxRaisesPerRound {
			return errBaseballRaiseCapped
		}
		if amount <= 0 || toCall+amount > p.GetChips() {
			return errBaseballBetRange
		}
		g.moveToPot(p, toCall+amount)
		g.currentBet = p.GetCurrentBet()
		g.raiseCount++
		g.appendLog(i, "raise", "", nil)
	default:
		return errBaseballUnknownAction
	}
	g.actedFlags[i] = true
	return nil
}

// moveToPot は席からポットへチップを移す。
func (g *BaseballPoker) moveToPot(p *BaseballPokerPlayer, amount int) {
	amount = min(amount, p.GetChips())
	if amount <= 0 {
		return
	}
	p.SubtractChips(amount)
	p.SetCurrentBet(p.GetCurrentBet() + amount)
	g.pot += amount
	if p.GetChips() == 0 {
		p.SetAllIn(true)
	}
}

// advanceAfterAction は 1 手のあとの進行を決める。
func (g *BaseballPoker) advanceAfterAction() {
	if g.activePlayers() <= 1 {
		g.finishHand()
		return
	}
	if g.bettingRoundComplete() {
		g.nextStreet()
		return
	}
	g.turn = g.nextActiveSeat(g.turn)
}

// bettingRoundComplete はこのラウンドが終わったかを返す。
func (g *BaseballPoker) bettingRoundComplete() bool {
	for i, p := range g.players {
		if p.GetFolded() || p.GetAllIn() {
			continue
		}
		if !g.actedFlags[i] || p.GetCurrentBet() != g.currentBet {
			return false
		}
	}
	return true
}

// nextStreet は次の札を配る。
//
// **7th ストリートは伏せて配る。** ここだけ向きが違うので、表札のイベントは
// 起きない ── 最後の 1 枚で買い増しを迫られるとベットラウンドが成立しない。
func (g *BaseballPoker) nextStreet() {
	if g.street >= baseballStreets+1 {
		g.finishHand()
		return
	}
	faceUp := g.street < baseballStreets
	for _, p := range g.players {
		if p.GetFolded() {
			continue
		}
		p.AddDealtCard(g.draw(), faceUp)
	}
	g.street++
	g.resetRound()
	if faceUp {
		g.resolveUpCardEvents(0)
		return
	}
	if g.activePlayers() <= 1 {
		g.finishHand()
	}
}

// finishHand はポットを分配してハンドを閉じる。
func (g *BaseballPoker) finishHand() {
	g.phase = BaseballPhaseShowdown
	g.buyer = -1
	g.buyCost = 0
	g.results = g.results[:0]
	for range g.players {
		g.results = append(g.results, BaseballResult{})
	}

	eligible := make([]int, 0, len(g.players))
	best := -1
	for i, p := range g.players {
		g.results[i].PlayerIdx = i
		if p.GetFolded() {
			continue
		}
		rank := p.EvaluateBest()
		g.results[i].HandRank = rank
		g.results[i].UsedWild = p.GetUsedWild()
		if rank > best {
			best = rank
			eligible = eligible[:0]
			eligible = append(eligible, i)
		} else if rank == best {
			eligible = append(eligible, i)
		}
	}
	if len(eligible) == 0 {
		// 全員降りた (起こり得ないが、ポットを取り残さない)。
		eligible = append(eligible, 0)
	}

	// **端数は若い席から配る。** ポットに残すと卓からチップが消える。
	share := g.pot / len(eligible)
	rest := g.pot % len(eligible)
	for n, i := range eligible {
		won := share
		if n < rest {
			won++
		}
		g.players[i].AddChips(won)
		g.results[i].WonAmount = won
	}
	g.pot = 0

	if g.aliveSeats() <= 1 || g.humanIsBroke() {
		g.finish()
	}
}

// NextHand は次のハンドを始める。
func (g *BaseballPoker) NextHand() error {
	if g.gameEndFlag {
		return errBaseballGameOver
	}
	if g.phase != BaseballPhaseShowdown {
		return errBaseballHandInProgress
	}
	g.startHand()
	return nil
}

// finish は終局にする。
func (g *BaseballPoker) finish() {
	g.gameEndFlag = true
	g.phase = BaseballPhaseGameEnd
	g.buyer = -1
}

// humanIsBroke は人間がアンティを払えなくなったかを返す。
func (g *BaseballPoker) humanIsBroke() bool {
	p := g.players[g.HumanSeat()]
	return p.GetChips() <= 0
}

// CpuPlay は CPU の手番を進める。
func (g *BaseballPoker) CpuPlay() { g.advanceCpu() }

// advanceCpu は人間の手番か終局まで CPU を進める。
func (g *BaseballPoker) advanceCpu() {
	for range baseballMaxCpuSteps {
		if g.gameEndFlag || g.phase == BaseballPhaseShowdown {
			return
		}
		// 買い増しを迫られているのが CPU なら、その場で答えさせる。
		if g.phase == BaseballPhaseBuyIn {
			if g.buyer < 0 || g.players[g.buyer].GetIsHuman() {
				return
			}
			_ = g.AnswerBuyIn(g.cpuBuyAnswer(g.buyer))
			continue
		}
		if g.phase != BaseballPhaseBetting || g.IsHumanTurn() {
			return
		}
		action, amount := g.cpuDecide(g.turn)
		if err := g.applyAction(g.turn, action, amount); err != nil {
			// 打てない手を選んだら降りずにチェック / コールへ落とす。
			_ = g.applyAction(g.turn, BaseballActionCall, 0)
		}
		g.advanceAfterAction()
	}
}

// cpuBuyAnswer は CPU の買い増し判断を返す。
//
// **手札の強さで決める。** ポットが自分のチップに対して重いときは降りる ──
// 一律に払わせると、買い増しがただの徴税になって選択でなくなる。
func (g *BaseballPoker) cpuBuyAnswer(i int) int {
	p := g.players[i]
	cost := min(g.buyCost, p.GetChips())
	if cost <= 0 {
		return BaseballBuyPay
	}
	rank := p.EvaluateBest()
	if rank >= PokerHandTwoPair {
		return BaseballBuyPay
	}
	if rank >= PokerHandOnePair && cost*2 <= p.GetChips() {
		return BaseballBuyPay
	}
	if cost*4 <= p.GetChips() {
		return BaseballBuyPay
	}
	return BaseballBuyFold
}

// cpuDecide は CPU のベット判断を返す。
func (g *BaseballPoker) cpuDecide(i int) (int, int) {
	p := g.players[i]
	toCall := g.currentBet - p.GetCurrentBet()
	rank := p.EvaluateBest()

	if toCall <= 0 {
		if rank >= PokerHandThreeOfAKind && g.raiseCount < baseballMaxRaisesPerRound {
			return BaseballActionBet, min(g.config.Ante*2, p.GetChips())
		}
		return BaseballActionCheck, 0
	}
	if rank >= PokerHandFlush && g.raiseCount < baseballMaxRaisesPerRound &&
		toCall+g.config.Ante <= p.GetChips() {
		return BaseballActionRaise, g.config.Ante
	}
	if rank >= PokerHandOnePair || toCall <= g.config.Ante {
		return BaseballActionCall, 0
	}
	return BaseballActionFold, 0
}

// firstActiveSeat はラウンドの最初に打つ席を返す。
func (g *BaseballPoker) firstActiveSeat() int {
	for i, p := range g.players {
		if !p.GetFolded() && !p.GetAllIn() {
			return i
		}
	}
	return 0
}

// nextActiveSeat は from の次に打つ席を返す。
func (g *BaseballPoker) nextActiveSeat(from int) int {
	n := len(g.players)
	for k := 1; k <= n; k++ {
		i := (from + k) % n
		if !g.players[i].GetFolded() && !g.players[i].GetAllIn() {
			return i
		}
	}
	return from
}

// activePlayers はまだ降りていない席の数を返す。
func (g *BaseballPoker) activePlayers() int {
	n := 0
	for _, p := range g.players {
		if !p.GetFolded() {
			n++
		}
	}
	return n
}

// aliveSeats はチップの残っている席の数を返す。
func (g *BaseballPoker) aliveSeats() int {
	n := 0
	for _, p := range g.players {
		if p.GetChips() > 0 {
			n++
		}
	}
	return n
}

// HumanSeat は人間の席を返す。
func (g *BaseballPoker) HumanSeat() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return 0
}

// IsHumanTurn は人間の手番かを返す。
func (g *BaseballPoker) IsHumanTurn() bool {
	return g.phase == BaseballPhaseBetting && g.turn == g.HumanSeat() &&
		!g.players[g.turn].GetFolded() && !g.players[g.turn].GetAllIn()
}

// IsHumanBuying は人間が買い増しを迫られているかを返す。
func (g *BaseballPoker) IsHumanBuying() bool {
	return g.phase == BaseballPhaseBuyIn && g.buyer == g.HumanSeat()
}

// WinnerSeat はいちばんチップの多い席を返す。
func (g *BaseballPoker) WinnerSeat() int {
	best, seat := -1, 0
	for i, p := range g.players {
		if p.GetChips() > best {
			best, seat = p.GetChips(), i
		}
	}
	return seat
}

// GetConfig は設定を返す。
func (g *BaseballPoker) GetConfig() BaseballPokerConfig { return g.config }

// SetConfig は設定を差し替える。
func (g *BaseballPoker) SetConfig(c BaseballPokerConfig) { g.config = c }

// GetPhase は現在のフェーズを返す。
func (g *BaseballPoker) GetPhase() BaseballPhase { return g.phase }

// GetGameEndFlag は終局かを返す。
func (g *BaseballPoker) GetGameEndFlag() bool { return g.gameEndFlag }

// GetPlayers は席の一覧を返す。
func (g *BaseballPoker) GetPlayers() []*BaseballPokerPlayer { return g.players }

// GetStreet は配り終えた表札の数を返す。
func (g *BaseballPoker) GetStreet() int { return g.street }

// GetPot は現在のポットを返す。
func (g *BaseballPoker) GetPot() int { return g.pot }

// GetCurrentBet はこのラウンドで合わせる額を返す。
func (g *BaseballPoker) GetCurrentBet() int { return g.currentBet }

// GetToCall は人間が払うべき額を返す。
func (g *BaseballPoker) GetToCall() int {
	p := g.players[g.HumanSeat()]
	return max(0, g.currentBet-p.GetCurrentBet())
}

// GetRaiseCount はこのラウンドのレイズ回数を返す。
func (g *BaseballPoker) GetRaiseCount() int { return g.raiseCount }

// CanRaise はまだレイズできるかを返す。
func (g *BaseballPoker) CanRaise() bool {
	return g.phase == BaseballPhaseBetting && g.raiseCount < baseballMaxRaisesPerRound
}

// GetTurnSeat は手番の席を返す。
func (g *BaseballPoker) GetTurnSeat() int { return g.turn }

// GetBuyerSeat は買い増しを迫られている席を返す (-1 なら誰もいない)。
func (g *BaseballPoker) GetBuyerSeat() int { return g.buyer }

// GetBuyCost は買い増しの額を返す。
func (g *BaseballPoker) GetBuyCost() int { return g.buyCost }

// GetHandNumber は何ハンド目かを返す。
func (g *BaseballPoker) GetHandNumber() int { return g.handNumber }

// GetResults は直近のハンド結果を返す。
func (g *BaseballPoker) GetResults() []BaseballResult { return g.results }

// GetRemainingCards は山の残り枚数を返す。
func (g *BaseballPoker) GetRemainingCards() int { return g.deck.GetRemainingCount() }

// GetActionLog は棋譜を返す。
func (g *BaseballPoker) GetActionLog() []*ActionLogEntry { return g.actionLog }

// appendLog は棋譜に 1 行足す。
func (g *BaseballPoker) appendLog(seat int, actionType, detail string, cards []*Card) {
	g.turnNumber++
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.turnNumber,
		PlayerIdx:  seat,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- 助言 ---

// BaseballHint は人間への助言。
type BaseballHint struct {
	// Action は薦める操作 ("fold" / "check" / "call" / "bet" / "raise" / "pay")。
	Action string
	// Reason は理由の識別子 (i18n キーの一部)。
	Reason string
}

// GetHint は人間への助言を返す。
//
// **ワイルドが 8 枚あることを前提に薦める。** 3 と 9 で役が跳ねるので、
// ワイルド無しの相場でフォールドすると降りすぎになる。
func (g *BaseballPoker) GetHint() *BaseballHint {
	if g.gameEndFlag {
		return nil
	}
	p := g.players[g.HumanSeat()]

	if g.IsHumanBuying() {
		cost := min(g.buyCost, p.GetChips())
		rank := p.EvaluateBest()
		if rank >= PokerHandTwoPair {
			return &BaseballHint{Action: "pay", Reason: "handIsWorthTheBuy"}
		}
		if cost*3 <= p.GetChips() {
			return &BaseballHint{Action: "pay", Reason: "buyIsCheapEnough"}
		}
		return &BaseballHint{Action: "fold", Reason: "buyCostsTooMuch"}
	}
	if g.phase != BaseballPhaseBetting || !g.IsHumanTurn() {
		return nil
	}

	rank := p.EvaluateBest()
	toCall := g.GetToCall()

	if toCall <= 0 {
		if rank >= PokerHandThreeOfAKind && g.CanRaise() {
			return &BaseballHint{Action: "bet", Reason: "strongEnoughToBet"}
		}
		return &BaseballHint{Action: "check", Reason: "seeAnotherCard"}
	}
	if rank >= PokerHandFlush && g.CanRaise() {
		return &BaseballHint{Action: "raise", Reason: "strongEnoughToRaise"}
	}
	// **ワイルドが 8 枚あるので、ツーペアでは足りないことが多い。**
	if rank >= PokerHandTwoPair {
		return &BaseballHint{Action: "call", Reason: "worthACall"}
	}
	if toCall <= g.config.Ante {
		return &BaseballHint{Action: "call", Reason: "cheapToStay"}
	}
	return &BaseballHint{Action: "fold", Reason: "wildsRaiseTheBar"}
}

// --- 永続化 ---

// baseballJSON is the JSON wire format for BaseballPoker.
type baseballJSON struct {
	Deck        *TrumpCards            `json:"dk"`
	Players     []*BaseballPokerPlayer `json:"pl"`
	Config      BaseballPokerConfig    `json:"cf"`
	Phase       int                    `json:"ph"`
	Street      int                    `json:"st"`
	Pot         int                    `json:"po"`
	CurrentBet  int                    `json:"cb"`
	RaiseCount  int                    `json:"rc"`
	Turn        int                    `json:"tu"`
	ActedFlags  []bool                 `json:"af"`
	Buyer       int                    `json:"by"`
	BuyCost     int                    `json:"bc"`
	HandNumber  int                    `json:"hn"`
	Results     []BaseballResult       `json:"rs"`
	GameEndFlag bool                   `json:"ge"`
	ActionLog   []*ActionLogEntry      `json:"al"`
	TurnNumber  int                    `json:"tn"`
}

// MarshalJSON implements json.Marshaler.
func (g *BaseballPoker) MarshalJSON() ([]byte, error) {
	return json.Marshal(baseballJSON{
		Deck: g.deck, Players: g.players, Config: g.config,
		Phase: int(g.phase), Street: g.street,
		Pot: g.pot, CurrentBet: g.currentBet, RaiseCount: g.raiseCount,
		Turn: g.turn, ActedFlags: g.actedFlags,
		Buyer: g.buyer, BuyCost: g.buyCost,
		HandNumber: g.handNumber, Results: g.results, GameEndFlag: g.gameEndFlag,
		ActionLog: g.actionLog, TurnNumber: g.turnNumber,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **買い増しの席はフェーズと整合していなければならない。** 席番号の範囲を
// 見るだけでは「ベット中なのに買い手が立っている」保存が通り、その席は
// 誰にも動かされないまま盤面が止まる ── 範囲検査ではなく組合せを見る。
func (g *BaseballPoker) UnmarshalJSON(data []byte) error {
	var j baseballJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := baseballValidate(&j); err != nil {
		return err
	}

	g.deck = j.Deck
	g.players = j.Players
	g.config = j.Config
	g.phase = BaseballPhase(j.Phase)
	g.street = j.Street
	g.pot = j.Pot
	g.currentBet = j.CurrentBet
	g.raiseCount = j.RaiseCount
	g.turn = j.Turn
	g.actedFlags = j.ActedFlags
	g.buyer = j.Buyer
	g.buyCost = j.BuyCost
	g.handNumber = j.HandNumber
	g.results = j.Results
	g.gameEndFlag = j.GameEndFlag
	g.actionLog = j.ActionLog
	g.turnNumber = j.TurnNumber

	if g.actedFlags == nil || len(g.actedFlags) != len(g.players) {
		g.actedFlags = make([]bool, len(g.players))
	}
	if g.results == nil {
		g.results = make([]BaseballResult, 0, len(g.players))
	}
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// baseballValidate は復元した盤面が破綻していないかを見る。
func baseballValidate(j *baseballJSON) error {
	if j.Deck == nil {
		return errBaseballSliceTooLong
	}
	if len(j.Players) > baseballMaxSliceLen || len(j.ActionLog) > baseballMaxSliceLen ||
		len(j.Results) > baseballMaxSliceLen || len(j.ActedFlags) > baseballMaxSliceLen {
		return errBaseballSliceTooLong
	}
	if len(j.Players) != j.Config.Seats {
		return errBaseballSeatCount
	}
	if j.Phase < 0 || BaseballPhase(j.Phase) > BaseballPhaseMax {
		return errBaseballPhaseRange
	}
	if j.Street < 0 || j.Street > baseballStreets+1 {
		return errBaseballStreetRange
	}
	if j.Turn < 0 || j.Turn >= len(j.Players) {
		return errBaseballTurnRange
	}
	if j.Pot < 0 || j.BuyCost < 0 || j.CurrentBet < 0 {
		return errBaseballNegativePot
	}
	if j.Buyer < -1 || j.Buyer >= len(j.Players) {
		return errBaseballBuyerRange
	}
	// **買い手とフェーズの整合。** 範囲検査では通ってしまう組合せを弾く。
	if (j.Buyer >= 0) != (BaseballPhase(j.Phase) == BaseballPhaseBuyIn) {
		return errBaseballBuyerPhase
	}
	return nil
}
