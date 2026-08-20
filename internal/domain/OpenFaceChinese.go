//go:build !js || !wasm || casino

// Package domain オープンフェイス・チャイニーズポーカー (Open Face Chinese / OFC) の
// ドメインモデル。
//
// OFC は中国式ポーカーの近代的変種。本実装は人間 1 + CPU の 2〜4 人（既定はヘッズアップ
// =2 人）で、52 枚デッキを使う。各プレイヤーは 3 つの段 — 上段 (front, 3 枚)、中段
// (middle, 5 枚)、下段 (back, 5 枚) — を、配られた／引いたカードを 1 枚ずつ表向きに置いて
// 構築する。一度置いたカードは動かせない。
//
// 進行: 最初に各プレイヤーへ 5 枚配り、その 5 枚を 1 枚ずつ任意の段に置く。以降は 1 枚ずつ
// 引いて置く動作を、全段が埋まる（合計 13 枚配置）まで繰り返す。CPU は段の容量とハンド
// 強度を意識したヒューリスティックで自動配置する。
//
// ハンド強度は下段 ≥ 中段 ≥ 上段 を満たさねばならず、満たさない場合そのプレイヤーは
// ファウル (foul) し、全段で対戦相手に負ける。
//
// 採点 (全段確定後): まずファウル判定を行い、各プレイヤー対の 3 段を 1 段 +1 点で総当たり
// 比較する（ファウルした側は各段で負け、両者ファウルなら相殺）。3 段すべて勝てば +3 の
// スクープボーナス。さらに段ごとのロイヤリティ (Chinese Poker 共通ヘルパー流用) を点数に
// 加える。各プレイヤーのラウンド得点 = 全相手との合計。
//
// ファンタジーランド: ファウルせず上段が QQ 以上（クイーンのペア以上、cpFrontRoyalty が
// 正となる ThreeCardHandPair でペア値 12 以上、または ThreeCardHandThreeOfAKind）の場合、
// 次ラウンドにファンタジーランド権を得る。本実装ではフラグのみ保持し、ファンタジーランド
// 中のプレイヤーには 13 枚を一括で配る（簡略化: 通常配置 UI をそのまま使い、特別な
// ボーナス再獲得ルールやセット枚数差異は実装しない）。
//
// マッチは固定ラウンド数 (既定 4) をプレイし、累積得点の最高者を勝者とする。全 CPU でも
// 各ラウンドが必ず 13 枚配置で終わるため、ラウンド数上限により必ず終了する。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
)

// OpenFaceChineseHandSize 1 プレイヤーが置くカード総数 (3+5+5)
const OpenFaceChineseHandSize = 13

// OpenFaceChineseFrontSize 上段サイズ
const OpenFaceChineseFrontSize = 3

// OpenFaceChineseMiddleSize 中段サイズ
const OpenFaceChineseMiddleSize = 5

// OpenFaceChineseBackSize 下段サイズ
const OpenFaceChineseBackSize = 5

// OpenFaceChineseInitialDeal 最初に配る枚数
const OpenFaceChineseInitialDeal = 5

// OFC の段インデックス
const (
	// OpenFaceChineseRowFront 上段
	OpenFaceChineseRowFront = 0
	// OpenFaceChineseRowMiddle 中段
	OpenFaceChineseRowMiddle = 1
	// OpenFaceChineseRowBack 下段
	OpenFaceChineseRowBack = 2
)

// OpenFaceChinesePhase ゲームフェーズ
type OpenFaceChinesePhase int

// OFC のフェーズ定数
const (
	// OpenFaceChinesePhasePlacing カード配置フェーズ
	OpenFaceChinesePhasePlacing OpenFaceChinesePhase = 0
	// OpenFaceChinesePhaseRoundEnd ラウンド終了（採点済み）フェーズ
	OpenFaceChinesePhaseRoundEnd OpenFaceChinesePhase = 1
	// OpenFaceChinesePhaseGameEnd マッチ終了フェーズ
	OpenFaceChinesePhaseGameEnd OpenFaceChinesePhase = 2
)

// OpenFaceChineseHint ヒント情報（推奨配置段）
type OpenFaceChineseHint struct {
	Row    int    // 推奨段 (0=front,1=middle,2=back)
	Reason string // ヒント理由キー
}

