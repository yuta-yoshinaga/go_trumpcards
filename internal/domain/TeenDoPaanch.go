//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// TeenDoPaanchPhase は 3-2-5 のゲームフェーズ。
type TeenDoPaanchPhase int

// TeenDoPaanch のフェーズ定数
const (
	// TeenDoPaanchPhaseTrump 切り札の宣言（ノルマ 5 の人が最初の 5 枚だけを見て決める）
	TeenDoPaanchPhaseTrump TeenDoPaanchPhase = iota
	// TeenDoPaanchPhasePlay プレイ中
	TeenDoPaanchPhasePlay
	// TeenDoPaanchPhaseRoundEnd ラウンド終了
	TeenDoPaanchPhaseRoundEnd
	// TeenDoPaanchPhaseGameEnd ゲーム終了
	TeenDoPaanchPhaseGameEnd
)

// TeenDoPaanchPlayerCnt はプレイヤー数（3 人固定）。
const TeenDoPaanchPlayerCnt = 3

// TeenDoPaanchTargets は 3 つのノルマ。**宣言ではなく最初から割り当てられます。**
// 名前の「3-2-5（ティーン・ドー・パーンチ）」がそのままこの 3 つの数です。
var TeenDoPaanchTargets = [TeenDoPaanchPlayerCnt]int{3, 2, 5}

// TeenDoPaanchTricksPerRound は 1 ラウンドのトリック数。
//
// **3 + 2 + 5 = 10 で、トリックが余りも不足もしません。**
const TeenDoPaanchTricksPerRound = 3 + 2 + 5

// TeenDoPaanchHandSize は各プレイヤーの手札枚数。
const TeenDoPaanchHandSize = TeenDoPaanchTricksPerRound

// TeenDoPaanchDeckSize は使用するデッキの枚数。
//
// **issue の「30枚（8〜A）」は算術が合いません。** 8,9,10,J,Q,K,A は 7 ランク
// なので 7 × 4 = **28 枚**にしかならず、3 人 × 10 枚に 2 枚足りません。実際の
// 3-2-5 は 28 枚に **7♠ と 7♥ の 2 枚**を足した 30 枚デッキを使い、これで
// ちょうど配り切れます（NewTrumpCardsTeenDoPaanch を参照）。
const TeenDoPaanchDeckSize = TeenDoPaanchPlayerCnt * TeenDoPaanchHandSize

// TeenDoPaanchFirstDeal は切り札を決める前に配る枚数。
//
// **手札が揃う前に切り札を決めるのが賭けどころ。** 10 枚見てから決めるのでは
// 意味が変わってしまいます。
const TeenDoPaanchFirstDeal = 5

// TeenDoPaanchDefaultRounds は既定のラウンド数。
//
// **3 ラウンドで全員が 3・2・5 を一度ずつ引き受けます。**
const TeenDoPaanchDefaultRounds = 3

// teenDoPaanchMaxSliceLen caps slice sizes during deserialisation.
const teenDoPaanchMaxSliceLen = 1000

// TeenDoPaanch は 3-2-5（ティーン・ドー・パーンチ）のゲームクラス。
//
// インド北部の 3 人専用トリックテイキング。30 枚を 10 枚ずつ配り、10 トリックを
// 打ちます。
//
// **ノルマは宣言するものではなく、最初から割り当てられています。** 3 人が
// それぞれ 3・2・5 トリックという**別々のノルマ**を負い、役割はラウンドごとに
// 回ります。全員が同じ目標を追う既存のトリックテイカーとはここが違います。
//
// **多く取ってもうれしくありません。** ノルマちょうど以上を取れば達成、
// 届かなければ未達成で、余分に取ったぶんは次のラウンドで**相手の良い札を
// 召し上げる権利**に変わります。
type TeenDoPaanch struct {
	players     []*TeenDoPaanchPlayer
	config      TeenDoPaanchConfig
	phase       TeenDoPaanchPhase
	trumpCards  *TrumpCards
	trumpSuit   int
	roundNumber int
	trickNumber int
	// fivePlayerIdx はノルマ 5 を負う席。**この人が切り札を決めます。**
	fivePlayerIdx    int
	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int
	// surplus は前ラウンドの過不足（+ が超過、- が不足）。次の配りで札の
	// やり取りに使い、使い切ったら 0 に戻します。
	surplus []int
	// lastExchange は直前に動いた札の枚数（0 = やり取り無し）。
	lastExchange int
	gameEndFlag  bool
	// winnerIdx は勝者 (-1: 未確定/同点)。
	winnerIdx int
	actionLogBase
}

