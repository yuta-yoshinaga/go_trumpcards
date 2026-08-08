//go:build !js || !wasm || classic

// Package domain ナップ (Nap / Napoleon) のドメインモデル。
//
// Nap はイギリス発祥の 5 枚手札のギャンブル系トリックテイキング。52 枚デッキを 5 枚ずつ
// 配り、各プレイヤーが取れると思うトリック数 (2〜4、または 5=Nap) を時計回りに 1 回だけ
// 入札する。最高入札者が宣言者となり、手札の最長スートを切り札として 5 トリックを戦う
// (本実装では切り札を自動選択)。宣言トリック数以上取れれば入札値のチップを獲得、失敗は
// 同額を相手が獲得する。Nap (5 宣言) は成功 +10 / 失敗時は各相手 +5 と高低差が大きい。
// 累積チップが目標 (既定 20) に達したプレイヤーが勝利する。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// NapPlayerCnt プレイヤー数 (人間 1 + CPU 3)
const NapPlayerCnt = 4

// NapHandSize 各プレイヤーの手札枚数
const NapHandSize = 5

// NapTrickCount 1 ラウンドのトリック数
const NapTrickCount = 5

// NapBid 入札種別 (値はそのまま目標トリック数; Pass=0)
type NapBid int

// Nap の入札定数
const (
	// NapBidPass パス
	NapBidPass NapBid = 0
	// NapBidTwo 2 トリック宣言
	NapBidTwo NapBid = 2
	// NapBidThree 3 トリック宣言
	NapBidThree NapBid = 3
	// NapBidFour 4 トリック宣言
	NapBidFour NapBid = 4
	// NapBidNap 5 トリック宣言 (Nap)
	NapBidNap NapBid = 5
)

// napBidTarget 契約の目標トリック数を返す。
func napBidTarget(b NapBid) int { return int(b) }

// NapPhase ゲームフェーズ
type NapPhase int

// Nap のフェーズ定数
const (
	// NapPhaseBid 入札フェーズ
	NapPhaseBid NapPhase = 0
	// NapPhasePlay トリックプレイフェーズ
	NapPhasePlay NapPhase = 1
	// NapPhaseTrickEnd トリック終了フェーズ
	NapPhaseTrickEnd NapPhase = 2
	// NapPhaseRoundEnd ラウンド終了フェーズ
	NapPhaseRoundEnd NapPhase = 3
	// NapPhaseGameEnd ゲーム終了フェーズ
	NapPhaseGameEnd NapPhase = 4
)

// NapHint ヒント情報
type NapHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// Nap ナップのゲームクラス
type Nap struct {
	trumpCards       *TrumpCards
	players          []*NapPlayer
	config           NapConfig
	phase            NapPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	bids             [NapPlayerCnt]NapBid
	bidDone          [NapPlayerCnt]bool
	declarerIdx      int // -1=未確定/全パス
	contract         NapBid
	trumpSuit        int
	playerScores     [NapPlayerCnt]int // 累積チップ
	roundTricks      [NapPlayerCnt]int
	gameEndFlag      bool
	winnerPlayer     int // -1=未確定
	actionLogBase
}

// NewNap コンストラクタ
func NewNap(trumpCards *TrumpCards, players []*NapPlayer, config NapConfig) *Nap {
	return &Nap{trumpCards: trumpCards, players: players, config: config, winnerPlayer: -1, declarerIdx: -1}
}

// NewDefaultNap 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultNap() *Nap {
	players := make([]*NapPlayer, NapPlayerCnt)
	players[0] = NewNapPlayer(true)
	for i := 1; i < NapPlayerCnt; i++ {
		players[i] = NewNapPlayer(false)
	}
	return NewNap(NewTrumpCards(0), players, DefaultNapConfig())
}

// Reset ゲーム初期化
func (g *Nap) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.playerScores = [NapPlayerCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する。
func (g *Nap) NextRound() {
	if g.phase != NapPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % NapPlayerCnt
	g.startRound()
}

// startRound 手札を配り、入札フェーズを開始する。
func (g *Nap) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.bids = [NapPlayerCnt]NapBid{}
	g.bidDone = [NapPlayerCnt]bool{}
	g.declarerIdx = -1
	g.contract = NapBidPass
	g.trumpSuit = 0
	g.roundTricks = [NapPlayerCnt]int{}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()
	g.sortAllHands()

	g.currentPlayerIdx = (g.dealerIdx + 1) % NapPlayerCnt // forehand bids first
	g.phase = NapPhaseBid
}