// OpenFaceChinese オープンフェイス・チャイニーズポーカーのゲームクラス
type OpenFaceChinese struct {
	trumpCards       *TrumpCards
	players          []*OpenFaceChinesePlayer
	config           OpenFaceChineseConfig
	phase            OpenFaceChinesePhase
	roundNumber      int
	currentPlayerIdx int
	dealerIdx        int
	gameEndFlag      bool
	winnerIdx        int // -1=未確定/引き分け
	actionLogBase
}

// NewOpenFaceChinese コンストラクタ
func NewOpenFaceChinese(trumpCards *TrumpCards, players []*OpenFaceChinesePlayer, config OpenFaceChineseConfig) *OpenFaceChinese {
	return &OpenFaceChinese{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		winnerIdx:  -1,
	}
}

// NewDefaultOpenFaceChinese 既定設定 (ヘッズアップ, 人間 1 + CPU 1) で生成する。
func NewDefaultOpenFaceChinese() *OpenFaceChinese {
	cfg := DefaultOpenFaceChineseConfig()
	return NewOpenFaceChinese(NewTrumpCards(0), ofcBuildPlayers(cfg.PlayerCount), cfg)
}

// ofcBuildPlayers 人間 1 + 残り CPU のプレイヤー列を生成する。
func ofcBuildPlayers(count int) []*OpenFaceChinesePlayer {
	if count < OpenFaceChinesePlayerMin {
		count = OpenFaceChinesePlayerMin
	}
	if count > OpenFaceChinesePlayerMax {
		count = OpenFaceChinesePlayerMax
	}
	players := make([]*OpenFaceChinesePlayer, count)
	players[0] = NewOpenFaceChinesePlayer(true)
	for i := 1; i < count; i++ {
		players[i] = NewOpenFaceChinesePlayer(false)
	}
	return players
}

// Reset ゲーム初期化
func (g *OpenFaceChinese) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.actionLog = nil
	for _, p := range g.players {
		p.SetTotalScore(0)
		p.SetFantasyland(false)
	}
	g.startRound()
}

