//go:build !js || !wasm || extra

// Package domain アナコンダ (Anaconda / Pass the Trash) のドメインモデル。
//
// Anaconda はアメリカのホームポーカーの変種「Pass the Trash」。標準 52 枚デッキを使い、
// 3〜7 人が全員アンティを入れてから 7 枚ずつ配られる。手札を隣へ回し (パス)、最良の 5 枚を
// 残し、1 枚ずつ公開しながら公開の合間にベッティングを行い、最強の 5 枚ポーカーハンドが
// ポットを総取りする。
//
// # 1 ラウンドの流れ
//
//  1. 全員がアンティをポットに払い、7 枚ずつ配られる (パスフェーズへ)。
//  2. パスフェーズ (3 → 2 → 1): 3 サブラウンド。各プレイヤーは選んだ N 枚を「左隣」の
//     プレイヤーへ同時に渡す。サブラウンド 1 は 3 枚、2 は 2 枚、3 は 1 枚。人間が渡す札を
//     選び、CPU は最も弱い札を渡す。
//  3. セットフェーズ: 各プレイヤーは手元の 7 枚から残す 5 枚を選ぶ (残り 2 枚は捨て札)。
//     人間が 5 枚を選び、CPU はポーカーランクで最良の 5 枚を残す。
//  4. ロールフェーズ (5 ロール): 残した 5 枚を 1 枚ずつ公開する。ロール 1 の前および各公開の
//     後にベッティングラウンド (チェック/コール・レイズ・フォールド) が入る。
//  5. ショーダウン: フォールドしなかったプレイヤーの中で最強の 5 枚ポーカーハンドがポットを
//     総取りする。途中で 1 人を残して全員フォールドしたら、その 1 人が即座にポットを獲得する。
//
// ゲームは「規定ラウンド数 (TargetRounds) を消化」または「アンティを払えるプレイヤーが
// AnacondaMinPlayerCount 未満 (人間の脱落を含む)」で終了し、チップ最多のプレイヤーが勝者となる。
//
// # デッキ
//
// 標準 52 枚 (ジョーカーなし)。NewTrumpCards(0) は extra ワーカーから到達可能。
//
// 本実装は extra ワーカーから到達可能なよう、ポーカー手役評価・ベッティング・ポット精算
// ロジックをすべてインラインで持つ (既存の Poker 評価器は casino ビルドタグで到達不可のため)。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

// AnacondaPhase はゲームフェーズ。
type AnacondaPhase int

// Anaconda のフェーズ定数。ワイヤー値はフロントエンドの enum と一致させる。
const (
	// AnacondaPhasePass パスフェーズ (3→2→1 のサブラウンドで札を左隣へ渡す)。ワイヤー値 0。
	AnacondaPhasePass AnacondaPhase = 0
	// AnacondaPhaseSet セットフェーズ (残す 5 枚を選ぶ)。ワイヤー値 1。
	AnacondaPhaseSet AnacondaPhase = 1
	// AnacondaPhaseRoll ロールフェーズ (1 枚ずつ公開 + ベッティング)。ワイヤー値 2。
	AnacondaPhaseRoll AnacondaPhase = 2
	// AnacondaPhaseResult ラウンド解決済み (結果表示; 次ラウンド待ち or ゲーム終了)。ワイヤー値 3。
	AnacondaPhaseResult AnacondaPhase = 3
)

// ポーカー手役カテゴリ (高いほど強い)。ワイヤー値はフロントエンドの enum と一致させる。
const (
	// AnacondaHighCard ハイカード (ノーペア)
	AnacondaHighCard = 0
	// AnacondaOnePair ワンペア
	AnacondaOnePair = 1
	// AnacondaTwoPair ツーペア
	AnacondaTwoPair = 2
	// AnacondaThreeKind スリーカード
	AnacondaThreeKind = 3
	// AnacondaStraight ストレート
	AnacondaStraight = 4
	// AnacondaFlush フラッシュ
	AnacondaFlush = 5
	// AnacondaFullHouse フルハウス
	AnacondaFullHouse = 6
	// AnacondaFourKind フォーカード
	AnacondaFourKind = 7
	// AnacondaStraightFlush ストレートフラッシュ
	AnacondaStraightFlush = 8
)

// AnacondaResult は人間プレイヤーから見たラウンド結果。
// GameResult は共有ファイル internal/domain/game_result.go に移動したので到達可能に
// なったが、この型名は JSON ペイロードに出るため統合していない（#4462）。値は
// GameResult と同一。
type AnacondaResult int

const (
	// AnacondaResultLose 負け (ショーダウンに残ったが敗北)
	AnacondaResultLose AnacondaResult = -1
	// AnacondaResultNone 結果なし (フォールド / 未参加 / 未解決)
	AnacondaResultNone AnacondaResult = 0
	// AnacondaResultWin 勝ち (ポット獲得)
	AnacondaResultWin AnacondaResult = 1
)

// Anaconda のゲーム定数。
const (
	// AnacondaDealSize は各プレイヤーへの配札枚数。
	AnacondaDealSize = 7
	// AnacondaKeepSize は残す (公開する) 枚数 = 5 枚ポーカーハンド。
	AnacondaKeepSize = 5
	// AnacondaPassStart は最初のパスサブラウンドで渡す枚数 (3 → 2 → 1)。
	AnacondaPassStart = 3
	// AnacondaMaxRaises は 1 ベッティングラウンドあたりの最大レイズ回数。
	AnacondaMaxRaises = 3
)

// anacondaMaxSliceLen はデシリアライズ時のスライス長の上限。
const anacondaMaxSliceLen = 5000

