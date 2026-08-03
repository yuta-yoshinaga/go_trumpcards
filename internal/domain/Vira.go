//go:build !js || !wasm || extra

// Package domain ヴィーラ (Vira) のドメインモデル。
//
// Vira は 19 世紀スウェーデンで国民的に遊ばれた 3 人用のオークション・トリック
// テイキング。52 枚デッキから各自 13 枚を配り (本実装ではタロン 13 枚は使わない
// 簡略化)、入札で最高宣言をしたプレイヤーが「宣言者」となり残り 2 人と対戦する。
//
// 本家のビッド表は 30 種を超えるが、本実装では主要な 4 種に絞る (issue #4413 の
// 「主要ビッドのみ」方針):
//
//	Gask   目標 7 トリック / 価値 2
//	Solo   目標 8 トリック / 価値 4
//	Vira   目標 10 トリック / 価値 8  (ゲーム名を冠する最上位の常用宣言)
//	Misere 0 トリック・切り札なし / 価値 6
//
// 精算は **ポット式**。これが Préférence との一番の違いで、宣言の達成/失敗が
// プレイヤー間の直接のやり取りではなくポットを経由する:
//
//   - 各局の開始時に全員がアンティ 1 をポットへ入れる
//   - 宣言者が達成すると、ポットを総取りし、さらに守備側 2 人から契約価値を受け取る
//   - 失敗すると、契約価値をポットへ積み増し、守備側 2 人へも契約価値を支払う
//
// ポットは次局へ持ち越されるので、失敗が続くほど次の宣言者の見返りが大きくなる。
// 規定局数を終えた時点で持ち点が最大のプレイヤーが勝者。
package domain

import (
	"fmt"
	"math/rand"
	"sort"
)

// ViraPlayerCnt プレイヤー数 (人間 1 + CPU 2)。
const ViraPlayerCnt = 3

// ViraHandSize 各プレイヤーの手札枚数。
const ViraHandSize = 13

// ViraTrickCount 1 ラウンドのトリック数。
const ViraTrickCount = 13

// ViraAnte 各局の開始時に全員が支払うアンティ。
const ViraAnte = 1

// ViraBid 入札種別。
type ViraBid int

const (
	// ViraBidPass パス。
	ViraBidPass ViraBid = iota
	// ViraBidGask 7 トリック宣言 (最も低い常用宣言)。
	ViraBidGask
	// ViraBidSolo 8 トリック宣言。
	ViraBidSolo
	// ViraBidMisere 0 トリック・切り札なし。
	ViraBidMisere
	// ViraBidVira 10 トリック宣言。ゲーム名を冠する最上位。
	ViraBidVira
)

// ViraBidTarget 契約の目標トリック数を返す。Misère と Pass は 0。
//
// **公開しているのはプレゼンタが同じ表を持たないため。**CUI 側に写すと、
// 階梯を変えたときに片方だけ直して黙って食い違う。
func ViraBidTarget(b ViraBid) int {
	switch b {
	case ViraBidGask:
		return 7
	case ViraBidSolo:
		return 8
	case ViraBidVira:
		return 10
	default: // Misère / Pass
		return 0
	}
}

// viraBidValue 契約の価値を返す。守備側との受け渡し額であり、
// 失敗時にポットへ積む額でもある。
func viraBidValue(b ViraBid) int {
	switch b {
	case ViraBidGask:
		return 2
	case ViraBidSolo:
		return 4
	case ViraBidMisere:
		return 6
	case ViraBidVira:
		return 8
	default:
		return 0
	}
}

// ViraBidNames 表示用のビッド名。
var ViraBidNames = []string{"Pass", "Gask", "Solo", "Misère", "Vira"}

// ViraPhase ゲームフェーズ。
type ViraPhase int

const (
	// ViraPhaseBid 入札フェーズ。
	ViraPhaseBid ViraPhase = iota
	// ViraPhasePlay トリックプレイ。
	ViraPhasePlay
	// ViraPhaseTrickEnd トリック終了 (結果表示待ち)。
	ViraPhaseTrickEnd
	// ViraPhaseRoundEnd ラウンド終了 (精算済み)。
	ViraPhaseRoundEnd
	// ViraPhaseGameEnd マッチ終了。
	ViraPhaseGameEnd
)

