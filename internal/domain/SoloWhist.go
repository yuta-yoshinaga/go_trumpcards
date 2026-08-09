//go:build !js || !wasm || classic

// Package domain ソロ・ホイスト (Solo Whist) のドメインモデル。
//
// Solo Whist はイギリス発祥の 4 人用トリックテイキングゲーム。52 枚デッキを 13 枚ずつ
// 配り、入札フェーズで各プレイヤーが Pass / Solo (単独 8 トリック) / Misère (0 トリック・
// 切り札なし) / Abundance (単独 9 トリック) を 1 回ずつ宣言する。最高入札者が宣言者と
// なり残り 3 人の防御側と対戦する (本実装では Prop&Cop の臨時同盟は省略)。Solo/Abundance
// では宣言者の最長スートが切り札となる。マストフォローでトリックを進め、契約の達成可否で
// プレイヤー別ゲーム点を加減し、累積が目標 (既定 21) に達したプレイヤーが勝利する。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// SoloWhistPlayerCnt プレイヤー数 (人間 1 + CPU 3)
const SoloWhistPlayerCnt = 4

// SoloWhistHandSize 各プレイヤーの手札枚数
const SoloWhistHandSize = 13

// SoloWhistTrickCount 1 ラウンドのトリック数
const SoloWhistTrickCount = 13

// SoloWhistBid 入札種別
type SoloWhistBid int

// Solo Whist の入札定数 (序列は値の昇順)。
const (
	// SoloWhistBidPass パス
	SoloWhistBidPass SoloWhistBid = 0
	// SoloWhistBidSolo 単独 8 トリック
	SoloWhistBidSolo SoloWhistBid = 1
	// SoloWhistBidMisere 0 トリック・切り札なし
	SoloWhistBidMisere SoloWhistBid = 2
	// SoloWhistBidAbundance 単独 9 トリック
	SoloWhistBidAbundance SoloWhistBid = 3
)

// soloWhistBidTarget 契約の目標トリック数を返す。
func soloWhistBidTarget(b SoloWhistBid) int {
	switch b {
	case SoloWhistBidSolo:
		return 8
	case SoloWhistBidAbundance:
		return 9
	default: // Misère / Pass
		return 0
	}
}

// soloWhistBidValue 契約の得点価値を返す。
func soloWhistBidValue(b SoloWhistBid) int {
	switch b {
	case SoloWhistBidSolo:
		return 2
	case SoloWhistBidMisere:
		return 3
	case SoloWhistBidAbundance:
		return 4
	default:
		return 0
	}
}

// SoloWhistPhase ゲームフェーズ
type SoloWhistPhase int

// Solo Whist のフェーズ定数
const (
	// SoloWhistPhaseBid 入札フェーズ
	SoloWhistPhaseBid SoloWhistPhase = 0
	// SoloWhistPhasePlay トリックプレイフェーズ
	SoloWhistPhasePlay SoloWhistPhase = 1
	// SoloWhistPhaseTrickEnd トリック終了フェーズ
	SoloWhistPhaseTrickEnd SoloWhistPhase = 2
	// SoloWhistPhaseRoundEnd ラウンド終了フェーズ
	SoloWhistPhaseRoundEnd SoloWhistPhase = 3
	// SoloWhistPhaseGameEnd ゲーム終了フェーズ
	SoloWhistPhaseGameEnd SoloWhistPhase = 4
)

// SoloWhistHint ヒント情報
type SoloWhistHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// SoloWhist ソロ・ホイストのゲームクラス
type SoloWhist struct {
	trumpCards       *TrumpCards
	players          []*SoloWhistPlayer
	config           SoloWhistConfig
	phase            SoloWhistPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	bids             [SoloWhistPlayerCnt]SoloWhistBid // 各プレイヤーの入札
	bidDone          [SoloWhistPlayerCnt]bool         // 入札済みフラグ
	declarerIdx      int                              // 宣言者 (-1=未確定/全パス)
	contract         SoloWhistBid                     // 確定した契約
	trumpSuit        int                              // 切り札スート (0=なし=Misère)
	playerScores     [SoloWhistPlayerCnt]int          // 累積ゲーム点
	roundTricks      [SoloWhistPlayerCnt]int          // 現ラウンドの獲得トリック数
	gameEndFlag      bool
	winnerPlayer     int // -1=未確定
	actionLogBase
}

