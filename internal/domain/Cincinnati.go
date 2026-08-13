//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// CincinnatiPhase はゲームの進行段階。
//
// **コミュニティを 1 枚めくるたびに 1 段階。** Holdem の 3-1-1 と違い、
// 5 回とも同じ形で進む。
type CincinnatiPhase int

const (
	// CincinnatiPhaseDeal は配る前。
	CincinnatiPhaseDeal CincinnatiPhase = iota
	// CincinnatiPhaseBetting はベットラウンド中。
	CincinnatiPhaseBetting
	// CincinnatiPhaseShowdown はショーダウン。
	CincinnatiPhaseShowdown
	// CincinnatiPhaseGameEnd はゲーム終了。
	CincinnatiPhaseGameEnd
)

// CincinnatiPhaseMax は最大のフェーズ値 (復元時の範囲検査に使う)。
const CincinnatiPhaseMax = CincinnatiPhaseGameEnd

// アクション定数 (共通定数のエイリアス)。
const (
	CincinnatiActionFold  = bettingActionFold
	CincinnatiActionCheck = bettingActionCheck
	CincinnatiActionCall  = bettingActionCall
	CincinnatiActionBet   = bettingActionBet
	CincinnatiActionRaise = bettingActionRaise
)

// cincinnatiMaxSliceLen は復元時に許すスライス長の上限。
const cincinnatiMaxSliceLen = 512

// cincinnatiMaxCpuSteps は CPU を進める 1 回あたりの上限。
const cincinnatiMaxCpuSteps = 128

// cincinnatiMaxRaisesPerRound は 1 ラウンドのレイズ上限。
//
// **上限が無いと終わらない。** 2 人が交互にレイズし続けられる規則にすると、
// チップが尽きるまでラウンドが閉じない経路ができる。
const cincinnatiMaxRaisesPerRound = 3

// エラー値。
var (
	errCincinnatiFinished    = errors.New("cincinnati: game already finished")
	errCincinnatiWrongPhase  = errors.New("cincinnati: not allowed in this phase")
	errCincinnatiNotYourRun  = errors.New("cincinnati: not your turn")
	errCincinnatiBadAction   = errors.New("cincinnati: unknown action")
	errCincinnatiCannotCheck = errors.New("cincinnati: cannot check facing a bet")
	errCincinnatiCannotBet   = errors.New("cincinnati: cannot bet facing a bet")
	errCincinnatiCannotRaise = errors.New("cincinnati: nothing to raise")
	errCincinnatiRaiseCapped = errors.New("cincinnati: the raise cap for this round is reached")
	errCincinnatiBetRange    = errors.New("cincinnati: bet out of range")
)

// Cincinnati はシンシナティの卓。
//
// **手札 5 枚 + コミュニティ 5 枚の 10 枚から最良の 5 枚を選ぶ。** 手札が
// 多いぶん、コミュニティを 1 枚も使わない役が普通に成立する ── そこが
// Holdem (ホールカード 2 枚) との一番の違いになる。
//
// **コミュニティは 1 枚ずつ 5 回めくり、そのたびにベットラウンドが入る。**
// 情報が少しずつ増えるので、Holdem の 3 枚一括より降りどころが多い。
type Cincinnati struct {
	deck    *TrumpCards
	players []*CincinnatiPlayer
	config  CincinnatiConfig

	phase CincinnatiPhase
	// community は表向きになったコミュニティ。**まだ伏せている札は入らない。**
	community []*Card
	// revealed は公開済みの枚数 (= 何回目のベットラウンドか)。
	revealed int

	pot         int
	currentBet  int
	raiseCount  int
	turn        int
	lastAggr    int
	actedFlags  []bool
	handNumber  int
	results     []CincinnatiResult
	gameEndFlag bool
	actionLog   []*ActionLogEntry
	turnNumber  int
}

// CincinnatiResult は 1 席のハンド結果。
type CincinnatiResult struct {
	// PlayerIdx は席番号。
	PlayerIdx int
	// HandRank は役のランク。
	HandRank int
	// WonAmount は獲得額。
	WonAmount int
}

// NewCincinnati は指定の山・席・設定で卓を構築する。
func NewCincinnati(deck *TrumpCards, players []*CincinnatiPlayer, config CincinnatiConfig) *Cincinnati {
	return &Cincinnati{
		deck: deck, players: players, config: config,
		community:  make([]*Card, 0, CincinnatiCommunityCards),
		actedFlags: make([]bool, len(players)),
		handNumber: 1,
	}
}