// NextRound 次のラウンドを開始する。
func (g *OpenFaceChinese) NextRound() {
	if g.phase != OpenFaceChinesePhaseRoundEnd || g.gameEndFlag {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % len(g.players)
	g.startRound()
}

// startRound 段をリセットし、初手を配って配置フェーズを開始する。
func (g *OpenFaceChinese) startRound() {
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards = NewTrumpCards(0)
	g.trumpCards.Shuffle()
	// ファンタジーランドのプレイヤーには 13 枚一括、通常は最初の 5 枚を配る。
	for _, p := range g.players {
		n := OpenFaceChineseInitialDeal
		if p.GetFantasyland() {
			n = OpenFaceChineseHandSize
		}
		for i := 0; i < n; i++ {
			if c := g.trumpCards.DrawCard(); c != nil {
				p.pending = append(p.pending, c)
			}
		}
	}
	g.currentPlayerIdx = (g.dealerIdx + 1) % len(g.players)
	g.phase = OpenFaceChinesePhasePlacing
	g.appendLog(-1, "deal", "new round dealt", nil)
}

// IsHumanTurn 現在の手番が人間か（配置フェーズ）。
func (g *OpenFaceChinese) IsHumanTurn() bool {
	if g.phase != OpenFaceChinesePhasePlacing {
		return false
	}
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// PlayerPlace 人間プレイヤーが先頭の保留カードを指定段 (0=front,1=middle,2=back) に置く。
func (g *OpenFaceChinese) PlayerPlace(row int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != OpenFaceChinesePhasePlacing {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.place(g.currentPlayerIdx, row)
}

// place 配置共通処理。1 枚置いた後、保留が尽きていれば手番を進め、ドローを補充する。
func (g *OpenFaceChinese) place(playerIdx, row int) error {
	p := g.players[playerIdx]
	card := firstPending(p)
	if err := p.placeCard(row); err != nil {
		return err
	}
	g.appendLog(playerIdx, "place", fmt.Sprintf("%s places %s on row %d", playerName(g.players, playerIdx), cardStr(card), row), []*Card{card})
	g.afterPlace(playerIdx)
	return nil
}

// firstPending 先頭の保留カードを返す（無ければ nil）。
func firstPending(p *OpenFaceChinesePlayer) *Card {
	if len(p.pending) == 0 {
		return nil
	}
	return p.pending[0]
}

// afterPlace 1 枚配置後の進行管理。保留が残っていれば同じ手番を続け、尽きたら次の
// ドローを補充するか手番を移し、全員 13 枚置き終わればラウンドを採点する。
func (g *OpenFaceChinese) afterPlace(playerIdx int) {
	p := g.players[playerIdx]
	if len(p.pending) > 0 {
		return // まだ置くカードが残っている（最初の 5 枚やファンタジーランド）。
	}
	if p.placedCount() >= OpenFaceChineseHandSize {
		// このプレイヤーは完了。次の未完了プレイヤーへ。
		g.advanceToNextPlayer()
		return
	}
	// 次の 1 枚を引く。
	if c := g.trumpCards.DrawCard(); c != nil {
		p.pending = append(p.pending, c)
	}
}

// advanceToNextPlayer まだ 13 枚置いていない次プレイヤーへ手番を移す。全員完了なら採点。
func (g *OpenFaceChinese) advanceToNextPlayer() {
	n := len(g.players)
	for k := 1; k <= n; k++ {
		ni := (g.currentPlayerIdx + k) % n
		if g.players[ni].placedCount() < OpenFaceChineseHandSize {
			g.currentPlayerIdx = ni
			// 次プレイヤーがまだ保留を持っていなければ 1 枚引く。
			if len(g.players[ni].pending) == 0 {
				if c := g.trumpCards.DrawCard(); c != nil {
					g.players[ni].pending = append(g.players[ni].pending, c)
				}
			}
			return
		}
	}
	g.scoreRound()
}

// CpuPlay 現在の手番が CPU の場合に保留カードを 1 枚配置する。
func (g *OpenFaceChinese) CpuPlay() {
	if g.gameEndFlag || g.phase != OpenFaceChinesePhasePlacing {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	row := g.cpuChooseRow(idx)
	_ = g.place(idx, row)
}

// cpuChooseRow CPU が保留カードを置く段を選ぶ。
func (g *OpenFaceChinese) cpuChooseRow(playerIdx int) int {
	p := g.players[playerIdx]
	card := firstPending(p)
	open := g.openRows(p)
	if len(open) == 0 {
		return OpenFaceChineseRowBack
	}
	if len(open) == 1 {
		return open[0]
	}
	if g.config.CpuDifficulty == OpenFaceChineseCpuDifficultyEasy || card == nil {
		return open[rand.Intn(len(open))]
	}
	// Normal と Hard はどちらも cpuPlaceSmart を用いる (意図的に同一)。
	// 段のファウル回避を優先する貪欲戦略で、現状この 1 種類で両難易度を兼ねる。
	return g.cpuPlaceSmart(p, card, open)
}

// openRows まだ満杯でない段のリストを下段→中段→上段の順で返す。
func (g *OpenFaceChinese) openRows(p *OpenFaceChinesePlayer) []int {
	var open []int
	for _, row := range []int{OpenFaceChineseRowBack, OpenFaceChineseRowMiddle, OpenFaceChineseRowFront} {
		if !p.rowFull(row) {
			open = append(open, row)
		}
	}
	return open
}

// cpuPlaceSmart 強い札は下段優先、弱い札は埋まり具合を見て上段へ振るヒューリスティック。
// 下段 ≥ 中段 ≥ 上段 を崩しにくいよう、強い札ほど下の段に置く。
func (g *OpenFaceChinese) cpuPlaceSmart(p *OpenFaceChinesePlayer, card *Card, open []int) int {
	v := ofcRankValue(card)
	// 強い札（J 以上）は空きがあれば下段、次に中段。
	if v >= 11 {
		for _, row := range []int{OpenFaceChineseRowBack, OpenFaceChineseRowMiddle, OpenFaceChineseRowFront} {
			if containsInt(open, row) {
				return row
			}
		}
	}
	// 弱い札は上段、次に中段、最後に下段。
	for _, row := range []int{OpenFaceChineseRowFront, OpenFaceChineseRowMiddle, OpenFaceChineseRowBack} {
		if containsInt(open, row) {
			return row
		}
	}
	return open[0]
}

// containsInt slice に v が含まれるか。
func containsInt(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// ofcRankValue カード値（A=14）を返す。
func ofcRankValue(c *Card) int {
	if c == nil {
		return 0
	}
	v := c.GetValue()
	if v == 1 {
		return 14
	}
	return v
}

// scoreRound 全プレイヤーの段を評価し、ファウル・段比較・ロイヤリティで得点を確定する。
func (g *OpenFaceChinese) scoreRound() {
	n := len(g.players)
	// ファウル判定とロイヤリティ算出。
	for _, p := range g.players {
		p.fouled = !ofcValidRows(p)
		if p.fouled {
			p.royalty = 0
		} else {
			p.royalty = ofcPlayerRoyalty(p)
		}
		p.roundScore = 0
	}
	// 総当たり段比較とロイヤリティ精算。段得点もロイヤリティ差も対戦相手間で
	// やり取りされるため、全プレイヤーの roundScore 合計は常に 0 になる
	// (中国式ポーカー / OFC のゼロサム精算)。ファウル者は royalty=0 のため、
	// 相手のロイヤリティを一方的に支払う形になる。
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			pi, pj := g.players[i], g.players[j]
			if pi == nil || pj == nil {
				continue
			}
			delta := ofcCompareScore(pi, pj) + (pi.royalty - pj.royalty)
			pi.roundScore += delta
			pj.roundScore -= delta
		}
	}
	for _, p := range g.players {
		if p == nil {
			continue
		}
		p.totalScore += p.roundScore
		// 次ラウンドのファンタジーランド権を判定する。
		p.fantasyland = !p.fouled && ofcQualifiesFantasyland(p)
	}
	g.appendLog(-1, "round_score", fmt.Sprintf("round %d scored", g.roundNumber), nil)
	g.phase = OpenFaceChinesePhaseRoundEnd
	g.checkGameEnd()
}

// ScoreRound 外部から採点を要求する。配置が全段完了している場合のみ採点する。
func (g *OpenFaceChinese) ScoreRound() {
	if g.phase != OpenFaceChinesePhasePlacing {
		return
	}
	for _, p := range g.players {
		if p.placedCount() < OpenFaceChineseHandSize {
			return
		}
	}
	g.scoreRound()
}

// checkGameEnd 目標ラウンド数に達したらマッチ終了とし、累積最高得点者を勝者とする。
func (g *OpenFaceChinese) checkGameEnd() {
	if g.roundNumber < g.config.TargetRounds {
		return
	}
	g.gameEndFlag = true
	g.phase = OpenFaceChinesePhaseGameEnd
	best, bestIdx, tie := -1<<31, -1, false
	for i, p := range g.players {
		switch {
		case p.totalScore > best:
			best = p.totalScore
			bestIdx = i
			tie = false
		case p.totalScore == best:
			tie = true
		}
	}
	if tie {
		g.winnerIdx = -1
	} else {
		g.winnerIdx = bestIdx
	}
	g.appendLog(-1, "game_end", "match finished", nil)
}

// --- Row validation / scoring helpers ---

// ofcValidRows 下段 ≥ 中段 ≥ 上段 を満たすか（ファウルでない）を返す。全段が満杯である
// ことを前提とする。
func ofcValidRows(p *OpenFaceChinesePlayer) bool {
	if len(p.front) != OpenFaceChineseFrontSize || len(p.middle) != OpenFaceChineseMiddleSize ||
		len(p.back) != OpenFaceChineseBackSize {
		return false
	}
	return cpValidateHands(p.front, p.middle, p.back)
}

// OpenFaceChinesePlacementFouls は card を row に置くと **確定的に** 反則
// (front > middle または middle > back) になるかを返す。
//
// **OFC の核心ルールは front ≦ middle ≦ back** で、破ると全段負け扱いになる。
// 置いた瞬間に確定する反則を事前に知らせるための判定 (#5676) -- Web は
// ofcPlacementFouls で同じことをしている。
//
// **まだ埋まっていない段は判定しない。**未確定を反則と呼ぶと、まだ挽回できる
// 配置まで避けさせてしまう。埋まった 2 段だけを比べるときは、残る段に比較を
// 無効化する手を仮置きする。
func OpenFaceChinesePlacementFouls(front, middle, back []*Card, card *Card, row int) bool {
	next := func(rowCards []*Card, target int) []*Card {
		if row != target {
			return rowCards
		}
		return append(append([]*Card(nil), rowCards...), card)
	}
	return ofcRowsAlreadyFouled(
		next(front, OpenFaceChineseRowFront),
		next(middle, OpenFaceChineseRowMiddle),
		next(back, OpenFaceChineseRowBack),
	)
}

// ofcRowsAlreadyFouled は埋まっている段どうしの強さの順がすでに崩れているかを返す。
func ofcRowsAlreadyFouled(front, middle, back []*Card) bool {
	frontFull := len(front) == OpenFaceChineseFrontSize
	middleFull := len(middle) == OpenFaceChineseMiddleSize
	backFull := len(back) == OpenFaceChineseBackSize

	switch {
	case frontFull && middleFull && backFull:
		return !cpValidateHands(front, middle, back)
	case middleFull && backFull:
		// front を無効化して middle と back だけを比べる。
		return !cpValidateHands(ofcNeutralFront(), middle, back)
	case frontFull && middleFull:
		// back を無効化して front と middle だけを比べる。
		return !cpValidateHands(front, middle, ofcStrongestBack())
	default:
		return false
	}
}

// ofcNeutralFront は比較を無効化するための最弱の上段を返す。
//
// **2-3-4 は使えない** -- 3 枚でも連番はストレートに数えられ、ハイカードの中段を
// 上回ってしまう。連番にもペアにもならない 2-4-7 を別スートで置く。
func ofcNeutralFront() []*Card {
	return []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignClover, 7, false),
	}
}

// ofcStrongestBack は比較を無効化するための最強の下段 (ロイヤルストレートフラッシュ)。
func ofcStrongestBack() []*Card {
	return []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 10, false),
	}
}

