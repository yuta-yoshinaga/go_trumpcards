//go:build !js || !wasm || extra2

// Package domain (Cuckoo) implements the European life-based survival game
// Cuckoo, also known as Chase the Ace or Ranter-Go-Round.
//
// Rules (faithful but bounded):
//
//   - 4 players (seat 0 human + 3 CPU). A standard 52-card deck. Ranks are
//     compared by raw value: Ace = 1 (lowest) … King = 13 (highest).
//   - Each player is dealt exactly ONE card and starts with InitialLives lives
//     (CuckooStartLives = 3 by default).
//   - Players act once per round in seat order starting from the dealer's left
//     and ending with the dealer. On their turn a player either KEEP their card
//     or SWAP it with their right-hand neighbour (the next active player).
//   - A neighbour holding a King may REFUSE the swap: the refusal reveals the
//     King and blocks the exchange (the requester keeps their card). CPUs holding
//     a King always auto-refuse; the human neighbour decides via PlayerRefuse /
//     PlayerAcceptSwap.
//   - Simplification of the dealer's special option: in traditional Cuckoo the
//     dealer may cut the deck for a fresh card. Here the LAST player to act (the
//     dealer) may SWAP with the stock instead of a neighbour — drawing a new
//     card from the deck. There is no neighbour to refuse a stock swap.
//   - After every active player has acted the round ends: the lowest card value
//     among active players is found and every player holding that lowest value
//     loses one life (ties → all of them lose a life). Players reaching 0 lives
//     are eliminated.
//   - Rounds repeat (re-dealing one card each) until a single player remains —
//     that player wins. Because at least one life is always lost each round, a
//     full-CPU game terminates; a defensive iteration guard backs this up.
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// CuckooPlayerCnt Cuckoo プレイヤー数 (人間 1 + CPU 3)
const CuckooPlayerCnt = 4

// CuckooKingValue King のランク値 (スワップ拒否が可能)
const CuckooKingValue = 13

// cuckooCpuSwapThreshold CPU がスワップを検討する手札の上限値 (これ以下なら交換を試みる)
const cuckooCpuSwapThreshold = 7

// cuckooMaxRounds ラウンド数の防御的上限 (毎ラウンド必ずライフが減るため通常到達しない)
const cuckooMaxRounds = 1000

// CuckooPhase ゲームフェーズ
type CuckooPhase int

// Cuckoo のフェーズ定数
const (
	// CuckooPhaseTurn 手番プレイヤーが Keep / Swap を選ぶフェーズ
	CuckooPhaseTurn CuckooPhase = 0
	// CuckooPhaseRefuse King を持つ隣人が拒否するか受けるかを選ぶフェーズ
	CuckooPhaseRefuse CuckooPhase = 1
	// CuckooPhaseRoundEnd ラウンド終了フェーズ
	CuckooPhaseRoundEnd CuckooPhase = 2
	// CuckooPhaseGameEnd ゲーム終了フェーズ
	CuckooPhaseGameEnd CuckooPhase = 3
)

// Cuckoo Cuckoo (カッコー / Chase the Ace) ゲームクラス
type Cuckoo struct {
	trumpCards       *TrumpCards
	players          []*CuckooPlayer
	config           CuckooConfig
	phase            CuckooPhase
	currentPlayerIdx int     // 現在の手番プレイヤー
	stock            []*Card // 残り山札 (ディーラーのストックスワップ用)
	dealerIdx        int     // ディーラー (最後に手番が回る)
	actedCount       int     // 当該ラウンドで手番を終えたアクティブプレイヤー数
	pendingSwapFrom  int     // Refuse フェーズで交換を要求している側 (-1=なし)
	pendingSwapTo    int     // Refuse フェーズで King を持つ隣人 (-1=なし)
	revealedKings    []bool  // 拒否で公開された King を保持中か
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	roundLowest      int   // 直近ラウンドの最低カード値 (-1=未確定)
	roundLosers      []int // 直近ラウンドでライフを失ったプレイヤー
	actionLog        []*ActionLogEntry
	rng              *rand.Rand
}

