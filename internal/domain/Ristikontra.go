//go:build !js || !wasm || extra2

// Package domain implements Ristikontra — a Finnish fishing / capture game.
//
// # Rules summary
//
// Ristikontra is always a **fixed 2-vs-2 partnership** (seats 0・2 against
// seats 1・3; seat 0 is the human) played with a standard 52-card deck.
// There is **no trump suit**. A face-up pile sits in the centre and on a turn
// the active player plays one hand card onto it.
//
// A player CAPTURES the whole pile when the played card's RANK equals the
// current pile-top's rank. That is the only way to capture.
//
// # Ristikontra (the counter) — the mechanic the game is named for
//
// risti-kontra means "cross-counter". Immediately after a capture, a player who
// lays another card of **the same rank that made that capture** takes the
// captured bundle away from the capturer, and it keeps going: the bundle can be
// stolen back and forth for as long as players keep matching that rank. The
// chain ends the moment somebody plays a different rank.
//
// # Divergence from the clone source (Pişti)
//
// This domain is derived from Pişti, and the differences are the point:
//
//   - **No wild Jack.** Pişti lets a Jack capture anything; here only an exact
//     rank match captures.
//   - **No Pişti bonus.** Pişti awards +10 (or +20 Jack-on-Jack) for capturing a
//     lone card; Ristikontra has no such award.
//   - **Fixed partnerships.** Pişti seats compete individually and the table can
//     be 2–4 players; Ristikontra is always 4 players in two teams.
//   - **The counter.** Pişti has no way to take a capture back.
//
// # Dealing
//
// Each player is dealt RistikontraHandSize (4) cards and
// RistikontraInitialPileSize (4) cards start the centre pile. When all hands are
// empty another 4 are dealt to each player from the stock; this repeats until
// the stock is exhausted.
//
// # Final scoring
//
// When the stock and all hands are exhausted, any remaining centre pile goes to
// the team of the player who captured last. **Each team then totals the cards it
// captured and the larger total wins** — there is no per-card point schedule and
// no per-seat contest. An equal split leaves the game drawn (GetWinners returns
// no seats).
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// RistikontraPhase はリスティコントラのフェーズを表す。
type RistikontraPhase string

// リスティコントラのフェーズ定数。
const (
	// RistikontraPhasePlay プレイ中 (カードを場へ出す)
	RistikontraPhasePlay RistikontraPhase = "play"
	// RistikontraPhaseRoundEnd ラウンド (配り直し) 締め処理。内部遷移用。
	RistikontraPhaseRoundEnd RistikontraPhase = "roundEnd"
	// RistikontraPhaseGameEnd ゲーム終了
	RistikontraPhaseGameEnd RistikontraPhase = "gameEnd"
)

// ristikontraMaxSliceLen はデシリアライズ時のスライス長上限。
const ristikontraMaxSliceLen = 1000

// ristikontraState はゲーム進行状態。
type ristikontraState struct {
	phase          RistikontraPhase
	currentTurn    int     // 現在の手番プレイヤー
	pile           []*Card // 場の山 (末尾が一番上)
	lastCaptureIdx int     // 最後に捕獲したプレイヤー (-1 = なし)
	// **直前の捕獲を覚えておく。** リスティコントラ (risti-kontra = 十字の
	// 打ち返し) の核心は、捕獲された直後に同ランクを出すと**山ごと奪い返せる**
	// こと。奪い返しは続く限り連鎖する。捕獲で山を空にしてしまうと、この
	// 打ち返しに使う札束が消えるので、別に持っておく。
	counterCards []*Card // 打ち返しの対象になっている捕獲済みの束 (nil = 対象なし)
	counterRank  int     // その捕獲を成立させたランク (0 = 対象なし)
	gameEndFlag  bool
	winners      []int
	actionLogBase
}

// Ristikontra はリスティコントラ ゲームの状態を保持する集約ルート。
type Ristikontra struct {
	trumpCards *TrumpCards
	players    []*RistikontraPlayer
	config     RistikontraConfig
	state      ristikontraState
}

