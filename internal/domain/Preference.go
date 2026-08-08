//go:build !js || !wasm || classic

// Package domain プレフェランス (Préférence) のドメインモデル。
//
// Préférence はロシア・オーストリアで親しまれる 3 人用の入札トリックテイキング。32 枚
// デッキを 10 枚ずつ配り (本実装ではウィドウ 2 枚は使用しない簡略化)、入札フェーズで各
// プレイヤーが目標トリック数 (6/7/8) または Misère (0 トリック・切り札なし) を 1 回だけ
// 宣言する。最高入札者が宣言者となり残り 2 人の守備側と対戦する。Six/Seven/Eight では
// 宣言者の最長スートが切り札となる (本実装では自動選択)。マストフォローで 10 トリックを
// 戦い、契約の達成可否でプレイヤー別ゲーム点を加減し、累積が目標 (既定 30) に達した
// プレイヤーが勝利する。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// PreferencePlayerCnt プレイヤー数 (人間 1 + CPU 2)
const PreferencePlayerCnt = 3

// PreferenceHandSize 各プレイヤーの手札枚数
const PreferenceHandSize = 10

// PreferenceTrickCount 1 ラウンドのトリック数
const PreferenceTrickCount = 10

// PreferenceBid 入札種別
type PreferenceBid int

// Préférence の入札定数 (序列は値の昇順; Misère は Six と Seven の間)。
const (
	// PreferenceBidPass パス
	PreferenceBidPass PreferenceBid = iota
	// PreferenceBidSix 6 トリック宣言
	PreferenceBidSix
	// PreferenceBidMisere Misère (0 トリック・切り札なし)
	PreferenceBidMisere
	// PreferenceBidSeven 7 トリック宣言
	PreferenceBidSeven
	// PreferenceBidEight 8 トリック宣言
	PreferenceBidEight
)

// preferenceBidTarget 契約の目標トリック数を返す。
func preferenceBidTarget(b PreferenceBid) int {
	switch b {
	case PreferenceBidSix:
		return 6
	case PreferenceBidSeven:
		return 7
	case PreferenceBidEight:
		return 8
	default: // Misère / Pass
		return 0
	}
}

// preferenceBidValue 契約の得点価値を返す。
func preferenceBidValue(b PreferenceBid) int {
	switch b {
	case PreferenceBidSix:
		return 6
	case PreferenceBidMisere:
		return 10
	case PreferenceBidSeven:
		return 7
	case PreferenceBidEight:
		return 8
	default:
		return 0
	}
}

// PreferencePhase ゲームフェーズ
type PreferencePhase int

// Préférence のフェーズ定数
const (
	// PreferencePhaseBid 入札フェーズ
	PreferencePhaseBid PreferencePhase = 0
	// PreferencePhasePlay トリックプレイフェーズ
	PreferencePhasePlay PreferencePhase = 1
	// PreferencePhaseTrickEnd トリック終了フェーズ
	PreferencePhaseTrickEnd PreferencePhase = 2
	// PreferencePhaseRoundEnd ラウンド終了フェーズ
	PreferencePhaseRoundEnd PreferencePhase = 3
	// PreferencePhaseGameEnd ゲーム終了フェーズ
	PreferencePhaseGameEnd PreferencePhase = 4
)

// PreferenceHint ヒント情報
type PreferenceHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// Preference プレフェランスのゲームクラス
type Preference struct {
	trumpCards       *TrumpCards
	players          []*PreferencePlayer
	config           PreferenceConfig
	phase            PreferencePhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	bids             [PreferencePlayerCnt]PreferenceBid
	bidDone          [PreferencePlayerCnt]bool
	declarerIdx      int // -1=未確定/全パス
	contract         PreferenceBid
	trumpSuit        int
	playerScores     [PreferencePlayerCnt]int
	roundTricks      [PreferencePlayerCnt]int
	gameEndFlag      bool
	winnerPlayer     int // -1=未確定
	actionLogBase
}

// NewPreference コンストラクタ
func NewPreference(trumpCards *TrumpCards, players []*PreferencePlayer, config PreferenceConfig) *Preference {
	return &Preference{trumpCards: trumpCards, players: players, config: config, winnerPlayer: -1, declarerIdx: -1}
}

// NewDefaultPreference 標準の 3 人構成 (人間 1, CPU 2) と既定設定で生成する。
func NewDefaultPreference() *Preference {
	players := make([]*PreferencePlayer, PreferencePlayerCnt)
	players[0] = NewPreferencePlayer(true)
	for i := 1; i < PreferencePlayerCnt; i++ {
		players[i] = NewPreferencePlayer(false)
	}
	return NewPreference(NewTrumpCardsBelote(), players, DefaultPreferenceConfig())
}

