//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// BhabhiPhase はバービーのゲームフェーズ。
type BhabhiPhase int

// Bhabhi のフェーズ定数
const (
	// BhabhiPhasePlay プレイ中
	BhabhiPhasePlay BhabhiPhase = iota
	// BhabhiPhaseGameEnd ゲーム終了（Bhabhi が確定した）
	BhabhiPhaseGameEnd
)

// BhabhiDefaultPlayers は既定の参加人数。
const BhabhiDefaultPlayers = 4

// BhabhiDeckSize は使用するデッキの枚数。
const BhabhiDeckSize = 52

// bhabhiMaxSliceLen caps slice sizes during deserialisation.
const bhabhiMaxSliceLen = 1000

// BhabhiStalemateTricks は引き分け（膠着）と判定するまでのトリック数。
//
// **バービーは必ず終わるゲームではない。** 場から札が落ちるのは全員が
// フォローできたトリックだけなので、4 つのスートに札が散って常に誰かが
// フォローできない配置になると、引き取りだけが延々と続いて手札の総数が
// 減らなくなる。実測（各人数 400 局、CPU 同士）:
//
//	n=3 膠着 0/400   決着トリック数 中央値 22 / 最大 30
//	n=4 膠着 18/400  中央値 23 / 最大 43
//	n=5 膠着 0/400   中央値 27 / 最大 54
//	n=6 膠着 3/400   中央値 32 / 最大 73
//	n=7 膠着 3/400   中央値 41 / 最大 86
//
// そこで上限を**実測最大の 3 倍以上**に置き、到達したら
// **いちばん手札の多い人を Bhabhi** として決着させる。普通の局では絶対に
// 発火しないが、発火したときに「誰も負けないまま止まる」ことは無くなる。
const BhabhiStalemateTricks = 300

// Bhabhi はバービー（Bhabhi / Get Away）のゲームクラス。
//
// インド・パキスタンの家庭で遊ばれる**回避型**。52 枚を参加人数で配り切り、
// リードのスートにフォローできなかった人が**場に出ている札を全部手札に
// 引き取る**。手札を出し切った人から抜けていき、**最後に 1 人だけ残った人が
// Bhabhi（敗者）**。勝者を決めるのではなく敗者を決めるゲームです。
//
// **フォローが揃ったトリックは捨てられる（引き取りではない）。** ここを
// 引き取りにすると場に出た札が必ず誰かの手札へ戻るので、手札の総数が 52 枚の
// まま減らず、**誰も上がれずゲームが終わりません**。issue #5244 の規則 4 は
// そう書かれていますが、そのままでは成立しないので採っていません。
type Bhabhi struct {
	players     []*BhabhiPlayer
	config      BhabhiConfig
	phase       BhabhiPhase
	pile        []*TrickCard // 場に出ている札（引き取り対象）
	leadSuit    int          // 現在のリードスート（0 = トリック未開始）
	currentIdx  int
	leadIdx     int
	trickNumber int
	// lastPickupIdx は直前に場札を引き取った人（-1 = まだ無い）。
	lastPickupIdx int
	// lastPickupSize は直前に引き取った枚数。
	lastPickupSize int
	finishedCnt    int
	gameEndFlag    bool
	// bhabhiIdx は敗者（-1 = 未確定）。
	bhabhiIdx int
	// stalemate は膠着で打ち切ったかどうか。
	stalemate bool
	actionLogBase
}

// NewBhabhi はコンストラクタ。
func NewBhabhi(players []*BhabhiPlayer, config BhabhiConfig) *Bhabhi {
	return &Bhabhi{
		players:       players,
		config:        config,
		leadSuit:      0,
		lastPickupIdx: -1,
		bhabhiIdx:     -1,
	}
}

// NewDefaultBhabhi は既定人数のセットアップを返す。
func NewDefaultBhabhi() *Bhabhi {
	return NewBhabhi(newBhabhiPlayers(BhabhiDefaultPlayers), DefaultBhabhiConfig())
}

