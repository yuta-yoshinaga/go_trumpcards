//go:build !js || !wasm || extra3

// Package domain プリメロ (Primero / Primiera) のドメインモデル。
//
// Primero はルネサンス期 (16 世紀) に広まったヴァイイング (vying) 系ポットゲームで、
// ポーカーの祖の一つ。40 枚デッキ (A,2,3,4,5,6,7,J,Q,K の 10 ランク × 4 スート) を使い、
// 2〜6 人が全員アンティを入れてから 4 枚ずつ配られる (共有カードはない)。
//
// # 1 ラウンドの流れ
//
//  1. 全員がアンティをポットに払い、4 枚ずつ配られる。
//  2. ディーラーの左隣から時計回りにベッティング: 各アクティブプレイヤーは
//     「コール (現在の賭けに合わせる)」「レイズ / ヴィ (賭けをアンティ分だけ上げる;
//     1 ラウンドあたり最大 PrimeroMaxRaises 回)」「フォールド (降りる)」を選ぶ。
//  3. 全員がフォールドせずに現在の賭けにマッチしたらベッティング終了。1 人だけ残れば
//     その 1 人がポットを総取り (クリーンウィン)。2 人以上残ればショーダウン。
//  4. ショーダウンでは最強手がポットを総取りする。
//
// # プライムポイント (古典 Primero)
//
//	7 → 21, 6 → 18, A(1) → 16, 5 → 15, 4 → 14, 3 → 13, 2 → 12, 絵札 J/Q/K → 10。
//
// # 手役 (CLEAN 解釈; 高い順)
//
//   - フルクサス (Fluxus / フラッシュ): 4 枚すべて同スート。プライムポイント合計で比較。
//   - スプレムス (Supremus): プリメロ (4 スートすべて 1 枚ずつ) かつポイント合計 >= 50。合計で比較。
//   - プリメロ (Primero): 4 スートすべて 1 枚ずつ、かつポイント合計 < 50。合計で比較。
//   - ヌメルス (Numerus): 上記以外。スコア = 最も価値の高い単一スートのポイント合計。
//     そのスコア、次いで総合計で比較。
//   - カテゴリ順: Fluxus(3) > Supremus(2) > Primero(1) > Numerus(0)。完全同点は
//     座席番号が最も小さいプレイヤーがポットを総取りする (決定的ルール)。
//
// # デッキ
//
// 40 枚 (A,2,3,4,5,6,7,J,Q,K × 4 スート)。newPrimeroDeck は標準 52 枚 (NewTrumpCards(0)) を
// 8,9,10 を除いてフィルタして構築する。
//
// 本実装は extra ワーカーから到達可能なよう、手役評価・ベッティング・ポット精算ロジックを
// すべてインラインで持つ (Poker / ThreeCardBrag は casino ビルドタグで到達不可のため)。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// PrimeroPhase はゲームフェーズ。
type PrimeroPhase int

// Primero のフェーズ定数。ワイヤー値はフロントエンドの enum と一致させる。
const (
	// PrimeroPhaseBetting ベッティング中 (配札済み; 手番プレイヤーの call/raise/fold 待ち)。ワイヤー値 0。
	PrimeroPhaseBetting PrimeroPhase = 0
	// PrimeroPhaseResult ラウンド解決済み (結果表示; 次ラウンド待ち or ゲーム終了)。ワイヤー値 1。
	PrimeroPhaseResult PrimeroPhase = 1
)

// 手役カテゴリ (高いほど強い)。
const (
	// PrimeroHandNumerus ヌメルス (上記いずれでもない)
	PrimeroHandNumerus = 0
	// PrimeroHandPrimero プリメロ (4 スート 1 枚ずつ、合計 < 50)
	PrimeroHandPrimero = 1
	// PrimeroHandSupremus スプレムス (プリメロで合計 >= 50)
	PrimeroHandSupremus = 2
	// PrimeroHandFluxus フルクサス (4 枚同スート)
	PrimeroHandFluxus = 3
)

// primeroSupremusThreshold は Supremus/Primero を分けるポイント合計しきい値。
const primeroSupremusThreshold = 50

// PrimeroResult は人間プレイヤーから見たラウンド結果。
// GameResult は共有ファイル internal/domain/game_result.go に移動したので到達可能に
// なったが、この型名は JSON ペイロードに出るため統合していない（#4462）。値は
// GameResult と同一。
type PrimeroResult int

