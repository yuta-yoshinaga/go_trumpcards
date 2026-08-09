//go:build !js || !wasm || extra2

// Package domain — 撿紅點 (Chinese Ten) のドメインモデル。
//
// 中国・台湾のフィッシング系ゲーム。52 枚を使い、**赤札だけが得点になる**。
//
// # 捕獲
//
//   - A〜9 は、場札との**合計がちょうど 10** になる札を取れる (A+9, 3+7, 5+5 …)
//   - 10・J・Q・K は**同ランク**の札しか取れない
//   - 出した 1 枚が取れるのは場札 **1 枚だけ**
//
// # 1 手番の構造 — issue #4378 が丸ごと落としている部分
//
// 手番は **2 段**ある:
//
//  1. 手札から 1 枚出す。取れれば両方を取り札へ、取れなければ場に置く
//  2. **続けて山札の一番上をめくり、同じ規則で場札を取る**。取れなければ場に残る
//
// この 2 段目がこのゲームの骨格である。issue の手順にはこれが無く、その結果
// 「手札がなくなれば再配布し、山札が尽きるまで繰り返す」と書かれているが、
// **再配布は無い** —— 手札 1 枚につき山札 1 枚が減るので両者は同時に尽きる。
// めくりを省くと場札が増え続け、取れる札が半分になって別のゲームになる。
//
// # 得点
//
// 赤札 (♥♦) だけが得点する。2〜8 は額面、9〜K は 10 点、A は 20 点。
// 合計は 70 + 100 + 40 = **210 点**で、引き分け点 ChineseTenTieScore = 105 は
// ちょうどその半分。資料の 105 と数え上げが一致することが、この得点表が正しい
// ことの裏付けになっている (TestChineseTen_TieScoreIsHalfTheRedPoints)。
//
// # 人数
//
// 本実装は **2 人戦**のみ。3 人戦では♠A が、4 人戦では黒の A 両方が得点に加わり、
// 上の 210/105 も変わる。まず完全に文書化されている 2 人戦を正確に実装する。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// ChineseTenPlayerCnt はプレイヤー数 (v1は2人固定)。
const ChineseTenPlayerCnt = 2

// ChineseTenHandSize は 2 人戦の配札枚数 (24 ÷ 人数)。
const ChineseTenHandSize = 12

// ChineseTenLayoutSize は開始時に場へ置く枚数。
const ChineseTenLayoutSize = 4

// ChineseTenDeckSize は 52 枚。
const ChineseTenDeckSize = 52

// ChineseTenTotalRedPoints は赤札の合計点。
const ChineseTenTotalRedPoints = 210

// ChineseTenTieScore は引き分けとなる点。赤札総点のちょうど半分。
const ChineseTenTieScore = ChineseTenTotalRedPoints / 2

// ChineseTenCardPoints は 1 枚の得点を返す。
//
// **赤札 (♥♦) だけが点になる。** 黒札は 0 点で、取っても得点には効かない
// (2 人戦の場合。3〜4 人戦では黒の A が加わるが本実装は 2 人戦のみ)。
func ChineseTenCardPoints(c *Card) int {
	if c == nil || !chineseTenIsRed(c) {
		return 0
	}
	switch v := c.GetValue(); {
	case v == 1:
		return 20
	case v >= 2 && v <= 8:
		return v
	default: // 9, 10, J, Q, K
		return 10
	}
}

// chineseTenIsRed はハートまたはダイヤかを返す。
func chineseTenIsRed(c *Card) bool {
	return c != nil && (c.GetDesign() == CardDesignHeart || c.GetDesign() == CardDesignDiamond)
}

// ChineseTenCaptures は played が target を取れるかを返す。
//
// A〜9 は合計 10、10〜K は同ランク。**この 2 つは別の規則**で、たとえば 10 は
// 「合計 10」では取れない (10 + 0 の札は存在しないが、規則としても同ランク限定)。
func ChineseTenCaptures(played, target *Card) bool {
	if played == nil || target == nil {
		return false
	}
	pv, tv := played.GetValue(), target.GetValue()
	if pv <= 9 {
		return tv <= 9 && pv+tv == 10
	}
	return pv == tv
}

