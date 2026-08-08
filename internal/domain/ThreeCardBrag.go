//go:build !js || !wasm || casino

// Package domain スリーカード・ブラグ (Three Card Brag) のドメインモデル。
//
// Three Card Brag はイギリス発祥のポーカーの祖となる 3 枚ベッティングゲーム。52 枚デッキを
// 使い、4 人で全員アンティを入れてから 3 枚ずつ配る。各プレイヤーは手札を見ない「Blind」か、
// 見る「Seen」かを選べる。Blind は Seen の半額 (本実装では Blind=stake / Seen=2*stake) で賭けられ、
// いつでも手札を見て Seen に昇格できる。
//
// ベッティングはアンティ後、時計回りにコール / レイズ / フォールドを繰り返す。最後の 1 人に
// なればポット獲得。残り 2 人になると Seen プレイヤーは「Show」(2*stake を払って勝負) を要求でき、
// 手を比べて強い方がポットを取る (引き分けは Show を要求した側の負け)。
//
// 役の強さ (高い順): Prial (スリーカード) > Running Flush (ストレートフラッシュ) > Run
// (ストレート) > Flush > Pair > High Card。特例として 3-3-3 が最強の Prial、A-2-3 が最強の Run。
//
// チップが尽きたプレイヤーは脱落し、最後に残った 1 人が試合の勝者となる。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

// 役カテゴリ (高いほど強い)
const (
	// ThreeCardBragHighCard ハイカード
	ThreeCardBragHighCard = 0
	// ThreeCardBragPair ペア
	ThreeCardBragPair = 1
	// ThreeCardBragFlush フラッシュ
	ThreeCardBragFlush = 2
	// ThreeCardBragRun ラン (ストレート)
	ThreeCardBragRun = 3
	// ThreeCardBragRunningFlush ランニングフラッシュ (ストレートフラッシュ)
	ThreeCardBragRunningFlush = 4
	// ThreeCardBragPrial プライアル (スリーカード)
	ThreeCardBragPrial = 5
)

// threeCardBragMaxActions は 1 ディールあたりの賭けアクション上限 (安全網)。
// 通常はフォールド / Show で終わるが、全員コールし続ける状況を強制的に決着させる。
const threeCardBragMaxActions = 24

// ThreeCardBragPhase ゲームフェーズ
type ThreeCardBragPhase int

// Three Card Brag のフェーズ定数
const (
	// ThreeCardBragPhaseBetting ベッティングフェーズ
	ThreeCardBragPhaseBetting ThreeCardBragPhase = iota
	// ThreeCardBragPhaseShowdown ショーダウン (手を公開)
	ThreeCardBragPhaseShowdown
	// ThreeCardBragPhaseRoundEnd ディール終了
	ThreeCardBragPhaseRoundEnd
	// ThreeCardBragPhaseGameEnd ゲーム終了
	ThreeCardBragPhaseGameEnd
)

// ThreeCardBragHint ヒント情報
type ThreeCardBragHint struct {
	Action string // "see"/"bet"/"raise"/"fold"/"show"
	Reason string // ヒント理由キー
}