// NewCuckoo コンストラクタ
func NewCuckoo(trumpCards *TrumpCards, players []*CuckooPlayer, config CuckooConfig) *Cuckoo {
	return &Cuckoo{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		winnerIdx:       -1,
		pendingSwapFrom: -1,
		pendingSwapTo:   -1,
		roundLowest:     -1,
		revealedKings:   make([]bool, len(players)),
		roundLosers:     make([]int, 0),
		rng:             rand.New(rand.NewSource(rand.Int63())),
	}
}

// SetRand テスト用に乱数源を差し替える
func (g *Cuckoo) SetRand(r *rand.Rand) { g.rng = r }

// NewDefaultCuckoo returns Cuckoo with the standard 4-player setup
// (1 human, 3 CPU) and DefaultCuckooConfig. Single source of truth for the
// CUI, Web, and Worker construction sites.
func NewDefaultCuckoo() *Cuckoo {
	players := []*CuckooPlayer{
		NewCuckooPlayer(true),
		NewCuckooPlayer(false),
		NewCuckooPlayer(false),
		NewCuckooPlayer(false),
	}
	return NewCuckoo(NewTrumpCards(0), players, DefaultCuckooConfig())
}

// Reset ゲームを初期化する
func (g *Cuckoo) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.actionLog = nil

	for _, p := range g.players {
		p.Reset()
		p.SetLives(g.config.InitialLives)
	}

	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Cuckoo) NextRound() {
	if g.phase != CuckooPhaseRoundEnd {
		return
	}
	if g.roundNumber >= cuckooMaxRounds {
		g.forceGameEnd()
		return
	}
	g.roundNumber++
	g.dealerIdx = g.nextActiveIdx(g.dealerIdx)
	g.startRound()
}

// startRound 1 ラウンド分のセットアップを行う (脱落者を除いて 1 枚ずつ配牌する)
func (g *Cuckoo) startRound() {
	g.pendingSwapFrom = -1
	g.pendingSwapTo = -1
	g.roundLowest = -1
	g.roundLosers = make([]int, 0)
	g.actedCount = 0
	for i := range g.revealedKings {
		g.revealedKings[i] = false
	}

	g.dealRound()

	g.currentPlayerIdx = g.firstActiveFrom(g.nextActiveIdx(g.dealerIdx))
	g.phase = CuckooPhaseTurn
}

// dealRound 山札を作り直し、脱落していないプレイヤーに 1 枚ずつ配る
func (g *Cuckoo) dealRound() {
	g.trumpCards.Replenish()
	g.stock = make([]*Card, 0, g.trumpCards.GetTotalCount())
	for {
		card := g.trumpCards.DrawCard()
		if card == nil {
			break
		}
		g.stock = append(g.stock, card)
	}
	g.rng.Shuffle(len(g.stock), func(i, j int) {
		g.stock[i], g.stock[j] = g.stock[j], g.stock[i]
	})

	for _, p := range g.players {
		p.Reset()
		if p.IsEliminated() {
			continue
		}
		if len(g.stock) > 0 {
			card := g.stock[len(g.stock)-1]
			g.stock = g.stock[:len(g.stock)-1]
			p.AddCard(card)
		}
	}
}

// --- Human actions ---

// PlayerKeep 人間プレイヤーが手札を保持してターンを進める
func (g *Cuckoo) PlayerKeep() error {
	if err := g.guardHumanTurn(); err != nil {
		return err
	}
	g.keep(g.currentPlayerIdx)
	return nil
}

// PlayerSwap 人間プレイヤーがスワップを試みる (隣人 or ディーラーはストック)
func (g *Cuckoo) PlayerSwap() error {
	if err := g.guardHumanTurn(); err != nil {
		return err
	}
	g.attemptSwap(g.currentPlayerIdx)
	return nil
}

// PlayerRefuse King を持つ人間の隣人がスワップを拒否する
func (g *Cuckoo) PlayerRefuse() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CuckooPhaseRefuse {
		return ErrWrongPhase
	}
	if g.pendingSwapTo < 0 || !g.players[g.pendingSwapTo].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.resolveRefuse(true)
	return nil
}

// PlayerAcceptSwap King を持つ人間の隣人がスワップを受け入れる
func (g *Cuckoo) PlayerAcceptSwap() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CuckooPhaseRefuse {
		return ErrWrongPhase
	}
	if g.pendingSwapTo < 0 || !g.players[g.pendingSwapTo].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.resolveRefuse(false)
	return nil
}

