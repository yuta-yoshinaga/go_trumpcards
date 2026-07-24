//go:build !js || !wasm || classic

// Package domain スポイル・ファイブ (Spoil Five / Maw) のドメインモデル。
//
// Spoil Five はアイルランド発祥の 5 人用トリックテイキング。52 枚デッキを 5 枚ずつ配り、
// 残り札の 1 枚をめくって切り札スートを決める。最大の特徴は固定最強牌と Reneging。
// 強さ順 (高→低): 切り札の 5 > 切り札の J > ♥A (常に第 3 位) > 切り札の A > 切り札 K >
// 切り札 Q > その他の切り札 > リードスートの札。上位 3 枚 (切り札 5・J・♥A) はフォロー
// 義務がなく温存できる (Reneging)。最初に 3 トリックを取ったプレイヤーが即勝者となり
// ポットを獲得する。誰も 3 トリックに届かなければ Spoil (流局) となりポットは次ラウンドへ
// 積み増される。累積得点が目標 (既定 30) に達したプレイヤーが勝利する。
//
// 簡略化: 赤スート/黒スートでのピップ順反転は実装せず一律 10-high とする。Jink (5 トリック
// 完全取り) ボーナスは未実装。切り札はめくり札で自動決定 (宣言フェーズなし)。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// SpoilFivePlayerCnt プレイヤー数 (人間 1 + CPU 4)
const SpoilFivePlayerCnt = 5

// SpoilFiveHandSize 各プレイヤーの手札枚数
const SpoilFiveHandSize = 5

// SpoilFiveTrickCount 1 ラウンドの最大トリック数
const SpoilFiveTrickCount = 5

// SpoilFiveWinTricks ラウンド勝利に必要なトリック数
const SpoilFiveWinTricks = 3

// SpoilFiveAntePerRound 1 ラウンドごとにポットへ積まれる額 (各プレイヤー 1)
const SpoilFiveAntePerRound = SpoilFivePlayerCnt

// SpoilFivePhase ゲームフェーズ
type SpoilFivePhase int

// Spoil Five のフェーズ定数
const (
	// SpoilFivePhasePlay トリックプレイフェーズ
	SpoilFivePhasePlay SpoilFivePhase = 0
	// SpoilFivePhaseTrickEnd トリック終了フェーズ
	SpoilFivePhaseTrickEnd SpoilFivePhase = 1
	// SpoilFivePhaseRoundEnd ラウンド終了フェーズ
	SpoilFivePhaseRoundEnd SpoilFivePhase = 2
	// SpoilFivePhaseGameEnd ゲーム終了フェーズ
	SpoilFivePhaseGameEnd SpoilFivePhase = 3
)

// SpoilFiveHint ヒント情報
type SpoilFiveHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// SpoilFiveTrickCard トリック中の 1 枚
type SpoilFiveTrickCard struct {
	PlayerIdx int   `json:"pi"`
	Card      *Card `json:"c"`
}

// SpoilFive スポイル・ファイブのゲームクラス
type SpoilFive struct {
	trumpCards       *TrumpCards
	players          []*SpoilFivePlayer
	config           SpoilFiveConfig
	phase            SpoilFivePhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*SpoilFiveTrickCard
	leadPlayerIdx    int
	dealerIdx        int
	trumpSuit        int
	pot              int
	roundWinnerIdx   int // 直近ラウンドの勝者 (-1=Spoil/未確定)
	gameEndFlag      bool
	winnerPlayer     int // -1=未確定
	actionLog        []*ActionLogEntry
}

// NewSpoilFive コンストラクタ
func NewSpoilFive(trumpCards *TrumpCards, players []*SpoilFivePlayer, config SpoilFiveConfig) *SpoilFive {
	return &SpoilFive{trumpCards: trumpCards, players: players, config: config, winnerPlayer: -1, roundWinnerIdx: -1}
}

// NewDefaultSpoilFive 標準の 5 人構成 (人間 1, CPU 4) と既定設定で生成する。
func NewDefaultSpoilFive() *SpoilFive {
	players := make([]*SpoilFivePlayer, SpoilFivePlayerCnt)
	players[0] = NewSpoilFivePlayer(true)
	for i := 1; i < SpoilFivePlayerCnt; i++ {
		players[i] = NewSpoilFivePlayer(false)
	}
	return NewSpoilFive(NewTrumpCards(0), players, DefaultSpoilFiveConfig())
}