// threeCardBragRank はカードのブラグ順位を返す (A=14, K=13 ... 2=2)。
func threeCardBragRank(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// ThreeCardBragEval は 3 枚の手の役カテゴリと比較用タイブレーク列を返す。
// タイブレークは降順で、カテゴリが同じならこの列を辞書順比較すれば勝敗が決まる。
func ThreeCardBragEval(cards []*Card) (int, []int) {
	if len(cards) != ThreeCardBragHandSize {
		return -1, nil
	}
	r := []int{threeCardBragRank(cards[0]), threeCardBragRank(cards[1]), threeCardBragRank(cards[2])}
	sort.Sort(sort.Reverse(sort.IntSlice(r)))
	flush := cards[0].GetDesign() == cards[1].GetDesign() && cards[1].GetDesign() == cards[2].GetDesign()
	prial := r[0] == r[1] && r[1] == r[2]
	isRun, runHigh := threeCardBragRun(r)

	switch {
	case prial:
		// 3-3-3 が最強。それ以外はランク順 (A-A-A 以下)。
		pr := r[0]
		if r[0] == 3 {
			pr = 15
		}
		return ThreeCardBragPrial, []int{pr}
	case isRun && flush:
		return ThreeCardBragRunningFlush, []int{runHigh}
	case isRun:
		return ThreeCardBragRun, []int{runHigh}
	case flush:
		return ThreeCardBragFlush, r
	case r[0] == r[1] || r[1] == r[2]:
		// ペア + キッカー。
		pairRank, kicker := r[1], r[2]
		if r[1] == r[2] {
			pairRank, kicker = r[1], r[0]
		}
		return ThreeCardBragPair, []int{pairRank, kicker}
	default:
		return ThreeCardBragHighCard, r
	}
}

// threeCardBragRun は降順ランク列がラン (連続) かを返す。A-2-3 は最強ラン (high=15)。
func threeCardBragRun(r []int) (bool, int) {
	// A-2-3 (降順では 14,3,2)
	if r[0] == 14 && r[1] == 3 && r[2] == 2 {
		return true, 15
	}
	if r[0]-1 == r[1] && r[1]-1 == r[2] {
		return true, r[0]
	}
	return false, 0
}

// ThreeCardBragCompare は手 a が手 b に勝てば 1、負ければ -1、引き分けは 0 を返す。
// 引数は ThreeCardBragEval が返す (カテゴリ, タイブレーク列) の組。
func ThreeCardBragCompare(catA int, tbA []int, catB int, tbB []int) int {
	if catA != catB {
		if catA > catB {
			return 1
		}
		return -1
	}
	for i := 0; i < len(tbA) && i < len(tbB); i++ {
		if tbA[i] != tbB[i] {
			if tbA[i] > tbB[i] {
				return 1
			}
			return -1
		}
	}
	return 0
}

// ThreeCardBrag スリーカード・ブラグ ゲームクラス
type ThreeCardBrag struct {
	trumpCards       *TrumpCards
	players          []*ThreeCardBragPlayer
	config           ThreeCardBragConfig
	phase            ThreeCardBragPhase
	roundNumber      int
	dealerIdx        int
	currentPlayerIdx int
	pot              int
	stake            int // 現在の賭け単位 (Blind=stake / Seen=2*stake)
	lastAggressorIdx int
	actionCount      int
	roundWinnerIdx   int // 直近ディールの勝者 (-1: 未確定)
	showdown         bool
	gameEndFlag      bool
	matchWinnerIdx   int // 試合の勝者 (-1: 未確定)
	actionLogBase
}

// NewThreeCardBrag コンストラクタ
func NewThreeCardBrag(trumpCards *TrumpCards, players []*ThreeCardBragPlayer, config ThreeCardBragConfig) *ThreeCardBrag {
	return &ThreeCardBrag{
		trumpCards:     trumpCards,
		players:        players,
		config:         config,
		roundWinnerIdx: -1,
		matchWinnerIdx: -1,
	}
}

// NewDefaultThreeCardBrag 標準の 4 人セットアップ (人間 idx 0 + CPU 3) を返す。
func NewDefaultThreeCardBrag() *ThreeCardBrag {
	cfg := DefaultThreeCardBragConfig()
	players := make([]*ThreeCardBragPlayer, ThreeCardBragPlayerCnt)
	for i := range players {
		players[i] = NewThreeCardBragPlayer(i == 0, cfg.StartingChips)
	}
	return NewThreeCardBrag(NewTrumpCards(0), players, cfg)
}

// Reset ゲーム初期化
func (g *ThreeCardBrag) Reset() {
	g.gameEndFlag = false
	g.matchWinnerIdx = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.actionLog = nil
	for _, p := range g.players {
		p.SetChips(g.config.StartingChips)
		p.SetOut(false)
	}
	g.startDeal()
}

// NextRound 次のディールを開始する
func (g *ThreeCardBrag) NextRound() {
	if g.phase != ThreeCardBragPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = g.nextActive(g.dealerIdx)
	g.startDeal()
}

// startDeal 1 ディールを開始する: アンティ徴収・配札・ベッティング開始。
func (g *ThreeCardBrag) startDeal() {
	g.pot = 0
	g.showdown = false
	g.roundWinnerIdx = -1
	g.actionCount = 0
	g.stake = g.config.Ante

	for _, p := range g.players {
		p.ResetForDeal()
	}
	// チップ不足のプレイヤーは脱落。
	for _, p := range g.players {
		if !p.GetOut() && p.GetChips() < g.config.Ante {
			p.SetOut(true)
		}
	}
	// アンティを払えるプレイヤーが 1 人以下なら試合終了 (1 人だけのベッティングループを防ぐ)。
	if g.aliveCount() <= 1 {
		g.gameEndFlag = true
		g.phase = ThreeCardBragPhaseGameEnd
		g.matchWinnerIdx = g.firstAlive()
		g.appendLog(g.matchWinnerIdx, "game_end", fmt.Sprintf("%s wins the match", playerName(g.players, g.matchWinnerIdx)), nil)
		return
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	// アンティ徴収 + 配札。
	for _, p := range g.players {
		if p.GetOut() {
			continue
		}
		p.SubtractChips(g.config.Ante)
		p.AddRoundBet(g.config.Ante)
		g.pot += g.config.Ante
		for i := 0; i < ThreeCardBragHandSize; i++ {
			if c := g.trumpCards.DrawCard(); c != nil {
				p.AddCard(c)
			}
		}
	}
	g.phase = ThreeCardBragPhaseBetting
	g.currentPlayerIdx = g.nextActive(g.dealerIdx)
	g.lastAggressorIdx = g.currentPlayerIdx
	g.appendLog(-1, "deal", fmt.Sprintf("Deal %d: ante %d, pot %d", g.roundNumber, g.config.Ante, g.pot), nil)
}

// callCost は playerIdx が現在の stake をコールするのに必要な額を返す (Blind は半額)。
func (g *ThreeCardBrag) callCost(playerIdx int) int {
	if g.players[playerIdx].GetSeen() {
		return g.stake * 2
	}
	return g.stake
}

// --- Actions ---

// PlayerSee 人間プレイヤーが手札を見て Seen に昇格する (Blind 時のみ)。手番は消費しない。
func (g *ThreeCardBrag) PlayerSee() error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	if g.players[g.currentPlayerIdx].GetSeen() {
		return NewDomainError(ErrInvalidPlay, "すでに手札を見ています")
	}
	g.players[g.currentPlayerIdx].SetSeen(true)
	g.appendLog(g.currentPlayerIdx, "see", fmt.Sprintf("%s sees their hand", playerName(g.players, g.currentPlayerIdx)), nil)
	return nil
}

// PlayerBet 人間プレイヤーが現在の stake をコール (またはアンティ後の最初の賭け) する。
func (g *ThreeCardBrag) PlayerBet() error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	return g.applyCall(g.currentPlayerIdx)
}

