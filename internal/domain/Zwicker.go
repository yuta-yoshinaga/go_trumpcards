//go:build !js || !wasm || extra2

// Package domain — ツヴィッカー (Zwicker) のドメインモデル。
//
// ドイツ北部シュレースヴィヒのカシノ系フィッシングゲーム。**55 枚**
// (52 + ジョーカー 3 枚)、4 人 2 対 2。
//
// # issue #4387 の仕様案との相違
//
// 看板メカニクスとされた 2 つが、どちらも原典 (pagat) に存在しない。
//
//   - issue は「ジョーカーを出すとき**任意の値を宣言**してワイルドにする」と
//     するが、**ジョーカーはワイルドではない**。3 枚に 15 / 20 / 25 が固定で
//     割り当てられ、選択の余地はない。値を選べるのは **A と絵札**のほうで、
//     A=1/11, J=2/12, Q=3/13, K=4/14 の 2 択を取る側が決める
//   - issue は「値が一致する組をまとめて取ることを **Zwick** と呼ぶ」とするが、
//     **Zwick は「場を空にした」ボーナス**の名前 (1 点)。複数組の同時取りは
//     カシノ系の普通のキャプチャで、Zwicker 固有ではない
//   - デッキはジョーカー 2 枚ではなく **3 枚**。配り方 (場 3 + 各自 4+4+5) が
//     山をちょうど使い切るのは 55 枚のときだけで、54 枚だと 51 が 4 で
//     割り切れない
//   - 2〜5 人ではなく **4 人 2 対 2** が本来の形
//   - 得点は「獲得枚数と特定の得点札」ではなく内訳が決まっており、合計
//     **ちょうど 30 点** + Zwick 各 1 点
//
// **cassino のエンジンは土台にできない。**cassino は絵札をランク一致でしか
// 扱わず合計に参加させないので (CassinoIsFaceCard / findFaceRankMatches)、
// 値関数が根本から違う。共有エンジンを差し替え可能にするほうが専用に書くより
// 大きな変更になるため、Zwicker 側に専用の値・捕獲判定を持つ。
package domain

import (
	"encoding/json"
	"fmt"
)

// ZwickerPlayerCnt はプレイヤー数 (4 人 2 対 2)。
const ZwickerPlayerCnt = 4

// ZwickerJokerCnt はジョーカーの枚数。55 枚デッキの根拠。
const ZwickerJokerCnt = 3

// ZwickerInitialTableSize は最初に場へ表向きに置く枚数。
const ZwickerInitialTableSize = 3

// zwickerDealSizes は 3 段階の配り方。場 3 枚を除いた 52 枚を 4 人で
// 13 枚ずつ、4 + 4 + 5 に割る。**合計が山とぴったり合う唯一の組み合わせ**。
var zwickerDealSizes = [3]int{4, 4, 5}

// Zwicker の得点定数。合計はちょうど 30 点になる。
const (
	// ZwickerScoreLargeJoker 大ジョーカー (25)
	ZwickerScoreLargeJoker = 7
	// ZwickerScoreMiddleJoker 中ジョーカー (20)
	ZwickerScoreMiddleJoker = 6
	// ZwickerScoreSmallJoker 小ジョーカー (15)
	ZwickerScoreSmallJoker = 5
	// ZwickerScoreDiamondTen ♦10
	ZwickerScoreDiamondTen = 3
	// ZwickerScoreSpadeTen ♠10
	ZwickerScoreSpadeTen = 1
	// ZwickerScoreSpadeTwo ♠2
	ZwickerScoreSpadeTwo = 1
	// ZwickerScoreAce 各エース
	ZwickerScoreAce = 1
	// ZwickerScoreMajority 枚数最多
	ZwickerScoreMajority = 3
	// ZwickerScoreZwick Zwick 1 回につき
	ZwickerScoreZwick = 1
	// ZwickerScoreTotal 札で取れる点の合計 (Zwick を除く)
	ZwickerScoreTotal = ZwickerScoreLargeJoker + ZwickerScoreMiddleJoker + ZwickerScoreSmallJoker +
		ZwickerScoreDiamondTen + ZwickerScoreSpadeTen + ZwickerScoreSpadeTwo +
		4*ZwickerScoreAce + ZwickerScoreMajority
)

// ZwickerPhase はゲームフェーズ。
type ZwickerPhase int