const (
	// PrimeroResultLose 負け (ベッティングに残ったが敗北)
	PrimeroResultLose PrimeroResult = -1
	// PrimeroResultNone 結果なし (フォールド / 未参加 / 未解決)
	PrimeroResultNone PrimeroResult = 0
	// PrimeroResultWin 勝ち (ポット獲得)
	PrimeroResultWin PrimeroResult = 1
)

// primeroDeckSize は 40 枚デッキのサイズ (A,2,3,4,5,6,7,J,Q,K × 4 スート)。
const primeroDeckSize = 40

// primeroMaxSliceLen はデシリアライズ時のスライス長の上限。
const primeroMaxSliceLen = 5000

// primeroMaxActions は 1 ラウンドあたりの賭けアクション上限 (安全網)。
const primeroMaxActions = 60

// CPU AI の手役判断しきい値 (ヌメルスの最強単一スートポイント合計)。
const (
	primeroCpuRaiseScore = 40 // 強い単一スート (ヌメルス) でレイズを検討
	primeroCpuCallScore  = 28 // そこそこの単一スート (ヌメルス) でコール
)

// デシリアライズ検証用のセンチネルエラー。
var (
	errPrimeroSnapshot      = errors.New("primero: invalid serialised game state")
	errPrimeroInvalidPlayer = errors.New("primero: invalid player state")
)

// PrimeroHint はヒント情報 (人間への call/raise/fold 助言)。
type PrimeroHint struct {
	Action string // 推奨アクション ("call"/"raise"/"fold")
	Reason string // ヒント理由キー
}

// primeroState はゲーム進行状態。
type primeroState struct {
	phase           PrimeroPhase
	roundNumber     int
	dealerIdx       int
	currentPlayer   int
	pot             int
	currentBet      int // 現在の必要総拠出額 (アンティ = 開始値)。call でこの額に合わせる
	raiseCount      int // このラウンドのレイズ回数
	actedSinceRaise int // 直近のレイズ (または配札) 以降にアクションしたアクティブ人数
	actionCount     int // 総アクション数 (安全網)
	winnerIdx       int // 直近ラウンドの勝者 (-1 = なし)
	matchWinnerIdx  int // ゲーム全体の勝者 (-1 = 未確定)
	result          PrimeroResult
	gameEndFlag     bool
	scored          bool // ラウンド結果を確定済みか (二重確定防止)
	actionLogBase
}

// Primero はプリメロの状態を保持する集約ルート。
type Primero struct {
	trumpCards *TrumpCards
	players    []*PrimeroPlayer
	config     PrimeroConfig
	state      primeroState
}

// NewPrimero はコンストラクタ。
func NewPrimero(trumpCards *TrumpCards, players []*PrimeroPlayer, config PrimeroConfig) *Primero {
	return &Primero{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		state: primeroState{
			phase:          PrimeroPhaseBetting,
			winnerIdx:      -1,
			matchWinnerIdx: -1,
			actionLogBase:  actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
		},
	}
}

// NewDefaultPrimero は標準構成 (人間 seat 0 + CPU) を生成する。CUI / Web / Worker 構築の単一情報源。
func NewDefaultPrimero() *Primero {
	cfg := DefaultPrimeroConfig()
	g := NewPrimero(newPrimeroDeck(), primeroNewPlayers(cfg), cfg)
	g.Reset()
	return g
}

// newPrimeroDeck は標準 52 枚 (NewTrumpCards(0)) から 8,9,10 を除いて 40 枚デッキを構築する。
func newPrimeroDeck() *TrumpCards {
	drop := map[int]bool{8: true, 9: true, 10: true}
	full := NewTrumpCards(0)
	filtered := make([]*Card, 0, primeroDeckSize)
	for _, c := range full.deck {
		if !drop[c.GetValue()] {
			filtered = append(filtered, c)
		}
	}
	full.deck = filtered
	full.deckCnt = len(filtered)
	full.deckInit()
	return full
}

// primeroNewPlayers は設定に基づいてプレイヤー列を生成する (seat 0 = 人間)。
func primeroNewPlayers(cfg PrimeroConfig) []*PrimeroPlayer {
	players := make([]*PrimeroPlayer, cfg.PlayerCount)
	for i := range players {
		players[i] = NewPrimeroPlayer(i == 0, cfg.StartingChips)
	}
	return players
}

// --- ゲーム進行 ---

