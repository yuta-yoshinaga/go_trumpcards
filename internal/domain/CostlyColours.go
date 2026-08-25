//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"fmt"
)

// 卓の形。
const (
	// CostlyColoursPlayerCnt は席数 (人間 1 + CPU 1)。
	CostlyColoursPlayerCnt = 2
	// CostlyColoursHandSize は配る枚数。**3 枚だけ。**
	CostlyColoursHandSize = 3
	// CostlyColoursShowSize はショーで数える枚数 (手札 3 + 表の 1)。
	CostlyColoursShowSize = 4
	// CostlyColoursDeckSize は使う札の枚数。
	CostlyColoursDeckSize = 52
)

// フェーズ。
//
// **交換 (mog) は独立したフェーズ。** 配ったあと打ち始める前に、相手と 1 枚
// 交換するかを決める ── ここを飛ばすと #5461 の「クリブが無いだけ」に戻る。
const (
	// CostlyColoursPhaseMog は交換の可否を決めている状態。
	CostlyColoursPhaseMog = "mog"
	// CostlyColoursPhasePlay は 31 まで数え上げている状態。
	CostlyColoursPhasePlay = "play"
	// CostlyColoursPhaseShow は 1 ディールの集計を見せている状態。
	CostlyColoursPhaseShow = "show"
	// CostlyColoursPhaseGameEnd は終局。
	CostlyColoursPhaseGameEnd = "gameEnd"
)

// CostlyColoursScoreLine は 1 つの得点項目。
type CostlyColoursScoreLine struct {
	// Key は項目の識別子。
	Key string
	// Points は席ごとの点。
	Points []int
}

// CostlyColoursDealResult は 1 ディールの集計。
type CostlyColoursDealResult struct {
	// Lines は項目別の内訳。
	Lines []CostlyColoursScoreLine
	// Totals は席ごとのこのディールの合計。
	Totals [CostlyColoursPlayerCnt]int
	// Combos は席ごとの色とスートの役の識別子。
	Combos [CostlyColoursPlayerCnt]string
}

// CostlyColours はコストリー・カラーズの状態を保持する集約ルート。
type CostlyColours struct {
	deck    []*Card
	drawIdx int
	players []*CostlyColoursPlayer
	config  CostlyColoursConfig
	// turnUp は表に返した「トランプ」の 1 枚。ショーではこれも数える。
	turnUp *Card
	// pile は今の数え上げに出た札。
	pile []*Card
	// total は今の数え上げの累計。**31 を超えられない。**
	total int
	// wentOut は「ゴー」を宣言した席 (-1 = 誰も宣言していない)。
	wentOut     int
	phase       string
	dealNumber  int
	dealerIdx   int
	currentIdx  int
	lastResult  *CostlyColoursDealResult
	gameEndFlag bool
	winnerIdx   int
	actionLogBase
}

// NewCostlyColours はコンストラクタ。
func NewCostlyColours(players []*CostlyColoursPlayer, config CostlyColoursConfig) *CostlyColours {
	return &CostlyColours{
		players:   players,
		config:    config,
		phase:     CostlyColoursPhaseMog,
		wentOut:   -1,
		winnerIdx: -1,
	}
}

// NewDefaultCostlyColours は既定の設定で生成する。
func NewDefaultCostlyColours() *CostlyColours {
	players := []*CostlyColoursPlayer{
		NewCostlyColoursPlayer(true), NewCostlyColoursPlayer(false),
	}
	return NewCostlyColours(players, DefaultCostlyColoursConfig())
}

// Reset はゲームを最初から始める。
//
// **親を席 1 にして席 0 から打たせる。** 非親 (エルダー) が先に打ち、交換を
// 持ちかけるのも非親なので、親を 0 にすると人間は最初の選択を持てない。
func (c *CostlyColours) Reset() {
	for _, p := range c.players {
		p.ResetDeal()
		p.ResetScore()
	}
	c.dealNumber = 1
	c.dealerIdx = CostlyColoursPlayerCnt - 1
	c.gameEndFlag = false
	c.winnerIdx = -1
	c.lastResult = nil
	c.actionLog = make([]*ActionLogEntry, 0)
	c.startDeal()
}