// anacondaMaxActions は 1 ベッティングラウンド (ストリート) あたりの賭けアクション上限 (安全網)。
const anacondaMaxActions = 100

// CPU AI のベッティング判断しきい値 (ハイカードの最強ランク; A=14, K=13, Q=12)。
const (
	anacondaCpuRaiseHigh = 13 // K 以上のハイカードでレイズを検討
	anacondaCpuCallHigh  = 12 // Q 以上のハイカードでコール
)

// デシリアライズ検証用のセンチネルエラー。
var (
	errAnacondaSnapshot      = errors.New("anaconda: invalid serialised game state")
	errAnacondaInvalidPlayer = errors.New("anaconda: invalid player state")
)

// AnacondaHint はヒント情報 (人間へのパス/セット/ベット助言)。
type AnacondaHint struct {
	Action      string // 推奨アクション ("pass"/"keep"/"call"/"raise"/"fold")
	CardIndices []int  // pass/keep の推奨カードインデックス (call/raise/fold では nil)
	Reason      string // ヒント理由キー
}

// anacondaState はゲーム進行状態。
type anacondaState struct {
	phase           AnacondaPhase
	roundNumber     int
	dealerIdx       int
	currentPlayer   int // ロールベッティングの手番プレイヤー
	passCount       int // 現在のパスサブラウンドで渡す枚数 (3/2/1; パス外では 0)
	rollIndex       int // ロールフェーズで公開済みの枚数 (0..AnacondaKeepSize)
	pot             int
	currentBet      int // 現在のストリートで必要な拠出額 (call でこの額に合わせる)
	raiseCount      int // このストリートのレイズ回数
	actedSinceRaise int // 直近のレイズ (またはストリート開始) 以降にアクションしたアクティブ人数
	actionCount     int // このストリートの総アクション数 (安全網)
	winnerIdx       int // 直近ラウンドの勝者 (-1 = なし)
	matchWinnerIdx  int // ゲーム全体の勝者 (-1 = 未確定)
	result          AnacondaResult
	gameEndFlag     bool
	scored          bool // ラウンド結果を確定済みか (二重確定防止)
	actionLogBase
}

// Anaconda はアナコンダの状態を保持する集約ルート。
type Anaconda struct {
	trumpCards *TrumpCards
	players    []*AnacondaPlayer
	config     AnacondaConfig
	state      anacondaState
}

// NewAnaconda はコンストラクタ。
func NewAnaconda(trumpCards *TrumpCards, players []*AnacondaPlayer, config AnacondaConfig) *Anaconda {
	return &Anaconda{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		state: anacondaState{
			phase:          AnacondaPhasePass,
			winnerIdx:      -1,
			matchWinnerIdx: -1,
			actionLogBase:  actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
		},
	}
}

// NewDefaultAnaconda は標準構成 (人間 seat 0 + CPU) を生成する。CUI / Web / Worker 構築の単一情報源。
func NewDefaultAnaconda() *Anaconda {
	cfg := DefaultAnacondaConfig()
	g := NewAnaconda(NewTrumpCards(0), anacondaNewPlayers(cfg), cfg)
	g.Reset()
	return g
}

// anacondaNewPlayers は設定に基づいてプレイヤー列を生成する (seat 0 = 人間)。
func anacondaNewPlayers(cfg AnacondaConfig) []*AnacondaPlayer {
	players := make([]*AnacondaPlayer, cfg.PlayerCount)
	for i := range players {
		players[i] = NewAnacondaPlayer(i == 0, cfg.StartingChips)
	}
	return players
}

// --- ゲーム進行 ---

