//go:build !js || !wasm || casino

// Package domain オークション・フォーティファイブズ (Auction Forty-Fives) のドメインモデル。
//
// Auction Forty-Fives はアイルランド・カナダ (ノバスコシア) の 4 人 2 チームの入札トリック
// テイキング。52 枚デッキを 5 枚ずつ配り、入札フェーズで各プレイヤーが 15/20/25 点 (または
// Pass) を 1 回宣言する。最高入札者のチームが落札チームとなり切り札を宣言する (本実装では
// 落札者の最長スートを自動選択し、カード交換は省略)。トリックの強さは Spoil Five 系の固定
// 最強牌ランク (切り札 5 > 切り札 J > ♥A > 切り札 A > K > Q > その他切り札 > リードスート)。
// 上位 3 枚 (切り札 5・J・♥A) はフォロー免除 (Reneging)。各トリックは 5 点で、切り札 5 を含む
// トリックが常に最高得点トリックとなる。落札チームが宣言点以上を取れば加点、失敗は宣言点を
// 失点。非落札チームは取得点を常に加算。先に 45 点へ到達したチームが勝利。Jink (25 を宣言し
// 全 5 トリック取得) は即勝利。
//
// 簡略化: 名誉点 (切り札 J/♥A の追加点) は実装せず一律 1 トリック 5 点とする。赤/黒スートでの
// 数札ピップ順反転は簡略化し一律 10-high とする。カード交換フェーズは省略。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// FortyFivesPlayerCnt プレイヤー数 (人間 1 + CPU 3)
const FortyFivesPlayerCnt = 4

// FortyFivesTeamCnt チーム数
const FortyFivesTeamCnt = 2

// FortyFivesHandSize 各プレイヤーの手札枚数
const FortyFivesHandSize = 5

// FortyFivesTrickCount 1 ラウンドのトリック数
const FortyFivesTrickCount = 5

// FortyFivesPointsPerTrick 1 トリックあたりの得点
const FortyFivesPointsPerTrick = 5

// FortyFivesBid 入札種別 (値はそのまま宣言点; Pass=0)
type FortyFivesBid int

// Forty-Fives の入札定数
const (
	// FortyFivesBidPass パス
	FortyFivesBidPass FortyFivesBid = 0
	// FortyFivesBidFifteen 15 点宣言
	FortyFivesBidFifteen FortyFivesBid = 15
	// FortyFivesBidTwenty 20 点宣言
	FortyFivesBidTwenty FortyFivesBid = 20
	// FortyFivesBidTwentyFive 25 点宣言 (Jink = 全トリック)
	FortyFivesBidTwentyFive FortyFivesBid = 25
)

// FortyFivesPhase ゲームフェーズ
type FortyFivesPhase int

// Forty-Fives のフェーズ定数
const (
	// FortyFivesPhaseBid 入札フェーズ
	FortyFivesPhaseBid FortyFivesPhase = 0
	// FortyFivesPhasePlay トリックプレイフェーズ
	FortyFivesPhasePlay FortyFivesPhase = 1
	// FortyFivesPhaseTrickEnd トリック終了フェーズ
	FortyFivesPhaseTrickEnd FortyFivesPhase = 2
	// FortyFivesPhaseRoundEnd ラウンド終了フェーズ
	FortyFivesPhaseRoundEnd FortyFivesPhase = 3
	// FortyFivesPhaseGameEnd ゲーム終了フェーズ
	FortyFivesPhaseGameEnd FortyFivesPhase = 4
)

// FortyFivesHint ヒント情報
type FortyFivesHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// FortyFives オークション・フォーティファイブズのゲームクラス
type FortyFives struct {
	trumpCards       *TrumpCards
	players          []*FortyFivesPlayer
	config           FortyFivesConfig
	phase            FortyFivesPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	bids             [FortyFivesPlayerCnt]FortyFivesBid
	bidDone          [FortyFivesPlayerCnt]bool
	declarerIdx      int // 落札者 (-1=未確定/全パス)
	contract         FortyFivesBid
	trumpSuit        int
	teamScores       [FortyFivesTeamCnt]int // 累積点
	roundTeamPts     [FortyFivesTeamCnt]int // 現ラウンドのチーム別トリック得点
	gameEndFlag      bool
	winnerTeam       int // -1=未確定
	actionLog        []*ActionLogEntry
}