// NextDeal は次のディールを始める。
func (c *CostlyColours) NextDeal() {
	if c.gameEndFlag || c.phase != CostlyColoursPhaseShow {
		return
	}
	c.dealNumber++
	c.dealerIdx = (c.dealerIdx + 1) % CostlyColoursPlayerCnt
	c.startDeal()
}

// startDeal は 3 枚ずつ配り、次の 1 枚を表に返す。
func (c *CostlyColours) startDeal() {
	for _, p := range c.players {
		p.ResetDeal()
	}
	src := NewTrumpCards(0)
	src.Shuffle()
	c.deck = make([]*Card, 0, CostlyColoursDeckSize)
	for {
		card := src.DrawCard()
		if card == nil {
			break
		}
		c.deck = append(c.deck, card)
	}
	c.drawIdx = 0
	for i := 0; i < CostlyColoursHandSize; i++ {
		for seat := 0; seat < CostlyColoursPlayerCnt; seat++ {
			idx := (c.dealerIdx + 1 + seat) % CostlyColoursPlayerCnt
			if card := c.draw(); card != nil {
				c.players[idx].AddCard(card)
			}
		}
	}
	// **次の 1 枚を表に返す。** これが「トランプ」で、ショーでも数える。
	c.turnUp = c.draw()
	c.pile = make([]*Card, 0, CostlyColoursDeckSize)
	c.total = 0
	c.wentOut = -1
	c.phase = CostlyColoursPhaseMog
	c.currentIdx = (c.dealerIdx + 1) % CostlyColoursPlayerCnt

	c.appendLog(-1, "deal", fmt.Sprintf("deal %d: dealer=%d, turn-up shown",
		c.dealNumber, c.dealerIdx), nil)
	// **表に返した J は親の 4 点。** 打ち始める前に入る ("for his heels")。
	if c.turnUp != nil && c.turnUp.GetValue() == 11 {
		c.players[c.dealerIdx].AddScore(CostlyHeelsPoints)
		c.appendLog(c.dealerIdx, "heels",
			fmt.Sprintf("player %d pegs %d for his heels", c.dealerIdx, CostlyHeelsPoints), nil)
		c.checkGameEnd()
	}
}

// draw は山から 1 枚引く。
func (c *CostlyColours) draw() *Card {
	if c.drawIdx >= len(c.deck) {
		return nil
	}
	card := c.deck[c.drawIdx]
	c.drawIdx++
	return card
}

// PlayerMog は人間が交換に応じるかを決める。
//
// **断れば相手に 1 点。** 交換そのものは 1 枚ずつの取り替えなので手札は
// 3 枚のまま。
func (c *CostlyColours) PlayerMog(accept bool) error {
	human := findHumanIdx(c.players)
	if human < 0 {
		return NewDomainErrorCode(ErrInvalidPlay, "costlycolours.errNoHuman", nil)
	}
	if c.gameEndFlag {
		return NewDomainErrorCode(ErrGameEnded, "costlycolours.errGameEnded", nil)
	}
	if c.phase != CostlyColoursPhaseMog {
		return NewDomainErrorCode(ErrWrongPhase, "costlycolours.errNotMogPhase", nil)
	}
	c.resolveMog(human, accept)
	return nil
}

// resolveMog は交換の可否を反映し、数え上げへ進む。
func (c *CostlyColours) resolveMog(seat int, accept bool) {
	other := (seat + 1) % CostlyColoursPlayerCnt
	if !accept {
		// **断った側ではなく、断られた側に 1 点。**
		c.players[other].AddScore(CostlyMogRefusalPoints)
		c.appendLog(seat, "mogRefused",
			fmt.Sprintf("player %d refuses; player %d pegs %d", seat, other, CostlyMogRefusalPoints), nil)
	} else {
		c.swapOneCard(seat, other)
		c.appendLog(seat, "mog", fmt.Sprintf("player %d and %d exchange a card", seat, other), nil)
	}
	c.phase = CostlyColoursPhasePlay
	c.currentIdx = (c.dealerIdx + 1) % CostlyColoursPlayerCnt
	c.checkGameEnd()
}