// Reset は新しいゲームを開始する。チップ・プレイヤー数を設定から作り直し、第 1 ラウンドを配る。
func (g *Anaconda) Reset() {
	g.players = anacondaNewPlayers(g.config)
	g.trumpCards = NewTrumpCards(0)
	g.state = anacondaState{
		phase:          AnacondaPhasePass,
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
func (g *Anaconda) NextRound() {
	if g.state.phase != AnacondaPhaseResult || g.state.gameEndFlag {
		return
	}
	g.state.roundNumber++
	g.state.dealerIdx = (g.state.dealerIdx + 1) % len(g.players)
	g.startRound()
}

// startRound は 1 ラウンドを準備する: 脱落判定・アンティ徴収・配札・パスフェーズへ遷移。
func (g *Anaconda) startRound() {
	g.state.winnerIdx = -1
	g.state.result = AnacondaResultNone
	g.state.pot = 0
	g.state.currentBet = 0
	g.state.raiseCount = 0
	g.state.actedSinceRaise = 0
	g.state.actionCount = 0
	g.state.rollIndex = 0
	g.state.passCount = AnacondaPassStart
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
	// 参加可能なプレイヤーが規定数未満、または人間が脱落 → ゲーム終了。
	if g.solventCount() < AnacondaMinPlayerCount || g.players[0].GetOut() {
		g.endGame()
		return
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	for _, p := range g.players {
		if p.GetOut() {
			continue
		}
		p.SubtractChips(g.config.Ante)
		p.AddRoundBet(g.config.Ante)
		g.state.pot += g.config.Ante
		for i := 0; i < AnacondaDealSize; i++ {
			if c := g.trumpCards.DrawCard(); c != nil {
				p.AddCard(c)
			}
		}
	}
	g.state.phase = AnacondaPhasePass
	g.appendLog(-1, "deal",
		fmt.Sprintf("Round %d: ante %d, pot %d, deal %d", g.state.roundNumber, g.config.Ante, g.state.pot, AnacondaDealSize), nil)
}

// --- パスフェーズ ---

// Pass は人間 (seat 0) が選んだ札を左隣へ渡し、パスサブラウンドを進める。
func (g *Anaconda) Pass(indices []int) error {
	if g.state.gameEndFlag {
		return NewDomainError(ErrGameEnded, "the game has already ended")
	}
	if g.state.phase != AnacondaPhasePass {
		return NewDomainError(ErrWrongPhase, "passing is only allowed during the pass phase")
	}
	human := g.players[0]
	if human.GetOut() {
		return NewDomainError(ErrInvalidPlay, "you are out of the game")
	}
	if err := anacondaValidateIndices(indices, g.state.passCount, human.GetCardsSize()); err != nil {
		return err
	}
	g.executePass(indices)
	if g.state.passCount <= 1 {
		g.enterSetPhase()
	} else {
		g.state.passCount--
	}
	return nil
}

// executePass は全参加プレイヤーの札を左隣へ同時に渡す。humanIndices は人間の選択。
func (g *Anaconda) executePass(humanIndices []int) {
	participants := g.participantSeats()
	n := len(participants)
	outgoing := make([][]*Card, n)
	for k, seat := range participants {
		p := g.players[seat]
		var idxs []int
		if p.GetIsHuman() {
			idxs = humanIndices
		} else {
			idxs = g.cpuPassIndices(p, g.state.passCount)
		}
		outgoing[k] = p.RemoveCards(idxs)
		g.appendLog(seat, "pass",
			fmt.Sprintf("%s passes %d card(s) left", playerName(g.players, seat), len(outgoing[k])), append([]*Card(nil), outgoing[k]...))
	}
	for k := range participants {
		recipient := participants[(k+1)%n]
		for _, c := range outgoing[k] {
			g.players[recipient].AddCard(c)
		}
	}
}

// cpuPassIndices は CPU が渡す最も弱い count 枚のインデックスを返す (ランク昇順)。
func (g *Anaconda) cpuPassIndices(p *AnacondaPlayer, count int) []int {
	type ci struct{ idx, rank int }
	list := make([]ci, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		list = append(list, ci{i, anacondaRank(p.GetCard(i))})
	}
	sort.Slice(list, func(a, b int) bool { return list[a].rank < list[b].rank })
	idxs := make([]int, 0, count)
	for i := 0; i < count && i < len(list); i++ {
		idxs = append(idxs, list[i].idx)
	}
	return idxs
}

// --- セットフェーズ ---

// enterSetPhase はセットフェーズへ遷移する。
func (g *Anaconda) enterSetPhase() {
	g.state.passCount = 0
	g.state.phase = AnacondaPhaseSet
	g.appendLog(-1, "set", "choose the best 5 cards to keep", nil)
}

// Keep は人間 (seat 0) が残す 5 枚を選び、CPU も最良の 5 枚を残してロールフェーズへ遷移する。
func (g *Anaconda) Keep(indices []int) error {
	if g.state.gameEndFlag {
		return NewDomainError(ErrGameEnded, "the game has already ended")
	}
	if g.state.phase != AnacondaPhaseSet {
		return NewDomainError(ErrWrongPhase, "keeping is only allowed during the set phase")
	}
	human := g.players[0]
	if human.GetOut() {
		return NewDomainError(ErrInvalidPlay, "you are out of the game")
	}
	if err := anacondaValidateIndices(indices, AnacondaKeepSize, human.GetCardsSize()); err != nil {
		return err
	}
	g.applyKeep(0, indices)
	for i, p := range g.players {
		if i == 0 || p.GetOut() {
			continue
		}
		g.applyKeep(i, g.cpuBestKeepIndices(p))
	}
	g.enterRollPhase()
	return nil
}

// applyKeep は seat の手札から keep 以外を捨て、残す 5 枚だけにする。
func (g *Anaconda) applyKeep(seat int, keep []int) {
	p := g.players[seat]
	keepSet := make(map[int]bool, len(keep))
	for _, k := range keep {
		keepSet[k] = true
	}
	discard := make([]int, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		if !keepSet[i] {
			discard = append(discard, i)
		}
	}
	removed := p.RemoveCards(discard)
	g.appendLog(seat, "keep",
		fmt.Sprintf("%s discards %d card(s)", playerName(g.players, seat), len(removed)), append([]*Card(nil), removed...))
}

// cpuBestKeepIndices は 7 枚から最良の 5 枚を残すインデックス列を返す (捨てる 2 枚を総当り)。
func (g *Anaconda) cpuBestKeepIndices(p *AnacondaPlayer) []int {
	size := p.GetCardsSize()
	bestKeep := make([]int, 0, AnacondaKeepSize)
	bestCat := -1
	var bestTb []int
	for a := 0; a < size; a++ {
		for b := a + 1; b < size; b++ {
			keep := make([]int, 0, AnacondaKeepSize)
			cards := make([]*Card, 0, AnacondaKeepSize)
			for i := 0; i < size; i++ {
				if i == a || i == b {
					continue
				}
				keep = append(keep, i)
				cards = append(cards, p.GetCard(i))
			}
			cat, tb := AnacondaEval(cards)
			if bestCat == -1 || AnacondaCompare(cat, tb, bestCat, bestTb) > 0 {
				bestCat, bestTb, bestKeep = cat, tb, keep
			}
		}
	}
	return bestKeep
}

// --- ロールフェーズ / ベッティング ---

// enterRollPhase はロールフェーズへ遷移し、最初のベッティングラウンドを開始する。
func (g *Anaconda) enterRollPhase() {
	g.state.phase = AnacondaPhaseRoll
	g.state.rollIndex = 0
	g.startBettingRound()
	g.driveCPU()
}

// startBettingRound は 1 ベッティングラウンド (ストリート) を準備する。
func (g *Anaconda) startBettingRound() {
	g.state.currentBet = 0
	g.state.raiseCount = 0
	g.state.actedSinceRaise = 0
	g.state.actionCount = 0
	for _, p := range g.players {
		p.SetStreetBet(0)
	}
	g.state.currentPlayer = g.nextActive(g.state.dealerIdx)
	g.appendLog(-1, "roll",
		fmt.Sprintf("betting round (revealed %d of %d)", g.state.rollIndex, AnacondaKeepSize), nil)
}

// PlayerCall は人間 (現在の手番) が現在の賭けにコール (チェック含む) する。
func (g *Anaconda) PlayerCall() error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	g.applyCall(g.state.currentPlayer)
	g.driveCPU()
	return nil
}

// PlayerRaise は人間 (現在の手番) が賭けをアンティ分だけ引き上げる。
func (g *Anaconda) PlayerRaise() error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	if g.state.raiseCount >= AnacondaMaxRaises {
		return NewDomainError(ErrInvalidPlay, "no more raises are allowed this round")
	}
	g.applyRaise(g.state.currentPlayer)
	g.driveCPU()
	return nil
}

// PlayerFold は人間 (現在の手番) が降りる。
func (g *Anaconda) PlayerFold() error {
	if err := g.checkTurn(); err != nil {
		return err
	}
	g.applyFold(g.state.currentPlayer)
	g.driveCPU()
	return nil
}

// checkTurn は現在の手番が人間かつロールフェーズかを検証する。
func (g *Anaconda) checkTurn() error {
	if g.state.gameEndFlag {
		return NewDomainError(ErrGameEnded, "the game has already ended")
	}
	if g.state.phase != AnacondaPhaseRoll {
		return NewDomainError(ErrWrongPhase, "betting is only allowed during the roll phase")
	}
	if !g.players[g.state.currentPlayer].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "it is not your turn")
	}
	return nil
}

