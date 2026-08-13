//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"math/rand"
)

// TuSacPhase は進行の段階。
type TuSacPhase int

const (
	// TuSacPhaseDraw は引くのを待つ場面。
	TuSacPhaseDraw TuSacPhase = iota
	// TuSacPhaseDiscard は捨てるのを待つ場面 (メルドはこの間に出せる)。
	TuSacPhaseDiscard
	// TuSacPhaseRoundEnd はラウンドの決着。
	TuSacPhaseRoundEnd
	// TuSacPhaseGameEnd は終局。
	TuSacPhaseGameEnd
)

// TuSacPhaseMax は最大のフェーズ値 (復元時の範囲検査に使う)。
const TuSacPhaseMax = TuSacPhaseGameEnd

// tuSacMaxSliceLen は復元時に受け付けるスライスの上限。
const tuSacMaxSliceLen = 512

// tuSacMaxCpuSteps は CPU を進める 1 回あたりの上限。
const tuSacMaxCpuSteps = 512

// エラー値。
var (
	errTuSacFinished     = errors.New("tusac: the game is over")
	errTuSacWrongPhase   = errors.New("tusac: not allowed in this phase")
	errTuSacNotYourTurn  = errors.New("tusac: it is not your turn")
	errTuSacStockEmpty   = errors.New("tusac: the stock is empty")
	errTuSacNoDiscard    = errors.New("tusac: the discard pile is empty")
	errTuSacBadCardIndex = errors.New("tusac: card index out of range")
	errTuSacNotAMeld     = errors.New("tusac: those cards are not a valid combination")
	errTuSacRoundRange   = errors.New("tusac: round number out of range")
	errTuSacTurnRange    = errors.New("tusac: turn seat out of range")
	errTuSacPhaseRange   = errors.New("tusac: phase out of range")
	errTuSacSliceTooLong = errors.New("tusac: slice too long")
	errTuSacSeatCount    = errors.New("tusac: seat count does not match the config")
	errTuSacDeckTooLarge = errors.New("tusac: the stock holds more cards than the deck")
)

// TuSacResult は 1 席のラウンド結果。
type TuSacResult struct {
	PlayerIdx int
	// MeldPoints は場に出した組み合わせの合計。
	MeldPoints int
	// HandPenalty は手に残った枚数ぶんの失点。
	HandPenalty int
	// RoundScore は差し引き。
	RoundScore int
	// WentOut は上がった席か。
	WentOut bool
}

// TuSac は四色牌 (Tứ Sắc) の卓。
//
// **112 枚の四色牌で組み合わせを作るラミー。** 引いて捨てるという手番の形は
// 標準デッキのラミーと同じだが、**数字の並びという概念が無い**ので「同スートの
// 連番」に当たるメルドが存在しない ── 作れるのは同色同種 3 枚、異色の車馬砲、
// 卒 5 枚の 3 つだけ。
//
// **ラウンドは山が尽きれば必ず終わる。** 上がりは手札 21 枚をちょうど
// 使い切る (3a + 5b = 20) 必要があって滅多に出ないので、終了性は上がりでは
// なく山の枯渇が担保する。
type TuSac struct {
	stock   []*Card
	discard []*Card
	players []*TuSacPlayer
	config  TuSacConfig

	phase       TuSacPhase
	turn        int
	roundNumber int
	// wentOut は上がった席 (-1 なら山切れで終わった)。
	wentOut int

	results     []TuSacResult
	gameEndFlag bool

	actionLog  []*ActionLogEntry
	turnNumber int
}

// NewTuSac は TuSac を構築する。
func NewTuSac(players []*TuSacPlayer, config TuSacConfig) *TuSac {
	return &TuSac{
		players:   players,
		config:    config,
		wentOut:   -1,
		results:   make([]TuSacResult, 0, len(players)),
		actionLog: make([]*ActionLogEntry, 0),
	}
}

// NewDefaultTuSac は既定設定の卓を返す。
func NewDefaultTuSac() *TuSac {
	cfg := DefaultTuSacConfig()
	return NewTuSac(NewTuSacPlayersForTable(cfg.Seats), cfg)
}

// Reset はゲームを最初から始める。
func (g *TuSac) Reset() {
	if err := g.config.Validate(); err != nil {
		g.config = DefaultTuSacConfig()
	}
	g.players = NewTuSacPlayersForTable(g.config.Seats)
	g.roundNumber = 0
	g.gameEndFlag = false
	g.actionLog = g.actionLog[:0]
	g.turnNumber = 0
	g.startRound()
}