// NewFortyFives コンストラクタ
func NewFortyFives(trumpCards *TrumpCards, players []*FortyFivesPlayer, config FortyFivesConfig) *FortyFives {
	return &FortyFives{trumpCards: trumpCards, players: players, config: config, winnerTeam: -1, declarerIdx: -1}
}

// NewDefaultFortyFives 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultFortyFives() *FortyFives {
	players := make([]*FortyFivesPlayer, FortyFivesPlayerCnt)
	players[0] = NewFortyFivesPlayer(true)
	for i := 1; i < FortyFivesPlayerCnt; i++ {
		players[i] = NewFortyFivesPlayer(false)
	}
	return NewFortyFives(NewTrumpCards(0), players, DefaultFortyFivesConfig())
}

// FortyFivesTeamOf プレイヤーが属するチーム (0 = 席0&2, 1 = 席1&3)
func FortyFivesTeamOf(playerIdx int) int { return playerIdx % FortyFivesTeamCnt }

// fortyFivesTeamName チーム番号を表示名 (A/B) に変換する (casino ワーカーで自己完結)。
func fortyFivesTeamName(team int) string {
	if team == 0 {
		return "A"
	}
	return "B"
}

// Reset ゲーム初期化
func (g *FortyFives) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.teamScores = [FortyFivesTeamCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する。
func (g *FortyFives) NextRound() {
	if g.phase != FortyFivesPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % FortyFivesPlayerCnt
	g.startRound()
}

// startRound 手札を配り、入札フェーズを開始する。
func (g *FortyFives) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.bids = [FortyFivesPlayerCnt]FortyFivesBid{}
	g.bidDone = [FortyFivesPlayerCnt]bool{}
	g.declarerIdx = -1
	g.contract = FortyFivesBidPass
	g.trumpSuit = 0
	g.roundTeamPts = [FortyFivesTeamCnt]int{}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()
	g.sortAllHands()

	g.currentPlayerIdx = (g.dealerIdx + 1) % FortyFivesPlayerCnt // forehand bids first
	g.phase = FortyFivesPhaseBid
}

