//go:build !js || !wasm || extra3

// Package domain ブイヨット (Bouillotte) のドメインモデル。
//
// Bouillotte は 18 世紀フランスのポーカーの祖となるヴァイイング (vying) 系ポットゲーム。
// 20 枚デッキ (A, K, Q, 9, 8 の 5 ランク × 4 スート) を使い、3〜4 人が全員アンティを
// 入れてから 3 枚ずつ配られ、さらに共有カード「retourne (ルトゥルヌ)」を 1 枚表向きに置く。
//
// # 1 ラウンドの流れ
//
//  1. 全員がアンティをポットに払い、3 枚ずつ配られ、retourne が 1 枚めくられる。
//  2. ディーラーの左隣から時計回りにベッティング: 各アクティブプレイヤーは
//     「コール (現在の賭けに合わせる)」「レイズ / ヴィ (賭けをアンティ分だけ上げる;
//     1 ラウンドあたり最大 BouillotteMaxRaises 回)」「フォールド (降りる)」を選ぶ。
//  3. 全員がフォールドせずに現在の賭けにマッチしたらベッティング終了。1 人だけ残れば
//     その 1 人がポットを総取り (クリーンウィン)。2 人以上残ればショーダウン。
//  4. ショーダウンでは最強手がポットを総取りする。
//
// # 手役 (CLEAN 解釈)
//
//   - ブルラン (Brelan / スリーカード) はすべてに勝つ。手札 3 枚が同ランクなら
//     「ブルラン・サンプル (simple)」、手札にペアがあり retourne がそのランクと一致
//     (retourne が 3 枚目を完成させる) なら「ブルラン・ファヴォリ (favori)」。同ランクの
//     サンプルに対してファヴォリが勝つ。ブルランはトリップのランクで比較する (A 最強)。
//   - ブルランでなければハイカード: 手札 3 枚 + retourne (共有カード) の中の最強カードで
//     比較し、同点は次点カードで比較する。スートは無関係。完全同点の場合は座席番号が
//     最も小さいプレイヤーがポットを総取りする (決定的ルール)。
//
// # デッキ
//
// 20 枚 (A, K, Q, 9, 8 × 4 スート)。newBouillotteDeck は標準 52 枚 (NewTrumpCards(0)) を
// これら 5 ランクにフィルタして構築する。
//
// 本実装は extra ワーカーから到達可能なよう、手役評価・ベッティング・ポット精算ロジックを
// すべてインラインで持つ (Poker / ThreeCardBrag は casino ビルドタグで到達不可のため)。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// BouillottePhase はゲームフェーズ。
type BouillottePhase int

// Bouillotte のフェーズ定数。ワイヤー値はフロントエンドの enum と一致させる。
const (
	// BouillottePhaseBetting ベッティング中 (配札済み; 手番プレイヤーの call/raise/fold 待ち)。ワイヤー値 0。
	BouillottePhaseBetting BouillottePhase = 0
	// BouillottePhaseResult ラウンド解決済み (結果表示; 次ラウンド待ち or ゲーム終了)。ワイヤー値 1。
	BouillottePhaseResult BouillottePhase = 1
)

// 手役カテゴリ (高いほど強い)。
const (
	// BouillotteHandHighCard ハイカード
	BouillotteHandHighCard = 0
	// BouillotteHandBrelan ブルラン (スリーカード)
	BouillotteHandBrelan = 1
)

// BouillotteResult は人間プレイヤーから見たラウンド結果。
// GameResult は共有ファイル internal/domain/game_result.go に移動したので到達可能に
// なったが、この型名は JSON ペイロードに出るため統合していない（#4462）。値は
// GameResult と同一。
type BouillotteResult int

const (
	// BouillotteResultLose 負け (ベッティングに残ったが敗北)
	BouillotteResultLose BouillotteResult = -1
	// BouillotteResultNone 結果なし (フォールド / 未参加 / 未解決)
	BouillotteResultNone BouillotteResult = 0
	// BouillotteResultWin 勝ち (ポット獲得)
	BouillotteResultWin BouillotteResult = 1
)