// ofcPlayerRoyalty 3 段のロイヤリティ合計を返す（Chinese Poker 共通ヘルパー流用）。
func ofcPlayerRoyalty(p *OpenFaceChinesePlayer) int {
	fr := evalThreeCardHand(p.front)
	mr := evalFiveCardHand(p.middle)
	br := evalFiveCardHand(p.back)
	return cpCalcRoyalty(fr, mr, br, p.front)
}

// ofcCompareScore プレイヤー a 視点での対 b 得点を返す。1 段勝ち +1、スクープ (3 段勝ち)
// で追加 +3 のボーナス。ファウルは各段で負け、両者ファウルは相殺 (0)。
func ofcCompareScore(a, b *OpenFaceChinesePlayer) int {
	switch {
	case a.fouled && b.fouled:
		return 0
	case a.fouled:
		return -6 // 相手の 3 段勝ち +3 スクープ
	case b.fouled:
		return 6
	}
	wins := 0
	if c := compareThreeCardHands(a.front, b.front); c > 0 {
		wins++
	} else if c < 0 {
		wins--
	}
	if c := cpCompareFiveCardHands(a.middle, b.middle); c > 0 {
		wins++
	} else if c < 0 {
		wins--
	}
	if c := cpCompareFiveCardHands(a.back, b.back); c > 0 {
		wins++
	} else if c < 0 {
		wins--
	}
	score := wins
	switch wins {
	case 3:
		score += 3 // スクープ
	case -3:
		score -= 3
	}
	return score
}