// newBhabhiPlayers は先頭を人間、残りを CPU として n 人ぶん作る。
func newBhabhiPlayers(n int) []*BhabhiPlayer {
	players := make([]*BhabhiPlayer, 0, n)
	players = append(players, NewBhabhiPlayer(true))
	for range n - 1 {
		players = append(players, NewBhabhiPlayer(false))
	}
	return players
}

// Reset はゲームを初期化する。
//
// **人数は設定から作り直す。** 設定だけ変えて席を作り直さないと、変更が
// 次のゲームに反映されません。
func (b *Bhabhi) Reset() {
	if len(b.players) != b.config.PlayerCnt {
		b.players = newBhabhiPlayers(b.config.PlayerCnt)
	}
	b.phase = BhabhiPhasePlay
	b.pile = nil
	b.leadSuit = 0
	b.trickNumber = 0
	b.lastPickupIdx = -1
	b.lastPickupSize = 0
	b.finishedCnt = 0
	b.gameEndFlag = false
	b.bhabhiIdx = -1
	b.stalemate = false
	b.actionLog = nil
	for _, p := range b.players {
		p.ResetGame()
	}
	b.deal()
	b.leadIdx = 0
	b.currentIdx = 0
	b.addLog(-1, "deal", fmt.Sprintf("%d 人に配り切りました", len(b.players)), nil)
}

// deal は 52 枚を参加人数で配り切る。
//
// **52 は 3/5/6/7 人で割り切れない。** 余りは先頭の席から 1 枚ずつ多く配る
// ので、手札枚数は最大 1 枚差になります。山札は残しません。
func (b *Bhabhi) deal() {
	deck := NewTrumpCards(0)
	deck.Shuffle()
	for i := 0; i < BhabhiDeckSize; i++ {
		c := deck.DrawCard()
		if c == nil {
			break
		}
		b.players[i%len(b.players)].AddCard(c)
	}
	b.sortAllHands()
}

// sortAllHands は手札をスート・ランク順に整える。
func (b *Bhabhi) sortAllHands() {
	for _, p := range b.players {
		sortPlayerHand(p, func(ci, cj *Card) bool {
			if ci.GetDesign() != cj.GetDesign() {
				return ci.GetDesign() < cj.GetDesign()
			}
			return bhabhiRank(ci) < bhabhiRank(cj)
		})
	}
}

// bhabhiRank は札の強さ。**A が最強**（A > K > Q > J > 10 > … > 2）。
func bhabhiRank(c *Card) int {
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// GetValidPlayIndices は playerIdx が出せる手札の添字を返す。
//
// **フォローできるならそのスートしか出せない。** フォローできないときだけ
// 何を出してもよく、そのときに場札を引き取ります。
func (b *Bhabhi) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(b.players) || b.gameEndFlag {
		return []int{}
	}
	p := b.players[playerIdx]
	if p.IsOut() {
		return []int{}
	}
	all := make([]int, 0, p.GetCardsSize())
	follow := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		all = append(all, i)
		if b.leadSuit != 0 && p.GetCard(i).GetDesign() == b.leadSuit {
			follow = append(follow, i)
		}
	}
	if len(follow) > 0 {
		return follow
	}
	return all
}

// IsHumanTurn は現在の手番が人間かを返す。
func (b *Bhabhi) IsHumanTurn() bool {
	if b.gameEndFlag || b.currentIdx < 0 || b.currentIdx >= len(b.players) {
		return false
	}
	return b.players[b.currentIdx].GetIsHuman()
}

// PlayerPlay は人間が 1 枚出す。
func (b *Bhabhi) PlayerPlay(cardIndex int) error {
	if !b.IsHumanTurn() {
		return errors.New("not your turn")
	}
	return b.play(b.currentIdx, cardIndex)
}

// CpuPlay は CPU が 1 枚出す。
func (b *Bhabhi) CpuPlay() {
	if b.gameEndFlag || b.IsHumanTurn() {
		return
	}
	_ = b.play(b.currentIdx, b.chooseCpuCard(b.currentIdx))
}