// Zwicker のフェーズ定数
const (
	// ZwickerPhasePlay 手番進行中
	ZwickerPhasePlay ZwickerPhase = iota
	// ZwickerPhaseRoundEnd 1 ディール終了 (精算済み)
	ZwickerPhaseRoundEnd
	// ZwickerPhaseGameEnd 決着
	ZwickerPhaseGameEnd
)

// ZwickerTeamOf は席番号からチーム (0 or 1) を返す。向かい合わせが味方。
func ZwickerTeamOf(seat int) int { return seat % 2 }

// ZwickerBuild は場に積まれた宣言値つきの山。
type ZwickerBuild struct {
	Owner int
	Value int
	Cards []*Card
}

// ZwickerRoundScore は 1 ディールの内訳。
type ZwickerRoundScore struct {
	// CardPoints はチーム別の得点札の合計 (枚数最多と Zwick を除く)。
	CardPoints [2]int
	// Cards はチーム別の獲得枚数。
	Cards [2]int
	// MajorityTeam は枚数最多のチーム (-1 = 同数で誰も取らない)。
	MajorityTeam int
	// Zwicks はチーム別の Zwick 回数。
	Zwicks [2]int
	// Total はチーム別のこのディールの得点。
	Total [2]int
}

// Zwicker はツヴィッカーのゲームクラス。
type Zwicker struct {
	trumpCards *TrumpCards
	players    []*ZwickerPlayer
	config     ZwickerConfig
	phase      ZwickerPhase

	tableCards []*Card
	builds     []*ZwickerBuild

	currentIdx int
	dealerIdx  int
	// dealStage は 3 段階の配り方のどこか (0,1,2)。
	dealStage int
	// lastCaptureIdx は最後に捕獲した席。ディール終了時の残り札の行き先。
	lastCaptureIdx int

	scores    [2]int
	lastRound *ZwickerRoundScore

	gameEndFlag bool
	winnerTeam  int
	actionLog   []*ActionLogEntry
}

// NewZwicker はコンストラクタ。
func NewZwicker(trumpCards *TrumpCards, players []*ZwickerPlayer, config ZwickerConfig) *Zwicker {
	return &Zwicker{
		trumpCards:     trumpCards,
		players:        players,
		config:         config,
		lastCaptureIdx: -1,
		winnerTeam:     -1,
	}
}

// NewDefaultZwicker は標準の 4 人セットアップを返す。
func NewDefaultZwicker() *Zwicker {
	players := make([]*ZwickerPlayer, 0, ZwickerPlayerCnt)
	players = append(players, NewZwickerPlayer(true))
	for range ZwickerPlayerCnt - 1 {
		players = append(players, NewZwickerPlayer(false))
	}
	return NewZwicker(NewTrumpCards(ZwickerJokerCnt), players, DefaultZwickerConfig())
}

// Reset はゲーム全体を初期化する。
func (z *Zwicker) Reset() {
	z.scores = [2]int{}
	z.dealerIdx = 0
	z.gameEndFlag = false
	z.winnerTeam = -1
	z.actionLog = nil
	z.lastRound = nil
	z.dealRound()
}

// dealRound は 1 ディールを開始する。
func (z *Zwicker) dealRound() {
	z.trumpCards.Replenish()
	z.trumpCards.Shuffle()
	z.tableCards = nil
	z.builds = nil
	z.dealStage = 0
	z.lastCaptureIdx = -1
	for _, p := range z.players {
		p.ResetRound()
	}

	for range ZwickerInitialTableSize {
		if c := z.trumpCards.DrawCard(); c != nil {
			z.tableCards = append(z.tableCards, c)
		}
	}
	z.dealHands()

	z.currentIdx = (z.dealerIdx + 1) % len(z.players)
	z.phase = ZwickerPhasePlay
	z.addLog(-1, "deal", fmt.Sprintf("%d cards to the table", len(z.tableCards)), z.tableCards)
}

// dealHands は現在の段階ぶんを各自に配り、段階を進める。
func (z *Zwicker) dealHands() {
	if z.dealStage >= len(zwickerDealSizes) {
		return
	}
	n := zwickerDealSizes[z.dealStage]
	for range n {
		for _, p := range z.players {
			if c := z.trumpCards.DrawCard(); c != nil {
				p.AddCard(c)
			}
		}
	}
	z.dealStage++
}