// bouillotteDeckSize は 20 枚デッキのサイズ (A, K, Q, 9, 8 × 4 スート)。
const bouillotteDeckSize = 20

// bouillotteMaxSliceLen はデシリアライズ時のスライス長の上限。
const bouillotteMaxSliceLen = 5000

// bouillotteMaxActions は 1 ラウンドあたりの賭けアクション上限 (安全網)。
const bouillotteMaxActions = 40

// CPU AI の手役判断しきい値 (ハイカードの最強ランク; A=14, K=13, Q=12, 9=9, 8=8)。
const (
	bouillotteCpuRaiseHigh = 13 // K 以上でレイズを検討
	bouillotteCpuCallHigh  = 12 // Q 以上でコール
)

// デシリアライズ検証用のセンチネルエラー。
var (
	errBouillotteSnapshot      = errors.New("bouillotte: invalid serialised game state")
	errBouillotteInvalidPlayer = errors.New("bouillotte: invalid player state")
)

// BouillotteHint はヒント情報 (人間への call/raise/fold 助言)。
type BouillotteHint struct {
	Action string // 推奨アクション ("call"/"raise"/"fold")
	Reason string // ヒント理由キー
}

// bouillotteState はゲーム進行状態。
type bouillotteState struct {
	phase           BouillottePhase
	roundNumber     int
	dealerIdx       int
	currentPlayer   int
	pot             int
	currentBet      int // 現在の必要総拠出額 (アンティ = 開始値)。call でこの額に合わせる
	raiseCount      int // このラウンドのレイズ回数
	actedSinceRaise int // 直近のレイズ (または配札) 以降にアクションしたアクティブ人数
	actionCount     int // 総アクション数 (安全網)
	retourne        *Card
	winnerIdx       int // 直近ラウンドの勝者 (-1 = なし)
	matchWinnerIdx  int // ゲーム全体の勝者 (-1 = 未確定)
	result          BouillotteResult
	gameEndFlag     bool
	scored          bool // ラウンド結果を確定済みか (二重確定防止)
	actionLogBase
}

// Bouillotte はブイヨットの状態を保持する集約ルート。
type Bouillotte struct {
	trumpCards *TrumpCards
	players    []*BouillottePlayer
	config     BouillotteConfig
	state      bouillotteState
}

// NewBouillotte はコンストラクタ。
func NewBouillotte(trumpCards *TrumpCards, players []*BouillottePlayer, config BouillotteConfig) *Bouillotte {
	return &Bouillotte{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		state: bouillotteState{
			phase:          BouillottePhaseBetting,
			winnerIdx:      -1,
			matchWinnerIdx: -1,
			actionLogBase:  actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
		},
	}
}

// NewDefaultBouillotte は標準構成 (人間 seat 0 + CPU) を生成する。CUI / Web / Worker 構築の単一情報源。
func NewDefaultBouillotte() *Bouillotte {
	cfg := DefaultBouillotteConfig()
	g := NewBouillotte(newBouillotteDeck(), bouillotteNewPlayers(cfg), cfg)
	g.Reset()
	return g
}

// newBouillotteDeck は標準 52 枚 (NewTrumpCards(0)) を A, K, Q, 9, 8 の 5 ランクに
// フィルタして 20 枚デッキを構築する。
func newBouillotteDeck() *TrumpCards {
	keep := map[int]bool{1: true, 13: true, 12: true, 9: true, 8: true} // A, K, Q, 9, 8
	full := NewTrumpCards(0)
	filtered := make([]*Card, 0, bouillotteDeckSize)
	for _, c := range full.deck {
		if keep[c.GetValue()] {
			filtered = append(filtered, c)
		}
	}
	full.deck = filtered
	full.deckCnt = len(filtered)
	full.deckInit()
	return full
}