// deal 各プレイヤーへ 5 枚を配る。
func (g *Nap) deal() {
	for i := 0; i < NapHandSize; i++ {
		for j := 0; j < NapPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % NapPlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
}

// --- Bidding ---

// IsHumanBidTurn 入札フェーズで人間の手番か。
func (g *Nap) IsHumanBidTurn() bool {
	return g.phase == NapPhaseBid && g.currentPlayerIdx >= 0 &&
		g.currentPlayerIdx < len(g.players) && g.players[g.currentPlayerIdx].GetIsHuman()
}

// highestBid 現在の最高入札と入札者を返す (-1=なし)。
func (g *Nap) highestBid() (NapBid, int) {
	best, bestIdx := NapBidPass, -1
	for i := 0; i < NapPlayerCnt; i++ {
		if g.bidDone[i] && g.bids[i] > best {
			best = g.bids[i]
			bestIdx = i
		}
	}
	return best, bestIdx
}

// PlayerBid 人間プレイヤーが入札する。
func (g *Nap) PlayerBid(bid NapBid) error {
	if g.phase != NapPhaseBid {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyBid(g.currentPlayerIdx, bid)
}

// CpuBid 入札フェーズで CPU が 1 件入札する。
func (g *Nap) CpuBid() {
	if g.phase != NapPhaseBid {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	_ = g.applyBid(idx, g.cpuChooseBid(idx))
}

// applyBid 入札を記録し、次の入札者へ進める。全員入札したら契約を確定する。
func (g *Nap) applyBid(idx int, bid NapBid) error {
	if bid != NapBidPass && (bid < NapBidTwo || bid > NapBidNap) {
		return NewDomainError(ErrInvalidPlay, "入札値が不正です")
	}
	high, _ := g.highestBid()
	if bid != NapBidPass && bid <= high {
		return NewDomainError(ErrInvalidPlay, "現在の入札を上回る必要があります")
	}
	g.bids[idx] = bid
	g.bidDone[idx] = true
	if bid != NapBidPass {
		g.appendLog(idx, "bid", fmt.Sprintf("%s bids %s", g.playerName(idx), napBidName(bid)), nil)
	} else {
		g.appendLog(idx, "bid", fmt.Sprintf("%s passes", g.playerName(idx)), nil)
	}
	for k := 1; k <= NapPlayerCnt; k++ {
		ni := (idx + k) % NapPlayerCnt
		if !g.bidDone[ni] {
			g.currentPlayerIdx = ni
			return nil
		}
	}
	g.resolveBidding()
	return nil
}

// resolveBidding 入札を締め、宣言者・契約・切り札を確定してプレイへ移る。
func (g *Nap) resolveBidding() {
	bid, idx := g.highestBid()
	if idx < 0 || bid == NapBidPass {
		g.declarerIdx = -1
		g.phase = NapPhaseRoundEnd
		g.appendLog(-1, "passed_out", "all players passed; round is void", nil)
		return
	}
	g.declarerIdx = idx
	g.contract = bid
	g.trumpSuit = g.longestSuit(idx)
	g.appendLog(idx, "contract",
		fmt.Sprintf("%s declares %s (trump %d)", g.playerName(idx), napBidName(bid), g.trumpSuit), nil)
	g.leadPlayerIdx = idx // declarer leads in Nap
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = NapPhasePlay
}

// longestSuit プレイヤーが最も多く持つスートを返す。
func (g *Nap) longestSuit(playerIdx int) int {
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

// cpuChooseBid CPU の入札を選ぶ。高位札と最長スートの長さから目標トリック数を見積もる。
func (g *Nap) cpuChooseBid(idx int) NapBid {
	high, _ := g.highestBid()
	suit := g.longestSuit(idx)
	cnt, highCards := 0, 0
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
	est := highCards + cnt/2 // 大雑把な取得見込みトリック数
	want := NapBidPass
	switch {
	case est >= 5:
		want = NapBidNap
	case est == 4:
		want = NapBidFour
	case est == 3:
		want = NapBidThree
	case est == 2:
		want = NapBidTwo
	}
	if want > high {
		return want
	}
	return NapBidPass
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Nap) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != NapPhasePlay {
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
func (g *Nap) CpuPlay() {
	if g.gameEndFlag || g.phase != NapPhasePlay {
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
func (g *Nap) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", g.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == NapPlayerCnt {
		g.phase = NapPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % NapPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する。
func (g *Nap) ResolveTrick() {
	if g.phase != NapPhaseTrickEnd || len(g.currentTrick) != NapPlayerCnt {
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
		fmt.Sprintf("%s wins trick %d", g.playerName(winnerIdx), g.trickNumber), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= NapTrickCount {
		g.phase = NapPhaseRoundEnd
	} else {
		g.phase = NapPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Nap) NextTrick() {
	if g.phase != NapPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = NapPhasePlay
}

// ScoreRound 契約の達成可否を判定し、チップを加算してマッチ終了を判定する。
func (g *Nap) ScoreRound() {
	if g.phase != NapPhaseRoundEnd {
		return
	}
	if g.declarerIdx >= 0 {
		won := g.contractMade()
		if won {
			g.playerScores[g.declarerIdx] += g.makeValue()
		} else {
			fv := g.failValue()
			for i := 0; i < NapPlayerCnt; i++ {
				if i != g.declarerIdx {
					g.playerScores[i] += fv
				}
			}
		}
		g.appendLog(-1, "round_score",
			fmt.Sprintf("round %d: %s %s (%d/%d tricks)",
				g.roundNumber, napBidName(g.contract),
				map[bool]string{true: "made", false: "failed"}[won],
				g.roundTricks[g.declarerIdx], napBidTarget(g.contract)), nil)
		g.checkGameEnd()
	}
}

// makeValue 契約達成時に宣言者が得るチップ数。
func (g *Nap) makeValue() int {
	if g.contract == NapBidNap {
		return 10
	}
	return napBidTarget(g.contract)
}

// failValue 契約失敗時に各相手が得るチップ数。
func (g *Nap) failValue() int {
	if g.contract == NapBidNap {
		return 5
	}
	return napBidTarget(g.contract)
}

// contractMade 宣言者が契約を達成したか。
func (g *Nap) contractMade() bool {
	return g.roundTricks[g.declarerIdx] >= napBidTarget(g.contract)
}

// checkGameEnd 目標チップ到達でマッチ終了を判定する。
func (g *Nap) checkGameEnd() {
	leader, best := -1, -1
	for i := 0; i < NapPlayerCnt; i++ {
		if g.playerScores[i] > best {
			best = g.playerScores[i]
			leader = i
		}
	}
	if best >= g.config.TargetPoints && leader >= 0 {
		g.gameEndFlag = true
		g.winnerPlayer = leader
		g.phase = NapPhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the match!", g.playerName(leader)), nil)
	}
}

// --- Trick / play helpers ---

// validatePlay マストフォローを検証する。
func (g *Nap) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if g.playerHasSuit(playerIdx, leadSuit) && card.GetDesign() != leadSuit {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasSuit プレイヤーが指定スートのカードを持っているか。
func (g *Nap) playerHasSuit(playerIdx, design int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == design {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する。切り札があれば最強切り札、なければ
// リードスートの最強札が勝つ。
func (g *Nap) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	bestRank := g.napRank(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		if tc.Card.GetDesign() != g.trumpSuit && tc.Card.GetDesign() != leadSuit {
			continue
		}
		if r := g.napRank(tc.Card); r > bestRank {
			bestRank = r
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// napRank トリック比較用ランク。切り札は非切り札より常に強い。A 高。
func (g *Nap) napRank(card *Card) int {
	base := napCardStrength(card.GetValue())
	if g.trumpSuit != 0 && card.GetDesign() == g.trumpSuit {
		return 100 + base
	}
	return base
}

// napCardStrength カード強度 (A 高: A>K>Q>J>10>...>2)。
func napCardStrength(value int) int {
	if value == 1 {
		return 14
	}
	return value
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Nap) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Nap) sortAllHands() {
	for _, p := range g.players {
		napSortHand(p)
	}
}

// napSortHand 手札をスート→強さ順にソートする。
func napSortHand(p *NapPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return napCardStrength(cards[i].GetValue()) > napCardStrength(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playerName プレイヤー名を返す。
func (g *Nap) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *Nap) indexOfPlayerInTrick(playerIdx int) int {
	for i, tc := range g.currentTrick {
		if tc.PlayerIdx == playerIdx {
			return i
		}
	}
	return -1
}

// trickTopRank 現在のトリック勝者の札のランクを返す。見つからない場合は極小値。
func (g *Nap) trickTopRank(winnerIdx int) int {
	idx := g.indexOfPlayerInTrick(winnerIdx)
	if idx < 0 {
		return -1 << 30
	}
	return g.napRank(g.currentTrick[idx].Card)
}

// findHumanIdx 人間プレイヤーのインデックス (-1=なし)。
func (g *Nap) findHumanIdx() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// --- CPU AI ---

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Nap) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == NapCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 宣言者かどうかとトリック状況を意識した戦略プレイ。
func (g *Nap) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	isDeclarer := playerIdx == g.declarerIdx
	if len(g.currentTrick) == 0 {
		if isDeclarer {
			return pickHighest(player, valid, func(c *Card) int { return g.napRank(c) })
		}
		return pickLowest(player, valid, func(c *Card) int { return g.napRank(c) })
	}
	winnerIdx := g.trickWinner()
	topRank := g.trickTopRank(winnerIdx)
	winners := napFilter(valid, func(idx int) bool { return g.napRank(player.GetCard(idx)) > topRank })
	// 宣言者は勝ちたい; 防御は宣言者が勝っているトリックを奪いたい。
	declarerWinning := winnerIdx == g.declarerIdx
	wantWin := isDeclarer || !declarerWinning
	if wantWin && len(winners) > 0 {
		return pickLowest(player, winners, func(c *Card) int { return g.napRank(c) })
	}
	return pickLowest(player, valid, func(c *Card) int { return g.napRank(c) })
}

// napFilter 述語を満たすインデックスを抽出する。
func napFilter(indices []int, pred func(int) bool) []int {
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
func (g *Nap) GetHint() *NapHint {
	human := g.findHumanIdx()
	if human < 0 || g.phase != NapPhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.getValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuPlaySmart(human, valid)
	return &NapHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Nap) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_high"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && card.GetDesign() != g.trumpSuit {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	if g.napRank(card) > g.trickTopRank(winnerIdx) {
		return "follow_win"
	}
	return "follow_duck"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Nap) GetPhase() NapPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Nap) SetPhase(phase NapPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Nap) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Nap) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *Nap) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Nap) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Nap) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Nap) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Nap) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Nap) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Nap) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Nap) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Nap) GetDealerIdx() int { return g.dealerIdx }

