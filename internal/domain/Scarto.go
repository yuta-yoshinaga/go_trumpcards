//go:build !js || !wasm || extra3

// Package domain スカルト (Scarto / Piedmontese Tarot) のドメインモデル。
//
// Scarto はイタリア (ピエモンテ) の 78 枚タロットデッキを用いる 3 人用トリックテイキング
// ゲーム。タロッキ系の中で最も単純な部類で、入札 (bidding) もパートナーシップも無く、
// 「親 (scartatore, ディーラー) が余剰札を伏せて捨てる (scarto) → 切り札優先のトリックを
// 取り合う」だけの構造を持つ。各プレイヤーは自分の獲得点を最大化する個人戦。
//
// # デッキ (78 枚)
//
//   - スート札 56 枚: design = 1..4 (4 スート)、value = 1..14
//     (1-10, Valet=11, Cavalier=12, Dame=13, Roi/King=14)。
//   - 切り札 (trumps) 21 枚: design = ScartoTrumpDesign (5)、value = 1..21。
//   - エクスキューズ (Excuse / Matto) 1 枚: design = ScartoExcuseDesign (6)、value = 0。
//
// ブー (bouts / honours) は 切り札 1・切り札 21・エクスキューズ の 3 枚。
//
// # ルールと簡略化 (本実装が採用する縮小版)
//
//   - 配札: 3 人へ 25 枚ずつ (75 枚)。残り 3 枚 (タロン) は親が拾い、親は一時的に 28 枚を
//     持つ。配り順は 5 枚パケットを 5 巡し各プレイヤーへ 25 枚、最後のタロン 3 枚を親へ渡す。
//   - スカルト (親の捨て札): 親は 3 枚を伏せて捨て、手札を 25 枚に戻す。**得点札 (King・
//     ブー・コート札 = Valet/Cavalier/Dame/Roi) は捨てられない**。捨てられるのは 0.5 点札
//     (非切り札のピップ 1..10) のみ。もしそれが 3 枚に満たない稀な場合に限り、非ブーの切り札を
//     フォールバックとして許可する。捨てた 3 枚は親の獲得札 (捕獲点) に計上される (フレンチ
//     タロットのエカルトと同じ扱い)。人間が親なら cardIndices で選択、CPU が親なら自動選択。
//   - トリックプレイ (25 トリック): リードスートに従う義務。ボイド時は切り札を出す義務。
//     切り札が場に出ていれば (またはオーバートランプ時) 可能なら上位切り札を出す義務。エクス
//     キューズはいつでも出せ (フォロー免除)、トリックを取られず、出したプレイヤー自身のトリック
//     山に残る (フレンチタロットと同じ)。最強切り札が勝ち、無ければリードスートの最強札が勝つ。
//   - カードポイント (ハーフポイント): Roi/各ブー = 4.5、Dame = 3.5、Cavalier = 2.5、
//     Valet = 1.5、その他 = 0.5。合計 91 点。内部では丸め誤差を避けるため 2 倍した整数
//     (ハーフポイント) で保持する (合計 182)。
//   - 精算 (デクレアラー/コントラクト無し): 各プレイヤーの捕獲ハーフポイント h_i と 3 人平均を
//     比較したゼロサム精算。deal 得点 score_i = ScartoPlayerCnt × h_i − Σh。Σscore_i = 0
//     が構造的に成立する (= 3 × (h_i − 平均) をハーフポイント単位で表したもの)。全 78 枚が
//     必ず分配される (75 枚がトリック + 3 枚が親のスカルト) ため、Σh = 182 で余り札は生じない。
//   - 累積得点: TargetDeals ディール後、累積得点最上位が勝者。ScartoResult は人間視点。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// ScartoPlayerCnt プレイヤー数 (人間 1 + CPU 2)
const ScartoPlayerCnt = 3

// ScartoHandSize 各プレイヤーがトリックプレイで持つ札の枚数
const ScartoHandSize = 25

// ScartoSurplus タロン (親が拾って捨てる余剰札) の枚数
const ScartoSurplus = 3

// ScartoDeckSize デッキ枚数 (78 枚タロットデッキ)
const ScartoDeckSize = 78

// ScartoTrickCount 1 ディールのトリック数
const ScartoTrickCount = 25

// ScartoDefaultDeals マッチを構成するディール数 (既定)
const ScartoDefaultDeals = 5

// ScartoSuitCnt スート数
const ScartoSuitCnt = 4

// ScartoTrumpDesign 切り札 (trump) の仮想デザイン値。1..4 はスート、5 が切り札。
const ScartoTrumpDesign = 5

// ScartoExcuseDesign エクスキューズ (Excuse / Matto) の仮想デザイン値。
const ScartoExcuseDesign = 6

// ScartoExcuseValue エクスキューズのカード値。
const ScartoExcuseValue = 0

// ScartoMaxTrump 切り札の最大値 (21)。
const ScartoMaxTrump = 21

// ScartoPetitValue プティ (最小の切り札, ブー) の値。
const ScartoPetitValue = 1

// ScartoKingValue スート札のキング (Roi) の値。
const ScartoKingValue = 14

// ScartoCourtMin コート札 (Valet/Cavalier/Dame/Roi) の最小値。11 以上がコート札。
const ScartoCourtMin = 11

// ScartoPhase ゲームフェーズ
type ScartoPhase int

// Scarto のフェーズ定数 (入札フェーズは無い)
const (
	// ScartoPhaseScarto 親のスカルト (捨て札) フェーズ
	ScartoPhaseScarto ScartoPhase = 0
	// ScartoPhasePlay トリックプレイフェーズ
	ScartoPhasePlay ScartoPhase = 1
	// ScartoPhaseTrickEnd トリック終了フェーズ
	ScartoPhaseTrickEnd ScartoPhase = 2
	// ScartoPhaseRoundEnd ディール終了フェーズ
	ScartoPhaseRoundEnd ScartoPhase = 3
	// ScartoPhaseGameEnd ゲーム終了フェーズ
	ScartoPhaseGameEnd ScartoPhase = 4
)