// Reset ゲーム初期化
func (g *SpoilFive) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.pot = 0
	for _, p := range g.players {
		p.SetScore(0)
	}
	g.actionLog = nil
	g.leadPlayerIdx = (g.dealerIdx + 1) % SpoilFivePlayerCnt
	g.startRound()
}

// NextRound 次のラウンドを開始する。
func (g *SpoilFive) NextRound() {
	if g.phase != SpoilFivePhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % SpoilFivePlayerCnt
	// 直近の勝者がいればその左隣、いなければ (Spoil) ディーラー左隣がリード。
	g.leadPlayerIdx = (g.dealerIdx + 1) % SpoilFivePlayerCnt
	g.startRound()
}

// startRound 手札を配り、切り札をめくってプレイフェーズを開始する。
func (g *SpoilFive) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.roundWinnerIdx = -1
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()
	// 残り札の 1 枚をめくって切り札を決める。
	if up := g.trumpCards.DrawCard(); up != nil {
		g.trumpSuit = up.GetDesign()
	}
	g.sortAllHands()
	// このラウンドのアンティをポットへ。
	g.pot += SpoilFiveAntePerRound

	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = SpoilFivePhasePlay
	g.appendLog(g.leadPlayerIdx, "round_start",
		fmt.Sprintf("round %d: trump %d, pot %d, %s leads",
			g.roundNumber, g.trumpSuit, g.pot, g.playerName(g.leadPlayerIdx)), nil)
}