// Take は手札 1 枚を playedValue として使い、場の札とビルドを取る。
//
// playedValue は A・絵札の 2 択のうちどちらで扱うか。数札とジョーカーでは
// 唯一の値と一致していなければならない。
func (z *Zwicker) Take(player, handIdx, playedValue int, tableIdxs, buildIdxs []int) error {
	card, err := z.checkPlay(player, handIdx)
	if err != nil {
		return err
	}
	if !ZwickerHasValue(card, playedValue) {
		return fmt.Errorf("that card cannot be played as %d", playedValue)
	}
	tIdxs, ok := zwickerSortedUnique(tableIdxs, len(z.tableCards))
	if !ok {
		return fmt.Errorf("bad table selection")
	}
	bIdxs, ok := zwickerSortedUnique(buildIdxs, len(z.builds))
	if !ok {
		return fmt.Errorf("bad build selection")
	}
	if len(tIdxs) == 0 && len(bIdxs) == 0 {
		return fmt.Errorf("a capture must take something")
	}

	// ビルドは宣言値ちょうどでしか取れない。
	for _, i := range bIdxs {
		if z.builds[i].Value != playedValue {
			return fmt.Errorf("build %d is worth %d, not %d", i, z.builds[i].Value, playedValue)
		}
	}
	// 場の単札は、選んだぶんを「合計が playedValue のグループ」に余さず
	// 分けられなければならない。
	if len(tIdxs) > 0 {
		chosen := make([]*Card, 0, len(tIdxs))
		for _, i := range tIdxs {
			chosen = append(chosen, z.tableCards[i])
		}
		if !zwickerCanPartition(chosen, playedValue) {
			return fmt.Errorf("those table cards do not add up to %d", playedValue)
		}
	}

	captured := []*Card{card}
	for _, i := range tIdxs {
		captured = append(captured, z.tableCards[i])
	}
	for _, i := range bIdxs {
		captured = append(captured, z.builds[i].Cards...)
	}

	p := z.GetPlayer(player)
	p.RemoveCard(handIdx)
	z.removeTableCards(tIdxs)
	z.removeBuilds(bIdxs)
	p.AddCaptured(captured)
	z.lastCaptureIdx = player

	detail := fmt.Sprintf("captures %d card(s) as %d", len(captured)-1, playedValue)
	// **Zwick は「場を空にした」ボーナス。**issue が言うような複数組同時取りの
	// 名前ではない。
	if len(z.tableCards) == 0 && len(z.builds) == 0 {
		p.AddZwick()
		detail += " and clears the table (Zwick)"
	}
	z.addLog(player, "take", detail, captured)
	z.advance()
	return nil
}

// Build は手札 1 枚と場の札を積んで宣言値のビルドを**新しく**作る。
//
// 既存ビルドへの合流は無い。同じ宣言値のビルドが 2 つ並ぶことはあるが、Take は
// 一致する値のビルドを複数まとめて取れるので困らない。宣言値と同じ値で取れる札
// を手札に残していなければならない。
func (z *Zwicker) Build(player, handIdx int, tableIdxs []int, declaredValue int) error {
	card, err := z.checkPlay(player, handIdx)
	if err != nil {
		return err
	}
	tIdxs, ok := zwickerSortedUnique(tableIdxs, len(z.tableCards))
	if !ok {
		return fmt.Errorf("bad table selection")
	}
	if len(tIdxs) == 0 {
		return fmt.Errorf("a build needs at least one table card")
	}
	if declaredValue <= 0 {
		return fmt.Errorf("a build needs a positive value")
	}

	group := make([]*Card, 0, len(tIdxs)+1)
	group = append(group, card)
	for _, i := range tIdxs {
		group = append(group, z.tableCards[i])
	}
	// ビルドは 1 つの山なので、合計がちょうど宣言値になる 1 グループに限る。
	if !zwickerSingleGroupSums(group, declaredValue) {
		return fmt.Errorf("those cards do not add up to %d", declaredValue)
	}
	// **宣言値で取れる札を手札に残しているかは必須の条件。**取れない値を
	// 宣言すると、そのビルドは相手に献上するだけになる。
	if !z.holdsValueBesides(player, handIdx, declaredValue) {
		return fmt.Errorf("you must keep a card worth %d to collect the build", declaredValue)
	}

	p := z.GetPlayer(player)
	p.RemoveCard(handIdx)
	z.removeTableCards(tIdxs)
	z.builds = append(z.builds, &ZwickerBuild{Owner: player, Value: declaredValue, Cards: group})
	z.addLog(player, "build", fmt.Sprintf("builds %d", declaredValue), group)
	z.advance()
	return nil
}

