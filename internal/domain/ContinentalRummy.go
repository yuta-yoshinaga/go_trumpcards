//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// フェーズ。**文字列で持つ。** 数値だと 0 が「未設定」と区別できない。
const (
	// ContinentalRummyPhaseDraw は山札か捨て札から 1 枚取る番。
	ContinentalRummyPhaseDraw = "draw"
	// ContinentalRummyPhaseDiscard は 1 枚捨てる番 (上がれるなら上がれる)。
	ContinentalRummyPhaseDiscard = "discard"
	// ContinentalRummyPhaseRoundEnd はラウンドの決着を見せている状態。
	ContinentalRummyPhaseRoundEnd = "roundEnd"
	// ContinentalRummyPhaseGameEnd は終局。
	ContinentalRummyPhaseGameEnd = "gameEnd"
)

// ContinentalRummyHumanIdx は人間の席。
const ContinentalRummyHumanIdx = 0

// 精算のボーナス。**負けた側の残り札は数えない。**
//
// この形式は勝った側が各相手から取り立てる方式で、点数は「どう上がったか」で
// 決まる。#5464 の「残存カードの点数を失点として加算し、最少失点が勝ち」は
// 別のラミーの精算で、ここには無い。
const (
	// ContinentalRummyWinPoints は上がりそのものの 1 点。
	ContinentalRummyWinPoints = 1
	// ContinentalRummyJokerPoints はメルドに使ったジョーカー 1 枚あたり。
	ContinentalRummyJokerPoints = 2
	// ContinentalRummyFirstTurnPoints は最初の手番で上がったときの加点。
	ContinentalRummyFirstTurnPoints = 7
	// ContinentalRummyDealtPoints は配られた 15 枚のまま上がったときの加点。
	ContinentalRummyDealtPoints = 10
	// ContinentalRummyNoJokerPoints はジョーカーを 1 枚も使わなかったときの加点。
	ContinentalRummyNoJokerPoints = 10
	// ContinentalRummyOneSuitPoints は 15 枚すべてが同じスートだったときの加点。
	ContinentalRummyOneSuitPoints = 10
)

// ContinentalRummyBonus は上がった手に付いた加点の内訳 1 行。
type ContinentalRummyBonus struct {
	// Key は加点の種類 ("win" / "joker" / "firstTurn" / "dealt" / "noJoker" / "oneSuit")。
	Key string `json:"k"`
	// Points はその加点。
	Points int `json:"p"`
}

// ContinentalRummyRoundResult は 1 ラウンドの決着。
type ContinentalRummyRoundResult struct {
	// WinnerIdx は上がった席。誰も上がらずに山が尽きたら -1。
	WinnerIdx int `json:"w"`
	// Bonuses は加点の内訳。
	Bonuses []ContinentalRummyBonus `json:"b"`
	// PerOpponent は相手 1 人あたりの取り立て額。
	PerOpponent int `json:"po"`
	// Total は上がった側が受け取った合計。
	Total int `json:"t"`
}

// ContinentalRummy はコンチネンタル・ラミーのゲーム。
//
// **セットはメルドにならず、途中で場に出すこともできない。** 15 枚を認められた
// 形に一度で並べて上がるか、何も出さないかのどちらかしかない。
type ContinentalRummy struct {
	actionLogBase

	players     []*ContinentalRummyPlayer
	stock       []*Card
	discardPile []*Card
	phase       string
	currentIdx  int
	dealerIdx   int
	roundNumber int
	// drewThisRound は席ごとに、このラウンドで一度でも引いたか。
	// 「配られた 15 枚のまま上がった」の判定に要る。
	drewThisRound []bool
	// recycles はこのラウンドで捨て札を山に戻した回数。
	recycles    int
	lastResult  *ContinentalRummyRoundResult
	gameEndFlag bool
	winnerIdx   int
	config      ContinentalRummyConfig
}

// NewDefaultContinentalRummy は既定の設定でゲームを作る。
func NewDefaultContinentalRummy() *ContinentalRummy {
	return NewContinentalRummy(DefaultContinentalRummyConfig())
}