// startRound は次のラウンドを配る。
func (g *TuSac) startRound() {
	deck := buildTuSacDeck()
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })

	g.roundNumber++
	g.phase = TuSacPhaseDraw
	g.turn = 0
	g.wentOut = -1
	g.results = g.results[:0]
	for _, p := range g.players {
		p.ResetForRound()
	}

	for range TuSacHandSize {
		for _, p := range g.players {
			p.AddCard(deck[0])
			deck = deck[1:]
		}
	}
	// **捨て札は 1 枚めくって始める。** 最初の席にも「捨て札から引く」という
	// 選択肢を与えるため ── 空だと最初の 1 手だけ選択肢が 1 つ少ない。
	g.discard = []*Card{deck[0]}
	g.stock = deck[1:]
	g.appendLog(-1, "deal", "dealt the hands", nil)
	g.runCpuTurns()
}

// Draw は山または捨て札から 1 枚引く。fromDiscard が真なら捨て札の一番上。
func (g *TuSac) Draw(fromDiscard bool) error {
	if g.gameEndFlag {
		return errTuSacFinished
	}
	if g.phase != TuSacPhaseDraw {
		return errTuSacWrongPhase
	}
	if !g.IsHumanTurn() {
		return errTuSacNotYourTurn
	}
	if err := g.drawFor(g.turn, fromDiscard); err != nil {
		return err
	}
	g.phase = TuSacPhaseDiscard
	return nil
}

// drawFor は席 i に 1 枚引かせる。
func (g *TuSac) drawFor(i int, fromDiscard bool) error {
	p := g.players[i]
	if fromDiscard {
		if len(g.discard) == 0 {
			return errTuSacNoDiscard
		}
		top := g.discard[len(g.discard)-1]
		g.discard = g.discard[:len(g.discard)-1]
		p.AddCard(top)
		g.appendLog(i, "draw", "from the discard", []*Card{top})
		return nil
	}
	if len(g.stock) == 0 {
		return errTuSacStockEmpty
	}
	c := g.stock[0]
	g.stock = g.stock[1:]
	p.AddCard(c)
	g.appendLog(i, "draw", "from the stock", nil)
	return nil
}

// Meld は手札の指定の札を組み合わせとして場に出す。
//
// **出したら戻せない。** 場に出た札は他の組み合わせに使えなくなるので、
// 出す判断そのものがこのゲームの読みどころになる。
func (g *TuSac) Meld(indexes []int) error {
	if g.gameEndFlag {
		return errTuSacFinished
	}
	if g.phase != TuSacPhaseDiscard {
		return errTuSacWrongPhase
	}
	if !g.IsHumanTurn() {
		return errTuSacNotYourTurn
	}
	return g.meldFor(g.turn, indexes)
}

// meldFor は席 i にメルドを出させる。
func (g *TuSac) meldFor(i int, indexes []int) error {
	p := g.players[i]
	picked, kind := TuSacFindMeld(p.GetCards(), indexes)
	if kind == TuSacMeldNone {
		return errTuSacNotAMeld
	}
	p.AddMeld(kind, picked)
	p.RemoveCardsAt(indexes)
	g.appendLog(i, "meld", TuSacMeldKindName(kind), picked)
	return nil
}

// Discard は手札から 1 枚捨てて手番を渡す。
func (g *TuSac) Discard(index int) error {
	if g.gameEndFlag {
		return errTuSacFinished
	}
	if g.phase != TuSacPhaseDiscard {
		return errTuSacWrongPhase
	}
	if !g.IsHumanTurn() {
		return errTuSacNotYourTurn
	}
	if err := g.discardFor(g.turn, index); err != nil {
		return err
	}
	g.advanceTurn()
	return nil
}

// discardFor は席 i に 1 枚捨てさせる。
func (g *TuSac) discardFor(i, index int) error {
	p := g.players[i]
	if index < 0 || index >= len(p.GetCards()) {
		return errTuSacBadCardIndex
	}
	c := p.GetCards()[index]
	p.RemoveCardsAt([]int{index})
	g.discard = append(g.discard, c)
	g.appendLog(i, "discard", "", []*Card{c})
	return nil
}

