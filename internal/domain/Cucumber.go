//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// CucumberPhase はゲームフェーズ。
type CucumberPhase int

// キューカンバーのフェーズ定数
const (
	// CucumberPhasePlay プレイ中
	CucumberPhasePlay CucumberPhase = 0
	// CucumberPhaseRoundEnd ラウンド終了 (最終トリックの失点が付いた)
	CucumberPhaseRoundEnd CucumberPhase = 1
	// CucumberPhaseGameEnd ゲーム終了
	CucumberPhaseGameEnd CucumberPhase = 2
)

// CucumberPhaseMin / Max はフェーズ列挙の範囲 (復元時の検証用)。
const (
	CucumberPhaseMin = CucumberPhasePlay
	CucumberPhaseMax = CucumberPhaseGameEnd
)

// cucumberMaxSliceLen は復元時に受け付けるスライス長の上限。
const cucumberMaxSliceLen = 4000

// CucumberHint は人間への助言。
type CucumberHint struct {
	// CardIndex は出すべき手札の位置。
	CardIndex *int
	// Reason は助言の理由。
	Reason string
}

// cucumberRank はランクを 2..14 で返す (A が最強)。
func cucumberRank(c *Card) int {
	if c == nil {
		return 0
	}
	if v := c.GetValue(); v == 1 {
		return 14
	} else {
		return v
	}
}

// Cucumber はキューカンバー (Cucumber / Gurka) のゲームクラス。
//
// スウェーデン・フィンランドのトリックテイキング。**スートは一切関係ありません。**
// フォローの規則が独特で、**そのトリックでいま出ている最高ランクより高い札を持って
// いれば必ずそれを出し、無ければ手札のいちばん低い札を出します**。選べるのは
// 「どの高い札を出すか」だけで、出さない自由はありません。
//
// # 罰は最後の 1 トリックだけ
//
// 7 トリック戦いますが、**失点が付くのは最終トリックを取った 1 人だけ**で、その
// トリックを取った札のランクぶん失点します。途中のトリックは何枚取っても
// 0 点——**終盤に高い札を残していることだけが危険**という、逆算の効く形です。
//
// # 配り
//
// **人数で割ろうとしないこと。** 52 枚は 3 / 5 / 6 人で割り切れないので、伝統
// どおり [CucumberHandSize] 枚固定で配り、残りは使いません。
//
// # 停止保証
//
// 最終トリックのランクは必ず 2 以上なので、**毎ラウンド誰かの失点が必ず増えます**。
// 失点は単調に増えるだけなので、いずれ [CucumberConfig.TargetScore] に到達します。
type Cucumber struct {
	trumpCards *TrumpCards
	players    []*CucumberPlayer
	config     CucumberConfig
	phase      CucumberPhase

	currentTrick     []*TrickCard
	leadPlayerIdx    int
	currentPlayerIdx int
	trickNumber      int
	roundNumber      int
	// lastTrickWinnerIdx は直前ラウンドで最終トリックを取った席 (-1 = 未)。
	lastTrickWinnerIdx int
	// lastPenalty は直前ラウンドで付いた失点。
	lastPenalty int
	gameEndFlag bool
	winnerIdx   int
	actionLogBase
}

// NewCucumber はコンストラクタ。
func NewCucumber(players []*CucumberPlayer, config CucumberConfig) *Cucumber {
	if config.Validate() != nil {
		config = DefaultCucumberConfig()
	}
	if players == nil {
		players = newCucumberSeats(config.PlayerCnt)
	}
	return &Cucumber{
		players:            players,
		config:             config,
		lastTrickWinnerIdx: -1,
		winnerIdx:          -1,
	}
}

// newCucumberSeats は席 0 を人間、以降を CPU とした座席を作る。
func newCucumberSeats(n int) []*CucumberPlayer {
	seats := make([]*CucumberPlayer, n)
	for i := range seats {
		seats[i] = NewCucumberPlayer(i == 0)
	}
	return seats
}

// NewDefaultCucumber は既定設定の Cucumber を返す。
func NewDefaultCucumber() *Cucumber {
	cfg := DefaultCucumberConfig()
	return NewCucumber(newCucumberSeats(cfg.PlayerCnt), cfg)
}

