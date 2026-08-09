//go:build !js || !wasm || casino

// TeenPatti (ティーンパッティ) のドメインモデル。
//
// Teen Patti はインド版 Three Card Brag。52 枚デッキを使い、4 人で全員ブート (アンティ) を入れて
// から 3 枚ずつ配る。各プレイヤーは手札を見ない「Blind」か見る「Seen」かを選べ、Blind は Seen の
// 半額 (本実装では Blind=stake / Seen=2*stake) で賭けられ、いつでも手札を見て Seen に昇格できる。
//
// ベッティングは時計回りにコール / レイズ / フォールド。最後の 1 人になればポット獲得。残り 2 人で
// Seen プレイヤーは「Show」(2*stake を払って勝負) を要求できる。さらに Teen Patti 固有の
// 「サイドショー (Side Show)」: 残り 3 人以上のとき Seen プレイヤーは直前の Seen プレイヤーに手比べを
// 申請でき、申請を受けた側が承諾すると弱い方が降りる (拒否すればそのまま続行)。
//
// 役判定は Three Card Brag と共有 (ThreeCardBragEval / ThreeCardBragCompare): Trail (スリーカード) >
// Pure Sequence > Sequence > Color > Pair > High Card。チップが尽きたプレイヤーは脱落し、最後に
// 残った 1 人が試合の勝者となる。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// teenPattiMaxActions は 1 ディールあたりの賭けアクション上限 (安全網)。
// 通常はフォールド / Show で終わるが、全員コールし続ける状況を強制的に決着させる。
const teenPattiMaxActions = 24

// TeenPattiPhase ゲームフェーズ
type TeenPattiPhase int

// Teen Patti のフェーズ定数
const (
	// TeenPattiPhaseBetting ベッティングフェーズ
	TeenPattiPhaseBetting TeenPattiPhase = iota
	// TeenPattiPhaseSideShow サイドショー申請に対する応答待ち
	TeenPattiPhaseSideShow
	// TeenPattiPhaseShowdown ショーダウン (手を公開)
	TeenPattiPhaseShowdown
	// TeenPattiPhaseRoundEnd ディール終了
	TeenPattiPhaseRoundEnd
	// TeenPattiPhaseGameEnd ゲーム終了
	TeenPattiPhaseGameEnd
)

// TeenPattiHint ヒント情報
type TeenPattiHint struct {
	Action string // "see"/"bet"/"raise"/"fold"/"show"
	Reason string // ヒント理由キー
}

// teenPattiSideShow は直近で成立した (承諾された) サイドショーの結果を保持する (表示用)。
// カード自体は保持せず参加者インデックスと敗者のみを持つ (WASM バイナリを軽く保つため)。
// プレゼンター側が保持中のカードを敗者・勝者インデックスから引いて役名と手札を組み立てる。
type teenPattiSideShow struct {
	Requester int `json:"rq"` // 申請者インデックス
	Target    int `json:"tg"` // 対象インデックス
	Loser     int `json:"ls"` // 敗者インデックス (Requester または Target)
}

// TeenPatti ティーンパッティ ゲームクラス
type TeenPatti struct {
	trumpCards        *TrumpCards
	players           []*TeenPattiPlayer
	config            TeenPattiConfig
	phase             TeenPattiPhase
	roundNumber       int
	dealerIdx         int
	currentPlayerIdx  int
	pot               int
	stake             int // 現在の賭け単位 (Blind=stake / Seen=2*stake)
	lastAggressorIdx  int
	actionCount       int
	roundWinnerIdx    int // 直近ディールの勝者 (-1: 未確定)
	showdown          bool
	sideShowRequester int // サイドショー申請者 (-1: なし)
	sideShowTarget    int // サイドショー対象 (-1: なし)
	gameEndFlag       bool
	matchWinnerIdx    int                // 試合の勝者 (-1: 未確定)
	lastSideShow      *teenPattiSideShow // 直近で成立したサイドショー結果 (nil: なし)
	actionLogBase
}

// NewTeenPatti コンストラクタ
func NewTeenPatti(trumpCards *TrumpCards, players []*TeenPattiPlayer, config TeenPattiConfig) *TeenPatti {
	return &TeenPatti{
		trumpCards:        trumpCards,
		players:           players,
		config:            config,
		roundWinnerIdx:    -1,
		matchWinnerIdx:    -1,
		sideShowRequester: -1,
		sideShowTarget:    -1,
	}
}