// NewSoloWhist コンストラクタ
func NewSoloWhist(trumpCards *TrumpCards, players []*SoloWhistPlayer, config SoloWhistConfig) *SoloWhist {
	return &SoloWhist{trumpCards: trumpCards, players: players, config: config, winnerPlayer: -1, declarerIdx: -1}
}

// NewDefaultSoloWhist 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultSoloWhist() *SoloWhist {
	players := make([]*SoloWhistPlayer, SoloWhistPlayerCnt)
	players[0] = NewSoloWhistPlayer(true)
	for i := 1; i < SoloWhistPlayerCnt; i++ {
		players[i] = NewSoloWhistPlayer(false)
	}
	return NewSoloWhist(NewTrumpCards(0), players, DefaultSoloWhistConfig())
}

// Reset ゲーム初期化
func (g *SoloWhist) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.playerScores = [SoloWhistPlayerCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *SoloWhist) NextRound() {
	if g.phase != SoloWhistPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % SoloWhistPlayerCnt
	g.startRound()
}

// startRound 手札を配り、入札フェーズを開始する。
func (g *SoloWhist) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.bids = [SoloWhistPlayerCnt]SoloWhistBid{}
	g.bidDone = [SoloWhistPlayerCnt]bool{}
	g.declarerIdx = -1
	g.contract = SoloWhistBidPass
	g.trumpSuit = 0
	g.roundTricks = [SoloWhistPlayerCnt]int{}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()
	g.sortAllHands()

	g.currentPlayerIdx = (g.dealerIdx + 1) % SoloWhistPlayerCnt // forehand bids first
	g.phase = SoloWhistPhaseBid
}