// Vira ヴィーラのゲーム本体。
type Vira struct {
	trumpCards       *TrumpCards
	players          []*ViraPlayer
	config           ViraConfig
	rng              *rand.Rand
	phase            ViraPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	bids             [ViraPlayerCnt]ViraBid
	bidDone          [ViraPlayerCnt]bool
	declarerIdx      int // -1 = 未確定 / 全パス
	contract         ViraBid
	trumpSuit        int
	pot              int
	playerScores     [ViraPlayerCnt]int
	roundTricks      [ViraPlayerCnt]int
	lastRoundDelta   [ViraPlayerCnt]int
	lastRoundMade    bool
	gameEndFlag      bool
	winnerPlayer     int // -1 = 未確定 (同点)
	actionLog        []*ActionLogEntry
	shuffled         []*Card // rng 差し替え時の並べ替え済み山
}

// NewVira コンストラクタ。
func NewVira(trumpCards *TrumpCards, players []*ViraPlayer, config ViraConfig) *Vira {
	return &Vira{trumpCards: trumpCards, players: players, config: config, rng: nil, declarerIdx: -1, winnerPlayer: -1}
}

// NewDefaultVira 標準の 3 人構成 (人間 1, CPU 2) と既定設定で生成する。
func NewDefaultVira() *Vira {
	players := make([]*ViraPlayer, ViraPlayerCnt)
	players[0] = NewViraPlayer(true)
	for i := 1; i < ViraPlayerCnt; i++ {
		players[i] = NewViraPlayer(false)
	}
	return NewVira(NewTrumpCards(0), players, DefaultViraConfig())
}

// SetRand 乱数生成器を差し替える (テスト用)。
func (g *Vira) SetRand(r *rand.Rand) { g.rng = r }

// shuffle 山札をシャッフルする。
//
// **rng は配り札の決定性のためにある。**テストが手札を固定できないと、
// 「Reset 直後に主張する」形のテストが確率的に落ちる (#4467 で一度 develop を
// 赤くしている)。`TrumpCards` 側に差し替え口が無いので、Cuckoo と同じく
// ゲーム側で持って山を並べ替える。
func (g *Vira) shuffle() {
	if g.rng == nil {
		g.trumpCards.Shuffle()
		return
	}
	deck := make([]*Card, 0, ViraPlayerCnt*ViraHandSize)
	for {
		c := g.trumpCards.DrawCard()
		if c == nil {
			break
		}
		deck = append(deck, c)
	}
	g.rng.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	g.shuffled = deck
}

// drawCard 山から 1 枚引く。rng を差し替えているときは並べ替え済みの山から取る。
func (g *Vira) drawCard() *Card {
	if g.shuffled == nil {
		return g.trumpCards.DrawCard()
	}
	if len(g.shuffled) == 0 {
		return nil
	}
	c := g.shuffled[0]
	g.shuffled = g.shuffled[1:]
	return c
}