// NewDefaultTeenPatti 標準の 4 人セットアップ (人間 idx 0 + CPU 3) を返す。
func NewDefaultTeenPatti() *TeenPatti {
	cfg := DefaultTeenPattiConfig()
	players := make([]*TeenPattiPlayer, TeenPattiPlayerCnt)
	for i := range players {
		players[i] = NewTeenPattiPlayer(i == 0, cfg.StartingChips)
	}
	return NewTeenPatti(NewTrumpCards(0), players, cfg)
}

// Reset ゲーム初期化
func (g *TeenPatti) Reset() {
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
func (g *TeenPatti) NextRound() {
	if g.phase != TeenPattiPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = g.nextActive(g.dealerIdx)
	g.startDeal()
}

// startDeal 1 ディールを開始する: アンティ徴収・配札・ベッティング開始。
func (g *TeenPatti) startDeal() {
	g.pot = 0
	g.showdown = false
	g.roundWinnerIdx = -1
	g.actionCount = 0
	g.sideShowRequester = -1
	g.sideShowTarget = -1
	g.lastSideShow = nil
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
		g.phase = TeenPattiPhaseGameEnd
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
		for i := 0; i < TeenPattiHandSize; i++ {
			if c := g.trumpCards.DrawCard(); c != nil {
				p.AddCard(c)
			}
		}
	}
	g.phase = TeenPattiPhaseBetting
	g.currentPlayerIdx = g.nextActive(g.dealerIdx)
	g.lastAggressorIdx = g.currentPlayerIdx
	g.appendLog(-1, "deal", fmt.Sprintf("Deal %d: ante %d, pot %d", g.roundNumber, g.config.Ante, g.pot), nil)
}

// callCost は playerIdx が現在の stake をコールするのに必要な額を返す (Blind は半額)。
func (g *TeenPatti) callCost(playerIdx int) int {
	if g.players[playerIdx].GetSeen() {
		return g.stake * 2
	}
	return g.stake
}

// --- Actions ---

// PlayerSee 人間プレイヤーが手札を見て Seen に昇格する (Blind 時のみ)。手番は消費しない。
func (g *TeenPatti) PlayerSee() error {
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
func (g *TeenPatti) PlayerBet() error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	return g.applyCall(g.currentPlayerIdx)
}

// PlayerRaise 人間プレイヤーが stake を newStake へ引き上げて賭ける。
func (g *TeenPatti) PlayerRaise(newStake int) error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	if newStake <= g.stake {
		return NewDomainError(ErrInvalidPlay, "レイズは現在の賭け単位より大きくする必要があります")
	}
	return g.applyRaise(g.currentPlayerIdx, newStake)
}

// PlayerFold 人間プレイヤーが降りる。
func (g *TeenPatti) PlayerFold() error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	g.applyFold(g.currentPlayerIdx)
	return nil
}

// PlayerShow 残り 2 人のとき、Seen プレイヤーが Show (勝負) を要求する。
func (g *TeenPatti) PlayerShow() error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	if !g.canShow(g.currentPlayerIdx) {
		return NewDomainError(ErrInvalidPlay, "Show は残り 2 人かつ Seen のときのみ要求できます")
	}
	g.applyShow(g.currentPlayerIdx)
	return nil
}

// PlayerRequestSideShow 人間の Seen プレイヤーが直前の Seen プレイヤーにサイドショーを申請する。
// 申請者は Seen のコール額を支払い、対象プレイヤーの応答待ち (SideShow フェーズ) になる。
func (g *TeenPatti) PlayerRequestSideShow() error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	if !g.canRequestSideShow(g.currentPlayerIdx) {
		return NewDomainError(ErrInvalidPlay, "サイドショーは Seen 同士・残り 3 人以上のときのみ申請できます")
	}
	g.applyRequestSideShow(g.currentPlayerIdx)
	return nil
}

// PlayerRespondSideShow サイドショー対象の人間プレイヤーが承諾 (accept=true) / 拒否する。
func (g *TeenPatti) PlayerRespondSideShow(accept bool) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TeenPattiPhaseSideShow {
		return ErrWrongPhase
	}
	if g.currentPlayerIdx != g.sideShowTarget || !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.applyRespondSideShow(accept)
	return nil
}