// NewTeenDoPaanch はコンストラクタ。
func NewTeenDoPaanch(players []*TeenDoPaanchPlayer, config TeenDoPaanchConfig) *TeenDoPaanch {
	return &TeenDoPaanch{
		players:       players,
		config:        config,
		fivePlayerIdx: 0,
		surplus:       make([]int, TeenDoPaanchPlayerCnt),
		winnerIdx:     -1,
	}
}

// NewDefaultTeenDoPaanch は標準の 3 人セットアップを返す。
func NewDefaultTeenDoPaanch() *TeenDoPaanch {
	players := make([]*TeenDoPaanchPlayer, 0, TeenDoPaanchPlayerCnt)
	players = append(players, NewTeenDoPaanchPlayer(true))
	for range TeenDoPaanchPlayerCnt - 1 {
		players = append(players, NewTeenDoPaanchPlayer(false))
	}
	return NewTeenDoPaanch(players, DefaultTeenDoPaanchConfig())
}

// teenDoPaanchRank は札の強さ。**A が最強**、7 が最弱。
func teenDoPaanchRank(c *Card) int {
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// Reset はゲームを初期化する。
func (g *TeenDoPaanch) Reset() {
	g.roundNumber = 0
	g.fivePlayerIdx = 0
	g.surplus = make([]int, TeenDoPaanchPlayerCnt)
	g.lastExchange = 0
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.actionLog = nil
	for _, p := range g.players {
		p.ResetGame()
	}
	g.startRound()
}

// startRound は 1 ラウンドを配り直す。
//
// **切り札を決める前に配るのは 5 枚だけ。** 残りは宣言のあとに配ります。
func (g *TeenDoPaanch) startRound() {
	g.phase = TeenDoPaanchPhaseTrump
	g.trumpSuit = 0
	g.trickNumber = 0
	g.currentTrick = nil
	g.lastExchange = 0
	for _, p := range g.players {
		p.ResetRound()
	}
	g.assignTargets()

	g.trumpCards = NewTrumpCardsTeenDoPaanch()
	g.trumpCards.Shuffle()
	for range TeenDoPaanchFirstDeal {
		for i := range TeenDoPaanchPlayerCnt {
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[i].AddCard(c)
			}
		}
	}
	g.sortAllHands()

	g.roundNumber++
	g.currentPlayerIdx = g.fivePlayerIdx
	g.leadPlayerIdx = g.fivePlayerIdx
	g.addLog(-1, "deal", fmt.Sprintf("ラウンド %d：ノルマ 3/2/5 を割り当てました", g.roundNumber), nil)
}

// assignTargets はノルマを席へ割り当てる。**5 を起点に 3・2 と回します。**
func (g *TeenDoPaanch) assignTargets() {
	for i := range TeenDoPaanchPlayerCnt {
		idx := (g.fivePlayerIdx + i) % TeenDoPaanchPlayerCnt
		// TeenDoPaanchTargets は {3,2,5} なので、5 が先頭に来るよう並べ替える。
		switch i {
		case 0:
			g.players[idx].SetTarget(5)
		case 1:
			g.players[idx].SetTarget(3)
		default:
			g.players[idx].SetTarget(2)
		}
	}
}

// DeclareTrump はノルマ 5 の人が切り札を宣言する。
func (g *TeenDoPaanch) DeclareTrump(suit int) error {
	if g.gameEndFlag {
		return errors.New("game is over")
	}
	if g.phase != TeenDoPaanchPhaseTrump {
		return errors.New("trump has already been declared")
	}
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return fmt.Errorf("invalid suit: %d", suit)
	}
	g.trumpSuit = suit
	g.completeDeal()
	g.exchangeCards()
	g.phase = TeenDoPaanchPhasePlay
	g.currentPlayerIdx = g.leadPlayerIdx
	g.addLog(g.fivePlayerIdx, "trump", fmt.Sprintf("切り札は %s", suitStr(suit)), nil)
	return nil
}