// NewRistikontra コンストラクタ。
func NewRistikontra(trumpCards *TrumpCards, players []*RistikontraPlayer, config RistikontraConfig) *Ristikontra {
	return &Ristikontra{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		state: ristikontraState{
			phase:          RistikontraPhasePlay,
			lastCaptureIdx: -1,
		},
	}
}

// NewDefaultRistikontra はデフォルト設定 (4人: 1 human + 3 CPU) のリスティコントラを返す。
func NewDefaultRistikontra() *Ristikontra {
	config := DefaultRistikontraConfig()
	players := makeRistikontraPlayers(config.PlayerCnt)
	return NewRistikontra(NewTrumpCards(0), players, config)
}

// makeRistikontraPlayers は seat 0 を人間とした cnt 人のプレイヤースライスを作る。
func makeRistikontraPlayers(cnt int) []*RistikontraPlayer {
	if cnt < RistikontraMinPlayerCnt {
		cnt = RistikontraMinPlayerCnt
	}
	if cnt > RistikontraMaxPlayerCnt {
		cnt = RistikontraMaxPlayerCnt
	}
	players := make([]*RistikontraPlayer, cnt)
	players[0] = NewRistikontraPlayer(true)
	for i := 1; i < cnt; i++ {
		players[i] = NewRistikontraPlayer(false)
	}
	return players
}

// Reset は新しいゲームを開始する。
func (g *Ristikontra) Reset() {
	g.players = makeRistikontraPlayers(g.config.PlayerCnt)
	g.trumpCards = NewTrumpCards(0)
	g.trumpCards.Shuffle()
	g.state = ristikontraState{
		phase:          RistikontraPhasePlay,
		currentTurn:    0,
		lastCaptureIdx: -1,
		actionLogBase:  actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
	}
	g.dealHands()
	g.dealInitialPile()
}

// NextRound はリスティコントラ では新規ゲーム開始 (Reset) と同義。
// 1 ゲームは山札を配り切るまでの 1 セッションで完結するため、終局後の続行は
// 新しいゲームの開始として扱う。
func (g *Ristikontra) NextRound() {
	g.Reset()
}

// dealHands は各プレイヤーへ RistikontraHandSize 枚配る。山札が尽きたら途中で終わる。
func (g *Ristikontra) dealHands() {
	for k := 0; k < RistikontraHandSize; k++ {
		for i := 0; i < len(g.players); i++ {
			card := g.trumpCards.DrawCard()
			if card == nil {
				return
			}
			g.players[i].AddCard(card)
		}
	}
}

// dealInitialPile はゲーム開始時に RistikontraInitialPileSize 枚を場へ置く。
// 一番上 (最後に置く札) がジャックの場合は、続けて山札から引いた非ジャック札と
// 入れ替え、ジャックを山の下に潜らせて非ジャックを上にする (issue #2329 のルール:
// 開始時の一番上がジャックにならないようにする)。山札が尽きた場合は引き直しを
// 諦め、ジャックのまま開始する。TrumpCards は引いた札を山へ戻せないため、
// 「戻して引き直す」代わりに余分に 1 枚引いて入れ替える等価な処理とする。
func (g *Ristikontra) dealInitialPile() {
	for i := 0; i < RistikontraInitialPileSize; i++ {
		card := g.trumpCards.DrawCard()
		if card == nil {
			break
		}
		g.state.pile = append(g.state.pile, card)
	}
	// **一番上のジャックを避ける処理は要らない。** クローン元のピシュティは
	// ジャックが万能の捕獲札なので、初手でいきなり全部さらわれないよう
	// 上に来ないようにしていた。リスティコントラのジャックはただの札。
	g.appendLog(-1, "deal", fmt.Sprintf("dealt %d pile cards", len(g.state.pile)), append([]*Card(nil), g.state.pile...))
}

// RistikontraTeamCnt はチーム数。席 0・2 = チーム 0、席 1・3 = チーム 1。
const RistikontraTeamCnt = 2