// PlayerRaise 人間プレイヤーが stake を newStake へ引き上げて賭ける。
func (g *ThreeCardBrag) PlayerRaise(newStake int) error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	if newStake <= g.stake {
		return NewDomainError(ErrInvalidPlay, "レイズは現在の賭け単位より大きくする必要があります")
	}
	return g.applyRaise(g.currentPlayerIdx, newStake)
}

// PlayerFold 人間プレイヤーが降りる。
func (g *ThreeCardBrag) PlayerFold() error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	g.applyFold(g.currentPlayerIdx)
	return nil
}

// PlayerShow 残り 2 人のとき、Seen プレイヤーが Show (勝負) を要求する。
func (g *ThreeCardBrag) PlayerShow() error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	if !g.canShow(g.currentPlayerIdx) {
		return NewDomainError(ErrInvalidPlay, "Show は残り 2 人かつ Seen のときのみ要求できます")
	}
	g.applyShow(g.currentPlayerIdx)
	return nil
}

// checkTurn は現在の手番が人間かつベッティングフェーズかを検証する。
func (g *ThreeCardBrag) checkTurn() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ThreeCardBragPhaseBetting {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return nil
}

// applyCall コールを適用する。チップ不足なら自動フォールド。
func (g *ThreeCardBrag) applyCall(idx int) error {
	cost := g.callCost(idx)
	p := g.players[idx]
	if p.GetChips() < cost {
		// 払えない場合はフォールド扱い。
		g.applyFold(idx)
		return nil
	}
	p.SubtractChips(cost)
	p.AddRoundBet(cost)
	g.pot += cost
	g.actionCount++
	g.appendLog(idx, "bet", fmt.Sprintf("%s bets %d (pot %d)", playerName(g.players, idx), cost, g.pot), nil)
	g.advanceOrResolve(idx)
	return nil
}