// PlayerDeclareTrump は人間が切り札を宣言する。
func (g *TeenDoPaanch) PlayerDeclareTrump(suit int) error {
	if g.phase != TeenDoPaanchPhaseTrump || g.fivePlayerIdx != 0 {
		return errors.New("not your call")
	}
	return g.DeclareTrump(suit)
}

// CpuDeclareTrump は CPU が切り札を宣言する。
func (g *TeenDoPaanch) CpuDeclareTrump() {
	if g.phase != TeenDoPaanchPhaseTrump || g.fivePlayerIdx == 0 || g.gameEndFlag {
		return
	}
	_ = g.DeclareTrump(g.chooseCpuTrump(g.fivePlayerIdx))
}

// chooseCpuTrump は CPU の切り札選び。**5 枚しか見えないので枚数がいちばん
// 確かな情報。** 同数なら強い札を持っているほうを採ります。
func (g *TeenDoPaanch) chooseCpuTrump(playerIdx int) int {
	p := g.players[playerIdx]
	count := map[int]int{}
	strength := map[int]int{}
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		count[c.GetDesign()]++
		strength[c.GetDesign()] += teenDoPaanchRank(c)
	}
	best, bestCount, bestStrength := CardDesignSpade, -1, -1
	for _, s := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if count[s] > bestCount || (count[s] == bestCount && strength[s] > bestStrength) {
			best, bestCount, bestStrength = s, count[s], strength[s]
		}
	}
	return best
}

// completeDeal は残りの札を配り切る。
func (g *TeenDoPaanch) completeDeal() {
	for g.trumpCards.GetRemainingCount() > 0 {
		for i := range TeenDoPaanchPlayerCnt {
			if g.players[i].GetCardsSize() >= TeenDoPaanchHandSize {
				continue
			}
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[i].AddCard(c)
			}
		}
	}
	g.sortAllHands()
}

// exchangeCards は前ラウンドの過不足で札をやり取りする。
//
// **ノルマを超えた人が、届かなかった人の良い札を召し上げます。** 超過ぶんだけ
// 相手の最強札を取り、代わりに自分の最弱札を渡します。多く取ることに意味が
// 生まれるのはここだけで、そのラウンドの得点にはなりません。
func (g *TeenDoPaanch) exchangeCards() {
	moved := 0
	for taker := range TeenDoPaanchPlayerCnt {
		for g.surplus[taker] > 0 {
			giver := g.nextDeficit()
			if giver < 0 {
				break
			}
			g.moveBestCard(giver, taker)
			g.surplus[taker]--
			g.surplus[giver]++
			moved++
		}
	}
	g.surplus = make([]int, TeenDoPaanchPlayerCnt)
	g.lastExchange = moved
	if moved > 0 {
		g.sortAllHands()
		g.addLog(-1, "exchange", fmt.Sprintf("前ラウンドの過不足で %d 枚を移しました", moved), nil)
	}
}

// nextDeficit はまだノルマに届かなかったぶんが残っている席を返す (-1 = 無し)。
func (g *TeenDoPaanch) nextDeficit() int {
	for i := range TeenDoPaanchPlayerCnt {
		if g.surplus[i] < 0 {
			return i
		}
	}
	return -1
}

// moveBestCard は giver の最強札を taker に渡し、taker の最弱札を giver に返す。
func (g *TeenDoPaanch) moveBestCard(giver, taker int) {
	gp, tp := g.players[giver], g.players[taker]
	if gp.GetCardsSize() == 0 || tp.GetCardsSize() == 0 {
		return
	}
	bi, brank := 0, -1
	for i := range gp.GetCardsSize() {
		if r := teenDoPaanchRank(gp.GetCard(i)); r > brank {
			bi, brank = i, r
		}
	}
	wi, wrank := 0, 1<<30
	for i := range tp.GetCardsSize() {
		if r := teenDoPaanchRank(tp.GetCard(i)); r < wrank {
			wi, wrank = i, r
		}
	}
	best := gp.RemoveCard(bi)
	worst := tp.RemoveCard(wi)
	if best != nil {
		tp.AddCard(best)
	}
	if worst != nil {
		gp.AddCard(worst)
	}
}

