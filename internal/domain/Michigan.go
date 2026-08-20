//go:build !js || !wasm || extra4

// Package domain ミシガン (Michigan / Newmarket) のドメインモデル。
//
// ミシガン (別名 Newmarket、Boodle、Chicago) は「ストップ (stops)」系のギャンブル・
// パーティゲーム。中央に 4 枚の「ブードル (boodle)」札を表向きに置き、各プレイヤーは
// アンティをブードルに分配して賭ける。その後、標準 52 枚を全員 + 1 つの伏せた
// 「デッドハンド (widow)」に配り切り、同スートの昇順シーケンスでカードを出していく。
// ブードルと一致する札を出したプレイヤーはそのブードルのチップを総取りする。
//
// # 1 ラウンドの流れ
//
//  1. ベッティング: 全員がアンティ分のチップを 4 つのブードルに分配して積む。
//     チップは獲得されるまでラウンドをまたいで持ち越される。
//  2. プレイ (ストップの仕組み):
//     a. ディーラーの左隣が「リード」する — 手札のいずれか 1 スートの最小札を出す。
//     b. シーケンスは同スートで 1 つずつ上がる (♥3 → ♥4 → ♥5 ...)。次の札を持つ
//     プレイヤーに手番が移り (席順ではなく所持者へ)、そのプレイヤーが出す。
//     c. 次の札がデッドハンドにある、または K を超えたら「ストップ」。最後に出した
//     プレイヤーが新しいシーケンスを始める (再び任意スートの最小札を出す)。
//     d. ブードルと一致する札を出したら、そのブードルのチップを即座に総取りする。
//  3. 誰かが手札を出し切った瞬間にラウンド終了。未獲得のブードルは持ち越し。
//
// # スコア
//
// チップ。規定ラウンド数を消化するとゲーム終了し、チップ最多のプレイヤーが勝者。
//
// 本実装は extra ワーカーから到達可能なよう、シーケンス (ストップ) ロジックを
// すべてインラインで持つ (Sevens は casino/classic ビルドタグで到達不可のため)。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

// MichiganPhase はゲームフェーズ。ワイヤー値はフロントエンドの enum と一致させる。
type MichiganPhase int

// Michigan のフェーズ定数。
const (
	// MichiganPhaseBet ベッティング中 (人間のブードル賭け待ち)。ワイヤー値 0。
	MichiganPhaseBet MichiganPhase = 0
	// MichiganPhasePlay プレイ中 (ストップ・シーケンス)。ワイヤー値 1。
	MichiganPhasePlay MichiganPhase = 1
	// MichiganPhaseResult ラウンド解決済み (結果表示; 次ラウンド待ち or ゲーム終了)。ワイヤー値 2。
	MichiganPhaseResult MichiganPhase = 2
)

// MichiganResult は人間プレイヤーから見たラウンド結果。
// GameResult は共有ファイル internal/domain/game_result.go に移動したので到達可能に
// なったが、この型名は JSON ペイロードに出るため統合していない（#4462）。値は
// GameResult と同一。
type MichiganResult int

const (
	// MichiganResultLose 負け (他プレイヤーの純増がより大きい)
	MichiganResultLose MichiganResult = -1
	// MichiganResultNone 結果なし (誰も利益を出していない / 引き分け)
	MichiganResultNone MichiganResult = 0
	// MichiganResultWin 勝ち (人間の純増が最大かつプラス)
	MichiganResultWin MichiganResult = 1
)

// michiganMaxSliceLen はデシリアライズ時のスライス長の上限。
const michiganMaxSliceLen = 5000

// michiganMaxPlays は 1 ラウンドあたりの安全網となる最大プレイ数 (52 枚 + 余裕)。
const michiganMaxPlays = 200

// デシリアライズ検証用のセンチネルエラー。
var (
	errMichiganSnapshot      = errors.New("michigan: invalid serialised game state")
	errMichiganInvalidPlayer = errors.New("michigan: invalid player state")
)

// MichiganHint はヒント情報 (人間へのプレイ助言)。
type MichiganHint struct {
	CardIndex int    // 推奨カードの手札インデックス
	Reason    string // ヒント理由キー ("forced"/"claim_boodle"/"lead_low")
}