// ChineseTenPhase はゲームフェーズ。
type ChineseTenPhase int

// ChineseTenのフェーズ定数
const (
	// ChineseTenPhasePlay 手札を出すフェーズ
	ChineseTenPhasePlay ChineseTenPhase = iota
	// ChineseTenPhaseSelect 出した札が複数の場札を取れるとき、どれを取るか選ぶ
	ChineseTenPhaseSelect
	// ChineseTenPhaseGameEnd 終局
	ChineseTenPhaseGameEnd
)

// newChineseTenDeck は 52 枚を生成する (シャッフル前)。
func newChineseTenDeck() []*Card {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	deck := make([]*Card, 0, ChineseTenDeckSize)
	for _, s := range suits {
		for v := 1; v <= 13; v++ {
			deck = append(deck, NewCard(s, v, true))
		}
	}
	return deck
}

// chineseTenShuffle は Fisher-Yates。domain の shuffleCards は casino タグの
// ファイルにあり extra2 ビルドから見えないため、専用名で持つ。
func chineseTenShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// ChineseTen は撿紅點のゲームクラス。
type ChineseTen struct {
	players     []*ChineseTenPlayer
	config      ChineseTenConfig
	phase       ChineseTenPhase
	stock       []*Card
	layout      []*Card
	captured    [][]*Card
	currentIdx  int
	pending     *Card // 選択待ちの、出した/めくった札
	pendingFlip bool  // pending が山札めくりに由来するか
	scores      []int
	gameEndFlag bool
	winnerIdx   int
	actionLogBase
}

// NewChineseTen はコンストラクタ。
func NewChineseTen(players []*ChineseTenPlayer, config ChineseTenConfig) *ChineseTen {
	return &ChineseTen{
		players:   players,
		config:    config,
		captured:  make([][]*Card, len(players)),
		scores:    make([]int, len(players)),
		winnerIdx: -1,
	}
}

// NewDefaultChineseTen は標準の 2 人セットアップを返す。
func NewDefaultChineseTen() *ChineseTen {
	return NewChineseTen(
		[]*ChineseTenPlayer{NewChineseTenPlayer(true), NewChineseTenPlayer(false)},
		DefaultChineseTenConfig(),
	)
}

// Reset はゲームを初期化する。
func (c *ChineseTen) Reset() {
	c.phase = ChineseTenPhasePlay
	c.pending = nil
	c.pendingFlip = false
	c.gameEndFlag = false
	c.winnerIdx = -1
	c.actionLog = nil
	c.layout = nil
	c.scores = make([]int, len(c.players))
	c.captured = make([][]*Card, len(c.players))
	for i := range c.captured {
		c.captured[i] = make([]*Card, 0, ChineseTenDeckSize)
	}
	for _, p := range c.players {
		p.ResetGame()
	}

	deck := newChineseTenDeck()
	chineseTenShuffle(deck)
	pos := 0
	for range ChineseTenHandSize {
		for _, p := range c.players {
			p.AddCard(deck[pos])
			pos++
		}
	}
	c.layout = append(c.layout, deck[pos:pos+ChineseTenLayoutSize]...)
	pos += ChineseTenLayoutSize
	c.stock = append([]*Card(nil), deck[pos:]...)

	c.currentIdx = 0
	c.addLog(-1, "deal", "cards dealt", nil)
}

// PlayCard は player が手札 handIdx の札を出す。
func (c *ChineseTen) PlayCard(player, handIdx int) error {
	if c.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if c.phase != ChineseTenPhasePlay {
		return fmt.Errorf("a selection is pending")
	}
	if player != c.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	p := c.GetPlayer(player)
	if p == nil {
		return fmt.Errorf("no such player: %d", player)
	}
	if handIdx < 0 || handIdx >= p.GetCardsSize() {
		return fmt.Errorf("card index %d out of range", handIdx)
	}
	card := p.RemoveCard(handIdx)
	if card == nil {
		return fmt.Errorf("card index %d is empty", handIdx)
	}
	c.resolve(player, card, false)
	return nil
}