// Reset はゲームを初期化する。
func (c *Cucumber) Reset() {
	for _, p := range c.players {
		p.ResetGame()
	}
	c.roundNumber = 0
	c.lastTrickWinnerIdx = -1
	c.lastPenalty = 0
	c.gameEndFlag = false
	c.winnerIdx = -1
	c.actionLog = nil
	c.addLog(-1, "start", fmt.Sprintf("キューカンバーを開始しました（%d 人、%d 点で終了）",
		c.config.PlayerCnt, c.config.TargetScore), nil)
	c.dealRound()
}

// dealRound は 1 ラウンド分を配る。
func (c *Cucumber) dealRound() {
	c.roundNumber++
	c.phase = CucumberPhasePlay
	c.currentTrick = nil
	c.trickNumber = 0

	c.trumpCards = NewTrumpCards(0)
	c.trumpCards.Shuffle()
	for _, p := range c.players {
		p.Reset()
	}
	// **7 枚固定。** 割り切れない配りを避けるための伝統的な枚数です。
	for range CucumberHandSize {
		for _, p := range c.players {
			if card := c.trumpCards.DrawCard(); card != nil {
				p.AddCard(card)
			}
		}
	}
	c.sortAllHands()

	// 前ラウンドの最終トリックを取った席が次のリード (初回は席 0)。
	lead := 0
	if c.lastTrickWinnerIdx >= 0 {
		lead = c.lastTrickWinnerIdx
	}
	c.leadPlayerIdx = lead
	c.currentPlayerIdx = lead
	c.addLog(-1, "deal", fmt.Sprintf("ラウンド %d を配りました", c.roundNumber), nil)
}

// sortAllHands は手札をランク順に整える。
func (c *Cucumber) sortAllHands() {
	for _, p := range c.players {
		sortPlayerHand(p, func(ci, cj *Card) bool {
			if cucumberRank(ci) != cucumberRank(cj) {
				return cucumberRank(ci) < cucumberRank(cj)
			}
			return ci.GetDesign() < cj.GetDesign()
		})
	}
}

// HighestInTrick はいまトリックに出ている最高ランクを返す (0 = まだ無い)。
func (c *Cucumber) HighestInTrick() int {
	best := 0
	for _, tc := range c.currentTrick {
		if r := cucumberRank(tc.Card); r > best {
			best = r
		}
	}
	return best
}

// GetValidPlayIndices は出せる手札の位置を返す。
//
// **選べるのは「どの高い札を出すか」だけ。** 更新できる札を 1 枚でも持っていれば
// その中から選ぶ義務があり、1 枚も無ければいちばん低い札に決まります。
func (c *Cucumber) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(c.players) {
		return nil
	}
	p := c.players[playerIdx]
	all := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		all = append(all, i)
	}
	if len(all) == 0 {
		return nil
	}
	// リードは何を出してもよい。
	high := c.HighestInTrick()
	if high == 0 {
		return all
	}
	if above := filterAbove(p, all, high, cucumberRank); len(above) > 0 {
		return above
	}
	// **更新できないなら、いちばん低い札 1 枚だけが合法。**
	if low := pickLowest(p, all, cucumberRank); low >= 0 {
		return []int{low}
	}
	return nil
}

// IsHumanTurn は現在の手番が人間かを返す。
func (c *Cucumber) IsHumanTurn() bool {
	if c.gameEndFlag {
		return false
	}
	switch c.phase {
	case CucumberPhasePlay:
		return c.players[c.currentPlayerIdx].GetIsHuman()
	case CucumberPhaseRoundEnd:
		// **失点が付いた事実を読む時間。** 押されるまで次は配りません。
		return true
	default:
		return false
	}
}

// PlayerPlay は人間が 1 枚出す。
func (c *Cucumber) PlayerPlay(cardIndex int) error {
	if c.phase != CucumberPhasePlay {
		return errors.New("いまは札を出す場面ではありません")
	}
	if !c.players[c.currentPlayerIdx].GetIsHuman() {
		return errors.New("あなたの番ではありません")
	}
	return c.play(c.currentPlayerIdx, cardIndex)
}

