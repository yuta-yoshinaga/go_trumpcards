//go:build !js || !wasm || extra3

// Package domain チルッラ (Cirulla) のドメインモデル。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// CirullaPlayerCnt はプレイヤー数 (ヘッズアップ)。
const CirullaPlayerCnt = 2

// CirullaHandSize は 1 回に配る枚数。
const CirullaHandSize = 3

// CirullaTableSize は最初に場へ置く枚数。
const CirullaTableSize = 4

// CirullaDeckSize は 40 枚デッキの枚数。
const CirullaDeckSize = 40

// フェーズ。
const (
	// CirullaPhasePlay プレイ中。
	CirullaPhasePlay = "play"
	// CirullaPhaseRoundEnd 1 ラウンド終了 (集計済み、次ラウンド待ち)。
	CirullaPhaseRoundEnd = "roundEnd"
	// CirullaPhaseGameEnd ゲーム終了。
	CirullaPhaseGameEnd = "gameEnd"
)

// CirullaScoreLine は 1 つの得点項目。
type CirullaScoreLine struct {
	// Key は項目の識別子 ("cards" / "denari" / "settebello" / "primiera" /
	// "piccola" / "grande" / "scope" / "bonus")。
	Key string
	// Points は席ごとの点。
	Points [CirullaPlayerCnt]int
}

// CirullaRoundResult は 1 ラウンドの集計結果。
type CirullaRoundResult struct {
	// Lines は項目別の内訳。
	Lines []CirullaScoreLine
	// Totals は席ごとのこのラウンドの合計。
	Totals [CirullaPlayerCnt]int
	// SweptDenari は全デナリを取った席 (-1 = なし)。
	SweptDenari int
}

// Cirulla はチルッラの状態を保持する集約ルート。
type Cirulla struct {
	deck    []*Card
	drawIdx int
	players []*CirullaPlayer
	config  CirullaConfig
	phase   string
	// table は場札。
	table []*Card
	// roundNumber は現在のラウンド。
	roundNumber   int
	dealerIdx     int
	currentPlayer int
	// lastCapturer は最後に捕獲した席。**ラウンド末の場札はここへ渡る。**
	lastCapturer int
	// lastBonus は直近の配札ボーナス (席別の識別子)。
	lastBonus   [CirullaPlayerCnt]string
	lastResult  *CirullaRoundResult
	gameEndFlag bool
	winnerIdx   int
	actionLogBase
}

// NewCirulla はコンストラクタ。
func NewCirulla(players []*CirullaPlayer, config CirullaConfig) *Cirulla {
	return &Cirulla{
		players:      players,
		config:       config,
		phase:        CirullaPhasePlay,
		lastCapturer: -1,
		winnerIdx:    -1,
	}
}

// NewDefaultCirulla は標準の 2 人構成 (1 human + 1 CPU) を生成する。
func NewDefaultCirulla() *Cirulla {
	players := make([]*CirullaPlayer, CirullaPlayerCnt)
	players[0] = NewCirullaPlayer(true)
	players[1] = NewCirullaPlayer(false)
	return NewCirulla(players, DefaultCirullaConfig())
}

// Reset は新しいゲームを開始する。
func (c *Cirulla) Reset() {
	for _, p := range c.players {
		p.ResetRound()
		p.ResetScore()
	}
	c.roundNumber = 1
	c.dealerIdx = 1 // 親の左隣 (席 0 = 人間) から打つ
	c.gameEndFlag = false
	c.winnerIdx = -1
	c.lastResult = nil
	c.actionLog = make([]*ActionLogEntry, 0)
	c.startRound()
}

// NextRound は次のラウンドを開始する。
func (c *Cirulla) NextRound() {
	if c.gameEndFlag || c.phase != CirullaPhaseRoundEnd {
		return
	}
	c.roundNumber++
	c.dealerIdx = (c.dealerIdx + 1) % CirullaPlayerCnt
	c.startRound()
}