// ofcQualifiesFantasyland 上段が QQ 以上（クイーンのペア以上、またはスリーカード）の場合 true。
func ofcQualifiesFantasyland(p *OpenFaceChinesePlayer) bool {
	if len(p.front) != OpenFaceChineseFrontSize {
		return false
	}
	rank := evalThreeCardHand(p.front)
	if rank == ThreeCardHandThreeOfAKind {
		return true
	}
	if rank != ThreeCardHandPair {
		return false
	}
	return cpPairValue(p.front) >= 12 // Q=12 以上
}

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨配置段を返す。
func (g *OpenFaceChinese) GetHint() *OpenFaceChineseHint {
	if g.phase != OpenFaceChinesePhasePlacing {
		return nil
	}
	idx := g.currentPlayerIdx
	if idx < 0 || idx >= len(g.players) || !g.players[idx].GetIsHuman() {
		return nil
	}
	p := g.players[idx]
	card := firstPending(p)
	open := g.openRows(p)
	if len(open) == 0 || card == nil {
		return nil
	}
	row := g.cpuPlaceSmart(p, card, open)
	return &OpenFaceChineseHint{Row: row, Reason: ofcHintReason(card, row)}
}

// ofcHintReason 推奨段に応じた理由キーを返す。
func ofcHintReason(card *Card, row int) string {
	strong := ofcRankValue(card) >= 11
	switch {
	case row == OpenFaceChineseRowBack && strong:
		return "strong_back"
	case row == OpenFaceChineseRowFront && !strong:
		return "weak_front"
	default:
		return "balance"
	}
}