// ScartoPhaseMin フェーズ下限 (検証用)
const ScartoPhaseMin = int(ScartoPhaseScarto)

// ScartoPhaseMax フェーズ上限 (検証用)
const ScartoPhaseMax = int(ScartoPhaseGameEnd)

// ScartoOutcome ディール結果 (人間視点の deal 精算符号)
type ScartoOutcome int

// Scarto のディール結果定数
const (
	// ScartoOutcomeNone 未確定 / 平均通り
	ScartoOutcomeNone ScartoOutcome = 0
	// ScartoOutcomeWin 人間が平均を上回った
	ScartoOutcomeWin ScartoOutcome = 1
	// ScartoOutcomeLoss 人間が平均を下回った
	ScartoOutcomeLoss ScartoOutcome = 2
)

// ScartoResult 人間視点のマッチ結果。
// GameResult は共有ファイル internal/domain/game_result.go に移動したので到達可能に
// なったが、この型名は JSON ペイロードに出るため統合していない（#4462）。値は
// GameResult と同一。
type ScartoResult int

// Scarto のマッチ結果定数
const (
	// ScartoResultLose 敗北
	ScartoResultLose ScartoResult = -1
	// ScartoResultNone 未確定 / 引き分け
	ScartoResultNone ScartoResult = 0
	// ScartoResultWin 勝利
	ScartoResultWin ScartoResult = 1
)

// ScartoHint ヒント情報
type ScartoHint struct {
	CardIndices []int  // 推奨カードインデックス (スカルト/プレイ)
	Reason      string // ヒント理由キー
}

// Scarto スカルトのゲームクラス
type Scarto struct {
	deck             []*Card
	deckDrawCnt      int
	players          []*ScartoPlayer
	config           ScartoConfig
	phase            ScartoPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	scarto           []*Card // 親が捨てた 3 枚 (親の捕獲点に計上)
	// --- scoring ---
	playerScores    [ScartoPlayerCnt]int
	dealScores      [ScartoPlayerCnt]int // 直近ディールの精算値 (表示/結果判定用)
	lastTrickWinner int
	outcome         ScartoOutcome
	result          ScartoResult
	scored          bool
	gameEndFlag     bool
	winnerPlayer    int
	actionLogBase
}

// NewScarto コンストラクタ
func NewScarto(players []*ScartoPlayer, config ScartoConfig) *Scarto {
	return &Scarto{
		players:         players,
		config:          config,
		winnerPlayer:    -1,
		lastTrickWinner: -1,
	}
}

// NewDefaultScarto 標準の 3 人構成 (人間 1, CPU 2) と既定設定で生成する。
func NewDefaultScarto() *Scarto {
	players := make([]*ScartoPlayer, ScartoPlayerCnt)
	players[0] = NewScartoPlayer(true)
	for i := 1; i < ScartoPlayerCnt; i++ {
		players[i] = NewScartoPlayer(false)
	}
	return NewScarto(players, DefaultScartoConfig())
}

// buildScartoDeck 78 枚タロットデッキを直接構築する。スート札 (design 1..4, value
// 1..14) 56 枚 + 切り札 (design 5, value 1..21) 21 枚 + エクスキューズ (design 6, value 0)。
func buildScartoDeck() []*Card {
	deck := make([]*Card, 0, ScartoDeckSize)
	for suit := 1; suit <= ScartoSuitCnt; suit++ {
		for val := 1; val <= ScartoKingValue; val++ {
			deck = append(deck, NewCard(suit, val, false))
		}
	}
	for val := 1; val <= ScartoMaxTrump; val++ {
		deck = append(deck, NewCard(ScartoTrumpDesign, val, false))
	}
	deck = append(deck, NewCard(ScartoExcuseDesign, ScartoExcuseValue, false))
	return deck
}

// Reset ゲーム初期化
func (g *Scarto) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.playerScores = [ScartoPlayerCnt]int{}
	g.dealScores = [ScartoPlayerCnt]int{}
	g.result = ScartoResultNone
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のディールを開始する
func (g *Scarto) NextRound() {
	if g.phase != ScartoPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % ScartoPlayerCnt
	g.startRound()
}

// startRound 手札を配り、スカルトフェーズを開始する。
func (g *Scarto) startRound() {
	g.trickNumber = 0
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.lastTrickWinner = -1
	g.scarto = nil
	g.outcome = ScartoOutcomeNone
	g.scored = false
	g.dealScores = [ScartoPlayerCnt]int{}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.deal()
	g.sortAllHands()
	g.currentPlayerIdx = g.dealerIdx
	g.phase = ScartoPhaseScarto
}

// deal 5 枚パケットを 5 巡して各プレイヤーへ 25 枚を配り、タロン 3 枚を親に渡す
// (親は一時的に 28 枚を持つ)。
func (g *Scarto) deal() {
	g.deck = buildScartoDeck()
	rand.Shuffle(len(g.deck), func(i, j int) {
		g.deck[i], g.deck[j] = g.deck[j], g.deck[i]
	})
	g.deckDrawCnt = 0
	const packet = 5
	rounds := ScartoHandSize / packet // 5 巡
	for r := 0; r < rounds; r++ {
		for j := 0; j < ScartoPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % ScartoPlayerCnt
			for k := 0; k < packet; k++ {
				if c := g.drawCard(); c != nil {
					g.players[idx].AddCard(c)
				}
			}
		}
	}
	// 残りのタロン (3 枚) を親へ。
	for g.deckDrawCnt < len(g.deck) {
		if c := g.drawCard(); c != nil {
			g.players[g.dealerIdx].AddCard(c)
		}
	}
}

// drawCard デッキから 1 枚配る (尽きたら nil)。
func (g *Scarto) drawCard() *Card {
	return drawFromDeck(g.deck, &g.deckDrawCnt)
}

// --- Scarto (dealer discard) ---