// bouillotteNewPlayers は設定に基づいてプレイヤー列を生成する (seat 0 = 人間)。
func bouillotteNewPlayers(cfg BouillotteConfig) []*BouillottePlayer {
	players := make([]*BouillottePlayer, cfg.PlayerCount)
	for i := range players {
		players[i] = NewBouillottePlayer(i == 0, cfg.StartingChips)
	}
	return players
}

// --- ゲーム進行 ---

// Reset は新しいゲームを開始する。チップ・プレイヤー数を設定から作り直し、第 1 ラウンドを配る。
func (g *Bouillotte) Reset() {
	g.players = bouillotteNewPlayers(g.config)
	g.trumpCards = newBouillotteDeck()
	g.state = bouillotteState{
		phase:          BouillottePhaseBetting,
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
func (g *Bouillotte) NextRound() {
	if g.state.phase != BouillottePhaseResult || g.state.gameEndFlag {
		return
	}
	g.state.roundNumber++
	g.state.dealerIdx = (g.state.dealerIdx + 1) % len(g.players)
	g.startRound()
}

// startRound は 1 ラウンドを準備する: 脱落判定・アンティ徴収・配札・retourne 公開・ベッティング開始。
func (g *Bouillotte) startRound() {
	// ラウンド単位の状態をクリア。
	g.state.winnerIdx = -1
	g.state.result = BouillotteResultNone
	g.state.retourne = nil
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
		for i := 0; i < BouillotteHandSize; i++ {
			if c := g.trumpCards.DrawCard(); c != nil {
				p.AddCard(c)
			}
		}
	}
	g.state.currentBet = g.config.Ante
	g.state.retourne = g.trumpCards.DrawCard()
	g.state.phase = BouillottePhaseBetting
	g.state.currentPlayer = g.nextActive(g.state.dealerIdx)
	g.appendLog(-1, "deal",
		fmt.Sprintf("Round %d: ante %d, pot %d, retourne up", g.state.roundNumber, g.config.Ante, g.state.pot), nil)
	// 最初の手番が CPU なら人間の手番になるまで進める。
	g.driveCPU()
}

// --- Actions (human) ---

// PlayerCall は人間 (現在の手番) が現在の賭けにコール (マッチ) する。
func (g *Bouillotte) PlayerCall() error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	g.applyCall(g.state.currentPlayer)
	g.driveCPU()
	return nil
}

// PlayerRaise は人間 (現在の手番) が賭けをアンティ分だけ引き上げる (ヴィ)。
func (g *Bouillotte) PlayerRaise() error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	if g.state.raiseCount >= BouillotteMaxRaises {
		return NewDomainError(ErrInvalidPlay, "no more raises are allowed this round")
	}
	g.applyRaise(g.state.currentPlayer)
	g.driveCPU()
	return nil
}

// PlayerFold は人間 (現在の手番) が降りる。
func (g *Bouillotte) PlayerFold() error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	g.applyFold(g.state.currentPlayer)
	g.driveCPU()
	return nil
}

// checkTurn は現在の手番が人間かつベッティングフェーズかを検証する。
func (g *Bouillotte) checkTurn() error {
	if g.state.gameEndFlag {
		return NewDomainError(ErrGameEnded, "the game has already ended")
	}
	if g.state.phase != BouillottePhaseBetting {
		return NewDomainError(ErrWrongPhase, "betting is only allowed during the betting phase")
	}
	if !g.players[g.state.currentPlayer].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "it is not your turn")
	}
	return nil
}

// applyCall はコール (現在の賭けにマッチ) を適用する。チップ不足なら自動フォールド。
func (g *Bouillotte) applyCall(idx int) {
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
	g.appendLog(idx, "call", fmt.Sprintf("%s calls %d (pot %d)", g.playerName(idx), need, g.state.pot), nil)
	g.advanceOrClose(idx)
}

// applyRaise はレイズ (ヴィ) を適用する。増分はアンティに等しい。払えなければコールに退避する。
func (g *Bouillotte) applyRaise(idx int) {
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
	g.appendLog(idx, "raise", fmt.Sprintf("%s raises to %d, pays %d (pot %d)", g.playerName(idx), newBet, need, g.state.pot), nil)
	g.advanceOrClose(idx)
}