// Reset ゲーム初期化。
func (g *Vira) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.pot = 0
	g.playerScores = [ViraPlayerCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する。規定局数に達していればマッチを終える。
func (g *Vira) NextRound() {
	if g.phase != ViraPhaseRoundEnd {
		return
	}
	if g.roundNumber >= g.config.TargetRounds {
		g.finishMatch()
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % ViraPlayerCnt
	g.startRound()
}

// startRound アンティを集め、手札を配り、入札フェーズを開始する。
func (g *Vira) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.bids = [ViraPlayerCnt]ViraBid{}
	g.bidDone = [ViraPlayerCnt]bool{}
	g.declarerIdx = -1
	g.contract = ViraBidPass
	g.trumpSuit = 0
	g.roundTricks = [ViraPlayerCnt]int{}
	g.lastRoundDelta = [ViraPlayerCnt]int{}
	g.lastRoundMade = false
	for _, p := range g.players {
		p.ResetRound()
	}

	// **アンティは全員から等しく取る。**ポットが次局へ持ち越されるので、
	// ここを飛ばすと失敗が続いたときの見返りが積み上がらない。
	for i := range g.players {
		g.playerScores[i] -= ViraAnte
		g.pot += ViraAnte
	}

	g.trumpCards.Replenish()
	g.shuffled = nil
	g.shuffle()
	g.deal()
	g.sortAllHands()

	g.leadPlayerIdx = (g.dealerIdx + 1) % ViraPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = ViraPhaseBid
	g.appendLog(-1, "deal", fmt.Sprintf("ラウンド %d 開始 (ポット %d)", g.roundNumber, g.pot), nil)
}

// deal 各プレイヤーへ 13 枚ずつ配る。残り 13 枚のタロンは使わない。
func (g *Vira) deal() {
	for i := 0; i < ViraHandSize; i++ {
		for j := 0; j < ViraPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % ViraPlayerCnt
			if c := g.drawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
}

// sortAllHands 全員の手札をスート・ランク順に整列する。
func (g *Vira) sortAllHands() {
	for _, p := range g.players {
		cards := make([]*Card, 0, p.GetCardsSize())
		for i := 0; i < p.GetCardsSize(); i++ {
			cards = append(cards, p.GetCard(i))
		}
		sort.SliceStable(cards, func(a, b int) bool {
			if cards[a].GetDesign() != cards[b].GetDesign() {
				return cards[a].GetDesign() < cards[b].GetDesign()
			}
			return viraCardStrength(cards[a].GetValue()) < viraCardStrength(cards[b].GetValue())
		})
		p.Reset()
		for _, c := range cards {
			p.AddCard(c)
		}
	}
}

// viraCardStrength カード強度 (A 高: A > K > Q > J > 10 > … > 2)。
func viraCardStrength(value int) int {
	if value == 1 {
		return 14 // Ace
	}
	return value
}

// IsHumanBidTurn 人間の入札手番か。
func (g *Vira) IsHumanBidTurn() bool {
	return g.phase == ViraPhaseBid && g.players[g.currentPlayerIdx].GetIsHuman() && !g.bidDone[g.currentPlayerIdx]
}

// highestBid 現在の最高入札とその席を返す。誰も宣言していなければ (Pass, -1)。
func (g *Vira) highestBid() (ViraBid, int) {
	best := ViraBidPass
	idx := -1
	for i, b := range g.bids {
		if b > best {
			best = b
			idx = i
		}
	}
	return best, idx
}

// PlayerBid 人間プレイヤーが入札する。
func (g *Vira) PlayerBid(bid ViraBid) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ViraPhaseBid {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyBid(g.currentPlayerIdx, bid)
}

// applyBid 入札を適用し、手番を進める。
func (g *Vira) applyBid(idx int, bid ViraBid) error {
	if bid < ViraBidPass || bid > ViraBidVira {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("不正な入札です: %d", bid))
	}
	if g.bidDone[idx] {
		return NewDomainError(ErrInvalidPlay, "既に入札済みです")
	}
	// **パス以外は現在の最高入札を上回らなければならない。**同値を許すと
	// 誰が宣言者になるかが席順だけで決まり、階梯が意味を失う。
	if bid != ViraBidPass {
		if best, _ := g.highestBid(); bid <= best {
			return NewDomainError(ErrInvalidPlay,
				fmt.Sprintf("%s を上回る宣言が必要です", ViraBidNames[best]))
		}
	}
	g.bids[idx] = bid
	g.bidDone[idx] = true
	g.appendLog(idx, "bid", ViraBidNames[bid], nil)

	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % ViraPlayerCnt
	if g.allBidsDone() {
		g.resolveBidding()
	}
	return nil
}

// allBidsDone 全員が入札を終えたか。
func (g *Vira) allBidsDone() bool {
	for _, done := range g.bidDone {
		if !done {
			return false
		}
	}
	return true
}

// CpuBid 現在の手番が CPU の場合に 1 回入札する。
func (g *Vira) CpuBid() {
	if g.gameEndFlag || g.phase != ViraPhaseBid {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() || g.bidDone[idx] {
		return
	}
	_ = g.applyBid(idx, g.cpuChooseBid(idx))
}

// cpuChooseBid CPU の入札を選ぶ。手札の強さから届きそうな最高の宣言を返す。
func (g *Vira) cpuChooseBid(idx int) ViraBid {
	best, _ := g.highestBid()
	p := g.players[idx]

	// 絵札と A の枚数、最長スートの長さで大まかに測る。
	honours := 0
	suitLen := [CardDesignMax + 1]int{}
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if viraCardStrength(c.GetValue()) >= 12 {
			honours++
		}
		suitLen[c.GetDesign()]++
	}
	longest := 0
	for _, n := range suitLen {
		if n > longest {
			longest = n
		}
	}

	// **低い手ほど Misère に向く。**絵札が無いことは「取れない」ではなく
	// 「取らずに済む」なので、宣言できない手ではなく別種の手になる。
	want := ViraBidPass
	switch {
	case honours >= 7 && longest >= 6:
		want = ViraBidVira
	case honours <= 1:
		want = ViraBidMisere
	case honours >= 5 && longest >= 5:
		want = ViraBidSolo
	case honours >= 4:
		want = ViraBidGask
	}
	if want <= best {
		return ViraBidPass
	}
	return want
}

// resolveBidding 入札を締め、宣言者と切り札を決める。全パスなら流局。
func (g *Vira) resolveBidding() {
	best, idx := g.highestBid()
	if idx < 0 {
		// **全パスは流局。**ポットはそのまま次局へ持ち越す。
		g.appendLog(-1, "allpass", "全員パス。ポットを持ち越して次のラウンドへ", nil)
		g.phase = ViraPhaseRoundEnd
		return
	}
	g.declarerIdx = idx
	g.contract = best
	if best != ViraBidMisere {
		g.trumpSuit = g.longestSuit(idx)
	}
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = ViraPhasePlay
	g.appendLog(idx, "declare",
		fmt.Sprintf("%s を宣言 (切り札 %s)", ViraBidNames[best], viraSuitName(g.trumpSuit)), nil)
}

// viraSuitName 切り札スートの表示名。0 は切り札なし。
func viraSuitName(suit int) string {
	switch suit {
	case CardDesignSpade:
		return "スペード"
	case CardDesignClover:
		return "クラブ"
	case CardDesignHeart:
		return "ハート"
	case CardDesignDiamond:
		return "ダイヤ"
	default:
		return "なし"
	}
}

// longestSuit プレイヤーの最長スートを返す。同数なら番号の小さい方。
func (g *Vira) longestSuit(playerIdx int) int {
	counts := [CardDesignMax + 1]int{}
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		counts[p.GetCard(i).GetDesign()]++
	}
	bestSuit, bestCnt := CardDesignSpade, -1
	for suit := 1; suit <= CardDesignMax; suit++ {
		if counts[suit] > bestCnt {
			bestSuit, bestCnt = suit, counts[suit]
		}
	}
	return bestSuit
}