// NextRound はラウンド終了状態から次のラウンドを配る。
func (c *Cucumber) NextRound() error {
	if c.gameEndFlag {
		return errors.New("ゲームは終了しています")
	}
	if c.phase != CucumberPhaseRoundEnd {
		return errors.New("いまはラウンドの区切りではありません")
	}
	c.dealRound()
	return nil
}

// CpuPlay は CPU が 1 枚出す。
func (c *Cucumber) CpuPlay() {
	if c.gameEndFlag || c.phase != CucumberPhasePlay || c.IsHumanTurn() {
		return
	}
	idx := c.currentPlayerIdx
	_ = c.play(idx, c.chooseCpuCard(idx))
}

// play は席 playerIdx に cardIndex を出させる。
func (c *Cucumber) play(playerIdx, cardIndex int) error {
	if playerIdx != c.currentPlayerIdx {
		return fmt.Errorf("いまは席 %d の番ではありません", playerIdx)
	}
	valid := c.GetValidPlayIndices(playerIdx)
	if !cucumberContains(valid, cardIndex) {
		p := c.players[playerIdx]
		if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
			return fmt.Errorf("手札の位置が範囲外です: %d", cardIndex)
		}
		// **出せないのではなく、出す札が決まっている。** 理由を言い分けます。
		//
		// **合法手が 1 つ = 更新できない、ではありません。** ちょうど 1 枚だけ
		// 更新できる形もあるので、その札が実際に更新するかどうかで見分けます。
		if c.forcedLowest(playerIdx, valid) {
			return errors.New("更新できる札がないので、いちばん低い札を出してください")
		}
		return errors.New("いま出ている最高ランクを超える札を出してください")
	}

	card := c.players[playerIdx].RemoveCard(cardIndex)
	c.currentTrick = append(c.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	c.addLog(playerIdx, "play", "カードを出しました", []*Card{card})

	if len(c.currentTrick) < len(c.players) {
		c.currentPlayerIdx = (c.currentPlayerIdx + 1) % len(c.players)
		return nil
	}
	c.resolveTrick()
	return nil
}

// IsForcedLowest は席 playerIdx が「更新できないので低い札に決まっている」かを返す。
//
// **この判定を各層で書き直さないこと。** 「合法手が 1 つ = 更新できない」は偽で、
// 実際そこで一度間違えました。判定はここ 1 か所に置き、UI は結果だけを受け取ります。
func (c *Cucumber) IsForcedLowest(playerIdx int) bool {
	if c.phase != CucumberPhasePlay {
		return false
	}
	return c.forcedLowest(playerIdx, c.GetValidPlayIndices(playerIdx))
}

// forcedLowest は「更新できないので低い札に決まっている」場面かを返す。
func (c *Cucumber) forcedLowest(playerIdx int, valid []int) bool {
	high := c.HighestInTrick()
	if high == 0 || len(valid) != 1 {
		return false
	}
	return cucumberRank(c.players[playerIdx].GetCard(valid[0])) <= high
}

// resolveTrick はトリックを解決する。
//
// **スートは関係ありません。** 出た札のうち最高ランクが勝ち、同ランクなら
// 先に出したほうが勝ちます。
func (c *Cucumber) resolveTrick() {
	winner := c.trickWinner()
	c.trickNumber++
	c.addLog(winner, "trick", fmt.Sprintf("%d 番目のトリックを取りました", c.trickNumber), nil)

	if c.players[winner].GetCardsSize() == 0 {
		// **失点が付くのは最終トリックだけ。**
		c.finishRound(winner)
		return
	}
	c.currentTrick = nil
	c.leadPlayerIdx = winner
	c.currentPlayerIdx = winner
}

// trickWinner は最高ランクを出した席を返す (同ランクは先着)。
func (c *Cucumber) trickWinner() int {
	best, bestRank := c.currentTrick[0].PlayerIdx, cucumberRank(c.currentTrick[0].Card)
	for _, tc := range c.currentTrick[1:] {
		if r := cucumberRank(tc.Card); r > bestRank {
			best, bestRank = tc.PlayerIdx, r
		}
	}
	return best
}