// --- Misc ---

// --- Getters ---

// GetPhase 現在のフェーズ取得
func (g *OpenFaceChinese) GetPhase() OpenFaceChinesePhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *OpenFaceChinese) SetPhase(phase OpenFaceChinesePhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *OpenFaceChinese) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *OpenFaceChinese) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *OpenFaceChinese) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *OpenFaceChinese) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *OpenFaceChinese) GetDealerIdx() int { return g.dealerIdx }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *OpenFaceChinese) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1=未確定/引き分け)
func (g *OpenFaceChinese) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *OpenFaceChinese) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *OpenFaceChinese) GetPlayer(i int) *OpenFaceChinesePlayer {
	return getPlayer(g.players, i)
}

// GetConfig 設定取得
func (g *OpenFaceChinese) GetConfig() OpenFaceChineseConfig { return g.config }

// SetConfig 設定変更。プレイヤー数が変わる場合は列を作り直す。
func (g *OpenFaceChinese) SetConfig(cfg OpenFaceChineseConfig) {
	g.config = cfg
	if len(g.players) != cfg.PlayerCount && cfg.PlayerCount >= OpenFaceChinesePlayerMin && cfg.PlayerCount <= OpenFaceChinesePlayerMax {
		g.players = ofcBuildPlayers(cfg.PlayerCount)
	}
}

// GetCurrentCard 現在の手番プレイヤーが置こうとしている保留カードを返す（無ければ nil）。
func (g *OpenFaceChinese) GetCurrentCard() *Card {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return nil
	}
	return firstPending(g.players[g.currentPlayerIdx])
}

// --- JSON ---

// openFaceChineseJSON is the JSON wire format for OpenFaceChinese.
type openFaceChineseJSON struct {
	TrumpCards       *TrumpCards              `json:"tc"`
	Players          []*OpenFaceChinesePlayer `json:"ps"`
	Config           OpenFaceChineseConfig    `json:"cf"`
	Phase            OpenFaceChinesePhase     `json:"ph"`
	RoundNumber      int                      `json:"rn"`
	CurrentPlayerIdx int                      `json:"ci"`
	DealerIdx        int                      `json:"di"`
	GameEndFlag      bool                     `json:"ge"`
	WinnerIdx        int                      `json:"wi"`
	ActionLog        []*ActionLogEntry        `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *OpenFaceChinese) MarshalJSON() ([]byte, error) {
	return json.Marshal(openFaceChineseJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DealerIdx:        g.dealerIdx,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		ActionLog:        g.actionLog,
	})
}

// openFaceChineseMaxSliceLen caps slice sizes during deserialisation.
const openFaceChineseMaxSliceLen = 5000

// errOpenFaceChineseOversized is the single sentinel error for oversized input arrays.
var errOpenFaceChineseOversized = errors.New("openfacechinese: input array exceeds maximum allowed size")

// errOpenFaceChineseInvalidPlayers is returned when the restored player count is invalid.
var errOpenFaceChineseInvalidPlayers = errors.New("openfacechinese: invalid player count")

// errOpenFaceChineseInvalidState is returned when a restored index/state field is out of range.
var errOpenFaceChineseInvalidState = errors.New("openfacechinese: invalid state values in json")

// UnmarshalJSON implements json.Unmarshaler.
func (g *OpenFaceChinese) UnmarshalJSON(data []byte) error {
	var j openFaceChineseJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > openFaceChineseMaxSliceLen || len(j.ActionLog) > openFaceChineseMaxSliceLen {
		return errOpenFaceChineseOversized
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if len(j.Players) < OpenFaceChinesePlayerMin || len(j.Players) > OpenFaceChinesePlayerMax {
		return errOpenFaceChineseInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errOpenFaceChineseInvalidPlayers
		}
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= len(j.Players) ||
		j.DealerIdx < 0 || j.DealerIdx >= len(j.Players) ||
		j.WinnerIdx < -1 || j.WinnerIdx >= len(j.Players) ||
		j.RoundNumber < 1 ||
		j.Phase < OpenFaceChinesePhasePlacing || j.Phase > OpenFaceChinesePhaseGameEnd {
		return errOpenFaceChineseInvalidState
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