// Reset は新しいゲームを開始する。チップ・プレイヤー数を設定から作り直し、第 1 ラウンドを配る。
func (g *Primero) Reset() {
	g.players = primeroNewPlayers(g.config)
	g.trumpCards = newPrimeroDeck()
	g.state = primeroState{
		phase:          PrimeroPhaseBetting,
		roundNumber:    1,
		dealerIdx:      0,
		winnerIdx:      -1,
		matchWinnerIdx: -1,
		actionLogBase:  actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
	}
	g.startRound()
}

// NextRound は同じチップを保持したまま次のラウンドを配る。Result フェーズかつゲーム
// 継続中のときのみ有効。
func (g *Primero) NextRound() {
	if g.state.phase != PrimeroPhaseResult || g.state.gameEndFlag {
		return
	}
	g.state.roundNumber++
	g.state.dealerIdx = (g.state.dealerIdx + 1) % len(g.players)
	g.startRound()
}

// startRound は 1 ラウンドを準備する: 脱落判定・アンティ徴収・配札・ベッティング開始。
func (g *Primero) startRound() {
	// ラウンド単位の状態をクリア。
	g.state.winnerIdx = -1
	g.state.result = PrimeroResultNone
	g.state.pot = 0
	g.state.currentBet = 0
	g.state.raiseCount = 0
	g.state.actedSinceRaise = 0
	g.state.actionCount = 0
	g.state.scored = false
	for _, p := range g.players {
		p.ResetForRound()
	}
	// アンティを払えないプレイヤーは脱落。
	for _, p := range g.players {
		if !p.GetOut() && p.GetChips() < g.config.Ante {
			p.SetOut(true)
		}
	}
	// 参加可能なプレイヤーが 2 人未満、または人間が脱落 → ゲーム終了。
	if g.solventCount() < 2 || g.players[0].GetOut() {
		g.endGame()
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
		g.state.pot += g.config.Ante
		for i := 0; i < PrimeroHandSize; i++ {
			if c := g.trumpCards.DrawCard(); c != nil {
				p.AddCard(c)
			}
		}
	}
	g.state.currentBet = g.config.Ante
	g.state.phase = PrimeroPhaseBetting
	g.state.currentPlayer = g.nextActive(g.state.dealerIdx)
	g.appendLog(-1, "deal",
		fmt.Sprintf("Round %d: ante %d, pot %d", g.state.roundNumber, g.config.Ante, g.state.pot), nil)
	// 最初の手番が CPU なら人間の手番になるまで進める。
	g.driveCPU()
}

// --- Actions (human) ---

// PlayerCall は人間 (現在の手番) が現在の賭けにコール (マッチ) する。
func (g *Primero) PlayerCall() error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	g.applyCall(g.state.currentPlayer)
	g.driveCPU()
	return nil
}

// PlayerRaise は人間 (現在の手番) が賭けをアンティ分だけ引き上げる (ヴィ)。
func (g *Primero) PlayerRaise() error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	if g.state.raiseCount >= PrimeroMaxRaises {
		return NewDomainError(ErrInvalidPlay, "no more raises are allowed this round")
	}
	g.applyRaise(g.state.currentPlayer)
	g.driveCPU()
	return nil
}

// PlayerFold は人間 (現在の手番) が降りる。
func (g *Primero) PlayerFold() error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	g.applyFold(g.state.currentPlayer)
	g.driveCPU()
	return nil
}

// checkTurn は現在の手番が人間かつベッティングフェーズかを検証する。
func (g *Primero) checkTurn() error {
	if g.state.gameEndFlag {
		return NewDomainError(ErrGameEnded, "the game has already ended")
	}
	if g.state.phase != PrimeroPhaseBetting {
		return NewDomainError(ErrWrongPhase, "betting is only allowed during the betting phase")
	}
	if !g.players[g.state.currentPlayer].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "it is not your turn")
	}
	return nil
}

// applyCall はコール (現在の賭けにマッチ) を適用する。チップ不足なら自動フォールド。
func (g *Primero) applyCall(idx int) {
	p := g.players[idx]
	need := g.state.currentBet - p.GetRoundBet()
	if need < 0 {
		need = 0
	}
	if need > 0 && p.GetChips() < need {
		// 払えない場合はフォールド扱い。
		g.applyFold(idx)
		return
	}
	if need > 0 {
		p.SubtractChips(need)
		p.AddRoundBet(need)
		g.state.pot += need
	}
	g.state.actedSinceRaise++
	g.state.actionCount++
	g.appendLog(idx, "call", fmt.Sprintf("%s calls %d (pot %d)", playerName(g.players, idx), need, g.state.pot), nil)
	g.advanceOrClose(idx)
}