// play は playerIdx に手札の cardIndex を出させる。
func (b *Bhabhi) play(playerIdx, cardIndex int) error {
	if b.gameEndFlag {
		return errors.New("game is over")
	}
	if playerIdx != b.currentIdx {
		return fmt.Errorf("not player %d's turn", playerIdx)
	}
	p := b.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return fmt.Errorf("invalid card index: %d", cardIndex)
	}
	if !bhabhiContains(b.GetValidPlayIndices(playerIdx), cardIndex) {
		return errors.New("must follow the led suit")
	}

	card := p.GetCard(cardIndex)
	// **フォローできないなら、この 1 枚を出したうえで場札を全部引き取る。**
	cannotFollow := b.leadSuit != 0 && card.GetDesign() != b.leadSuit

	p.RemoveCard(cardIndex)
	b.pile = append(b.pile, &TrickCard{PlayerIdx: playerIdx, Card: card})
	if b.leadSuit == 0 {
		b.leadSuit = card.GetDesign()
	}
	b.addLog(playerIdx, "play", cardStr(card), []*Card{card})

	if cannotFollow {
		b.pickUpPile(playerIdx)
		return nil
	}
	b.advance()
	return nil
}

// advance は次の手番へ進め、一周したらトリックを解決する。
func (b *Bhabhi) advance() {
	// **手札が尽きた人はその場で抜ける。** 引き取りが起きないかぎり戻れません。
	b.markFinished()
	if b.gameEndFlag {
		return
	}
	if b.trickComplete() {
		b.resolveTrick()
		return
	}
	next := b.nextInTrick(b.currentIdx)
	if next < 0 {
		b.resolveTrick()
		return
	}
	b.currentIdx = next
}

// trickComplete はこのトリックに参加できる全員が出し終えたかを返す。
func (b *Bhabhi) trickComplete() bool {
	return len(b.pile) >= b.trickParticipants()
}

// trickParticipants はこのトリックに出す人数。
//
// **リード時点で手札を持っていた人数**で数える。途中で上がった人を除いて
// しまうと、最後の 1 枚を出した人のぶんが数えられずトリックが閉じません。
func (b *Bhabhi) trickParticipants() int {
	seen := map[int]bool{}
	for _, tc := range b.pile {
		seen[tc.PlayerIdx] = true
	}
	n := len(seen)
	for i, p := range b.players {
		if !seen[i] && !p.IsOut() {
			n++
		}
	}
	return n
}

// nextInTrick はまだこのトリックに出していない次の人を返す（-1 = 全員出した）。
func (b *Bhabhi) nextInTrick(from int) int {
	played := map[int]bool{}
	for _, tc := range b.pile {
		played[tc.PlayerIdx] = true
	}
	n := len(b.players)
	for i := 1; i <= n; i++ {
		j := (from + i) % n
		if !b.players[j].IsOut() && !played[j] {
			return j
		}
	}
	return -1
}

// resolveTrick は全員フォローできたトリックを解決する。
//
// **場札は捨てる。** 引き取らせるとカードが場から消えず、誰も上がれません。
func (b *Bhabhi) resolveTrick() {
	winner := b.trickWinner()
	discarded := len(b.pile)
	b.pile = nil
	b.leadSuit = 0
	b.trickNumber++
	b.addLog(winner, "trick", fmt.Sprintf("%d 枚を場から流しました", discarded), nil)

	b.markFinished()
	if b.gameEndFlag {
		return
	}
	if b.trickNumber >= BhabhiStalemateTricks {
		b.finishStalemate()
		return
	}
	// **最強を出した人が次のリード。** 上がっていたら次の生存者へ回す。
	if winner >= 0 && !b.players[winner].IsOut() {
		b.leadIdx = winner
	} else {
		b.leadIdx = b.nextAlive(winner)
	}
	b.currentIdx = b.leadIdx
}