// deal 各プレイヤーへ 13 枚を配る。
func (g *SoloWhist) deal() {
	for i := 0; i < SoloWhistHandSize; i++ {
		for j := 0; j < SoloWhistPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % SoloWhistPlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
}

// --- Bidding ---

// IsHumanBidTurn 入札フェーズで人間の手番か。
func (g *SoloWhist) IsHumanBidTurn() bool {
	return g.phase == SoloWhistPhaseBid && g.currentPlayerIdx >= 0 &&
		g.currentPlayerIdx < len(g.players) && g.players[g.currentPlayerIdx].GetIsHuman()
}

// highestBid 現在の最高入札と入札者を返す (-1=なし)。
func (g *SoloWhist) highestBid() (SoloWhistBid, int) {
	best, bestIdx := SoloWhistBidPass, -1
	for i := 0; i < SoloWhistPlayerCnt; i++ {
		if g.bidDone[i] && g.bids[i] > best {
			best = g.bids[i]
			bestIdx = i
		}
	}
	return best, bestIdx
}

// PlayerBid 人間プレイヤーが入札する。
func (g *SoloWhist) PlayerBid(bid SoloWhistBid) error {
	if g.phase != SoloWhistPhaseBid {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyBid(g.currentPlayerIdx, bid)
}

// CpuBid 入札フェーズで CPU が 1 件入札する。
func (g *SoloWhist) CpuBid() {
	if g.phase != SoloWhistPhaseBid {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	_ = g.applyBid(idx, g.cpuChooseBid(idx))
}

// applyBid 入札を記録し、次の入札者へ進める。全員入札したら契約を確定する。
func (g *SoloWhist) applyBid(idx int, bid SoloWhistBid) error {
	high, _ := g.highestBid()
	// パス以外は現在の最高入札を上回る必要がある。
	if bid != SoloWhistBidPass && bid <= high {
		return NewDomainError(ErrInvalidPlay, "現在の入札を上回る必要があります")
	}
	g.bids[idx] = bid
	g.bidDone[idx] = true
	if bid != SoloWhistBidPass {
		g.appendLog(idx, "bid", fmt.Sprintf("%s bids %s", playerName(g.players, idx), soloWhistBidName(bid)), nil)
	} else {
		g.appendLog(idx, "bid", fmt.Sprintf("%s passes", playerName(g.players, idx)), nil)
	}
	// 次の未入札プレイヤーへ。
	for k := 1; k <= SoloWhistPlayerCnt; k++ {
		ni := (idx + k) % SoloWhistPlayerCnt
		if !g.bidDone[ni] {
			g.currentPlayerIdx = ni
			return nil
		}
	}
	g.resolveBidding()
	return nil
}

// resolveBidding 入札を締め、宣言者・契約・切り札を確定してプレイへ移る。
func (g *SoloWhist) resolveBidding() {
	bid, idx := g.highestBid()
	if idx < 0 || bid == SoloWhistBidPass {
		// 全員パス: ラウンドを流す。
		g.declarerIdx = -1
		g.phase = SoloWhistPhaseRoundEnd
		g.appendLog(-1, "passed_out", "all players passed; round is void", nil)
		return
	}
	g.declarerIdx = idx
	g.contract = bid
	if bid == SoloWhistBidMisere {
		g.trumpSuit = 0
	} else {
		g.trumpSuit = g.longestSuit(idx)
	}
	g.appendLog(idx, "contract",
		fmt.Sprintf("%s declares %s (trump %d)", playerName(g.players, idx), soloWhistBidName(bid), g.trumpSuit), nil)
	g.leadPlayerIdx = (g.dealerIdx + 1) % SoloWhistPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = SoloWhistPhasePlay
}

// longestSuit プレイヤーが最も多く持つスートを返す。
func (g *SoloWhist) longestSuit(playerIdx int) int {
	return longestSuit(g.players[playerIdx])
}

// cpuChooseBid CPU の入札を選ぶ。最長スートが長ければ Solo、極端に弱ければ Misère。
func (g *SoloWhist) cpuChooseBid(idx int) SoloWhistBid {
	high, _ := g.highestBid()
	suit := g.longestSuit(idx)
	cnt := 0
	highCards := 0
	p := g.players[idx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == suit {
			cnt++
		}
		if c.GetValue() == 1 || c.GetValue() >= 12 { // A,Q,K
			highCards++
		}
	}
	want := SoloWhistBidPass
	switch {
	case cnt >= 6 && highCards >= 4:
		want = SoloWhistBidAbundance
	case cnt >= 5 && highCards >= 2:
		want = SoloWhistBidSolo
	case highCards == 0:
		want = SoloWhistBidMisere
	}
	if want > high {
		return want
	}
	return SoloWhistBidPass
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *SoloWhist) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SoloWhistPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	card := player.GetCard(cardIndex)
	if err := g.validatePlay(g.currentPlayerIdx, card); err != nil {
		return err
	}
	played := player.RemoveCard(cardIndex)
	g.playCard(g.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行する。
func (g *SoloWhist) CpuPlay() {
	if g.gameEndFlag || g.phase != SoloWhistPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	cardIdx := g.cpuSelectPlayCard(idx)
	played := g.players[idx].RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	g.playCard(idx, played)
}

// playCard カードをプレイする共通処理。
func (g *SoloWhist) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == SoloWhistPlayerCnt {
		g.phase = SoloWhistPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % SoloWhistPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する。
func (g *SoloWhist) ResolveTrick() {
	if g.phase != SoloWhistPhaseTrickEnd || len(g.currentTrick) != SoloWhistPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
	}
	g.players[winnerIdx].AddTrick(trickCards)
	g.roundTricks[winnerIdx]++
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", playerName(g.players, winnerIdx), g.trickNumber), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= SoloWhistTrickCount {
		g.phase = SoloWhistPhaseRoundEnd
	} else {
		g.phase = SoloWhistPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *SoloWhist) NextTrick() {
	if g.phase != SoloWhistPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = SoloWhistPhasePlay
}

// ScoreRound 契約の達成可否を判定し、ゲーム点を加減してマッチ終了を判定する。
func (g *SoloWhist) ScoreRound() {
	if g.phase != SoloWhistPhaseRoundEnd {
		return
	}
	if g.declarerIdx >= 0 {
		value := soloWhistBidValue(g.contract)
		won := g.contractMade()
		if won {
			g.playerScores[g.declarerIdx] += value
		} else {
			for i := 0; i < SoloWhistPlayerCnt; i++ {
				if i != g.declarerIdx {
					g.playerScores[i] += value
				}
			}
		}
		g.appendLog(-1, "round_score",
			fmt.Sprintf("round %d: %s %s (%d tricks)",
				g.roundNumber, soloWhistBidName(g.contract),
				map[bool]string{true: "made", false: "failed"}[won], g.roundTricks[g.declarerIdx]), nil)
		g.checkGameEnd()
	}
}

// contractMade 宣言者が契約を達成したか。
func (g *SoloWhist) contractMade() bool {
	tricks := g.roundTricks[g.declarerIdx]
	switch g.contract {
	case SoloWhistBidMisere:
		return tricks == 0
	case SoloWhistBidSolo:
		return tricks >= soloWhistBidTarget(SoloWhistBidSolo)
	case SoloWhistBidAbundance:
		return tricks >= soloWhistBidTarget(SoloWhistBidAbundance)
	default:
		return false
	}
}

// checkGameEnd 目標点到達でマッチ終了を判定する。
func (g *SoloWhist) checkGameEnd() {
	leader, best := -1, -1
	for i := 0; i < SoloWhistPlayerCnt; i++ {
		if g.playerScores[i] > best {
			best = g.playerScores[i]
			leader = i
		}
	}
	if best >= g.config.TargetPoints && leader >= 0 {
		g.gameEndFlag = true
		g.winnerPlayer = leader
		g.phase = SoloWhistPhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the match!", playerName(g.players, leader)), nil)
	}
}

// --- Trick / play helpers ---

// validatePlay マストフォローを検証する。
func (g *SoloWhist) validatePlay(playerIdx int, card *Card) error {
	return validateFollowSuit(g.currentTrick, g.players, playerIdx, card)
}

// trickWinner トリックの勝者を決定する。切り札があれば最強切り札、なければ
// リードスートの最強札が勝つ。
func (g *SoloWhist) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	bestRank := g.soloWhistRank(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		if g.trumpSuit != 0 && tc.Card.GetDesign() != g.trumpSuit && tc.Card.GetDesign() != leadSuit {
			continue
		}
		if g.trumpSuit == 0 && tc.Card.GetDesign() != leadSuit {
			continue
		}
		if r := g.soloWhistRank(tc.Card); r > bestRank {
			bestRank = r
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// soloWhistRank トリック比較用ランク。切り札は非切り札より常に強い。A 高。
func (g *SoloWhist) soloWhistRank(card *Card) int {
	base := soloWhistCardStrength(card.GetValue())
	if g.trumpSuit != 0 && card.GetDesign() == g.trumpSuit {
		return 100 + base
	}
	return base
}

// soloWhistCardStrength カード強度 (A 高: A>K>Q>J>10>...>2)。
func soloWhistCardStrength(value int) int {
	if value == 1 {
		return 14
	}
	return value
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *SoloWhist) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.validatePlay(playerIdx, c) == nil })
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *SoloWhist) sortAllHands() {
	for _, p := range g.players {
		soloWhistSortHand(p)
	}
}

// soloWhistSortHand 手札をスート→強さ順にソートする。
func soloWhistSortHand(p *SoloWhistPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return soloWhistCardStrength(cards[i].GetValue()) > soloWhistCardStrength(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *SoloWhist) indexOfPlayerInTrick(playerIdx int) int {
	return indexOfPlayerInTrick(g.currentTrick, playerIdx)
}

// trickTopRank 現在のトリック勝者の札のランクを返す。見つからない場合は極小値。
func (g *SoloWhist) trickTopRank(winnerIdx int) int {
	idx := g.indexOfPlayerInTrick(winnerIdx)
	if idx < 0 {
		return -1 << 30
	}
	return g.soloWhistRank(g.currentTrick[idx].Card)
}

// --- CPU AI ---

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *SoloWhist) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == SoloWhistCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 宣言者かどうかとトリック状況を意識した戦略プレイ。
func (g *SoloWhist) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	isDeclarer := playerIdx == g.declarerIdx
	misere := g.contract == SoloWhistBidMisere
	// Misère の宣言者は勝ちたくないので最も弱い札を出す。
	if isDeclarer && misere {
		return pickLowest(player, valid, func(c *Card) int { return g.soloWhistRank(c) })
	}
	if len(g.currentTrick) == 0 {
		// リード: 宣言者は強い札、防御は弱い札。
		if isDeclarer {
			return pickHighest(player, valid, func(c *Card) int { return g.soloWhistRank(c) })
		}
		return pickLowest(player, valid, func(c *Card) int { return g.soloWhistRank(c) })
	}
	winnerIdx := g.trickWinner()
	topRank := g.trickTopRank(winnerIdx)
	winners := soloWhistFilter(valid, func(idx int) bool { return g.soloWhistRank(player.GetCard(idx)) > topRank })
	wantWin := isDeclarer != misere // 宣言者(非Misère)は勝ちたい; 防御は宣言者を負かしたい
	if wantWin && len(winners) > 0 {
		return pickLowest(player, winners, func(c *Card) int { return g.soloWhistRank(c) })
	}
	return pickLowest(player, valid, func(c *Card) int { return g.soloWhistRank(c) })
}

// soloWhistFilter 述語を満たすインデックスを抽出する。
func soloWhistFilter(indices []int, pred func(int) bool) []int {
	var out []int
	for _, idx := range indices {
		if pred(idx) {
			out = append(out, idx)
		}
	}
	return out
}

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨プレイを返す。
func (g *SoloWhist) GetHint() *SoloWhistHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.phase != SoloWhistPhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.getValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuPlaySmart(human, valid)
	return &SoloWhistHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *SoloWhist) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && (g.trumpSuit == 0 || card.GetDesign() != g.trumpSuit) {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	if g.soloWhistRank(card) > g.trickTopRank(winnerIdx) {
		return "follow_win"
	}
	return "follow_duck"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *SoloWhist) GetPhase() SoloWhistPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *SoloWhist) SetPhase(phase SoloWhistPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *SoloWhist) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *SoloWhist) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *SoloWhist) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *SoloWhist) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *SoloWhist) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *SoloWhist) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *SoloWhist) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *SoloWhist) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *SoloWhist) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *SoloWhist) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *SoloWhist) GetDealerIdx() int { return g.dealerIdx }