// sortAllHands は手札をスート・ランク順に整える。
func (g *TeenDoPaanch) sortAllHands() {
	for _, p := range g.players {
		sortPlayerHand(p, func(ci, cj *Card) bool {
			if ci.GetDesign() != cj.GetDesign() {
				return ci.GetDesign() < cj.GetDesign()
			}
			return teenDoPaanchRank(ci) < teenDoPaanchRank(cj)
		})
	}
}

// GetValidPlayIndices は playerIdx が出せる手札の添字を返す。
//
// **フォロー義務あり。** リードスートを持っていればそれしか出せません。
func (g *TeenDoPaanch) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= TeenDoPaanchPlayerCnt || g.gameEndFlag {
		return []int{}
	}
	p := g.players[playerIdx]
	if g.phase != TeenDoPaanchPhasePlay {
		return []int{}
	}
	all := make([]int, 0, p.GetCardsSize())
	follow := make([]int, 0, p.GetCardsSize())
	leadSuit := 0
	if len(g.currentTrick) > 0 {
		leadSuit = g.currentTrick[0].Card.GetDesign()
	}
	for i := range p.GetCardsSize() {
		all = append(all, i)
		if leadSuit != 0 && p.GetCard(i).GetDesign() == leadSuit {
			follow = append(follow, i)
		}
	}
	if len(follow) > 0 {
		return follow
	}
	return all
}

// IsHumanTurn は現在の手番が人間かを返す。
func (g *TeenDoPaanch) IsHumanTurn() bool {
	if g.gameEndFlag || g.phase != TeenDoPaanchPhasePlay {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// IsHumanTrumpTurn は人間が切り札を宣言する番かを返す。
func (g *TeenDoPaanch) IsHumanTrumpTurn() bool {
	return !g.gameEndFlag && g.phase == TeenDoPaanchPhaseTrump && g.fivePlayerIdx == 0
}

// PlayerPlay は人間が 1 枚出す。
func (g *TeenDoPaanch) PlayerPlay(cardIndex int) error {
	if !g.IsHumanTurn() {
		return errors.New("not your turn")
	}
	return g.play(0, cardIndex)
}

// CpuPlay は CPU が 1 枚出す。
func (g *TeenDoPaanch) CpuPlay() {
	if g.gameEndFlag || g.phase != TeenDoPaanchPhasePlay || g.IsHumanTurn() {
		return
	}
	_ = g.play(g.currentPlayerIdx, g.chooseCpuCard(g.currentPlayerIdx))
}

// play は playerIdx に手札の cardIndex を出させる。
func (g *TeenDoPaanch) play(playerIdx, cardIndex int) error {
	if g.gameEndFlag {
		return errors.New("game is over")
	}
	if g.phase != TeenDoPaanchPhasePlay {
		return errors.New("not the play phase")
	}
	if playerIdx != g.currentPlayerIdx {
		return fmt.Errorf("not player %d's turn", playerIdx)
	}
	p := g.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return fmt.Errorf("invalid card index: %d", cardIndex)
	}
	if !teenDoPaanchContains(g.GetValidPlayIndices(playerIdx), cardIndex) {
		return errors.New("must follow the led suit")
	}

	card := p.RemoveCard(cardIndex)
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.addLog(playerIdx, "play", cardStr(card), []*Card{card})

	if len(g.currentTrick) < TeenDoPaanchPlayerCnt {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % TeenDoPaanchPlayerCnt
		return nil
	}
	g.resolveTrick()
	return nil
}

// resolveTrick はトリックを解決する。
func (g *TeenDoPaanch) resolveTrick() {
	winner := g.trickWinner()
	cards := make([]*Card, 0, TeenDoPaanchPlayerCnt)
	for _, tc := range g.currentTrick {
		cards = append(cards, tc.Card)
	}
	g.players[winner].AddTrick(cards)
	g.currentTrick = nil
	g.trickNumber++
	g.leadPlayerIdx = winner
	g.currentPlayerIdx = winner
	g.addLog(winner, "trick", fmt.Sprintf("トリック %d を取りました", g.trickNumber), nil)

	if g.trickNumber >= TeenDoPaanchTricksPerRound {
		g.finishRound()
	}
}

// trickWinner は切り札 > リードスートの順で最強札を出した人を返す。
func (g *TeenDoPaanch) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return g.leadPlayerIdx
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	best, bestRank, bestTrump := g.currentTrick[0].PlayerIdx, -1, false
	for _, tc := range g.currentTrick {
		suit, rank := tc.Card.GetDesign(), teenDoPaanchRank(tc.Card)
		isTrump := g.trumpSuit != 0 && suit == g.trumpSuit
		switch {
		case isTrump && !bestTrump:
			best, bestRank, bestTrump = tc.PlayerIdx, rank, true
		case isTrump == bestTrump && suit == leadSuit && !bestTrump && rank > bestRank:
			best, bestRank = tc.PlayerIdx, rank
		case isTrump && bestTrump && rank > bestRank:
			best, bestRank = tc.PlayerIdx, rank
		}
	}
	return best
}