// swapOneCard は 2 席が 1 枚ずつ取り替える。
//
// **手札は 3 枚のまま。** 渡すのは互いにいちばん要らない札 ── ここでは
// 数え上げで潰しの利かない大きい札を出す。
func (c *CostlyColours) swapOneCard(a, b int) {
	ia, ib := costlyWorstCardIdx(c.players[a].GetHand()), costlyWorstCardIdx(c.players[b].GetHand())
	if ia < 0 || ib < 0 {
		return
	}
	ca := c.players[a].RemoveCard(ia)
	cb := c.players[b].RemoveCard(ib)
	if ca != nil {
		c.players[b].AddCard(ca)
		c.players[b].SetMoggedIn(true)
	}
	if cb != nil {
		c.players[a].AddCard(cb)
		c.players[a].SetMoggedIn(true)
	}
}

// costlyWorstCardIdx は手放してよい札の位置を返す。
//
// **J と 2 は残す。** どちらも手札にあるだけで点になるので、渡すと相手に
// その点を献上することになる。
func costlyWorstCardIdx(hand []*Card) int {
	best, bestScore := -1, -1
	for i, card := range hand {
		if card == nil {
			continue
		}
		score := CostlyCardValue(card)
		if CostlyIsJackOrDeuce(card) {
			score = -1 // 最後まで残す。
		}
		if score > bestScore {
			best, bestScore = i, score
		}
	}
	return best
}

// PlayerPlay は人間が 1 枚出す。
func (c *CostlyColours) PlayerPlay(handIdx int) error {
	human := findHumanIdx(c.players)
	if human < 0 {
		return NewDomainErrorCode(ErrInvalidPlay, "costlycolours.errNoHuman", nil)
	}
	if c.gameEndFlag {
		return NewDomainErrorCode(ErrGameEnded, "costlycolours.errGameEnded", nil)
	}
	if c.phase != CostlyColoursPhasePlay {
		return NewDomainErrorCode(ErrWrongPhase, "costlycolours.errNotPlayPhase", nil)
	}
	if c.currentIdx != human {
		return NewDomainErrorCode(ErrNotHumanTurn, "costlycolours.errNotYourTurn", nil)
	}
	return c.applyPlay(human, handIdx)
}

// applyPlay は 1 枚を数え上げへ足す。
func (c *CostlyColours) applyPlay(seat, handIdx int) error {
	p := c.players[seat]
	hand := p.GetHand()
	if handIdx < 0 || handIdx >= len(hand) {
		return NewDomainErrorCode(ErrInvalidCard, "costlycolours.errCardRange",
			map[string]string{"idx": fmt.Sprint(handIdx)})
	}
	card := hand[handIdx]
	// **31 を超える札は出せない。**
	if c.total+CostlyCardValue(card) > CostlyThirtyOne {
		return NewDomainErrorCode(ErrInvalidPlay, "costlycolours.errWouldExceed",
			map[string]string{"total": fmt.Sprint(c.total)})
	}

	p.RemoveCard(handIdx)
	p.AddPlayed(card)
	c.pile = append(c.pile, card)
	c.total += CostlyCardValue(card)

	pts, reasons := CostlyPlayScore(c.pile, c.total)
	if pts > 0 {
		p.AddScore(pts)
		c.appendLog(seat, "peg", fmt.Sprintf("player %d pegs %d (%v)", seat, pts, reasons), []*Card{card})
	} else {
		c.appendLog(seat, "play", fmt.Sprintf("player %d plays, total %d", seat, c.total), []*Card{card})
	}
	c.checkGameEnd()
	if c.gameEndFlag {
		return nil
	}
	c.advanceAfterPlay(seat)
	return nil
}