// GetDeclarerIdx 宣言者インデックス取得 (-1=未確定)
func (g *SoloWhist) GetDeclarerIdx() int { return g.declarerIdx }

// SetDeclarerIdx 宣言者インデックス設定 (テスト用)
func (g *SoloWhist) SetDeclarerIdx(idx int) { g.declarerIdx = idx }

// GetContract 確定した契約を取得
func (g *SoloWhist) GetContract() SoloWhistBid { return g.contract }

// SetContract 契約設定 (テスト用)
func (g *SoloWhist) SetContract(b SoloWhistBid) { g.contract = b }

// GetBids 各プレイヤーの入札を取得
func (g *SoloWhist) GetBids() [SoloWhistPlayerCnt]SoloWhistBid { return g.bids }

// GetBidDone は各プレイヤーが入札済みか（true=済み）を返す。未入札とパスを
// 区別できるよう、プレゼンター向けに公開する。
func (g *SoloWhist) GetBidDone() [SoloWhistPlayerCnt]bool { return g.bidDone }

// GetTrumpSuit 切り札スート取得 (0=なし)
func (g *SoloWhist) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *SoloWhist) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetPlayerScores プレイヤー別累積点取得
func (g *SoloWhist) GetPlayerScores() [SoloWhistPlayerCnt]int { return g.playerScores }