// michiganState はゲーム進行状態。
type michiganState struct {
	phase           MichiganPhase
	roundNumber     int
	dealerIdx       int
	currentPlayer   int
	leadPlayerIdx   int // このラウンドの最初のリード席 (ディーラーの左隣)
	seqSuit         int // 現在のシーケンスのスート (0 = 新シーケンス待ち, 1..4 = 進行中)
	seqHighValue    int // 現在のシーケンスで出た最大値 (0 = なし)
	lastPlayerIdx   int // 直近に出したプレイヤー (ストップ時に新シーケンスを始める)
	boodles         []*MichiganBoodle
	deadHand        []*Card // 伏せたデッドハンド (誰も持たず、決してプレイされない)
	roundStartChips []int   // ラウンド開始時 (賭け前) の各プレイヤーのチップ
	humanBetPlaced  bool    // 人間がこのラウンドのブードル賭けを済ませたか
	winnerIdx       int     // 直近ラウンドで手札を出し切ったプレイヤー (-1 = なし)
	matchWinnerIdx  int     // ゲーム全体の勝者 (-1 = 未確定)
	result          MichiganResult
	gameEndFlag     bool
	scored          bool // ラウンド結果を確定済みか (二重確定防止)
	actionLogBase
}

// Michigan はミシガンの状態を保持する集約ルート。
type Michigan struct {
	trumpCards *TrumpCards
	players    []*MichiganPlayer
	config     MichiganConfig
	state      michiganState
}

// NewMichigan はコンストラクタ。
func NewMichigan(trumpCards *TrumpCards, players []*MichiganPlayer, config MichiganConfig) *Michigan {
	return &Michigan{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		state: michiganState{
			phase:          MichiganPhaseBet,
			winnerIdx:      -1,
			matchWinnerIdx: -1,
			boodles:        newMichiganBoodles(),
			actionLogBase:  actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
		},
	}
}

// NewDefaultMichigan は標準構成 (人間 seat 0 + CPU) を生成する。CUI / Web / Worker 構築の単一情報源。
func NewDefaultMichigan() *Michigan {
	cfg := DefaultMichiganConfig()
	g := NewMichigan(NewTrumpCards(0), michiganNewPlayers(cfg), cfg)
	g.Reset()
	return g
}

// michiganNewPlayers は設定に基づいてプレイヤー列を生成する (seat 0 = 人間)。
func michiganNewPlayers(cfg MichiganConfig) []*MichiganPlayer {
	players := make([]*MichiganPlayer, cfg.PlayerCount)
	for i := range players {
		players[i] = NewMichiganPlayer(i == 0, cfg.StartingChips)
	}
	return players
}

// --- ゲーム進行 ---

// Reset は新しいゲームを開始する。チップ・プレイヤー数・ブードルを設定から作り直し、第 1 ラウンドを配る。
func (g *Michigan) Reset() {
	g.players = michiganNewPlayers(g.config)
	g.trumpCards = NewTrumpCards(0)
	g.state = michiganState{
		phase:          MichiganPhaseBet,
		roundNumber:    1,
		dealerIdx:      0,
		winnerIdx:      -1,
		matchWinnerIdx: -1,
		boodles:        newMichiganBoodles(),
		actionLogBase:  actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
	}
	g.startRound()
}

// NextRound は同じチップ・持ち越しブードルのまま次のラウンドを配る。Result フェーズかつ
// ゲーム継続中のときのみ有効。
func (g *Michigan) NextRound() {
	if g.state.phase != MichiganPhaseResult || g.state.gameEndFlag {
		return
	}
	g.state.roundNumber++
	g.state.dealerIdx = (g.state.dealerIdx + 1) % len(g.players)
	g.startRound()
}