// NewContinentalRummy はコンストラクタ。
func NewContinentalRummy(cfg ContinentalRummyConfig) *ContinentalRummy {
	c := &ContinentalRummy{config: cfg, winnerIdx: -1}
	c.players = make([]*ContinentalRummyPlayer, ContinentalRummyPlayerCnt)
	for i := range c.players {
		c.players[i] = NewContinentalRummyPlayer(i == ContinentalRummyHumanIdx)
	}
	return c
}

// Reset はゲームを最初から始める。
func (c *ContinentalRummy) Reset() {
	for _, p := range c.players {
		p.ResetRound()
		p.AddScore(-p.GetScore())
	}
	c.roundNumber = 0
	c.dealerIdx = ContinentalRummyPlayerCnt - 1
	c.gameEndFlag = false
	c.winnerIdx = -1
	c.lastResult = nil
	c.actionLog = nil
	c.startRound()
}

// startRound は 1 ラウンドぶんを配る。
func (c *ContinentalRummy) startRound() {
	c.roundNumber++
	c.dealerIdx = (c.dealerIdx + 1) % ContinentalRummyPlayerCnt
	c.lastResult = nil
	c.drewThisRound = make([]bool, ContinentalRummyPlayerCnt)
	c.recycles = 0
	for _, p := range c.players {
		p.ResetRound()
	}

	// **2〜5 人卓は 2 組 + ジョーカー 2 枚。** 人数 − 1 組ではない。
	deck := NewTrumpCardsWithDecks(ContinentalRummyDeckCnt, ContinentalRummyJokerCnt)
	deck.Shuffle()
	c.stock = make([]*Card, 0, deck.GetTotalCount())
	for card := deck.DrawCard(); card != nil; card = deck.DrawCard() {
		c.stock = append(c.stock, card)
	}

	// **3 枚ずつ 5 回配る。** 1 枚ずつ 15 回ではない (原典の配り方)。
	for round := 0; round < ContinentalRummyHandSize/ContinentalRummyDealChunk; round++ {
		for i := 0; i < ContinentalRummyPlayerCnt; i++ {
			seat := (c.dealerIdx + 1 + i) % ContinentalRummyPlayerCnt
			for k := 0; k < ContinentalRummyDealChunk; k++ {
				c.players[seat].AddCard(c.drawStock())
			}
		}
	}
	c.discardPile = []*Card{c.drawStock()}
	c.currentIdx = (c.dealerIdx + 1) % ContinentalRummyPlayerCnt
	c.phase = ContinentalRummyPhaseDraw
	c.appendLog(-1, "deal", fmt.Sprintf("round %d dealt, %d in stock", c.roundNumber, len(c.stock)), nil)
	c.runCpuTurns()
}

// drawStock は山札から 1 枚取る。尽きていれば捨て札を裏返して積み直す。
//
// **原典は山が尽きたときのことを書いていない。** 書いていないからといって
// そこで流局にすると、実測で **200 局のうち 161 局が誰も上がれずに流れた** ──
// 15 枚を 3〜5 枚の並びだけで揃える形式では、45 枚の山は 1 周で足りない。
// 誰も届かない目標を出し続けるくらいなら、ラミー系で普通に行われている
// 「捨て札を裏返して積み直す」を採る。頭の 1 枚は場に残す。
//
// 積み直しは continentalRummyMaxRecycles 回まで。そこで打ち切るのは、
// 上限の無いループを Worker に持ち込まないため。
func (c *ContinentalRummy) drawStock() *Card {
	if len(c.stock) == 0 {
		c.recycleDiscards()
	}
	if len(c.stock) == 0 {
		return nil
	}
	card := c.stock[0]
	c.stock = c.stock[1:]
	return card
}

// continentalRummyMaxRecycles は 1 ラウンドで捨て札を積み直せる回数。
const continentalRummyMaxRecycles = 3