// trickWinner はリードスートの最強札を出した人を返す。
func (b *Bhabhi) trickWinner() int {
	best, bestRank := -1, -1
	for _, tc := range b.pile {
		if tc.Card.GetDesign() != b.leadSuit {
			continue
		}
		if r := bhabhiRank(tc.Card); r > bestRank {
			best, bestRank = tc.PlayerIdx, r
		}
	}
	return best
}

// pickUpPile は playerIdx に場札を全部引き取らせ、そのままリードさせる。
func (b *Bhabhi) pickUpPile(playerIdx int) {
	p := b.players[playerIdx]
	taken := len(b.pile)
	for _, tc := range b.pile {
		p.AddCard(tc.Card)
	}
	p.AddPickup()
	b.lastPickupIdx = playerIdx
	b.lastPickupSize = taken
	b.pile = nil
	b.leadSuit = 0
	b.trickNumber++
	b.sortAllHands()
	b.addLog(playerIdx, "pickup", fmt.Sprintf("%d 枚を引き取りました", taken), nil)

	// **引き取った人は必ず手札を持っている**ので、上がりの判定は他の席だけ。
	b.markFinished()
	if b.gameEndFlag {
		return
	}
	if b.trickNumber >= BhabhiStalemateTricks {
		b.finishStalemate()
		return
	}
	// **引き取った人は次のリードを取らない。** 取らせると終わらなくなる:
	// あるスートが場に 1 枚しか残っていないとき、それを持っている人がリード
	// → 他方はフォローできず引き取ってその 1 枚を得る → 今度はそちらがリード
	// …… と札が 2 人のあいだを往復し続ける。実測で 200 局中 146 局が
	// この形で止まらなかった。リードを次の席へ渡すと、引き取った直後の人が
	// 同じ札を出し直せないので循環が切れる。
	b.leadIdx = b.nextAlive(playerIdx)
	if b.leadIdx < 0 {
		b.leadIdx = playerIdx
	}
	b.currentIdx = b.leadIdx
}

// markFinished は手札が尽きた席を上がりにし、残り 1 人になったら終局する。
func (b *Bhabhi) markFinished() {
	for i, p := range b.players {
		if p.IsOut() || p.GetCardsSize() > 0 {
			continue
		}
		b.finishedCnt++
		p.SetRank(b.finishedCnt)
		p.SetIsFinished(true)
		b.addLog(i, "finish", fmt.Sprintf("%d 位で上がりました", b.finishedCnt), nil)
	}
	if b.aliveCount() <= 1 {
		b.finishGame()
	}
}

// aliveCount はまだ手札が残っている人数を返す。
func (b *Bhabhi) aliveCount() int {
	n := 0
	for _, p := range b.players {
		if !p.IsOut() {
			n++
		}
	}
	return n
}

// nextAlive は from の次の、まだ手札が残っている人を返す（-1 = 誰もいない）。
func (b *Bhabhi) nextAlive(from int) int {
	n := len(b.players)
	if from < 0 {
		from = 0
	}
	for i := 1; i <= n; i++ {
		j := (from + i) % n
		if !b.players[j].IsOut() {
			return j
		}
	}
	return -1
}

// finishGame は最後に残った 1 人を Bhabhi にして終局する。
func (b *Bhabhi) finishGame() {
	b.phase = BhabhiPhaseGameEnd
	b.gameEndFlag = true
	b.bhabhiIdx = -1
	for i, p := range b.players {
		if !p.IsOut() {
			b.bhabhiIdx = i
			break
		}
	}
	b.stalemate = false
	b.addLog(b.bhabhiIdx, "result", "Bhabhi が確定しました", nil)
}

// finishStalemate は膠着で打ち切り、**いちばん手札の多い人**を Bhabhi にする。
//
// 同数なら席順の早いほうを負けにする（決定的にするため）。
// 詳細は BhabhiStalemateTricks を参照。
func (b *Bhabhi) finishStalemate() {
	b.phase = BhabhiPhaseGameEnd
	b.gameEndFlag = true
	b.stalemate = true
	b.bhabhiIdx = -1
	most := -1
	for i, p := range b.players {
		if p.IsOut() {
			continue
		}
		if n := p.GetCardsSize(); n > most {
			b.bhabhiIdx, most = i, n
		}
	}
	b.addLog(b.bhabhiIdx, "result",
		fmt.Sprintf("%d トリックで膠着。手札が最も多い席を Bhabhi とします", b.trickNumber), nil)
}