// advanceTurn は次の席に手番を渡し、終了条件を見る。
func (g *TuSac) advanceTurn() {
	// **手札を使い切った席が上がり。** 3a + 5b = 20 をちょうど満たす必要が
	// あるので滅多に出ないが、出たらそのラウンドは即終わる。
	if len(g.players[g.turn].GetCards()) == 0 {
		g.wentOut = g.turn
		g.finishRound()
		return
	}
	g.turn = (g.turn + 1) % len(g.players)
	g.phase = TuSacPhaseDraw
	g.runCpuTurns()
}

// runCpuTurns は人間の手番か決着まで CPU を進める。
func (g *TuSac) runCpuTurns() {
	for range tuSacMaxCpuSteps {
		if g.gameEndFlag || g.phase == TuSacPhaseRoundEnd {
			return
		}
		// **山を見るのは「人間の手番か」より先。** 後にすると、山が尽きた
		// 瞬間が人間の手番だったときにここで戻ってしまい、引けない `Draw` を
		// 待つ盤面で止まる ── ラウンドを閉じる判断はこの 1 か所に集める。
		if len(g.stock) == 0 {
			g.finishRound()
			return
		}
		if g.IsHumanTurn() {
			return
		}
		i := g.turn
		// **CPU は山から引く。** 捨て札を拾う判断は組み合わせの読みが要るので、
		// 場当たりに拾わせると人間より弱くなるどころか進行が濁る。
		if err := g.drawFor(i, false); err != nil {
			g.finishRound()
			return
		}
		g.phase = TuSacPhaseDiscard
		g.cpuMeld(i)
		_ = g.discardFor(i, g.cpuDiscardIndex(i))
		if len(g.players[i].GetCards()) == 0 {
			g.wentOut = i
			g.finishRound()
			return
		}
		g.turn = (g.turn + 1) % len(g.players)
		g.phase = TuSacPhaseDraw
	}
}

// cpuMeld は出せる組み合わせをすべて出す。
func (g *TuSac) cpuMeld(i int) {
	for range TuSacHandSize {
		indexes := g.findAnyMeld(g.players[i].GetCards())
		if indexes == nil {
			return
		}
		_ = g.meldFor(i, indexes)
	}
}

// findAnyMeld は手札から出せる組み合わせを 1 つ探す (無ければ nil)。
//
// **大きいメルドから探す。** 卒 5 枚は同色同種より点が高いので、先に 3 枚の
// 組を出してしまうと、卒がばらけて 5 枚に届かなくなることがある。
func (g *TuSac) findAnyMeld(hand []*Card) []int {
	// 卒 5 枚。
	soldiers := make([]int, 0, TuSacSoldierSetSize)
	for i, c := range hand {
		if c != nil && c.GetValue() == TuSacPieceSoldier {
			soldiers = append(soldiers, i)
			if len(soldiers) == TuSacSoldierSetSize {
				return soldiers
			}
		}
	}

	// 同色同種 3 枚。
	byKey := map[[2]int][]int{}
	for i, c := range hand {
		if c == nil {
			continue
		}
		k := [2]int{c.GetDesign(), c.GetValue()}
		byKey[k] = append(byKey[k], i)
		if len(byKey[k]) == TuSacSetSize {
			return byKey[k]
		}
	}

	// 異色の車馬砲。
	byPiece := map[int][]int{}
	for i, c := range hand {
		if c != nil && TuSacIsChariotHorseCannon(c.GetValue()) {
			byPiece[c.GetValue()] = append(byPiece[c.GetValue()], i)
		}
	}
	for _, ci := range byPiece[TuSacPieceChariot] {
		for _, hi := range byPiece[TuSacPieceHorse] {
			for _, ni := range byPiece[TuSacPieceCannon] {
				idx := []int{ci, hi, ni}
				if _, kind := TuSacFindMeld(hand, idx); kind == TuSacMeldChariotTrio {
					return idx
				}
			}
		}
	}
	return nil
}

// cpuDiscardIndex は捨てる札を選ぶ。
//
// **組み合わせに近い札を残す。** 同じ色・同じ駒の仲間が少ない札から捨てる。
func (g *TuSac) cpuDiscardIndex(i int) int {
	hand := g.players[i].GetCards()
	if len(hand) == 0 {
		return 0
	}
	counts := map[[2]int]int{}
	for _, c := range hand {
		if c != nil {
			counts[[2]int{c.GetDesign(), c.GetValue()}]++
		}
	}
	worst, worstCount := 0, 1<<30
	for k, c := range hand {
		if c == nil {
			continue
		}
		n := counts[[2]int{c.GetDesign(), c.GetValue()}]
		if c.GetValue() == TuSacPieceSoldier {
			n += TuSacSoldierSetSize // 卒は 5 枚組があるので残しやすくする
		}
		if n < worstCount {
			worst, worstCount = k, n
		}
	}
	return worst
}