// ristikontraTeamOf は席番号からチーム番号を返す。**常に 2 対 2 の固定
// パートナーシップ**で、クローン元のピシュティのような可変人数戦ではない。
func ristikontraTeamOf(seat int) int { return seat % RistikontraTeamCnt }

// calcTeamCardCounts はチームごとの獲得枚数を返す。勝敗はこの合計で決まる。
func (g *Ristikontra) calcTeamCardCounts() []int {
	counts := make([]int, RistikontraTeamCnt)
	for i, pl := range g.players {
		counts[ristikontraTeamOf(i)] += pl.CapturedCount()
	}
	return counts
}

// GetTeamCardCounts はチームごとの獲得枚数を返す (表示用)。
func (g *Ristikontra) GetTeamCardCounts() []int { return g.calcTeamCardCounts() }

// pileTop は場の一番上の札を返す (なければ nil)。
func (g *Ristikontra) pileTop() *Card {
	if len(g.state.pile) == 0 {
		return nil
	}
	return g.state.pile[len(g.state.pile)-1]
}

// allHandsEmpty は全員の手札が空かどうか。
func (g *Ristikontra) allHandsEmpty() bool {
	return allHandsEmpty(g.players)
}

// PlayerPlay は人間プレイヤーが手札 cardIndex を場へ出す。
func (g *Ristikontra) PlayerPlay(cardIndex int) error {
	if g.state.gameEndFlag {
		return ErrGameEnded
	}
	if g.state.phase != RistikontraPhasePlay {
		return NewDomainError(ErrWrongPhase, "not in play phase")
	}
	if !g.players[g.state.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyPlay(g.state.currentTurn, cardIndex)
}

// CpuPlay は CPU のターンを 1 回進める。
func (g *Ristikontra) CpuPlay() {
	if g.state.gameEndFlag || g.state.phase != RistikontraPhasePlay {
		return
	}
	p := g.players[g.state.currentTurn]
	if p.GetIsHuman() || p.GetCardsSize() == 0 {
		return
	}
	idx := g.chooseCpuCard(g.state.currentTurn)
	_ = g.applyPlay(g.state.currentTurn, idx)
}

// applyPlay は cardIndex の手札を場へ出し、捕獲判定・打ち返し判定・手番進行を行う。
func (g *Ristikontra) applyPlay(playerIdx, cardIndex int) error {
	player := g.players[playerIdx]
	card := player.GetCard(cardIndex)
	if card == nil {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", cardIndex))
	}

	top := g.pileTop()
	counters := g.isCounter(card)
	captures := !counters && g.isCapture(card, top)

	_ = player.RemoveCard(cardIndex)

	switch {
	case counters:
		// **打ち返し。** 直前の捕獲を成立させたランクをそのまま被せると、
		// 相手が取った束ごと自分のものになる。奪われた側の獲得枚数から
		// 束を引き、こちらに移す。連鎖はここから続く。
		stolen := g.state.counterCards
		if prev := g.state.lastCaptureIdx; prev >= 0 && prev < len(g.players) {
			g.players[prev].RemoveCaptured(stolen)
		}
		taken := append(append([]*Card(nil), stolen...), card)
		player.AddCaptured(taken)
		g.state.lastCaptureIdx = playerIdx
		g.state.counterCards = taken
		g.state.counterRank = card.GetValue()
		g.appendLog(playerIdx, "counter",
			fmt.Sprintf("countered for %d card(s)", len(taken)), taken)

	case captures:
		g.state.pile = append(g.state.pile, card)
		taken := append([]*Card(nil), g.state.pile...)
		player.AddCaptured(taken)
		g.state.pile = g.state.pile[:0]
		g.state.lastCaptureIdx = playerIdx
		// 次の 1 手だけ、この束は打ち返しの的になる。
		g.state.counterCards = taken
		g.state.counterRank = card.GetValue()
		g.appendLog(playerIdx, "capture",
			fmt.Sprintf("captured %d card(s)", len(taken)), taken)

	default:
		// 捕獲も打ち返しも起きなかったので、打ち返しの機会は閉じる。
		g.state.counterCards = nil
		g.state.counterRank = 0
		g.state.pile = append(g.state.pile, card)
		g.appendLog(playerIdx, "play", "played onto pile", []*Card{card})
	}

	g.advanceTurn()
	return nil
}

// isCapture は card を top に重ねたとき捕獲が成立するかを返す。
// 場が空 (top == nil) の場合は捕獲不成立。
func (g *Ristikontra) isCapture(card, top *Card) bool {
	if top == nil {
		return false
	}
	// **ジャックはワイルドではない。** クローン元のピシュティはジャックを
	// 万能の捕獲札にし、単独札を取ると Pişti ボーナスが付くが、
	// リスティコントラにはどちらも無く、成立するのは同ランクだけ。
	return card.GetValue() == top.GetValue()
}

// isCounter は card が直前の捕獲を奪い返せるかを返す。
//
// 捕獲を成立させたランクと同じ札を、その直後に出すと、**捕獲された束ごと**
// 自分のものになる。これが「リスティコントラ (打ち返し)」で、続く限り連鎖する。
func (g *Ristikontra) isCounter(card *Card) bool {
	return g.state.counterCards != nil && card != nil &&
		card.GetValue() == g.state.counterRank
}

// advanceTurn は手番を次に進め、必要なら配り直し・終局処理を行う。
func (g *Ristikontra) advanceTurn() {
	g.state.currentTurn = (g.state.currentTurn + 1) % len(g.players)
	if !g.allHandsEmpty() {
		return
	}
	if g.trumpCards.GetRemainingCount() > 0 {
		g.dealHands()
		return
	}
	g.finishGame()
}

// finishGame は終局処理: 残りの場札を最後の捕獲者へ渡し、最終得点を確定する。
func (g *Ristikontra) finishGame() {
	g.state.phase = RistikontraPhaseRoundEnd
	if g.state.lastCaptureIdx >= 0 && len(g.state.pile) > 0 {
		leftover := append([]*Card(nil), g.state.pile...)
		g.players[g.state.lastCaptureIdx].AddCaptured(leftover)
		g.appendLog(g.state.lastCaptureIdx, "lastTake", fmt.Sprintf("last-take: %d card(s)", len(leftover)), leftover)
	}
	g.state.pile = g.state.pile[:0]

	// **勝敗はチーム単位。** 席ごとの枚数ではなく、パートナーと合わせた
	// 合計で決める。クローン元のピシュティは各席が独立に競うので、ここが
	// いちばん大きな作り替えになる。
	teamCounts := g.calcTeamCardCounts()
	bestTeam, bestCount, tied := -1, -1, false
	for team, c := range teamCounts {
		switch {
		case c > bestCount:
			bestTeam, bestCount, tied = team, c, false
		case c == bestCount:
			tied = true
		}
	}
	winners := make([]int, 0)
	if !tied && bestTeam >= 0 {
		for i := range g.players {
			if ristikontraTeamOf(i) == bestTeam {
				winners = append(winners, i)
			}
		}
	}
	g.state.winners = winners
	maxScore := bestCount
	g.state.gameEndFlag = true
	g.state.phase = RistikontraPhaseGameEnd
	g.appendLog(-1, "gameEnd", fmt.Sprintf("game ended (top score %d)", maxScore), nil)
}

// calcFinalScore は各プレイヤーの最終得点を計算する。
// インデックスはプレイヤーシートに対応する。
func (g *Ristikontra) calcFinalScore() []int {
	// **表示する得点と勝敗の判定を同じ数字にする。**
	//
	// これはクローン元 (ピシュティ) の席ごとのカード点 (A +1 / 2♣ +2 /
	// 10♦ +3 / J +1 に最多捕獲 +3) をそのまま持ってきていた。finishGame は
	// チームの獲得枚数で勝者を決めるので、**負けたチームのほうが画面上の
	// 得点は高い**という盤面が普通に出る。勝敗と表示が食い違うのは、
	// プレイヤーにとっては「勝ったはずなのに負けと言われる」に等しい。
	//
	// リスティコントラの結果は枚数そのものなので、席にはその席が属する
	// チームの合計を入れる。GetProvisionalScores と同じ数字が、そのまま
	// 最終結果になる。
	counts := g.calcTeamCardCounts()
	scores := make([]int, len(g.players))
	for i := range g.players {
		scores[i] = counts[ristikontraTeamOf(i)]
	}
	return scores
}

// mostCapturedSeat は最多捕獲の単独リーダーの席を返す。同数、または誰も
// 捕獲していなければ -1。
func (g *Ristikontra) mostCapturedSeat() int {
	best, bestSeat, tie := 0, -1, false
	for i, p := range g.players {
		cnt := p.CapturedCount()
		if bestSeat == -1 || cnt > best {
			best, bestSeat, tie = cnt, i, false
		} else if cnt == best {
			tie = true
		}
	}
	if tie || best == 0 {
		return -1
	}
	return bestSeat
}

// GetProvisionalScores は対局中の暫定スコアを返す。
//
// 各席には**その席が属するチームの合計枚数**が入る。勝敗はチーム単位なので、
// 席ごとの枚数を見せると「自分は取っているのに負けている」が読めなくなる。
func (g *Ristikontra) GetProvisionalScores() []int {
	// **チームの獲得枚数をそのまま返す。** 札ごとの点数も Pişti ボーナスも
	// 無いので、途中経過は「今どちらのチームが何枚持っているか」で言い切れる。
	// クローン元のピシュティは最後にカード点を数えるため暫定値が近似だったが、
	// ここは近似ではない。
	counts := g.calcTeamCardCounts()
	out := make([]int, len(g.players))
	for i := range g.players {
		out[i] = counts[ristikontraTeamOf(i)]
	}
	return out
}

// GetProvisionalLeader は暫定の最多捕獲リーダーの席を返す (同数なら -1)。
func (g *Ristikontra) GetProvisionalLeader() int { return g.mostCapturedSeat() }

// chooseCpuCard は CPU の手番で出す手札インデックスを選ぶ。
//   - Easy   : 合法手 (常に全札合法) からランダム。
//   - Normal : 捕獲できる札を優先、無ければランクの低い札を捨てる。
//   - Hard   : **打ち返しを最優先**、次に捕獲、無ければランクの低い札を捨てる。
//
// Hard の分岐はクローン元では「場が 1 枚なら Pişti を狙う」だったが、それは
// ボーナスがあってこその優先順位で、しかも直後の一般分岐と同じ札を返す
// 二度手間だった。このゲームで一番大きいのは打ち返しなので、そこを見る。
func (g *Ristikontra) chooseCpuCard(playerIdx int) int {
	player := g.players[playerIdx]
	size := player.GetCardsSize()
	if size == 0 {
		return 0
	}
	top := g.pileTop()

	switch g.config.CpuDifficulty {
	case RistikontraDifficultyEasy:
		return rand.Intn(size)
	case RistikontraDifficultyHard:
		// 直前の捕獲を奪えるなら、それが盤面で一番大きい振れ幅。
		if idx := g.findCounterCard(player); idx >= 0 {
			return idx
		}
		if idx := g.findCapturingCard(player, top); idx >= 0 {
			return idx
		}
		return g.lowestValueCardIdx(player)
	default: // Normal
		if idx := g.findCapturingCard(player, top); idx >= 0 {
			return idx
		}
		return g.lowestValueCardIdx(player)
	}
}

// findCounterCard は直前の捕獲を奪える手札のインデックスを返す (なければ -1)。
func (g *Ristikontra) findCounterCard(player *RistikontraPlayer) int {
	if g.state.counterCards == nil {
		return -1
	}
	for i := 0; i < player.GetCardsSize(); i++ {
		if c := player.GetCard(i); c != nil && c.GetValue() == g.state.counterRank {
			return i
		}
	}
	return -1
}

// findCapturingCard は捕獲できる手札のインデックスを返す (なければ -1)。
// 複数ある場合はジャック以外の rank-match を優先し、ジャックは温存する。
func (g *Ristikontra) findCapturingCard(player *RistikontraPlayer, top *Card) int {
	if top == nil {
		return -1
	}
	// **ジャックは万能札ではない。** クローン元のピシュティはここで
	// 「同ランクが無ければジャックを出す」と探していたが、
	// リスティコントラで捕獲できるのは同ランクだけ。
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c == nil {
			continue
		}
		if c.GetValue() == top.GetValue() {
			return i
		}
	}
	return -1
}