// finishRound は最終トリックの失点を付けてラウンドを終える。
func (c *Cucumber) finishRound(winner int) {
	// **取った札のランクぶん失点する。** 高い札で取るほど痛い。
	penalty := 0
	for _, tc := range c.currentTrick {
		if tc.PlayerIdx == winner {
			if r := cucumberRank(tc.Card); r > penalty {
				penalty = r
			}
		}
	}
	c.players[winner].AddPenalty(penalty)
	c.lastTrickWinnerIdx = winner
	c.lastPenalty = penalty
	c.phase = CucumberPhaseRoundEnd
	c.addLog(winner, "penalty", fmt.Sprintf("最終トリックを取り %d 点の失点", penalty), nil)

	if c.reachedTarget() {
		c.finish()
	}
}

// reachedTarget は誰かが失点上限に達したかを返す。
func (c *Cucumber) reachedTarget() bool {
	for _, p := range c.players {
		if p.GetPenalty() >= c.config.TargetScore {
			return true
		}
	}
	return false
}

// finish は失点のいちばん少ない席を勝ちにして終局する。
func (c *Cucumber) finish() {
	c.phase = CucumberPhaseGameEnd
	c.gameEndFlag = true
	c.winnerIdx = c.leaderIdx()
	c.addLog(c.winnerIdx, "result",
		fmt.Sprintf("失点 %d 点でいちばん少なく終えました", c.players[c.winnerIdx].GetPenalty()), nil)
}

// leaderIdx は失点のいちばん少ない席を返す (同点なら若い席)。
func (c *Cucumber) leaderIdx() int {
	best := 0
	for i, p := range c.players {
		if p.GetPenalty() < c.players[best].GetPenalty() {
			best = i
		}
	}
	return best
}

// GiveUp は投了する。
func (c *Cucumber) GiveUp() {
	if c.gameEndFlag {
		return
	}
	c.phase = CucumberPhaseGameEnd
	c.gameEndFlag = true
	// 人間以外で失点のいちばん少ない席を勝ちにする。
	best := -1
	for i := 1; i < len(c.players); i++ {
		if best < 0 || c.players[i].GetPenalty() < c.players[best].GetPenalty() {
			best = i
		}
	}
	if best < 0 {
		best = 0
	}
	c.winnerIdx = best
	c.addLog(0, "giveup", "投了しました", nil)
}

// chooseCpuCard は CPU が出す札を選ぶ。
//
// **終盤の高い札だけが危険。** 更新できるうちは安い更新札で凌ぎ、最終トリックに
// 高い札を持ち越さないようにします。
func (c *Cucumber) chooseCpuCard(playerIdx int) int {
	valid := c.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := c.players[playerIdx]
	// 最終トリックは取ると失点なので、更新できるなら**いちばん低い更新札**にする。
	if pick := pickLowest(p, valid, cucumberRank); pick >= 0 {
		return pick
	}
	return valid[0]
}

// GetHint は人間への助言を返す。
func (c *Cucumber) GetHint() *CucumberHint {
	if c.gameEndFlag || c.phase != CucumberPhasePlay || !c.IsHumanTurn() {
		return nil
	}
	valid := c.GetValidPlayIndices(0)
	if len(valid) == 0 {
		return nil
	}
	idx := c.chooseCpuCard(0)
	reason := "cucumberBeat"
	switch {
	case c.forcedLowest(0, valid):
		// 更新できないので低い札に決まっている。
		reason = "cucumberForced"
	case c.HighestInTrick() == 0:
		reason = "cucumberLead"
	}
	return &CucumberHint{CardIndex: &idx, Reason: reason}
}

// cucumberContains は xs に v が含まれるかを返す。
//
// **他ゲームのヘルパは使えません。** 近いものは別のビルドタグの中にあり、
// ホストビルドだけ通って Worker が落ちます。
func cucumberContains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// addLog は棋譜に 1 行足す。
func (c *Cucumber) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	c.appendLog(playerIdx, actionType, detail, cards)
}

// GetConfig は設定を返す。
func (c *Cucumber) GetConfig() CucumberConfig { return c.config }