// GiveUp は投了する。**投了した人が Bhabhi。**
func (b *Bhabhi) GiveUp() {
	if b.gameEndFlag {
		return
	}
	b.phase = BhabhiPhaseGameEnd
	b.gameEndFlag = true
	b.bhabhiIdx = 0
	b.stalemate = false
	b.addLog(0, "giveup", "投了しました", nil)
}

// chooseCpuCard は CPU の手。
//
//   - フォローできないなら**いちばん高い札**を捨てる。どのみち場札を引き取る
//     ので、手札に残しておきたくない札を落とす場面。
//   - フォローできるなら、**まだ誰かが出す予定なら安く、自分が最後なら**
//     引き取りを避けたい相手に押し付ける意味が無いので素直に安く出す。
//     ただし場札が大きいときは、取っても捨てられるだけなので勝ちにいく。
func (b *Bhabhi) chooseCpuCard(playerIdx int) int {
	valid := b.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := b.players[playerIdx]
	canFollow := b.leadSuit != 0 && p.GetCard(valid[0]).GetDesign() == b.leadSuit

	if b.leadSuit == 0 {
		return b.chooseCpuLead(playerIdx, valid)
	}

	pick, pickRank := valid[0], bhabhiRank(p.GetCard(valid[0]))
	for _, i := range valid[1:] {
		r := bhabhiRank(p.GetCard(i))
		if !canFollow {
			if r > pickRank { // 引き取り確定 → 高い札を落とす
				pick, pickRank = i, r
			}
			continue
		}
		if r < pickRank { // フォローできる → 安く出す
			pick, pickRank = i, r
		}
	}
	return pick
}

// chooseCpuLead はリードの手。
//
// **いちばん枚数の多いスートから安く出す。** 手札に 1 枚しかないスートを
// リードすると誰かがフォローできず引き取って終わり、場札が減らない。出回って
// いるスートほど全員がフォローでき、トリックが閉じて札が場から落ちる。
// 膠着そのものを消せるわけではない（BhabhiStalemateTricks を参照）。
func (b *Bhabhi) chooseCpuLead(playerIdx int, valid []int) int {
	p := b.players[playerIdx]
	counts := map[int]int{}
	for _, i := range valid {
		counts[p.GetCard(i).GetDesign()]++
	}
	pick, pickCount, pickRank := valid[0], counts[p.GetCard(valid[0]).GetDesign()], bhabhiRank(p.GetCard(valid[0]))
	for _, i := range valid[1:] {
		c, r := counts[p.GetCard(i).GetDesign()], bhabhiRank(p.GetCard(i))
		if c > pickCount || (c == pickCount && r < pickRank) {
			pick, pickCount, pickRank = i, c, r
		}
	}
	return pick
}

// GetHint は人間への助言を返す。
func (b *Bhabhi) GetHint() *BhabhiHint {
	if b.gameEndFlag || !b.IsHumanTurn() {
		return nil
	}
	idx := b.chooseCpuCard(0)
	valid := b.GetValidPlayIndices(0)
	if len(valid) == 0 {
		return nil
	}
	var reason string
	switch {
	case b.leadSuit == 0:
		reason = "bhabhiLead"
	case b.players[0].GetCard(valid[0]).GetDesign() != b.leadSuit:
		// フォローできる札があれば GetValidPlayIndices はそれだけを返すので、
		// 先頭が別スートなら**フォローできない**ことが確定している。
		reason = "bhabhiDumpHigh"
	default:
		reason = "bhabhiDuck"
	}
	return &BhabhiHint{CardIndex: &idx, Reason: reason}
}