// PlayerScarto 人間の親が 3 枚を伏せて捨てる。
func (g *Scarto) PlayerScarto(cardIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ScartoPhaseScarto {
		return ErrWrongPhase
	}
	if !g.isHumanScartoTurn() {
		return ErrNotHumanTurn
	}
	return g.doScarto(cardIndices)
}

// CpuScarto CPU の親が 3 枚を自動で捨てる。
func (g *Scarto) CpuScarto() {
	if g.gameEndFlag || g.phase != ScartoPhaseScarto {
		return
	}
	if g.isHumanScartoTurn() {
		return
	}
	_ = g.doScarto(g.cpuSelectScarto(g.dealerIdx))
}

// doScarto スカルトの共通処理。捨てた 3 枚を親の得点札 (scarto) とする。
func (g *Scarto) doScarto(cardIndices []int) error {
	player := g.players[g.dealerIdx]
	if len(cardIndices) != ScartoSurplus {
		return NewDomainError(ErrInvalidCard, "ちょうど 3 枚を捨ててください")
	}
	seen := make(map[int]bool, ScartoSurplus)
	for _, idx := range cardIndices {
		if idx < 0 || idx >= player.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidCard, "同じカードを 2 回選べません")
		}
		seen[idx] = true
	}
	if err := g.validateScarto(player, cardIndices); err != nil {
		return err
	}
	discarded := player.RemoveCards(cardIndices)
	g.scarto = discarded
	g.appendLog(g.dealerIdx, "scarto",
		fmt.Sprintf("%s discards %d cards (scarto)", playerName(g.players, g.dealerIdx), len(discarded)), discarded)
	g.sortAllHands()
	g.startPlay()
	return nil
}

// validateScarto スカルトの合法性を検証する。得点札 (King/ブー/コート札) は捨てられない。
// 捨てられるのは 0.5 点札 (非切り札のピップ)。それが 3 枚に満たない場合のみ非ブー切り札を許可。
func (g *Scarto) validateScarto(player *ScartoPlayer, cardIndices []int) error {
	discardable := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		if scartoDiscardable(player.GetCard(i)) {
			discardable++
		}
	}
	allowTrump := discardable < ScartoSurplus
	for _, idx := range cardIndices {
		c := player.GetCard(idx)
		if c == nil {
			return NewDomainError(ErrInvalidCard, "カードが不正です")
		}
		if scartoIsExcuse(c) {
			return NewDomainError(ErrInvalidPlay, "エクスキューズは捨てられません")
		}
		if scartoIsBout(c) {
			return NewDomainError(ErrInvalidPlay, "ブー (切り札1/21) は捨てられません")
		}
		if scartoIsTrump(c) {
			if !allowTrump {
				return NewDomainError(ErrInvalidPlay, "切り札は (やむを得ない場合を除き) 捨てられません")
			}
			continue
		}
		if c.GetValue() >= ScartoCourtMin {
			return NewDomainError(ErrInvalidPlay, "得点札 (King/コート札) は捨てられません")
		}
	}
	return nil
}

// scartoDiscardable 通常スカルトに出せる札か (非切り札・非エクスキューズ・非コートのピップ)。
func scartoDiscardable(c *Card) bool {
	if c == nil || scartoIsTrump(c) || scartoIsExcuse(c) {
		return false
	}
	return c.GetValue() < ScartoCourtMin
}

// --- Play ---