// finishRound はラウンドを締めて得点をつける。
func (g *TuSac) finishRound() {
	g.phase = TuSacPhaseRoundEnd
	g.results = g.results[:0]
	for i, p := range g.players {
		meld := p.MeldPoints()
		penalty := len(p.GetCards())
		round := meld - penalty
		p.SetRoundScore(round)
		p.AddScore(round)
		g.results = append(g.results, TuSacResult{
			PlayerIdx: i, MeldPoints: meld, HandPenalty: penalty,
			RoundScore: round, WentOut: i == g.wentOut,
		})
	}
	g.appendLog(-1, "round", "round over", nil)

	if g.roundNumber >= g.config.Rounds {
		g.finish()
	}
}

// NextRound は次のラウンドを始める。
func (g *TuSac) NextRound() error {
	if g.gameEndFlag {
		return errTuSacFinished
	}
	if g.phase != TuSacPhaseRoundEnd {
		return errTuSacWrongPhase
	}
	g.startRound()
	return nil
}

// finish は終局にする。
func (g *TuSac) finish() {
	g.gameEndFlag = true
	g.phase = TuSacPhaseGameEnd
}

// HumanSeat は人間の席を返す。
func (g *TuSac) HumanSeat() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return 0
}

// IsHumanTurn は人間の手番かを返す。
func (g *TuSac) IsHumanTurn() bool {
	return !g.gameEndFlag && g.phase != TuSacPhaseRoundEnd && g.turn == g.HumanSeat()
}

// WinnerSeat は得点がいちばん高い席を返す。
func (g *TuSac) WinnerSeat() int {
	best, seat := -1<<30, 0
	for i, p := range g.players {
		if p.GetScore() > best {
			best, seat = p.GetScore(), i
		}
	}
	return seat
}

// GetConfig は設定を返す。
func (g *TuSac) GetConfig() TuSacConfig { return g.config }

// SetConfig は設定を差し替える。
func (g *TuSac) SetConfig(c TuSacConfig) { g.config = c }

// GetPhase は現在のフェーズを返す。
func (g *TuSac) GetPhase() TuSacPhase { return g.phase }

// GetGameEndFlag は終局かを返す。
func (g *TuSac) GetGameEndFlag() bool { return g.gameEndFlag }

// GetPlayers は席の一覧を返す。
func (g *TuSac) GetPlayers() []*TuSacPlayer { return g.players }

// GetTurnSeat は手番の席を返す。
func (g *TuSac) GetTurnSeat() int { return g.turn }

// GetRoundNumber は何ラウンド目かを返す。
func (g *TuSac) GetRoundNumber() int { return g.roundNumber }

// GetStockCount は山の残り枚数を返す。
func (g *TuSac) GetStockCount() int { return len(g.stock) }

// GetDiscardTop は捨て札の一番上を返す (無ければ nil)。
func (g *TuSac) GetDiscardTop() *Card {
	if len(g.discard) == 0 {
		return nil
	}
	return g.discard[len(g.discard)-1]
}

// GetDiscardCount は捨て札の枚数を返す。
func (g *TuSac) GetDiscardCount() int { return len(g.discard) }

// GetWentOutSeat は上がった席を返す (-1 なら山切れ)。
func (g *TuSac) GetWentOutSeat() int { return g.wentOut }

// GetResults は直近のラウンド結果を返す。
func (g *TuSac) GetResults() []TuSacResult { return g.results }

// GetActionLog は棋譜を返す。
func (g *TuSac) GetActionLog() []*ActionLogEntry { return g.actionLog }