// NapDeclarerProgress は宣言者の契約達成状況。
type NapDeclarerProgress struct {
	// Won は宣言者がこれまでに取ったトリック数。
	Won int
	// Needed は契約に必要なトリック数。
	Needed int
	// Remaining は残りのトリック数。
	Remaining int
	// Unreachable はもう契約に届かないことが確定したか。
	Unreachable bool
}

// GetDeclarerProgress は宣言者の契約達成状況を返す。宣言者が決まっていない、
// またはプレイ中/トリック終了以外のフェーズでは nil。
//
// **CUI は宣言者が何トリック取ったかを一切知らせていなかった (#4763)。**
// Web は nap-declarer-progress で常時出している。CLI プレイヤーは自分で
// トリック数を数えるしかなかった。
//
// Nap は1ラウンド5トリックで、契約値 (2/3/4/5) がそのまま必要トリック数。
func (g *Nap) GetDeclarerProgress() *NapDeclarerProgress {
	if g.phase != NapPhasePlay && g.phase != NapPhaseTrickEnd {
		return nil
	}
	if g.declarerIdx < 0 || g.declarerIdx >= len(g.players) {
		return nil
	}
	played := 0
	for _, p := range g.players {
		played += p.GetTrickCount()
	}
	remaining := NapTrickCount - played
	if remaining < 0 {
		remaining = 0
	}
	won := g.players[g.declarerIdx].GetTrickCount()
	needed := int(g.contract)
	return &NapDeclarerProgress{
		Won: won, Needed: needed, Remaining: remaining,
		// **「もう届かない」は残りを全部取っても足りないときだけ。**早すぎる
		// 判定は、まだ勝てるラウンドを投げさせる。
		Unreachable: needed-won > remaining,
	}
}