// resolve は 1 枚出した (またはめくった) 結果を場へ反映する。
func (c *ChineseTen) resolve(player int, card *Card, fromFlip bool) {
	matches := c.matchIndices(card)
	switch len(matches) {
	case 0:
		c.layout = append(c.layout, card)
		c.addLog(player, "place", "no capture; the card joins the layout", []*Card{card})
	case 1:
		c.capture(player, card, matches[0])
	default:
		// **めくった札にも選択させる。** めくりは手番の後半であって別の誰かの
		// 手番ではないので、選ぶのは同じプレイヤーである。CPU の番なら CPU が
		// 解決するので入力待ちにはならない。
		//
		// 自動で最高点を取る実装にしていたが、それだと pendingFlip が常に false
		// になり、姉妹ゲームの Mushi (同じフィッシング系) とも挙動が食い違って
		// いた。出典はめくりの選択者を明示していないが、卓上では出した本人が
		// 選ぶし、選ばせない理由も無い。
		c.pending = card
		c.pendingFlip = fromFlip
		c.phase = ChineseTenPhaseSelect
		return
	}
	c.afterResolve(player, fromFlip)
}

// SelectCapture は選択フェーズで場札 layoutIdx を取る。
func (c *ChineseTen) SelectCapture(player, layoutIdx int) error {
	if c.phase != ChineseTenPhaseSelect {
		return fmt.Errorf("no selection is pending")
	}
	if player != c.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	if layoutIdx < 0 || layoutIdx >= len(c.layout) {
		return fmt.Errorf("layout index %d out of range", layoutIdx)
	}
	if !ChineseTenCaptures(c.pending, c.layout[layoutIdx]) {
		return fmt.Errorf("layout card %d cannot be captured by that card", layoutIdx)
	}

	card, fromFlip := c.pending, c.pendingFlip
	c.pending, c.pendingFlip = nil, false
	c.phase = ChineseTenPhasePlay
	c.capture(player, card, layoutIdx)
	c.afterResolve(player, fromFlip)
	return nil
}

// matchIndices は card が取れる場札の添字を返す。
func (c *ChineseTen) matchIndices(card *Card) []int {
	var out []int
	for i, l := range c.layout {
		if ChineseTenCaptures(card, l) {
			out = append(out, i)
		}
	}
	return out
}

// bestMatch は候補のうち最も得点の高い場札の添字を返す。
func (c *ChineseTen) bestMatch(candidates []int) int {
	best, bestPts := candidates[0], -1
	for _, i := range candidates {
		if pts := ChineseTenCardPoints(c.layout[i]); pts > bestPts {
			best, bestPts = i, pts
		}
	}
	return best
}

// capture は出した札と場札 1 枚を取り札へ移す。
func (c *ChineseTen) capture(player int, card *Card, layoutIdx int) {
	taken := []*Card{card, c.layout[layoutIdx]}
	c.layout = append(c.layout[:layoutIdx], c.layout[layoutIdx+1:]...)
	c.captured[player] = append(c.captured[player], taken...)
	c.scores[player] += ChineseTenCardPoints(taken[0]) + ChineseTenCardPoints(taken[1])
	c.addLog(player, "capture", "captures a pair", taken)
}

// afterResolve は 1 枚ぶんの処理後の進行 (山札めくり → 手番交代)。
func (c *ChineseTen) afterResolve(player int, wasFlip bool) {
	if !wasFlip {
		if len(c.stock) > 0 {
			flip := c.stock[0]
			c.stock = c.stock[1:]
			c.resolve(player, flip, true)
			return
		}
	}
	if c.phase == ChineseTenPhaseSelect {
		return
	}
	c.currentIdx = (player + 1) % len(c.players)
	if c.handsEmpty() {
		c.finishGame()
	}
}