// applyCall はコール (現在の賭けにマッチ; 差額 0 ならチェック) を適用する。払えなければ自動フォールド。
func (g *Anaconda) applyCall(idx int) {
	p := g.players[idx]
	need := g.state.currentBet - p.GetStreetBet()
	if need < 0 {
		need = 0
	}
	if need > 0 && p.GetChips() < need {
		g.applyFold(idx)
		return
	}
	if need > 0 {
		p.SubtractChips(need)
		p.AddStreetBet(need)
		p.AddRoundBet(need)
		g.state.pot += need
	}
	g.state.actedSinceRaise++
	g.state.actionCount++
	g.appendLog(idx, "call", fmt.Sprintf("%s calls %d (pot %d)", playerName(g.players, idx), need, g.state.pot), nil)
	g.advanceOrClose(idx)
}

// applyRaise はレイズを適用する。増分はアンティに等しい。払えなければコールに退避する。
func (g *Anaconda) applyRaise(idx int) {
	p := g.players[idx]
	newBet := g.state.currentBet + g.config.Ante
	need := newBet - p.GetStreetBet()
	if p.GetChips() < need {
		g.applyCall(idx)
		return
	}
	g.state.currentBet = newBet
	g.state.raiseCount++
	p.SubtractChips(need)
	p.AddStreetBet(need)
	p.AddRoundBet(need)
	g.state.pot += need
	g.state.actedSinceRaise = 1
	g.state.actionCount++
	g.appendLog(idx, "raise", fmt.Sprintf("%s raises to %d, pays %d (pot %d)", playerName(g.players, idx), newBet, need, g.state.pot), nil)
	g.advanceOrClose(idx)
}

// applyFold はフォールドを適用する。
func (g *Anaconda) applyFold(idx int) {
	g.players[idx].SetFolded(true)
	g.state.actionCount++
	g.appendLog(idx, "fold", fmt.Sprintf("%s folds", playerName(g.players, idx)), nil)
	g.advanceOrClose(idx)
}

// advanceOrClose は次の手番へ進めるか、ベッティングラウンドを締める。
func (g *Anaconda) advanceOrClose(fromIdx int) {
	if g.activeCount() <= 1 ||
		g.state.actedSinceRaise >= g.activeCount() ||
		g.state.actionCount >= anacondaMaxActions {
		g.closeBettingRound()
		return
	}
	g.state.currentPlayer = g.nextActive(fromIdx)
}

// closeBettingRound はベッティングラウンドを締め、次の公開へ進むかショーダウンで精算する。
func (g *Anaconda) closeBettingRound() {
	if g.activeCount() <= 1 || g.state.rollIndex >= AnacondaKeepSize {
		g.resolveShowdown()
		return
	}
	g.state.rollIndex++
	g.startBettingRound()
}

// resolveShowdown はロール終了時にポットを精算し、フェーズを Result へ遷移する。
// scored フラグで二重精算を防ぐ (フェーズ入場時に 1 回だけ発火)。
func (g *Anaconda) resolveShowdown() {
	if g.state.scored {
		return
	}
	active := g.activeSeats()
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
	g.state.rollIndex = AnacondaKeepSize
	g.state.scored = true
	g.state.phase = AnacondaPhaseResult
	g.checkGameEnd()
}

// setHumanResult は人間 (seat 0) の勝敗結果を設定する。
func (g *Anaconda) setHumanResult(winnerIdx int) {
	human := g.players[0]
	if human.GetOut() || human.GetFolded() {
		g.state.result = AnacondaResultNone
		return
	}
	if winnerIdx == 0 {
		g.state.result = AnacondaResultWin
	} else {
		g.state.result = AnacondaResultLose
	}
}