// finishRound はラウンドを精算する。
//
// **ノルマちょうど以上で達成。** 多く取っても得点は増えず、超過ぶんは次の
// ラウンドで相手の良い札を召し上げる権利に変わります。
func (g *TeenDoPaanch) finishRound() {
	g.phase = TeenDoPaanchPhaseRoundEnd
	for i, p := range g.players {
		diff := p.GetTrickCount() - p.GetTarget()
		g.surplus[i] = diff
		if diff >= 0 {
			p.AddMet()
			g.addLog(i, "score", fmt.Sprintf("ノルマ %d に対し %d トリック：達成",
				p.GetTarget(), p.GetTrickCount()), nil)
		} else {
			g.addLog(i, "score", fmt.Sprintf("ノルマ %d に対し %d トリック：未達成",
				p.GetTarget(), p.GetTrickCount()), nil)
		}
	}
	if g.roundNumber >= g.config.Rounds {
		g.finishGame()
	}
}

// NextRound は次のラウンドを開始する。
func (g *TeenDoPaanch) NextRound() {
	if g.gameEndFlag || g.phase != TeenDoPaanchPhaseRoundEnd {
		return
	}
	// **ノルマ 5 は席を移ります。** 3 ラウンドで全員が一巡します。
	g.fivePlayerIdx = (g.fivePlayerIdx + 1) % TeenDoPaanchPlayerCnt
	g.startRound()
}

// finishGame は終局処理。**ノルマ達成回数がいちばん多い人の勝ち。**
func (g *TeenDoPaanch) finishGame() {
	g.phase = TeenDoPaanchPhaseGameEnd
	g.gameEndFlag = true
	best, bestMet, tie := -1, -1, false
	for i, p := range g.players {
		switch {
		case p.GetMet() > bestMet:
			best, bestMet, tie = i, p.GetMet(), false
		case p.GetMet() == bestMet:
			tie = true
		}
	}
	if tie {
		best = -1
	}
	g.winnerIdx = best
	g.addLog(-1, "result", "ゲーム終了", nil)
}

// GiveUp は投了する。
func (g *TeenDoPaanch) GiveUp() {
	if g.gameEndFlag {
		return
	}
	g.phase = TeenDoPaanchPhaseGameEnd
	g.gameEndFlag = true
	g.winnerIdx = -1
	g.addLog(0, "giveup", "投了しました", nil)
}

// chooseCpuCard は CPU の手。
//
// **多く取ってもうれしくないゲームなので、ノルマに届いたら降りにいきます。**
func (g *TeenDoPaanch) chooseCpuCard(playerIdx int) int {
	valid := g.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := g.players[playerIdx]
	wantTrick := p.GetTrickCount() < p.GetTarget()

	pick, pickRank := valid[0], teenDoPaanchRank(p.GetCard(valid[0]))
	for _, i := range valid[1:] {
		r := teenDoPaanchRank(p.GetCard(i))
		if wantTrick && r > pickRank {
			pick, pickRank = i, r
		} else if !wantTrick && r < pickRank {
			pick, pickRank = i, r
		}
	}
	return pick
}

// TeenDoPaanchHint は 3-2-5 の助言。
type TeenDoPaanchHint struct {
	CardIndex *int
	Reason    string
	// Suit は切り札に勧めるスート（プレイ中は 0）。
	Suit int
}