// deal 各プレイヤーへ 5 枚を配る。
func (g *FortyFives) deal() {
	for i := 0; i < FortyFivesHandSize; i++ {
		for j := 0; j < FortyFivesPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % FortyFivesPlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
}

// --- Bidding ---

// IsHumanBidTurn 入札フェーズで人間の手番か。
func (g *FortyFives) IsHumanBidTurn() bool {
	return g.phase == FortyFivesPhaseBid && g.currentPlayerIdx >= 0 &&
		g.currentPlayerIdx < len(g.players) && g.players[g.currentPlayerIdx].GetIsHuman()
}

// highestBid 現在の最高入札と入札者を返す (-1=なし)。
func (g *FortyFives) highestBid() (FortyFivesBid, int) {
	best, bestIdx := FortyFivesBidPass, -1
	for i := 0; i < FortyFivesPlayerCnt; i++ {
		if g.bidDone[i] && g.bids[i] > best {
			best = g.bids[i]
			bestIdx = i
		}
	}
	return best, bestIdx
}

// PlayerBid 人間プレイヤーが入札する。
func (g *FortyFives) PlayerBid(bid FortyFivesBid) error {
	if g.phase != FortyFivesPhaseBid {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyBid(g.currentPlayerIdx, bid)
}

// CpuBid 入札フェーズで CPU が 1 件入札する。
func (g *FortyFives) CpuBid() {
	if g.phase != FortyFivesPhaseBid {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	_ = g.applyBid(idx, g.cpuChooseBid(idx))
}

// applyBid 入札を記録し、次の入札者へ進める。全員入札したら契約を確定する。
func (g *FortyFives) applyBid(idx int, bid FortyFivesBid) error {
	if bid != FortyFivesBidPass && bid != FortyFivesBidFifteen && bid != FortyFivesBidTwenty && bid != FortyFivesBidTwentyFive {
		return NewDomainError(ErrInvalidPlay, "入札値が不正です")
	}
	high, _ := g.highestBid()
	if bid != FortyFivesBidPass && bid <= high {
		return NewDomainError(ErrInvalidPlay, "現在の入札を上回る必要があります")
	}
	g.bids[idx] = bid
	g.bidDone[idx] = true
	if bid != FortyFivesBidPass {
		g.appendLog(idx, "bid", fmt.Sprintf("%s bids %d", g.playerName(idx), int(bid)), nil)
	} else {
		g.appendLog(idx, "bid", fmt.Sprintf("%s passes", g.playerName(idx)), nil)
	}
	for k := 1; k <= FortyFivesPlayerCnt; k++ {
		ni := (idx + k) % FortyFivesPlayerCnt
		if !g.bidDone[ni] {
			g.currentPlayerIdx = ni
			return nil
		}
	}
	g.resolveBidding()
	return nil
}

// resolveBidding 入札を締め、落札者・契約・切り札を確定してプレイへ移る。
func (g *FortyFives) resolveBidding() {
	bid, idx := g.highestBid()
	if idx < 0 || bid == FortyFivesBidPass {
		g.declarerIdx = -1
		g.phase = FortyFivesPhaseRoundEnd
		g.appendLog(-1, "passed_out", "all players passed; round is void", nil)
		return
	}
	g.declarerIdx = idx
	g.contract = bid
	g.trumpSuit = g.longestSuit(idx)
	g.appendLog(idx, "contract",
		fmt.Sprintf("%s (team %s) bids %d, trump %d", g.playerName(idx), fortyFivesTeamName(FortyFivesTeamOf(idx)), int(bid), g.trumpSuit), nil)
	g.leadPlayerIdx = idx // declarer leads
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = FortyFivesPhasePlay
}

// longestSuit プレイヤーが最も多く持つスートを返す。
func (g *FortyFives) longestSuit(playerIdx int) int {
	counts := map[int]int{}
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		counts[p.GetCard(i).GetDesign()]++
	}
	bestSuit, bestCnt := CardDesignSpade, -1
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if counts[suit] > bestCnt {
			bestCnt = counts[suit]
			bestSuit = suit
		}
	}
	return bestSuit
}

// cpuChooseBid CPU の入札を選ぶ。高位札と最長スートから見積もる。
func (g *FortyFives) cpuChooseBid(idx int) FortyFivesBid {
	high, _ := g.highestBid()
	suit := g.longestSuit(idx)
	cnt, highCards := 0, 0
	p := g.players[idx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == suit {
			cnt++
		}
		if c.GetValue() == 1 || c.GetValue() == 5 || c.GetValue() >= 11 {
			highCards++
		}
	}
	want := FortyFivesBidPass
	switch {
	case cnt >= 4 && highCards >= 3:
		want = FortyFivesBidTwentyFive
	case cnt >= 3 && highCards >= 2:
		want = FortyFivesBidTwenty
	case cnt >= 3 || highCards >= 2:
		want = FortyFivesBidFifteen
	}
	if want > high {
		return want
	}
	return FortyFivesBidPass
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *FortyFives) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != FortyFivesPhasePlay {
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
func (g *FortyFives) CpuPlay() {
	if g.gameEndFlag || g.phase != FortyFivesPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	cardIdx := g.cpuSelectPlayCard(idx)
	played := g.players[idx].RemoveCard(cardIdx)
	g.playCard(idx, played)
}

// playCard カードをプレイする共通処理。
func (g *FortyFives) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", g.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == FortyFivesPlayerCnt {
		g.phase = FortyFivesPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % FortyFivesPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定し、得点を加算する。
func (g *FortyFives) ResolveTrick() {
	if g.phase != FortyFivesPhaseTrickEnd || len(g.currentTrick) != FortyFivesPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
	}
	g.players[winnerIdx].AddTrick(trickCards)
	g.players[winnerIdx].IncRoundTricks()
	g.roundTeamPts[FortyFivesTeamOf(winnerIdx)] += FortyFivesPointsPerTrick
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (+%d)", g.playerName(winnerIdx), g.trickNumber, FortyFivesPointsPerTrick), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= FortyFivesTrickCount {
		g.phase = FortyFivesPhaseRoundEnd
	} else {
		g.phase = FortyFivesPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *FortyFives) NextTrick() {
	if g.phase != FortyFivesPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = FortyFivesPhasePlay
}

// ScoreRound ラウンド結果を判定し、チーム点を加減してマッチ終了を判定する。
func (g *FortyFives) ScoreRound() {
	if g.phase != FortyFivesPhaseRoundEnd {
		return
	}
	if g.declarerIdx >= 0 {
		bidTeam := FortyFivesTeamOf(g.declarerIdx)
		otherTeam := 1 - bidTeam
		bidVal := int(g.contract)
		jink := g.contract == FortyFivesBidTwentyFive && g.teamTricks(bidTeam) == FortyFivesTrickCount

		// 非落札チームは取得点を常に加算。
		g.teamScores[otherTeam] += g.roundTeamPts[otherTeam]
		// 落札チーム: 宣言点以上なら加点、未達は宣言点を失点。
		if g.roundTeamPts[bidTeam] >= bidVal {
			g.teamScores[bidTeam] += g.roundTeamPts[bidTeam]
		} else {
			g.teamScores[bidTeam] -= bidVal
		}
		g.appendLog(-1, "round_score",
			fmt.Sprintf("round %d: team %s bid %d, got %d; team %s got %d",
				g.roundNumber, fortyFivesTeamName(bidTeam), bidVal, g.roundTeamPts[bidTeam],
				fortyFivesTeamName(otherTeam), g.roundTeamPts[otherTeam]), nil)

		if jink {
			// Jink: 25 を宣言して全トリック → 即勝利。
			g.gameEndFlag = true
			g.winnerTeam = bidTeam
			g.phase = FortyFivesPhaseGameEnd
			g.appendLog(-1, "jink", fmt.Sprintf("Jink! team %s sweeps and wins", fortyFivesTeamName(bidTeam)), nil)
			return
		}
		g.checkGameEnd()
	}
}

// teamTricks チームが現ラウンドで取ったトリック数の合計。
func (g *FortyFives) teamTricks(team int) int {
	n := 0
	for i := 0; i < FortyFivesPlayerCnt; i++ {
		if FortyFivesTeamOf(i) == team {
			n += g.players[i].GetRoundTricks()
		}
	}
	return n
}

// checkGameEnd 目標点到達でマッチ終了を判定する。
func (g *FortyFives) checkGameEnd() {
	leader, best := -1, -1<<30
	for t := 0; t < FortyFivesTeamCnt; t++ {
		if g.teamScores[t] > best {
			best = g.teamScores[t]
			leader = t
		}
	}
	if best >= g.config.TargetPoints && leader >= 0 {
		g.gameEndFlag = true
		g.winnerTeam = leader
		g.phase = FortyFivesPhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("Team %s wins the match!", fortyFivesTeamName(leader)), nil)
	}
}