// checkGameEnd は停止条件 (規定ラウンド到達 or 参加可能者が規定数未満) を判定する。
func (g *Anaconda) checkGameEnd() {
	if g.state.roundNumber >= g.config.TargetRounds || g.solventCount() < AnacondaMinPlayerCount {
		g.endGame()
	}
}

// endGame はゲームを終了し、チップ最多のプレイヤーを勝者に設定する。
func (g *Anaconda) endGame() {
	g.state.gameEndFlag = true
	g.state.phase = AnacondaPhaseResult
	g.state.matchWinnerIdx = g.richestIdx()
	g.appendLog(g.state.matchWinnerIdx, "game_end",
		fmt.Sprintf("%s wins the game", playerName(g.players, g.state.matchWinnerIdx)), nil)
}

// --- CPU ---

// driveCPU は現在の手番が CPU の間、CPU のアクションを実行し続ける。人間の手番か
// ロール解決で停止する。
func (g *Anaconda) driveCPU() {
	guard := 0
	for g.state.phase == AnacondaPhaseRoll && !g.state.gameEndFlag {
		idx := g.state.currentPlayer
		if g.players[idx].GetIsHuman() {
			return
		}
		g.cpuAct(idx)
		guard++
		if guard > anacondaMaxActions*4 {
			g.closeBettingRound()
			return
		}
	}
}

// cpuAct は CPU が 1 アクション (call/raise/fold) を選ぶ。
func (g *Anaconda) cpuAct(idx int) {
	p := g.players[idx]
	cat, tb := g.evalPlayer(idx)
	need := g.state.currentBet - p.GetStreetBet()
	if need < 0 {
		need = 0
	}
	canRaise := g.state.raiseCount < AnacondaMaxRaises && p.GetChips() >= need+g.config.Ante
	high := 0
	if len(tb) > 0 {
		high = tb[0]
	}
	switch {
	case cat >= AnacondaThreeKind:
		if canRaise {
			g.applyRaise(idx)
			return
		}
		g.applyCall(idx)
	case cat >= AnacondaOnePair || high >= anacondaCpuRaiseHigh:
		if canRaise && cat >= AnacondaTwoPair && g.state.raiseCount < 2 {
			g.applyRaise(idx)
			return
		}
		g.applyCall(idx)
	case need == 0:
		g.applyCall(idx)
	case high >= anacondaCpuCallHigh && need <= g.config.Ante*2:
		g.applyCall(idx)
	default:
		g.applyFold(idx)
	}
}

// --- ポーカー手役評価 (インライン) ---

// anacondaRank はカードのランクを返す (A=14, K=13 ... 2=2)。
func anacondaRank(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// AnacondaEval は 5 枚の手の役カテゴリと比較用タイブレーク列 (比較順) を返す。
// カテゴリが同じならこの列を辞書順比較すれば勝敗が決まる。
//
//   - StraightFlush / Straight: タイブレーク = [ストレートの最高ランク] (A-2-3-4-5 は 5)
//   - その他: 出現回数の多い→ランクの高い順に並べた「相異なるランク」列
//     (例 FullHouse = [トリップ, ペア]、TwoPair = [高ペア, 低ペア, キッカー]、
//     HighCard / Flush = 5 ランクの降順)
func AnacondaEval(cards []*Card) (int, []int) {
	if len(cards) != AnacondaKeepSize {
		return -1, nil
	}
	ranks := make([]int, 0, AnacondaKeepSize)
	suits := make([]int, 0, AnacondaKeepSize)
	for _, c := range cards {
		if c == nil {
			return -1, nil
		}
		ranks = append(ranks, anacondaRank(c))
		suits = append(suits, c.GetDesign())
	}

	counts := make(map[int]int, AnacondaKeepSize)
	for _, r := range ranks {
		counts[r]++
	}
	type rc struct{ rank, cnt int }
	groups := make([]rc, 0, len(counts))
	for r, c := range counts {
		groups = append(groups, rc{r, c})
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].cnt != groups[j].cnt {
			return groups[i].cnt > groups[j].cnt
		}
		return groups[i].rank > groups[j].rank
	})
	tb := make([]int, 0, len(groups))
	for _, gr := range groups {
		tb = append(tb, gr.rank)
	}

	flush := anacondaIsFlush(suits)
	straight, straightHigh := anacondaStraight(ranks)

	switch {
	case straight && flush:
		return AnacondaStraightFlush, []int{straightHigh}
	case groups[0].cnt == 4:
		return AnacondaFourKind, tb
	case groups[0].cnt == 3 && len(groups) > 1 && groups[1].cnt == 2:
		return AnacondaFullHouse, tb
	case flush:
		return AnacondaFlush, tb
	case straight:
		return AnacondaStraight, []int{straightHigh}
	case groups[0].cnt == 3:
		return AnacondaThreeKind, tb
	case groups[0].cnt == 2 && len(groups) > 1 && groups[1].cnt == 2:
		return AnacondaTwoPair, tb
	case groups[0].cnt == 2:
		return AnacondaOnePair, tb
	default:
		return AnacondaHighCard, tb
	}
}

// anacondaIsFlush は 5 枚が同一スートかを返す。
func anacondaIsFlush(suits []int) bool {
	for i := 1; i < len(suits); i++ {
		if suits[i] != suits[0] {
			return false
		}
	}
	return true
}