// appendLog は棋譜に 1 行足す。
func (g *TuSac) appendLog(seat int, actionType, detail string, cards []*Card) {
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

// TuSacHint は人間への助言。
type TuSacHint struct {
	// Action は薦める操作 ("draw" / "meld" / "discard" / "next")。
	Action string
	// Indexes は薦める札の位置 (Action が "meld" / "discard" のとき)。
	Indexes []int
	// Reason は理由の識別子 (i18n キーの一部)。
	Reason string
}

// GetHint は人間への助言を返す。
func (g *TuSac) GetHint() *TuSacHint {
	if g.gameEndFlag {
		return nil
	}
	if g.phase == TuSacPhaseRoundEnd {
		return &TuSacHint{Action: "next", Reason: "roundIsOver"}
	}
	if !g.IsHumanTurn() {
		return nil
	}
	p := g.players[g.HumanSeat()]

	if g.phase == TuSacPhaseDraw {
		// **捨て札が組み合わせを完成させるなら拾う。**
		if top := g.GetDiscardTop(); top != nil {
			probe := make([]*Card, 0, len(p.GetCards())+1)
			probe = append(probe, p.GetCards()...)
			probe = append(probe, top)
			if g.findAnyMeld(probe) != nil && g.findAnyMeld(p.GetCards()) == nil {
				return &TuSacHint{Action: "draw", Reason: "discardCompletesAMeld"}
			}
		}
		return &TuSacHint{Action: "draw", Reason: "drawFromTheStock"}
	}

	if idx := g.findAnyMeld(p.GetCards()); idx != nil {
		return &TuSacHint{Action: "meld", Indexes: idx, Reason: "youCanLayAMeld"}
	}
	return &TuSacHint{
		Action:  "discard",
		Indexes: []int{g.cpuDiscardIndex(g.HumanSeat())},
		Reason:  "discardTheLoneCard",
	}
}

// --- 永続化 ---

// tuSacJSON is the JSON wire format for TuSac.
type tuSacJSON struct {
	Stock       []*Card           `json:"st"`
	Discard     []*Card           `json:"dc"`
	Players     []*TuSacPlayer    `json:"pl"`
	Config      TuSacConfig       `json:"cf"`
	Phase       int               `json:"ph"`
	Turn        int               `json:"tu"`
	RoundNumber int               `json:"rn"`
	WentOut     int               `json:"wo"`
	Results     []TuSacResult     `json:"rs"`
	GameEndFlag bool              `json:"ge"`
	ActionLog   []*ActionLogEntry `json:"al"`
	TurnNumber  int               `json:"tn"`
}

// MarshalJSON implements json.Marshaler.
func (g *TuSac) MarshalJSON() ([]byte, error) {
	return json.Marshal(tuSacJSON{
		Stock: g.stock, Discard: g.discard, Players: g.players, Config: g.config,
		Phase: int(g.phase), Turn: g.turn, RoundNumber: g.roundNumber,
		WentOut: g.wentOut, Results: g.results, GameEndFlag: g.gameEndFlag,
		ActionLog: g.actionLog, TurnNumber: g.turnNumber,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *TuSac) UnmarshalJSON(data []byte) error {
	var j tuSacJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := tuSacValidate(&j); err != nil {
		return err
	}

	g.stock = j.Stock
	g.discard = j.Discard
	g.players = j.Players
	g.config = j.Config
	g.phase = TuSacPhase(j.Phase)
	g.turn = j.Turn
	g.roundNumber = j.RoundNumber
	g.wentOut = j.WentOut
	g.results = j.Results
	g.gameEndFlag = j.GameEndFlag
	g.actionLog = j.ActionLog
	g.turnNumber = j.TurnNumber

	if g.results == nil {
		g.results = make([]TuSacResult, 0, len(g.players))
	}
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// tuSacValidate は復元した盤面が破綻していないかを見る。
func tuSacValidate(j *tuSacJSON) error {
	if len(j.Players) > tuSacMaxSliceLen || len(j.ActionLog) > tuSacMaxSliceLen ||
		len(j.Results) > tuSacMaxSliceLen {
		return errTuSacSliceTooLong
	}
	if len(j.Stock) > TuSacDeckSize || len(j.Discard) > TuSacDeckSize {
		return errTuSacDeckTooLarge
	}
	if len(j.Players) != j.Config.Seats {
		return errTuSacSeatCount
	}
	if j.Phase < 0 || TuSacPhase(j.Phase) > TuSacPhaseMax {
		return errTuSacPhaseRange
	}
	if j.Turn < 0 || j.Turn >= len(j.Players) {
		return errTuSacTurnRange
	}
	if j.RoundNumber < 1 || j.RoundNumber > j.Config.Rounds {
		return errTuSacRoundRange
	}
	// **上がった席は実在の席か -1。** 範囲外だと結果の表示が別の席を指す。
	if j.WentOut < -1 || j.WentOut >= len(j.Players) {
		return errTuSacTurnRange
	}
	return nil
}