// SetConfig は設定を更新する。
func (c *Cucumber) SetConfig(cfg CucumberConfig) {
	if err := cfg.Validate(); err != nil {
		return
	}
	if cfg.PlayerCnt != c.config.PlayerCnt {
		c.players = newCucumberSeats(cfg.PlayerCnt)
	}
	c.config = cfg
}

// GetPhase は現在のフェーズを返す。
func (c *Cucumber) GetPhase() CucumberPhase { return c.phase }

// GetGameEndFlag は終局フラグを返す。
func (c *Cucumber) GetGameEndFlag() bool { return c.gameEndFlag }

// GetCurrentTrick は現在のトリックを返す。
func (c *Cucumber) GetCurrentTrick() []*TrickCard { return c.currentTrick }

// GetCurrentPlayerIdx は現在の手番を返す。
func (c *Cucumber) GetCurrentPlayerIdx() int { return c.currentPlayerIdx }

// GetLeadPlayerIdx はリード席を返す。
func (c *Cucumber) GetLeadPlayerIdx() int { return c.leadPlayerIdx }

// GetTrickNumber は解決済みのトリック数を返す。
func (c *Cucumber) GetTrickNumber() int { return c.trickNumber }

// GetRoundNumber はラウンド数を返す。
func (c *Cucumber) GetRoundNumber() int { return c.roundNumber }

// GetLastTrickWinnerIdx は直前ラウンドで最終トリックを取った席を返す (-1 = 未)。
func (c *Cucumber) GetLastTrickWinnerIdx() int { return c.lastTrickWinnerIdx }

// GetLastPenalty は直前ラウンドで付いた失点を返す。
func (c *Cucumber) GetLastPenalty() int { return c.lastPenalty }

// GetPlayerCnt は人数を返す。
func (c *Cucumber) GetPlayerCnt() int { return len(c.players) }

// GetPlayer は席 i のプレイヤーを返す。
func (c *Cucumber) GetPlayer(i int) *CucumberPlayer {
	if i < 0 || i >= len(c.players) {
		return nil
	}
	return c.players[i]
}

// GetWinnerIdx は勝者の席を返す (-1 = 未確定)。
func (c *Cucumber) GetWinnerIdx() int { return c.winnerIdx }