// guardHumanTurn ターンフェーズの人間操作に共通する前提チェック
func (g *Cuckoo) guardHumanTurn() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CuckooPhaseTurn {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return nil
}

// --- Shared move helpers ---

// keep 手札を保持してターンを進める
func (g *Cuckoo) keep(idx int) {
	g.appendLog(idx, "keep", fmt.Sprintf("%s keeps", g.playerName(idx)), nil)
	g.advanceTurn()
}

// attemptSwap idx がスワップを試みる。ディーラー (最後の手番) はストックと交換、
// それ以外は次のアクティブプレイヤーと交換する。隣人が King を持つ場合は
// Refuse フェーズへ移行する (CPU は自動拒否)。
func (g *Cuckoo) attemptSwap(idx int) {
	if idx == g.dealerIdx {
		g.swapWithStock(idx)
		g.advanceTurn()
		return
	}
	neighbour := g.nextActiveIdx(idx)
	if neighbour == idx {
		// 他にアクティブプレイヤーがいない: 保持と同義
		g.keep(idx)
		return
	}
	g.pendingSwapFrom = idx
	g.pendingSwapTo = neighbour
	if g.players[neighbour].HasKing() {
		if g.players[neighbour].GetIsHuman() {
			g.phase = CuckooPhaseRefuse
			return
		}
		// CPU は King で自動拒否
		g.resolveRefuse(true)
		return
	}
	g.performSwap(idx, neighbour)
	g.clearPending()
	g.advanceTurn()
}

// performSwap 2 プレイヤーの手札を交換する
func (g *Cuckoo) performSwap(a, b int) {
	ca := g.players[a].Card()
	cb := g.players[b].Card()
	g.players[a].SetCard(cb)
	g.players[b].SetCard(ca)
	g.appendLog(a, "swap", fmt.Sprintf("%s swaps with %s", g.playerName(a), g.playerName(b)), nil)
}

// swapWithStock ディーラーが山札から新しいカードを引いて交換する
func (g *Cuckoo) swapWithStock(idx int) {
	if len(g.stock) == 0 {
		g.appendLog(idx, "keep", fmt.Sprintf("%s keeps (stock empty)", g.playerName(idx)), nil)
		return
	}
	newCard := g.stock[len(g.stock)-1]
	g.stock = g.stock[:len(g.stock)-1]
	old := g.players[idx].Card()
	g.players[idx].SetCard(newCard)
	if old != nil {
		g.stock = append([]*Card{old}, g.stock...)
	}
	g.appendLog(idx, "swap_stock", fmt.Sprintf("%s swaps with the stock", g.playerName(idx)), nil)
}

// resolveRefuse 拒否フェーズを解決する。refused=true なら King 公開で交換は不成立、
// false なら交換成立。いずれもターンを進める。
func (g *Cuckoo) resolveRefuse(refused bool) {
	from := g.pendingSwapFrom
	to := g.pendingSwapTo
	if refused {
		if to >= 0 && to < len(g.revealedKings) {
			g.revealedKings[to] = true
		}
		g.appendLog(to, "refuse", fmt.Sprintf("%s reveals a King and refuses", g.playerName(to)), nil)
	} else {
		g.performSwap(from, to)
	}
	g.clearPending()
	g.advanceTurn()
}

// clearPending 保留中のスワップ状態をクリアする
func (g *Cuckoo) clearPending() {
	g.pendingSwapFrom = -1
	g.pendingSwapTo = -1
}

// --- CPU ---

// CpuPlay 現在の手番が CPU の場合に 1 アクション実行する。
// ターンフェーズでは Keep/Swap を判断し、Refuse フェーズ (CPU 隣人) は
// 通常 attemptSwap 内で自動解決されるため到達しないが防御的に処理する。
func (g *Cuckoo) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	switch g.phase {
	case CuckooPhaseTurn:
		if g.players[g.currentPlayerIdx].GetIsHuman() {
			return
		}
		g.cpuTurn()
	case CuckooPhaseRefuse:
		if g.pendingSwapTo < 0 || g.players[g.pendingSwapTo].GetIsHuman() {
			return
		}
		g.resolveRefuse(true)
	}
}