// GetHint は人間への助言を返す。
func (g *TeenDoPaanch) GetHint() *TeenDoPaanchHint {
	if g.gameEndFlag {
		return nil
	}
	if g.IsHumanTrumpTurn() {
		return &TeenDoPaanchHint{Reason: "teendopaanchSelectTrump", Suit: g.chooseCpuTrump(0)}
	}
	if !g.IsHumanTurn() {
		return nil
	}
	valid := g.GetValidPlayIndices(0)
	if len(valid) == 0 {
		return nil
	}
	idx := g.chooseCpuCard(0)
	p := g.players[0]
	reason := "teendopaanchDuck"
	if p.GetTrickCount() < p.GetTarget() {
		reason = "teendopaanchWinTrick"
	}
	return &TeenDoPaanchHint{CardIndex: &idx, Reason: reason}
}

// teenDoPaanchContains は xs が v を含むかを返す。
func teenDoPaanchContains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// addLog は棋譜に 1 行足す。
func (g *TeenDoPaanch) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.appendLog(playerIdx, actionType, detail, cards)
}

// --- アクセサ ---------------------------------------------------------------

// GetConfig はゲーム設定を返す。
func (g *TeenDoPaanch) GetConfig() TeenDoPaanchConfig { return g.config }

// SetConfig はゲーム設定を設定する。
func (g *TeenDoPaanch) SetConfig(cfg TeenDoPaanchConfig) { g.config = cfg }

// GetPhase は現在のフェーズを返す。
func (g *TeenDoPaanch) GetPhase() TeenDoPaanchPhase { return g.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *TeenDoPaanch) GetGameEndFlag() bool { return g.gameEndFlag }

// GetRoundNumber は現在のラウンド番号を返す。
func (g *TeenDoPaanch) GetRoundNumber() int { return g.roundNumber }

// GetTrickNumber は現在のトリック番号を返す。
func (g *TeenDoPaanch) GetTrickNumber() int { return g.trickNumber }

// GetTrumpSuit は切り札のスートを返す（0 = 未宣言）。
func (g *TeenDoPaanch) GetTrumpSuit() int { return g.trumpSuit }

// GetFivePlayerIdx はノルマ 5 を負う席（切り札を決める席）を返す。
func (g *TeenDoPaanch) GetFivePlayerIdx() int { return g.fivePlayerIdx }

// GetCurrentPlayerIdx は現在の手番を返す。
func (g *TeenDoPaanch) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// GetLeadPlayerIdx はリードプレイヤーを返す。
func (g *TeenDoPaanch) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// GetCurrentTrick は現在のトリックを返す。
func (g *TeenDoPaanch) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// GetLastExchange は直前のラウンド間で動いた札の枚数を返す。
func (g *TeenDoPaanch) GetLastExchange() int { return g.lastExchange }

// GetPlayerCnt はプレイヤー数を返す。
func (g *TeenDoPaanch) GetPlayerCnt() int { return TeenDoPaanchPlayerCnt }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *TeenDoPaanch) GetPlayer(i int) *TeenDoPaanchPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetWinnerIdx は勝者を返す（-1 = 未確定/同点）。
func (g *TeenDoPaanch) GetWinnerIdx() int { return g.winnerIdx }

// GetActionLog は棋譜を返す。
func (g *TeenDoPaanch) GetActionLog() []*ActionLogEntry { return g.actionLog }