// --- Trick / play helpers (Spoil Five fixed-trump rank + Reneging) ---

// isTopTrump 上位 3 枚の切り札 (切り札 5・切り札 J・♥A) か。
func (g *FortyFives) isTopTrump(card *Card) bool {
	d, v := card.GetDesign(), card.GetValue()
	if d == g.trumpSuit && (v == 5 || v == 11) {
		return true
	}
	return d == CardDesignHeart && v == 1
}

// isTrumpCard 切り札扱いの札か (切り札スート、または ♥A)。
func (g *FortyFives) isTrumpCard(card *Card) bool {
	return card.GetDesign() == g.trumpSuit || (card.GetDesign() == CardDesignHeart && card.GetValue() == 1)
}

// validatePlay マストフォロー + Reneging (上位切り札はフォロー免除) を検証する。
func (g *FortyFives) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadCard := g.currentTrick[0].Card
	leadIsTrump := g.isTrumpCard(leadCard)
	leadSuit := leadCard.GetDesign()

	hasFollowable := false
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if leadIsTrump {
			if g.isTrumpCard(c) && !g.isTopTrump(c) {
				hasFollowable = true
				break
			}
		} else {
			if c.GetDesign() == leadSuit && !g.isTrumpCard(c) {
				hasFollowable = true
				break
			}
		}
	}
	if !hasFollowable {
		return nil
	}
	if leadIsTrump {
		if !g.isTrumpCard(card) {
			return NewDomainError(ErrInvalidPlay, "切り札に従ってください")
		}
	} else {
		if card.GetDesign() != leadSuit || g.isTrumpCard(card) {
			return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
		}
	}
	return nil
}