// startRound は 1 ラウンドを準備する: 状態クリア → CPU の自動ベット → 人間のベット待ち。
func (g *Michigan) startRound() {
	g.state.winnerIdx = -1
	g.state.result = MichiganResultNone
	g.state.seqSuit = 0
	g.state.seqHighValue = 0
	g.state.deadHand = make([]*Card, 0)
	g.state.scored = false
	g.state.humanBetPlaced = false
	for _, b := range g.state.boodles {
		b.claimedBy = -1
	}
	for _, p := range g.players {
		p.ResetForRound()
	}
	// ラウンド開始時 (賭け前) のチップを記録する (結果判定用)。
	g.state.roundStartChips = make([]int, len(g.players))
	for i, p := range g.players {
		g.state.roundStartChips[i] = p.GetChips()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.state.phase = MichiganPhaseBet
	g.state.currentPlayer = 0
	g.appendLog(-1, "round_start", fmt.Sprintf("Round %d: place your boodle bets", g.state.roundNumber), nil)
	// CPU は自動でブードルに賭ける。
	for i := 1; i < len(g.players); i++ {
		g.autoBet(i)
	}
	// 人間が賭けられない (残高 0) 場合は空の賭けを置いてプレイへ進む。
	if g.humanBudget() == 0 {
		g.state.humanBetPlaced = true
		g.finishBetting()
	}
}

// humanBudget は人間 (seat 0) が今ラウンド賭けるべき額 (min(ante, chips)) を返す。
func (g *Michigan) humanBudget() int {
	chips := g.players[0].GetChips()
	if chips < g.config.Ante {
		return chips
	}
	return g.config.Ante
}

// michiganEvenSplit は budget を 4 つのブードルにできるだけ均等に分配する。
func michiganEvenSplit(budget int) []int {
	dist := make([]int, MichiganBoodleCount)
	q := budget / MichiganBoodleCount
	r := budget % MichiganBoodleCount
	for i := 0; i < MichiganBoodleCount; i++ {
		dist[i] = q
		if i < r {
			dist[i]++
		}
	}
	return dist
}

// autoBet は CPU (seat) の自動ベット (均等分配) を行う。
func (g *Michigan) autoBet(seat int) {
	budget := g.playerBudget(seat)
	if budget <= 0 {
		return
	}
	g.applyBet(seat, michiganEvenSplit(budget))
}

// playerBudget は seat が今ラウンド賭けるべき額 (min(ante, chips)) を返す。
func (g *Michigan) playerBudget(seat int) int {
	chips := g.players[seat].GetChips()
	if chips < g.config.Ante {
		return chips
	}
	return g.config.Ante
}

// applyBet は dist に従ってチップをブードルに積む (呼び出し側で合計検証済み)。
func (g *Michigan) applyBet(seat int, dist []int) {
	p := g.players[seat]
	total := 0
	for i, amt := range dist {
		if amt <= 0 {
			continue
		}
		p.SubtractChips(amt)
		g.state.boodles[i].chips += amt
		total += amt
	}
	p.AddRoundBet(total)
	g.appendLog(seat, "bet", fmt.Sprintf("%s bets %d across boodles", playerName(g.players, seat), total), nil)
}

// PlaceHumanBet は人間 (seat 0) のブードル賭けを適用する。bets は 4 要素、各非負、
// 合計は humanBudget() と一致する必要がある。
func (g *Michigan) PlaceHumanBet(bets []int) error {
	if g.state.gameEndFlag {
		return NewDomainError(ErrGameEnded, "the game has already ended")
	}
	if g.state.phase != MichiganPhaseBet {
		return NewDomainError(ErrWrongPhase, "betting is only allowed during the betting phase")
	}
	if g.state.humanBetPlaced {
		return NewDomainError(ErrInvalidPlay, "you have already placed your bet this round")
	}
	if len(bets) != MichiganBoodleCount {
		return NewDomainError(ErrInvalidPlay, "you must place a bet on each of the four boodles")
	}
	sum := 0
	for _, amt := range bets {
		if amt < 0 {
			return NewDomainError(ErrInvalidPlay, "boodle bets must be non-negative")
		}
		sum += amt
	}
	if sum != g.humanBudget() {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("boodle bets must total %d", g.humanBudget()))
	}
	g.applyBet(0, bets)
	g.state.humanBetPlaced = true
	g.finishBetting()
	return nil
}

// finishBetting は全員のベット完了後にカードを配ってプレイフェーズを開始する。
func (g *Michigan) finishBetting() {
	g.deal()
	g.state.phase = MichiganPhasePlay
	g.state.leadPlayerIdx = (g.state.dealerIdx + 1) % len(g.players)
	g.state.currentPlayer = g.state.leadPlayerIdx
	g.state.lastPlayerIdx = g.state.leadPlayerIdx
	g.state.seqSuit = 0
	g.state.seqHighValue = 0
	g.appendLog(-1, "deal", fmt.Sprintf("Round %d: cards dealt, player %d leads", g.state.roundNumber, g.state.leadPlayerIdx), nil)
	g.driveCPU()
}

// deal は 52 枚を N 人 + 1 つのデッドハンドにラウンドロビンで配り切る。
func (g *Michigan) deal() {
	n := len(g.players)
	recipients := n + 1 // 最後の受け手 (index n) はデッドハンド。
	i := 0
	for {
		c := g.trumpCards.DrawCard()
		if c == nil {
			break
		}
		r := i % recipients
		if r < n {
			g.players[r].AddCard(c)
		} else {
			g.state.deadHand = append(g.state.deadHand, c)
		}
		i++
	}
	for _, p := range g.players {
		michiganSortHand(p)
	}
	michiganSortCards(g.state.deadHand)
}

// michiganSortCards はカード列をスート昇順・値昇順に並べ替える。
func michiganSortCards(cards []*Card) {
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return cards[i].GetValue() < cards[j].GetValue()
	})
}

// michiganSortHand はプレイヤーの手札をスート昇順・値昇順に並べ替える。
func michiganSortHand(p *MichiganPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := range cards {
		cards[i] = p.GetCard(i)
	}
	michiganSortCards(cards)
	p.ClearHand()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// --- Play (human) ---

// PlayCard は人間 (現在の手番) が手札インデックス idx のカードを出す。
func (g *Michigan) PlayCard(idx int) error {
	if g.state.gameEndFlag {
		return NewDomainError(ErrGameEnded, "the game has already ended")
	}
	if g.state.phase != MichiganPhasePlay {
		return NewDomainError(ErrWrongPhase, "playing is only allowed during the play phase")
	}
	if !g.players[g.state.currentPlayer].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "it is not your turn")
	}
	if !g.isLegalPlay(g.state.currentPlayer, idx) {
		return NewDomainError(ErrInvalidPlay, "that card cannot be played now")
	}
	g.doPlay(g.state.currentPlayer, idx)
	g.driveCPU()
	return nil
}