// recycleDiscards は捨て札 (頭の 1 枚を除く) を混ぜて山に戻す。
func (c *ContinentalRummy) recycleDiscards() {
	if c.recycles >= continentalRummyMaxRecycles || len(c.discardPile) < 2 {
		return
	}
	top := c.discardPile[len(c.discardPile)-1]
	rest := c.discardPile[:len(c.discardPile)-1]
	shuffled := make([]*Card, len(rest))
	copy(shuffled, rest)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	c.stock = shuffled
	c.discardPile = []*Card{top}
	c.recycles++
	c.appendLog(-1, "recycle",
		fmt.Sprintf("discards turned back into a stock of %d (recycle %d/%d)",
			len(c.stock), c.recycles, continentalRummyMaxRecycles), nil)
}

// DrawStock は山札から 1 枚取る。
func (c *ContinentalRummy) DrawStock() error { return c.draw(true) }

// DrawDiscard は捨て札の一番上を取る。
func (c *ContinentalRummy) DrawDiscard() error { return c.draw(false) }

// draw は 1 枚引く。**山と捨て札を別々の入口にする**のは、既定のまま届いた
// 要求がどちらかに黙って倒れないようにするため。
func (c *ContinentalRummy) draw(fromStock bool) error {
	if err := c.guardTurn(ContinentalRummyPhaseDraw); err != nil {
		return err
	}
	var card *Card
	if fromStock {
		card = c.drawStock()
		if card == nil {
			// **山が尽きたらラウンドは流れる。** 捨て札を混ぜ直したりはしない。
			c.finishRound(-1)
			return nil
		}
	} else {
		if len(c.discardPile) == 0 {
			return NewDomainErrorCode(ErrDeckExhausted, "continentalrummy.errDiscardEmpty", nil)
		}
		card = c.discardPile[len(c.discardPile)-1]
		c.discardPile = c.discardPile[:len(c.discardPile)-1]
	}
	c.players[c.currentIdx].AddCard(card)
	c.drewThisRound[c.currentIdx] = true
	c.phase = ContinentalRummyPhaseDiscard
	return nil
}

// Discard は i 番目の手札を捨てて次の席へ渡す。
func (c *ContinentalRummy) Discard(i int) error {
	if err := c.guardTurn(ContinentalRummyPhaseDiscard); err != nil {
		return err
	}
	card := c.players[c.currentIdx].RemoveCard(i)
	if card == nil {
		return NewDomainErrorCode(ErrInvalidCard, "continentalrummy.errCardRange",
			map[string]string{"val": fmt.Sprint(i)})
	}
	c.discardPile = append(c.discardPile, card)
	c.advance()
	return nil
}

// GoOut は 15 枚を並べて上がる。捨てる 1 枚を i で指定する。
//
// **部分メルドは無い。** 残り 15 枚が認められた形にならなければ断る。
func (c *ContinentalRummy) GoOut(i int) error {
	if err := c.guardTurn(ContinentalRummyPhaseDiscard); err != nil {
		return err
	}
	p := c.players[c.currentIdx]
	if i < 0 || i >= p.GetCardsSize() {
		return NewDomainErrorCode(ErrInvalidCard, "continentalrummy.errCardRange",
			map[string]string{"val": fmt.Sprint(i)})
	}
	if !c.layDownAndGoOut(c.currentIdx, i) {
		return NewDomainErrorCode(ErrInvalidPlay, "continentalrummy.errNotAGoOut", nil)
	}
	return nil
}

// layDownAndGoOut は 15 枚を並べて上がる。上がれない手なら false。
//
// **並べた札は手札から消える。** 場に出したのに手元にも残っていると、次に
// 誰かが数えたときに 15 枚が二重に数えられる。人間と CPU で同じ道を通す。
func (c *ContinentalRummy) layDownAndGoOut(seat, discardIdx int) bool {
	p := c.players[seat]
	rest := make([]*Card, 0, ContinentalRummyHandSize)
	for j := 0; j < p.GetCardsSize(); j++ {
		if j != discardIdx {
			rest = append(rest, p.GetCard(j))
		}
	}
	groups, ok := FindContinentalRummyGoOut(rest)
	if !ok {
		return false
	}
	melds := make([][]*Card, 0, len(groups))
	for _, g := range groups {
		run := make([]*Card, 0, len(g))
		for _, idx := range g {
			run = append(run, rest[idx])
		}
		melds = append(melds, run)
	}
	p.SetMelds(melds)
	c.discardPile = append(c.discardPile, p.GetCard(discardIdx))
	p.ClearHand()
	c.finishRound(seat)
	return true
}

