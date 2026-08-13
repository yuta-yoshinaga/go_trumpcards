//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// IronCrossPhase はゲームの進行段階。
type IronCrossPhase int

const (
	// IronCrossPhaseBetting はベットラウンド中。
	IronCrossPhaseBetting IronCrossPhase = iota
	// IronCrossPhaseChoose は縦か横かを選ぶ段階。
	IronCrossPhaseChoose
	// IronCrossPhaseShowdown はショーダウン。
	IronCrossPhaseShowdown
	// IronCrossPhaseGameEnd はゲーム終了。
	IronCrossPhaseGameEnd
)

// IronCrossPhaseMax は最大のフェーズ値 (復元時の範囲検査に使う)。
const IronCrossPhaseMax = IronCrossPhaseGameEnd

// アクション定数 (共通定数のエイリアス)。
const (
	IronCrossActionFold  = bettingActionFold
	IronCrossActionCheck = bettingActionCheck
	IronCrossActionCall  = bettingActionCall
	IronCrossActionBet   = bettingActionBet
	IronCrossActionRaise = bettingActionRaise
)

// ironCrossMaxSliceLen は復元時に許すスライス長の上限。
const ironCrossMaxSliceLen = 512

// ironCrossMaxCpuSteps は CPU を進める 1 回あたりの上限。
const ironCrossMaxCpuSteps = 128

// ironCrossMaxRaisesPerRound は 1 ラウンドのレイズ上限。
const ironCrossMaxRaisesPerRound = 3

// エラー値。
var (
	errIronCrossFinished    = errors.New("ironcross: game already finished")
	errIronCrossWrongPhase  = errors.New("ironcross: not allowed in this phase")
	errIronCrossNotYourRun  = errors.New("ironcross: not your turn")
	errIronCrossBadAction   = errors.New("ironcross: unknown action")
	errIronCrossCannotCheck = errors.New("ironcross: cannot check facing a bet")
	errIronCrossCannotBet   = errors.New("ironcross: cannot bet facing a bet")
	errIronCrossCannotRaise = errors.New("ironcross: nothing to raise")
	errIronCrossRaiseCapped = errors.New("ironcross: the raise cap for this round is reached")
	errIronCrossBetRange    = errors.New("ironcross: bet out of range")
	errIronCrossBadLine     = errors.New("ironcross: choose the vertical or the horizontal line")
)

// IronCrossResult は 1 席のハンド結果。
type IronCrossResult struct {
	PlayerIdx int
	// Line はその席が使った列。
	Line      IronCrossLine
	HandRank  int
	WonAmount int
}

// IronCross はアイアンクロス (クリスクロス) の卓。
//
// **十字の 5 枚は全員が見るが、使えるのは縦か横の 3 枚だけ。**
// 中央の 1 枚だけが両方に入るので、「どちらを選ぶか」で手が変わる ──
// 全員が同じ 5 枚を共有する Holdem との違いはここに尽きる。
//
//	    [1]
//	[3] [0] [4]
//	    [2]
//
// **十字は 1 枚ずつ開く。** 開くたびにベットラウンドが入り、5 枚出そろってから
// 縦横を選ぶ。選ぶ時点では全部見えているので、選択は運ではなく判断になる。
type IronCross struct {
	deck    *TrumpCards
	players []*IronCrossPlayer
	config  IronCrossConfig

	phase IronCrossPhase
	// cross は十字の 5 枚。**開いていない位置は nil。**
	cross []*Card
	// revealed は開いた枚数。
	revealed int

	pot         int
	currentBet  int
	raiseCount  int
	turn        int
	actedFlags  []bool
	handNumber  int
	results     []IronCrossResult
	gameEndFlag bool
	actionLog   []*ActionLogEntry
	turnNumber  int
}

// NewIronCross は指定の山・席・設定で卓を構築する。
func NewIronCross(deck *TrumpCards, players []*IronCrossPlayer, config IronCrossConfig) *IronCross {
	return &IronCross{
		deck: deck, players: players, config: config,
		cross:      make([]*Card, IronCrossCommunityCards),
		actedFlags: make([]bool, len(players)),
		handNumber: 1,
	}
}