// applyRequestSideShow サイドショー申請を適用する。
func (g *TeenPatti) applyRequestSideShow(idx int) {
	cost := g.stake * 2 // 申請者は Seen
	p := g.players[idx]
	if p.GetChips() < cost {
		cost = p.GetChips()
	}
	p.SubtractChips(cost)
	p.AddRoundBet(cost)
	g.pot += cost
	g.actionCount++
	target := g.prevActiveSeen(idx)
	g.sideShowRequester = idx
	g.sideShowTarget = target
	g.phase = TeenPattiPhaseSideShow
	g.currentPlayerIdx = target
	g.appendLog(idx, "sideshow_request",
		fmt.Sprintf("%s requests a side show with %s", playerName(g.players, idx), playerName(g.players, target)), nil)
}

// applyRespondSideShow サイドショーへの応答を適用する。
func (g *TeenPatti) applyRespondSideShow(accept bool) {
	req, tgt := g.sideShowRequester, g.sideShowTarget
	g.sideShowRequester = -1
	g.sideShowTarget = -1
	if accept {
		cmp := g.compareHands(req, tgt)
		loser := req
		if cmp > 0 {
			loser = tgt
		}
		// 表示用に成立したサイドショーの結果を保持する (参加者のカードは deal 更新まで保持される)。
		g.lastSideShow = &teenPattiSideShow{Requester: req, Target: tgt, Loser: loser}
		g.appendLog(tgt, "sideshow_accept",
			fmt.Sprintf("%s accepts; %s loses the side show", playerName(g.players, tgt), playerName(g.players, loser)), nil)
		g.players[loser].SetFolded(true)
		if g.activeCount() == 1 {
			g.endDeal([]int{g.firstActive()})
			return
		}
	} else {
		g.appendLog(tgt, "sideshow_decline", fmt.Sprintf("%s declines the side show", playerName(g.players, tgt)), nil)
	}
	// 申請者の賭けは済んでいるので、申請者の次の手番から再開する。
	g.phase = TeenPattiPhaseBetting
	g.currentPlayerIdx = g.nextActive(req)
}

// canRequestSideShow は idx がサイドショーを申請できるか (Seen 同士・3 人以上) を返す。
func (g *TeenPatti) canRequestSideShow(idx int) bool {
	if g.activeCount() < 3 || !g.players[idx].GetSeen() {
		return false
	}
	t := g.prevActiveSeen(idx)
	return t >= 0 && t != idx
}

// prevActiveSeen は idx の手前 (反時計回り) で最初の未フォールド・Seen プレイヤーを返す (-1: なし)。
func (g *TeenPatti) prevActiveSeen(idx int) int {
	for i := 1; i <= TeenPattiPlayerCnt; i++ {
		j := (idx - i + TeenPattiPlayerCnt) % TeenPattiPlayerCnt
		if j == idx {
			break
		}
		p := g.players[j]
		if !p.GetOut() && !p.GetFolded() && p.GetSeen() {
			return j
		}
	}
	return -1
}

// checkTurn は現在の手番が人間かつベッティングフェーズかを検証する。
func (g *TeenPatti) checkTurn() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TeenPattiPhaseBetting {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return nil
}