// isLegalPlay は seat が手札インデックス idx を今出せるかを判定する。
func (g *Michigan) isLegalPlay(seat, idx int) bool {
	for _, v := range g.playableIndices(seat) {
		if v == idx {
			return true
		}
	}
	return false
}

// doPlay は 1 枚のプレイを処理する: ブードル精算 → シーケンス更新 → ストップ/継続/ラウンド終了。
func (g *Michigan) doPlay(seat, idx int) {
	p := g.players[seat]
	removed := p.RemoveCard(idx)
	if removed == nil {
		return
	}
	// ブードル精算。
	for _, b := range g.state.boodles {
		if b.claimedBy == -1 && b.card.GetDesign() == removed.GetDesign() && b.card.GetValue() == removed.GetValue() {
			won := b.chips
			b.chips = 0
			b.claimedBy = seat
			if won > 0 {
				p.AddChips(won)
			}
			g.appendLog(seat, "boodle", fmt.Sprintf("%s hits a boodle and collects %d", playerName(g.players, seat), won), []*Card{removed})
		}
	}
	// シーケンス更新。
	g.state.seqSuit = removed.GetDesign()
	g.state.seqHighValue = removed.GetValue()
	g.state.lastPlayerIdx = seat
	g.appendLog(seat, "play", fmt.Sprintf("%s plays %s", playerName(g.players, seat), michiganCardStr(removed)), []*Card{removed})
	// 手札を出し切ったらラウンド終了。
	if p.GetCardsSize() == 0 {
		g.enterRoundEnd(seat)
		return
	}
	// 次の札を持つプレイヤーへ手番を移す。
	nextVal := g.state.seqHighValue + 1
	if nextVal <= 13 {
		if holder := g.holderOf(g.state.seqSuit, nextVal); holder >= 0 {
			g.state.currentPlayer = holder
			return
		}
	}
	// ストップ: 最後に出したプレイヤーが新しいシーケンスを始める。
	g.appendLog(-1, "stop", "sequence stopped", nil)
	g.state.seqSuit = 0
	g.state.seqHighValue = 0
	g.state.currentPlayer = g.state.lastPlayerIdx
}

// holderOf は (suit, value) のカードを持つプレイヤー席を返す (誰も持たない = デッドハンド = -1)。
func (g *Michigan) holderOf(suit, value int) int {
	for seat, p := range g.players {
		for i := 0; i < p.GetCardsSize(); i++ {
			c := p.GetCard(i)
			if c.GetDesign() == suit && c.GetValue() == value {
				return seat
			}
		}
	}
	return -1
}

// playableIndices は seat が現在の手番で出せる手札インデックス列を返す (手番でなければ空)。
func (g *Michigan) playableIndices(seat int) []int {
	if g.state.phase != MichiganPhasePlay || seat != g.state.currentPlayer {
		return []int{}
	}
	p := g.players[seat]
	if g.state.seqSuit == 0 {
		// リード: 各スートの最小札。
		res := make([]int, 0, MichiganBoodleCount)
		for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
			lowIdx := -1
			lowVal := CardValueMax + 1
			for i := 0; i < p.GetCardsSize(); i++ {
				c := p.GetCard(i)
				if c.GetDesign() == suit && c.GetValue() < lowVal {
					lowVal = c.GetValue()
					lowIdx = i
				}
			}
			if lowIdx >= 0 {
				res = append(res, lowIdx)
			}
		}
		sort.Ints(res)
		return res
	}
	// 進行中: 昇順で次の 1 枚のみが合法 (手番プレイヤーは必ず所持)。
	want := g.state.seqHighValue + 1
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == g.state.seqSuit && c.GetValue() == want {
			return []int{i}
		}
	}
	return []int{}
}

// matchesUnclaimedBoodle は card が未獲得かつチップの積まれたブードルと一致するかを返す。
func (g *Michigan) matchesUnclaimedBoodle(card *Card) bool {
	for _, b := range g.state.boodles {
		if b.claimedBy == -1 && b.chips > 0 &&
			b.card.GetDesign() == card.GetDesign() && b.card.GetValue() == card.GetValue() {
			return true
		}
	}
	return false
}

// --- CPU ---