// applyRaise レイズを適用する。
func (g *ThreeCardBrag) applyRaise(idx, newStake int) error {
	g.stake = newStake
	cost := g.callCost(idx)
	p := g.players[idx]
	if p.GetChips() < cost {
		return NewDomainError(ErrInvalidPlay, "チップが不足しています")
	}
	p.SubtractChips(cost)
	p.AddRoundBet(cost)
	g.pot += cost
	g.lastAggressorIdx = idx
	g.actionCount++
	g.appendLog(idx, "raise", fmt.Sprintf("%s raises to %d, bets %d (pot %d)", playerName(g.players, idx), newStake, cost, g.pot), nil)
	g.advanceOrResolve(idx)
	return nil
}

// applyFold フォールドを適用する。
func (g *ThreeCardBrag) applyFold(idx int) {
	g.players[idx].SetFolded(true)
	g.appendLog(idx, "fold", fmt.Sprintf("%s folds", playerName(g.players, idx)), nil)
	if g.activeCount() == 1 {
		g.endDeal([]int{g.firstActive()})
		return
	}
	g.advanceOrResolve(idx)
}

// applyShow Show を適用し、残り 2 人のショーダウンを行う。
func (g *ThreeCardBrag) applyShow(idx int) {
	cost := g.stake * 2
	p := g.players[idx]
	if p.GetChips() < cost {
		cost = p.GetChips()
	}
	p.SubtractChips(cost)
	p.AddRoundBet(cost)
	g.pot += cost
	g.appendLog(idx, "show", fmt.Sprintf("%s pays %d to see (pot %d)", playerName(g.players, idx), cost, g.pot), nil)
	opp := g.otherActive(idx)
	g.showdown = true
	g.phase = ThreeCardBragPhaseShowdown
	cmp := g.compareHands(idx, opp)
	switch {
	case cmp > 0:
		g.endDeal([]int{idx})
	case cmp < 0:
		g.endDeal([]int{opp})
	default:
		// 引き分けは Show を要求した側の負け。
		g.endDeal([]int{opp})
	}
}

// advanceOrResolve 次の手番へ進める。安全網の上限に達したら強制ショーダウン。
func (g *ThreeCardBrag) advanceOrResolve(fromIdx int) {
	if g.actionCount >= threeCardBragMaxActions {
		g.forcedShowdown()
		return
	}
	g.currentPlayerIdx = g.nextActive(fromIdx)
}

// forcedShowdown 残り全アクティブを公開し、最強手にポットを与える (安全網)。
func (g *ThreeCardBrag) forcedShowdown() {
	g.showdown = true
	g.phase = ThreeCardBragPhaseShowdown
	winners := g.bestActiveHands()
	g.endDeal(winners)
}

// --- CPU ---