// anacondaStraight は 5 枚がストレートかと最高ランクを返す。A-2-3-4-5 (ホイール) は最高 5、
// 10-J-Q-K-A は最高 14。それ以外のラップは不可。
func anacondaStraight(ranks []int) (bool, int) {
	uniq := make(map[int]bool, len(ranks))
	for _, r := range ranks {
		uniq[r] = true
	}
	if len(uniq) != AnacondaKeepSize {
		return false, 0
	}
	rs := make([]int, 0, AnacondaKeepSize)
	for r := range uniq {
		rs = append(rs, r)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(rs)))
	if rs[0]-rs[len(rs)-1] == AnacondaKeepSize-1 {
		return true, rs[0]
	}
	// ホイール A-2-3-4-5 → ランク列 {14,5,4,3,2}
	if rs[0] == 14 && rs[1] == 5 && rs[2] == 4 && rs[3] == 3 && rs[4] == 2 {
		return true, 5
	}
	return false, 0
}

// evalPlayer は playerIdx の手役を返す。
func (g *Anaconda) evalPlayer(idx int) (int, []int) {
	p := g.players[idx]
	cards := make([]*Card, 0, AnacondaKeepSize)
	for i := 0; i < p.GetCardsSize(); i++ {
		cards = append(cards, p.GetCard(i))
	}
	return AnacondaEval(cards)
}

// AnacondaCompare は手 a が手 b に勝てば 1、負ければ -1、引き分けは 0 を返す。
func AnacondaCompare(catA int, tbA []int, catB int, tbB []int) int {
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
func (g *Anaconda) bestHand(seats []int) int {
	best := seats[0]
	bestCat, bestTb := g.evalPlayer(best)
	for _, idx := range seats[1:] {
		cat, tb := g.evalPlayer(idx)
		if AnacondaCompare(cat, tb, bestCat, bestTb) > 0 {
			best, bestCat, bestTb = idx, cat, tb
		}
	}
	return best
}

// --- ヘルパー ---

// participantSeats は非脱落プレイヤーのインデックス列 (昇順) を返す。
func (g *Anaconda) participantSeats() []int {
	out := make([]int, 0, len(g.players))
	for i, p := range g.players {
		if !p.GetOut() {
			out = append(out, i)
		}
	}
	return out
}

// activeSeats は未フォールド・非脱落プレイヤーのインデックス列 (昇順) を返す。
func (g *Anaconda) activeSeats() []int {
	out := make([]int, 0, len(g.players))
	for i, p := range g.players {
		if !p.GetOut() && !p.GetFolded() {
			out = append(out, i)
		}
	}
	return out
}

// activeCount は未フォールド・非脱落プレイヤー数を返す。
func (g *Anaconda) activeCount() int {
	return countPlayers(g.players, func(p *AnacondaPlayer) bool { return !p.GetOut() && !p.GetFolded() })
}

// nextActive は from の次の未フォールド・非脱落プレイヤーを返す。
func (g *Anaconda) nextActive(from int) int {
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
func (g *Anaconda) solventCount() int {
	return countPlayers(g.players, func(p *AnacondaPlayer) bool { return !p.GetOut() && p.GetChips() >= g.config.Ante })
}

// richestIdx はチップが最多のプレイヤーのインデックスを返す (同数は座席番号の小さい方)。
func (g *Anaconda) richestIdx() int {
	return maxIndexBy(g.players, func(p *AnacondaPlayer) int { return p.GetChips() })
}

// anacondaValidateIndices はカードインデックス列の妥当性 (枚数・範囲・重複なし) を検証する。
func anacondaValidateIndices(indices []int, want, handSize int) error {
	if len(indices) != want {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("select exactly %d card(s)", want))
	}
	seen := make(map[int]bool, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= handSize {
			return NewDomainError(ErrInvalidPlay, "card index out of range")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidPlay, "duplicate card index")
		}
		seen[idx] = true
	}
	return nil
}

func (g *Anaconda) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.state.appendLog(playerIdx, actionType, detail, cards)
}

// --- 公開カード ---

// GetRevealedCards はフロント表示用に idx のプレイヤーの公開カードを返す。人間は常に自分の
// 全手札、CPU はロールフェーズで公開済みの枚数、結果フェーズでは非フォールド・非脱落なら全 5 枚。
func (g *Anaconda) GetRevealedCards(idx int) []*Card {
	if idx < 0 || idx >= len(g.players) {
		return nil
	}
	p := g.players[idx]
	if p.GetIsHuman() {
		return anacondaHandSlice(p, p.GetCardsSize())
	}
	switch g.state.phase {
	case AnacondaPhaseResult:
		if !p.GetOut() && !p.GetFolded() {
			return anacondaHandSlice(p, p.GetCardsSize())
		}
		return nil
	case AnacondaPhaseRoll:
		return anacondaHandSlice(p, g.state.rollIndex)
	default:
		return nil
	}
}

// anacondaHandSlice は p の手札の先頭 n 枚を返す。
func anacondaHandSlice(p *AnacondaPlayer, n int) []*Card {
	if n > p.GetCardsSize() {
		n = p.GetCardsSize()
	}
	if n <= 0 {
		return nil
	}
	out := make([]*Card, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, p.GetCard(i))
	}
	return out
}

// IsHandFullyRevealed は idx の手札が完全公開 (5 枚公開) されているかを返す。
func (g *Anaconda) IsHandFullyRevealed(idx int) bool {
	return len(g.GetRevealedCards(idx)) == AnacondaKeepSize
}

// --- Hint ---