// guardTurn は終局・フェーズ・手番をまとめて見る。
func (c *ContinentalRummy) guardTurn(want string) error {
	if c.gameEndFlag {
		return NewDomainErrorCode(ErrGameEnded, "continentalrummy.errGameEnded", nil)
	}
	if c.phase != want {
		return NewDomainErrorCode(ErrWrongPhase, "continentalrummy.errWrongPhase", nil)
	}
	if c.currentIdx != ContinentalRummyHumanIdx {
		return NewDomainErrorCode(ErrNotHumanTurn, "continentalrummy.errNotYourTurn", nil)
	}
	return nil
}

// advance は手番を次の席へ渡し、CPU の番を進める。
func (c *ContinentalRummy) advance() {
	c.currentIdx = (c.currentIdx + 1) % ContinentalRummyPlayerCnt
	c.phase = ContinentalRummyPhaseDraw
	c.runCpuTurns()
}

// finishRound はラウンドを畳んで精算する。winner < 0 なら流局。
func (c *ContinentalRummy) finishRound(winner int) {
	c.phase = ContinentalRummyPhaseRoundEnd
	res := &ContinentalRummyRoundResult{WinnerIdx: winner}
	if winner >= 0 {
		res.Bonuses = c.bonusesFor(winner)
		for _, b := range res.Bonuses {
			res.PerOpponent += b.Points
		}
		// **勝った側が各相手から取り立てる。** 負けた側の残り札は数えない。
		res.Total = res.PerOpponent * (ContinentalRummyPlayerCnt - 1)
		c.players[winner].AddScore(res.Total)
		c.appendLog(winner, "goOut",
			fmt.Sprintf("seat %d goes out for %d from each of %d opponent(s)",
				winner, res.PerOpponent, ContinentalRummyPlayerCnt-1), nil)
	} else {
		c.appendLog(-1, "washout", "stock exhausted with nobody out", nil)
	}
	c.lastResult = res
	if c.roundNumber >= c.config.TotalRounds {
		c.endGame()
	}
}

// bonusesFor は上がった手に付く加点の内訳を返す。
func (c *ContinentalRummy) bonusesFor(seat int) []ContinentalRummyBonus {
	p := c.players[seat]
	out := []ContinentalRummyBonus{{Key: "win", Points: ContinentalRummyWinPoints}}

	jokers, suits := 0, map[int]bool{}
	for _, run := range p.GetMelds() {
		for _, card := range run {
			if IsContinentalRummyJoker(card) {
				jokers++
				continue
			}
			suits[card.GetDesign()] = true
		}
	}
	if jokers > 0 {
		out = append(out, ContinentalRummyBonus{Key: "joker", Points: jokers * ContinentalRummyJokerPoints})
	} else {
		out = append(out, ContinentalRummyBonus{Key: "noJoker", Points: ContinentalRummyNoJokerPoints})
	}
	// **配られた 15 枚のまま上がるのと、1 手番で上がるのは別の加点。**
	// 前者は一度も引いていないこと、後者は最初の手番で決めたこと。
	if !c.drewThisRound[seat] {
		out = append(out, ContinentalRummyBonus{Key: "dealt", Points: ContinentalRummyDealtPoints})
	} else if c.roundTurnsTaken(seat) == 1 {
		out = append(out, ContinentalRummyBonus{Key: "firstTurn", Points: ContinentalRummyFirstTurnPoints})
	}
	if len(suits) == 1 {
		out = append(out, ContinentalRummyBonus{Key: "oneSuit", Points: ContinentalRummyOneSuitPoints})
	}
	return out
}