// startPlay プレイフェーズを開始する。親の左隣がリードする。
func (g *Scarto) startPlay() {
	g.sortAllHands()
	g.trickNumber = 1
	g.currentTrick = nil
	g.leadPlayerIdx = (g.dealerIdx + 1) % ScartoPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = ScartoPhasePlay
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Scarto) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ScartoPhasePlay {
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

// CpuPlay CPU プレイヤーが 1 ターン実行する。
func (g *Scarto) CpuPlay() {
	if g.gameEndFlag || g.phase != ScartoPhasePlay {
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
func (g *Scarto) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), scartoCardStr(card)), []*Card{card})
	if len(g.currentTrick) == ScartoPlayerCnt {
		g.phase = ScartoPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % ScartoPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する。エクスキューズは出した本人が保持し、残りを
// 勝者のトリック山へ。最終トリックなら RoundEnd に入り得点計算を発火する。
func (g *Scarto) ResolveTrick() {
	if g.phase != ScartoPhaseTrickEnd || len(g.currentTrick) != ScartoPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	excuseOwner := -1
	var excuseCard *Card
	won := make([]*Card, 0, ScartoPlayerCnt)
	allCards := make([]*Card, 0, ScartoPlayerCnt)
	for _, tc := range g.currentTrick {
		if tc == nil {
			continue
		}
		allCards = append(allCards, tc.Card)
		if scartoIsExcuse(tc.Card) {
			excuseOwner = tc.PlayerIdx
			excuseCard = tc.Card
			continue
		}
		won = append(won, tc.Card)
	}
	g.players[winnerIdx].AddTrick(won)
	if excuseOwner >= 0 && excuseCard != nil {
		g.players[excuseOwner].AddTrick([]*Card{excuseCard})
	}
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", playerName(g.players, winnerIdx), g.trickNumber), allCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= ScartoTrickCount {
		g.lastTrickWinner = winnerIdx
		g.phase = ScartoPhaseRoundEnd
		g.enterRoundEnd()
	} else {
		g.phase = ScartoPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Scarto) NextTrick() {
	if g.phase != ScartoPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = ScartoPhasePlay
}

// ScoreRound RoundEnd フェーズでの得点計算を行う (enterRoundEnd を idempotent に呼ぶ)。
func (g *Scarto) ScoreRound() {
	if g.phase != ScartoPhaseRoundEnd {
		return
	}
	g.enterRoundEnd()
}

// enterRoundEnd RoundEnd 突入時に一度だけ得点計算と精算を行う (scored フラグでガード)。
func (g *Scarto) enterRoundEnd() {
	if g.scored {
		return
	}
	g.scored = true
	half := g.capturedHalfPoints()
	g.dealScores = scartoSettleDeal(half)
	for i := 0; i < ScartoPlayerCnt; i++ {
		g.playerScores[i] += g.dealScores[i]
	}
	g.outcome = g.humanOutcome()
	g.appendLog(-1, "round_score",
		fmt.Sprintf("deal %d: captured half-points %v -> deal scores %v",
			g.roundNumber, half, g.dealScores), nil)
	g.checkGameEnd()
}

// capturedHalfPoints 各プレイヤーが捕獲したカードのハーフポイント合計を返す。親のスカルト
// 3 枚は親の捕獲点に加算される。全 78 枚が分配されるため合計は必ず 182 になる。
func (g *Scarto) capturedHalfPoints() [ScartoPlayerCnt]int {
	var half [ScartoPlayerCnt]int
	for i := 0; i < ScartoPlayerCnt && i < len(g.players); i++ {
		sum := 0
		for _, trick := range g.players[i].GetTricksTaken() {
			for _, c := range trick {
				sum += scartoCardHalfPoints(c)
			}
		}
		half[i] = sum
	}
	if g.dealerIdx >= 0 && g.dealerIdx < ScartoPlayerCnt {
		for _, c := range g.scarto {
			half[g.dealerIdx] += scartoCardHalfPoints(c)
		}
	}
	return half
}

// humanOutcome 人間の deal 精算符号から結果を返す。
func (g *Scarto) humanOutcome() ScartoOutcome {
	human := findHumanIdx(g.players)
	if human < 0 {
		return ScartoOutcomeNone
	}
	switch {
	case g.dealScores[human] > 0:
		return ScartoOutcomeWin
	case g.dealScores[human] < 0:
		return ScartoOutcomeLoss
	default:
		return ScartoOutcomeNone
	}
}

// checkGameEnd 規定ディール数を終えたらマッチ終了を判定し、累積得点最上位を勝者とする。
func (g *Scarto) checkGameEnd() {
	if g.roundNumber < g.config.TargetDeals {
		return
	}
	leader, best := 0, g.playerScores[0]
	tie := false
	for i := 1; i < ScartoPlayerCnt; i++ {
		if g.playerScores[i] > best {
			best = g.playerScores[i]
			leader = i
			tie = false
		} else if g.playerScores[i] == best {
			tie = true
		}
	}
	g.gameEndFlag = true
	g.phase = ScartoPhaseGameEnd
	g.result = g.humanResult(leader, tie)
	if tie {
		g.winnerPlayer = -1
		g.appendLog(-1, "game_end", "the match ends in a draw", nil)
	} else {
		g.winnerPlayer = leader
		g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the match!", playerName(g.players, leader)), nil)
	}
}

// humanResult 人間 (seat 0) 視点でマッチ結果を返す。単独トップなら Win、トップ同点なら None。
func (g *Scarto) humanResult(leader int, tie bool) ScartoResult {
	human := findHumanIdx(g.players)
	if human < 0 {
		return ScartoResultNone
	}
	if g.playerScores[human] == g.playerScores[leader] {
		if tie {
			return ScartoResultNone
		}
		return ScartoResultWin
	}
	return ScartoResultLose
}

// --- Scoring helper (pure) ---

// scartoSettleDeal 各プレイヤーの捕獲ハーフポイントからゼロサムの deal 得点を計算する純粋関数。
// score_i = ScartoPlayerCnt × half_i − Σhalf。Σscore_i = 0 が構造的に成立する。
func scartoSettleDeal(half [ScartoPlayerCnt]int) [ScartoPlayerCnt]int {
	total := 0
	for _, h := range half {
		total += h
	}
	var out [ScartoPlayerCnt]int
	for i, h := range half {
		out[i] = ScartoPlayerCnt*h - total
	}
	return out
}

// --- Card classification / points ---

// scartoIsTrump 切り札か。
func scartoIsTrump(c *Card) bool {
	return c != nil && c.GetDesign() == ScartoTrumpDesign
}

// scartoIsExcuse エクスキューズか。
func scartoIsExcuse(c *Card) bool {
	return c != nil && c.GetDesign() == ScartoExcuseDesign
}

// scartoIsBout ブー (切り札1 / 切り札21 / Excuse) か。
func scartoIsBout(c *Card) bool {
	if c == nil {
		return false
	}
	if scartoIsExcuse(c) {
		return true
	}
	return scartoIsTrump(c) && (c.GetValue() == ScartoPetitValue || c.GetValue() == ScartoMaxTrump)
}

// scartoCardHalfPoints カードのハーフポイント (点数×2) を返す。
// Roi/各ブー = 9, Dame = 7, Cavalier = 5, Valet = 3, その他 = 1。
func scartoCardHalfPoints(c *Card) int {
	if c == nil {
		return 0
	}
	if scartoIsBout(c) {
		return 9
	}
	if scartoIsTrump(c) || scartoIsExcuse(c) {
		return 1
	}
	switch c.GetValue() {
	case ScartoKingValue: // Roi
		return 9
	case 13: // Dame
		return 7
	case 12: // Cavalier
		return 5
	case 11: // Valet
		return 3
	default:
		return 1
	}
}

// --- Trick logic ---

// ledSuit 現在のトリックのリードスートを返す。最初の非エクスキューズ札の design。
// エクスキューズのみでスートが未確定なら -1。
func (g *Scarto) ledSuit() int {
	for _, tc := range g.currentTrick {
		if tc == nil || tc.Card == nil {
			continue
		}
		if !scartoIsExcuse(tc.Card) {
			return tc.Card.GetDesign()
		}
	}
	return -1
}

// highestTrumpInTrick 現在のトリック中の最強切り札の値を返す (0=切り札なし)。
func (g *Scarto) highestTrumpInTrick() int {
	best := 0
	for _, tc := range g.currentTrick {
		if tc == nil {
			continue
		}
		if scartoIsTrump(tc.Card) && tc.Card.GetValue() > best {
			best = tc.Card.GetValue()
		}
	}
	return best
}

// validatePlay マストフォロー + 切り札義務 + オーバートランプ義務を検証する。
func (g *Scarto) validatePlay(playerIdx int, card *Card) error {
	valid := g.getValidPlayIndices(playerIdx)
	player := g.players[playerIdx]
	for _, idx := range valid {
		if player.GetCard(idx) == card {
			return nil
		}
	}
	return NewDomainError(ErrInvalidPlay, "フォロー義務・切り札義務・オーバートランプ義務に反しています")
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Scarto) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	n := player.GetCardsSize()
	all := make([]int, 0, n)
	for i := 0; i < n; i++ {
		all = append(all, i)
	}
	if len(g.currentTrick) == 0 {
		return all
	}
	led := g.ledSuit()
	if led == -1 {
		return all
	}
	excuseIdx := -1
	for i := 0; i < n; i++ {
		if scartoIsExcuse(player.GetCard(i)) {
			excuseIdx = i
		}
	}
	highestTrump := g.highestTrumpInTrick()
	var base []int
	if led == ScartoTrumpDesign {
		base = g.trumpFollowIndices(player, highestTrump)
	} else {
		base = g.suitFollowIndices(player, led, highestTrump)
	}
	if excuseIdx >= 0 {
		base = scartoAppendUnique(base, excuseIdx)
	}
	if len(base) == 0 {
		return all
	}
	return base
}