// deal 各プレイヤーへ 5 枚を配る。
func (g *SpoilFive) deal() {
	for i := 0; i < SpoilFiveHandSize; i++ {
		for j := 0; j < SpoilFivePlayerCnt; j++ {
			idx := (g.leadPlayerIdx + j) % SpoilFivePlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *SpoilFive) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SpoilFivePhasePlay {
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
func (g *SpoilFive) CpuPlay() {
	if g.gameEndFlag || g.phase != SpoilFivePhasePlay {
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
func (g *SpoilFive) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &SpoilFiveTrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", g.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == SpoilFivePlayerCnt {
		g.phase = SpoilFivePhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % SpoilFivePlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する。3 トリック到達で即ラウンド終了。
func (g *SpoilFive) ResolveTrick() {
	if g.phase != SpoilFivePhaseTrickEnd || len(g.currentTrick) != SpoilFivePlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
	}
	g.players[winnerIdx].AddTrick(trickCards)
	g.players[winnerIdx].IncRoundTricks()
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", g.playerName(winnerIdx), g.trickNumber), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.players[winnerIdx].GetRoundTricks() >= SpoilFiveWinTricks {
		// 3 トリック到達: 即勝利。
		g.roundWinnerIdx = winnerIdx
		g.phase = SpoilFivePhaseRoundEnd
	} else if g.trickNumber >= SpoilFiveTrickCount {
		// 5 トリック消化したが誰も 3 に届かず: Spoil (流局)。
		g.roundWinnerIdx = -1
		g.phase = SpoilFivePhaseRoundEnd
	} else {
		g.phase = SpoilFivePhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *SpoilFive) NextTrick() {
	if g.phase != SpoilFivePhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = SpoilFivePhasePlay
}

// ScoreRound ラウンド結果を処理する。勝者はポットを獲得、Spoil ならポットは持ち越し。
func (g *SpoilFive) ScoreRound() {
	if g.phase != SpoilFivePhaseRoundEnd {
		return
	}
	if g.roundWinnerIdx >= 0 {
		g.players[g.roundWinnerIdx].SetScore(g.players[g.roundWinnerIdx].GetScore() + g.pot)
		g.appendLog(g.roundWinnerIdx, "round_win",
			fmt.Sprintf("%s wins the pot of %d", g.playerName(g.roundWinnerIdx), g.pot), nil)
		g.pot = 0
		g.checkGameEnd()
	} else {
		g.appendLog(-1, "spoil", fmt.Sprintf("spoiled! pot of %d carries over", g.pot), nil)
	}
}

// checkGameEnd 目標点到達でマッチ終了を判定する。
func (g *SpoilFive) checkGameEnd() {
	leader, best := -1, -1
	for i := 0; i < SpoilFivePlayerCnt; i++ {
		if g.players[i].GetScore() > best {
			best = g.players[i].GetScore()
			leader = i
		}
	}
	if best >= g.config.TargetPoints && leader >= 0 {
		g.gameEndFlag = true
		g.winnerPlayer = leader
		g.phase = SpoilFivePhaseGameEnd
		g.appendLog(leader, "game_end", fmt.Sprintf("%s wins the match!", g.playerName(leader)), nil)
	}
}

// --- Trick / play helpers ---

// isTopTrump 上位 3 枚の切り札 (切り札 5・切り札 J・♥A) か。
func (g *SpoilFive) isTopTrump(card *Card) bool {
	d, v := card.GetDesign(), card.GetValue()
	if d == g.trumpSuit && (v == 5 || v == 11) {
		return true
	}
	return d == CardDesignHeart && v == 1 // ♥A は常に上位切り札
}

// isTrumpCard 切り札扱いの札か (切り札スート、または ♥A)。
func (g *SpoilFive) isTrumpCard(card *Card) bool {
	return card.GetDesign() == g.trumpSuit || (card.GetDesign() == CardDesignHeart && card.GetValue() == 1)
}

// validatePlay マストフォロー + Reneging (上位切り札はフォロー免除) を検証する。
func (g *SpoilFive) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadCard := g.currentTrick[0].Card
	leadIsTrump := g.isTrumpCard(leadCard)
	leadSuit := leadCard.GetDesign()

	// プレイヤーがフォロー義務を負う札を持っているか。
	// 切り札リード時: 上位切り札 (Reneging 可) は義務札に数えない。
	// 非切り札リード時: そのスートの札 (♥A は切り札なので除外) が義務札。
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
		return nil // フォローできる義務札がない → 何でも出せる
	}
	// 義務札がある: フォローしているか (切り札リードなら切り札、非切り札なら同スート) を確認。
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
func (g *SpoilFive) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	bestRank := g.spoilRank(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		// 切り札扱い、またはリードスートの札のみ勝者候補。
		if !g.isTrumpCard(tc.Card) && tc.Card.GetDesign() != leadSuit {
			continue
		}
		if r := g.spoilRank(tc.Card); r > bestRank {
			bestRank = r
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// spoilRank Spoil Five の固定ランクを返す (高いほど強い)。
func (g *SpoilFive) spoilRank(card *Card) int {
	d, v := card.GetDesign(), card.GetValue()
	switch {
	case d == g.trumpSuit && v == 5:
		return 1000 // 切り札の 5 (最強)
	case d == g.trumpSuit && v == 11:
		return 999 // 切り札の J
	case d == CardDesignHeart && v == 1:
		return 998 // ♥A (常に第 3 位)
	case d == g.trumpSuit && v == 1:
		return 997 // 切り札の A
	case d == g.trumpSuit && v == 13:
		return 996 // 切り札 K
	case d == g.trumpSuit && v == 12:
		return 995 // 切り札 Q
	case d == g.trumpSuit:
		return 900 + spoilPip(v) // その他の切り札 (10-high 簡略)
	default:
		return spoilPlainStrength(v) // 非切り札 (リードスート判定は trickWinner 側)
	}
}

// spoilPip 切り札の数札の相対強さ (10>9>...>2, 簡略)。
func spoilPip(v int) int {
	if v == 1 {
		return 14
	}
	return v
}

// spoilPlainStrength 非切り札の強さ (A 高)。
func spoilPlainStrength(v int) int {
	if v == 1 {
		return 14
	}
	return v
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *SpoilFive) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *SpoilFive) sortAllHands() {
	for _, p := range g.players {
		g.spoilSortHand(p)
	}
}

// spoilSortHand 手札を強さ順 (降順) にソートする。
func (g *SpoilFive) spoilSortHand(p *SpoilFivePlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		return g.spoilRank(cards[i]) > g.spoilRank(cards[j])
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playerName プレイヤー名を返す。
func (g *SpoilFive) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *SpoilFive) indexOfPlayerInTrick(playerIdx int) int {
	for i, tc := range g.currentTrick {
		if tc.PlayerIdx == playerIdx {
			return i
		}
	}
	return -1
}

// trickTopRank 現在のトリック勝者の札のランクを返す。見つからない場合は極小値。
func (g *SpoilFive) trickTopRank(winnerIdx int) int {
	idx := g.indexOfPlayerInTrick(winnerIdx)
	if idx < 0 {
		return -1 << 30
	}
	return g.spoilRank(g.currentTrick[idx].Card)
}

// findHumanIdx 人間プレイヤーのインデックス (-1=なし)。
func (g *SpoilFive) findHumanIdx() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// appendLog 棋譜にエントリを追加する。
func (g *SpoilFive) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
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
func (g *SpoilFive) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == SpoilFiveCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart トリックを取りに行く戦略プレイ。
func (g *SpoilFive) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		// リード: 強い札で取りに行く。
		return g.maxBy(player, valid, func(c *Card) int { return g.spoilRank(c) })
	}
	winnerIdx := g.trickWinner()
	topRank := g.trickTopRank(winnerIdx)
	winners := spoilFilter(valid, func(idx int) bool { return g.spoilRank(player.GetCard(idx)) > topRank })
	if len(winners) > 0 {
		// 最小コストで勝てる札。
		return g.minBy(player, winners, func(c *Card) int { return g.spoilRank(c) })
	}
	// 勝てない: 最弱札を捨てる。
	return g.minBy(player, valid, func(c *Card) int { return g.spoilRank(c) })
}

// minBy score が最小となるインデックスを返す。
func (g *SpoilFive) minBy(player *SpoilFivePlayer, indices []int, score func(*Card) int) int {
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
func (g *SpoilFive) maxBy(player *SpoilFivePlayer, indices []int, score func(*Card) int) int {
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

// spoilFilter 述語を満たすインデックスを抽出する。
func spoilFilter(indices []int, pred func(int) bool) []int {
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
func (g *SpoilFive) GetHint() *SpoilFiveHint {
	human := g.findHumanIdx()
	if human < 0 || g.phase != SpoilFivePhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.getValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuPlaySmart(human, valid)
	return &SpoilFiveHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *SpoilFive) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_high"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	winnerIdx := g.trickWinner()
	if g.spoilRank(card) > g.trickTopRank(winnerIdx) &&
		(g.isTrumpCard(card) || card.GetDesign() == g.currentTrick[0].Card.GetDesign()) {
		return "take_trick"
	}
	return "discard_low"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *SpoilFive) GetPhase() SpoilFivePhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *SpoilFive) SetPhase(phase SpoilFivePhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *SpoilFive) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *SpoilFive) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *SpoilFive) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *SpoilFive) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *SpoilFive) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *SpoilFive) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *SpoilFive) GetCurrentTrick() []*SpoilFiveTrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *SpoilFive) SetCurrentTrick(trick []*SpoilFiveTrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *SpoilFive) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *SpoilFive) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *SpoilFive) GetDealerIdx() int { return g.dealerIdx }