// teenDoPaanchJSON は KV スナップショットの表現。
//
// **非公開フィールドは全部ここに載せる。** Worker はリクエストごとに JSON から
// ゲームを作り直すので、載せ忘れたものは毎回消えます (#4478)。
type teenDoPaanchJSON struct {
	TrumpCards       *TrumpCards           `json:"tc"`
	Players          []*TeenDoPaanchPlayer `json:"pl"`
	Config           TeenDoPaanchConfig    `json:"cf"`
	Phase            TeenDoPaanchPhase     `json:"ph"`
	TrumpSuit        int                   `json:"ts"`
	RoundNumber      int                   `json:"rn"`
	TrickNumber      int                   `json:"tn"`
	FivePlayerIdx    int                   `json:"fp"`
	CurrentTrick     []*TrickCard          `json:"ct"`
	CurrentPlayerIdx int                   `json:"ci"`
	LeadPlayerIdx    int                   `json:"li"`
	Surplus          []int                 `json:"sp"`
	LastExchange     int                   `json:"le"`
	GameEndFlag      bool                  `json:"ge"`
	WinnerIdx        int                   `json:"wi"`
	ActionLog        []*ActionLogEntry     `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (g *TeenDoPaanch) MarshalJSON() ([]byte, error) {
	return json.Marshal(&teenDoPaanchJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		TrumpSuit:        g.trumpSuit,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		FivePlayerIdx:    g.fivePlayerIdx,
		CurrentTrick:     g.currentTrick,
		CurrentPlayerIdx: g.currentPlayerIdx,
		LeadPlayerIdx:    g.leadPlayerIdx,
		Surplus:          g.surplus,
		LastExchange:     g.lastExchange,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		ActionLog:        g.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元
func (g *TeenDoPaanch) UnmarshalJSON(data []byte) error {
	var j teenDoPaanchJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < TeenDoPaanchPhaseTrump || j.Phase > TeenDoPaanchPhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	// **切り札は宣言フェーズのあいだだけ 0。** 素通しすると trickWinner が
	// どの札も切り札と見なさなくなり、勝敗が黙って変わる (#5302〜#5305)。
	if j.Phase == TeenDoPaanchPhaseTrump {
		if j.TrumpSuit != 0 {
			return fmt.Errorf("trump suit %d before it was declared", j.TrumpSuit)
		}
	} else if j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond {
		return fmt.Errorf("invalid trump suit: %d", j.TrumpSuit)
	}
	if j.RoundNumber < 1 || j.RoundNumber > TeenDoPaanchRoundsMax {
		return fmt.Errorf("invalid round number: %d", j.RoundNumber)
	}
	if j.TrickNumber < 0 || j.TrickNumber > TeenDoPaanchTricksPerRound {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if len(j.CurrentTrick) > TeenDoPaanchPlayerCnt {
		return fmt.Errorf("current trick holds %d cards", len(j.CurrentTrick))
	}
	if len(j.ActionLog) > teenDoPaanchMaxSliceLen {
		return errors.New("teendopaanch: input array exceeds maximum allowed size")
	}
	for name, idx := range map[string]int{
		"current player": j.CurrentPlayerIdx,
		"lead player":    j.LeadPlayerIdx,
		"five player":    j.FivePlayerIdx,
	} {
		if idx < 0 || idx >= TeenDoPaanchPlayerCnt {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= TeenDoPaanchPlayerCnt {
		return fmt.Errorf("invalid winner: %d", j.WinnerIdx)
	}
	// **勝者が決まっているのは終局後だけ。**
	if !j.GameEndFlag && j.WinnerIdx != -1 {
		return fmt.Errorf("winner %d before the game ended", j.WinnerIdx)
	}
	// **過不足は合計 0。** 誰かの超過は必ず誰かの不足なので、崩れていたら壊れている。
	if len(j.Surplus) != TeenDoPaanchPlayerCnt {
		return fmt.Errorf("surplus holds %d entries", len(j.Surplus))
	}
	sum := 0
	for _, v := range j.Surplus {
		if v < -TeenDoPaanchTricksPerRound || v > TeenDoPaanchTricksPerRound {
			return fmt.Errorf("invalid surplus: %d", v)
		}
		sum += v
	}
	if sum != 0 {
		return fmt.Errorf("surplus sums to %d, not 0", sum)
	}
	if j.LastExchange < 0 || j.LastExchange > TeenDoPaanchTricksPerRound {
		return fmt.Errorf("invalid exchange size: %d", j.LastExchange)
	}

	if j.TrumpCards != nil {
		g.trumpCards = j.TrumpCards
	}
	if len(j.Players) == TeenDoPaanchPlayerCnt {
		g.players = j.Players
	}
	g.config = j.Config
	g.phase = j.Phase
	g.trumpSuit = j.TrumpSuit
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.fivePlayerIdx = j.FivePlayerIdx
	g.currentTrick = j.CurrentTrick
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.surplus = j.Surplus
	g.lastExchange = j.LastExchange
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.actionLog = j.ActionLog
	return nil
}