// Trail は手札 1 枚を場に置くだけで手番を終える。
func (z *Zwicker) Trail(player, handIdx int) error {
	card, err := z.checkPlay(player, handIdx)
	if err != nil {
		return err
	}
	z.GetPlayer(player).RemoveCard(handIdx)
	z.tableCards = append(z.tableCards, card)
	z.addLog(player, "trail", "trails a card", []*Card{card})
	z.advance()
	return nil
}

// checkPlay は出せる状態かを確かめ、対象の札を返す。
func (z *Zwicker) checkPlay(player, handIdx int) (*Card, error) {
	if z.gameEndFlag {
		return nil, fmt.Errorf("the game is over")
	}
	if z.phase != ZwickerPhasePlay {
		return nil, fmt.Errorf("the deal is not in progress")
	}
	if player != z.currentIdx {
		return nil, fmt.Errorf("it is not player %d's turn", player)
	}
	p := z.GetPlayer(player)
	if p == nil || handIdx < 0 || handIdx >= p.GetCardsSize() {
		return nil, fmt.Errorf("card index %d out of range", handIdx)
	}
	return p.GetCard(handIdx), nil
}

// holdsValueBesides は handIdx 以外の手札に value で使える札があるかを返す。
func (z *Zwicker) holdsValueBesides(player, handIdx, value int) bool {
	p := z.GetPlayer(player)
	for i := range p.GetCardsSize() {
		if i == handIdx {
			continue
		}
		if ZwickerHasValue(p.GetCard(i), value) {
			return true
		}
	}
	return false
}

// zwickerSingleGroupSums は cards の合計をちょうど target にできるかを返す。
// 各札は 2 つの値を持ちうるので、値の選び方まで探索する。
func zwickerSingleGroupSums(cards []*Card, target int) bool {
	sums := map[int]struct{}{0: {}}
	for _, c := range cards {
		next := make(map[int]struct{}, len(sums)*2)
		for s := range sums {
			for _, v := range ZwickerCardValues(c) {
				if s+v <= target {
					next[s+v] = struct{}{}
				}
			}
		}
		sums = next
		if len(sums) == 0 {
			return false
		}
	}
	_, ok := sums[target]
	return ok
}

// removeTableCards は昇順の添字集合を場から取り除く。
func (z *Zwicker) removeTableCards(idxs []int) {
	if len(idxs) == 0 {
		return
	}
	drop := make(map[int]struct{}, len(idxs))
	for _, i := range idxs {
		drop[i] = struct{}{}
	}
	kept := make([]*Card, 0, len(z.tableCards))
	for i, c := range z.tableCards {
		if _, gone := drop[i]; !gone {
			kept = append(kept, c)
		}
	}
	z.tableCards = kept
}

// removeBuilds は昇順の添字集合をビルドから取り除く。
func (z *Zwicker) removeBuilds(idxs []int) {
	if len(idxs) == 0 {
		return
	}
	drop := make(map[int]struct{}, len(idxs))
	for _, i := range idxs {
		drop[i] = struct{}{}
	}
	kept := make([]*ZwickerBuild, 0, len(z.builds))
	for i, b := range z.builds {
		if _, gone := drop[i]; !gone {
			kept = append(kept, b)
		}
	}
	z.builds = kept
}

// advance は手番を進め、必要なら配り直しかディール終了に入る。
func (z *Zwicker) advance() {
	z.currentIdx = (z.currentIdx + 1) % len(z.players)
	if !z.allHandsEmpty() {
		return
	}
	if z.dealStage < len(zwickerDealSizes) {
		z.dealHands()
		z.addLog(-1, "deal", fmt.Sprintf("deals stage %d", z.dealStage), nil)
		return
	}
	z.finishRound()
}