// applyRaise はレイズ (ヴィ) を適用する。増分はアンティに等しい。払えなければコールに退避する。
func (g *Primero) applyRaise(idx int) {
	p := g.players[idx]
	newBet := g.state.currentBet + g.config.Ante
	need := newBet - p.GetRoundBet()
	if p.GetChips() < need {
		// レイズを払えない場合はコール (さらに払えなければフォールド) に退避。
		g.applyCall(idx)
		return
	}
	g.state.currentBet = newBet
	g.state.raiseCount++
	p.SubtractChips(need)
	p.AddRoundBet(need)
	g.state.pot += need
	// レイズ後は本人を含め 1 人がアクション済み。他のアクティブは応答が必要。
	g.state.actedSinceRaise = 1
	g.state.actionCount++
	g.appendLog(idx, "raise", fmt.Sprintf("%s raises to %d, pays %d (pot %d)", playerName(g.players, idx), newBet, need, g.state.pot), nil)
	g.advanceOrClose(idx)
}

// applyFold はフォールドを適用する。
func (g *Primero) applyFold(idx int) {
	g.players[idx].SetFolded(true)
	g.state.actionCount++
	g.appendLog(idx, "fold", fmt.Sprintf("%s folds", playerName(g.players, idx)), nil)
	g.advanceOrClose(idx)
}

// advanceOrClose は次の手番へ進めるか、ベッティングを締めて精算する。
func (g *Primero) advanceOrClose(fromIdx int) {
	// 残り 1 人ならクリーンウィン、全アクティブがアクション済みならショーダウン。
	if g.activeCount() <= 1 ||
		g.state.actedSinceRaise >= g.activeCount() ||
		g.state.actionCount >= primeroMaxActions {
		g.resolveRound()
		return
	}
	g.state.currentPlayer = g.nextActive(fromIdx)
}

// resolveRound はベッティング終了時にポットを精算し、フェーズを Result へ遷移する。
// scored フラグで二重精算を防ぐ (フェーズ入場時に 1 回だけ発火)。
func (g *Primero) resolveRound() {
	if g.state.scored {
		return
	}
	active := g.activeSeats()
	// 初期値 -1 は len(active)==0 の安全網 (通常は発生しない: 最後のフォールドは
	// クリーンウィンになる)。1 人ならその席、複数ならベストハンドが勝者。
	winner := -1
	if len(active) == 1 {
		winner = active[0]
	} else if len(active) >= 2 {
		winner = g.bestHand(active)
	}
	g.state.winnerIdx = winner
	if winner >= 0 {
		g.players[winner].AddChips(g.state.pot)
		g.appendLog(winner, "win",
			fmt.Sprintf("%s wins the pot (%d)", playerName(g.players, winner), g.state.pot), nil)
	}
	g.setHumanResult(winner)
	g.state.pot = 0
	g.state.scored = true
	g.state.phase = PrimeroPhaseResult
	g.checkGameEnd()
}

// setHumanResult は人間 (seat 0) の勝敗結果を設定する。
func (g *Primero) setHumanResult(winnerIdx int) {
	human := g.players[0]
	if human.GetOut() || human.GetFolded() {
		g.state.result = PrimeroResultNone
		return
	}
	if winnerIdx == 0 {
		g.state.result = PrimeroResultWin
	} else {
		g.state.result = PrimeroResultLose
	}
}

// checkGameEnd は停止条件 (規定ラウンド到達 or 参加可能者 2 人未満) を判定し、
// 満たせばゲームを終了させる。
func (g *Primero) checkGameEnd() {
	if g.state.roundNumber >= g.config.TargetRounds || g.solventCount() < 2 {
		g.endGame()
	}
}

// endGame はゲームを終了し、チップ最多のプレイヤーを勝者に設定する。
func (g *Primero) endGame() {
	g.state.gameEndFlag = true
	g.state.phase = PrimeroPhaseResult
	g.state.matchWinnerIdx = g.richestIdx()
	g.appendLog(g.state.matchWinnerIdx, "game_end",
		fmt.Sprintf("%s wins the game", playerName(g.players, g.state.matchWinnerIdx)), nil)
}

// --- CPU ---

// driveCPU は現在の手番が CPU の間、CPU のアクションを実行し続ける。人間の手番か
// ラウンド解決で停止する。
func (g *Primero) driveCPU() {
	guard := 0
	for g.state.phase == PrimeroPhaseBetting && !g.state.gameEndFlag {
		idx := g.state.currentPlayer
		if g.players[idx].GetIsHuman() {
			return
		}
		g.cpuAct(idx)
		guard++
		if guard > primeroMaxActions*2 {
			// 安全網 (通常到達しない): 強制的にラウンドを締める。
			g.resolveRound()
			return
		}
	}
}