// applyFold はフォールドを適用する。
func (g *Bouillotte) applyFold(idx int) {
	g.players[idx].SetFolded(true)
	g.state.actionCount++
	g.appendLog(idx, "fold", fmt.Sprintf("%s folds", g.playerName(idx)), nil)
	g.advanceOrClose(idx)
}

// advanceOrClose は次の手番へ進めるか、ベッティングを締めて精算する。
func (g *Bouillotte) advanceOrClose(fromIdx int) {
	// 残り 1 人ならクリーンウィン、全アクティブがアクション済みならショーダウン。
	if g.activeCount() <= 1 ||
		g.state.actedSinceRaise >= g.activeCount() ||
		g.state.actionCount >= bouillotteMaxActions {
		g.resolveRound()
		return
	}
	g.state.currentPlayer = g.nextActive(fromIdx)
}

// resolveRound はベッティング終了時にポットを精算し、フェーズを Result へ遷移する。
// scored フラグで二重精算を防ぐ (フェーズ入場時に 1 回だけ発火)。
func (g *Bouillotte) resolveRound() {
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
			fmt.Sprintf("%s wins the pot (%d)", g.playerName(winner), g.state.pot), nil)
	}
	g.setHumanResult(winner)
	g.state.pot = 0
	g.state.scored = true
	g.state.phase = BouillottePhaseResult
	g.checkGameEnd()
}

// setHumanResult は人間 (seat 0) の勝敗結果を設定する。
func (g *Bouillotte) setHumanResult(winnerIdx int) {
	human := g.players[0]
	if human.GetOut() || human.GetFolded() {
		g.state.result = BouillotteResultNone
		return
	}
	if winnerIdx == 0 {
		g.state.result = BouillotteResultWin
	} else {
		g.state.result = BouillotteResultLose
	}
}

// checkGameEnd は停止条件 (規定ラウンド到達 or 参加可能者 2 人未満) を判定し、
// 満たせばゲームを終了させる。
func (g *Bouillotte) checkGameEnd() {
	if g.state.roundNumber >= g.config.TargetRounds || g.solventCount() < 2 {
		g.endGame()
	}
}

// endGame はゲームを終了し、チップ最多のプレイヤーを勝者に設定する。
func (g *Bouillotte) endGame() {
	g.state.gameEndFlag = true
	g.state.phase = BouillottePhaseResult
	g.state.matchWinnerIdx = g.richestIdx()
	g.appendLog(g.state.matchWinnerIdx, "game_end",
		fmt.Sprintf("%s wins the game", g.playerName(g.state.matchWinnerIdx)), nil)
}

// --- CPU ---

// driveCPU は現在の手番が CPU の間、CPU のアクションを実行し続ける。人間の手番か
// ラウンド解決で停止する。
func (g *Bouillotte) driveCPU() {
	guard := 0
	for g.state.phase == BouillottePhaseBetting && !g.state.gameEndFlag {
		idx := g.state.currentPlayer
		if g.players[idx].GetIsHuman() {
			return
		}
		g.cpuAct(idx)
		guard++
		if guard > bouillotteMaxActions*2 {
			// 安全網 (通常到達しない): 強制的にラウンドを締める。
			g.resolveRound()
			return
		}
	}
}