// startRound は山を作り、3 枚ずつ配って場に 4 枚置く。
func (c *Cirulla) startRound() {
	for _, p := range c.players {
		p.ResetRound()
	}
	// **山は 40 枚を引き切って使う。** TrumpCards の DrawCard から自前の
	// スライスへ移しておくことで、残り枚数を数えたり保存したりできる。
	src := NewTrumpCardsScopa()
	src.Shuffle()
	c.deck = make([]*Card, 0, CirullaDeckSize)
	for {
		card := src.DrawCard()
		if card == nil {
			break
		}
		c.deck = append(c.deck, card)
	}
	c.drawIdx = 0
	c.table = make([]*Card, 0, CirullaTableSize)
	c.lastCapturer = -1
	c.lastBonus = [CirullaPlayerCnt]string{}

	c.dealHands()
	for i := 0; i < CirullaTableSize; i++ {
		if card := c.draw(); card != nil {
			c.table = append(c.table, card)
		}
	}
	c.currentPlayer = (c.dealerIdx + 1) % CirullaPlayerCnt
	c.phase = CirullaPhasePlay
	c.appendLog(-1, "deal", fmt.Sprintf("round %d: dealer=%d, table=%d",
		c.roundNumber, c.dealerIdx, len(c.table)), nil)
}

// dealHands は各席へ 3 枚配り、配札ボーナスを判定する。
//
// **ボーナスは配るたびに見る。** 最初の 3 枚だけの話ではなく、山が尽きるまで
// 配り直すたびに成立しうる。
func (c *Cirulla) dealHands() {
	for j := 0; j < CirullaPlayerCnt; j++ {
		idx := (c.dealerIdx + 1 + j) % CirullaPlayerCnt
		hand := make([]*Card, 0, CirullaHandSize)
		for i := 0; i < CirullaHandSize; i++ {
			card := c.draw()
			if card == nil {
				break
			}
			c.players[idx].AddCard(card)
			hand = append(hand, card)
		}
		name, points := CirullaDealBonus(hand)
		c.lastBonus[idx] = name
		if points > 0 {
			c.players[idx].AddBonusPoints(points)
			c.appendLog(idx, "bonus",
				fmt.Sprintf("player %d scores %s for %d", idx, name, points), hand)
		}
	}
}

// draw は山から 1 枚引く。
func (c *Cirulla) draw() *Card {
	if c.drawIdx >= len(c.deck) {
		return nil
	}
	card := c.deck[c.drawIdx]
	c.drawIdx++
	return card
}

// PlayerPlay は人間が 1 枚出す。captureIdxs は取る場札 (空なら場に置く)。
func (c *Cirulla) PlayerPlay(handIdx int, captureIdxs []int) error {
	if c.gameEndFlag {
		return ErrGameEnded
	}
	if c.phase != CirullaPhasePlay {
		return ErrWrongPhase
	}
	if !c.players[c.currentPlayer].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return c.applyPlay(c.currentPlayer, handIdx, captureIdxs)
}

// CpuPlay は CPU が 1 枚出す。
func (c *Cirulla) CpuPlay() {
	if c.gameEndFlag || c.phase != CirullaPhasePlay {
		return
	}
	if c.players[c.currentPlayer].GetIsHuman() {
		return
	}
	handIdx, captureIdxs := c.cpuChoose(c.currentPlayer)
	_ = c.applyPlay(c.currentPlayer, handIdx, captureIdxs)
}