// cpuAct は CPU が 1 アクション (call/raise/fold) を選ぶ。
func (g *Primero) cpuAct(idx int) {
	p := g.players[idx]
	cat, tb := g.evalPlayer(idx)
	need := g.state.currentBet - p.GetRoundBet()
	if need < 0 {
		need = 0
	}
	canRaise := g.state.raiseCount < PrimeroMaxRaises && p.GetChips() >= need+g.config.Ante
	score := 0
	if len(tb) > 0 {
		score = tb[0]
	}
	switch {
	case cat >= PrimeroHandSupremus:
		// 非常に強い (フルクサス / スプレムス): 可能ならレイズ、無理ならコール。
		if canRaise {
			g.applyRaise(idx)
			return
		}
		g.applyCall(idx)
	case cat == PrimeroHandPrimero:
		// 強い (プリメロ): レイズ控えめ、そうでなければコール。
		if canRaise && g.state.raiseCount < 2 {
			g.applyRaise(idx)
			return
		}
		g.applyCall(idx)
	case need == 0:
		// タダで残れる (チェック相当) → コール。
		g.applyCall(idx)
	case score >= primeroCpuCallScore && need <= g.config.Ante*2:
		// 中程度 (強い単一スート) かつ安い → コール。
		g.applyCall(idx)
	default:
		// 弱い手で賭けが必要 → フォールド。
		g.applyFold(idx)
	}
}

// --- 手役評価 (インライン) ---

// primeroPoints はカードのプライムポイントを返す
// (7→21, 6→18, A→16, 5→15, 4→14, 3→13, 2→12, 絵札→10)。
func primeroPoints(c *Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 7:
		return 21
	case 6:
		return 18
	case 1:
		return 16
	case 5:
		return 15
	case 4:
		return 14
	case 3:
		return 13
	case 2:
		return 12
	case 11, 12, 13:
		return 10
	default:
		return 0
	}
}

// PrimeroEval は 4 枚の手の役カテゴリと比較用タイブレーク列を返す。
// カテゴリが同じならこの列を辞書順比較すれば勝敗が決まる。
//
//   - Fluxus:   タイブレーク = [プライムポイント合計]
//   - Supremus: タイブレーク = [プライムポイント合計]
//   - Primero:  タイブレーク = [プライムポイント合計]
//   - Numerus:  タイブレーク = [最強単一スートのポイント合計, 総合計]
func PrimeroEval(cards []*Card) (int, []int) {
	if len(cards) != PrimeroHandSize {
		return -1, nil
	}
	// スート (design 1..4) ごとのポイント合計。index 0 は未使用。
	var suitSum [5]int
	var suitCount [5]int
	total := 0
	distinct := 0
	for _, c := range cards {
		if c == nil {
			continue
		}
		pts := primeroPoints(c)
		total += pts
		d := c.GetDesign()
		if d < 1 || d > 4 {
			continue
		}
		if suitCount[d] == 0 {
			distinct++
		}
		suitCount[d]++
		suitSum[d] += pts
	}
	switch distinct {
	case 1:
		// フルクサス: 4 枚同スート。
		return PrimeroHandFluxus, []int{total}
	case 4:
		// 4 スート 1 枚ずつ: 合計で Supremus / Primero を判定。
		if total >= primeroSupremusThreshold {
			return PrimeroHandSupremus, []int{total}
		}
		return PrimeroHandPrimero, []int{total}
	default:
		// ヌメルス: 最強単一スートのポイント合計。
		maxSuit := 0
		for d := 1; d <= 4; d++ {
			if suitSum[d] > maxSuit {
				maxSuit = suitSum[d]
			}
		}
		return PrimeroHandNumerus, []int{maxSuit, total}
	}
}

// evalPlayer は playerIdx の手役を返す。
func (g *Primero) evalPlayer(idx int) (int, []int) {
	p := g.players[idx]
	cards := make([]*Card, 0, PrimeroHandSize)
	for i := 0; i < p.GetCardsSize(); i++ {
		cards = append(cards, p.GetCard(i))
	}
	return PrimeroEval(cards)
}

// PrimeroCompare は手 a が手 b に勝てば 1、負ければ -1、引き分けは 0 を返す。
func PrimeroCompare(catA int, tbA []int, catB int, tbB []int) int {
	if catA != catB {
		if catA > catB {
			return 1
		}
		return -1
	}
	for i := 0; i < len(tbA) && i < len(tbB); i++ {
		if tbA[i] != tbB[i] {
			if tbA[i] > tbB[i] {
				return 1
			}
			return -1
		}
	}
	return 0
}