// handsEmpty は全員の手札が尽きたかを返す。
func (c *ChineseTen) handsEmpty() bool {
	for _, p := range c.players {
		if p.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// finishGame は終局処理。得点は赤札の合計で、105 (総点の半分) が引き分け。
func (c *ChineseTen) finishGame() {
	c.gameEndFlag = true
	c.phase = ChineseTenPhaseGameEnd
	switch {
	case c.scores[0] > c.scores[1]:
		c.winnerIdx = 0
	case c.scores[1] > c.scores[0]:
		c.winnerIdx = 1
	default:
		c.winnerIdx = -1
	}
	c.addLog(-1, "game", "game over", nil)
}

// addLog は棋譜へ 1 行追加する。
func (c *ChineseTen) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	c.appendLog(playerIdx, actionType, detail, cards)
}

// ---- CPU ----

// ChineseTenCpuAction は CPU が選んだ手。
type ChineseTenCpuAction struct {
	// HandIdx は出す手札の添字 (選択フェーズでは -1)。
	HandIdx int
	// LayoutIdx は選択フェーズで取る場札の添字 (それ以外は -1)。
	LayoutIdx int
}

// ChineseTenCpuDecide は idx の CPU が取る手を決める。
//
// 選択フェーズなら最も点の高い場札を取る。それ以外は、最も点の高い捕獲になる
// 手札を出し、どれも取れなければ最も点の低い札 (＝相手に渡しても損の小さい札) を
// 場に置く。
func (c *ChineseTen) ChineseTenCpuDecide(idx int) ChineseTenCpuAction {
	if c.phase == ChineseTenPhaseSelect {
		best, bestPts := -1, -1
		for _, i := range c.matchIndices(c.pending) {
			if pts := ChineseTenCardPoints(c.layout[i]); pts > bestPts {
				best, bestPts = i, pts
			}
		}
		return ChineseTenCpuAction{HandIdx: -1, LayoutIdx: best}
	}

	p := c.GetPlayer(idx)
	if p == nil || p.GetCardsSize() == 0 {
		return ChineseTenCpuAction{HandIdx: -1, LayoutIdx: -1}
	}

	bestIdx, bestGain := -1, -1
	discardIdx, discardPts := 0, -1
	for i := range p.GetCardsSize() {
		card := p.GetCard(i)
		if card == nil {
			continue
		}
		gain := 0
		if matches := c.matchIndices(card); len(matches) > 0 {
			gain = ChineseTenCardPoints(card) + ChineseTenCardPoints(c.layout[c.bestMatch(matches)])
		}
		if gain > bestGain {
			bestIdx, bestGain = i, gain
		}
		if pts := ChineseTenCardPoints(card); discardPts == -1 || pts < discardPts {
			discardIdx, discardPts = i, pts
		}
	}
	if bestGain > 0 {
		return ChineseTenCpuAction{HandIdx: bestIdx, LayoutIdx: -1}
	}
	return ChineseTenCpuAction{HandIdx: discardIdx, LayoutIdx: -1}
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (c *ChineseTen) GetPlayers() []*ChineseTenPlayer { return c.players }

// GetPlayer は idx のプレイヤーを返す。範囲外は nil。
func (c *ChineseTen) GetPlayer(idx int) *ChineseTenPlayer {
	return getPlayer(c.players, idx)
}

// GetLayout は場札を返す。
func (c *ChineseTen) GetLayout() []*Card { return c.layout }

// GetStockCount は山札の残り枚数を返す。
func (c *ChineseTen) GetStockCount() int { return len(c.stock) }

// GetCaptured は idx の取り札を返す。
func (c *ChineseTen) GetCaptured(idx int) []*Card {
	if idx < 0 || idx >= len(c.captured) {
		return nil
	}
	return c.captured[idx]
}

// GetScore は idx の得点を返す。
func (c *ChineseTen) GetScore(idx int) int {
	return elemAt(c.scores, idx)
}

// SetScore は idx の得点を設定する (テスト用)。
func (c *ChineseTen) SetScore(idx, score int) {
	if idx < 0 || idx >= len(c.scores) {
		return
	}
	c.scores[idx] = score
}

// GetPhase は現在のフェーズを返す。
func (c *ChineseTen) GetPhase() ChineseTenPhase { return c.phase }

// GetCurrentPlayerIdx は手番のプレイヤー添字を返す。
func (c *ChineseTen) GetCurrentPlayerIdx() int { return c.currentIdx }

// GetPendingCard は選択待ちの札を返す (無ければ nil)。
func (c *ChineseTen) GetPendingCard() *Card { return c.pending }

// GetSelectableIndices は選択フェーズで取れる場札の添字を返す。
func (c *ChineseTen) GetSelectableIndices() []int {
	if c.phase != ChineseTenPhaseSelect {
		return nil
	}
	return c.matchIndices(c.pending)
}

// GetGameEndFlag は終局しているかを返す。
func (c *ChineseTen) GetGameEndFlag() bool { return c.gameEndFlag }

// GetWinnerIdx は勝者の添字を返す。引き分けは -1。
func (c *ChineseTen) GetWinnerIdx() int { return c.winnerIdx }

// GetConfig はゲーム設定を返す。
func (c *ChineseTen) GetConfig() ChineseTenConfig { return c.config }

// SetConfig はゲーム設定を差し替える。
func (c *ChineseTen) SetConfig(cfg ChineseTenConfig) { c.config = cfg }

// ---- JSON ----

// chineseTenJSON は KV のワイヤ形式。Worker は毎リクエストここから組み直すので、
// ここに無いものは次のリクエストでは存在しない。
type chineseTenJSON struct {
	Players     []*ChineseTenPlayer `json:"pl"`
	Config      ChineseTenConfig    `json:"cf"`
	Phase       ChineseTenPhase     `json:"ph"`
	Stock       []*Card             `json:"st"`
	Layout      []*Card             `json:"ly"`
	Captured    [][]*Card           `json:"cp"`
	CurrentIdx  int                 `json:"ci"`
	Pending     *Card               `json:"pd"`
	PendingFlip bool                `json:"pf"`
	Scores      []int               `json:"sc"`
	GameEndFlag bool                `json:"ge"`
	WinnerIdx   int                 `json:"wi"`
	ActionLog   []*ActionLogEntry   `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (c *ChineseTen) MarshalJSON() ([]byte, error) {
	return json.Marshal(chineseTenJSON{
		Players:     c.players,
		Config:      c.config,
		Phase:       c.phase,
		Stock:       c.stock,
		Layout:      c.layout,
		Captured:    c.captured,
		CurrentIdx:  c.currentIdx,
		Pending:     c.pending,
		PendingFlip: c.pendingFlip,
		Scores:      c.scores,
		GameEndFlag: c.gameEndFlag,
		WinnerIdx:   c.winnerIdx,
		ActionLog:   c.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// Worker はこれを KV の未検証バイト列に対して毎リクエスト実行する。添字は信用せず
// 丸め、取り札スライスは人数ぶんに揃える。
func (c *ChineseTen) UnmarshalJSON(data []byte) error {
	var j chineseTenJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) == 0 {
		return fmt.Errorf("chineseten: no players in snapshot")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("chineseten: %w", err)
	}
	c.players = j.Players
	c.config = j.Config
	c.phase = j.Phase
	c.stock = j.Stock
	c.layout = j.Layout
	c.pending = j.Pending
	c.pendingFlip = j.PendingFlip
	c.gameEndFlag = j.GameEndFlag
	c.actionLog = j.ActionLog

	n := len(c.players)
	c.currentIdx = chineseTenClampSeat(j.CurrentIdx, n)
	c.winnerIdx = chineseTenClampSeat(j.WinnerIdx, n)

	c.scores = make([]int, n)
	copy(c.scores, j.Scores)
	c.captured = make([][]*Card, n)
	for i := range c.captured {
		if i < len(j.Captured) && j.Captured[i] != nil {
			c.captured[i] = j.Captured[i]
		} else {
			c.captured[i] = make([]*Card, 0, ChineseTenDeckSize)
		}
	}
	return nil
}

// chineseTenClampSeat は範囲外のプレイヤー添字を -1 に丸める。
func chineseTenClampSeat(idx, n int) int {
	if idx < 0 || idx >= n {
		return -1
	}
	return idx
}

// SetLayoutForTest は場札を差し替える (テスト用)。
func (c *ChineseTen) SetLayoutForTest(cards []*Card) { c.layout = cards }

// SetCurrentPlayerForTest は手番を設定する (テスト用)。
func (c *ChineseTen) SetCurrentPlayerForTest(idx int) { c.currentIdx = idx }

// SetStockForTest は山札を差し替える (テスト用)。
func (c *ChineseTen) SetStockForTest(cards []*Card) { c.stock = cards }