// lowestValueCardIdx は捨てるのに一番惜しくない手札のインデックスを返す。
// 同値なら最初に見つかったものを返す。
//
// **札ごとの点数は無いので、ランクの低さで選ぶ。** クローン元のピシュティは
// A +1 / 2♣ +2 / 10♦ +3 / J +1 という配点があり、それを避けて捨てていた。
// リスティコントラの結果は枚数だけなので、その配点で選ぶと理由の無い基準に
// なる (2♣ を後生大事に抱える、など)。低いランクほど後で捕獲に使いにくいので、
// 先に手放す。
func (g *Ristikontra) lowestValueCardIdx(player *RistikontraPlayer) int {
	best := 0
	bestRank := 1 << 30
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c == nil {
			continue
		}
		if v := c.GetValue(); v < bestRank {
			bestRank = v
			best = i
		}
	}
	return best
}

// appendLog は棋譜にエントリを追加する。
func (g *Ristikontra) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.state.appendLog(playerIdx, actionType, detail, cards)
}

// --- 状態アクセサ ---

// IsHumanTurn は現在の手番が人間かどうかを返す。
func (g *Ristikontra) IsHumanTurn() bool {
	if g.state.gameEndFlag {
		return false
	}
	return g.players[g.state.currentTurn].GetIsHuman()
}