// cpuAct は CPU が 1 アクション (call/raise/fold) を選ぶ。
func (g *Bouillotte) cpuAct(idx int) {
	p := g.players[idx]
	cat, tb := g.evalPlayer(idx)
	need := g.state.currentBet - p.GetRoundBet()
	if need < 0 {
		need = 0
	}
	canRaise := g.state.raiseCount < BouillotteMaxRaises && p.GetChips() >= need+g.config.Ante
	high := 0
	if len(tb) > 0 {
		high = tb[0]
	}
	switch {
	case cat == BouillotteHandBrelan:
		// 非常に強い: 可能ならレイズ、無理ならコール。
		if canRaise {
			g.applyRaise(idx)
			return
		}
		g.applyCall(idx)
	case high >= bouillotteCpuRaiseHigh:
		// 強い (A/K ハイ): レイズ控えめ、そうでなければコール。
		if canRaise && g.state.raiseCount < 2 {
			g.applyRaise(idx)
			return
		}
		g.applyCall(idx)
	case need == 0:
		// タダで残れる (チェック相当) → コール。
		g.applyCall(idx)
	case high >= bouillotteCpuCallHigh && need <= g.config.Ante*2:
		// 中程度 (Q ハイ) かつ安い → コール。
		g.applyCall(idx)
	default:
		// 弱い手で賭けが必要 → フォールド。
		g.applyFold(idx)
	}
}

// --- 手役評価 (インライン) ---

// bouillotteRank はカードのブイヨット順位を返す (A=14, K=13, Q=12, 9=9, 8=8)。
func bouillotteRank(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// BouillotteEval は 3 枚の手 + retourne の役カテゴリと比較用タイブレーク列を返す。
// カテゴリが同じならこの列を辞書順比較すれば勝敗が決まる。
//
//   - Brelan: タイブレーク = [トリップのランク, ファヴォリフラグ(1=favori, 0=simple)]
//   - HighCard: タイブレーク = 手札 3 枚 + retourne のランク降順列
func BouillotteEval(cards []*Card, retourne *Card) (int, []int) {
	if len(cards) != BouillotteHandSize {
		return -1, nil
	}
	r := []int{bouillotteRank(cards[0]), bouillotteRank(cards[1]), bouillotteRank(cards[2])}
	bouillotteSortDesc(r)
	rr := bouillotteRank(retourne)

	// スリーカード (ブルラン・サンプル)。
	if r[0] == r[1] && r[1] == r[2] {
		return BouillotteHandBrelan, []int{r[0], 0}
	}
	// ファヴォリ: 手札にペアがあり retourne がそのランクと一致。
	if rr > 0 {
		cnt := 0
		for _, x := range r {
			if x == rr {
				cnt++
			}
		}
		if cnt == 2 {
			return BouillotteHandBrelan, []int{rr, 1}
		}
	}
	// ハイカード: 手札 3 枚 + retourne の降順列。
	all := []int{r[0], r[1], r[2]}
	if rr > 0 {
		all = append(all, rr)
	}
	bouillotteSortDesc(all)
	return BouillotteHandHighCard, all
}

// bouillotteSortDesc はスライスを降順にソートする (小さな固定長スライス向けの単純挿入ソート)。
func bouillotteSortDesc(s []int) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] > s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// evalPlayer は playerIdx の手役 (retourne を共有カードとして含む) を返す。
func (g *Bouillotte) evalPlayer(idx int) (int, []int) {
	p := g.players[idx]
	cards := make([]*Card, 0, BouillotteHandSize)
	for i := 0; i < p.GetCardsSize(); i++ {
		cards = append(cards, p.GetCard(i))
	}
	return BouillotteEval(cards, g.state.retourne)
}