// driveCPU は現在の手番が CPU の間、CPU のプレイを実行し続ける。人間の手番か
// ラウンド解決で停止する。
func (g *Michigan) driveCPU() {
	guard := 0
	for g.state.phase == MichiganPhasePlay && !g.state.gameEndFlag {
		seat := g.state.currentPlayer
		if g.players[seat].GetIsHuman() {
			return
		}
		g.cpuPlay(seat)
		guard++
		if guard > michiganMaxPlays {
			return
		}
	}
}

// cpuPlay は CPU (seat) が 1 枚を選んで出す。
func (g *Michigan) cpuPlay(seat int) {
	idx := g.chooseCpuPlay(seat)
	if idx < 0 {
		return
	}
	g.doPlay(seat, idx)
}

// chooseCpuPlay は CPU (seat) の出す手札インデックスを選ぶ (合法手なし = -1)。
func (g *Michigan) chooseCpuPlay(seat int) int {
	pi := g.playableIndices(seat)
	if len(pi) == 0 {
		return -1
	}
	if g.state.seqSuit != 0 {
		return pi[0] // 進行中は強制の 1 枚。
	}
	// リード: チップの積まれたブードルを取れる手を優先、なければ最小値の手。
	best := pi[0]
	bestVal := g.players[seat].GetCard(pi[0]).GetValue()
	for _, idx := range pi {
		card := g.players[seat].GetCard(idx)
		if g.matchesUnclaimedBoodle(card) {
			return idx
		}
		if card.GetValue() < bestVal {
			best = idx
			bestVal = card.GetValue()
		}
	}
	return best
}

// --- Round / game end ---

// enterRoundEnd は手札を出し切ったプレイヤーが出た時点でラウンドを解決する。
// scored フラグで二重解決を防ぐ (フェーズ入場時に 1 回だけ発火)。
func (g *Michigan) enterRoundEnd(goOutSeat int) {
	if g.state.scored {
		return
	}
	g.state.phase = MichiganPhaseResult
	g.state.winnerIdx = goOutSeat
	g.state.result = g.computeHumanResult()
	g.state.scored = true
	g.appendLog(goOutSeat, "round_end", fmt.Sprintf("%s empties their hand; round over", playerName(g.players, goOutSeat)), nil)
	g.checkGameEnd()
}

// computeHumanResult は人間 (seat 0) の今ラウンドの純増チップから勝敗を導く。
func (g *Michigan) computeHumanResult() MichiganResult {
	if len(g.state.roundStartChips) != len(g.players) {
		return MichiganResultNone
	}
	humanNet := g.players[0].GetChips() - g.state.roundStartChips[0]
	for i := 1; i < len(g.players); i++ {
		net := g.players[i].GetChips() - g.state.roundStartChips[i]
		if net > humanNet {
			return MichiganResultLose
		}
	}
	if humanNet > 0 {
		return MichiganResultWin
	}
	return MichiganResultNone
}

// checkGameEnd は停止条件 (規定ラウンド到達) を判定し、満たせばゲームを終了させる。
func (g *Michigan) checkGameEnd() {
	if g.state.roundNumber >= g.config.TargetRounds {
		g.endGame()
	}
}

// endGame はゲームを終了し、チップ最多のプレイヤーを勝者に設定する。
func (g *Michigan) endGame() {
	g.state.gameEndFlag = true
	g.state.phase = MichiganPhaseResult
	g.state.matchWinnerIdx = g.richestIdx()
	g.appendLog(g.state.matchWinnerIdx, "game_end",
		fmt.Sprintf("%s wins the game", playerName(g.players, g.state.matchWinnerIdx)), nil)
}

// richestIdx はチップが最多のプレイヤーのインデックスを返す (同数は座席番号の小さい方)。
func (g *Michigan) richestIdx() int {
	return maxIndexBy(g.players, func(p *MichiganPlayer) int { return p.GetChips() })
}

// --- Hint ---

// GetHint はプレイ中の人間 (seat 0 の手番) にプレイ助言を返す。
func (g *Michigan) GetHint() *MichiganHint {
	if g.state.phase != MichiganPhasePlay || g.state.gameEndFlag || !g.IsHumanTurn() {
		return nil
	}
	pi := g.playableIndices(0)
	if len(pi) == 0 {
		return nil
	}
	if g.state.seqSuit != 0 {
		return &MichiganHint{CardIndex: pi[0], Reason: "forced"}
	}
	for _, idx := range pi {
		if g.matchesUnclaimedBoodle(g.players[0].GetCard(idx)) {
			return &MichiganHint{CardIndex: idx, Reason: "claim_boodle"}
		}
	}
	best := pi[0]
	bestVal := g.players[0].GetCard(pi[0]).GetValue()
	for _, idx := range pi {
		if v := g.players[0].GetCard(idx).GetValue(); v < bestVal {
			best = idx
			bestVal = v
		}
	}
	return &MichiganHint{CardIndex: best, Reason: "lead_low"}
}

// --- ヘルパー ---