// appendLog 棋譜に 1 行追加する。
func (g *Vira) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: len(g.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// finishMatch マッチを終え、持ち点最大のプレイヤーを勝者にする。
func (g *Vira) finishMatch() {
	g.gameEndFlag = true
	g.phase = ViraPhaseGameEnd
	bestScore := g.playerScores[0]
	winner := 0
	tie := false
	for i := 1; i < ViraPlayerCnt; i++ {
		switch {
		case g.playerScores[i] > bestScore:
			bestScore, winner, tie = g.playerScores[i], i, false
		case g.playerScores[i] == bestScore:
			tie = true
		}
	}
	// **同点なら勝者なし。**席順で決めると、常に若い席が得をする。
	if tie {
		g.winnerPlayer = -1
	} else {
		g.winnerPlayer = winner
	}
	g.appendLog(-1, "gameend", fmt.Sprintf("マッチ終了 (ポット残 %d)", g.pot), nil)
}

// IsHumanTurn 人間のプレイ手番か。
func (g *Vira) IsHumanTurn() bool {
	return g.phase == ViraPhasePlay && g.players[g.currentPlayerIdx].GetIsHuman()
}

// PlayerPlay 人間プレイヤーが手札の 1 枚を出す。
func (g *Vira) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ViraPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	if err := g.validatePlay(g.currentPlayerIdx, player.GetCard(cardIndex)); err != nil {
		return err
	}
	g.playCard(g.currentPlayerIdx, player.RemoveCard(cardIndex))
	return nil
}