// CpuAct 現在の手番が CPU の場合に 1 アクション実行する。
func (g *ThreeCardBrag) CpuAct() {
	if g.gameEndFlag || g.phase != ThreeCardBragPhaseBetting {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	p := g.players[idx]
	// Blind の場合はまず手札を見る (簡易 AI)。
	if !p.GetSeen() {
		p.SetSeen(true)
		g.appendLog(idx, "see", fmt.Sprintf("%s sees their hand", playerName(g.players, idx)), nil)
	}
	cat, _ := g.handEval(idx)
	cost := g.callCost(idx)
	twoLeft := g.activeCount() == 2
	switch {
	case cat >= ThreeCardBragRun:
		// 強い手: 残り 2 人なら Show、そうでなければレイズ。
		if twoLeft && g.canShow(idx) {
			g.applyShow(idx)
			return
		}
		if g.stake < g.config.Ante*8 && p.GetChips() >= g.callCost(idx)*2 {
			_ = g.applyRaise(idx, g.stake+g.config.Ante)
			return
		}
		_ = g.applyCall(idx)
	case cat >= ThreeCardBragPair:
		// 中程度: コール (高すぎればフォールド)。
		if cost > p.GetChips() || cost > g.config.Ante*6 {
			g.applyFold(idx)
			return
		}
		_ = g.applyCall(idx)
	default:
		// 弱い手: 安ければコール、高ければフォールド。
		if cost <= g.config.Ante {
			_ = g.applyCall(idx)
			return
		}
		g.applyFold(idx)
	}
}

// --- Showdown / scoring ---

// endDeal ディールを終了し、ポットを winners に分配して試合終了を判定する。
func (g *ThreeCardBrag) endDeal(winners []int) {
	if len(winners) == 0 {
		winners = []int{g.firstActive()}
	}
	share := g.pot / len(winners)
	rem := g.pot - share*len(winners)
	for i, w := range winners {
		amt := share
		if i == 0 {
			amt += rem
		}
		g.players[w].AddChips(amt)
	}
	g.roundWinnerIdx = winners[0]
	g.appendLog(winners[0], "win", fmt.Sprintf("%s wins the pot (%d)", playerName(g.players, winners[0]), g.pot), nil)

	// チップ 0 は脱落。
	for _, p := range g.players {
		if !p.GetOut() && p.GetChips() <= 0 {
			p.SetOut(true)
		}
	}
	if g.aliveCount() <= 1 {
		g.gameEndFlag = true
		g.phase = ThreeCardBragPhaseGameEnd
		g.matchWinnerIdx = g.firstAlive()
		g.appendLog(g.matchWinnerIdx, "game_end", fmt.Sprintf("%s wins the match", playerName(g.players, g.matchWinnerIdx)), nil)
		return
	}
	g.phase = ThreeCardBragPhaseRoundEnd
}

// bestActiveHands は未フォールドプレイヤーのうち最強手のインデックス列を返す (引き分けは複数)。
func (g *ThreeCardBrag) bestActiveHands() []int {
	best := []int{}
	var bestCat int
	var bestTb []int
	for i, p := range g.players {
		if p.GetOut() || p.GetFolded() {
			continue
		}
		cat, tb := g.handEval(i)
		if len(best) == 0 {
			best, bestCat, bestTb = []int{i}, cat, tb
			continue
		}
		switch ThreeCardBragCompare(cat, tb, bestCat, bestTb) {
		case 1:
			best, bestCat, bestTb = []int{i}, cat, tb
		case 0:
			best = append(best, i)
		}
	}
	return best
}

// handEval は playerIdx の手の役を返す。
func (g *ThreeCardBrag) handEval(playerIdx int) (int, []int) {
	p := g.players[playerIdx]
	cards := make([]*Card, 0, ThreeCardBragHandSize)
	for i := 0; i < p.GetCardsSize(); i++ {
		cards = append(cards, p.GetCard(i))
	}
	return ThreeCardBragEval(cards)
}

// compareHands は手 a と手 b を比較する (a 勝ち=1, b 勝ち=-1, 引き分け=0)。
func (g *ThreeCardBrag) compareHands(a, b int) int {
	catA, tbA := g.handEval(a)
	catB, tbB := g.handEval(b)
	return ThreeCardBragCompare(catA, tbA, catB, tbB)
}

// --- helpers ---

// canShow は idx が Show を要求できるか (残り 2 人 & Seen) を返す。
func (g *ThreeCardBrag) canShow(idx int) bool {
	return g.activeCount() == 2 && g.players[idx].GetSeen()
}

// activeCount 未フォールド・非脱落のプレイヤー数。
func (g *ThreeCardBrag) activeCount() int {
	n := 0
	for _, p := range g.players {
		if !p.GetOut() && !p.GetFolded() {
			n++
		}
	}
	return n
}

// firstActive 最初の未フォールド・非脱落プレイヤー。
func (g *ThreeCardBrag) firstActive() int {
	for i, p := range g.players {
		if !p.GetOut() && !p.GetFolded() {
			return i
		}
	}
	return 0
}

// otherActive idx 以外の未フォールド・非脱落プレイヤー (残り 2 人の相手)。
func (g *ThreeCardBrag) otherActive(idx int) int {
	for i, p := range g.players {
		if i != idx && !p.GetOut() && !p.GetFolded() {
			return i
		}
	}
	return idx
}

// nextActive from の次の未フォールド・非脱落プレイヤー。
func (g *ThreeCardBrag) nextActive(from int) int {
	for i := 1; i <= ThreeCardBragPlayerCnt; i++ {
		idx := (from + i) % ThreeCardBragPlayerCnt
		if !g.players[idx].GetOut() && !g.players[idx].GetFolded() {
			return idx
		}
	}
	return from
}

// aliveCount チップを持つ (非脱落) プレイヤー数。
func (g *ThreeCardBrag) aliveCount() int {
	n := 0
	for _, p := range g.players {
		if !p.GetOut() {
			n++
		}
	}
	return n
}

// firstAlive 最初の非脱落プレイヤー。
func (g *ThreeCardBrag) firstAlive() int {
	for i, p := range g.players {
		if !p.GetOut() {
			return i
		}
	}
	return 0
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *ThreeCardBrag) GetPhase() ThreeCardBragPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *ThreeCardBrag) SetPhase(phase ThreeCardBragPhase) { g.phase = phase }