// NewDefaultIronCross は既定の卓を構築する。席 0 が人間。
func NewDefaultIronCross() *IronCross {
	cfg := DefaultIronCrossConfig()
	return NewIronCross(NewTrumpCards(0),
		NewIronCrossPlayersForTable(cfg.Seats, cfg.InitialChips), cfg)
}

// Reset はゲームを初期化する。
func (g *IronCross) Reset() {
	for _, p := range g.players {
		p.SetChips(g.config.InitialChips)
	}
	g.handNumber = 1
	g.gameEndFlag = false
	g.actionLog = nil
	g.turnNumber = 0
	g.appendLog(-1, "reset", "game reset", nil)
	g.startHand()
}

// startHand は 1 ハンド配る。
func (g *IronCross) startHand() {
	g.deck.Replenish()
	g.deck.Shuffle()
	g.cross = make([]*Card, IronCrossCommunityCards)
	g.revealed = 0
	g.pot = 0
	g.currentBet = 0
	g.raiseCount = 0
	g.results = nil
	g.actedFlags = make([]bool, len(g.players))

	for _, p := range g.players {
		p.ResetForHand()
	}
	for _, p := range g.players {
		ante := min(g.config.Ante, p.GetChips())
		p.SubtractChips(ante)
		g.pot += ante
		if p.GetChips() == 0 {
			p.SetAllIn(true)
		}
	}
	for range IronCrossHoleCards {
		for _, p := range g.players {
			p.AddCard(g.deck.DrawCard())
		}
	}

	g.phase = IronCrossPhaseBetting
	g.turn = g.firstActiveSeat()
	g.appendLog(-1, "deal", fmt.Sprintf("hand %d, ante %d", g.handNumber, g.config.Ante), nil)
	g.advanceCpu()
}

// --- ベッティング ---

// PlayerAction は人間の手を処理する。
func (g *IronCross) PlayerAction(action, amount int) error {
	if g.gameEndFlag {
		return errIronCrossFinished
	}
	if g.phase != IronCrossPhaseBetting {
		return errIronCrossWrongPhase
	}
	if g.turn != g.HumanSeat() {
		return errIronCrossNotYourRun
	}
	if err := g.applyAction(g.turn, action, amount); err != nil {
		return err
	}
	g.advanceAfterAction()
	g.advanceCpu()
	return nil
}

// applyAction は席 i の手を適用する。
func (g *IronCross) applyAction(i, action, amount int) error {
	p := g.players[i]
	toCall := g.currentBet - p.GetCurrentBet()

	switch action {
	case IronCrossActionFold:
		p.SetFolded(true)
		g.appendLog(i, "fold", fmt.Sprintf("seat %d folds", i), nil)
	case IronCrossActionCheck:
		if toCall > 0 {
			return errIronCrossCannotCheck
		}
		g.appendLog(i, "check", fmt.Sprintf("seat %d checks", i), nil)
	case IronCrossActionCall:
		g.moveToPot(p, toCall)
		g.appendLog(i, "call", fmt.Sprintf("seat %d calls %d", i, toCall), nil)
	case IronCrossActionBet:
		if g.currentBet > 0 {
			return errIronCrossCannotBet
		}
		if amount < g.config.Ante || amount > p.GetChips() {
			return errIronCrossBetRange
		}
		g.moveToPot(p, amount)
		g.currentBet = p.GetCurrentBet()
		g.raiseCount++
		g.appendLog(i, "bet", fmt.Sprintf("seat %d bets %d", i, amount), nil)
	case IronCrossActionRaise:
		if g.currentBet == 0 {
			return errIronCrossCannotRaise
		}
		if g.raiseCount >= ironCrossMaxRaisesPerRound {
			return errIronCrossRaiseCapped
		}
		if amount < g.config.Ante || amount > p.GetChips()-toCall {
			return errIronCrossBetRange
		}
		g.moveToPot(p, toCall+amount)
		g.currentBet = p.GetCurrentBet()
		g.raiseCount++
		g.appendLog(i, "raise", fmt.Sprintf("seat %d raises %d", i, amount), nil)
	default:
		return errIronCrossBadAction
	}
	g.actedFlags[i] = true
	return nil
}