// bestHand はアクティブプレイヤーのうち最強手のインデックスを返す。完全同点は
// 座席番号の最も小さいプレイヤーが総取りする (seats は昇順、strictly-greater でのみ更新)。
func (g *Primero) bestHand(seats []int) int {
	best := seats[0]
	bestCat, bestTb := g.evalPlayer(best)
	for _, idx := range seats[1:] {
		cat, tb := g.evalPlayer(idx)
		if PrimeroCompare(cat, tb, bestCat, bestTb) > 0 {
			best, bestCat, bestTb = idx, cat, tb
		}
	}
	return best
}

// --- ヘルパー ---

// activeSeats は未フォールド・非脱落プレイヤーのインデックス列 (昇順) を返す。
func (g *Primero) activeSeats() []int {
	out := make([]int, 0, len(g.players))
	for i, p := range g.players {
		if !p.GetOut() && !p.GetFolded() {
			out = append(out, i)
		}
	}
	return out
}

// activeCount は未フォールド・非脱落プレイヤー数を返す。
func (g *Primero) activeCount() int {
	n := 0
	for _, p := range g.players {
		if !p.GetOut() && !p.GetFolded() {
			n++
		}
	}
	return n
}

// nextActive は from の次の未フォールド・非脱落プレイヤーを返す。
func (g *Primero) nextActive(from int) int {
	n := len(g.players)
	for i := 1; i <= n; i++ {
		idx := (from + i) % n
		if !g.players[idx].GetOut() && !g.players[idx].GetFolded() {
			return idx
		}
	}
	return from
}

// solventCount はアンティを払える (非脱落かつチップ >= アンティ) プレイヤー数を返す。
func (g *Primero) solventCount() int {
	n := 0
	for _, p := range g.players {
		if !p.GetOut() && p.GetChips() >= g.config.Ante {
			n++
		}
	}
	return n
}

// richestIdx はチップが最多のプレイヤーのインデックスを返す (同数は座席番号の小さい方)。
func (g *Primero) richestIdx() int {
	best := 0
	for i, p := range g.players {
		if p.GetChips() > g.players[best].GetChips() {
			best = i
		}
	}
	return best
}

func (g *Primero) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.state.appendLog(playerIdx, actionType, detail, cards)
}

// --- Hint ---

// GetHint はベッティング中の人間 (seat 0 の手番) に call/raise/fold の助言を返す。
func (g *Primero) GetHint() *PrimeroHint {
	if g.state.phase != PrimeroPhaseBetting || g.state.gameEndFlag || g.state.currentPlayer != 0 {
		return nil
	}
	p := g.players[0]
	if p.GetOut() || p.GetFolded() {
		return nil
	}
	cat, tb := g.evalPlayer(0)
	score := 0
	if len(tb) > 0 {
		score = tb[0]
	}
	need := g.state.currentBet - p.GetRoundBet()
	switch {
	case cat >= PrimeroHandSupremus:
		return &PrimeroHint{Action: "raise", Reason: "strong_hand"}
	case cat == PrimeroHandPrimero:
		return &PrimeroHint{Action: "raise", Reason: "strong_hand"}
	case score >= primeroCpuRaiseScore:
		return &PrimeroHint{Action: "raise", Reason: "strong_hand"}
	case score >= primeroCpuCallScore:
		return &PrimeroHint{Action: "call", Reason: "medium_hand"}
	case need <= 0:
		return &PrimeroHint{Action: "call", Reason: "medium_hand"}
	default:
		return &PrimeroHint{Action: "fold", Reason: "weak_hand"}
	}
}

// --- 状態アクセサ ---

// GetPhase は現在のフェーズを返す。
func (g *Primero) GetPhase() PrimeroPhase { return g.state.phase }

// SetPhase はフェーズを設定する (テスト用)。
func (g *Primero) SetPhase(p PrimeroPhase) { g.state.phase = p }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *Primero) GetGameEndFlag() bool { return g.state.gameEndFlag }

// GetRoundNumber は現在のラウンド番号を返す。
func (g *Primero) GetRoundNumber() int { return g.state.roundNumber }

// GetDealerIdx はディーラーの座席番号を返す。
func (g *Primero) GetDealerIdx() int { return g.state.dealerIdx }