// trumpFollowIndices 切り札がリードされた場合の合法な非エクスキューズ札を返す。
func (g *Scarto) trumpFollowIndices(player *ScartoPlayer, highestTrump int) []int {
	trumps := g.suitOf(player, ScartoTrumpDesign)
	if len(trumps) == 0 {
		return g.nonExcuseIndices(player)
	}
	higher := scartoFilter(trumps, func(idx int) bool {
		c := player.GetCard(idx)
		return c != nil && c.GetValue() > highestTrump
	})
	if len(higher) > 0 {
		return higher
	}
	return trumps
}

// suitFollowIndices スートがリードされた場合の合法な非エクスキューズ札を返す。
func (g *Scarto) suitFollowIndices(player *ScartoPlayer, led, highestTrump int) []int {
	ledCards := g.suitOf(player, led)
	if len(ledCards) > 0 {
		return ledCards
	}
	trumps := g.suitOf(player, ScartoTrumpDesign)
	if len(trumps) == 0 {
		return g.nonExcuseIndices(player)
	}
	higher := scartoFilter(trumps, func(idx int) bool {
		c := player.GetCard(idx)
		return c != nil && c.GetValue() > highestTrump
	})
	if highestTrump > 0 && len(higher) > 0 {
		return higher
	}
	return trumps
}

// suitOf 指定 design の (非エクスキューズ) 手札インデックスを返す。
func (g *Scarto) suitOf(player *ScartoPlayer, design int) []int {
	var out []int
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c == nil || scartoIsExcuse(c) {
			continue
		}
		if c.GetDesign() == design {
			out = append(out, i)
		}
	}
	return out
}

// nonExcuseIndices エクスキューズを除く全手札インデックスを返す。
func (g *Scarto) nonExcuseIndices(player *ScartoPlayer) []int {
	var out []int
	for i := 0; i < player.GetCardsSize(); i++ {
		if !scartoIsExcuse(player.GetCard(i)) {
			out = append(out, i)
		}
	}
	return out
}

// trickWinner トリックの勝者を決定する。切り札があれば最強切り札、無ければリードスートの最強札。
// エクスキューズは決して勝たない。
func (g *Scarto) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	led := g.ledSuit()
	winIdx := g.currentTrick[0].PlayerIdx
	winRank := -1
	for _, tc := range g.currentTrick {
		if tc == nil {
			continue
		}
		r := scartoWinRank(tc.Card, led)
		if r > winRank {
			winRank = r
			winIdx = tc.PlayerIdx
		}
	}
	return winIdx
}

// scartoWinRank トリック勝敗比較用のランクを返す (高いほど強い)。エクスキューズ = -1、
// 切り札 = 1000+値、リードスート = 値、それ以外 = -1。
func scartoWinRank(c *Card, led int) int {
	if c == nil || scartoIsExcuse(c) {
		return -1
	}
	if scartoIsTrump(c) {
		return 1000 + c.GetValue()
	}
	if c.GetDesign() == led {
		return c.GetValue()
	}
	return -1
}

// --- CPU AI ---

// cpuSelectScarto CPU の親が捨てる 3 枚のインデックスを選ぶ。得点札 (King/コート/ブー) は
// 温存し、価値の低い 0.5 点札から捨てる。足りない場合のみ非ブー切り札で補う。
func (g *Scarto) cpuSelectScarto(playerIdx int) []int {
	p := g.players[playerIdx]
	n := p.GetCardsSize()
	idxs := make([]int, n)
	for i := range idxs {
		idxs[i] = i
	}
	keepValue := func(c *Card) int {
		if c == nil {
			return -1
		}
		if scartoIsExcuse(c) {
			return 100000
		}
		if scartoIsBout(c) {
			return 80000
		}
		if !scartoIsTrump(c) && c.GetValue() >= ScartoCourtMin {
			return 90000
		}
		if scartoIsTrump(c) {
			return 10000 + c.GetValue()
		}
		return c.GetValue()*10 + scartoCardHalfPoints(c)
	}
	sort.SliceStable(idxs, func(a, b int) bool {
		return keepValue(p.GetCard(idxs[a])) < keepValue(p.GetCard(idxs[b]))
	})
	discardable := make([]int, 0, n)
	trumpFallback := make([]int, 0, n)
	for _, idx := range idxs {
		c := p.GetCard(idx)
		if scartoDiscardable(c) {
			discardable = append(discardable, idx)
		} else if scartoIsTrump(c) && !scartoIsBout(c) {
			trumpFallback = append(trumpFallback, idx)
		}
	}
	chosen := make([]int, 0, ScartoSurplus)
	for _, idx := range discardable {
		if len(chosen) >= ScartoSurplus {
			break
		}
		chosen = append(chosen, idx)
	}
	for _, idx := range trumpFallback {
		if len(chosen) >= ScartoSurplus {
			break
		}
		chosen = append(chosen, idx)
	}
	return chosen
}

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Scarto) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == ScartoCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 個人戦の自己利益プレイ。リードは安い札で温存、フォローは勝てるなら最弱の
// 勝ち札で取りに行き、取れなければ最安札を捨てる。
func (g *Scarto) cpuPlaySmart(playerIdx int, valid []int) int {
	p := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		return g.minByPoints(playerIdx, valid)
	}
	led := g.ledSuit()
	winnerIdx := g.trickWinner()
	pos := g.indexInTrick(winnerIdx)
	if pos < 0 {
		return g.minByPoints(playerIdx, valid)
	}
	winCard := g.currentTrick[pos].Card
	winners := scartoFilter(valid, func(idx int) bool {
		c := p.GetCard(idx)
		return c != nil && scartoWinRank(c, led) > scartoWinRank(winCard, led)
	})
	if len(winners) > 0 {
		return g.minByRank(playerIdx, winners)
	}
	return g.minByPoints(playerIdx, valid)
}