// cpuTurn CPU の Keep/Swap 判断 (低い札なら交換を試みる)
func (g *Cuckoo) cpuTurn() {
	idx := g.currentPlayerIdx
	if g.cpuWantsSwap(idx) {
		g.attemptSwap(idx)
		return
	}
	g.keep(idx)
}

// cpuWantsSwap CPU がスワップを望むか。低い手札 (<=閾値) で交換を試みる。
// Easy はランダム、Hard は閾値を厳しめに。
func (g *Cuckoo) cpuWantsSwap(idx int) bool {
	v := g.players[idx].CardValue()
	switch g.config.CpuDifficulty {
	case CuckooCpuDifficultyEasy:
		return g.rng.Intn(2) == 0
	case CuckooCpuDifficultyHard:
		return v <= cuckooCpuSwapThreshold+1
	default:
		return v <= cuckooCpuSwapThreshold
	}
}

// --- Round / game resolution ---

// advanceTurn 次のアクティブプレイヤーへ。全アクティブプレイヤーが手番を
// 終えたらラウンドを精算する。
func (g *Cuckoo) advanceTurn() {
	g.actedCount++
	if g.actedCount >= g.activeCount() {
		g.endRound()
		return
	}
	g.currentPlayerIdx = g.nextActiveIdx(g.currentPlayerIdx)
	g.phase = CuckooPhaseTurn
}

// endRound 最低カード値のプレイヤーがライフを失い、脱落判定を行う
func (g *Cuckoo) endRound() {
	lowest := -1
	for _, p := range g.players {
		if p.IsEliminated() {
			continue
		}
		v := p.CardValue()
		if lowest < 0 || v < lowest {
			lowest = v
		}
	}
	g.roundLowest = lowest

	g.roundLosers = make([]int, 0)
	for i, p := range g.players {
		if p.IsEliminated() {
			continue
		}
		if p.CardValue() == lowest {
			p.LoseLife()
			g.roundLosers = append(g.roundLosers, i)
			g.appendLog(i, "lose_life", fmt.Sprintf("%s loses a life (lowest: %d)", g.playerName(i), lowest), nil)
		}
	}

	// 総崩れ防止のタイブレーク: 残った全員が最低値で並び、このラウンドで全員が
	// 脱落すると勝者が決まらなくなる。その場合は並んだ中からスート最強の 1 人を
	// 生き残らせ、単独勝者を保証する (サドンデス)。
	if g.activeCount() == 0 && len(g.roundLosers) > 0 {
		survivor := g.roundLosers[0]
		for _, i := range g.roundLosers {
			si, ci := g.players[survivor].Card(), g.players[i].Card()
			if ci != nil && (si == nil || ci.GetDesign() > si.GetDesign()) {
				survivor = i
			}
		}
		g.players[survivor].SetLives(1)
		g.appendLog(survivor, "survive", fmt.Sprintf("%s survives the tie-break", g.playerName(survivor)), nil)
	}

	g.finishRound()
}

// finishRound ゲーム終了判定を行い、続行ならラウンド終了フェーズへ移る
func (g *Cuckoo) finishRound() {
	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = CuckooPhaseRoundEnd
	}
}

// checkGameEnd 残りプレイヤーが 1 人以下、または人間が脱落したらゲーム終了
func (g *Cuckoo) checkGameEnd() {
	active := g.activeCount()
	humanIdx := g.humanIdx()
	humanOut := humanIdx >= 0 && g.players[humanIdx].IsEliminated()

	if active > 1 && !humanOut {
		return
	}

	g.gameEndFlag = true
	g.phase = CuckooPhaseGameEnd
	g.winnerIdx = g.leaderIdx()
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", g.playerName(g.winnerIdx)), nil)
}

// forceGameEnd 防御的なラウンド上限到達時に強制終了する
func (g *Cuckoo) forceGameEnd() {
	g.gameEndFlag = true
	g.phase = CuckooPhaseGameEnd
	g.winnerIdx = g.leaderIdx()
}

// leaderIdx 最もライフが多いプレイヤー (同点は若いインデックス) を返す
func (g *Cuckoo) leaderIdx() int {
	best := 0
	bestLives := g.players[0].GetLives()
	for i := 1; i < len(g.players); i++ {
		if g.players[i].GetLives() > bestLives {
			bestLives = g.players[i].GetLives()
			best = i
		}
	}
	return best
}