// allHandsEmpty は全員の手札が空かを返す。
func (z *Zwicker) allHandsEmpty() bool {
	for _, p := range z.players {
		if p.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// finishRound はディールを精算する。
//
// 残った場札は**最後に捕獲した席**が取る。原典はここを明記していないので、
// カシノ系の慣習に合わせた。枚数最多の 3 点に効くので、伏せずに書いておく。
func (z *Zwicker) finishRound() {
	if last := z.lastCaptureIdx; last >= 0 && (len(z.tableCards) > 0 || len(z.builds) > 0) {
		leftovers := append([]*Card(nil), z.tableCards...)
		for _, b := range z.builds {
			leftovers = append(leftovers, b.Cards...)
		}
		z.GetPlayer(last).AddCaptured(leftovers)
		z.addLog(last, "sweep", "takes what is left on the table", leftovers)
		z.tableCards = nil
		z.builds = nil
	}

	score := z.scoreRound()
	z.lastRound = score
	z.scores[0] += score.Total[0]
	z.scores[1] += score.Total[1]
	z.addLog(-1, "round_end", fmt.Sprintf("team scores %d - %d", score.Total[0], score.Total[1]), nil)

	z.phase = ZwickerPhaseRoundEnd
	z.checkGameEnd()
}

// scoreRound は 1 ディールの内訳を計算する。
func (z *Zwicker) scoreRound() *ZwickerRoundScore {
	out := &ZwickerRoundScore{MajorityTeam: -1}
	for i, p := range z.players {
		team := ZwickerTeamOf(i)
		out.Zwicks[team] += p.GetZwicks()
		for _, c := range p.GetCaptured() {
			out.Cards[team]++
			out.CardPoints[team] += ZwickerScoreOfCard(c)
		}
	}
	switch {
	case out.Cards[0] > out.Cards[1]:
		out.MajorityTeam = 0
	case out.Cards[1] > out.Cards[0]:
		out.MajorityTeam = 1
	}
	for t := range 2 {
		out.Total[t] = out.CardPoints[t] + out.Zwicks[t]*ZwickerScoreZwick
	}
	if out.MajorityTeam >= 0 {
		out.Total[out.MajorityTeam] += ZwickerScoreMajority
	}
	return out
}

// checkGameEnd は目標点に達したチームがあれば決着させる。
func (z *Zwicker) checkGameEnd() {
	target := z.config.TargetScore
	a, b := z.scores[0] >= target, z.scores[1] >= target
	if !a && !b {
		return
	}
	switch {
	case a && !b:
		z.winnerTeam = 0
	case b && !a:
		z.winnerTeam = 1
	case z.scores[0] > z.scores[1]:
		z.winnerTeam = 0
	case z.scores[1] > z.scores[0]:
		z.winnerTeam = 1
	default:
		// 同点で同時到達。次のディールで決める。
		return
	}
	z.gameEndFlag = true
	z.phase = ZwickerPhaseGameEnd
	z.addLog(-1, "game_end", fmt.Sprintf("team %d wins", z.winnerTeam), nil)
}

// NextRound は次のディールを配る。
func (z *Zwicker) NextRound() error {
	if z.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if z.phase != ZwickerPhaseRoundEnd {
		return fmt.Errorf("the deal is still in progress")
	}
	z.dealerIdx = (z.dealerIdx + 1) % len(z.players)
	z.dealRound()
	return nil
}

// ---- CPU ----

// ZwickerCpuAction は CPU が選んだ手。
type ZwickerCpuAction struct {
	// Type は "take" / "trail"。ビルドは v1 では組まない。
	Type      string
	HandIdx   int
	Value     int
	TableIdxs []int
}

// ZwickerCpuDecide は idx の CPU が取る手を決める。
//
// 取れるなら最も得点の高い取り方を選び、取れなければ最も安い札を捨てる。
// ビルドは組まない -- 宣言値を保持する札を残す条件を満たしつつ相手に取られない
// 値を選ぶのは、この難易度では割に合わない。
func (z *Zwicker) ZwickerCpuDecide(idx int) ZwickerCpuAction {
	p := z.GetPlayer(idx)
	if p == nil || p.GetCardsSize() == 0 {
		return ZwickerCpuAction{Type: "trail", HandIdx: -1}
	}
	best := ZwickerCpuAction{Type: "trail", HandIdx: 0}
	bestScore := -1
	for h := range p.GetCardsSize() {
		for _, v := range ZwickerCardValues(p.GetCard(h)) {
			idxs := z.bestCaptureFor(v)
			if idxs == nil {
				continue
			}
			s := 0
			for _, i := range idxs {
				s += ZwickerScoreOfCard(z.tableCards[i]) * 10
				s++
			}
			if s > bestScore {
				bestScore = s
				best = ZwickerCpuAction{Type: "take", HandIdx: h, Value: v, TableIdxs: idxs}
			}
		}
	}
	if bestScore >= 0 {
		return best
	}
	return ZwickerCpuAction{Type: "trail", HandIdx: z.cheapestCard(idx)}
}

// bestCaptureFor は value ちょうどで取れる場札の添字集合を 1 つ返す。
// 大きく取れるほうを優先する。
func (z *Zwicker) bestCaptureFor(value int) []int {
	var best []int
	n := len(z.tableCards)
	if n == 0 || n > 12 {
		// 場が大きすぎるときは全探索しない。実戦では 12 枚を超えない。
		n = min(n, 12)
	}
	for mask := 1; mask < 1<<n; mask++ {
		idxs := make([]int, 0, n)
		cards := make([]*Card, 0, n)
		for i := range n {
			if mask&(1<<i) != 0 {
				idxs = append(idxs, i)
				cards = append(cards, z.tableCards[i])
			}
		}
		if !zwickerCanPartition(cards, value) {
			continue
		}
		if len(idxs) > len(best) {
			best = idxs
		}
	}
	return best
}

// cheapestCard は捨てても痛くない手札の添字を返す。
func (z *Zwicker) cheapestCard(idx int) int {
	p := z.GetPlayer(idx)
	best, bestScore := 0, 1<<30
	for i := range p.GetCardsSize() {
		s := ZwickerScoreOfCard(p.GetCard(i))
		if s < bestScore {
			best, bestScore = i, s
		}
	}
	return best
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (z *Zwicker) GetPlayers() []*ZwickerPlayer { return z.players }

// GetPlayer は idx のプレイヤーを返す。
func (z *Zwicker) GetPlayer(idx int) *ZwickerPlayer {
	return getPlayer(z.players, idx)
}

// GetPhase は現在のフェーズを返す。
func (z *Zwicker) GetPhase() ZwickerPhase { return z.phase }

// GetCurrentPlayerIdx は手番のプレイヤー添字を返す。
func (z *Zwicker) GetCurrentPlayerIdx() int { return z.currentIdx }

// GetTableCards は場の単札を返す。
func (z *Zwicker) GetTableCards() []*Card { return z.tableCards }

// GetBuilds は場のビルドを返す。
func (z *Zwicker) GetBuilds() []*ZwickerBuild { return z.builds }

// GetStockCount は山札の残り枚数を返す。
func (z *Zwicker) GetStockCount() int { return z.trumpCards.GetRemainingCount() }

// GetTeamScore はチームの累計得点を返す。
func (z *Zwicker) GetTeamScore(team int) int {
	if team < 0 || team > 1 {
		return 0
	}
	return z.scores[team]
}

// GetLastRoundScore は直近ディールの内訳を返す (未精算なら nil)。
func (z *Zwicker) GetLastRoundScore() *ZwickerRoundScore { return z.lastRound }

// GetGameEndFlag は決着しているかを返す。
func (z *Zwicker) GetGameEndFlag() bool { return z.gameEndFlag }

// GetWinnerTeam は勝ったチームを返す (-1: 未決着)。
func (z *Zwicker) GetWinnerTeam() int { return z.winnerTeam }

// GetConfig はゲーム設定を返す。
func (z *Zwicker) GetConfig() ZwickerConfig { return z.config }

// SetConfig はゲーム設定をセットする。
func (z *Zwicker) SetConfig(c ZwickerConfig) { z.config = c }

// GetActionLog は棋譜を返す。
func (z *Zwicker) GetActionLog() []*ActionLogEntry { return z.actionLog }

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (z *Zwicker) SetPhaseForTest(p ZwickerPhase) { z.phase = p }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (z *Zwicker) SetCurrentPlayerForTest(idx int) { z.currentIdx = idx }

// SetTableCardsForTest はテスト用に場を差し替える。
func (z *Zwicker) SetTableCardsForTest(cards []*Card) { z.tableCards = cards }

// SetTeamScoreForTest はテスト用に累計得点を差し替える。
func (z *Zwicker) SetTeamScoreForTest(team, score int) { z.scores[team] = score }

// SetDealStageForTest はテスト用に配り段階を差し替える。
func (z *Zwicker) SetDealStageForTest(stage int) { z.dealStage = stage }

// addLog は棋譜に 1 件追加する。
func (z *Zwicker) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	z.actionLog = append(z.actionLog, &ActionLogEntry{
		TurnNumber: len(z.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// zwickerJSON is the JSON wire format for Zwicker.
type zwickerJSON struct {
	TrumpCards *TrumpCards        `json:"tc"`
	Players    []*ZwickerPlayer   `json:"pl"`
	Config     ZwickerConfig      `json:"cfg"`
	Phase      ZwickerPhase       `json:"ph"`
	Table      []*Card            `json:"tb"`
	Builds     []*ZwickerBuild    `json:"bd"`
	Current    int                `json:"cur"`
	Dealer     int                `json:"dl"`
	DealStage  int                `json:"st"`
	LastCap    int                `json:"lc"`
	Scores     [2]int             `json:"sc"`
	LastRound  *ZwickerRoundScore `json:"lr"`
	GameEnd    bool               `json:"ge"`
	WinnerTeam int                `json:"wt"`
	ActionLog  []*ActionLogEntry  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (z *Zwicker) MarshalJSON() ([]byte, error) {
	return json.Marshal(zwickerJSON{
		TrumpCards: z.trumpCards, Players: z.players, Config: z.config, Phase: z.phase,
		Table: z.tableCards, Builds: z.builds, Current: z.currentIdx, Dealer: z.dealerIdx,
		DealStage: z.dealStage, LastCap: z.lastCaptureIdx, Scores: z.scores,
		LastRound: z.lastRound, GameEnd: z.gameEndFlag, WinnerTeam: z.winnerTeam,
		ActionLog: z.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// KV から戻る生バイト列は信用できないので、席数に合わせて詰め直し、設定を
// 検証する。**dealStage が落ちると配り直しが無限に走る**ので範囲も確かめる。
func (z *Zwicker) UnmarshalJSON(data []byte) error {
	var raw zwickerJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Players) != ZwickerPlayerCnt {
		return fmt.Errorf("expected %d players, got %d", ZwickerPlayerCnt, len(raw.Players))
	}
	if err := raw.Config.Validate(); err != nil {
		return err
	}
	if raw.Phase < ZwickerPhasePlay || raw.Phase > ZwickerPhaseGameEnd {
		return fmt.Errorf("unknown phase: %d", raw.Phase)
	}
	if raw.DealStage < 0 || raw.DealStage > len(zwickerDealSizes) {
		return fmt.Errorf("bad deal stage: %d", raw.DealStage)
	}

	z.trumpCards = raw.TrumpCards
	if z.trumpCards == nil {
		z.trumpCards = NewTrumpCards(ZwickerJokerCnt)
	}
	z.players = raw.Players
	z.config = raw.Config
	z.phase = raw.Phase
	z.tableCards = raw.Table
	z.dealStage = raw.DealStage
	z.scores = raw.Scores
	z.lastRound = raw.LastRound
	z.gameEndFlag = raw.GameEnd
	z.actionLog = raw.ActionLog

	z.currentIdx = clampZwickerIdx(raw.Current, len(z.players))
	z.dealerIdx = clampZwickerIdx(raw.Dealer, len(z.players))
	z.lastCaptureIdx = raw.LastCap
	if z.lastCaptureIdx < -1 || z.lastCaptureIdx >= len(z.players) {
		z.lastCaptureIdx = -1
	}
	z.winnerTeam = raw.WinnerTeam
	if z.winnerTeam < -1 || z.winnerTeam > 1 {
		z.winnerTeam = -1
	}

	z.builds = make([]*ZwickerBuild, 0, len(raw.Builds))
	for _, b := range raw.Builds {
		if b == nil || len(b.Cards) == 0 || b.Value <= 0 {
			continue
		}
		if b.Owner < 0 || b.Owner >= len(z.players) {
			continue
		}
		z.builds = append(z.builds, b)
	}
	return nil
}

// clampZwickerIdx は席番号を 0..n-1 に収める。
func clampZwickerIdx(idx, n int) int {
	if idx < 0 || idx >= n {
		return 0
	}
	return idx
}