// GetHint はフェーズに応じた助言を返す。
func (g *Anaconda) GetHint() *AnacondaHint {
	if g.state.gameEndFlag {
		return nil
	}
	human := g.players[0]
	if human.GetOut() {
		return nil
	}
	switch g.state.phase {
	case AnacondaPhasePass:
		if human.GetCardsSize() != AnacondaDealSize {
			return nil
		}
		return &AnacondaHint{Action: "pass", CardIndices: g.cpuPassIndices(human, g.state.passCount), Reason: "pass_weakest"}
	case AnacondaPhaseSet:
		if human.GetCardsSize() != AnacondaDealSize {
			return nil
		}
		return &AnacondaHint{Action: "keep", CardIndices: g.cpuBestKeepIndices(human), Reason: "keep_best"}
	case AnacondaPhaseRoll:
		if g.state.currentPlayer != 0 || human.GetFolded() {
			return nil
		}
		return g.rollHint()
	default:
		return nil
	}
}

// rollHint はロールフェーズの人間手番に call/raise/fold の助言を返す。
func (g *Anaconda) rollHint() *AnacondaHint {
	p := g.players[0]
	cat, tb := g.evalPlayer(0)
	high := 0
	if len(tb) > 0 {
		high = tb[0]
	}
	need := g.state.currentBet - p.GetStreetBet()
	switch {
	case cat >= AnacondaThreeKind:
		return &AnacondaHint{Action: "raise", Reason: "strong_hand"}
	case cat >= AnacondaOnePair || high >= anacondaCpuRaiseHigh:
		return &AnacondaHint{Action: "call", Reason: "medium_hand"}
	case need <= 0:
		return &AnacondaHint{Action: "call", Reason: "medium_hand"}
	default:
		return &AnacondaHint{Action: "fold", Reason: "weak_hand"}
	}
}

// --- 状態アクセサ ---

// GetPhase は現在のフェーズを返す。
func (g *Anaconda) GetPhase() AnacondaPhase { return g.state.phase }

// SetPhase はフェーズを設定する (テスト用)。
func (g *Anaconda) SetPhase(p AnacondaPhase) { g.state.phase = p }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *Anaconda) GetGameEndFlag() bool { return g.state.gameEndFlag }

// GetRoundNumber は現在のラウンド番号を返す。
func (g *Anaconda) GetRoundNumber() int { return g.state.roundNumber }

// GetDealerIdx はディーラーの座席番号を返す。
func (g *Anaconda) GetDealerIdx() int { return g.state.dealerIdx }

// SetDealerIdx はディーラーの座席番号を設定する (テスト用)。
func (g *Anaconda) SetDealerIdx(idx int) { g.state.dealerIdx = idx }

// GetCurrentPlayerIdx は現在の手番プレイヤーの座席番号を返す。
func (g *Anaconda) GetCurrentPlayerIdx() int { return g.state.currentPlayer }

// SetCurrentPlayerIdx は現在の手番プレイヤーを設定する (テスト用)。
func (g *Anaconda) SetCurrentPlayerIdx(idx int) { g.state.currentPlayer = idx }

// GetPassCount は現在のパスサブラウンドで渡す枚数を返す (3/2/1; パス外では 0)。
func (g *Anaconda) GetPassCount() int { return g.state.passCount }

// SetPassCount はパスサブラウンドの枚数を設定する (テスト用)。
func (g *Anaconda) SetPassCount(v int) { g.state.passCount = v }

// GetRollIndex はロールフェーズで公開済みの枚数を返す (0..AnacondaKeepSize)。
func (g *Anaconda) GetRollIndex() int { return g.state.rollIndex }

// SetRollIndex は公開済み枚数を設定する (テスト用)。
func (g *Anaconda) SetRollIndex(v int) { g.state.rollIndex = v }

// GetPot は現在のポットを返す。
func (g *Anaconda) GetPot() int { return g.state.pot }

// SetPot はポットを設定する (テスト用)。
func (g *Anaconda) SetPot(v int) { g.state.pot = v }

// GetCurrentBet は現在のストリートで必要な拠出額を返す。
func (g *Anaconda) GetCurrentBet() int { return g.state.currentBet }

// SetCurrentBet は現在のストリートで必要な拠出額を設定する (テスト用)。
func (g *Anaconda) SetCurrentBet(v int) { g.state.currentBet = v }

// GetRaiseCount はこのストリートのレイズ回数を返す。
func (g *Anaconda) GetRaiseCount() int { return g.state.raiseCount }

// GetMaxRaises は 1 ストリートあたりの最大レイズ回数を返す。
func (g *Anaconda) GetMaxRaises() int { return AnacondaMaxRaises }

// GetAnte はアンティ額を返す。
func (g *Anaconda) GetAnte() int { return g.config.Ante }

// GetWinnerIdx は直近ラウンドの勝者を返す (-1 = なし)。
func (g *Anaconda) GetWinnerIdx() int { return g.state.winnerIdx }

// GetMatchWinnerIdx はゲーム全体の勝者を返す (-1 = 未確定)。
func (g *Anaconda) GetMatchWinnerIdx() int { return g.state.matchWinnerIdx }

// GetResult は人間から見たラウンド結果を返す。
func (g *Anaconda) GetResult() AnacondaResult { return g.state.result }

// GetPlayerCnt はプレイヤー数を返す。
func (g *Anaconda) GetPlayerCnt() int { return len(g.players) }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *Anaconda) GetPlayer(i int) *AnacondaPlayer {
	return getPlayer(g.players, i)
}

// GetChips は人間 (seat 0) の保有チップを返す。
func (g *Anaconda) GetChips() int {
	return chipsOfFirst(g.players)
}