// trickWinner トリックの勝者を決定する。
func (g *FortyFives) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	bestRank := g.fortyFivesRank(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		if !g.isTrumpCard(tc.Card) && tc.Card.GetDesign() != leadSuit {
			continue
		}
		if r := g.fortyFivesRank(tc.Card); r > bestRank {
			bestRank = r
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// fortyFivesRank Spoil Five 系の固定ランクを返す (高いほど強い)。
func (g *FortyFives) fortyFivesRank(card *Card) int {
	d, v := card.GetDesign(), card.GetValue()
	switch {
	case d == g.trumpSuit && v == 5:
		return 1000
	case d == g.trumpSuit && v == 11:
		return 999
	case d == CardDesignHeart && v == 1:
		return 998
	case d == g.trumpSuit && v == 1:
		return 997
	case d == g.trumpSuit && v == 13:
		return 996
	case d == g.trumpSuit && v == 12:
		return 995
	case d == g.trumpSuit:
		return 900 + fortyFivesPip(v)
	default:
		return fortyFivesPip(v)
	}
}

// fortyFivesPip 数札の相対強さ (A 高, 10-high 簡略)。
func fortyFivesPip(v int) int {
	if v == 1 {
		return 14
	}
	return v
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *FortyFives) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *FortyFives) sortAllHands() {
	for _, p := range g.players {
		g.fortyFivesSortHand(p)
	}
}

// fortyFivesSortHand 手札を強さ順 (降順) にソートする。
func (g *FortyFives) fortyFivesSortHand(p *FortyFivesPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		return g.fortyFivesRank(cards[i]) > g.fortyFivesRank(cards[j])
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playerName プレイヤー名を返す。
func (g *FortyFives) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *FortyFives) indexOfPlayerInTrick(playerIdx int) int {
	for i, tc := range g.currentTrick {
		if tc.PlayerIdx == playerIdx {
			return i
		}
	}
	return -1
}

// trickTopRank 現在のトリック勝者の札のランクを返す。見つからない場合は極小値。
func (g *FortyFives) trickTopRank(winnerIdx int) int {
	idx := g.indexOfPlayerInTrick(winnerIdx)
	if idx < 0 {
		return -1 << 30
	}
	return g.fortyFivesRank(g.currentTrick[idx].Card)
}

// findHumanIdx 人間プレイヤーのインデックス (-1=なし)。
func (g *FortyFives) findHumanIdx() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// appendLog 棋譜にエントリを追加する。
func (g *FortyFives) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: len(g.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- CPU AI ---

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *FortyFives) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == FortyFivesCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart トリックとチーム関係を意識した戦略プレイ。
func (g *FortyFives) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		return g.maxBy(player, valid, func(c *Card) int { return g.fortyFivesRank(c) })
	}
	winnerIdx := g.trickWinner()
	topRank := g.trickTopRank(winnerIdx)
	partnerWinning := FortyFivesTeamOf(winnerIdx) == FortyFivesTeamOf(playerIdx) && winnerIdx != playerIdx
	winners := fortyFivesFilter(valid, func(idx int) bool { return g.fortyFivesRank(player.GetCard(idx)) > topRank })
	if partnerWinning {
		// 味方が勝っている: 最弱札を温存・捨てる。
		return g.minBy(player, valid, func(c *Card) int { return g.fortyFivesRank(c) })
	}
	if len(winners) > 0 {
		return g.minBy(player, winners, func(c *Card) int { return g.fortyFivesRank(c) })
	}
	return g.minBy(player, valid, func(c *Card) int { return g.fortyFivesRank(c) })
}