// advanceAfterPlay は次の手番を決め、必要なら数え上げを畳んでショーへ進む。
func (c *CostlyColours) advanceAfterPlay(seat int) {
	if c.handsEmpty() {
		c.awardLatter(seat)
		c.finishDeal()
		return
	}
	if c.total == CostlyThirtyOne {
		// **31 ちょうどで数え上げは終わり。** 次は 0 から数え直す。
		c.resetCount(seat)
		return
	}
	other := (seat + 1) % CostlyColoursPlayerCnt
	if c.canPlay(other) {
		c.currentIdx = other
		return
	}
	// **「ゴー」の 1 点は 1 回の行き詰まりにつき 1 度だけ。** 相手が詰まった
	// あとも出し続けられるが、1 枚ごとに再び 1 点が入るわけではない ──
	// wentOut が立っている間は同じ行き詰まりの続きなので数えない。
	if c.canPlay(seat) {
		if c.wentOut != other {
			c.players[seat].AddScore(CostlyGoPoints)
			c.wentOut = other
			c.appendLog(other, "go", fmt.Sprintf("player %d says go; player %d pegs %d",
				other, seat, CostlyGoPoints), nil)
			c.checkGameEnd()
		}
		c.currentIdx = seat
		return
	}
	// どちらも出せない ── 数え上げを畳む。
	c.awardLatter(seat)
	if c.handsEmpty() {
		c.finishDeal()
		return
	}
	c.resetCount(seat)
}

// awardLatter は 31 に届かず打ち切ったときの 1 点を渡す。
func (c *CostlyColours) awardLatter(seat int) {
	if c.total == CostlyThirtyOne {
		return
	}
	c.players[seat].AddScore(CostlyLatterPoints)
	c.appendLog(seat, "latter",
		fmt.Sprintf("player %d pegs %d for the latter", seat, CostlyLatterPoints), nil)
	c.checkGameEnd()
}

// resetCount は数え上げを 0 に戻し、相手から再開する。
func (c *CostlyColours) resetCount(seat int) {
	c.pile = make([]*Card, 0, CostlyColoursDeckSize)
	c.total = 0
	c.wentOut = -1
	next := (seat + 1) % CostlyColoursPlayerCnt
	if !c.canPlay(next) {
		next = seat
	}
	c.currentIdx = next
}

// canPlay は席 seat が 31 を超えずに出せる札を持つかを返す。
func (c *CostlyColours) canPlay(seat int) bool {
	for _, card := range c.players[seat].GetHand() {
		if c.total+CostlyCardValue(card) <= CostlyThirtyOne {
			return true
		}
	}
	return false
}