// GetRoundNumber 現在のディール番号取得
func (g *ThreeCardBrag) GetRoundNumber() int { return g.roundNumber }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *ThreeCardBrag) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *ThreeCardBrag) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *ThreeCardBrag) GetDealerIdx() int { return g.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (g *ThreeCardBrag) SetDealerIdx(idx int) { g.dealerIdx = idx }

// GetPot 現在のポット取得
func (g *ThreeCardBrag) GetPot() int { return g.pot }

// SetPot ポット設定 (テスト用)
func (g *ThreeCardBrag) SetPot(v int) { g.pot = v }

// GetStake 現在の賭け単位取得
func (g *ThreeCardBrag) GetStake() int { return g.stake }

// SetStake 賭け単位設定 (テスト用)
func (g *ThreeCardBrag) SetStake(v int) { g.stake = v }

// GetRoundWinnerIdx 直近ディールの勝者取得 (-1: 未確定)
func (g *ThreeCardBrag) GetRoundWinnerIdx() int { return g.roundWinnerIdx }

// IsShowdown ショーダウンが行われたか
func (g *ThreeCardBrag) IsShowdown() bool { return g.showdown }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *ThreeCardBrag) GetGameEndFlag() bool { return g.gameEndFlag }

// GetMatchWinnerIdx 試合の勝者取得 (-1: 未確定)
func (g *ThreeCardBrag) GetMatchWinnerIdx() int { return g.matchWinnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *ThreeCardBrag) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *ThreeCardBrag) GetPlayer(i int) *ThreeCardBragPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetConfig 設定取得
func (g *ThreeCardBrag) GetConfig() ThreeCardBragConfig { return g.config }

// SetConfig 設定変更
func (g *ThreeCardBrag) SetConfig(cfg ThreeCardBragConfig) { g.config = cfg }

// IsHumanTurn 現在の手番が人間かどうか
func (g *ThreeCardBrag) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// CanShow 現在の手番プレイヤーが Show を要求できるか (Web/CUI 用)。
func (g *ThreeCardBrag) CanShow() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.canShow(g.currentPlayerIdx)
}