// michiganCardStr はカードを棋譜用文字列に変換する。
func michiganCardStr(c *Card) string {
	if c == nil {
		return "??"
	}
	suits := []string{"joker", "spade", "clover", "heart", "diamond"}
	d := c.GetDesign()
	if d < 0 || d >= len(suits) {
		d = 0
	}
	return fmt.Sprintf("%s %d", suits[d], c.GetValue())
}

// michiganValidCard は復元カードが有効なスート (1..4) と値 (1..CardValueMax) を
// 持つかを検証する (JSON 復元時の防御用: 範囲外カードは後段の design 参照で破綻する)。
func michiganValidCard(c *Card) bool {
	if c == nil {
		return false
	}
	d := c.GetDesign()
	v := c.GetValue()
	return d >= CardDesignSpade && d <= CardDesignDiamond && v >= 1 && v <= CardValueMax
}

func (g *Michigan) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.state.appendLog(playerIdx, actionType, detail, cards)
}

// --- 状態アクセサ ---

// GetPhase は現在のフェーズを返す。
func (g *Michigan) GetPhase() MichiganPhase { return g.state.phase }

// SetPhase はフェーズを設定する (テスト用)。
func (g *Michigan) SetPhase(p MichiganPhase) { g.state.phase = p }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *Michigan) GetGameEndFlag() bool { return g.state.gameEndFlag }

// GetRoundNumber は現在のラウンド番号を返す。
func (g *Michigan) GetRoundNumber() int { return g.state.roundNumber }

// GetDealerIdx はディーラーの座席番号を返す。
func (g *Michigan) GetDealerIdx() int { return g.state.dealerIdx }

// SetDealerIdx はディーラーの座席番号を設定する (テスト用)。
func (g *Michigan) SetDealerIdx(idx int) { g.state.dealerIdx = idx }

// GetCurrentPlayerIdx は現在の手番プレイヤーの座席番号を返す。
func (g *Michigan) GetCurrentPlayerIdx() int { return g.state.currentPlayer }

// SetCurrentPlayerIdx は現在の手番プレイヤーを設定する (テスト用)。
func (g *Michigan) SetCurrentPlayerIdx(idx int) { g.state.currentPlayer = idx }

// GetLeadPlayerIdx はこのラウンドの最初のリード席を返す。
func (g *Michigan) GetLeadPlayerIdx() int { return g.state.leadPlayerIdx }

// GetAnte はアンティ額を返す。
func (g *Michigan) GetAnte() int { return g.config.Ante }

// GetBetBudget は人間 (seat 0) が今ラウンド賭けるべき額 (min(ante, chips)) を返す。
func (g *Michigan) GetBetBudget() int { return g.humanBudget() }

// GetHumanBetPlaced は人間が今ラウンドの賭けを済ませたかを返す。
func (g *Michigan) GetHumanBetPlaced() bool { return g.state.humanBetPlaced }

// GetBoodleCnt はブードルの数を返す (常に 4)。
func (g *Michigan) GetBoodleCnt() int { return len(g.state.boodles) }

// GetBoodle は指定インデックスのブードルを返す。
func (g *Michigan) GetBoodle(i int) *MichiganBoodle {
	if i < 0 || i >= len(g.state.boodles) {
		return nil
	}
	return g.state.boodles[i]
}

// GetSeqSuit は現在のシーケンスのスートを返す (0 = 新シーケンス待ち, 1..4 = 進行中)。
func (g *Michigan) GetSeqSuit() int { return g.state.seqSuit }

// GetSeqHighValue は現在のシーケンスで出た最大値を返す (0 = なし)。
func (g *Michigan) GetSeqHighValue() int { return g.state.seqHighValue }

// GetDeadHandCount はデッドハンドの枚数を返す。
func (g *Michigan) GetDeadHandCount() int { return len(g.state.deadHand) }

// GetPlayerCnt はプレイヤー数を返す。
func (g *Michigan) GetPlayerCnt() int { return len(g.players) }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *Michigan) GetPlayer(i int) *MichiganPlayer {
	return getPlayer(g.players, i)
}

// GetChips は人間 (seat 0) の保有チップを返す。
func (g *Michigan) GetChips() int {
	return chipsOfFirst(g.players)
}

// IsHumanTurn は現在人間 (seat 0) の入力待ちかどうかを返す。
func (g *Michigan) IsHumanTurn() bool {
	if g.state.gameEndFlag {
		return false
	}
	switch g.state.phase {
	case MichiganPhaseBet:
		return !g.state.humanBetPlaced
	case MichiganPhasePlay:
		idx := g.state.currentPlayer
		if idx < 0 || idx >= len(g.players) {
			return false
		}
		return g.players[idx].GetIsHuman()
	default:
		return false
	}
}