// handsEmpty は全席の手札が尽きたかを返す。
func (c *CostlyColours) handsEmpty() bool {
	for _, p := range c.players {
		if p.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// PlayableIdxs は席 seat が出せる手札の位置を返す。
func (c *CostlyColours) PlayableIdxs(seat int) []int {
	if seat < 0 || seat >= len(c.players) || c.phase != CostlyColoursPhasePlay {
		return nil
	}
	out := make([]int, 0, CostlyColoursHandSize)
	for i, card := range c.players[seat].GetHand() {
		if c.total+CostlyCardValue(card) <= CostlyThirtyOne {
			out = append(out, i)
		}
	}
	return out
}

// finishDeal はショーを数える。
//
// **ショーは手札 3 枚 + 表の 1 枚。** 「4 枚同スート」の役が手札だけでは
// 成立しないので、Cribbage のスターターと同じく表の 1 枚が加わると読む。
// 手札は出し切って空なので、出した札を並べ直して数える。
func (c *CostlyColours) finishDeal() {
	res := &CostlyColoursDealResult{}
	trump := -1
	if c.turnUp != nil {
		trump = c.turnUp.GetDesign()
	}

	jackDeuce := make([]int, CostlyColoursPlayerCnt)
	rank := make([]int, CostlyColoursPlayerCnt)
	colour := make([]int, CostlyColoursPlayerCnt)

	for i, p := range c.players {
		// 自分の 3 枚 (出した札 + 残った札)。
		hand := append([]*Card(nil), p.GetPlayed()...)
		hand = append(hand, p.GetHand()...)
		// ショーで数える 4 枚 = 自分の 3 枚 + 表の 1 枚。
		show := append([]*Card(nil), hand...)
		if c.turnUp != nil {
			show = append(show, c.turnUp)
		}
		// **J / 2 の点は「手札にある」ぶんだけ。** 表に返った J は既に親の
		// 4 点 ("heels") になっているので、ここで二重に数えない。
		jackDeuce[i] = CostlyJackDeucePoints(hand, trump)
		_, rank[i] = CostlyRankCombo(show)
		combo, pts := CostlyColourCombo(show)
		colour[i], res.Combos[i] = pts, combo
	}

	res.Lines = []CostlyColoursScoreLine{
		{Key: "jackDeuce", Points: jackDeuce},
		{Key: "rank", Points: rank},
		{Key: "colour", Points: colour},
	}
	for _, line := range res.Lines {
		for i, pt := range line.Points {
			res.Totals[i] += pt
		}
	}
	// **非親から数える。** クリベッジ系の決まりで、非親の分だけで目標点に
	// 届いたらそこで終わり ── 親の手は数えない。打っている最中の点と同じで、
	// 「先に届いたほうが勝ち」を集計でも守る。
	c.lastResult = res
	c.appendLog(-1, "show", fmt.Sprintf("deal %d show: %d - %d",
		c.dealNumber, res.Totals[0], res.Totals[1]), nil)
	c.phase = CostlyColoursPhaseShow
	elder := (c.dealerIdx + 1) % CostlyColoursPlayerCnt
	for _, i := range []int{elder, c.dealerIdx} {
		c.players[i].AddScore(res.Totals[i])
		c.checkGameEnd()
		if c.gameEndFlag {
			return
		}
	}
}

// checkGameEnd は目標点に届いたかを見る。
//
// **点は打っている最中にも入る。** ペグボードのゲームなので、31 を作った
// 瞬間に勝ちが決まることがある ── ディールの切れ目でしか見ないと打ち過ぎる。
// 同点では終わらない。
func (c *CostlyColours) checkGameEnd() {
	if c.gameEndFlag {
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
	if best < c.config.TargetScore || tied {
		return
	}
	c.gameEndFlag = true
	c.winnerIdx = bestIdx
	c.phase = CostlyColoursPhaseGameEnd
	c.appendLog(-1, "gameEnd", fmt.Sprintf("player %d wins the match", bestIdx), nil)
}

// IsHumanTurn は人間の手番かを返す。
func (c *CostlyColours) IsHumanTurn() bool {
	human := findHumanIdx(c.players)
	if human < 0 || c.gameEndFlag {
		return false
	}
	switch c.phase {
	case CostlyColoursPhaseMog:
		return true
	case CostlyColoursPhasePlay:
		return c.currentIdx == human
	}
	return false
}

// GetConfig はゲーム設定を返す。
func (c *CostlyColours) GetConfig() CostlyColoursConfig { return c.config }

// SetConfig はゲーム設定を差し替える。
func (c *CostlyColours) SetConfig(cfg CostlyColoursConfig) { c.config = cfg }

// GetGameEndFlag は終局フラグを返す。
func (c *CostlyColours) GetGameEndFlag() bool { return c.gameEndFlag }

// GetPhase は現在のフェーズを返す。
func (c *CostlyColours) GetPhase() string { return c.phase }

// GetTurnUp は表に返した 1 枚を返す。
func (c *CostlyColours) GetTurnUp() *Card { return c.turnUp }

// GetPile は今の数え上げに出た札を返す。
func (c *CostlyColours) GetPile() []*Card { return c.pile }

// GetTotal は今の数え上げの累計を返す。
func (c *CostlyColours) GetTotal() int { return c.total }

// GetWentOut は「ゴー」を宣言した席を返す (-1 = なし)。
func (c *CostlyColours) GetWentOut() int { return c.wentOut }

// GetDealNumber は現在のディールを返す。
func (c *CostlyColours) GetDealNumber() int { return c.dealNumber }

// GetDealerIdx は親の席を返す。
func (c *CostlyColours) GetDealerIdx() int { return c.dealerIdx }

// GetCurrentPlayerIdx は手番の席を返す。
func (c *CostlyColours) GetCurrentPlayerIdx() int { return c.currentIdx }

// GetPlayerCnt は席数を返す。
func (c *CostlyColours) GetPlayerCnt() int { return len(c.players) }

// GetPlayer は指定席のプレイヤーを返す。
func (c *CostlyColours) GetPlayer(i int) *CostlyColoursPlayer {
	if i < 0 || i >= len(c.players) {
		return nil
	}
	return c.players[i]
}

// GetLastResult は直前のディールの集計を返す。
func (c *CostlyColours) GetLastResult() *CostlyColoursDealResult { return c.lastResult }

// GetWinnerIdx は勝者の席を返す (-1 = 未決)。
func (c *CostlyColours) GetWinnerIdx() int { return c.winnerIdx }

// SetTurnUpForTest は表の 1 枚を差し替える (テスト用)。
func (c *CostlyColours) SetTurnUpForTest(card *Card) { c.turnUp = card }

// SetTotalForTest は数え上げの累計を差し替える (テスト用)。
func (c *CostlyColours) SetTotalForTest(n int) { c.total = n }

// SetPhaseForTest はフェーズを差し替える (テスト用)。
func (c *CostlyColours) SetPhaseForTest(p string) { c.phase = p }

// SetCurrentForTest は手番の席を差し替える (テスト用)。
func (c *CostlyColours) SetCurrentForTest(i int) { c.currentIdx = i }

// costlyColoursJSON is the JSON wire format for CostlyColours.
type costlyColoursJSON struct {
	Deck        []*Card                  `json:"dk"`
	DrawIdx     int                      `json:"di"`
	Players     []*CostlyColoursPlayer   `json:"pl"`
	Config      CostlyColoursConfig      `json:"cf"`
	TurnUp      *Card                    `json:"tu"`
	Pile        []*Card                  `json:"pi"`
	Total       int                      `json:"to"`
	WentOut     int                      `json:"wo"`
	Phase       string                   `json:"ph"`
	DealNumber  int                      `json:"dn"`
	DealerIdx   int                      `json:"dl"`
	CurrentIdx  int                      `json:"cu"`
	LastResult  *CostlyColoursDealResult `json:"lr"`
	GameEndFlag bool                     `json:"ge"`
	WinnerIdx   int                      `json:"wi"`
	ActionLog   []*ActionLogEntry        `json:"al"`
}

// MarshalJSON implements json.Marshaler.
//
// **非公開フィールドだけの型は MarshalJSON が無いと `{}` になる。** 表の
// 1 枚が消えると、復元した盤ではショーの色役もトランプの J / 2 も数えられない。
func (c *CostlyColours) MarshalJSON() ([]byte, error) {
	return json.Marshal(costlyColoursJSON{
		Deck: c.deck, DrawIdx: c.drawIdx, Players: c.players, Config: c.config,
		TurnUp: c.turnUp, Pile: c.pile, Total: c.total, WentOut: c.wentOut,
		Phase: c.phase, DealNumber: c.dealNumber, DealerIdx: c.dealerIdx,
		CurrentIdx: c.currentIdx, LastResult: c.lastResult,
		GameEndFlag: c.gameEndFlag, WinnerIdx: c.winnerIdx, ActionLog: c.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *CostlyColours) UnmarshalJSON(data []byte) error {
	var j costlyColoursJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.players = j.Players
	if len(c.players) != CostlyColoursPlayerCnt {
		return fmt.Errorf("costlycolours: expected %d players, got %d",
			CostlyColoursPlayerCnt, len(c.players))
	}
	c.deck, c.drawIdx, c.config = j.Deck, j.DrawIdx, j.Config
	c.turnUp, c.pile, c.total, c.wentOut = j.TurnUp, j.Pile, j.Total, j.WentOut
	if c.pile == nil {
		c.pile = make([]*Card, 0)
	}
	c.phase, c.dealNumber = j.Phase, j.DealNumber
	c.dealerIdx, c.currentIdx = j.DealerIdx, j.CurrentIdx
	c.lastResult, c.gameEndFlag, c.winnerIdx = j.LastResult, j.GameEndFlag, j.WinnerIdx
	c.actionLog = j.ActionLog
	if c.actionLog == nil {
		c.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// FinishDealForTest はショーを数える (テスト用)。
func (c *CostlyColours) FinishDealForTest() { c.finishDeal() }