// applyCall コールを適用する。チップ不足なら自動フォールド。
func (g *TeenPatti) applyCall(idx int) error {
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
func (g *TeenPatti) applyRaise(idx, newStake int) error {
	// **検証してから書き換える。**以前は g.stake を先に更新しており、チップ不足で
	// 弾いた後も上がったままだった。以降の全員がその額でコールを迫られ、
	// 賭け単位が 1 から 290 に化けるような壊れ方をする。Web 側に上限が無く
	// (#4729)、この経路は誰でも踏めた。
	p := g.players[idx]
	cost := newStake
	if p.GetSeen() {
		cost = newStake * 2
	}
	if p.GetChips() < cost {
		return NewDomainError(ErrInvalidPlay, "チップが不足しています")
	}
	g.stake = newStake
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
func (g *TeenPatti) applyFold(idx int) {
	g.players[idx].SetFolded(true)
	g.appendLog(idx, "fold", fmt.Sprintf("%s folds", playerName(g.players, idx)), nil)
	if g.activeCount() == 1 {
		g.endDeal([]int{g.firstActive()})
		return
	}
	g.advanceOrResolve(idx)
}

// applyShow Show を適用し、残り 2 人のショーダウンを行う。
func (g *TeenPatti) applyShow(idx int) {
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
	g.phase = TeenPattiPhaseShowdown
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
func (g *TeenPatti) advanceOrResolve(fromIdx int) {
	if g.actionCount >= teenPattiMaxActions {
		g.forcedShowdown()
		return
	}
	g.currentPlayerIdx = g.nextActive(fromIdx)
}

// forcedShowdown 残り全アクティブを公開し、最強手にポットを与える (安全網)。
func (g *TeenPatti) forcedShowdown() {
	g.showdown = true
	g.phase = TeenPattiPhaseShowdown
	winners := g.bestActiveHands()
	g.endDeal(winners)
}

// --- CPU ---

// CpuAct 現在の手番が CPU の場合に 1 アクション実行する (ベッティング / サイドショー応答)。
func (g *TeenPatti) CpuAct() {
	if g.gameEndFlag {
		return
	}
	// サイドショー応答待ち: 対象が CPU なら手の強さで承諾/拒否する。
	if g.phase == TeenPattiPhaseSideShow {
		if g.currentPlayerIdx == g.sideShowTarget && !g.players[g.sideShowTarget].GetIsHuman() {
			cat, _ := g.handEval(g.sideShowTarget)
			g.applyRespondSideShow(cat >= ThreeCardBragPair)
		}
		return
	}
	if g.phase != TeenPattiPhaseBetting {
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
func (g *TeenPatti) endDeal(winners []int) {
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
		g.phase = TeenPattiPhaseGameEnd
		g.matchWinnerIdx = g.firstAlive()
		g.appendLog(g.matchWinnerIdx, "game_end", fmt.Sprintf("%s wins the match", playerName(g.players, g.matchWinnerIdx)), nil)
		return
	}
	g.phase = TeenPattiPhaseRoundEnd
}

// bestActiveHands は未フォールドプレイヤーのうち最強手のインデックス列を返す (引き分けは複数)。
func (g *TeenPatti) bestActiveHands() []int {
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
func (g *TeenPatti) handEval(playerIdx int) (int, []int) {
	p := g.players[playerIdx]
	cards := make([]*Card, 0, TeenPattiHandSize)
	for i := 0; i < p.GetCardsSize(); i++ {
		cards = append(cards, p.GetCard(i))
	}
	return ThreeCardBragEval(cards)
}

// compareHands は手 a と手 b を比較する (a 勝ち=1, b 勝ち=-1, 引き分け=0)。
func (g *TeenPatti) compareHands(a, b int) int {
	catA, tbA := g.handEval(a)
	catB, tbB := g.handEval(b)
	return ThreeCardBragCompare(catA, tbA, catB, tbB)
}

// --- helpers ---

// canShow は idx が Show を要求できるか (残り 2 人 & Seen) を返す。
func (g *TeenPatti) canShow(idx int) bool {
	return g.activeCount() == 2 && g.players[idx].GetSeen()
}

// activeCount 未フォールド・非脱落のプレイヤー数。
func (g *TeenPatti) activeCount() int {
	n := 0
	for _, p := range g.players {
		if !p.GetOut() && !p.GetFolded() {
			n++
		}
	}
	return n
}

// firstActive 最初の未フォールド・非脱落プレイヤー。
func (g *TeenPatti) firstActive() int {
	for i, p := range g.players {
		if !p.GetOut() && !p.GetFolded() {
			return i
		}
	}
	return 0
}

// otherActive idx 以外の未フォールド・非脱落プレイヤー (残り 2 人の相手)。
func (g *TeenPatti) otherActive(idx int) int {
	for i, p := range g.players {
		if i != idx && !p.GetOut() && !p.GetFolded() {
			return i
		}
	}
	return idx
}

// nextActive from の次の未フォールド・非脱落プレイヤー。
func (g *TeenPatti) nextActive(from int) int {
	for i := 1; i <= TeenPattiPlayerCnt; i++ {
		idx := (from + i) % TeenPattiPlayerCnt
		if !g.players[idx].GetOut() && !g.players[idx].GetFolded() {
			return idx
		}
	}
	return from
}

// aliveCount チップを持つ (非脱落) プレイヤー数。
func (g *TeenPatti) aliveCount() int {
	n := 0
	for _, p := range g.players {
		if !p.GetOut() {
			n++
		}
	}
	return n
}

// firstAlive 最初の非脱落プレイヤー。
func (g *TeenPatti) firstAlive() int {
	for i, p := range g.players {
		if !p.GetOut() {
			return i
		}
	}
	return 0
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *TeenPatti) GetPhase() TeenPattiPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *TeenPatti) SetPhase(phase TeenPattiPhase) { g.phase = phase }

// GetRoundNumber 現在のディール番号取得
func (g *TeenPatti) GetRoundNumber() int { return g.roundNumber }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *TeenPatti) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *TeenPatti) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *TeenPatti) GetDealerIdx() int { return g.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (g *TeenPatti) SetDealerIdx(idx int) { g.dealerIdx = idx }

// GetPot 現在のポット取得
func (g *TeenPatti) GetPot() int { return g.pot }

// SetPot ポット設定 (テスト用)
func (g *TeenPatti) SetPot(v int) { g.pot = v }

// GetRaiseRange は指定プレイヤーがいまレイズできる額の範囲を返す。
// レイズできないときは ok=false。
//
// **判定は callCost と同じ式から出す。**Seen は倍払うので上限が半分になる。
// この式が CUI の表示と Web の入力上限で割れると、「入力できたのに弾かれる」
// ずれになる (#4729)。
func (g *TeenPatti) GetRaiseRange(playerIdx int) (minRaise, maxRaise int, ok bool) {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return 0, 0, false
	}
	p := g.players[playerIdx]
	minRaise = g.stake + 1
	maxRaise = p.GetChips()
	if p.GetSeen() {
		maxRaise = p.GetChips() / 2
	}
	if maxRaise < minRaise {
		return minRaise, maxRaise, false
	}
	return minRaise, maxRaise, true
}

// GetStake 現在の賭け単位取得
func (g *TeenPatti) GetStake() int { return g.stake }

// SetStake 賭け単位設定 (テスト用)
func (g *TeenPatti) SetStake(v int) { g.stake = v }

// GetRoundWinnerIdx 直近ディールの勝者取得 (-1: 未確定)
func (g *TeenPatti) GetRoundWinnerIdx() int { return g.roundWinnerIdx }

// IsShowdown ショーダウンが行われたか
func (g *TeenPatti) IsShowdown() bool { return g.showdown }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *TeenPatti) GetGameEndFlag() bool { return g.gameEndFlag }