// SetPlayerScores プレイヤー別累積点設定 (テスト用)
func (g *SoloWhist) SetPlayerScores(s [SoloWhistPlayerCnt]int) { g.playerScores = s }

// GetRoundTricks 現ラウンドの獲得トリック数取得
func (g *SoloWhist) GetRoundTricks() [SoloWhistPlayerCnt]int { return g.roundTricks }

// SetRoundTricks 現ラウンドの獲得トリック数設定 (テスト用)
func (g *SoloWhist) SetRoundTricks(s [SoloWhistPlayerCnt]int) { g.roundTricks = s }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *SoloWhist) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝利プレイヤー取得 (-1=未確定)
func (g *SoloWhist) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayerCnt プレイヤー数取得
func (g *SoloWhist) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *SoloWhist) GetPlayer(i int) *SoloWhistPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間か (プレイフェーズ)。
func (g *SoloWhist) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *SoloWhist) GetConfig() SoloWhistConfig { return g.config }

// SetConfig 設定変更
func (g *SoloWhist) SetConfig(cfg SoloWhistConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *SoloWhist) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != SoloWhistPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// soloWhistJSON is the JSON wire format for SoloWhist.
type soloWhistJSON struct {
	TrumpCards       *TrumpCards                      `json:"tc"`
	Players          []*SoloWhistPlayer               `json:"ps"`
	Config           SoloWhistConfig                  `json:"cf"`
	Phase            SoloWhistPhase                   `json:"ph"`
	RoundNumber      int                              `json:"rn"`
	TrickNumber      int                              `json:"tn"`
	CurrentPlayerIdx int                              `json:"ci"`
	CurrentTrick     []*TrickCard                     `json:"ct"`
	LeadPlayerIdx    int                              `json:"li"`
	DealerIdx        int                              `json:"di"`
	Bids             [SoloWhistPlayerCnt]SoloWhistBid `json:"bd"`
	BidDone          [SoloWhistPlayerCnt]bool         `json:"bf"`
	DeclarerIdx      int                              `json:"dc"`
	Contract         SoloWhistBid                     `json:"co"`
	TrumpSuit        int                              `json:"ts"`
	PlayerScores     [SoloWhistPlayerCnt]int          `json:"sc"`
	RoundTricks      [SoloWhistPlayerCnt]int          `json:"rt"`
	GameEndFlag      bool                             `json:"ge"`
	WinnerPlayer     int                              `json:"wp"`
	ActionLog        []*ActionLogEntry                `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *SoloWhist) MarshalJSON() ([]byte, error) {
	return json.Marshal(soloWhistJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		LeadPlayerIdx:    g.leadPlayerIdx,
		DealerIdx:        g.dealerIdx,
		Bids:             g.bids,
		BidDone:          g.bidDone,
		DeclarerIdx:      g.declarerIdx,
		Contract:         g.contract,
		TrumpSuit:        g.trumpSuit,
		PlayerScores:     g.playerScores,
		RoundTricks:      g.roundTricks,
		GameEndFlag:      g.gameEndFlag,
		WinnerPlayer:     g.winnerPlayer,
		ActionLog:        g.actionLog,
	})
}

// soloWhistMaxSliceLen caps slice sizes during deserialisation.
const soloWhistMaxSliceLen = 5000

// errSoloWhistOversized is the single sentinel error for oversized input arrays.
var errSoloWhistOversized = errors.New("solowhist: input array exceeds maximum allowed size")

// errSoloWhistInvalidPlayers is returned when restored state lacks exactly SoloWhistPlayerCnt players.
var errSoloWhistInvalidPlayers = errors.New("solowhist: invalid player count")

// errSoloWhistInvalidTrick is returned when a restored trick card or its card is nil.
var errSoloWhistInvalidTrick = errors.New("solowhist: invalid trick card")

// UnmarshalJSON implements json.Unmarshaler.
func (g *SoloWhist) UnmarshalJSON(data []byte) error {
	var j soloWhistJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > soloWhistMaxSliceLen || len(j.CurrentTrick) > soloWhistMaxSliceLen ||
		len(j.ActionLog) > soloWhistMaxSliceLen {
		return errSoloWhistOversized
	}
	if len(j.Players) != SoloWhistPlayerCnt {
		return errSoloWhistInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errSoloWhistInvalidPlayers
		}
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil {
			return errSoloWhistInvalidTrick
		}
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*TrickCard, 0)
	}
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.bids = j.Bids
	g.bidDone = j.BidDone
	g.declarerIdx = j.DeclarerIdx
	g.contract = j.Contract
	g.trumpSuit = j.TrumpSuit
	g.playerScores = j.PlayerScores
	g.roundTricks = j.RoundTricks
	g.gameEndFlag = j.GameEndFlag
	g.winnerPlayer = j.WinnerPlayer
	g.actionLog = j.ActionLog
	return nil
}

// soloWhistBidName 入札種別の表示名を返す。
func soloWhistBidName(b SoloWhistBid) string {
	switch b {
	case SoloWhistBidSolo:
		return "Solo"
	case SoloWhistBidMisere:
		return "Misère"
	case SoloWhistBidAbundance:
		return "Abundance"
	default:
		return "Pass"
	}
}