// IsHumanTurn は現在がロールフェーズの人間 (seat 0) 手番かどうかを返す。
func (g *Anaconda) IsHumanTurn() bool {
	if g.state.gameEndFlag {
		return false
	}
	switch g.state.phase {
	case AnacondaPhasePass, AnacondaPhaseSet:
		// Pass and Set are simultaneous phases: the human (seat 0) selects and
		// submits once, so they are always "on turn" until the phase advances.
		return !g.players[0].GetFolded() && !g.players[0].GetOut()
	case AnacondaPhaseRoll:
		idx := g.state.currentPlayer
		if idx < 0 || idx >= len(g.players) {
			return false
		}
		return g.players[idx].GetIsHuman()
	default:
		return false
	}
}

// CanRaise は現在の手番プレイヤーがレイズ可能か (回数上限未満かつ増分を払える) を返す。
func (g *Anaconda) CanRaise() bool {
	if !g.IsHumanTurn() || g.state.raiseCount >= AnacondaMaxRaises {
		return false
	}
	p := g.players[g.state.currentPlayer]
	need := (g.state.currentBet + g.config.Ante) - p.GetStreetBet()
	return p.GetChips() >= need
}

// GetConfig はローカルルール設定を返す。
func (g *Anaconda) GetConfig() AnacondaConfig { return g.config }

// SetConfig はローカルルール設定を変更する。
func (g *Anaconda) SetConfig(cfg AnacondaConfig) { g.config = cfg }

// GetActionLog は棋譜を返す。
func (g *Anaconda) GetActionLog() []*ActionLogEntry { return g.state.actionLog }

// ResolveShowdownForTest は手札を設定済みの状態でショーダウンを解決する (テスト用)。
func (g *Anaconda) ResolveShowdownForTest() { g.resolveShowdown() }

// EnterRollForTest はロールフェーズ入場 (最初のベッティングラウンド開始 + CPU 進行) を
// 手動で発火する (テスト用)。
func (g *Anaconda) EnterRollForTest() { g.enterRollPhase() }

// --- JSON Serialization ---

// anacondaJSON is the JSON wire format for Anaconda.
type anacondaJSON struct {
	TrumpCards      *TrumpCards       `json:"tc"`
	Players         []*AnacondaPlayer `json:"ps"`
	Config          AnacondaConfig    `json:"cf"`
	Phase           AnacondaPhase     `json:"ph"`
	RoundNumber     int               `json:"rn"`
	DealerIdx       int               `json:"di"`
	CurrentPlayer   int               `json:"ci"`
	PassCount       int               `json:"pn"`
	RollIndex       int               `json:"ri"`
	Pot             int               `json:"pt"`
	CurrentBet      int               `json:"cb"`
	RaiseCount      int               `json:"rc"`
	ActedSinceRaise int               `json:"as"`
	ActionCount     int               `json:"ac"`
	WinnerIdx       int               `json:"wi"`
	MatchWinnerIdx  int               `json:"mw"`
	Result          AnacondaResult    `json:"re"`
	GameEndFlag     bool              `json:"ge"`
	Scored          bool              `json:"sc"`
	ActionLog       []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Anaconda) MarshalJSON() ([]byte, error) {
	return json.Marshal(anacondaJSON{
		TrumpCards:      g.trumpCards,
		Players:         g.players,
		Config:          g.config,
		Phase:           g.state.phase,
		RoundNumber:     g.state.roundNumber,
		DealerIdx:       g.state.dealerIdx,
		CurrentPlayer:   g.state.currentPlayer,
		PassCount:       g.state.passCount,
		RollIndex:       g.state.rollIndex,
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

// anacondaValidPhase は有効なフェーズかどうか。
func anacondaValidPhase(p AnacondaPhase) bool {
	return p >= AnacondaPhasePass && p <= AnacondaPhaseResult
}

// UnmarshalJSON implements json.Unmarshaler. 不正な永続化データを拒否するための
// バリデーションを行う (KV 復元時のインデックス範囲・スライス健全性)。
func (g *Anaconda) UnmarshalJSON(data []byte) error {
	var j anacondaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("anaconda: invalid config: %w", err)
	}
	n := len(j.Players)
	if n < AnacondaMinPlayerCount || n > AnacondaMaxPlayerCount || n != j.Config.PlayerCount {
		return errAnacondaSnapshot
	}
	if len(j.ActionLog) > anacondaMaxSliceLen {
		return errAnacondaSnapshot
	}
	if !anacondaValidPhase(j.Phase) {
		return errAnacondaSnapshot
	}
	if j.RoundNumber < 1 || j.Pot < 0 || j.CurrentBet < 0 || j.RaiseCount < 0 ||
		j.ActedSinceRaise < 0 || j.ActionCount < 0 {
		return errAnacondaSnapshot
	}
	if j.PassCount < 0 || j.PassCount > AnacondaPassStart {
		return errAnacondaSnapshot
	}
	if j.RollIndex < 0 || j.RollIndex > AnacondaKeepSize {
		return errAnacondaSnapshot
	}
	if j.DealerIdx < 0 || j.DealerIdx >= n || j.CurrentPlayer < 0 || j.CurrentPlayer >= n {
		return errAnacondaSnapshot
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= n || j.MatchWinnerIdx < -1 || j.MatchWinnerIdx >= n {
		return errAnacondaSnapshot
	}
	if j.Result < AnacondaResultLose || j.Result > AnacondaResultWin {
		return errAnacondaSnapshot
	}
	for _, p := range j.Players {
		if p == nil {
			return errAnacondaSnapshot
		}
	}
	for _, e := range j.ActionLog {
		if e == nil {
			return errAnacondaSnapshot
		}
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	g.config = j.Config
	g.state = anacondaState{
		phase:           j.Phase,
		roundNumber:     j.RoundNumber,
		dealerIdx:       j.DealerIdx,
		currentPlayer:   j.CurrentPlayer,
		passCount:       j.PassCount,
		rollIndex:       j.RollIndex,
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