// Reset ゲーム初期化
func (g *Preference) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.playerScores = [PreferencePlayerCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する。
func (g *Preference) NextRound() {
	if g.phase != PreferencePhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % PreferencePlayerCnt
	g.startRound()
}

// startRound 手札を配り、入札フェーズを開始する。
func (g *Preference) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.bids = [PreferencePlayerCnt]PreferenceBid{}
	g.bidDone = [PreferencePlayerCnt]bool{}
	g.declarerIdx = -1
	g.contract = PreferenceBidPass
	g.trumpSuit = 0
	g.roundTricks = [PreferencePlayerCnt]int{}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()
	g.sortAllHands()

	g.currentPlayerIdx = (g.dealerIdx + 1) % PreferencePlayerCnt // forehand bids first
	g.phase = PreferencePhaseBid
}

// deal 各プレイヤーへ 10 枚を配る (ウィドウ 2 枚は未使用)。
func (g *Preference) deal() {
	for i := 0; i < PreferenceHandSize; i++ {
		for j := 0; j < PreferencePlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % PreferencePlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
}

// --- Bidding ---

// IsHumanBidTurn 入札フェーズで人間の手番か。
func (g *Preference) IsHumanBidTurn() bool {
	return g.phase == PreferencePhaseBid && g.currentPlayerIdx >= 0 &&
		g.currentPlayerIdx < len(g.players) && g.players[g.currentPlayerIdx].GetIsHuman()
}

// highestBid 現在の最高入札と入札者を返す (-1=なし)。
func (g *Preference) highestBid() (PreferenceBid, int) {
	best, bestIdx := PreferenceBidPass, -1
	for i := 0; i < PreferencePlayerCnt; i++ {
		if g.bidDone[i] && g.bids[i] > best {
			best = g.bids[i]
			bestIdx = i
		}
	}
	return best, bestIdx
}

// PlayerBid 人間プレイヤーが入札する。
func (g *Preference) PlayerBid(bid PreferenceBid) error {
	if g.phase != PreferencePhaseBid {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyBid(g.currentPlayerIdx, bid)
}

// CpuBid 入札フェーズで CPU が 1 件入札する。
func (g *Preference) CpuBid() {
	if g.phase != PreferencePhaseBid {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	_ = g.applyBid(idx, g.cpuChooseBid(idx))
}

// applyBid 入札を記録し、次の入札者へ進める。全員入札したら契約を確定する。
func (g *Preference) applyBid(idx int, bid PreferenceBid) error {
	if bid < PreferenceBidPass || bid > PreferenceBidEight {
		return NewDomainError(ErrInvalidPlay, "入札値が不正です")
	}
	high, _ := g.highestBid()
	if bid != PreferenceBidPass && bid <= high {
		return NewDomainError(ErrInvalidPlay, "現在の入札を上回る必要があります")
	}
	g.bids[idx] = bid
	g.bidDone[idx] = true
	if bid != PreferenceBidPass {
		g.appendLog(idx, "bid", fmt.Sprintf("%s bids %s", g.playerName(idx), preferenceBidName(bid)), nil)
	} else {
		g.appendLog(idx, "bid", fmt.Sprintf("%s passes", g.playerName(idx)), nil)
	}
	for k := 1; k <= PreferencePlayerCnt; k++ {
		ni := (idx + k) % PreferencePlayerCnt
		if !g.bidDone[ni] {
			g.currentPlayerIdx = ni
			return nil
		}
	}
	g.resolveBidding()
	return nil
}

// resolveBidding 入札を締め、宣言者・契約・切り札を確定してプレイへ移る。
func (g *Preference) resolveBidding() {
	bid, idx := g.highestBid()
	if idx < 0 || bid == PreferenceBidPass {
		g.declarerIdx = -1
		g.phase = PreferencePhaseRoundEnd
		g.appendLog(-1, "passed_out", "all players passed; round is void", nil)
		return
	}
	g.declarerIdx = idx
	g.contract = bid
	if bid == PreferenceBidMisere {
		g.trumpSuit = 0
	} else {
		g.trumpSuit = g.longestSuit(idx)
	}
	g.appendLog(idx, "contract",
		fmt.Sprintf("%s declares %s (trump %d)", g.playerName(idx), preferenceBidName(bid), g.trumpSuit), nil)
	g.leadPlayerIdx = (g.dealerIdx + 1) % PreferencePlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = PreferencePhasePlay
}

// longestSuit プレイヤーが最も多く持つスートを返す。
func (g *Preference) longestSuit(playerIdx int) int {
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

// cpuChooseBid CPU の入札を選ぶ。高位札と最長スートの長さから見積もる。
func (g *Preference) cpuChooseBid(idx int) PreferenceBid {
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
	est := highCards + cnt/2
	want := PreferenceBidPass
	switch {
	case est >= 8:
		want = PreferenceBidEight
	case est == 7:
		want = PreferenceBidSeven
	case est >= 5:
		want = PreferenceBidSix
	case highCards == 0:
		want = PreferenceBidMisere
	}
	if want > high {
		return want
	}
	return PreferenceBidPass
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Preference) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != PreferencePhasePlay {
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
func (g *Preference) CpuPlay() {
	if g.gameEndFlag || g.phase != PreferencePhasePlay {
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
func (g *Preference) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", g.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == PreferencePlayerCnt {
		g.phase = PreferencePhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % PreferencePlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する。
func (g *Preference) ResolveTrick() {
	if g.phase != PreferencePhaseTrickEnd || len(g.currentTrick) != PreferencePlayerCnt {
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
	if g.trickNumber >= PreferenceTrickCount {
		g.phase = PreferencePhaseRoundEnd
	} else {
		g.phase = PreferencePhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Preference) NextTrick() {
	if g.phase != PreferencePhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = PreferencePhasePlay
}

// ScoreRound 契約の達成可否を判定し、ゲーム点を加減してマッチ終了を判定する。
func (g *Preference) ScoreRound() {
	if g.phase != PreferencePhaseRoundEnd {
		return
	}
	if g.declarerIdx >= 0 {
		value := preferenceBidValue(g.contract)
		won := g.contractMade()
		if won {
			g.playerScores[g.declarerIdx] += value
		} else {
			for i := 0; i < PreferencePlayerCnt; i++ {
				if i != g.declarerIdx {
					g.playerScores[i] += value
				}
			}
		}
		g.appendLog(-1, "round_score",
			fmt.Sprintf("round %d: %s %s (%d/%d tricks)",
				g.roundNumber, preferenceBidName(g.contract),
				map[bool]string{true: "made", false: "failed"}[won],
				g.roundTricks[g.declarerIdx], preferenceBidTarget(g.contract)), nil)
		g.checkGameEnd()
	}
}

// contractMade 宣言者が契約を達成したか。
func (g *Preference) contractMade() bool {
	tricks := g.roundTricks[g.declarerIdx]
	if g.contract == PreferenceBidMisere {
		return tricks == 0
	}
	return tricks >= preferenceBidTarget(g.contract)
}

// checkGameEnd 目標点到達でマッチ終了を判定する。
func (g *Preference) checkGameEnd() {
	leader, best := -1, -1
	for i := 0; i < PreferencePlayerCnt; i++ {
		if g.playerScores[i] > best {
			best = g.playerScores[i]
			leader = i
		}
	}
	if best >= g.config.TargetPoints && leader >= 0 {
		g.gameEndFlag = true
		g.winnerPlayer = leader
		g.phase = PreferencePhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the match!", g.playerName(leader)), nil)
	}
}

// --- Trick / play helpers ---

// validatePlay マストフォローを検証する。
func (g *Preference) validatePlay(playerIdx int, card *Card) error {
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
func (g *Preference) playerHasSuit(playerIdx, design int) bool {
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
func (g *Preference) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	bestRank := g.preferenceRank(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		if g.trumpSuit != 0 && tc.Card.GetDesign() != g.trumpSuit && tc.Card.GetDesign() != leadSuit {
			continue
		}
		if g.trumpSuit == 0 && tc.Card.GetDesign() != leadSuit {
			continue
		}
		if r := g.preferenceRank(tc.Card); r > bestRank {
			bestRank = r
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// preferenceRank トリック比較用ランク。切り札は非切り札より常に強い。A 高。
func (g *Preference) preferenceRank(card *Card) int {
	base := preferenceCardStrength(card.GetValue())
	if g.trumpSuit != 0 && card.GetDesign() == g.trumpSuit {
		return 100 + base
	}
	return base
}

// preferenceCardStrength カード強度 (A 高: A>K>Q>J>10>9>8>7)。
func preferenceCardStrength(value int) int {
	if value == 1 {
		return 14
	}
	return value
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Preference) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Preference) sortAllHands() {
	for _, p := range g.players {
		preferenceSortHand(p)
	}
}

// preferenceSortHand 手札をスート→強さ順にソートする。
func preferenceSortHand(p *PreferencePlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return preferenceCardStrength(cards[i].GetValue()) > preferenceCardStrength(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playerName プレイヤー名を返す。
func (g *Preference) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *Preference) indexOfPlayerInTrick(playerIdx int) int {
	for i, tc := range g.currentTrick {
		if tc.PlayerIdx == playerIdx {
			return i
		}
	}
	return -1
}

// trickTopRank 現在のトリック勝者の札のランクを返す。見つからない場合は極小値。
func (g *Preference) trickTopRank(winnerIdx int) int {
	idx := g.indexOfPlayerInTrick(winnerIdx)
	if idx < 0 {
		return -1 << 30
	}
	return g.preferenceRank(g.currentTrick[idx].Card)
}

// --- CPU AI ---

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Preference) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == PreferenceCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 宣言者かどうかとトリック状況を意識した戦略プレイ。
func (g *Preference) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	isDeclarer := playerIdx == g.declarerIdx
	misere := g.contract == PreferenceBidMisere
	if isDeclarer && misere {
		return pickLowest(player, valid, func(c *Card) int { return g.preferenceRank(c) })
	}
	if len(g.currentTrick) == 0 {
		if isDeclarer {
			return pickHighest(player, valid, func(c *Card) int { return g.preferenceRank(c) })
		}
		return pickLowest(player, valid, func(c *Card) int { return g.preferenceRank(c) })
	}
	winnerIdx := g.trickWinner()
	topRank := g.trickTopRank(winnerIdx)
	winners := preferenceFilter(valid, func(idx int) bool { return g.preferenceRank(player.GetCard(idx)) > topRank })
	wantWin := isDeclarer != misere
	if wantWin && len(winners) > 0 {
		return pickLowest(player, winners, func(c *Card) int { return g.preferenceRank(c) })
	}
	return pickLowest(player, valid, func(c *Card) int { return g.preferenceRank(c) })
}

// preferenceFilter 述語を満たすインデックスを抽出する。
func preferenceFilter(indices []int, pred func(int) bool) []int {
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
func (g *Preference) GetHint() *PreferenceHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.phase != PreferencePhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.getValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuPlaySmart(human, valid)
	return &PreferenceHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Preference) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		// 宣言者 (Misère 以外) はリードで強い札を出して主導権を握る。
		if playerIdx == g.declarerIdx && g.contract != PreferenceBidMisere {
			return "lead_high"
		}
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && (g.trumpSuit == 0 || card.GetDesign() != g.trumpSuit) {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	if g.preferenceRank(card) > g.trickTopRank(winnerIdx) {
		return "follow_win"
	}
	return "follow_duck"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Preference) GetPhase() PreferencePhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Preference) SetPhase(phase PreferencePhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Preference) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Preference) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *Preference) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Preference) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Preference) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Preference) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Preference) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Preference) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Preference) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Preference) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Preference) GetDealerIdx() int { return g.dealerIdx }