// BouillotteCompare は手 a が手 b に勝てば 1、負ければ -1、引き分けは 0 を返す。
func BouillotteCompare(catA int, tbA []int, catB int, tbB []int) int {
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
func (g *Bouillotte) bestHand(seats []int) int {
	best := seats[0]
	bestCat, bestTb := g.evalPlayer(best)
	for _, idx := range seats[1:] {
		cat, tb := g.evalPlayer(idx)
		if BouillotteCompare(cat, tb, bestCat, bestTb) > 0 {
			best, bestCat, bestTb = idx, cat, tb
		}
	}
	return best
}

// --- ヘルパー ---

// activeSeats は未フォールド・非脱落プレイヤーのインデックス列 (昇順) を返す。
func (g *Bouillotte) activeSeats() []int {
	out := make([]int, 0, len(g.players))
	for i, p := range g.players {
		if !p.GetOut() && !p.GetFolded() {
			out = append(out, i)
		}
	}
	return out
}

// activeCount は未フォールド・非脱落プレイヤー数を返す。
func (g *Bouillotte) activeCount() int {
	n := 0
	for _, p := range g.players {
		if !p.GetOut() && !p.GetFolded() {
			n++
		}
	}
	return n
}

// nextActive は from の次の未フォールド・非脱落プレイヤーを返す。
func (g *Bouillotte) nextActive(from int) int {
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
func (g *Bouillotte) solventCount() int {
	n := 0
	for _, p := range g.players {
		if !p.GetOut() && p.GetChips() >= g.config.Ante {
			n++
		}
	}
	return n
}

// richestIdx はチップが最多のプレイヤーのインデックスを返す (同数は座席番号の小さい方)。
func (g *Bouillotte) richestIdx() int {
	best := 0
	for i, p := range g.players {
		if p.GetChips() > g.players[best].GetChips() {
			best = i
		}
	}
	return best
}

// playerName は表示用のプレイヤー名を返す。
func (g *Bouillotte) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

func (g *Bouillotte) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.state.appendLog(playerIdx, actionType, detail, cards)
}

// --- Hint ---

// GetHint はベッティング中の人間 (seat 0 の手番) に call/raise/fold の助言を返す。
func (g *Bouillotte) GetHint() *BouillotteHint {
	if g.state.phase != BouillottePhaseBetting || g.state.gameEndFlag || g.state.currentPlayer != 0 {
		return nil
	}
	p := g.players[0]
	if p.GetOut() || p.GetFolded() {
		return nil
	}
	cat, tb := g.evalPlayer(0)
	high := 0
	if len(tb) > 0 {
		high = tb[0]
	}
	need := g.state.currentBet - p.GetRoundBet()
	switch {
	case cat == BouillotteHandBrelan:
		return &BouillotteHint{Action: "raise", Reason: "strong_hand"}
	case high >= bouillotteCpuRaiseHigh:
		return &BouillotteHint{Action: "raise", Reason: "strong_hand"}
	case high >= bouillotteCpuCallHigh:
		return &BouillotteHint{Action: "call", Reason: "medium_hand"}
	case need <= 0:
		return &BouillotteHint{Action: "call", Reason: "medium_hand"}
	default:
		return &BouillotteHint{Action: "fold", Reason: "weak_hand"}
	}
}

// --- 状態アクセサ ---

// GetPhase は現在のフェーズを返す。
func (g *Bouillotte) GetPhase() BouillottePhase { return g.state.phase }

// SetPhase はフェーズを設定する (テスト用)。
func (g *Bouillotte) SetPhase(p BouillottePhase) { g.state.phase = p }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *Bouillotte) GetGameEndFlag() bool { return g.state.gameEndFlag }

// GetRoundNumber は現在のラウンド番号を返す。
func (g *Bouillotte) GetRoundNumber() int { return g.state.roundNumber }

// GetDealerIdx はディーラーの座席番号を返す。
func (g *Bouillotte) GetDealerIdx() int { return g.state.dealerIdx }

// SetDealerIdx はディーラーの座席番号を設定する (テスト用)。
func (g *Bouillotte) SetDealerIdx(idx int) { g.state.dealerIdx = idx }

// GetCurrentPlayerIdx は現在の手番プレイヤーの座席番号を返す。
func (g *Bouillotte) GetCurrentPlayerIdx() int { return g.state.currentPlayer }

// SetCurrentPlayerIdx は現在の手番プレイヤーを設定する (テスト用)。
func (g *Bouillotte) SetCurrentPlayerIdx(idx int) { g.state.currentPlayer = idx }

// GetPot は現在のポットを返す。
func (g *Bouillotte) GetPot() int { return g.state.pot }

// SetPot はポットを設定する (テスト用)。
func (g *Bouillotte) SetPot(v int) { g.state.pot = v }

// GetCurrentBet は現在の必要総拠出額を返す。
func (g *Bouillotte) GetCurrentBet() int { return g.state.currentBet }