// GetCurrentTurn は現在の手番プレイヤーインデックスを返す。
func (g *Ristikontra) GetCurrentTurn() int { return g.state.currentTurn }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *Ristikontra) GetGameEndFlag() bool { return g.state.gameEndFlag }

// GetPhase は現在のフェーズを返す。
func (g *Ristikontra) GetPhase() RistikontraPhase { return g.state.phase }

// GetPile は場の山を返す (末尾が一番上)。
func (g *Ristikontra) GetPile() []*Card { return g.state.pile }

// GetPileTop は場の一番上の札を返す (なければ nil)。
// GetCounterRank は打ち返しの対象になっているランクを返す (0 = 対象なし)。
//
// **直前の捕獲だけが的になる。** このランクを今この場で出せば束ごと奪えるので、
// UI とヒントはここを見て「奪える手がある」と伝えられる。
func (g *Ristikontra) GetCounterRank() int { return g.state.counterRank }

func (g *Ristikontra) GetPileTop() *Card { return g.pileTop() }

// GetLastCaptureIdx は最後に捕獲したプレイヤーを返す (-1 = なし)。
func (g *Ristikontra) GetLastCaptureIdx() int { return g.state.lastCaptureIdx }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *Ristikontra) GetPlayer(idx int) *RistikontraPlayer {
	return getPlayer(g.players, idx)
}