// cucumberJSON is the JSON wire format for Cucumber.
type cucumberJSON struct {
	Players            []*CucumberPlayer `json:"pl"`
	Config             CucumberConfig    `json:"cf"`
	Phase              CucumberPhase     `json:"ph"`
	CurrentTrick       []*TrickCard      `json:"ct"`
	LeadPlayerIdx      int               `json:"lp"`
	CurrentIdx         int               `json:"ci"`
	TrickNumber        int               `json:"tn"`
	RoundNumber        int               `json:"rn"`
	LastTrickWinnerIdx int               `json:"lw"`
	LastPenalty        int               `json:"lpn"`
	GameEndFlag        bool              `json:"ge"`
	WinnerIdx          int               `json:"wi"`
	ActionLog          []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (c *Cucumber) MarshalJSON() ([]byte, error) {
	return json.Marshal(cucumberJSON{
		Players:            c.players,
		Config:             c.config,
		Phase:              c.phase,
		CurrentTrick:       c.currentTrick,
		LeadPlayerIdx:      c.leadPlayerIdx,
		CurrentIdx:         c.currentPlayerIdx,
		TrickNumber:        c.trickNumber,
		RoundNumber:        c.roundNumber,
		LastTrickWinnerIdx: c.lastTrickWinnerIdx,
		LastPenalty:        c.lastPenalty,
		GameEndFlag:        c.gameEndFlag,
		WinnerIdx:          c.winnerIdx,
		ActionLog:          c.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Cucumber) UnmarshalJSON(data []byte) error {
	var j cucumberJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if len(j.Players) != j.Config.PlayerCnt {
		return fmt.Errorf("seat count %d does not match the configured %d", len(j.Players), j.Config.PlayerCnt)
	}
	if j.Phase < CucumberPhaseMin || j.Phase > CucumberPhaseMax {
		return fmt.Errorf("phase out of range: %d", j.Phase)
	}
	if len(j.ActionLog) > cucumberMaxSliceLen {
		return fmt.Errorf("action log too long: %d", len(j.ActionLog))
	}
	// **トリックに座れるのは 1 人 1 枚まで。**
	if len(j.CurrentTrick) > len(j.Players) {
		return fmt.Errorf("the trick holds %d cards for %d seats", len(j.CurrentTrick), len(j.Players))
	}
	seen := make(map[int]bool, len(j.CurrentTrick))
	for i, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil {
			return fmt.Errorf("trick entry %d is missing its card", i)
		}
		if tc.PlayerIdx < 0 || tc.PlayerIdx >= len(j.Players) {
			return fmt.Errorf("trick entry %d has seat %d out of range", i, tc.PlayerIdx)
		}
		if seen[tc.PlayerIdx] {
			return fmt.Errorf("seat %d played twice in the same trick", tc.PlayerIdx)
		}
		seen[tc.PlayerIdx] = true
	}
	if j.CurrentIdx < 0 || j.CurrentIdx >= len(j.Players) {
		return fmt.Errorf("current player index out of range: %d", j.CurrentIdx)
	}
	if j.LeadPlayerIdx < 0 || j.LeadPlayerIdx >= len(j.Players) {
		return fmt.Errorf("lead player index out of range: %d", j.LeadPlayerIdx)
	}
	if j.LastTrickWinnerIdx < -1 || j.LastTrickWinnerIdx >= len(j.Players) {
		return fmt.Errorf("last trick winner index out of range: %d", j.LastTrickWinnerIdx)
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= len(j.Players) {
		return fmt.Errorf("winner index out of range: %d", j.WinnerIdx)
	}
	if j.GameEndFlag != (j.Phase == CucumberPhaseGameEnd) {
		return fmt.Errorf("the game-end flag and the phase disagree (flag=%v, phase=%d)", j.GameEndFlag, j.Phase)
	}
	if j.GameEndFlag != (j.WinnerIdx >= 0) {
		return fmt.Errorf("a finished game has a winner and an unfinished one does not (flag=%v, winner=%d)",
			j.GameEndFlag, j.WinnerIdx)
	}
	if j.TrickNumber < 0 || j.TrickNumber > CucumberHandSize {
		return fmt.Errorf("trick number out of range: %d", j.TrickNumber)
	}
	if j.RoundNumber < 0 {
		return fmt.Errorf("round number cannot be negative: %d", j.RoundNumber)
	}
	// **失点が付いたなら、それを負った席が居ます。**
	if j.LastPenalty < 0 {
		return fmt.Errorf("the last penalty cannot be negative: %d", j.LastPenalty)
	}
	if (j.LastPenalty > 0) != (j.LastTrickWinnerIdx >= 0) {
		return fmt.Errorf("a penalty was scored exactly when a seat took the last trick (penalty=%d, seat=%d)",
			j.LastPenalty, j.LastTrickWinnerIdx)
	}

	// **手札は全員そろって減ります。** トリックに出た枚数のぶんだけ差が付く形しか
	// ありえないので、7 枚から解決済みトリック数を引いた枚数の前後に収まります。
	for i, p := range j.Players {
		if p == nil {
			return fmt.Errorf("seat %d is missing", i)
		}
		want := CucumberHandSize - j.TrickNumber
		if n := p.GetCardsSize(); n != want && n != want-1 {
			return fmt.Errorf("seat %d holds %d cards after %d tricks, want %d or %d",
				i, n, j.TrickNumber, want-1, want)
		}
	}

	c.players = j.Players
	c.config = j.Config
	c.phase = j.Phase
	c.currentTrick = j.CurrentTrick
	c.leadPlayerIdx = j.LeadPlayerIdx
	c.currentPlayerIdx = j.CurrentIdx
	c.trickNumber = j.TrickNumber
	c.roundNumber = j.RoundNumber
	c.lastTrickWinnerIdx = j.LastTrickWinnerIdx
	c.lastPenalty = j.LastPenalty
	c.gameEndFlag = j.GameEndFlag
	c.winnerIdx = j.WinnerIdx
	c.actionLog = j.ActionLog
	if c.trumpCards == nil {
		c.trumpCards = NewTrumpCards(0)
	}
	return nil
}