// GetDeclarerIdx 宣言者インデックス取得 (-1=未確定)
func (g *Preference) GetDeclarerIdx() int { return g.declarerIdx }

// SetDeclarerIdx 宣言者インデックス設定 (テスト用)
func (g *Preference) SetDeclarerIdx(idx int) { g.declarerIdx = idx }

// GetContract 確定した契約を取得
func (g *Preference) GetContract() PreferenceBid { return g.contract }

// SetContract 契約設定 (テスト用)
func (g *Preference) SetContract(b PreferenceBid) { g.contract = b }

// GetBids 各プレイヤーの入札を取得
func (g *Preference) GetBids() [PreferencePlayerCnt]PreferenceBid { return g.bids }

// GetTrumpSuit 切り札スート取得 (0=なし)
func (g *Preference) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *Preference) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetPlayerScores プレイヤー別累積点取得
func (g *Preference) GetPlayerScores() [PreferencePlayerCnt]int { return g.playerScores }

// SetPlayerScores プレイヤー別累積点設定 (テスト用)
func (g *Preference) SetPlayerScores(s [PreferencePlayerCnt]int) { g.playerScores = s }

// GetRoundTricks 現ラウンドの獲得トリック数取得
func (g *Preference) GetRoundTricks() [PreferencePlayerCnt]int { return g.roundTricks }