// minBy score が最小となるインデックスを返す。
func (g *FortyFives) minBy(player *FortyFivesPlayer, indices []int, score func(*Card) int) int {
	best := indices[0]
	bestScore := score(player.GetCard(best))
	for _, idx := range indices[1:] {
		if s := score(player.GetCard(idx)); s < bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// maxBy score が最大となるインデックスを返す。
func (g *FortyFives) maxBy(player *FortyFivesPlayer, indices []int, score func(*Card) int) int {
	best := indices[0]
	bestScore := score(player.GetCard(best))
	for _, idx := range indices[1:] {
		if s := score(player.GetCard(idx)); s > bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// fortyFivesFilter 述語を満たすインデックスを抽出する。
func fortyFivesFilter(indices []int, pred func(int) bool) []int {
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
func (g *FortyFives) GetHint() *FortyFivesHint {
	human := g.findHumanIdx()
	if human < 0 || g.phase != FortyFivesPhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.getValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuPlaySmart(human, valid)
	return &FortyFivesHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *FortyFives) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_high"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	winnerIdx := g.trickWinner()
	if g.fortyFivesRank(card) > g.trickTopRank(winnerIdx) &&
		(g.isTrumpCard(card) || card.GetDesign() == g.currentTrick[0].Card.GetDesign()) {
		return "take_trick"
	}
	return "discard_low"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *FortyFives) GetPhase() FortyFivesPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *FortyFives) SetPhase(phase FortyFivesPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *FortyFives) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *FortyFives) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *FortyFives) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *FortyFives) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *FortyFives) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *FortyFives) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *FortyFives) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *FortyFives) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *FortyFives) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *FortyFives) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *FortyFives) GetDealerIdx() int { return g.dealerIdx }

// GetDeclarerIdx 落札者インデックス取得 (-1=未確定)
func (g *FortyFives) GetDeclarerIdx() int { return g.declarerIdx }

// SetDeclarerIdx 落札者インデックス設定 (テスト用)
func (g *FortyFives) SetDeclarerIdx(idx int) { g.declarerIdx = idx }

// GetContract 確定した契約を取得
func (g *FortyFives) GetContract() FortyFivesBid { return g.contract }

// SetContract 契約設定 (テスト用)
func (g *FortyFives) SetContract(b FortyFivesBid) { g.contract = b }

// GetBids 各プレイヤーの入札を取得
func (g *FortyFives) GetBids() [FortyFivesPlayerCnt]FortyFivesBid { return g.bids }

// GetBidDone は各プレイヤーが入札を済ませたか（true=済み）を返す。未入札と
// 「パス（bid=0）」を区別できるよう、プレゼンター向けに公開する。
func (g *FortyFives) GetBidDone() [FortyFivesPlayerCnt]bool { return g.bidDone }

// GetTrumpSuit 切り札スート取得 (0=なし)
func (g *FortyFives) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *FortyFives) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetTeamScores チーム別累積点取得
func (g *FortyFives) GetTeamScores() [FortyFivesTeamCnt]int { return g.teamScores }

// SetTeamScores チーム別累積点設定 (テスト用)
func (g *FortyFives) SetTeamScores(s [FortyFivesTeamCnt]int) { g.teamScores = s }

// GetRoundTeamPoints 現ラウンドのチーム別得点取得
func (g *FortyFives) GetRoundTeamPoints() [FortyFivesTeamCnt]int { return g.roundTeamPts }

// SetRoundTeamPoints 現ラウンドのチーム別得点設定 (テスト用)
func (g *FortyFives) SetRoundTeamPoints(s [FortyFivesTeamCnt]int) { g.roundTeamPts = s }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *FortyFives) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1=未確定)
func (g *FortyFives) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (g *FortyFives) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *FortyFives) GetPlayer(i int) *FortyFivesPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在の手番が人間か (プレイフェーズ)。
func (g *FortyFives) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *FortyFives) GetConfig() FortyFivesConfig { return g.config }

// SetConfig 設定変更
func (g *FortyFives) SetConfig(cfg FortyFivesConfig) { g.config = cfg }