// SetCurrentBet は現在の必要総拠出額を設定する (テスト用)。
func (g *Bouillotte) SetCurrentBet(v int) { g.state.currentBet = v }

// GetRaiseCount はこのラウンドのレイズ回数を返す。
func (g *Bouillotte) GetRaiseCount() int { return g.state.raiseCount }

// GetMaxRaises は 1 ラウンドあたりの最大レイズ回数を返す。
func (g *Bouillotte) GetMaxRaises() int { return BouillotteMaxRaises }

// GetAnte はアンティ額を返す。
func (g *Bouillotte) GetAnte() int { return g.config.Ante }

// GetRetourne は共有カード (retourne) を返す (未配札時は nil)。
func (g *Bouillotte) GetRetourne() *Card { return g.state.retourne }

// SetRetourne は共有カード (retourne) を設定する (テスト用)。
func (g *Bouillotte) SetRetourne(c *Card) { g.state.retourne = c }

// GetWinnerIdx は直近ラウンドの勝者を返す (-1 = なし)。
func (g *Bouillotte) GetWinnerIdx() int { return g.state.winnerIdx }

// GetMatchWinnerIdx はゲーム全体の勝者を返す (-1 = 未確定)。
func (g *Bouillotte) GetMatchWinnerIdx() int { return g.state.matchWinnerIdx }

// GetResult は人間から見たラウンド結果を返す。
func (g *Bouillotte) GetResult() BouillotteResult { return g.state.result }

// GetPlayerCnt はプレイヤー数を返す。
func (g *Bouillotte) GetPlayerCnt() int { return len(g.players) }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *Bouillotte) GetPlayer(i int) *BouillottePlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetChips は人間 (seat 0) の保有チップを返す。
func (g *Bouillotte) GetChips() int {
	if len(g.players) == 0 {
		return 0
	}
	return g.players[0].GetChips()
}

// IsHumanTurn は現在の手番が人間 (seat 0) かどうかを返す。
func (g *Bouillotte) IsHumanTurn() bool {
	if g.state.phase != BouillottePhaseBetting || g.state.gameEndFlag {
		return false
	}
	idx := g.state.currentPlayer
	if idx < 0 || idx >= len(g.players) {
		return false
	}
	return g.players[idx].GetIsHuman()
}

// CanRaise は現在の手番プレイヤーがレイズ可能か (回数上限未満かつ増分を払える) を返す。
func (g *Bouillotte) CanRaise() bool {
	if !g.IsHumanTurn() || g.state.raiseCount >= BouillotteMaxRaises {
		return false
	}
	p := g.players[g.state.currentPlayer]
	need := (g.state.currentBet + g.config.Ante) - p.GetRoundBet()
	return p.GetChips() >= need
}

// GetConfig はローカルルール設定を返す。
func (g *Bouillotte) GetConfig() BouillotteConfig { return g.config }

// SetConfig はローカルルール設定を変更する。
func (g *Bouillotte) SetConfig(cfg BouillotteConfig) { g.config = cfg }

// GetActionLog は棋譜を返す。
func (g *Bouillotte) GetActionLog() []*ActionLogEntry { return g.state.actionLog }

// ResolveForTest は手札・retourne を設定済みの状態でラウンドを解決する (テスト用)。乱数配札を
// 迂回して勝敗解決を決定的に検証するためのショートカット。
func (g *Bouillotte) ResolveForTest() { g.resolveRound() }

// ClearBettingForTest は 1 ラウンドの賭けブックキーピング (raiseCount /
// actedSinceRaise / actionCount) をゼロに戻す (テスト用)。Reset() はディーラー左の
// CPU を人間の手番まで自動進行させるため残留カウンタが乗る。決定的な途中状態を
// 組み立てるテストはこれで残留をクリアする。
func (g *Bouillotte) ClearBettingForTest() {
	g.state.raiseCount = 0
	g.state.actedSinceRaise = 0
	g.state.actionCount = 0
}