// GetPlayerCnt はプレイヤー数を返す。
func (g *Ristikontra) GetPlayerCnt() int { return len(g.players) }

// GetRemainingDeck は山札の残り枚数を返す。
func (g *Ristikontra) GetRemainingDeck() int { return g.trumpCards.GetRemainingCount() }

// GetConfig は設定を返す。
func (g *Ristikontra) GetConfig() RistikontraConfig { return g.config }

// SetConfig は設定を変更する。
func (g *Ristikontra) SetConfig(config RistikontraConfig) { g.config = config }

// SetGameEndFlagForTest はテスト用に終了フラグを設定する。
func (g *Ristikontra) SetGameEndFlagForTest(v bool) { g.state.gameEndFlag = v }

// GetActionLog は棋譜を返す。
func (g *Ristikontra) GetActionLog() []*ActionLogEntry { return g.state.actionLog }

// GetWinners はゲーム終了時の勝者シートのリストを返す (同点なら複数)。
func (g *Ristikontra) GetWinners() []int { return g.state.winners }

// GetFinalScores は現在の最終得点を計算して返す (途中経過でも算出可能)。
func (g *Ristikontra) GetFinalScores() []int { return g.calcFinalScore() }

// --- JSON Serialization ---

// ristikontraJSON is the JSON wire format for Ristikontra.
type ristikontraJSON struct {
	TrumpCards     *TrumpCards          `json:"tc"`
	Players        []*RistikontraPlayer `json:"pl"`
	Config         RistikontraConfig    `json:"cf"`
	Phase          RistikontraPhase     `json:"ph"`
	CurrentTurn    int                  `json:"ct"`
	Pile           []*Card              `json:"pi"`
	LastCaptureIdx int                  `json:"lc"`
	GameEndFlag    bool                 `json:"ge"`
	Winners        []int                `json:"wn"`
	ActionLog      []*ActionLogEntry    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Ristikontra) MarshalJSON() ([]byte, error) {
	return json.Marshal(ristikontraJSON{
		TrumpCards:     g.trumpCards,
		Players:        g.players,
		Config:         g.config,
		Phase:          g.state.phase,
		CurrentTurn:    g.state.currentTurn,
		Pile:           g.state.pile,
		LastCaptureIdx: g.state.lastCaptureIdx,
		GameEndFlag:    g.state.gameEndFlag,
		Winners:        g.state.winners,
		ActionLog:      g.state.actionLog,
	})
}