// activeCount 脱落していないプレイヤー数を返す
func (g *Cuckoo) activeCount() int {
	n := 0
	for _, p := range g.players {
		if !p.IsEliminated() {
			n++
		}
	}
	return n
}

// firstActiveFrom from を含めて最初のアクティブプレイヤーのインデックスを返す
func (g *Cuckoo) firstActiveFrom(from int) int {
	n := len(g.players)
	for step := 0; step < n; step++ {
		idx := (from + step) % n
		if !g.players[idx].IsEliminated() {
			return idx
		}
	}
	return from
}

// nextActiveIdx from の次のアクティブプレイヤーのインデックスを返す
func (g *Cuckoo) nextActiveIdx(from int) int {
	n := len(g.players)
	for step := 1; step <= n; step++ {
		idx := (from + step) % n
		if !g.players[idx].IsEliminated() {
			return idx
		}
	}
	return from
}

// humanIdx 人間プレイヤーのインデックスを返す (-1 = 不在)
func (g *Cuckoo) humanIdx() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// --- Getters ---

// GetPhase 現在のフェーズを取得する
func (g *Cuckoo) GetPhase() CuckooPhase { return g.phase }

// SetPhase フェーズを設定する (テスト用)
func (g *Cuckoo) SetPhase(phase CuckooPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号を取得する
func (g *Cuckoo) GetRoundNumber() int { return g.roundNumber }

// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
func (g *Cuckoo) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックスを設定する (テスト用)
func (g *Cuckoo) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDealerIdx ディーラーインデックスを取得する
func (g *Cuckoo) GetDealerIdx() int { return g.dealerIdx }

// GetStockCount 残り山札の枚数を取得する
func (g *Cuckoo) GetStockCount() int { return len(g.stock) }

// GetGameEndFlag ゲーム終了フラグを取得する
func (g *Cuckoo) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックスを取得する (-1 = 未確定)
func (g *Cuckoo) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数を取得する
func (g *Cuckoo) GetPlayerCnt() int { return len(g.players) }

// GetPlayer 指定インデックスのプレイヤーを取得する
func (g *Cuckoo) GetPlayer(i int) *CuckooPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在の手番が人間かを返す (ターン or 拒否フェーズ)
func (g *Cuckoo) IsHumanTurn() bool {
	switch g.phase {
	case CuckooPhaseTurn:
		if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
			return false
		}
		return g.players[g.currentPlayerIdx].GetIsHuman()
	case CuckooPhaseRefuse:
		if g.pendingSwapTo < 0 || g.pendingSwapTo >= len(g.players) {
			return false
		}
		return g.players[g.pendingSwapTo].GetIsHuman()
	default:
		return false
	}
}

// GetPendingSwapFrom スワップを要求中のプレイヤーインデックスを取得する (-1=なし)
func (g *Cuckoo) GetPendingSwapFrom() int { return g.pendingSwapFrom }

// GetPendingSwapTo スワップ要求先 (King 保持の隣人) を取得する (-1=なし)
func (g *Cuckoo) GetPendingSwapTo() int { return g.pendingSwapTo }

// IsKingRevealed 指定プレイヤーの King が拒否により公開されているかを取得する
func (g *Cuckoo) IsKingRevealed(i int) bool {
	if i < 0 || i >= len(g.revealedKings) {
		return false
	}
	return g.revealedKings[i]
}

// GetRoundLowest 直近ラウンドの最低カード値を取得する (-1=未確定)
func (g *Cuckoo) GetRoundLowest() int { return g.roundLowest }

// GetRoundLosers 直近ラウンドでライフを失ったプレイヤーのインデックス一覧を取得する
func (g *Cuckoo) GetRoundLosers() []int { return g.roundLosers }

// GetConfig ゲーム設定を取得する
func (g *Cuckoo) GetConfig() CuckooConfig { return g.config }

// SetConfig ゲーム設定を設定する
func (g *Cuckoo) SetConfig(cfg CuckooConfig) { g.config = cfg }

// GetActionLog 棋譜を取得する
func (g *Cuckoo) GetActionLog() []*ActionLogEntry { return g.actionLog }