// indexInTrick currentTrick 内で playerIdx の位置を返す (-1=なし)。
func (g *Scarto) indexInTrick(playerIdx int) int {
	for i, tc := range g.currentTrick {
		if tc != nil && tc.PlayerIdx == playerIdx {
			return i
		}
	}
	return -1
}

// minByRank 勝敗ランク最小の札を返す。
func (g *Scarto) minByRank(playerIdx int, indices []int) int {
	p := g.players[playerIdx]
	led := g.ledSuit()
	best := indices[0]
	bestScore := scartoPlayRank(p.GetCard(best), led)
	for _, idx := range indices[1:] {
		if s := scartoPlayRank(p.GetCard(idx), led); s < bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// minByPoints ハーフポイント最小の札を返す。
func (g *Scarto) minByPoints(playerIdx int, indices []int) int {
	p := g.players[playerIdx]
	best := indices[0]
	bestScore := scartoPointKey(p.GetCard(best))
	for _, idx := range indices[1:] {
		if s := scartoPointKey(p.GetCard(idx)); s < bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// scartoPointKey 点札温存判断用のキー。ブー/エクスキューズは高く扱う。
func scartoPointKey(c *Card) int {
	if c == nil {
		return 0
	}
	if scartoIsBout(c) {
		return 100 + scartoCardHalfPoints(c)
	}
	return scartoCardHalfPoints(c)
}

// scartoPlayRank プレイ順比較用のランク (エクスキューズは最弱扱い)。
func scartoPlayRank(c *Card, led int) int {
	if c == nil {
		return -100
	}
	if scartoIsExcuse(c) {
		return -100
	}
	if scartoIsTrump(c) {
		return 1000 + c.GetValue()
	}
	if c.GetDesign() == led {
		return c.GetValue()
	}
	return c.GetValue()
}

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨アクションを返す。
func (g *Scarto) GetHint() *ScartoHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.gameEndFlag {
		return nil
	}
	switch g.phase {
	case ScartoPhaseScarto:
		if g.dealerIdx != human {
			return nil
		}
		return &ScartoHint{CardIndices: g.cpuSelectScarto(human), Reason: "scarto_weak"}
	case ScartoPhasePlay:
		if g.currentPlayerIdx != human {
			return nil
		}
		valid := g.getValidPlayIndices(human)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuSelectPlayCard(human)
		return &ScartoHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
	}
	return nil
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Scarto) playHintReason(playerIdx, chosenIdx int) string {
	player := g.players[playerIdx]
	card := player.GetCard(chosenIdx)
	if len(g.currentTrick) == 0 {
		return "lead_low"
	}
	if scartoIsExcuse(card) {
		return "play_excuse"
	}
	led := g.ledSuit()
	winnerIdx := g.trickWinner()
	pos := g.indexInTrick(winnerIdx)
	if pos >= 0 {
		winCard := g.currentTrick[pos].Card
		if scartoWinRank(card, led) > scartoWinRank(winCard, led) {
			return "follow_win"
		}
	}
	return "follow_duck"
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Scarto) sortAllHands() {
	for _, p := range g.players {
		scartoSortHand(p)
	}
}

// scartoSortHand 手札をスート→値でソートする (切り札・エクスキューズは末尾)。
func scartoSortHand(p *ScartoPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		di, dj := cards[i].GetDesign(), cards[j].GetDesign()
		if di != dj {
			return di < dj
		}
		return cards[i].GetValue() < cards[j].GetValue()
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// isHumanScartoTurn 現在のスカルト手番が人間 (=人間が親) か。
func (g *Scarto) isHumanScartoTurn() bool {
	if g.dealerIdx < 0 || g.dealerIdx >= len(g.players) {
		return false
	}
	return g.players[g.dealerIdx].GetIsHuman()
}

// appendLog 棋譜にエントリを追加する。
func (g *Scarto) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.appendLogAt(len(g.actionLog)+1, playerIdx, actionType, detail, cards)
}

// scartoCardStr カードのログ表示文字列 (切り札・エクスキューズ対応)。
func scartoCardStr(c *Card) string {
	if c == nil {
		return "??"
	}
	if scartoIsExcuse(c) {
		return "Excuse"
	}
	if scartoIsTrump(c) {
		return fmt.Sprintf("T%d", c.GetValue())
	}
	suits := map[int]string{
		CardDesignSpade:   "♠",
		CardDesignClover:  "♣",
		CardDesignHeart:   "♥",
		CardDesignDiamond: "♦",
	}
	s, ok := suits[c.GetDesign()]
	if !ok {
		s = "?"
	}
	return fmt.Sprintf("%s%d", s, c.GetValue())
}

// scartoFilter 述語を満たすインデックスを抽出する。
func scartoFilter(indices []int, pred func(int) bool) []int {
	var out []int
	for _, idx := range indices {
		if pred(idx) {
			out = append(out, idx)
		}
	}
	return out
}

// scartoAppendUnique スライスに未含有のインデックスを追加する。
func scartoAppendUnique(indices []int, idx int) []int {
	for _, v := range indices {
		if v == idx {
			return indices
		}
	}
	return append(indices, idx)
}

// --- State getters / setters ---

// GetPhase 現在のフェーズ取得
func (g *Scarto) GetPhase() ScartoPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Scarto) SetPhase(phase ScartoPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Scarto) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Scarto) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *Scarto) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Scarto) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Scarto) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Scarto) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Scarto) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Scarto) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Scarto) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Scarto) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx 親 (scartatore) インデックス取得
func (g *Scarto) GetDealerIdx() int { return g.dealerIdx }