// applyPlay は 1 手を適用する。
func (c *Cirulla) applyPlay(playerIdx, handIdx int, captureIdxs []int) error {
	player := c.players[playerIdx]
	if handIdx < 0 || handIdx >= player.GetCardsSize() {
		return NewDomainErrorCode(ErrInvalidCard, "cirulla.errCardRange", nil)
	}
	card := player.GetCard(handIdx)

	if len(captureIdxs) == 0 {
		// **取れるのに置くことはできない。** 取れる手があるなら取る。
		if len(EnumerateCirullaCaptures(card, c.table)) > 0 {
			return NewDomainErrorCode(ErrInvalidPlay, "cirulla.errMustCapture", nil)
		}
		played := player.RemoveCard(handIdx)
		c.table = append(c.table, played)
		c.appendLog(playerIdx, "discard",
			fmt.Sprintf("player %d lays %s", playerIdx, cardStr(played)), []*Card{played})
		c.advance()
		return nil
	}

	if !IsValidCirullaCapture(card, c.table, captureIdxs) {
		return NewDomainErrorCode(ErrInvalidPlay, "cirulla.errInvalidCapture", nil)
	}
	played := player.RemoveCard(handIdx)
	taken := c.removeFromTable(captureIdxs)
	player.AddCaptured(append(taken, played))
	c.lastCapturer = playerIdx
	c.appendLog(playerIdx, "capture",
		fmt.Sprintf("player %d takes %d card(s) with %s", playerIdx, len(taken), cardStr(played)),
		append([]*Card{played}, taken...))

	// **場を空にしたらスコパ。** ただし山も手札も尽きた最後の手は数えない。
	if len(c.table) == 0 && !c.isFinalPlay() {
		player.AddScopa()
		c.appendLog(playerIdx, "scopa", fmt.Sprintf("player %d sweeps the table", playerIdx), nil)
	}
	c.advance()
	return nil
}