// moveToPot は席からポットへチップを移す。
func (g *IronCross) moveToPot(p *IronCrossPlayer, amount int) {
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

// advanceAfterAction は手番を進め、ラウンドが閉じていれば次の段階へ移る。
func (g *IronCross) advanceAfterAction() {
	if g.activePlayers() <= 1 {
		g.finishHand()
		return
	}
	if g.bettingRoundComplete() {
		g.nextStage()
		return
	}
	g.turn = g.nextActiveSeat(g.turn)
}

// bettingRoundComplete はラウンドが閉じたかを返す。
//
// **全員が 1 度は動き、かつ賭け額が揃っていること。** 片方だけだと、
// レイズに応じていない席を残したまま次の札を開いてしまう。
func (g *IronCross) bettingRoundComplete() bool {
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

// nextStage は十字を 1 枚開くか、選択の段階へ進む。
func (g *IronCross) nextStage() {
	if g.revealed >= IronCrossCommunityCards {
		// **5 枚そろってから選ぶ。** 全部見えている状態で選ばせるので、
		// 縦横の判断は運ではなく読みになる。
		g.phase = IronCrossPhaseChoose
		g.turn = g.HumanSeat()
		g.appendLog(-1, "choose", "all five are up; pick a line", nil)
		g.chooseForCpus()
		return
	}
	g.cross[g.revealOrder()] = g.deck.DrawCard()
	g.revealed++
	g.currentBet = 0
	g.raiseCount = 0
	g.actedFlags = make([]bool, len(g.players))
	for _, p := range g.players {
		p.SetCurrentBet(0)
	}
	g.turn = g.firstActiveSeat()
	g.appendLog(-1, "reveal", fmt.Sprintf("cross card %d", g.revealed), nil)

	if g.activePlayers() <= 1 {
		g.finishHand()
	}
}

// revealOrder は次に開く十字の位置を返す。
//
// **中央を最後に開く。** 中央は縦にも横にも入る唯一の札なので、最後まで
// 伏せておくと「どちらを選ぶか」が最後の 1 枚まで決まらない ── 端から開いて
// 中央で決着させるのが、この配置がいちばん効く順序。
func (g *IronCross) revealOrder() int {
	order := []int{IronCrossTop, IronCrossBottom, IronCrossLeft, IronCrossRight, IronCrossCenter}
	return order[min(g.revealed, len(order)-1)]
}

// chooseForCpus は CPU の列を決める。**最良のほうを採る。**
func (g *IronCross) chooseForCpus() {
	for i, p := range g.players {
		if i == g.HumanSeat() || p.GetFolded() {
			continue
		}
		p.SetLine(IronCrossLineNone)
		p.EvaluateBest(g.cross)
	}
}

// ChooseLine は人間が使う列を決め、決着まで進める。
func (g *IronCross) ChooseLine(l IronCrossLine) error {
	if g.gameEndFlag {
		return errIronCrossFinished
	}
	if g.phase != IronCrossPhaseChoose {
		return errIronCrossWrongPhase
	}
	if l != IronCrossLineVertical && l != IronCrossLineHorizontal {
		return errIronCrossBadLine
	}
	p := g.players[g.HumanSeat()]
	p.SetLine(l)
	g.appendLog(g.HumanSeat(), "line", fmt.Sprintf("seat %d takes the %s line",
		g.HumanSeat(), IronCrossLineName(l)), nil)
	g.finishHand()
	return nil
}

// finishHand は役を比べてポットを配る。
func (g *IronCross) finishHand() {
	g.phase = IronCrossPhaseShowdown
	g.results = make([]IronCrossResult, 0, len(g.players))

	bestRank, winners := -1, []int(nil)
	for i, p := range g.players {
		if p.GetFolded() {
			g.results = append(g.results, IronCrossResult{PlayerIdx: i, HandRank: -1})
			continue
		}
		rank := p.EvaluateBest(g.cross)
		g.results = append(g.results, IronCrossResult{PlayerIdx: i, Line: p.GetLine(), HandRank: rank})
		switch {
		case rank > bestRank:
			bestRank, winners = rank, []int{i}
		case rank == bestRank:
			winners = append(winners, i)
		}
	}

	// **端数は若い席から配る。** ポットに残すと卓から消える。
	if len(winners) > 0 {
		share := g.pot / len(winners)
		rest := g.pot - share*len(winners)
		for n, w := range winners {
			amount := share
			if n < rest {
				amount++
			}
			g.players[w].AddChips(amount)
			g.results[w].WonAmount = amount
		}
	}
	g.pot = 0
	g.appendLog(-1, "showdown", fmt.Sprintf("hand %d settled", g.handNumber), nil)
}

// NextHand は次のハンドを始める。
func (g *IronCross) NextHand() error {
	if g.gameEndFlag {
		return errIronCrossFinished
	}
	if g.phase != IronCrossPhaseShowdown {
		return errIronCrossWrongPhase
	}
	if g.aliveSeats() < IronCrossMinSeats || g.players[g.HumanSeat()].GetChips() <= 0 {
		g.finish()
		return nil
	}
	g.handNumber++
	g.startHand()
	return nil
}

// finish はゲームを終える。
func (g *IronCross) finish() {
	g.gameEndFlag = true
	g.phase = IronCrossPhaseGameEnd
	g.appendLog(-1, "gameEnd", fmt.Sprintf("winner seat %d", g.WinnerSeat()), nil)
}

// --- CPU ---

// CpuPlay は CPU の席を進める。
func (g *IronCross) CpuPlay() { g.advanceCpu() }

// advanceCpu は人間の手番になるかハンドが終わるまで CPU を進める。
func (g *IronCross) advanceCpu() {
	for range ironCrossMaxCpuSteps {
		if g.gameEndFlag || g.phase != IronCrossPhaseBetting || g.IsHumanTurn() {
			return
		}
		i := g.turn
		action, amount := g.cpuDecide(i)
		if err := g.applyAction(i, action, amount); err != nil {
			_ = g.applyAction(i, IronCrossActionFold, 0)
		}
		g.advanceAfterAction()
	}
}

// cpuDecide は CPU の手を決める。
//
// **見えている十字で、縦横の良いほうを見て決める。** 選択の余地があるぶん、
// 片方だけで測ると弱く見積もる。
func (g *IronCross) cpuDecide(i int) (int, int) {
	p := g.players[i]
	vRank, _ := p.EvaluateLine(g.cross, IronCrossLineVertical)
	hRank, _ := p.EvaluateLine(g.cross, IronCrossLineHorizontal)
	rank := max(vRank, hRank)
	toCall := g.currentBet - p.GetCurrentBet()

	if toCall <= 0 {
		if rank >= PokerHandTwoPair && g.raiseCount < ironCrossMaxRaisesPerRound {
			return IronCrossActionBet, g.config.Ante
		}
		return IronCrossActionCheck, 0
	}
	if rank >= PokerHandThreeOfAKind && g.raiseCount < ironCrossMaxRaisesPerRound {
		return IronCrossActionRaise, g.config.Ante
	}
	if rank >= PokerHandOnePair || toCall <= g.config.Ante {
		return IronCrossActionCall, 0
	}
	return IronCrossActionFold, 0
}

// --- 席の走査 ---

func (g *IronCross) firstActiveSeat() int {
	for i, p := range g.players {
		if !p.GetFolded() && !p.GetAllIn() {
			return i
		}
	}
	return 0
}

func (g *IronCross) nextActiveSeat(from int) int {
	for step := 1; step <= len(g.players); step++ {
		idx := (from + step) % len(g.players)
		p := g.players[idx]
		if !p.GetFolded() && !p.GetAllIn() {
			return idx
		}
	}
	return from
}

func (g *IronCross) activePlayers() int {
	n := 0
	for _, p := range g.players {
		if !p.GetFolded() {
			n++
		}
	}
	return n
}

func (g *IronCross) aliveSeats() int {
	n := 0
	for _, p := range g.players {
		if p.GetChips() > 0 {
			n++
		}
	}
	return n
}

// --- 参照 ---

// HumanSeat は人間の席を返す。
func (g *IronCross) HumanSeat() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return 0
}

// IsHumanTurn は人間の操作待ちかを返す。
func (g *IronCross) IsHumanTurn() bool {
	return g.phase == IronCrossPhaseBetting && g.turn == g.HumanSeat()
}

// IsChoosing は人間が列を選ぶ場面かを返す。
func (g *IronCross) IsChoosing() bool { return g.phase == IronCrossPhaseChoose }

// WinnerSeat はチップがいちばん多い席を返す。同点なら若い席。
func (g *IronCross) WinnerSeat() int {
	best, bestChips := 0, -1
	for i, p := range g.players {
		if p.GetChips() > bestChips {
			best, bestChips = i, p.GetChips()
		}
	}
	return best
}

// GetConfig はゲーム設定を返す。
func (g *IronCross) GetConfig() IronCrossConfig { return g.config }

// SetConfig はゲーム設定を設定する。
func (g *IronCross) SetConfig(c IronCrossConfig) { g.config = c }

// GetPhase は現在のフェーズを返す。
func (g *IronCross) GetPhase() IronCrossPhase { return g.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *IronCross) GetGameEndFlag() bool { return g.gameEndFlag }

// GetPlayers は席の一覧を返す。
func (g *IronCross) GetPlayers() []*IronCrossPlayer { return g.players }

// GetCross は十字の 5 枚を返す。**開いていない位置は nil。**
func (g *IronCross) GetCross() []*Card { return g.cross }

// GetRevealedCount は開いた枚数を返す。
func (g *IronCross) GetRevealedCount() int { return g.revealed }

// GetPot はポットを返す。
func (g *IronCross) GetPot() int { return g.pot }

// GetCurrentBet はこのラウンドの現在の賭け額を返す。
func (g *IronCross) GetCurrentBet() int { return g.currentBet }

// GetToCall は人間がコールに要する額を返す。
func (g *IronCross) GetToCall() int {
	p := g.players[g.HumanSeat()]
	return max(0, g.currentBet-p.GetCurrentBet())
}

// GetRaiseCount はこのラウンドのレイズ回数を返す。
func (g *IronCross) GetRaiseCount() int { return g.raiseCount }

// CanRaise はいまレイズできるかを返す。
func (g *IronCross) CanRaise() bool {
	return g.currentBet > 0 && g.raiseCount < ironCrossMaxRaisesPerRound
}

// GetTurnSeat はいまの手番を返す。
func (g *IronCross) GetTurnSeat() int { return g.turn }

// GetHandNumber はハンド数を返す。
func (g *IronCross) GetHandNumber() int { return g.handNumber }

// GetResults はハンドの結果を返す。
func (g *IronCross) GetResults() []IronCrossResult { return g.results }

// GetRemainingCards は山の残り枚数を返す。
func (g *IronCross) GetRemainingCards() int { return g.deck.GetRemainingCount() }

// GetActionLog は棋譜を返す。
func (g *IronCross) GetActionLog() []*ActionLogEntry { return g.actionLog }

// appendLog は棋譜に 1 行足す。
func (g *IronCross) appendLog(seat int, actionType, detail string, cards []*Card) {
	g.turnNumber++
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.turnNumber,
		PlayerIdx:  seat,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
	if len(g.actionLog) > ironCrossMaxSliceLen {
		g.actionLog = g.actionLog[len(g.actionLog)-ironCrossMaxSliceLen:]
	}
}

// --- 助言 ---

// IronCrossHint は人間への助言。
type IronCrossHint struct {
	// Action は薦める操作 ("fold" / "check" / "call" / "bet" / "raise" / "line")。
	Action string
	// Line は薦める列 (Action が "line" のときのみ)。
	Line IronCrossLine
	// Reason は理由の識別子 (i18n キーの一部)。
	Reason string
}

// GetHint は人間への助言を返す。
//
// **選ぶ場面では、強いほうの列を名指しする。** ここがこのゲームで唯一
// 取り返しのつかない選択なので、迷わせない。
func (g *IronCross) GetHint() *IronCrossHint {
	if g.gameEndFlag {
		return nil
	}
	p := g.players[g.HumanSeat()]

	if g.phase == IronCrossPhaseChoose {
		vRank, vBest := p.EvaluateLine(g.cross, IronCrossLineVertical)
		hRank, hBest := p.EvaluateLine(g.cross, IronCrossLineHorizontal)
		if hRank > vRank || (hRank == vRank && compareHighCardsSlice(hBest, vBest) > 0) {
			return &IronCrossHint{Action: "line", Line: IronCrossLineHorizontal, Reason: "horizontalIsBetter"}
		}
		return &IronCrossHint{Action: "line", Line: IronCrossLineVertical, Reason: "verticalIsBetter"}
	}
	if g.phase != IronCrossPhaseBetting || !g.IsHumanTurn() {
		return nil
	}

	vRank, _ := p.EvaluateLine(g.cross, IronCrossLineVertical)
	hRank, _ := p.EvaluateLine(g.cross, IronCrossLineHorizontal)
	rank := max(vRank, hRank)
	toCall := g.GetToCall()

	if toCall <= 0 {
		if rank >= PokerHandTwoPair && g.raiseCount < ironCrossMaxRaisesPerRound {
			return &IronCrossHint{Action: "bet", Reason: "strongEnoughToBet"}
		}
		return &IronCrossHint{Action: "check", Reason: "seeAnotherCard"}
	}
	if rank >= PokerHandThreeOfAKind && g.CanRaise() {
		return &IronCrossHint{Action: "raise", Reason: "strongEnoughToRaise"}
	}
	if rank >= PokerHandOnePair {
		return &IronCrossHint{Action: "call", Reason: "worthACall"}
	}
	if toCall <= g.config.Ante {
		return &IronCrossHint{Action: "call", Reason: "cheapToStay"}
	}
	return &IronCrossHint{Action: "fold", Reason: "notWorthIt"}
}

// --- 永続化 ---

// ironCrossJSON is the JSON wire format for IronCross.
type ironCrossJSON struct {
	Deck        *TrumpCards        `json:"dk"`
	Players     []*IronCrossPlayer `json:"pl"`
	Config      IronCrossConfig    `json:"cf"`
	Phase       int                `json:"ph"`
	Cross       []*Card            `json:"cr"`
	Revealed    int                `json:"rv"`
	Pot         int                `json:"po"`
	CurrentBet  int                `json:"cb"`
	RaiseCount  int                `json:"rc"`
	Turn        int                `json:"tu"`
	ActedFlags  []bool             `json:"af"`
	HandNumber  int                `json:"hn"`
	Results     []IronCrossResult  `json:"rs"`
	GameEndFlag bool               `json:"ge"`
	ActionLog   []*ActionLogEntry  `json:"al"`
	TurnNumber  int                `json:"tn"`
}

// MarshalJSON implements json.Marshaler.
func (g *IronCross) MarshalJSON() ([]byte, error) {
	return json.Marshal(ironCrossJSON{
		Deck: g.deck, Players: g.players, Config: g.config,
		Phase: int(g.phase), Cross: g.cross, Revealed: g.revealed,
		Pot: g.pot, CurrentBet: g.currentBet, RaiseCount: g.raiseCount,
		Turn: g.turn, ActedFlags: g.actedFlags, HandNumber: g.handNumber,
		Results: g.results, GameEndFlag: g.gameEndFlag,
		ActionLog: g.actionLog, TurnNumber: g.turnNumber,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **開いた枚数と、十字に実際に入っている札の数は必ず一致する。**
// ずれた保存を通すと、伏せたままの位置を役の判定に使うことになる。
func (g *IronCross) UnmarshalJSON(data []byte) error {
	var j ironCrossJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := ironCrossValidate(&j); err != nil {
		return err
	}

	g.deck = j.Deck
	if g.deck == nil {
		g.deck = NewTrumpCards(0)
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = IronCrossPhase(j.Phase)
	g.cross = j.Cross
	g.revealed = j.Revealed
	g.pot = j.Pot
	g.currentBet = j.CurrentBet
	g.raiseCount = j.RaiseCount
	g.turn = j.Turn
	g.actedFlags = j.ActedFlags
	g.handNumber = j.HandNumber
	g.results = j.Results
	g.gameEndFlag = j.GameEndFlag
	g.actionLog = j.ActionLog
	g.turnNumber = j.TurnNumber
	return nil
}

// ironCrossValidate は保存データの範囲と整合を検証する。
func ironCrossValidate(j *ironCrossJSON) error {
	if err := j.Config.Validate(); err != nil {
		return err
	}
	seats := len(j.Players)
	if seats != j.Config.Seats {
		return fmt.Errorf("ironcross: %d players for a %d-seat table", seats, j.Config.Seats)
	}
	for i, p := range j.Players {
		if p == nil {
			return fmt.Errorf("ironcross: seat %d is missing", i)
		}
	}
	if j.Phase < int(IronCrossPhaseBetting) || j.Phase > int(IronCrossPhaseMax) {
		return fmt.Errorf("ironcross: phase out of range: %d", j.Phase)
	}
	if j.Turn < 0 || j.Turn >= seats {
		return fmt.Errorf("ironcross: turn seat out of range: %d", j.Turn)
	}
	if j.HandNumber < 1 {
		return fmt.Errorf("ironcross: hand number out of range: %d", j.HandNumber)
	}
	for _, v := range []struct {
		name string
		n    int
	}{{"pot", j.Pot}, {"current bet", j.CurrentBet}, {"raise count", j.RaiseCount}} {
		if v.n < 0 {
			return fmt.Errorf("ironcross: %s must not be negative: %d", v.name, v.n)
		}
	}
	if j.RaiseCount > ironCrossMaxRaisesPerRound {
		return fmt.Errorf("ironcross: raise count exceeds the cap: %d", j.RaiseCount)
	}
	if j.Revealed < 0 || j.Revealed > IronCrossCommunityCards {
		return fmt.Errorf("ironcross: revealed count out of range: %d", j.Revealed)
	}
	if len(j.Cross) != 0 && len(j.Cross) != IronCrossCommunityCards {
		return fmt.Errorf("ironcross: the cross holds %d slots, want %d",
			len(j.Cross), IronCrossCommunityCards)
	}
	// **開いた枚数と実際の札の数が一致すること。**
	if placed := ironCrossPlacedCount(j.Cross); placed != j.Revealed {
		return fmt.Errorf("ironcross: %d cards placed for %d revealed", placed, j.Revealed)
	}
	if len(j.ActedFlags) != 0 && len(j.ActedFlags) != seats {
		return fmt.Errorf("ironcross: %d acted flags for %d seats", len(j.ActedFlags), seats)
	}
	if len(j.Results) != 0 && len(j.Results) != seats {
		return fmt.Errorf("ironcross: %d results for %d seats", len(j.Results), seats)
	}
	if len(j.ActionLog) > ironCrossMaxSliceLen {
		return fmt.Errorf("ironcross: action log too long: %d", len(j.ActionLog))
	}
	return nil
}

// ironCrossPlacedCount は十字に実際に置かれている札の数を返す。
func ironCrossPlacedCount(cross []*Card) int {
	n := 0
	for _, c := range cross {
		if c != nil {
			n++
		}
	}
	return n
}