// SetDealerIdx 親インデックス設定 (テスト用)
func (g *Scarto) SetDealerIdx(idx int) { g.dealerIdx = idx }

// GetScartoCount 親が捨てたスカルト札の枚数取得
func (g *Scarto) GetScartoCount() int { return len(g.scarto) }

// GetScarto スカルト札取得 (テスト用)
func (g *Scarto) GetScarto() []*Card { return g.scarto }

// SetScarto スカルト札設定 (テスト用)
func (g *Scarto) SetScarto(cards []*Card) { g.scarto = cards }

// GetPlayerScores プレイヤー別累積得点取得
func (g *Scarto) GetPlayerScores() [ScartoPlayerCnt]int { return g.playerScores }

// SetPlayerScores プレイヤー別累積得点設定 (テスト用)
func (g *Scarto) SetPlayerScores(s [ScartoPlayerCnt]int) { g.playerScores = s }

// GetDealScores 直近ディールの精算値取得
func (g *Scarto) GetDealScores() [ScartoPlayerCnt]int { return g.dealScores }

// GetCardPoints プレイヤー i が獲得したハーフポイント合計を返す (表示用)。
// 親のスカルト札も親の分に加算する。
func (g *Scarto) GetCardPoints(i int) int {
	if i < 0 || i >= len(g.players) {
		return 0
	}
	sum := 0
	for _, trick := range g.players[i].GetTricksTaken() {
		for _, c := range trick {
			sum += scartoCardHalfPoints(c)
		}
	}
	if i == g.dealerIdx {
		for _, c := range g.scarto {
			sum += scartoCardHalfPoints(c)
		}
	}
	return sum
}

// GetOutcome 直近ディールの結果取得
func (g *Scarto) GetOutcome() ScartoOutcome { return g.outcome }

// GetResult 人間視点のマッチ結果取得
func (g *Scarto) GetResult() ScartoResult { return g.result }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Scarto) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝利プレイヤー取得 (-1=未確定)
func (g *Scarto) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayerCnt プレイヤー数取得
func (g *Scarto) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Scarto) GetPlayer(i int) *ScartoPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番 (Play) が人間か。
func (g *Scarto) IsHumanTurn() bool {
	if g.phase != ScartoPhasePlay {
		return false
	}
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// IsHumanScartoTurn 現在のスカルト手番が人間 (=人間が親) か。
func (g *Scarto) IsHumanScartoTurn() bool {
	if g.phase != ScartoPhaseScarto {
		return false
	}
	return g.isHumanScartoTurn()
}

// GetConfig 設定取得
func (g *Scarto) GetConfig() ScartoConfig { return g.config }

// SetConfig 設定変更
func (g *Scarto) SetConfig(cfg ScartoConfig) { g.config = cfg }

// GetActionLog 棋譜取得
func (g *Scarto) GetActionLog() []*ActionLogEntry {
	return sliceOrEmpty(g.actionLog)
}

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Scarto) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != ScartoPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- Test / helper public wrappers ---

// TrickWinnerPublic 現在のトリックの勝者を返す (テスト用)。
func (g *Scarto) TrickWinnerPublic() int { return g.trickWinner() }

// LedSuitPublic 現在のトリックのリードスートを返す (テスト用)。
func (g *Scarto) LedSuitPublic() int { return g.ledSuit() }

// ScartoSettleDeal はディール精算の純粋関数の公開ラッパー (テスト用)。
func ScartoSettleDeal(half [ScartoPlayerCnt]int) [ScartoPlayerCnt]int { return scartoSettleDeal(half) }

// ScartoCardHalfPointsPublic はカードのハーフポイントを返す (テスト用)。
func ScartoCardHalfPointsPublic(c *Card) int { return scartoCardHalfPoints(c) }

// ScartoIsBoutPublic はカードがブーか返す (テスト用)。
func ScartoIsBoutPublic(c *Card) bool { return scartoIsBout(c) }

// ScartoIsTrumpPublic はカードが切り札か返す (テスト用)。
func ScartoIsTrumpPublic(c *Card) bool { return scartoIsTrump(c) }

// ScartoIsExcusePublic はカードがエクスキューズか返す (テスト用)。
func ScartoIsExcusePublic(c *Card) bool { return scartoIsExcuse(c) }

// ScartoDiscardablePublic はカードが通常スカルトに出せるか返す (テスト用)。
func ScartoDiscardablePublic(c *Card) bool { return scartoDiscardable(c) }

// BuildScartoDeckPublic は 78 枚デッキを構築する (テスト用)。
func BuildScartoDeckPublic() []*Card { return buildScartoDeck() }

// --- JSON ---