// GetHint 人間プレイヤーへのヒントを取得する
func (g *ThreeCardBrag) GetHint() *ThreeCardBragHint {
	if g.phase != ThreeCardBragPhaseBetting || g.currentPlayerIdx != 0 {
		return nil
	}
	p := g.players[0]
	if !p.GetSeen() {
		return &ThreeCardBragHint{Action: "see", Reason: "see_first"}
	}
	cat, _ := g.handEval(0)
	switch {
	case cat >= ThreeCardBragRun:
		if g.canShow(0) {
			return &ThreeCardBragHint{Action: "show", Reason: "strong_hand"}
		}
		return &ThreeCardBragHint{Action: "raise", Reason: "strong_hand"}
	case cat >= ThreeCardBragPair:
		return &ThreeCardBragHint{Action: "bet", Reason: "medium_hand"}
	default:
		return &ThreeCardBragHint{Action: "fold", Reason: "weak_hand"}
	}
}

// --- JSON ---

type threeCardBragJSON struct {
	TrumpCards       *TrumpCards            `json:"tc"`
	Players          []*ThreeCardBragPlayer `json:"ps"`
	Config           ThreeCardBragConfig    `json:"cf"`
	Phase            ThreeCardBragPhase     `json:"ph"`
	RoundNumber      int                    `json:"rn"`
	DealerIdx        int                    `json:"di"`
	CurrentPlayerIdx int                    `json:"ci"`
	Pot              int                    `json:"pt"`
	Stake            int                    `json:"sk"`
	LastAggressorIdx int                    `json:"la"`
	ActionCount      int                    `json:"ac"`
	RoundWinnerIdx   int                    `json:"rw"`
	Showdown         bool                   `json:"sd"`
	GameEndFlag      bool                   `json:"ge"`
	MatchWinnerIdx   int                    `json:"mw"`
	ActionLog        []*ActionLogEntry      `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *ThreeCardBrag) MarshalJSON() ([]byte, error) {
	return json.Marshal(threeCardBragJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		DealerIdx:        g.dealerIdx,
		CurrentPlayerIdx: g.currentPlayerIdx,
		Pot:              g.pot,
		Stake:            g.stake,
		LastAggressorIdx: g.lastAggressorIdx,
		ActionCount:      g.actionCount,
		RoundWinnerIdx:   g.roundWinnerIdx,
		Showdown:         g.showdown,
		GameEndFlag:      g.gameEndFlag,
		MatchWinnerIdx:   g.matchWinnerIdx,
		ActionLog:        g.actionLog,
	})
}

const threeCardBragMaxSliceLen = 5000

var errThreeCardBragSnapshot = errors.New("threecardbrag: invalid serialised game state")

func threeCardBragIdxInRange(i int) bool { return i >= 0 && i < ThreeCardBragPlayerCnt }

// UnmarshalJSON implements json.Unmarshaler.
func (g *ThreeCardBrag) UnmarshalJSON(data []byte) error {
	var j threeCardBragJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != ThreeCardBragPlayerCnt ||
		len(j.ActionLog) > threeCardBragMaxSliceLen ||
		!threeCardBragIdxInRange(j.CurrentPlayerIdx) || !threeCardBragIdxInRange(j.DealerIdx) ||
		!threeCardBragIdxInRange(j.LastAggressorIdx) ||
		j.RoundWinnerIdx < -1 || j.RoundWinnerIdx >= ThreeCardBragPlayerCnt ||
		j.MatchWinnerIdx < -1 || j.MatchWinnerIdx >= ThreeCardBragPlayerCnt ||
		j.RoundNumber < 1 || j.Pot < 0 || j.Stake < 0 || j.ActionCount < 0 ||
		j.Phase < ThreeCardBragPhaseBetting || j.Phase > ThreeCardBragPhaseGameEnd {
		return errThreeCardBragSnapshot
	}
	for _, p := range j.Players {
		if p == nil {
			return errThreeCardBragSnapshot
		}
	}
	for _, entry := range j.ActionLog {
		if entry == nil {
			return errThreeCardBragSnapshot
		}
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.dealerIdx = j.DealerIdx
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.pot = j.Pot
	g.stake = j.Stake
	g.lastAggressorIdx = j.LastAggressorIdx
	g.actionCount = j.ActionCount
	g.roundWinnerIdx = j.RoundWinnerIdx
	g.showdown = j.Showdown
	g.gameEndFlag = j.GameEndFlag
	g.matchWinnerIdx = j.MatchWinnerIdx
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