// SetDealerIdx はディーラーの座席番号を設定する (テスト用)。
func (g *Primero) SetDealerIdx(idx int) { g.state.dealerIdx = idx }

// GetCurrentPlayerIdx は現在の手番プレイヤーの座席番号を返す。
func (g *Primero) GetCurrentPlayerIdx() int { return g.state.currentPlayer }

// SetCurrentPlayerIdx は現在の手番プレイヤーを設定する (テスト用)。
func (g *Primero) SetCurrentPlayerIdx(idx int) { g.state.currentPlayer = idx }

// GetPot は現在のポットを返す。
func (g *Primero) GetPot() int { return g.state.pot }

// SetPot はポットを設定する (テスト用)。
func (g *Primero) SetPot(v int) { g.state.pot = v }

// GetCurrentBet は現在の必要総拠出額を返す。
func (g *Primero) GetCurrentBet() int { return g.state.currentBet }

// SetCurrentBet は現在の必要総拠出額を設定する (テスト用)。
func (g *Primero) SetCurrentBet(v int) { g.state.currentBet = v }

// GetRaiseCount はこのラウンドのレイズ回数を返す。
func (g *Primero) GetRaiseCount() int { return g.state.raiseCount }

// GetMaxRaises は 1 ラウンドあたりの最大レイズ回数を返す。
func (g *Primero) GetMaxRaises() int { return PrimeroMaxRaises }

// GetAnte はアンティ額を返す。
func (g *Primero) GetAnte() int { return g.config.Ante }

// GetWinnerIdx は直近ラウンドの勝者を返す (-1 = なし)。
func (g *Primero) GetWinnerIdx() int { return g.state.winnerIdx }

// GetMatchWinnerIdx はゲーム全体の勝者を返す (-1 = 未確定)。
func (g *Primero) GetMatchWinnerIdx() int { return g.state.matchWinnerIdx }

// GetResult は人間から見たラウンド結果を返す。
func (g *Primero) GetResult() PrimeroResult { return g.state.result }

// GetPlayerCnt はプレイヤー数を返す。
func (g *Primero) GetPlayerCnt() int { return len(g.players) }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *Primero) GetPlayer(i int) *PrimeroPlayer {
	return getPlayer(g.players, i)
}

// GetChips は人間 (seat 0) の保有チップを返す。
func (g *Primero) GetChips() int {
	if len(g.players) == 0 {
		return 0
	}
	return g.players[0].GetChips()
}

// IsHumanTurn は現在の手番が人間 (seat 0) かどうかを返す。
func (g *Primero) IsHumanTurn() bool {
	if g.state.phase != PrimeroPhaseBetting || g.state.gameEndFlag {
		return false
	}
	idx := g.state.currentPlayer
	if idx < 0 || idx >= len(g.players) {
		return false
	}
	return g.players[idx].GetIsHuman()
}

// CanRaise は現在の手番プレイヤーがレイズ可能か (回数上限未満かつ増分を払える) を返す。
func (g *Primero) CanRaise() bool {
	if !g.IsHumanTurn() || g.state.raiseCount >= PrimeroMaxRaises {
		return false
	}
	p := g.players[g.state.currentPlayer]
	need := (g.state.currentBet + g.config.Ante) - p.GetRoundBet()
	return p.GetChips() >= need
}

// GetConfig はローカルルール設定を返す。
func (g *Primero) GetConfig() PrimeroConfig { return g.config }

// SetConfig はローカルルール設定を変更する。
func (g *Primero) SetConfig(cfg PrimeroConfig) { g.config = cfg }

// GetActionLog は棋譜を返す。
func (g *Primero) GetActionLog() []*ActionLogEntry { return g.state.actionLog }

// ResolveForTest は手札を設定済みの状態でラウンドを解決する (テスト用)。乱数配札を
// 迂回して勝敗解決を決定的に検証するためのショートカット。
func (g *Primero) ResolveForTest() { g.resolveRound() }

// ClearBettingForTest は 1 ラウンドの賭けブックキーピング (raiseCount /
// actedSinceRaise / actionCount) をゼロに戻す (テスト用)。Reset() はディーラー左の
// CPU を人間の手番まで自動進行させるため残留カウンタが乗る。決定的な途中状態を
// 組み立てるテストはこれで残留をクリアする。
func (g *Primero) ClearBettingForTest() {
	g.state.raiseCount = 0
	g.state.actedSinceRaise = 0
	g.state.actionCount = 0
}

// --- JSON Serialization ---