// SetRoundTricks 現ラウンドの獲得トリック数設定 (テスト用)
func (g *Preference) SetRoundTricks(s [PreferencePlayerCnt]int) { g.roundTricks = s }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Preference) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝利プレイヤー取得 (-1=未確定)
func (g *Preference) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayerCnt プレイヤー数取得
func (g *Preference) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Preference) GetPlayer(i int) *PreferencePlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在の手番が人間か (プレイフェーズ)。
func (g *Preference) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Preference) GetConfig() PreferenceConfig { return g.config }

// SetConfig 設定変更
func (g *Preference) SetConfig(cfg PreferenceConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Preference) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != PreferencePhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// preferenceJSON is the JSON wire format for Preference.
type preferenceJSON struct {
	TrumpCards       *TrumpCards                        `json:"tc"`
	Players          []*PreferencePlayer                `json:"ps"`
	Config           PreferenceConfig                   `json:"cf"`
	Phase            PreferencePhase                    `json:"ph"`
	RoundNumber      int                                `json:"rn"`
	TrickNumber      int                                `json:"tn"`
	CurrentPlayerIdx int                                `json:"ci"`
	CurrentTrick     []*TrickCard                       `json:"ct"`
	LeadPlayerIdx    int                                `json:"li"`
	DealerIdx        int                                `json:"di"`
	Bids             [PreferencePlayerCnt]PreferenceBid `json:"bd"`
	BidDone          [PreferencePlayerCnt]bool          `json:"bf"`
	DeclarerIdx      int                                `json:"dc"`
	Contract         PreferenceBid                      `json:"co"`
	TrumpSuit        int                                `json:"ts"`
	PlayerScores     [PreferencePlayerCnt]int           `json:"sc"`
	RoundTricks      [PreferencePlayerCnt]int           `json:"rt"`
	GameEndFlag      bool                               `json:"ge"`
	WinnerPlayer     int                                `json:"wp"`
	ActionLog        []*ActionLogEntry                  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Preference) MarshalJSON() ([]byte, error) {
	return json.Marshal(preferenceJSON{
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

// preferenceMaxSliceLen caps slice sizes during deserialisation.
const preferenceMaxSliceLen = 5000

// errPreferenceOversized is the single sentinel error for oversized input arrays.
var errPreferenceOversized = errors.New("preference: input array exceeds maximum allowed size")

// errPreferenceInvalidPlayers is returned when restored state lacks exactly PreferencePlayerCnt players.
var errPreferenceInvalidPlayers = errors.New("preference: invalid player count")

// errPreferenceInvalidTrick is returned when a restored trick card is nil/out of range.
var errPreferenceInvalidTrick = errors.New("preference: invalid trick card")

// errPreferenceInvalidState is returned when a restored index/state field is out of range.
var errPreferenceInvalidState = errors.New("preference: invalid state values in json")

// UnmarshalJSON implements json.Unmarshaler.
func (g *Preference) UnmarshalJSON(data []byte) error {
	var j preferenceJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > preferenceMaxSliceLen || len(j.CurrentTrick) > preferenceMaxSliceLen ||
		len(j.ActionLog) > preferenceMaxSliceLen {
		return errPreferenceOversized
	}
	if len(j.Players) != PreferencePlayerCnt {
		return errPreferenceInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errPreferenceInvalidPlayers
		}
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= PreferencePlayerCnt ||
		j.LeadPlayerIdx < 0 || j.LeadPlayerIdx >= PreferencePlayerCnt ||
		j.DealerIdx < 0 || j.DealerIdx >= PreferencePlayerCnt ||
		j.DeclarerIdx < -1 || j.DeclarerIdx >= PreferencePlayerCnt ||
		j.WinnerPlayer < -1 || j.WinnerPlayer >= PreferencePlayerCnt ||
		j.TrumpSuit < 0 || j.TrumpSuit > 4 ||
		j.RoundNumber < 1 ||
		j.TrickNumber < 1 || j.TrickNumber > PreferenceTrickCount ||
		j.Phase < PreferencePhaseBid || j.Phase > PreferencePhaseGameEnd {
		return errPreferenceInvalidState
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= PreferencePlayerCnt {
			return errPreferenceInvalidTrick
		}
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsBelote()
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

// preferenceBidName 入札種別の表示名を返す。
func preferenceBidName(b PreferenceBid) string {
	switch b {
	case PreferenceBidSix:
		return "Six"
	case PreferenceBidMisere:
		return "Misère"
	case PreferenceBidSeven:
		return "Seven"
	case PreferenceBidEight:
		return "Eight"
	default:
		return "Pass"
	}
}