// BhabhiHint はバービーの助言。
type BhabhiHint struct {
	CardIndex *int
	Reason    string
}

// bhabhiContains は xs が v を含むかを返す。
func bhabhiContains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// --- アクセサ ---------------------------------------------------------------

// GetConfig はゲーム設定を返す。
func (b *Bhabhi) GetConfig() BhabhiConfig { return b.config }

// SetConfig はゲーム設定を設定する。
func (b *Bhabhi) SetConfig(cfg BhabhiConfig) { b.config = cfg }

// GetPhase は現在のフェーズを返す。
func (b *Bhabhi) GetPhase() BhabhiPhase { return b.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (b *Bhabhi) GetGameEndFlag() bool { return b.gameEndFlag }

// GetPlayerCnt はプレイヤー数を返す。
func (b *Bhabhi) GetPlayerCnt() int { return len(b.players) }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (b *Bhabhi) GetPlayer(i int) *BhabhiPlayer {
	if i < 0 || i >= len(b.players) {
		return nil
	}
	return b.players[i]
}

// GetCurrentPlayerIdx は現在の手番を返す。
func (b *Bhabhi) GetCurrentPlayerIdx() int { return b.currentIdx }

// GetLeadPlayerIdx はリードプレイヤーを返す。
func (b *Bhabhi) GetLeadPlayerIdx() int { return b.leadIdx }

// GetLeadSuit はリードスートを返す（0 = トリック未開始）。
func (b *Bhabhi) GetLeadSuit() int { return b.leadSuit }

// GetPile は場に出ている札を返す。
func (b *Bhabhi) GetPile() []*TrickCard { return b.pile }

// GetTrickNumber はこれまでに閉じたトリック数を返す。
func (b *Bhabhi) GetTrickNumber() int { return b.trickNumber }

// GetLastPickupIdx は直前に場札を引き取った人を返す（-1 = まだ無い）。
func (b *Bhabhi) GetLastPickupIdx() int { return b.lastPickupIdx }

// GetLastPickupSize は直前に引き取った枚数を返す。
func (b *Bhabhi) GetLastPickupSize() int { return b.lastPickupSize }

// GetAliveCount はまだ手札が残っている人数を返す。
func (b *Bhabhi) GetAliveCount() int { return b.aliveCount() }

// IsStalemate は膠着で打ち切られたかを返す。
func (b *Bhabhi) IsStalemate() bool { return b.stalemate }

// GetBhabhiIdx は敗者を返す（-1 = 未確定）。
func (b *Bhabhi) GetBhabhiIdx() int { return b.bhabhiIdx }

// GetActionLog は棋譜を返す。
func (b *Bhabhi) GetActionLog() []*ActionLogEntry { return b.actionLog }

// addLog は棋譜に 1 行足す。
func (b *Bhabhi) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	b.appendLog(playerIdx, actionType, detail, cards)
}