// primeroJSON is the JSON wire format for Primero.
type primeroJSON struct {
	TrumpCards      *TrumpCards       `json:"tc"`
	Players         []*PrimeroPlayer  `json:"ps"`
	Config          PrimeroConfig     `json:"cf"`
	Phase           PrimeroPhase      `json:"ph"`
	RoundNumber     int               `json:"rn"`
	DealerIdx       int               `json:"di"`
	CurrentPlayer   int               `json:"ci"`
	Pot             int               `json:"pt"`
	CurrentBet      int               `json:"cb"`
	RaiseCount      int               `json:"rc"`
	ActedSinceRaise int               `json:"as"`
	ActionCount     int               `json:"an"`
	WinnerIdx       int               `json:"wi"`
	MatchWinnerIdx  int               `json:"mw"`
	Result          PrimeroResult     `json:"re"`
	GameEndFlag     bool              `json:"ge"`
	Scored          bool              `json:"sc"`
	ActionLog       []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Primero) MarshalJSON() ([]byte, error) {
	return json.Marshal(primeroJSON{
		TrumpCards:      g.trumpCards,
		Players:         g.players,
		Config:          g.config,
		Phase:           g.state.phase,
		RoundNumber:     g.state.roundNumber,
		DealerIdx:       g.state.dealerIdx,
		CurrentPlayer:   g.state.currentPlayer,
		Pot:             g.state.pot,
		CurrentBet:      g.state.currentBet,
		RaiseCount:      g.state.raiseCount,
		ActedSinceRaise: g.state.actedSinceRaise,
		ActionCount:     g.state.actionCount,
		WinnerIdx:       g.state.winnerIdx,
		MatchWinnerIdx:  g.state.matchWinnerIdx,
		Result:          g.state.result,
		GameEndFlag:     g.state.gameEndFlag,
		Scored:          g.state.scored,
		ActionLog:       g.state.actionLog,
	})
}

// primeroValidPhase は有効なフェーズかどうか。
func primeroValidPhase(p PrimeroPhase) bool {
	return p == PrimeroPhaseBetting || p == PrimeroPhaseResult
}

// UnmarshalJSON implements json.Unmarshaler. 不正な永続化データを拒否するための
// バリデーションを行う (KV 復元時のインデックス範囲・スライス健全性)。
func (g *Primero) UnmarshalJSON(data []byte) error {
	var j primeroJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("primero: invalid config: %w", err)
	}
	n := len(j.Players)
	if n < PrimeroMinPlayerCount || n > PrimeroMaxPlayerCount || n != j.Config.PlayerCount {
		return errPrimeroSnapshot
	}
	if len(j.ActionLog) > primeroMaxSliceLen {
		return errPrimeroSnapshot
	}
	if !primeroValidPhase(j.Phase) {
		return errPrimeroSnapshot
	}
	if j.RoundNumber < 1 || j.Pot < 0 || j.CurrentBet < 0 || j.RaiseCount < 0 ||
		j.ActedSinceRaise < 0 || j.ActionCount < 0 {
		return errPrimeroSnapshot
	}
	if j.DealerIdx < 0 || j.DealerIdx >= n || j.CurrentPlayer < 0 || j.CurrentPlayer >= n {
		return errPrimeroSnapshot
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= n || j.MatchWinnerIdx < -1 || j.MatchWinnerIdx >= n {
		return errPrimeroSnapshot
	}
	if j.Result < PrimeroResultLose || j.Result > PrimeroResultWin {
		return errPrimeroSnapshot
	}
	for _, p := range j.Players {
		if p == nil {
			return errPrimeroSnapshot
		}
	}
	for _, e := range j.ActionLog {
		if e == nil {
			return errPrimeroSnapshot
		}
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newPrimeroDeck()
	}
	g.players = j.Players
	g.config = j.Config
	g.state = primeroState{
		phase:           j.Phase,
		roundNumber:     j.RoundNumber,
		dealerIdx:       j.DealerIdx,
		currentPlayer:   j.CurrentPlayer,
		pot:             j.Pot,
		currentBet:      j.CurrentBet,
		raiseCount:      j.RaiseCount,
		actedSinceRaise: j.ActedSinceRaise,
		actionCount:     j.ActionCount,
		winnerIdx:       j.WinnerIdx,
		matchWinnerIdx:  j.MatchWinnerIdx,
		result:          j.Result,
		gameEndFlag:     j.GameEndFlag,
		scored:          j.Scored,
		actionLogBase:   actionLogBase{actionLog: j.ActionLog},
	}
	if g.state.actionLog == nil {
		g.state.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