// ristikontraValidPhase は有効なフェーズかどうか。
func ristikontraValidPhase(p RistikontraPhase) bool {
	switch p {
	case RistikontraPhasePlay, RistikontraPhaseRoundEnd, RistikontraPhaseGameEnd:
		return true
	default:
		return false
	}
}

// UnmarshalJSON implements json.Unmarshaler. 不正な永続化データを拒否するための
// バリデーションを行う。
func (g *Ristikontra) UnmarshalJSON(data []byte) error {
	var j ristikontraJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > ristikontraMaxSliceLen || len(j.Pile) > ristikontraMaxSliceLen ||
		len(j.ActionLog) > ristikontraMaxSliceLen || len(j.Winners) > ristikontraMaxSliceLen {
		return fmt.Errorf("ristikontra: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("ristikontra: invalid config: %w", err)
	}
	if j.TrumpCards == nil {
		return fmt.Errorf("ristikontra: missing trump cards in state")
	}
	if len(j.Players) < RistikontraMinPlayerCnt || len(j.Players) > RistikontraMaxPlayerCnt {
		return fmt.Errorf("ristikontra: player count out of range")
	}
	for _, p := range j.Players {
		if p == nil {
			return fmt.Errorf("ristikontra: nil player in state")
		}
	}
	if !ristikontraValidPhase(j.Phase) {
		return fmt.Errorf("ristikontra: invalid phase")
	}
	if j.CurrentTurn < 0 || j.CurrentTurn >= len(j.Players) {
		return fmt.Errorf("ristikontra: current turn out of range")
	}
	if j.LastCaptureIdx < -1 || j.LastCaptureIdx >= len(j.Players) {
		return fmt.Errorf("ristikontra: last capture index out of range")
	}
	for _, w := range j.Winners {
		if w < 0 || w >= len(j.Players) {
			return fmt.Errorf("ristikontra: winner index out of range")
		}
	}

	g.trumpCards = j.TrumpCards
	g.players = j.Players
	g.config = j.Config
	g.state = ristikontraState{
		phase:          j.Phase,
		currentTurn:    j.CurrentTurn,
		pile:           j.Pile,
		lastCaptureIdx: j.LastCaptureIdx,
		gameEndFlag:    j.GameEndFlag,
		winners:        j.Winners,
		actionLogBase:  actionLogBase{actionLog: j.ActionLog},
	}
	if g.state.pile == nil {
		g.state.pile = make([]*Card, 0)
	}
	if g.state.actionLog == nil {
		g.state.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