// NewDefaultCincinnati は既定の卓を構築する。席 0 が人間。
func NewDefaultCincinnati() *Cincinnati {
	cfg := DefaultCincinnatiConfig()
	return NewCincinnati(NewTrumpCards(0),
		NewCincinnatiPlayersForTable(cfg.Seats, cfg.InitialChips), cfg)
}

// Reset はゲームを初期化する。
func (g *Cincinnati) Reset() {
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
func (g *Cincinnati) startHand() {
	g.deck.Replenish()
	g.deck.Shuffle()
	g.community = g.community[:0]
	g.revealed = 0
	g.pot = 0
	g.currentBet = 0
	g.raiseCount = 0
	g.results = nil
	g.actedFlags = make([]bool, len(g.players))

	for _, p := range g.players {
		p.ResetForHand()
	}
	// **アンティは全員から。** 参加費が無いと降り続けるのが最適になる。
	for _, p := range g.players {
		ante := min(g.config.Ante, p.GetChips())
		p.SubtractChips(ante)
		g.pot += ante
		if p.GetChips() == 0 {
			p.SetAllIn(true)
		}
	}
	for range CincinnatiHoleCards {
		for _, p := range g.players {
			p.AddCard(g.deck.DrawCard())
		}
	}

	g.phase = CincinnatiPhaseBetting
	g.turn = g.firstActiveSeat()
	g.lastAggr = g.turn
	g.appendLog(-1, "deal", fmt.Sprintf("hand %d, ante %d", g.handNumber, g.config.Ante), nil)
	g.advanceCpu()
}

// --- ベッティング ---

// PlayerAction は人間の手を処理する。
func (g *Cincinnati) PlayerAction(action, amount int) error {
	if g.gameEndFlag {
		return errCincinnatiFinished
	}
	if g.phase != CincinnatiPhaseBetting {
		return errCincinnatiWrongPhase
	}
	if g.turn != g.HumanSeat() {
		return errCincinnatiNotYourRun
	}
	if err := g.applyAction(g.turn, action, amount); err != nil {
		return err
	}
	g.advanceAfterAction()
	g.advanceCpu()
	return nil
}

// applyAction は席 i の手を適用する。
func (g *Cincinnati) applyAction(i, action, amount int) error {
	p := g.players[i]
	toCall := g.currentBet - p.GetCurrentBet()

	switch action {
	case CincinnatiActionFold:
		p.SetFolded(true)
		g.appendLog(i, "fold", fmt.Sprintf("seat %d folds", i), nil)
	case CincinnatiActionCheck:
		if toCall > 0 {
			return errCincinnatiCannotCheck
		}
		g.appendLog(i, "check", fmt.Sprintf("seat %d checks", i), nil)
	case CincinnatiActionCall:
		g.moveToPot(p, toCall)
		g.appendLog(i, "call", fmt.Sprintf("seat %d calls %d", i, toCall), nil)
	case CincinnatiActionBet:
		if g.currentBet > 0 {
			return errCincinnatiCannotBet
		}
		if amount < g.config.Ante || amount > p.GetChips() {
			return errCincinnatiBetRange
		}
		g.moveToPot(p, amount)
		g.currentBet = p.GetCurrentBet()
		g.raiseCount++
		g.lastAggr = i
		g.appendLog(i, "bet", fmt.Sprintf("seat %d bets %d", i, amount), nil)
	case CincinnatiActionRaise:
		if g.currentBet == 0 {
			return errCincinnatiCannotRaise
		}
		if g.raiseCount >= cincinnatiMaxRaisesPerRound {
			return errCincinnatiRaiseCapped
		}
		if amount < g.config.Ante || amount > p.GetChips()-toCall {
			return errCincinnatiBetRange
		}
		g.moveToPot(p, toCall+amount)
		g.currentBet = p.GetCurrentBet()
		g.raiseCount++
		g.lastAggr = i
		g.appendLog(i, "raise", fmt.Sprintf("seat %d raises %d", i, amount), nil)
	default:
		return errCincinnatiBadAction
	}
	g.actedFlags[i] = true
	return nil
}

// moveToPot は席からポットへチップを移す。**持ち分を超えては動かさない。**
func (g *Cincinnati) moveToPot(p *CincinnatiPlayer, amount int) {
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
func (g *Cincinnati) advanceAfterAction() {
	if g.activePlayers() <= 1 {
		g.showdown()
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
// レイズに応じていない席を残したまま次の札をめくってしまう。
func (g *Cincinnati) bettingRoundComplete() bool {
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

// nextStage は次の 1 枚をめくるか、ショーダウンへ進む。
func (g *Cincinnati) nextStage() {
	if g.revealed >= CincinnatiCommunityCards {
		g.showdown()
		return
	}
	g.community = append(g.community, g.deck.DrawCard())
	g.revealed++
	g.currentBet = 0
	g.raiseCount = 0
	g.actedFlags = make([]bool, len(g.players))
	for _, p := range g.players {
		p.SetCurrentBet(0)
	}
	g.turn = g.firstActiveSeat()
	g.lastAggr = g.turn
	g.appendLog(-1, "reveal", fmt.Sprintf("community card %d", g.revealed),
		[]*Card{g.community[len(g.community)-1]})

	// **公開しきってもベットラウンドは 1 回入る。** 5 枚目の後に賭けずに
	// 決着すると、最後の情報にだけ賭けられないことになる。
	if g.activePlayers() <= 1 {
		g.showdown()
	}
}

// showdown は役を比べてポットを配る。
func (g *Cincinnati) showdown() {
	g.phase = CincinnatiPhaseShowdown
	g.results = make([]CincinnatiResult, 0, len(g.players))

	bestRank, winners := -1, []int(nil)
	for i, p := range g.players {
		if p.GetFolded() {
			g.results = append(g.results, CincinnatiResult{PlayerIdx: i, HandRank: -1})
			continue
		}
		rank := p.EvaluateBest(g.community)
		g.results = append(g.results, CincinnatiResult{PlayerIdx: i, HandRank: rank})
		switch {
		case rank > bestRank:
			bestRank, winners = rank, []int{i}
		case rank == bestRank:
			winners = append(winners, i)
		}
	}

	// **端数は若い席から配る。** 割り切れない分をポットに残すと、そのぶんが
	// 卓から消える (総量が合わなくなる)。
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
	g.appendLog(-1, "showdown", fmt.Sprintf("hand %d settled", g.handNumber), g.community)
}

// NextHand は次のハンドを始める。
func (g *Cincinnati) NextHand() error {
	if g.gameEndFlag {
		return errCincinnatiFinished
	}
	if g.phase != CincinnatiPhaseShowdown {
		return errCincinnatiWrongPhase
	}
	if g.aliveSeats() < CincinnatiMinSeats || g.players[g.HumanSeat()].GetChips() <= 0 {
		g.finish()
		return nil
	}
	g.handNumber++
	g.startHand()
	return nil
}

// finish はゲームを終える。
func (g *Cincinnati) finish() {
	g.gameEndFlag = true
	g.phase = CincinnatiPhaseGameEnd
	g.appendLog(-1, "gameEnd", fmt.Sprintf("winner seat %d", g.WinnerSeat()), nil)
}

// --- CPU ---

// CpuPlay は CPU の席を進める。
func (g *Cincinnati) CpuPlay() { g.advanceCpu() }

// advanceCpu は人間の手番になるかハンドが終わるまで CPU を進める。
func (g *Cincinnati) advanceCpu() {
	for range cincinnatiMaxCpuSteps {
		if g.gameEndFlag || g.phase != CincinnatiPhaseBetting || g.IsHumanTurn() {
			return
		}
		i := g.turn
		action, amount := g.cpuDecide(i)
		if err := g.applyAction(i, action, amount); err != nil {
			// **拒まれたら降りる。** 判断を誤っても盤面は必ず進む。
			_ = g.applyAction(i, CincinnatiActionFold, 0)
		}
		g.advanceAfterAction()
	}
}

// cpuDecide は CPU の手を決める。
//
// **見えている 10 枚での最良役だけで決める。** 相手の手札は伏せたままなので、
// ここで覗くと勝てない相手になる。
func (g *Cincinnati) cpuDecide(i int) (int, int) {
	p := g.players[i]
	rank := p.EvaluateBest(g.community)
	toCall := g.currentBet - p.GetCurrentBet()

	if toCall <= 0 {
		if rank >= PokerHandTwoPair && g.raiseCount < cincinnatiMaxRaisesPerRound {
			return CincinnatiActionBet, g.config.Ante
		}
		return CincinnatiActionCheck, 0
	}
	if rank >= PokerHandThreeOfAKind && g.raiseCount < cincinnatiMaxRaisesPerRound {
		return CincinnatiActionRaise, g.config.Ante
	}
	if rank >= PokerHandOnePair || toCall <= g.config.Ante {
		return CincinnatiActionCall, 0
	}
	return CincinnatiActionFold, 0
}

// --- 席の走査 ---

// firstActiveSeat は最初に動く席を返す。
func (g *Cincinnati) firstActiveSeat() int {
	for i, p := range g.players {
		if !p.GetFolded() && !p.GetAllIn() {
			return i
		}
	}
	return 0
}

// nextActiveSeat は i の次に動ける席を返す。
func (g *Cincinnati) nextActiveSeat(from int) int {
	for step := 1; step <= len(g.players); step++ {
		idx := (from + step) % len(g.players)
		p := g.players[idx]
		if !p.GetFolded() && !p.GetAllIn() {
			return idx
		}
	}
	return from
}

// activePlayers は降りていない席の数を返す。
func (g *Cincinnati) activePlayers() int {
	n := 0
	for _, p := range g.players {
		if !p.GetFolded() {
			n++
		}
	}
	return n
}

// aliveSeats はチップの残っている席の数を返す。
func (g *Cincinnati) aliveSeats() int {
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
func (g *Cincinnati) HumanSeat() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return 0
}

// IsHumanTurn は人間の操作待ちかを返す。
func (g *Cincinnati) IsHumanTurn() bool {
	return g.phase == CincinnatiPhaseBetting && g.turn == g.HumanSeat()
}

// WinnerSeat はチップがいちばん多い席を返す。同点なら若い席。
func (g *Cincinnati) WinnerSeat() int {
	best, bestChips := 0, -1
	for i, p := range g.players {
		if p.GetChips() > bestChips {
			best, bestChips = i, p.GetChips()
		}
	}
	return best
}

// GetConfig はゲーム設定を返す。
func (g *Cincinnati) GetConfig() CincinnatiConfig { return g.config }

// SetConfig はゲーム設定を設定する。
func (g *Cincinnati) SetConfig(c CincinnatiConfig) { g.config = c }

// GetPhase は現在のフェーズを返す。
func (g *Cincinnati) GetPhase() CincinnatiPhase { return g.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *Cincinnati) GetGameEndFlag() bool { return g.gameEndFlag }

// GetPlayers は席の一覧を返す。
func (g *Cincinnati) GetPlayers() []*CincinnatiPlayer { return g.players }

// GetCommunityCards は表向きのコミュニティを返す。
func (g *Cincinnati) GetCommunityCards() []*Card { return g.community }

// GetRevealedCount は公開済みのコミュニティ枚数を返す。
func (g *Cincinnati) GetRevealedCount() int { return g.revealed }

// GetPot はポットを返す。
func (g *Cincinnati) GetPot() int { return g.pot }

// GetCurrentBet はこのラウンドの現在の賭け額を返す。
func (g *Cincinnati) GetCurrentBet() int { return g.currentBet }

// GetToCall は人間が現在コールに要する額を返す。
func (g *Cincinnati) GetToCall() int {
	p := g.players[g.HumanSeat()]
	return max(0, g.currentBet-p.GetCurrentBet())
}

// GetRaiseCount はこのラウンドのレイズ回数を返す。
func (g *Cincinnati) GetRaiseCount() int { return g.raiseCount }

// CanRaise はいまレイズできるかを返す。
func (g *Cincinnati) CanRaise() bool {
	return g.currentBet > 0 && g.raiseCount < cincinnatiMaxRaisesPerRound
}

// GetTurnSeat はいまの手番を返す。
func (g *Cincinnati) GetTurnSeat() int { return g.turn }

// GetHandNumber はハンド数を返す。
func (g *Cincinnati) GetHandNumber() int { return g.handNumber }

// GetResults はハンドの結果を返す。
func (g *Cincinnati) GetResults() []CincinnatiResult { return g.results }

// GetRemainingCards は山の残り枚数を返す。
func (g *Cincinnati) GetRemainingCards() int { return g.deck.GetRemainingCount() }

// GetActionLog は棋譜を返す。
func (g *Cincinnati) GetActionLog() []*ActionLogEntry { return g.actionLog }

// appendLog は棋譜に 1 行足す。
func (g *Cincinnati) appendLog(seat int, actionType, detail string, cards []*Card) {
	g.turnNumber++
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.turnNumber,
		PlayerIdx:  seat,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
	if len(g.actionLog) > cincinnatiMaxSliceLen {
		g.actionLog = g.actionLog[len(g.actionLog)-cincinnatiMaxSliceLen:]
	}
}

// --- 永続化 ---

// cincinnatiJSON is the JSON wire format for Cincinnati.
type cincinnatiJSON struct {
	Deck        *TrumpCards         `json:"dk"`
	Players     []*CincinnatiPlayer `json:"pl"`
	Config      CincinnatiConfig    `json:"cf"`
	Phase       int                 `json:"ph"`
	Community   []*Card             `json:"cm"`
	Revealed    int                 `json:"rv"`
	Pot         int                 `json:"po"`
	CurrentBet  int                 `json:"cb"`
	RaiseCount  int                 `json:"rc"`
	Turn        int                 `json:"tu"`
	ActedFlags  []bool              `json:"af"`
	HandNumber  int                 `json:"hn"`
	Results     []CincinnatiResult  `json:"rs"`
	GameEndFlag bool                `json:"ge"`
	ActionLog   []*ActionLogEntry   `json:"al"`
	TurnNumber  int                 `json:"tn"`
}

// MarshalJSON implements json.Marshaler.
func (g *Cincinnati) MarshalJSON() ([]byte, error) {
	return json.Marshal(cincinnatiJSON{
		Deck: g.deck, Players: g.players, Config: g.config,
		Phase: int(g.phase), Community: g.community, Revealed: g.revealed,
		Pot: g.pot, CurrentBet: g.currentBet, RaiseCount: g.raiseCount,
		Turn: g.turn, ActedFlags: g.actedFlags, HandNumber: g.handNumber,
		Results: g.results, GameEndFlag: g.gameEndFlag,
		ActionLog: g.actionLog, TurnNumber: g.turnNumber,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **公開枚数とコミュニティの実枚数は必ず一致する。** ここがずれた保存を通すと、
// 「まだ伏せている札」を役の判定に使ってしまう ── 添字としては正当なので
// 範囲チェックでは捕まらない。
func (g *Cincinnati) UnmarshalJSON(data []byte) error {
	var j cincinnatiJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := cincinnatiValidate(&j); err != nil {
		return err
	}

	g.deck = j.Deck
	if g.deck == nil {
		g.deck = NewTrumpCards(0)
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = CincinnatiPhase(j.Phase)
	g.community = j.Community
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

// cincinnatiValidate は保存データの範囲と整合を検証する。
func cincinnatiValidate(j *cincinnatiJSON) error {
	if err := j.Config.Validate(); err != nil {
		return err
	}
	seats := len(j.Players)
	if seats != j.Config.Seats {
		return fmt.Errorf("cincinnati: %d players for a %d-seat table", seats, j.Config.Seats)
	}
	for i, p := range j.Players {
		if p == nil {
			return fmt.Errorf("cincinnati: seat %d is missing", i)
		}
	}
	if j.Phase < int(CincinnatiPhaseDeal) || j.Phase > int(CincinnatiPhaseMax) {
		return fmt.Errorf("cincinnati: phase out of range: %d", j.Phase)
	}
	if j.Turn < 0 || j.Turn >= seats {
		return fmt.Errorf("cincinnati: turn seat out of range: %d", j.Turn)
	}
	if j.HandNumber < 1 {
		return fmt.Errorf("cincinnati: hand number out of range: %d", j.HandNumber)
	}
	for _, v := range []struct {
		name string
		n    int
	}{{"pot", j.Pot}, {"current bet", j.CurrentBet}, {"raise count", j.RaiseCount}} {
		if v.n < 0 {
			return fmt.Errorf("cincinnati: %s must not be negative: %d", v.name, v.n)
		}
	}
	if j.RaiseCount > cincinnatiMaxRaisesPerRound {
		return fmt.Errorf("cincinnati: raise count exceeds the cap: %d", j.RaiseCount)
	}
	// **公開枚数とコミュニティの実枚数が一致すること。**
	if j.Revealed < 0 || j.Revealed > CincinnatiCommunityCards {
		return fmt.Errorf("cincinnati: revealed count out of range: %d", j.Revealed)
	}
	if len(j.Community) != j.Revealed {
		return fmt.Errorf("cincinnati: %d community cards for %d revealed",
			len(j.Community), j.Revealed)
	}
	if len(j.ActedFlags) != 0 && len(j.ActedFlags) != seats {
		return fmt.Errorf("cincinnati: %d acted flags for %d seats", len(j.ActedFlags), seats)
	}
	if len(j.Results) != 0 && len(j.Results) != seats {
		return fmt.Errorf("cincinnati: %d results for %d seats", len(j.Results), seats)
	}
	if len(j.ActionLog) > cincinnatiMaxSliceLen {
		return fmt.Errorf("cincinnati: action log too long: %d", len(j.ActionLog))
	}
	return nil
}