// GetActionLog 棋譜取得
func (g *FortyFives) GetActionLog() []*ActionLogEntry { return g.actionLog }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *FortyFives) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != FortyFivesPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// fortyFivesJSON is the JSON wire format for FortyFives.
type fortyFivesJSON struct {
	TrumpCards       *TrumpCards                        `json:"tc"`
	Players          []*FortyFivesPlayer                `json:"ps"`
	Config           FortyFivesConfig                   `json:"cf"`
	Phase            FortyFivesPhase                    `json:"ph"`
	RoundNumber      int                                `json:"rn"`
	TrickNumber      int                                `json:"tn"`
	CurrentPlayerIdx int                                `json:"ci"`
	CurrentTrick     []*TrickCard                       `json:"ct"`
	LeadPlayerIdx    int                                `json:"li"`
	DealerIdx        int                                `json:"di"`
	Bids             [FortyFivesPlayerCnt]FortyFivesBid `json:"bd"`
	BidDone          [FortyFivesPlayerCnt]bool          `json:"bf"`
	DeclarerIdx      int                                `json:"dc"`
	Contract         FortyFivesBid                      `json:"co"`
	TrumpSuit        int                                `json:"ts"`
	TeamScores       [FortyFivesTeamCnt]int             `json:"sc"`
	RoundTeamPts     [FortyFivesTeamCnt]int             `json:"rp"`
	GameEndFlag      bool                               `json:"ge"`
	WinnerTeam       int                                `json:"wt"`
	ActionLog        []*ActionLogEntry                  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *FortyFives) MarshalJSON() ([]byte, error) {
	return json.Marshal(fortyFivesJSON{
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
		TeamScores:       g.teamScores,
		RoundTeamPts:     g.roundTeamPts,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// fortyFivesMaxSliceLen caps slice sizes during deserialisation.
const fortyFivesMaxSliceLen = 5000

// errFortyFivesOversized is the single sentinel error for oversized input arrays.
var errFortyFivesOversized = errors.New("fortyfives: input array exceeds maximum allowed size")

// errFortyFivesInvalidPlayers is returned when restored state lacks exactly FortyFivesPlayerCnt players.
var errFortyFivesInvalidPlayers = errors.New("fortyfives: invalid player count")

// errFortyFivesInvalidTrick is returned when a restored trick card is nil/out of range.
var errFortyFivesInvalidTrick = errors.New("fortyfives: invalid trick card")

// errFortyFivesInvalidState is returned when a restored index/state field is out of range.
var errFortyFivesInvalidState = errors.New("fortyfives: invalid state values in json")

// UnmarshalJSON implements json.Unmarshaler.
func (g *FortyFives) UnmarshalJSON(data []byte) error {
	var j fortyFivesJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > fortyFivesMaxSliceLen || len(j.CurrentTrick) > fortyFivesMaxSliceLen ||
		len(j.ActionLog) > fortyFivesMaxSliceLen {
		return errFortyFivesOversized
	}
	if len(j.Players) != FortyFivesPlayerCnt {
		return errFortyFivesInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errFortyFivesInvalidPlayers
		}
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= FortyFivesPlayerCnt ||
		j.LeadPlayerIdx < 0 || j.LeadPlayerIdx >= FortyFivesPlayerCnt ||
		j.DealerIdx < 0 || j.DealerIdx >= FortyFivesPlayerCnt ||
		j.DeclarerIdx < -1 || j.DeclarerIdx >= FortyFivesPlayerCnt ||
		j.WinnerTeam < -1 || j.WinnerTeam >= FortyFivesTeamCnt ||
		j.TrumpSuit < 0 || j.TrumpSuit > 4 ||
		j.RoundNumber < 1 ||
		j.TrickNumber < 1 || j.TrickNumber > FortyFivesTrickCount ||
		j.Phase < FortyFivesPhaseBid || j.Phase > FortyFivesPhaseGameEnd {
		return errFortyFivesInvalidState
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= FortyFivesPlayerCnt {
			return errFortyFivesInvalidTrick
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
	g.teamScores = j.TeamScores
	g.roundTeamPts = j.RoundTeamPts
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	return nil
}