// GetMatchWinnerIdx 試合の勝者取得 (-1: 未確定)
func (g *TeenPatti) GetMatchWinnerIdx() int { return g.matchWinnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *TeenPatti) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *TeenPatti) GetPlayer(i int) *TeenPattiPlayer {
	return getPlayer(g.players, i)
}

// GetConfig 設定取得
func (g *TeenPatti) GetConfig() TeenPattiConfig { return g.config }

// SetConfig 設定変更
func (g *TeenPatti) SetConfig(cfg TeenPattiConfig) { g.config = cfg }

// IsHumanTurn 現在の手番が人間かどうか
func (g *TeenPatti) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// CanShow 現在の手番プレイヤーが Show を要求できるか (Web/CUI 用)。
func (g *TeenPatti) CanShow() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.canShow(g.currentPlayerIdx)
}

// CanRequestSideShow 現在の手番プレイヤーがサイドショーを申請できるか (Web/CUI 用)。
func (g *TeenPatti) CanRequestSideShow() bool {
	if g.phase != TeenPattiPhaseBetting || g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.canRequestSideShow(g.currentPlayerIdx)
}

// GetSideShowRequester サイドショー申請者インデックス取得 (-1: なし)
func (g *TeenPatti) GetSideShowRequester() int { return g.sideShowRequester }

// GetSideShowTarget サイドショー対象インデックス取得 (-1: なし)
func (g *TeenPatti) GetSideShowTarget() int { return g.sideShowTarget }

// GetLastSideShow は直近で成立 (承諾) したサイドショーの申請者・対象・敗者を返す。
// 成立したサイドショーがまだ無い (またはディール更新で消去済み) のとき ok=false。
func (g *TeenPatti) GetLastSideShow() (requester, target, loser int, ok bool) {
	if g.lastSideShow == nil {
		return -1, -1, -1, false
	}
	return g.lastSideShow.Requester, g.lastSideShow.Target, g.lastSideShow.Loser, true
}

// GetHint 人間プレイヤーへのヒントを取得する
func (g *TeenPatti) GetHint() *TeenPattiHint {
	if g.phase != TeenPattiPhaseBetting || g.currentPlayerIdx != 0 {
		return nil
	}
	p := g.players[0]
	if !p.GetSeen() {
		return &TeenPattiHint{Action: "see", Reason: "see_first"}
	}
	cat, _ := g.handEval(0)
	switch {
	case cat >= ThreeCardBragRun:
		if g.canShow(0) {
			return &TeenPattiHint{Action: "show", Reason: "strong_hand"}
		}
		return &TeenPattiHint{Action: "raise", Reason: "strong_hand"}
	case cat >= ThreeCardBragPair:
		return &TeenPattiHint{Action: "bet", Reason: "medium_hand"}
	default:
		return &TeenPattiHint{Action: "fold", Reason: "weak_hand"}
	}
}