// playerName プレイヤー名を返す
func (g *Cuckoo) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// appendLog 棋譜にエントリを追加する
func (g *Cuckoo) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: len(g.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- JSON ---

// cuckooJSON is the JSON wire format for Cuckoo.
type cuckooJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*CuckooPlayer   `json:"pl"`
	Config           CuckooConfig      `json:"cf"`
	Phase            CuckooPhase       `json:"ph"`
	CurrentPlayerIdx int               `json:"ci"`
	Stock            []*Card           `json:"st"`
	DealerIdx        int               `json:"di"`
	ActedCount       int               `json:"ac"`
	PendingSwapFrom  int               `json:"pf"`
	PendingSwapTo    int               `json:"pt"`
	RevealedKings    []bool            `json:"rk"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	RoundNumber      int               `json:"rn"`
	RoundLowest      int               `json:"rlo"`
	RoundLosers      []int             `json:"rl"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// cuckooMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const cuckooMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (g *Cuckoo) MarshalJSON() ([]byte, error) {
	return json.Marshal(cuckooJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		Stock:            g.stock,
		DealerIdx:        g.dealerIdx,
		ActedCount:       g.actedCount,
		PendingSwapFrom:  g.pendingSwapFrom,
		PendingSwapTo:    g.pendingSwapTo,
		RevealedKings:    g.revealedKings,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		RoundNumber:      g.roundNumber,
		RoundLowest:      g.roundLowest,
		RoundLosers:      g.roundLosers,
		ActionLog:        g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler with defensive validation.
func (g *Cuckoo) UnmarshalJSON(data []byte) error {
	var j cuckooJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > cuckooMaxSliceLen || len(j.Stock) > cuckooMaxSliceLen ||
		len(j.ActionLog) > cuckooMaxSliceLen || len(j.RoundLosers) > cuckooMaxSliceLen ||
		len(j.RevealedKings) > cuckooMaxSliceLen {
		return fmt.Errorf("cuckoo: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("cuckoo: invalid config: %w", err)
	}
	if len(j.Players) != CuckooPlayerCnt {
		return fmt.Errorf("cuckoo: invalid player count: %d", len(j.Players))
	}
	for i, p := range j.Players {
		if p == nil {
			return fmt.Errorf("cuckoo: player %d is nil", i)
		}
	}
	if j.Phase < CuckooPhaseTurn || j.Phase > CuckooPhaseGameEnd {
		return fmt.Errorf("cuckoo: invalid phase: %d", j.Phase)
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= CuckooPlayerCnt {
		return fmt.Errorf("cuckoo: current player index out of range: %d", j.CurrentPlayerIdx)
	}
	if j.DealerIdx < 0 || j.DealerIdx >= CuckooPlayerCnt {
		return fmt.Errorf("cuckoo: dealer index out of range: %d", j.DealerIdx)
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= CuckooPlayerCnt {
		return fmt.Errorf("cuckoo: winner index out of range: %d", j.WinnerIdx)
	}
	if j.PendingSwapFrom < -1 || j.PendingSwapFrom >= CuckooPlayerCnt {
		return fmt.Errorf("cuckoo: pending swap from index out of range: %d", j.PendingSwapFrom)
	}
	if j.PendingSwapTo < -1 || j.PendingSwapTo >= CuckooPlayerCnt {
		return fmt.Errorf("cuckoo: pending swap to index out of range: %d", j.PendingSwapTo)
	}
	for _, idx := range j.RoundLosers {
		if idx < 0 || idx >= CuckooPlayerCnt {
			return fmt.Errorf("cuckoo: round loser index out of range: %d", idx)
		}
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.stock = j.Stock
	if g.stock == nil {
		g.stock = make([]*Card, 0)
	}
	g.dealerIdx = j.DealerIdx
	g.actedCount = j.ActedCount
	g.pendingSwapFrom = j.PendingSwapFrom
	g.pendingSwapTo = j.PendingSwapTo
	g.revealedKings = j.RevealedKings
	if len(g.revealedKings) != CuckooPlayerCnt {
		g.revealedKings = make([]bool, CuckooPlayerCnt)
	}
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.roundNumber = j.RoundNumber
	g.roundLowest = j.RoundLowest
	g.roundLosers = j.RoundLosers
	if g.roundLosers == nil {
		g.roundLosers = make([]int, 0)
	}
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(rand.Int63()))
	}
	return nil
}