// --- JSON Serialization ---

// bouillotteJSON is the JSON wire format for Bouillotte.
type bouillotteJSON struct {
	TrumpCards      *TrumpCards         `json:"tc"`
	Players         []*BouillottePlayer `json:"ps"`
	Config          BouillotteConfig    `json:"cf"`
	Phase           BouillottePhase     `json:"ph"`
	RoundNumber     int                 `json:"rn"`
	DealerIdx       int                 `json:"di"`
	CurrentPlayer   int                 `json:"ci"`
	Pot             int                 `json:"pt"`
	CurrentBet      int                 `json:"cb"`
	RaiseCount      int                 `json:"rc"`
	ActedSinceRaise int                 `json:"as"`
	ActionCount     int                 `json:"an"`
	Retourne        *Card               `json:"rt"`
	WinnerIdx       int                 `json:"wi"`
	MatchWinnerIdx  int                 `json:"mw"`
	Result          BouillotteResult    `json:"re"`
	GameEndFlag     bool                `json:"ge"`
	Scored          bool                `json:"sc"`
	ActionLog       []*ActionLogEntry   `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Bouillotte) MarshalJSON() ([]byte, error) {
	return json.Marshal(bouillotteJSON{
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
		Retourne:        g.state.retourne,
		WinnerIdx:       g.state.winnerIdx,
		MatchWinnerIdx:  g.state.matchWinnerIdx,
		Result:          g.state.result,
		GameEndFlag:     g.state.gameEndFlag,
		Scored:          g.state.scored,
		ActionLog:       g.state.actionLog,
	})
}

// bouillotteValidPhase は有効なフェーズかどうか。
func bouillotteValidPhase(p BouillottePhase) bool {
	return p == BouillottePhaseBetting || p == BouillottePhaseResult
}

// UnmarshalJSON implements json.Unmarshaler. 不正な永続化データを拒否するための
// バリデーションを行う (KV 復元時のインデックス範囲・スライス健全性)。
func (g *Bouillotte) UnmarshalJSON(data []byte) error {
	var j bouillotteJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("bouillotte: invalid config: %w", err)
	}
	n := len(j.Players)
	if n < BouillotteMinPlayerCount || n > BouillotteMaxPlayerCount || n != j.Config.PlayerCount {
		return errBouillotteSnapshot
	}
	if len(j.ActionLog) > bouillotteMaxSliceLen {
		return errBouillotteSnapshot
	}
	if !bouillotteValidPhase(j.Phase) {
		return errBouillotteSnapshot
	}
	if j.RoundNumber < 1 || j.Pot < 0 || j.CurrentBet < 0 || j.RaiseCount < 0 ||
		j.ActedSinceRaise < 0 || j.ActionCount < 0 {
		return errBouillotteSnapshot
	}
	if j.DealerIdx < 0 || j.DealerIdx >= n || j.CurrentPlayer < 0 || j.CurrentPlayer >= n {
		return errBouillotteSnapshot
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= n || j.MatchWinnerIdx < -1 || j.MatchWinnerIdx >= n {
		return errBouillotteSnapshot
	}
	if j.Result < BouillotteResultLose || j.Result > BouillotteResultWin {
		return errBouillotteSnapshot
	}
	for _, p := range j.Players {
		if p == nil {
			return errBouillotteSnapshot
		}
	}
	for _, e := range j.ActionLog {
		if e == nil {
			return errBouillotteSnapshot
		}
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newBouillotteDeck()
	}
	g.players = j.Players
	g.config = j.Config
	g.state = bouillotteState{
		phase:           j.Phase,
		roundNumber:     j.RoundNumber,
		dealerIdx:       j.DealerIdx,
		currentPlayer:   j.CurrentPlayer,
		pot:             j.Pot,
		currentBet:      j.CurrentBet,
		raiseCount:      j.RaiseCount,
		actedSinceRaise: j.ActedSinceRaise,
		actionCount:     j.ActionCount,
		retourne:        j.Retourne,
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