// GetPlayableIndices は人間 (seat 0) が今出せる手札インデックス列を返す (手番でなければ空)。
func (g *Michigan) GetPlayableIndices() []int {
	if g.state.phase != MichiganPhasePlay || g.state.currentPlayer != 0 || !g.IsHumanTurn() {
		return []int{}
	}
	return g.playableIndices(0)
}

// GetWinnerIdx は直近ラウンドで手札を出し切ったプレイヤーを返す (-1 = なし)。
func (g *Michigan) GetWinnerIdx() int { return g.state.winnerIdx }

// GetMatchWinnerIdx はゲーム全体の勝者を返す (-1 = 未確定)。
func (g *Michigan) GetMatchWinnerIdx() int { return g.state.matchWinnerIdx }

// GetResult は人間から見たラウンド結果を返す。
func (g *Michigan) GetResult() MichiganResult { return g.state.result }

// GetConfig はローカルルール設定を返す。
func (g *Michigan) GetConfig() MichiganConfig { return g.config }

// SetConfig はローカルルール設定を変更する。
func (g *Michigan) SetConfig(cfg MichiganConfig) { g.config = cfg }

// GetActionLog は棋譜を返す。
func (g *Michigan) GetActionLog() []*ActionLogEntry { return g.state.actionLog }

// --- テスト用ヘルパー ---

// SetPlayStateForTest はプレイ状態 (シーケンス・手番) を直接設定する (テスト用)。
// 直前のラウンド解決状態 (scored / winnerIdx / result) もクリアし、決定的なプレイ
// シナリオを隔離して組み立てられるようにする。
func (g *Michigan) SetPlayStateForTest(seqSuit, seqHigh, current, last int) {
	g.state.phase = MichiganPhasePlay
	g.state.seqSuit = seqSuit
	g.state.seqHighValue = seqHigh
	g.state.currentPlayer = current
	g.state.lastPlayerIdx = last
	g.state.scored = false
	g.state.winnerIdx = -1
	g.state.result = MichiganResultNone
}

// AddDeadCardForTest はデッドハンドにカードを追加する (テスト用)。
func (g *Michigan) AddDeadCardForTest(c *Card) {
	g.state.deadHand = append(g.state.deadHand, c)
}

// SetBoodleForTest はブードルのチップ・獲得状態を設定する (テスト用)。
func (g *Michigan) SetBoodleForTest(i, chips, claimedBy int) {
	if i < 0 || i >= len(g.state.boodles) {
		return
	}
	g.state.boodles[i].chips = chips
	g.state.boodles[i].claimedBy = claimedBy
}

// SetRoundStartChipsForTest はラウンド開始チップを設定する (テスト用)。
func (g *Michigan) SetRoundStartChipsForTest(chips []int) {
	g.state.roundStartChips = chips
}

// DoPlayForTest は seat の手札インデックス idx を出す (乱数配札を迂回した決定的検証用)。
func (g *Michigan) DoPlayForTest(seat, idx int) { g.doPlay(seat, idx) }

// DriveCPUForTest は CPU 手番を進める (テスト用)。
func (g *Michigan) DriveCPUForTest() { g.driveCPU() }

// --- JSON Serialization ---

// michiganJSON is the JSON wire format for Michigan.
type michiganJSON struct {
	TrumpCards      *TrumpCards       `json:"tc"`
	Players         []*MichiganPlayer `json:"ps"`
	Config          MichiganConfig    `json:"cf"`
	Phase           MichiganPhase     `json:"ph"`
	RoundNumber     int               `json:"rn"`
	DealerIdx       int               `json:"di"`
	CurrentPlayer   int               `json:"ci"`
	LeadPlayerIdx   int               `json:"li"`
	SeqSuit         int               `json:"sq"`
	SeqHighValue    int               `json:"sh"`
	LastPlayerIdx   int               `json:"lp"`
	Boodles         []*MichiganBoodle `json:"bd"`
	DeadHand        []*Card           `json:"dh"`
	RoundStartChips []int             `json:"rs"`
	HumanBetPlaced  bool              `json:"hb"`
	WinnerIdx       int               `json:"wi"`
	MatchWinnerIdx  int               `json:"mw"`
	Result          MichiganResult    `json:"re"`
	GameEndFlag     bool              `json:"ge"`
	Scored          bool              `json:"sc"`
	ActionLog       []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Michigan) MarshalJSON() ([]byte, error) {
	return json.Marshal(michiganJSON{
		TrumpCards:      g.trumpCards,
		Players:         g.players,
		Config:          g.config,
		Phase:           g.state.phase,
		RoundNumber:     g.state.roundNumber,
		DealerIdx:       g.state.dealerIdx,
		CurrentPlayer:   g.state.currentPlayer,
		LeadPlayerIdx:   g.state.leadPlayerIdx,
		SeqSuit:         g.state.seqSuit,
		SeqHighValue:    g.state.seqHighValue,
		LastPlayerIdx:   g.state.lastPlayerIdx,
		Boodles:         g.state.boodles,
		DeadHand:        g.state.deadHand,
		RoundStartChips: g.state.roundStartChips,
		HumanBetPlaced:  g.state.humanBetPlaced,
		WinnerIdx:       g.state.winnerIdx,
		MatchWinnerIdx:  g.state.matchWinnerIdx,
		Result:          g.state.result,
		GameEndFlag:     g.state.gameEndFlag,
		Scored:          g.state.scored,
		ActionLog:       g.state.actionLog,
	})
}