// scartoJSON is the JSON wire format for Scarto.
type scartoJSON struct {
	Deck             []*Card              `json:"dk"`
	DeckDrawCnt      int                  `json:"dw"`
	Players          []*ScartoPlayer      `json:"ps"`
	Config           ScartoConfig         `json:"cf"`
	Phase            ScartoPhase          `json:"ph"`
	RoundNumber      int                  `json:"rn"`
	TrickNumber      int                  `json:"tn"`
	CurrentPlayerIdx int                  `json:"ci"`
	CurrentTrick     []*TrickCard         `json:"ct"`
	LeadPlayerIdx    int                  `json:"li"`
	DealerIdx        int                  `json:"di"`
	Scarto           []*Card              `json:"sk"`
	PlayerScores     [ScartoPlayerCnt]int `json:"sc"`
	DealScores       [ScartoPlayerCnt]int `json:"ds"`
	LastTrickWinner  int                  `json:"lt"`
	Outcome          ScartoOutcome        `json:"oc"`
	Result           ScartoResult         `json:"rs"`
	Scored           bool                 `json:"sd"`
	GameEndFlag      bool                 `json:"ge"`
	WinnerPlayer     int                  `json:"wp"`
	ActionLog        []*ActionLogEntry    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Scarto) MarshalJSON() ([]byte, error) {
	return json.Marshal(scartoJSON{
		Deck:             g.deck,
		DeckDrawCnt:      g.deckDrawCnt,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		LeadPlayerIdx:    g.leadPlayerIdx,
		DealerIdx:        g.dealerIdx,
		Scarto:           g.scarto,
		PlayerScores:     g.playerScores,
		DealScores:       g.dealScores,
		LastTrickWinner:  g.lastTrickWinner,
		Outcome:          g.outcome,
		Result:           g.result,
		Scored:           g.scored,
		GameEndFlag:      g.gameEndFlag,
		WinnerPlayer:     g.winnerPlayer,
		ActionLog:        g.actionLog,
	})
}

// scartoMaxSliceLen caps slice sizes during deserialisation.
const scartoMaxSliceLen = 5000

// 各種デシリアライズ検証エラー。
var (
	errScartoOversized      = errors.New("scarto: input array exceeds maximum allowed size")
	errScartoInvalidPlayers = errors.New("scarto: invalid player count")
	errScartoInvalidTrick   = errors.New("scarto: invalid trick card")
	errScartoInvalidCard    = errors.New("scarto: invalid card element")
	errScartoInvalidIndex   = errors.New("scarto: index field out of range")
	errScartoInvalidPhase   = errors.New("scarto: phase out of range")
	errScartoInvalidOutcome = errors.New("scarto: outcome/result value out of range")
)

// scartoValidCard デシリアライズ時のカード妥当性を検証する (nil 拒否, 値域チェック)。
func scartoValidCard(c *Card) bool {
	if c == nil {
		return false
	}
	d, v := c.GetDesign(), c.GetValue()
	switch d {
	case ScartoExcuseDesign:
		return v == ScartoExcuseValue
	case ScartoTrumpDesign:
		return v >= 1 && v <= ScartoMaxTrump
	default:
		return d >= 1 && d <= ScartoSuitCnt && v >= 1 && v <= ScartoKingValue
	}
}

// scartoCheckCards スライスの各要素のカード妥当性を検証する。
func scartoCheckCards(cards []*Card) error {
	for _, c := range cards {
		if !scartoValidCard(c) {
			return errScartoInvalidCard
		}
	}
	return nil
}

// scartoInRange v が [0, PlayerCnt) か。
func scartoInRange(v int) bool { return v >= 0 && v < ScartoPlayerCnt }

// scartoInRangeOrUnset v が -1 (未設定) または [0, PlayerCnt) か。
func scartoInRangeOrUnset(v int) bool { return v == -1 || scartoInRange(v) }

// UnmarshalJSON implements json.Unmarshaler.
func (g *Scarto) UnmarshalJSON(data []byte) error {
	var j scartoJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > scartoMaxSliceLen || len(j.CurrentTrick) > scartoMaxSliceLen ||
		len(j.ActionLog) > scartoMaxSliceLen || len(j.Scarto) > scartoMaxSliceLen ||
		len(j.Deck) > scartoMaxSliceLen {
		return errScartoOversized
	}
	if len(j.Players) != ScartoPlayerCnt {
		return errScartoInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errScartoInvalidPlayers
		}
	}
	for _, c := range j.Deck {
		if !scartoValidCard(c) {
			return errScartoInvalidCard
		}
	}
	if err := scartoCheckCards(j.Scarto); err != nil {
		return err
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || !scartoValidCard(tc.Card) {
			return errScartoInvalidTrick
		}
		if !scartoInRange(tc.PlayerIdx) {
			return errScartoInvalidTrick
		}
	}
	if !scartoInRange(j.CurrentPlayerIdx) || !scartoInRange(j.DealerIdx) {
		return errScartoInvalidIndex
	}
	if !scartoInRangeOrUnset(j.LeadPlayerIdx) || !scartoInRangeOrUnset(j.LastTrickWinner) ||
		!scartoInRangeOrUnset(j.WinnerPlayer) {
		return errScartoInvalidIndex
	}
	if int(j.Phase) < ScartoPhaseMin || int(j.Phase) > ScartoPhaseMax {
		return errScartoInvalidPhase
	}
	// プレイ以降はリードプレイヤーが確定していなければならない。
	if j.Phase >= ScartoPhasePlay && j.Phase <= ScartoPhaseRoundEnd {
		if !scartoInRange(j.LeadPlayerIdx) {
			return errScartoInvalidIndex
		}
	}
	if j.Outcome < ScartoOutcomeNone || j.Outcome > ScartoOutcomeLoss {
		return errScartoInvalidOutcome
	}
	if j.Result < ScartoResultLose || j.Result > ScartoResultWin {
		return errScartoInvalidOutcome
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.deck = j.Deck
	if g.deck == nil {
		g.deck = make([]*Card, 0)
	}
	g.deckDrawCnt = j.DeckDrawCnt
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
	g.scarto = j.Scarto
	if g.scarto == nil {
		g.scarto = make([]*Card, 0)
	}
	g.playerScores = j.PlayerScores
	g.dealScores = j.DealScores
	g.lastTrickWinner = j.LastTrickWinner
	g.outcome = j.Outcome
	g.result = j.Result
	g.scored = j.Scored
	g.gameEndFlag = j.GameEndFlag
	g.winnerPlayer = j.WinnerPlayer
	g.actionLog = j.ActionLog
	return nil
}