// GetDeclarerIdx 宣言者インデックス取得 (-1=未確定)
func (g *Nap) GetDeclarerIdx() int { return g.declarerIdx }

// SetDeclarerIdx 宣言者インデックス設定 (テスト用)
func (g *Nap) SetDeclarerIdx(idx int) { g.declarerIdx = idx }

// GetContract 確定した契約を取得
func (g *Nap) GetContract() NapBid { return g.contract }

// SetContract 契約設定 (テスト用)
func (g *Nap) SetContract(b NapBid) { g.contract = b }

// GetBids 各プレイヤーの入札を取得
func (g *Nap) GetBids() [NapPlayerCnt]NapBid { return g.bids }

// GetTrumpSuit 切り札スート取得 (0=なし)
func (g *Nap) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *Nap) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetPlayerScores プレイヤー別累積チップ取得
func (g *Nap) GetPlayerScores() [NapPlayerCnt]int { return g.playerScores }

// SetPlayerScores プレイヤー別累積チップ設定 (テスト用)
func (g *Nap) SetPlayerScores(s [NapPlayerCnt]int) { g.playerScores = s }

// GetRoundTricks 現ラウンドの獲得トリック数取得
func (g *Nap) GetRoundTricks() [NapPlayerCnt]int { return g.roundTricks }