// --- JSON ---

type teenPattiJSON struct {
	TrumpCards        *TrumpCards        `json:"tc"`
	Players           []*TeenPattiPlayer `json:"ps"`
	Config            TeenPattiConfig    `json:"cf"`
	Phase             TeenPattiPhase     `json:"ph"`
	RoundNumber       int                `json:"rn"`
	DealerIdx         int                `json:"di"`
	CurrentPlayerIdx  int                `json:"ci"`
	Pot               int                `json:"pt"`
	Stake             int                `json:"sk"`
	LastAggressorIdx  int                `json:"la"`
	ActionCount       int                `json:"ac"`
	RoundWinnerIdx    int                `json:"rw"`
	Showdown          bool               `json:"sd"`
	SideShowRequester int                `json:"sr"`
	SideShowTarget    int                `json:"st"`
	GameEndFlag       bool               `json:"ge"`
	MatchWinnerIdx    int                `json:"mw"`
	LastSideShow      *teenPattiSideShow `json:"ls,omitempty"`
	ActionLog         []*ActionLogEntry  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *TeenPatti) MarshalJSON() ([]byte, error) {
	return json.Marshal(teenPattiJSON{
		TrumpCards:        g.trumpCards,
		Players:           g.players,
		Config:            g.config,
		Phase:             g.phase,
		RoundNumber:       g.roundNumber,
		DealerIdx:         g.dealerIdx,
		CurrentPlayerIdx:  g.currentPlayerIdx,
		Pot:               g.pot,
		Stake:             g.stake,
		LastAggressorIdx:  g.lastAggressorIdx,
		ActionCount:       g.actionCount,
		RoundWinnerIdx:    g.roundWinnerIdx,
		Showdown:          g.showdown,
		SideShowRequester: g.sideShowRequester,
		SideShowTarget:    g.sideShowTarget,
		GameEndFlag:       g.gameEndFlag,
		MatchWinnerIdx:    g.matchWinnerIdx,
		LastSideShow:      g.lastSideShow,
		ActionLog:         g.actionLog,
	})
}

const teenPattiMaxSliceLen = 5000

var errTeenPattiSnapshot = errors.New("teenpatti: invalid serialised game state")

func teenPattiIdxInRange(i int) bool { return i >= 0 && i < TeenPattiPlayerCnt }

// UnmarshalJSON implements json.Unmarshaler.
func (g *TeenPatti) UnmarshalJSON(data []byte) error {
	var j teenPattiJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != TeenPattiPlayerCnt ||
		len(j.ActionLog) > teenPattiMaxSliceLen ||
		!teenPattiIdxInRange(j.CurrentPlayerIdx) || !teenPattiIdxInRange(j.DealerIdx) ||
		!teenPattiIdxInRange(j.LastAggressorIdx) ||
		j.RoundWinnerIdx < -1 || j.RoundWinnerIdx >= TeenPattiPlayerCnt ||
		j.MatchWinnerIdx < -1 || j.MatchWinnerIdx >= TeenPattiPlayerCnt ||
		j.RoundNumber < 1 || j.Pot < 0 || j.Stake < 0 || j.ActionCount < 0 ||
		j.SideShowRequester < -1 || j.SideShowRequester >= TeenPattiPlayerCnt ||
		j.SideShowTarget < -1 || j.SideShowTarget >= TeenPattiPlayerCnt ||
		j.Phase < TeenPattiPhaseBetting || j.Phase > TeenPattiPhaseGameEnd {
		return errTeenPattiSnapshot
	}
	if j.LastSideShow != nil {
		ss := j.LastSideShow
		if !teenPattiIdxInRange(ss.Requester) || !teenPattiIdxInRange(ss.Target) ||
			!teenPattiIdxInRange(ss.Loser) || (ss.Loser != ss.Requester && ss.Loser != ss.Target) {
			return errTeenPattiSnapshot
		}
	}
	for _, p := range j.Players {
		if p == nil {
			return errTeenPattiSnapshot
		}
	}
	for _, entry := range j.ActionLog {
		if entry == nil {
			return errTeenPattiSnapshot
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
	g.sideShowRequester = j.SideShowRequester
	g.sideShowTarget = j.SideShowTarget
	g.gameEndFlag = j.GameEndFlag
	g.matchWinnerIdx = j.MatchWinnerIdx
	g.lastSideShow = j.LastSideShow
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