// CpuPlay 現在の手番が CPU の場合に 1 枚出す。
func (g *Vira) CpuPlay() {
	if g.gameEndFlag || g.phase != ViraPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	played := g.players[idx].RemoveCard(g.cpuSelectPlayCard(idx))
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	g.playCard(idx, played)
}

// validatePlay マストフォローを検証する。
func (g *Vira) validatePlay(playerIdx int, card *Card) error {
	if card == nil {
		return NewDomainError(ErrInvalidCard, "カードがありません")
	}
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if g.playerHasSuit(playerIdx, leadSuit) && card.GetDesign() != leadSuit {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasSuit プレイヤーが指定スートを持っているか。
func (g *Vira) playerHasSuit(playerIdx, design int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == design {
			return true
		}
	}
	return false
}

// GetValidPlayIndices 出せる手札の位置を返す。
func (g *Vira) GetValidPlayIndices(playerIdx int) []int {
	p := g.players[playerIdx]
	var all []int
	for i := 0; i < p.GetCardsSize(); i++ {
		all = append(all, i)
	}
	if len(g.currentTrick) == 0 {
		return all
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	var follow []int
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == leadSuit {
			follow = append(follow, i)
		}
	}
	if len(follow) > 0 {
		return follow
	}
	return all
}

// playCard 1 枚を場に出し、トリックが揃えばフェーズを進める。
func (g *Vira) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", "", []*Card{card})
	if len(g.currentTrick) < ViraPlayerCnt {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % ViraPlayerCnt
		return
	}
	g.phase = ViraPhaseTrickEnd
}

// ResolveTrick トリックの勝者を決め、次のトリックへ進む。
func (g *Vira) ResolveTrick() {
	if g.phase != ViraPhaseTrickEnd {
		return
	}
	winnerIdx := g.trickWinner()
	cards := make([]*Card, len(g.currentTrick))
	for i, tc := range g.currentTrick {
		cards[i] = tc.Card
	}
	g.players[winnerIdx].AddTrick(cards)
	g.roundTricks[winnerIdx]++
	g.appendLog(winnerIdx, "trickwin", fmt.Sprintf("トリック %d を獲得", g.trickNumber), cards)

	// **場は NextTrick までクリアしない。**共通の CPU ループ
	// (`runCpuTurnsLoop`) は TrickEnd フェーズで結果を見せてから NextTrick を
	// 呼ぶ形なので、ここで消すと勝ち札が一瞬も表示されない。
	g.leadPlayerIdx = winnerIdx
	g.currentPlayerIdx = winnerIdx
	if g.trickNumber >= ViraTrickCount {
		g.settleRound()
	}
}

// NextTrick 次のトリックを開始する。
func (g *Vira) NextTrick() {
	if g.phase != ViraPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = ViraPhasePlay
}

// ScoreRound ラウンド終了時にマッチ終了を判定する。
//
// 精算そのものは最終トリックの解決時 (`settleRound`) に済んでいる。ここは
// 共通インタラクタが「ラウンドを締める」ために呼ぶ入り口で、規定局数に
// 達していればマッチを終える。
func (g *Vira) ScoreRound() {
	if g.phase != ViraPhaseRoundEnd {
		return
	}
	if g.roundNumber >= g.config.TargetRounds {
		g.finishMatch()
	}
}