// SetRoundTricks 現ラウンドの獲得トリック数設定 (テスト用)
func (g *Nap) SetRoundTricks(s [NapPlayerCnt]int) { g.roundTricks = s }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Nap) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝利プレイヤー取得 (-1=未確定)
func (g *Nap) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayerCnt プレイヤー数取得
func (g *Nap) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Nap) GetPlayer(i int) *NapPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在の手番が人間か (プレイフェーズ)。
func (g *Nap) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Nap) GetConfig() NapConfig { return g.config }

// SetConfig 設定変更
func (g *Nap) SetConfig(cfg NapConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Nap) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != NapPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// napJSON is the JSON wire format for Nap.
type napJSON struct {
	TrumpCards       *TrumpCards          `json:"tc"`
	Players          []*NapPlayer         `json:"ps"`
	Config           NapConfig            `json:"cf"`
	Phase            NapPhase             `json:"ph"`
	RoundNumber      int                  `json:"rn"`
	TrickNumber      int                  `json:"tn"`
	CurrentPlayerIdx int                  `json:"ci"`
	CurrentTrick     []*TrickCard         `json:"ct"`
	LeadPlayerIdx    int                  `json:"li"`
	DealerIdx        int                  `json:"di"`
	Bids             [NapPlayerCnt]NapBid `json:"bd"`
	BidDone          [NapPlayerCnt]bool   `json:"bf"`
	DeclarerIdx      int                  `json:"dc"`
	Contract         NapBid               `json:"co"`
	TrumpSuit        int                  `json:"ts"`
	PlayerScores     [NapPlayerCnt]int    `json:"sc"`
	RoundTricks      [NapPlayerCnt]int    `json:"rt"`
	GameEndFlag      bool                 `json:"ge"`
	WinnerPlayer     int                  `json:"wp"`
	ActionLog        []*ActionLogEntry    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Nap) MarshalJSON() ([]byte, error) {
	return json.Marshal(napJSON{
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

// napMaxSliceLen caps slice sizes during deserialisation.
const napMaxSliceLen = 5000

// errNapOversized is the single sentinel error for oversized input arrays.
var errNapOversized = errors.New("nap: input array exceeds maximum allowed size")

// errNapInvalidPlayers is returned when restored state lacks exactly NapPlayerCnt players.
var errNapInvalidPlayers = errors.New("nap: invalid player count")

// errNapInvalidTrick is returned when a restored trick card or its card is nil/out of range.
var errNapInvalidTrick = errors.New("nap: invalid trick card")

// errNapInvalidState is returned when a restored index/state field is out of range.
var errNapInvalidState = errors.New("nap: invalid state values in json")

// UnmarshalJSON implements json.Unmarshaler.
func (g *Nap) UnmarshalJSON(data []byte) error {
	var j napJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > napMaxSliceLen || len(j.CurrentTrick) > napMaxSliceLen ||
		len(j.ActionLog) > napMaxSliceLen {
		return errNapOversized
	}
	if len(j.Players) != NapPlayerCnt {
		return errNapInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errNapInvalidPlayers
		}
	}
	// Bounds-check restored indices/state to prevent panics from corrupt KV data.
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= NapPlayerCnt ||
		j.LeadPlayerIdx < 0 || j.LeadPlayerIdx >= NapPlayerCnt ||
		j.DealerIdx < 0 || j.DealerIdx >= NapPlayerCnt ||
		j.DeclarerIdx < -1 || j.DeclarerIdx >= NapPlayerCnt ||
		j.WinnerPlayer < -1 || j.WinnerPlayer >= NapPlayerCnt ||
		j.TrumpSuit < 0 || j.TrumpSuit > 4 ||
		j.RoundNumber < 1 ||
		j.TrickNumber < 1 || j.TrickNumber > NapTrickCount ||
		j.Phase < NapPhaseBid || j.Phase > NapPhaseGameEnd {
		return errNapInvalidState
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= NapPlayerCnt {
			return errNapInvalidTrick
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

// napBidName 入札種別の表示名を返す。
func napBidName(b NapBid) string {
	switch b {
	case NapBidTwo:
		return "Two"
	case NapBidThree:
		return "Three"
	case NapBidFour:
		return "Four"
	case NapBidNap:
		return "Nap"
	default:
		return "Pass"
	}
}