// roundTurnsTaken はその席がこのラウンドで何手番打ったかを棋譜から数える。
func (c *ContinentalRummy) roundTurnsTaken(seat int) int {
	n := 0
	for _, e := range c.GetActionLog() {
		if e.PlayerIdx == seat && e.ActionType == "discard" {
			n++
		}
	}
	return n + 1 // いま打っている手番ぶん
}

// NextRound は次のラウンドへ進む。
func (c *ContinentalRummy) NextRound() {
	if c.gameEndFlag || c.phase != ContinentalRummyPhaseRoundEnd {
		return
	}
	c.startRound()
}

// endGame は終局して勝者を決める。**一番多く集めた席の勝ち。**
func (c *ContinentalRummy) endGame() {
	c.gameEndFlag = true
	c.phase = ContinentalRummyPhaseGameEnd
	best, bestIdx := -1, -1
	tie := false
	for i, p := range c.players {
		switch {
		case p.GetScore() > best:
			best, bestIdx, tie = p.GetScore(), i, false
		case p.GetScore() == best:
			tie = true
		}
	}
	if tie {
		bestIdx = -1
	}
	c.winnerIdx = bestIdx
	c.appendLog(-1, "gameEnd", fmt.Sprintf("game ends after %d round(s)", c.roundNumber), nil)
}

// アクセサ。
func (c *ContinentalRummy) GetPhase() string         { return c.phase }
func (c *ContinentalRummy) GetPlayerCnt() int        { return len(c.players) }
func (c *ContinentalRummy) GetCurrentPlayerIdx() int { return c.currentIdx }
func (c *ContinentalRummy) GetDealerIdx() int        { return c.dealerIdx }
func (c *ContinentalRummy) GetRoundNumber() int      { return c.roundNumber }
func (c *ContinentalRummy) GetStockCount() int       { return len(c.stock) }

// GetRecycleCount はこのラウンドで捨て札を積み直した回数を返す。
func (c *ContinentalRummy) GetRecycleCount() int                        { return c.recycles }
func (c *ContinentalRummy) GetGameEndFlag() bool                        { return c.gameEndFlag }
func (c *ContinentalRummy) GetWinnerIdx() int                           { return c.winnerIdx }
func (c *ContinentalRummy) GetConfig() ContinentalRummyConfig           { return c.config }
func (c *ContinentalRummy) SetConfig(cfg ContinentalRummyConfig)        { c.config = cfg }
func (c *ContinentalRummy) GetLastResult() *ContinentalRummyRoundResult { return c.lastResult }

// GetPlayer は i 番目の席を返す。範囲外なら nil。
func (c *ContinentalRummy) GetPlayer(i int) *ContinentalRummyPlayer {
	if i < 0 || i >= len(c.players) {
		return nil
	}
	return c.players[i]
}

// GetDiscardTop は捨て札の一番上を返す。空なら nil。
func (c *ContinentalRummy) GetDiscardTop() *Card {
	if len(c.discardPile) == 0 {
		return nil
	}
	return c.discardPile[len(c.discardPile)-1]
}

// IsHumanTurn は人間の番かを返す。
func (c *ContinentalRummy) IsHumanTurn() bool {
	return !c.gameEndFlag && c.currentIdx == ContinentalRummyHumanIdx &&
		(c.phase == ContinentalRummyPhaseDraw || c.phase == ContinentalRummyPhaseDiscard)
}

// CanGoOut は人間がいま上がれるかを返す。上がれるなら捨てる札の候補も返す。
func (c *ContinentalRummy) CanGoOut() (int, bool) {
	if c.phase != ContinentalRummyPhaseDiscard {
		return -1, false
	}
	return continentalGoOutDiscard(c.players[c.currentIdx].GetHand())
}

// continentalGoOutDiscard は 16 枚から「これを捨てれば上がれる」1 枚を探す。
func continentalGoOutDiscard(hand []*Card) (int, bool) {
	if len(hand) != ContinentalRummyHandSize+1 {
		return -1, false
	}
	for i := range hand {
		rest := make([]*Card, 0, ContinentalRummyHandSize)
		for j, card := range hand {
			if j != i {
				rest = append(rest, card)
			}
		}
		if CanContinentalRummyGoOut(rest) {
			return i, true
		}
	}
	return -1, false
}