// trickWinner 現在のトリックの勝者を返す。
func (g *Vira) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	bestRank := g.viraRank(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		// 切り札でもリードスートでもない札は勝てない。
		if g.trumpSuit != 0 && tc.Card.GetDesign() != g.trumpSuit && tc.Card.GetDesign() != leadSuit {
			continue
		}
		if g.trumpSuit == 0 && tc.Card.GetDesign() != leadSuit {
			continue
		}
		if r := g.viraRank(tc.Card); r > bestRank {
			bestRank = r
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// viraRank トリック比較用ランク。切り札は非切り札より常に強い。
func (g *Vira) viraRank(card *Card) int {
	base := viraCardStrength(card.GetValue())
	if g.trumpSuit != 0 && card.GetDesign() == g.trumpSuit {
		return 100 + base
	}
	return base
}

// cpuSelectPlayCard CPU が出す札を選ぶ。
func (g *Vira) cpuSelectPlayCard(idx int) int {
	valid := g.GetValidPlayIndices(idx)
	if len(valid) == 0 {
		return 0
	}
	if g.config.CpuDifficulty == ViraCpuDifficultyEasy {
		return valid[0]
	}
	p := g.players[idx]
	// **Misère の守備側は逆を行く。**宣言者に取らせたいので、勝てる札を優先する。
	// 宣言者自身は取りたくないので一番弱い札を出す。
	wantHigh := g.contract == ViraBidMisere && idx != g.declarerIdx
	best := valid[0]
	for _, i := range valid[1:] {
		cur := g.viraRank(p.GetCard(i))
		ref := g.viraRank(p.GetCard(best))
		if (wantHigh && cur > ref) || (!wantHigh && cur < ref) {
			best = i
		}
	}
	return best
}

// settleRound ラウンドを精算する。ポットを介した受け渡しを行う。
func (g *Vira) settleRound() {
	g.phase = ViraPhaseRoundEnd
	if g.declarerIdx < 0 {
		return
	}
	value := viraBidValue(g.contract)
	won := g.roundTricks[g.declarerIdx]
	made := won >= ViraBidTarget(g.contract)
	if g.contract == ViraBidMisere {
		// Misère は 1 トリックも取らないことが条件。
		made = won == 0
	}
	g.lastRoundMade = made

	before := g.playerScores
	if made {
		// **ポットは総取り。**アンティが積み上がっているほど見返りが大きい。
		g.playerScores[g.declarerIdx] += g.pot
		g.pot = 0
		for i := range g.players {
			if i == g.declarerIdx {
				continue
			}
			g.playerScores[i] -= value
			g.playerScores[g.declarerIdx] += value
		}
	} else {
		// 失敗すると契約価値をポットへ積み、守備側にも同額を払う。
		g.playerScores[g.declarerIdx] -= value
		g.pot += value
		for i := range g.players {
			if i == g.declarerIdx {
				continue
			}
			g.playerScores[i] += value
			g.playerScores[g.declarerIdx] -= value
		}
	}
	for i := range g.playerScores {
		g.lastRoundDelta[i] = g.playerScores[i] - before[i]
	}
	g.appendLog(g.declarerIdx, "settle",
		fmt.Sprintf("%s %s (%d トリック, ポット %d)", ViraBidNames[g.contract], viraMadeLabel(made), won, g.pot), nil)
}

// viraMadeLabel 達成可否の表示。
func viraMadeLabel(made bool) string {
	if made {
		return "成功"
	}
	return "失敗"
}

// GetPhase 現在のフェーズ。
func (g *Vira) GetPhase() ViraPhase { return g.phase }

// SetPhase フェーズを設定する (テスト用)。
func (g *Vira) SetPhase(p ViraPhase) { g.phase = p }

// GetRoundNumber 現在のラウンド番号。
func (g *Vira) GetRoundNumber() int { return g.roundNumber }

// GetTrickNumber 現在のトリック番号。
func (g *Vira) GetTrickNumber() int { return g.trickNumber }

// GetCurrentPlayerIdx 現在の手番。
func (g *Vira) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx 手番を設定する (テスト用)。
func (g *Vira) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 場に出ている札。
func (g *Vira) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// GetLeadPlayerIdx リードした席。
func (g *Vira) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// GetDealerIdx ディーラーの席。
func (g *Vira) GetDealerIdx() int { return g.dealerIdx }

// GetDeclarerIdx 宣言者の席。-1 は未確定/流局。
func (g *Vira) GetDeclarerIdx() int { return g.declarerIdx }

// GetContract 成立した契約。
func (g *Vira) GetContract() ViraBid { return g.contract }

// GetBids 各席の入札。
func (g *Vira) GetBids() [ViraPlayerCnt]ViraBid { return g.bids }

// GetBidDone 各席が入札を終えたか。
func (g *Vira) GetBidDone() [ViraPlayerCnt]bool { return g.bidDone }

// GetTrumpSuit 切り札スート。0 は切り札なし。
func (g *Vira) GetTrumpSuit() int { return g.trumpSuit }

// GetPot 現在のポット。
func (g *Vira) GetPot() int { return g.pot }

// GetPlayerScores 各席の持ち点。
func (g *Vira) GetPlayerScores() [ViraPlayerCnt]int { return g.playerScores }

// GetRoundTricks このラウンドで各席が取ったトリック数。
func (g *Vira) GetRoundTricks() [ViraPlayerCnt]int { return g.roundTricks }

// GetLastRoundDelta 直前のラウンドでの持ち点の増減。
func (g *Vira) GetLastRoundDelta() [ViraPlayerCnt]int { return g.lastRoundDelta }

// GetLastRoundMade 直前のラウンドで契約が達成されたか。
func (g *Vira) GetLastRoundMade() bool { return g.lastRoundMade }

// GetGameEndFlag マッチが終了したか。
func (g *Vira) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝者の席。-1 は未確定または同点。
func (g *Vira) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayers 全プレイヤー。
func (g *Vira) GetPlayers() []*ViraPlayer { return g.players }

// GetPlayer 指定席のプレイヤー。範囲外は nil。
func (g *Vira) GetPlayer(idx int) *ViraPlayer {
	if idx < 0 || idx >= len(g.players) {
		return nil
	}
	return g.players[idx]
}

// GetConfig 設定。
func (g *Vira) GetConfig() ViraConfig { return g.config }

// SetConfig 設定を差し替える。
//
// **検証はここではしない。**共通の `resetWithValidatedConfig` が呼ぶ前に
// `Validate` を通す契約なので、ここで戻り値を持つと共通ヘルパーに載らない。
// 直接呼ぶ場合も、呼び出し側が `Validate` を通してから渡すこと。
func (g *Vira) SetConfig(c ViraConfig) { g.config = c }

// GetActionLog 棋譜。
func (g *Vira) GetActionLog() []*ActionLogEntry {
	if g.actionLog == nil {
		return []*ActionLogEntry{}
	}
	return g.actionLog
}

// ForcePassForTest 指定席を強制的にパスさせる (テスト用)。
//
// CPU の入札は手札の強さから決まるため、テストから「全員パス」を確実に作る
// 手段が無い。入札の適用経路そのものは applyBid を通るので、検証したい
// 「全パスなら流局しポットは持ち越す」の筋道は本番と同じものを通る。
func (g *Vira) ForcePassForTest(idx int) error {
	if idx < 0 || idx >= ViraPlayerCnt {
		return NewDomainError(ErrInvalidPlay, "席が範囲外です")
	}
	return g.applyBid(idx, ViraBidPass)
}

// ViraHint ヒント情報。
type ViraHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// GetPlayerCnt プレイヤー数。
func (g *Vira) GetPlayerCnt() int { return len(g.players) }

// GetPlayableIndices プレイ可能なカードの位置。
func (g *Vira) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	return g.GetValidPlayIndices(playerIdx)
}

// findHumanIdx 人間プレイヤーの席。いなければ -1。
func (g *Vira) findHumanIdx() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// GetHint 人間プレイヤーへのヒント。手番でなければ nil。
func (g *Vira) GetHint() *ViraHint {
	human := g.findHumanIdx()
	if human < 0 || g.phase != ViraPhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.GetValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuSelectPlayCard(human)
	return &ViraHint{CardIndices: []int{idx}, Reason: g.playHintReason(human)}
}

// playHintReason ヒント理由キーを判定する。
//
// **Misère は宣言者と守備側で意味が逆になる。**宣言者は 1 枚も取ってはならず、
// 守備側は取らせたい。同じ「強い札を出せ」でも狙いが反対なので、キーを分ける。
func (g *Vira) playHintReason(playerIdx int) string {
	misere := g.contract == ViraBidMisere
	declarer := playerIdx == g.declarerIdx
	switch {
	case misere && declarer:
		return "misere_duck"
	case misere:
		return "misere_force"
	case len(g.currentTrick) == 0 && declarer:
		return "lead_high"
	case len(g.currentTrick) == 0:
		return "lead_low"
	case declarer:
		return "follow_win"
	default:
		return "follow_block"
	}
}