// GetTrumpSuit 切り札スート取得
func (g *SpoilFive) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *SpoilFive) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetPot 現在のポット額取得
func (g *SpoilFive) GetPot() int { return g.pot }

// SetPot ポット額設定 (テスト用)
func (g *SpoilFive) SetPot(n int) { g.pot = n }

// GetRoundWinnerIdx 直近ラウンド勝者取得 (-1=Spoil/未確定)
func (g *SpoilFive) GetRoundWinnerIdx() int { return g.roundWinnerIdx }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *SpoilFive) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝利プレイヤー取得 (-1=未確定)
func (g *SpoilFive) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayerCnt プレイヤー数取得
func (g *SpoilFive) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *SpoilFive) GetPlayer(i int) *SpoilFivePlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在の手番が人間か。
func (g *SpoilFive) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *SpoilFive) GetConfig() SpoilFiveConfig { return g.config }

// SetConfig 設定変更
func (g *SpoilFive) SetConfig(cfg SpoilFiveConfig) { g.config = cfg }

// GetActionLog 棋譜取得
func (g *SpoilFive) GetActionLog() []*ActionLogEntry { return g.actionLog }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *SpoilFive) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != SpoilFivePhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// spoilFiveJSON is the JSON wire format for SpoilFive.
type spoilFiveJSON struct {
	TrumpCards       *TrumpCards           `json:"tc"`
	Players          []*SpoilFivePlayer    `json:"ps"`
	Config           SpoilFiveConfig       `json:"cf"`
	Phase            SpoilFivePhase        `json:"ph"`
	RoundNumber      int                   `json:"rn"`
	TrickNumber      int                   `json:"tn"`
	CurrentPlayerIdx int                   `json:"ci"`
	CurrentTrick     []*SpoilFiveTrickCard `json:"ct"`
	LeadPlayerIdx    int                   `json:"li"`
	DealerIdx        int                   `json:"di"`
	TrumpSuit        int                   `json:"ts"`
	Pot              int                   `json:"po"`
	RoundWinnerIdx   int                   `json:"rw"`
	GameEndFlag      bool                  `json:"ge"`
	WinnerPlayer     int                   `json:"wp"`
	ActionLog        []*ActionLogEntry     `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *SpoilFive) MarshalJSON() ([]byte, error) {
	return json.Marshal(spoilFiveJSON{
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
		TrumpSuit:        g.trumpSuit,
		Pot:              g.pot,
		RoundWinnerIdx:   g.roundWinnerIdx,
		GameEndFlag:      g.gameEndFlag,
		WinnerPlayer:     g.winnerPlayer,
		ActionLog:        g.actionLog,
	})
}

// spoilFiveMaxSliceLen caps slice sizes during deserialisation.
const spoilFiveMaxSliceLen = 5000

// errSpoilFiveOversized is the single sentinel error for oversized input arrays.
var errSpoilFiveOversized = errors.New("spoilfive: input array exceeds maximum allowed size")

// errSpoilFiveInvalidPlayers is returned when restored state lacks exactly SpoilFivePlayerCnt players.
var errSpoilFiveInvalidPlayers = errors.New("spoilfive: invalid player count")

// errSpoilFiveInvalidTrick is returned when a restored trick card is nil/out of range.
var errSpoilFiveInvalidTrick = errors.New("spoilfive: invalid trick card")

// errSpoilFiveInvalidState is returned when a restored index/state field is out of range.
var errSpoilFiveInvalidState = errors.New("spoilfive: invalid state values in json")

// UnmarshalJSON implements json.Unmarshaler.
func (g *SpoilFive) UnmarshalJSON(data []byte) error {
	var j spoilFiveJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > spoilFiveMaxSliceLen || len(j.CurrentTrick) > spoilFiveMaxSliceLen ||
		len(j.ActionLog) > spoilFiveMaxSliceLen {
		return errSpoilFiveOversized
	}
	if len(j.Players) != SpoilFivePlayerCnt {
		return errSpoilFiveInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errSpoilFiveInvalidPlayers
		}
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= SpoilFivePlayerCnt ||
		j.LeadPlayerIdx < 0 || j.LeadPlayerIdx >= SpoilFivePlayerCnt ||
		j.DealerIdx < 0 || j.DealerIdx >= SpoilFivePlayerCnt ||
		j.RoundWinnerIdx < -1 || j.RoundWinnerIdx >= SpoilFivePlayerCnt ||
		j.WinnerPlayer < -1 || j.WinnerPlayer >= SpoilFivePlayerCnt ||
		j.TrumpSuit < 1 || j.TrumpSuit > 4 ||
		j.Pot < 0 ||
		j.RoundNumber < 1 ||
		j.TrickNumber < 1 || j.TrickNumber > SpoilFiveTrickCount ||
		j.Phase < SpoilFivePhasePlay || j.Phase > SpoilFivePhaseGameEnd {
		return errSpoilFiveInvalidState
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= SpoilFivePlayerCnt {
			return errSpoilFiveInvalidTrick
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
		g.currentTrick = make([]*SpoilFiveTrickCard, 0)
	}
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.trumpSuit = j.TrumpSuit
	g.pot = j.Pot
	g.roundWinnerIdx = j.RoundWinnerIdx
	g.gameEndFlag = j.GameEndFlag
	g.winnerPlayer = j.WinnerPlayer
	g.actionLog = j.ActionLog
	return nil
}