// SetStockForTest はテスト用に山札を差し替える。
func (c *ContinentalRummy) SetStockForTest(cards []*Card) { c.stock = cards }

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (c *ContinentalRummy) SetPhaseForTest(p string) { c.phase = p }

// SetCurrentIdxForTest はテスト用に手番を差し替える。
func (c *ContinentalRummy) SetCurrentIdxForTest(i int) { c.currentIdx = i }

// continentalRummyJSON は保存用の姿。
//
// **非公開フィールドしか無い型は MarshalJSON が無いと `{}` になる。**
type continentalRummyJSON struct {
	Hands       [][]*Card                    `json:"h"`
	Melds       [][][]*Card                  `json:"m"`
	Scores      []int                        `json:"s"`
	Stock       []*Card                      `json:"st"`
	Discard     []*Card                      `json:"d"`
	Phase       string                       `json:"p"`
	CurrentIdx  int                          `json:"ci"`
	DealerIdx   int                          `json:"di"`
	RoundNumber int                          `json:"rn"`
	Drew        []bool                       `json:"dr"`
	Recycles    int                          `json:"rc"`
	LastResult  *ContinentalRummyRoundResult `json:"lr"`
	GameEndFlag bool                         `json:"ge"`
	WinnerIdx   int                          `json:"wi"`
	Config      ContinentalRummyConfig       `json:"c"`
	ActionLog   []*ActionLogEntry            `json:"al"`
}

// MarshalJSON は盤面を JSON にする。
func (c *ContinentalRummy) MarshalJSON() ([]byte, error) {
	j := continentalRummyJSON{
		Stock: c.stock, Discard: c.discardPile, Phase: c.phase,
		CurrentIdx: c.currentIdx, DealerIdx: c.dealerIdx, RoundNumber: c.roundNumber,
		Drew: c.drewThisRound, Recycles: c.recycles, LastResult: c.lastResult, GameEndFlag: c.gameEndFlag,
		WinnerIdx: c.winnerIdx, Config: c.config, ActionLog: c.GetActionLog(),
	}
	for _, p := range c.players {
		j.Hands = append(j.Hands, p.GetHand())
		j.Melds = append(j.Melds, p.GetMelds())
		j.Scores = append(j.Scores, p.GetScore())
	}
	return json.Marshal(j)
}

// UnmarshalJSON は JSON から盤面を戻す。
func (c *ContinentalRummy) UnmarshalJSON(data []byte) error {
	var j continentalRummyJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.players = make([]*ContinentalRummyPlayer, ContinentalRummyPlayerCnt)
	for i := range c.players {
		c.players[i] = NewContinentalRummyPlayer(i == ContinentalRummyHumanIdx)
		if i < len(j.Hands) {
			for _, card := range j.Hands[i] {
				c.players[i].AddCard(card)
			}
		}
		if i < len(j.Melds) {
			c.players[i].SetMelds(j.Melds[i])
		}
		if i < len(j.Scores) {
			c.players[i].AddScore(j.Scores[i])
		}
	}
	c.stock, c.discardPile, c.phase = j.Stock, j.Discard, j.Phase
	c.currentIdx, c.dealerIdx, c.roundNumber = j.CurrentIdx, j.DealerIdx, j.RoundNumber
	c.drewThisRound, c.recycles, c.lastResult = j.Drew, j.Recycles, j.LastResult
	c.gameEndFlag, c.winnerIdx, c.config = j.GameEndFlag, j.WinnerIdx, j.Config
	c.actionLog = j.ActionLog
	return nil
}

// discardCountForTest はテスト用に捨て札の枚数を返す。
func (c *ContinentalRummy) discardCountForTest() int { return len(c.discardPile) }

// SetDiscardForTest はテスト用に捨て札を差し替える。
func (c *ContinentalRummy) SetDiscardForTest(cards []*Card) { c.discardPile = cards }