// michiganValidPhase は有効なフェーズかどうか。
func michiganValidPhase(p MichiganPhase) bool {
	return p == MichiganPhaseBet || p == MichiganPhasePlay || p == MichiganPhaseResult
}

// UnmarshalJSON implements json.Unmarshaler. 不正な永続化データを拒否するための
// バリデーションを行う (KV 復元時のインデックス範囲・スライス健全性)。
func (g *Michigan) UnmarshalJSON(data []byte) error {
	var j michiganJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("michigan: invalid config: %w", err)
	}
	n := len(j.Players)
	if n < MichiganMinPlayerCount || n > MichiganMaxPlayerCount || n != j.Config.PlayerCount {
		return errMichiganSnapshot
	}
	if len(j.ActionLog) > michiganMaxSliceLen || len(j.DeadHand) > michiganMaxSliceLen {
		return errMichiganSnapshot
	}
	if !michiganValidPhase(j.Phase) {
		return errMichiganSnapshot
	}
	if j.RoundNumber < 1 {
		return errMichiganSnapshot
	}
	if j.DealerIdx < 0 || j.DealerIdx >= n || j.CurrentPlayer < 0 || j.CurrentPlayer >= n ||
		j.LeadPlayerIdx < 0 || j.LeadPlayerIdx >= n || j.LastPlayerIdx < 0 || j.LastPlayerIdx >= n {
		return errMichiganSnapshot
	}
	if j.SeqSuit < 0 || j.SeqSuit > CardDesignDiamond || j.SeqHighValue < 0 || j.SeqHighValue > CardValueMax {
		return errMichiganSnapshot
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= n || j.MatchWinnerIdx < -1 || j.MatchWinnerIdx >= n {
		return errMichiganSnapshot
	}
	if j.Result < MichiganResultLose || j.Result > MichiganResultWin {
		return errMichiganSnapshot
	}
	// ブードル: 省略時は空 4 枚、指定時は必ず 4 枚。
	boodles := j.Boodles
	if len(boodles) == 0 {
		boodles = newMichiganBoodles()
	} else if len(boodles) != MichiganBoodleCount {
		return errMichiganSnapshot
	}
	for _, b := range boodles {
		if b == nil || !michiganValidCard(b.card) || b.chips < 0 || b.claimedBy < -1 || b.claimedBy >= n {
			return errMichiganSnapshot
		}
	}
	for _, p := range j.Players {
		if p == nil {
			return errMichiganSnapshot
		}
	}
	for _, c := range j.DeadHand {
		if !michiganValidCard(c) {
			return errMichiganSnapshot
		}
	}
	for _, e := range j.ActionLog {
		if e == nil {
			return errMichiganSnapshot
		}
	}
	// roundStartChips: 省略時はゼロ、指定時は必ず n 要素。
	rs := j.RoundStartChips
	if len(rs) == 0 {
		rs = make([]int, n)
	} else if len(rs) != n {
		return errMichiganSnapshot
	}
	for _, c := range rs {
		if c < 0 {
			return errMichiganSnapshot
		}
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	g.config = j.Config
	deadHand := j.DeadHand
	if deadHand == nil {
		deadHand = make([]*Card, 0)
	}
	actionLog := j.ActionLog
	if actionLog == nil {
		actionLog = make([]*ActionLogEntry, 0)
	}
	g.state = michiganState{
		phase:           j.Phase,
		roundNumber:     j.RoundNumber,
		dealerIdx:       j.DealerIdx,
		currentPlayer:   j.CurrentPlayer,
		leadPlayerIdx:   j.LeadPlayerIdx,
		seqSuit:         j.SeqSuit,
		seqHighValue:    j.SeqHighValue,
		lastPlayerIdx:   j.LastPlayerIdx,
		boodles:         boodles,
		deadHand:        deadHand,
		roundStartChips: rs,
		humanBetPlaced:  j.HumanBetPlaced,
		winnerIdx:       j.WinnerIdx,
		matchWinnerIdx:  j.MatchWinnerIdx,
		result:          j.Result,
		gameEndFlag:     j.GameEndFlag,
		scored:          j.Scored,
		actionLogBase:   actionLogBase{actionLog: actionLog},
	}
	return nil
}