// isFinalPlay は山も手札も尽きたかを返す。
func (c *Cirulla) isFinalPlay() bool {
	if c.drawIdx < len(c.deck) {
		return false
	}
	for _, p := range c.players {
		if p.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// removeFromTable は指定のインデックスを場から取り除いて返す。
func (c *Cirulla) removeFromTable(idxs []int) []*Card {
	keep := make([]*Card, 0, len(c.table))
	take := make([]*Card, 0, len(idxs))
	selected := make(map[int]bool, len(idxs))
	for _, i := range idxs {
		selected[i] = true
	}
	for i, card := range c.table {
		if selected[i] {
			take = append(take, card)
			continue
		}
		keep = append(keep, card)
	}
	c.table = keep
	return take
}

// advance は次の手番へ進める。手札が尽きたら配り直し、山も尽きたら集計する。
func (c *Cirulla) advance() {
	c.currentPlayer = (c.currentPlayer + 1) % CirullaPlayerCnt
	if c.handsEmpty() {
		if c.drawIdx < len(c.deck) {
			c.dealHands()
			c.currentPlayer = (c.dealerIdx + 1) % CirullaPlayerCnt
			return
		}
		c.finishRound()
	}
}

// handsEmpty は全員の手札が空かを返す。
func (c *Cirulla) handsEmpty() bool {
	for _, p := range c.players {
		if p.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// finishRound はラウンドを締めて集計する。
//
// **場に残った札は最後に取った人のもの。** 拾わないと 40 枚の勘定が合わず、
// 「最多枚数」の判定が狂う。
func (c *Cirulla) finishRound() {
	if len(c.table) > 0 && c.lastCapturer >= 0 {
		c.players[c.lastCapturer].AddCaptured(c.table)
		c.appendLog(c.lastCapturer, "sweep",
			fmt.Sprintf("player %d takes the %d card(s) left on the table",
				c.lastCapturer, len(c.table)), c.table)
		c.table = make([]*Card, 0)
	}

	result := c.scoreRound()
	c.lastResult = result
	for i, p := range c.players {
		p.AddScore(result.Totals[i])
	}
	c.phase = CirullaPhaseRoundEnd
	c.appendLog(-1, "roundEnd", fmt.Sprintf("round %d: %d - %d",
		c.roundNumber, result.Totals[0], result.Totals[1]), nil)
	c.checkGameEnd(result)
}

// scoreRound はラウンドの得点を項目別に集計する。
func (c *Cirulla) scoreRound() *CirullaRoundResult {
	res := &CirullaRoundResult{Lines: make([]CirullaScoreLine, 0, 8), SweptDenari: -1}

	add := func(key string, points [CirullaPlayerCnt]int) {
		res.Lines = append(res.Lines, CirullaScoreLine{Key: key, Points: points})
		for i := range points {
			res.Totals[i] += points[i]
		}
	}

	// 最多枚数・最多デナリは同数なら誰にも入らない。
	add("cards", cirullaAwardMax(c.countBy(func(*Card) bool { return true })))
	denari := c.countBy(CirullaIsDenari)
	add("denari", cirullaAwardMax(denari))
	for i, n := range denari {
		if n == 10 {
			res.SweptDenari = i
		}
	}
	add("settebello", c.awardHolder(CirullaIsSetteBello))

	primiera := [CirullaPlayerCnt]int{}
	for i, p := range c.players {
		primiera[i] = CirullaPrimiera(p.GetCaptured())
	}
	add("primiera", cirullaAwardMax(primiera))

	piccola := [CirullaPlayerCnt]int{}
	grande := [CirullaPlayerCnt]int{}
	scope := [CirullaPlayerCnt]int{}
	bonus := [CirullaPlayerCnt]int{}
	for i, p := range c.players {
		piccola[i] = CirullaPiccola(p.GetCaptured())
		if CirullaHasGrande(p.GetCaptured()) {
			grande[i] = CirullaGrandePoints
		}
		scope[i] = p.GetScope()
		bonus[i] = p.GetBonusPoints()
	}
	add("piccola", piccola)
	add("grande", grande)
	add("scope", scope)
	add("bonus", bonus)
	return res
}

// countBy は席ごとに条件を満たす獲得札の枚数を返す。
func (c *Cirulla) countBy(pred func(*Card) bool) [CirullaPlayerCnt]int {
	out := [CirullaPlayerCnt]int{}
	for i, p := range c.players {
		for _, card := range p.GetCaptured() {
			if pred(card) {
				out[i]++
			}
		}
	}
	return out
}

// awardHolder は条件を満たす札を持つ席に 1 点を与える。
func (c *Cirulla) awardHolder(pred func(*Card) bool) [CirullaPlayerCnt]int {
	out := [CirullaPlayerCnt]int{}
	for i, p := range c.players {
		for _, card := range p.GetCaptured() {
			if pred(card) {
				out[i] = 1
				break
			}
		}
	}
	return out
}

// cirullaAwardMax は最多の席に 1 点を与える。**同数なら誰にも入らない。**
func cirullaAwardMax(counts [CirullaPlayerCnt]int) [CirullaPlayerCnt]int {
	out := [CirullaPlayerCnt]int{}
	if counts[0] > counts[1] {
		out[0] = 1
	} else if counts[1] > counts[0] {
		out[1] = 1
	}
	return out
}

// checkGameEnd は終局を判定する。
//
// **全デナリを取ると即勝ち。** 点が足りていなくても、その時点で決まる。
func (c *Cirulla) checkGameEnd(result *CirullaRoundResult) {
	if result != nil && result.SweptDenari >= 0 {
		c.finishGame(result.SweptDenari)
		return
	}
	best, bestIdx, tied := -1, -1, false
	for i, p := range c.players {
		switch {
		case p.GetScore() > best:
			best, bestIdx, tied = p.GetScore(), i, false
		case p.GetScore() == best:
			tied = true
		}
	}
	if best >= c.config.TargetScore && !tied {
		c.finishGame(bestIdx)
	}
}

// finishGame は終局する。
func (c *Cirulla) finishGame(winner int) {
	c.gameEndFlag = true
	c.winnerIdx = winner
	c.phase = CirullaPhaseGameEnd
	c.appendLog(-1, "gameEnd", fmt.Sprintf("player %d wins the match", winner), nil)
}

// IsHumanTurn は人間の手番かを返す。
func (c *Cirulla) IsHumanTurn() bool {
	if c.gameEndFlag || c.phase != CirullaPhasePlay {
		return false
	}
	if c.currentPlayer < 0 || c.currentPlayer >= len(c.players) {
		return false
	}
	return c.players[c.currentPlayer].GetIsHuman()
}

// GetCaptureOptions は手札 handIdx を出したときの捕獲候補を返す。
func (c *Cirulla) GetCaptureOptions(playerIdx, handIdx int) [][]int {
	if playerIdx < 0 || playerIdx >= len(c.players) {
		return nil
	}
	p := c.players[playerIdx]
	if handIdx < 0 || handIdx >= p.GetCardsSize() {
		return nil
	}
	return EnumerateCirullaCaptures(p.GetCard(handIdx), c.table)
}

// CirullaHint は人間への推奨手。
type CirullaHint struct {
	// HandIdx は勧める手札。
	HandIdx int
	// CaptureIdxs は勧める捕獲対象 (空なら場に置く)。
	CaptureIdxs []int
	// Reason は理由の識別子。
	Reason string
}

// GetHint は人間への推奨手を返す。
func (c *Cirulla) GetHint() *CirullaHint {
	human := findHumanIdx(c.players)
	if human < 0 || c.gameEndFlag {
		return &CirullaHint{HandIdx: -1, Reason: "none"}
	}
	switch c.phase {
	case CirullaPhasePlay:
		if c.currentPlayer != human || c.players[human].GetCardsSize() == 0 {
			return &CirullaHint{HandIdx: -1, Reason: "none"}
		}
		// **助言は CPU の難易度に引きずらせない。**
		handIdx, capture := c.smartChoose(human)
		reason := "lay_off"
		if len(capture) > 0 {
			reason = "capture"
			if len(capture) == len(c.table) {
				reason = "sweep"
			}
		}
		return &CirullaHint{HandIdx: handIdx, CaptureIdxs: capture, Reason: reason}
	case CirullaPhaseRoundEnd:
		return &CirullaHint{HandIdx: -1, Reason: "next_round"}
	default:
		return &CirullaHint{HandIdx: -1, Reason: "none"}
	}
}

// --- 参照 ---

// GetConfig はゲーム設定を返す。
func (c *Cirulla) GetConfig() CirullaConfig { return c.config }

// SetConfig はゲーム設定を設定する。
func (c *Cirulla) SetConfig(cfg CirullaConfig) { c.config = cfg }

// GetPhase は現在のフェーズを返す。
func (c *Cirulla) GetPhase() string { return c.phase }

// GetPlayerCnt は席数を返す。
func (c *Cirulla) GetPlayerCnt() int { return len(c.players) }

// GetPlayer は席のプレイヤーを返す。
func (c *Cirulla) GetPlayer(i int) *CirullaPlayer {
	if i < 0 || i >= len(c.players) {
		return nil
	}
	return c.players[i]
}

// GetPlayers は全プレイヤーを返す。
func (c *Cirulla) GetPlayers() []*CirullaPlayer { return c.players }

// GetTable は場札を返す。
func (c *Cirulla) GetTable() []*Card { return c.table }

// GetRoundNumber は現在のラウンドを返す。
func (c *Cirulla) GetRoundNumber() int { return c.roundNumber }

// GetDealerIdx は親の席を返す。
func (c *Cirulla) GetDealerIdx() int { return c.dealerIdx }

// GetCurrentPlayerIdx は手番の席を返す。
func (c *Cirulla) GetCurrentPlayerIdx() int { return c.currentPlayer }

// GetLastCapturer は最後に捕獲した席を返す (-1 = なし)。
func (c *Cirulla) GetLastCapturer() int { return c.lastCapturer }

// GetLastBonus は席ごとの直近の配札ボーナス識別子を返す。
func (c *Cirulla) GetLastBonus() []string { return c.lastBonus[:] }

// GetDeckRemaining は山の残り枚数を返す。
func (c *Cirulla) GetDeckRemaining() int { return len(c.deck) - c.drawIdx }

// GetLastResult は直前ラウンドの集計を返す。
func (c *Cirulla) GetLastResult() *CirullaRoundResult { return c.lastResult }

// GetGameEndFlag は終局フラグを返す。
func (c *Cirulla) GetGameEndFlag() bool { return c.gameEndFlag }

// GetWinnerIdx は勝者の席を返す (-1 = 未決)。
func (c *Cirulla) GetWinnerIdx() int { return c.winnerIdx }

// --- 永続化 ---

// cirullaJSON is the JSON wire format for Cirulla.
type cirullaJSON struct {
	Deck          []*Card                  `json:"dk"`
	DrawIdx       int                      `json:"dx"`
	Players       []*CirullaPlayer         `json:"pl"`
	Config        CirullaConfig            `json:"cf"`
	Phase         string                   `json:"ph"`
	Table         []*Card                  `json:"tb"`
	RoundNumber   int                      `json:"rn"`
	DealerIdx     int                      `json:"di"`
	CurrentPlayer int                      `json:"cp"`
	LastCapturer  int                      `json:"lc"`
	LastBonus     [CirullaPlayerCnt]string `json:"lb"`
	LastResult    *CirullaRoundResult      `json:"lr"`
	GameEndFlag   bool                     `json:"ge"`
	WinnerIdx     int                      `json:"wi"`
	ActionLog     []*ActionLogEntry        `json:"al"`
}

// cirullaMaxSliceLen は復元時に受け入れるスライスの上限。
const cirullaMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
//
// **非公開フィールドだけの型は MarshalJSON が無いと `{}` になる。**
func (c *Cirulla) MarshalJSON() ([]byte, error) {
	return json.Marshal(cirullaJSON{
		Deck:          c.deck,
		DrawIdx:       c.drawIdx,
		Players:       c.players,
		Config:        c.config,
		Phase:         c.phase,
		Table:         c.table,
		RoundNumber:   c.roundNumber,
		DealerIdx:     c.dealerIdx,
		CurrentPlayer: c.currentPlayer,
		LastCapturer:  c.lastCapturer,
		LastBonus:     c.lastBonus,
		LastResult:    c.lastResult,
		GameEndFlag:   c.gameEndFlag,
		WinnerIdx:     c.winnerIdx,
		ActionLog:     c.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Cirulla) UnmarshalJSON(data []byte) error {
	var j cirullaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Deck) > cirullaMaxSliceLen || len(j.Table) > cirullaMaxSliceLen ||
		len(j.ActionLog) > cirullaMaxSliceLen {
		return fmt.Errorf("cirulla: input array exceeds maximum allowed size")
	}
	if len(j.Players) != CirullaPlayerCnt {
		return fmt.Errorf("cirulla: invalid player count %d, expected %d", len(j.Players), CirullaPlayerCnt)
	}
	c.deck = j.Deck
	c.drawIdx = j.DrawIdx
	c.players = j.Players
	c.config = j.Config
	c.phase = j.Phase
	c.table = j.Table
	if c.table == nil {
		c.table = make([]*Card, 0)
	}
	c.roundNumber = j.RoundNumber
	c.dealerIdx = j.DealerIdx
	c.currentPlayer = j.CurrentPlayer
	c.lastCapturer = j.LastCapturer
	c.lastBonus = j.LastBonus
	c.lastResult = j.LastResult
	c.gameEndFlag = j.GameEndFlag
	c.winnerIdx = j.WinnerIdx
	c.actionLog = j.ActionLog
	if c.actionLog == nil {
		c.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// cirullaRoundResultJSON is the JSON wire format for CirullaRoundResult.
type cirullaRoundResultJSON struct {
	Lines       []CirullaScoreLine    `json:"l"`
	Totals      [CirullaPlayerCnt]int `json:"t"`
	SweptDenari int                   `json:"s"`
}

// MarshalJSON implements json.Marshaler.
func (r *CirullaRoundResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(cirullaRoundResultJSON{Lines: r.Lines, Totals: r.Totals, SweptDenari: r.SweptDenari})
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *CirullaRoundResult) UnmarshalJSON(data []byte) error {
	var j cirullaRoundResultJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	r.Lines = j.Lines
	r.Totals = j.Totals
	r.SweptDenari = j.SweptDenari
	return nil
}

// cirullaRandIntn は rand.Intn の薄いラッパ (n <= 0 を握りつぶす)。
func cirullaRandIntn(n int) int {
	if n <= 0 {
		return 0
	}
	return rand.Intn(n)
}

// SetTableForTest は場札を差し替える (テスト用)。
//
// **場は配りで決まるので、狙った盤面は組めない。** 捕獲規則を確かめるには
// ここで固定するしかない。
func (c *Cirulla) SetTableForTest(cards []*Card) { c.table = cards }