// bhabhiJSON は KV スナップショットの表現。
//
// **非公開フィールドは全部ここに載せる。** Worker はリクエストごとに JSON
// からゲームを作り直すので、載せ忘れたものは毎回消えます (#4478)。
type bhabhiJSON struct {
	Players        []*BhabhiPlayer   `json:"pl"`
	Config         BhabhiConfig      `json:"cf"`
	Phase          BhabhiPhase       `json:"ph"`
	Pile           []*TrickCard      `json:"pi"`
	LeadSuit       int               `json:"ls"`
	CurrentIdx     int               `json:"ci"`
	LeadIdx        int               `json:"li"`
	TrickNumber    int               `json:"tn"`
	LastPickupIdx  int               `json:"lpi"`
	LastPickupSize int               `json:"lps"`
	FinishedCnt    int               `json:"fc"`
	GameEndFlag    bool              `json:"ge"`
	BhabhiIdx      int               `json:"bi"`
	Stalemate      bool              `json:"sm"`
	ActionLog      []*ActionLogEntry `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (b *Bhabhi) MarshalJSON() ([]byte, error) {
	return json.Marshal(&bhabhiJSON{
		Players:        b.players,
		Config:         b.config,
		Phase:          b.phase,
		Pile:           b.pile,
		LeadSuit:       b.leadSuit,
		CurrentIdx:     b.currentIdx,
		LeadIdx:        b.leadIdx,
		TrickNumber:    b.trickNumber,
		LastPickupIdx:  b.lastPickupIdx,
		LastPickupSize: b.lastPickupSize,
		FinishedCnt:    b.finishedCnt,
		GameEndFlag:    b.gameEndFlag,
		BhabhiIdx:      b.bhabhiIdx,
		Stalemate:      b.stalemate,
		ActionLog:      b.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元
func (b *Bhabhi) UnmarshalJSON(data []byte) error {
	var j bhabhiJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < BhabhiPhasePlay || j.Phase > BhabhiPhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	n := len(j.Players)
	if n < BhabhiMinPlayers || n > BhabhiMaxPlayers {
		return fmt.Errorf("invalid player count: %d", n)
	}
	if n != j.Config.PlayerCnt {
		return fmt.Errorf("player count %d does not match config %d", n, j.Config.PlayerCnt)
	}
	// **リードスートは 0（トリック未開始）か実在するスート。** 素通しすると
	// フォロー義務がどの札にも掛からなくなり、規則が黙って消えます。
	// このゲームは 0 が正当な状態なので、場札の有無と突き合わせる。
	if j.LeadSuit == 0 {
		if len(j.Pile) > 0 {
			return fmt.Errorf("pile holds %d cards with no led suit", len(j.Pile))
		}
	} else {
		if j.LeadSuit < CardDesignSpade || j.LeadSuit > CardDesignDiamond {
			return fmt.Errorf("invalid led suit: %d", j.LeadSuit)
		}
		if len(j.Pile) == 0 {
			return fmt.Errorf("led suit %d with an empty pile", j.LeadSuit)
		}
	}
	if len(j.Pile) > BhabhiDeckSize {
		return fmt.Errorf("pile holds %d cards", len(j.Pile))
	}
	if len(j.ActionLog) > bhabhiMaxSliceLen {
		return errors.New("bhabhi: input array exceeds maximum allowed size")
	}
	if j.TrickNumber < 0 || j.TrickNumber > BhabhiStalemateTricks {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if j.FinishedCnt < 0 || j.FinishedCnt > n {
		return fmt.Errorf("invalid finished count: %d", j.FinishedCnt)
	}
	for name, idx := range map[string]int{"current player": j.CurrentIdx, "lead player": j.LeadIdx} {
		if idx < 0 || idx >= n {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	for name, idx := range map[string]int{"last pickup": j.LastPickupIdx, "bhabhi": j.BhabhiIdx} {
		if idx < -1 || idx >= n {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	// **敗者が決まっているのは終局後だけ。** 進行中に載っていたら壊れている。
	if !j.GameEndFlag && j.BhabhiIdx != -1 {
		return fmt.Errorf("bhabhi %d before the game ended", j.BhabhiIdx)
	}
	if j.Stalemate && !j.GameEndFlag {
		return errors.New("stalemate before the game ended")
	}
	if j.LastPickupSize < 0 || j.LastPickupSize > BhabhiDeckSize {
		return fmt.Errorf("invalid pickup size: %d", j.LastPickupSize)
	}

	b.players = j.Players
	b.config = j.Config
	b.phase = j.Phase
	b.pile = j.Pile
	b.leadSuit = j.LeadSuit
	b.currentIdx = j.CurrentIdx
	b.leadIdx = j.LeadIdx
	b.trickNumber = j.TrickNumber
	b.lastPickupIdx = j.LastPickupIdx
	b.lastPickupSize = j.LastPickupSize
	b.finishedCnt = j.FinishedCnt
	b.gameEndFlag = j.GameEndFlag
	b.bhabhiIdx = j.BhabhiIdx
	b.stalemate = j.Stalemate
	b.actionLog = j.ActionLog
	return nil
}
