//go:build !js || !wasm || extra4

package domain

import (
	"errors"
	"fmt"
)

// Russian Bank (Crapette) 盤面定数。
const (
	// RussianBankPlayerCnt プレイヤー数 (人間 1 + CPU 1 固定)。
	RussianBankPlayerCnt = 2
	// RussianBankTableauCnt 共有タブローの列数。
	RussianBankTableauCnt = 4
	// RussianBankFoundationCnt 共有ファウンデーションの本数 (2 デッキ = スート毎 2 本)。
	RussianBankFoundationCnt = 8
	// russianBankMaxSliceLen JSON 復元時のスライス長上限 (改竄ガード)。
	russianBankMaxSliceLen = 10000
)

// RussianBankPhase ゲームフェーズ。
type RussianBankPhase int

// Russian Bank のフェーズ定数。
const (
	// RussianBankPhaseIdle 未開始 (Reset 前)。
	RussianBankPhaseIdle RussianBankPhase = iota
	// RussianBankPhasePlaying プレイ中。
	RussianBankPhasePlaying
	// RussianBankPhaseGameEnd 決着 (いずれかがリザーブを空にした)。
	RussianBankPhaseGameEnd
)

// RussianBankZone カード移動元ゾーン種別。
type RussianBankZone int

// Russian Bank のゾーン定数。
const (
	// RussianBankZoneReserve リザーブ (ストック)。
	RussianBankZoneReserve RussianBankZone = iota
	// RussianBankZoneWaste 廃札。
	RussianBankZoneWaste
	// RussianBankZoneTableau 共有タブロー。
	RussianBankZoneTableau
)

// RussianBankSource 移動元の指定。現在の手番プレイヤーを起点とする。
//   - Reserve / Waste: FromOpponent=true で相手の山を指す。
//   - Tableau: 共有タブローの Col 列。
type RussianBankSource struct {
	Zone         RussianBankZone
	FromOpponent bool
	Col          int
}

var errRussianBank = errors.New("russianbank: illegal move")

// RussianBank Russian Bank (Crapette) ゲーム本体。状態のみを保持する。
type RussianBank struct {
	decks       []*TrumpCards
	players     []*RussianBankPlayer
	tableau     [RussianBankTableauCnt][]*Card
	foundations [RussianBankFoundationCnt][]*Card
	config      RussianBankConfig
	phase       RussianBankPhase
	current     int // 手番プレイヤー seat
	winner      int // -1 = 未確定
	moveCount   int
	passStreak  int                       // 連続スタック・パス数 (両者詰みの停滞検出用)
	stopPoints  [RussianBankPlayerCnt]int // stop で咎めた回数 (副次スコア)
	actionLog   []*ActionLogEntry
	history     []*russianBankSnapshot // 人間の単一ステップ Undo 用
}

// russianBankSnapshot Undo 用スナップショット。
type russianBankSnapshot struct {
	stateJSON []byte
}

// NewRussianBank 設定を指定して生成する。
func NewRussianBank(cfg RussianBankConfig) *RussianBank {
	return &RussianBank{config: cfg, phase: RussianBankPhaseIdle, winner: -1}
}

// NewDefaultRussianBank 既定設定の RussianBank を返す。
func NewDefaultRussianBank() *RussianBank {
	return NewRussianBank(DefaultRussianBankConfig())
}

// Reset 現在の設定で新しいゲームを開始する。
func (g *RussianBank) Reset() {
	if err := g.config.Validate(); err != nil {
		g.config = DefaultRussianBankConfig()
	}
	g.players = make([]*RussianBankPlayer, RussianBankPlayerCnt)
	g.decks = make([]*TrumpCards, RussianBankPlayerCnt)
	for i := 0; i < RussianBankPlayerCnt; i++ {
		g.players[i] = NewRussianBankPlayer(g.defaultPlayerName(i), i != 0, i)
		g.decks[i] = NewTrumpCards(0)
	}
	g.startGame()
}

// ResetWithConfig 設定を適用して Reset する。
func (g *RussianBank) ResetWithConfig(cfg RussianBankConfig) {
	if err := cfg.Validate(); err != nil {
		cfg = DefaultRussianBankConfig()
	}
	g.config = cfg
	g.Reset()
}

// startGame 盤面を初期化して配る。
func (g *RussianBank) startGame() {
	g.phase = RussianBankPhasePlaying
	g.current = 0
	g.winner = -1
	g.moveCount = 0
	g.passStreak = 0
	g.stopPoints = [RussianBankPlayerCnt]int{}
	g.actionLog = nil
	g.history = nil
	for i := range g.tableau {
		g.tableau[i] = nil
	}
	for i := range g.foundations {
		g.foundations[i] = nil
	}
	for i, p := range g.players {
		p.resetPiles()
		deck := g.decks[i]
		deck.Shuffle()
		for k := 0; k < RussianBankReserveSize; k++ {
			p.pushReserve(deck.DrawCard())
		}
		for deck.GetRemainingCount() > 0 {
			p.pushHand(deck.DrawCard())
		}
	}
	g.appendLog(-1, "deal", "新しいゲームを開始しました", nil)
}

func (g *RussianBank) defaultPlayerName(idx int) string {
	if idx == 0 {
		return "You"
	}
	return "CPU"
}

// --- accessors ---

// GetPhase 現在のフェーズ。
func (g *RussianBank) GetPhase() RussianBankPhase { return g.phase }

// GetCurrentPlayer 手番プレイヤー seat。
func (g *RussianBank) GetCurrentPlayer() int { return g.current }

// GetWinner 勝者 seat (未確定は -1)。
func (g *RussianBank) GetWinner() int { return g.winner }

// GetMoveCount 累計手数。
func (g *RussianBank) GetMoveCount() int { return g.moveCount }

// GetConfig 設定を返す。
func (g *RussianBank) GetConfig() RussianBankConfig { return g.config }

// SetConfig 設定を差し替える (Reset は行わない)。
func (g *RussianBank) SetConfig(cfg RussianBankConfig) { g.config = cfg }

// GetGameEndFlag 決着済みなら true。
func (g *RussianBank) GetGameEndFlag() bool { return g.phase == RussianBankPhaseGameEnd }

// IsHumanTurn 現在の手番が人間 (seat 0) かを返す。
func (g *RussianBank) IsHumanTurn() bool { return g.current == 0 }

// CanCallStop 人間が今 stop を宣言できるか (CPU が強制ファウンデーション手を残している)。
func (g *RussianBank) CanCallStop() bool {
	return g.phase == RussianBankPhasePlaying && g.current == 0 && g.hasForcedFoundationMove(g.opponent(0))
}

// GetPlayers プレイヤー列を返す。
func (g *RussianBank) GetPlayers() []*RussianBankPlayer { return g.players }

// GetPlayer seat のプレイヤーを返す (範囲外は nil)。
func (g *RussianBank) GetPlayer(seat int) *RussianBankPlayer {
	return getPlayer(g.players, seat)
}

// GetTableau 共有タブローを返す。
func (g *RussianBank) GetTableau() [RussianBankTableauCnt][]*Card { return g.tableau }

// GetFoundations 共有ファウンデーションを返す。
func (g *RussianBank) GetFoundations() [RussianBankFoundationCnt][]*Card { return g.foundations }

// GetStopPoints seat の stop 得点を返す。
func (g *RussianBank) GetStopPoints(seat int) int {
	if seat < 0 || seat >= RussianBankPlayerCnt {
		return 0
	}
	return g.stopPoints[seat]
}

// GetActionLog アクションログを返す。
func (g *RussianBank) GetActionLog() []*ActionLogEntry { return g.actionLog }

func (g *RussianBank) appendLog(seat int, action, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.moveCount,
		PlayerIdx:  seat,
		ActionType: action,
		Detail:     detail,
		Cards:      cards,
	})
}

func (g *RussianBank) opponent(seat int) int { return (seat + 1) % RussianBankPlayerCnt }

// --- rules ---

// rbIsBlack 黒スート (スペード・クラブ) なら true。
func rbIsBlack(c *Card) bool {
	return c.GetDesign() == CardDesignSpade || c.GetDesign() == CardDesignClover
}

// rbCanPlaceTableau タブロー col に card を置けるか。空列は任意のカード、
// それ以外は交互色かつ降順 (value = top-1)。
func (g *RussianBank) rbCanPlaceTableau(card *Card, col int) bool {
	if card == nil || col < 0 || col >= RussianBankTableauCnt {
		return false
	}
	pile := g.tableau[col]
	if len(pile) == 0 {
		return true
	}
	top := pile[len(pile)-1]
	return rbIsBlack(card) != rbIsBlack(top) && card.GetValue() == top.GetValue()-1
}

// rbCanPlaceFoundation ファウンデーション fIdx に card を置けるか。空本は A のみ、
// それ以外は同スートかつ昇順 (value = top+1)。
func (g *RussianBank) rbCanPlaceFoundation(card *Card, fIdx int) bool {
	if card == nil || fIdx < 0 || fIdx >= RussianBankFoundationCnt {
		return false
	}
	pile := g.foundations[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == 1
	}
	top := pile[len(pile)-1]
	return card.GetDesign() == top.GetDesign() && card.GetValue() == top.GetValue()+1
}

// rbFoundationFor card を置ける最初のファウンデーション番号を返す (なければ -1)。
func (g *RussianBank) rbFoundationFor(card *Card) int {
	for i := 0; i < RussianBankFoundationCnt; i++ {
		if g.rbCanPlaceFoundation(card, i) {
			return i
		}
	}
	return -1
}

// --- source resolution ---

// peekSource 移動元のトップカードを覗く (取り出さない)。
func (g *RussianBank) peekSource(src RussianBankSource) *Card {
	switch src.Zone {
	case RussianBankZoneReserve:
		return g.sourcePlayer(src).ReserveTop()
	case RussianBankZoneWaste:
		return g.sourcePlayer(src).WasteTop()
	case RussianBankZoneTableau:
		if src.Col < 0 || src.Col >= RussianBankTableauCnt {
			return nil
		}
		return rbTopCard(g.tableau[src.Col])
	default:
		return nil
	}
}

// sourcePlayer 移動元の所有プレイヤーを返す。
func (g *RussianBank) sourcePlayer(src RussianBankSource) *RussianBankPlayer {
	seat := g.current
	if src.FromOpponent {
		seat = g.opponent(g.current)
	}
	return g.players[seat]
}

// takeSource 移動元のトップカードを取り出す。
func (g *RussianBank) takeSource(src RussianBankSource) *Card {
	switch src.Zone {
	case RussianBankZoneReserve:
		return g.sourcePlayer(src).popReserve()
	case RussianBankZoneWaste:
		return g.sourcePlayer(src).popWaste()
	case RussianBankZoneTableau:
		if src.Col < 0 || src.Col >= RussianBankTableauCnt {
			return nil
		}
		c := rbTopCard(g.tableau[src.Col])
		if c != nil {
			g.tableau[src.Col][len(g.tableau[src.Col])-1] = nil
			g.tableau[src.Col] = g.tableau[src.Col][:len(g.tableau[src.Col])-1]
		}
		return c
	default:
		return nil
	}
}

func rbSourceName(src RussianBankSource) string {
	owner := "own"
	if src.FromOpponent {
		owner = "opp"
	}
	switch src.Zone {
	case RussianBankZoneReserve:
		return owner + " reserve"
	case RussianBankZoneWaste:
		return owner + " waste"
	case RussianBankZoneTableau:
		return fmt.Sprintf("tableau %d", src.Col)
	default:
		return "?"
	}
}

// --- public actions ---

// MoveToFoundation 移動元のトップを置ける任意のファウンデーションに移す。
func (g *RussianBank) MoveToFoundation(src RussianBankSource) error {
	if err := g.assertPlayable(); err != nil {
		return err
	}
	card := g.peekSource(src)
	if card == nil {
		return errRussianBank
	}
	fIdx := g.rbFoundationFor(card)
	if fIdx < 0 {
		return errRussianBank
	}
	g.takeSnapshot()
	g.takeSource(src)
	g.foundations[fIdx] = append(g.foundations[fIdx], card)
	g.moveCount++
	g.appendLog(g.current, "foundation", fmt.Sprintf("%s → foundation %d", rbSourceName(src), fIdx), nil)
	g.afterMove()
	return nil
}

// MoveToTableau 移動元のトップを共有タブロー col に移す。
func (g *RussianBank) MoveToTableau(src RussianBankSource, col int) error {
	if err := g.assertPlayable(); err != nil {
		return err
	}
	card := g.peekSource(src)
	if card == nil {
		return errRussianBank
	}
	// 同じタブロー列への移動は無効。
	if src.Zone == RussianBankZoneTableau && src.Col == col {
		return errRussianBank
	}
	if !g.rbCanPlaceTableau(card, col) {
		return errRussianBank
	}
	g.takeSnapshot()
	g.takeSource(src)
	g.tableau[col] = append(g.tableau[col], card)
	g.moveCount++
	g.appendLog(g.current, "tableau", fmt.Sprintf("%s → tableau %d", rbSourceName(src), col), nil)
	g.afterMove()
	return nil
}

// Discard 手札トップを自分の廃札に送り、手番を終える。手札が空なら手番のみ渡す。
func (g *RussianBank) Discard() error {
	if err := g.assertPlayable(); err != nil {
		return err
	}
	g.takeSnapshot()
	p := g.players[g.current]
	if c := p.popHand(); c != nil {
		p.pushWaste(c)
		g.moveCount++
		g.passStreak = 0
		g.appendLog(g.current, "discard", "手札を廃札に送り手番終了", nil)
	} else {
		g.passStreak++
		g.appendLog(g.current, "pass", "手番をパス", nil)
		if g.checkStalemate() {
			return nil
		}
	}
	g.endTurn()
	return nil
}

// checkStalemate 両者が連続でスタック・パスした場合に停滞として決着させる。
// 残リザーブが少ない側を勝者とし、同数なら引き分け (winner=-1)。
func (g *RussianBank) checkStalemate() bool {
	if g.passStreak < RussianBankPlayerCnt {
		return false
	}
	g.phase = RussianBankPhaseGameEnd
	r0, r1 := g.players[0].ReserveSize(), g.players[1].ReserveSize()
	switch {
	case r0 < r1:
		g.winner = 0
	case r1 < r0:
		g.winner = 1
	default:
		g.winner = -1
	}
	g.appendLog(-1, "stalemate", "両者とも手詰まりのため停滞で決着", nil)
	return true
}

// CallStop 人間が CPU の取りこぼし (強制ファウンデーション手の見逃し) を咎める。
// 直前に CPU が手番を終え、かつ CPU のリザーブ/廃札トップにファウンデーションへ
// 出せるカードが残っている場合のみ有効。成功で stop 得点 +1 し手番を人間に保つ。
func (g *RussianBank) CallStop() error {
	if err := g.assertPlayable(); err != nil {
		return err
	}
	if g.current != 0 {
		return errors.New("russianbank: stop is only callable on your own turn after the opponent moved")
	}
	opp := g.opponent(g.current)
	if !g.hasForcedFoundationMove(opp) {
		return errors.New("russianbank: no violation to call")
	}
	g.stopPoints[g.current]++
	g.appendLog(g.current, "stop", "CPU の取りこぼしを咎めました (+1)", nil)
	g.history = nil // stop はアンドゥ対象外
	return nil
}

// assertPlayable プレイ可能状態か検証する。
func (g *RussianBank) assertPlayable() error {
	if g.phase != RussianBankPhasePlaying {
		return errors.New("russianbank: game is not in playing phase")
	}
	return nil
}

// afterMove 移動後の勝利判定。盤面手は進展なので停滞カウントをリセットする。
func (g *RussianBank) afterMove() {
	g.passStreak = 0
	if g.players[g.current].ReserveSize() == 0 {
		g.winner = g.current
		g.phase = RussianBankPhaseGameEnd
		g.appendLog(g.current, "win", fmt.Sprintf("%s がリザーブを空にして勝利", g.players[g.current].GetName()), nil)
	}
}

// endTurn 手番を相手に渡す。
func (g *RussianBank) endTurn() {
	if g.phase != RussianBankPhasePlaying {
		return
	}
	g.current = g.opponent(g.current)
	g.history = nil
}

// --- legal move detection ---

// hasForcedFoundationMove seat のリザーブ/廃札トップにファウンデーションへ
// 出せるカードがあるか。
func (g *RussianBank) hasForcedFoundationMove(seat int) bool {
	p := g.players[seat]
	if c := p.ReserveTop(); c != nil && g.rbFoundationFor(c) >= 0 {
		return true
	}
	if c := p.WasteTop(); c != nil && g.rbFoundationFor(c) >= 0 {
		return true
	}
	return false
}

// rbMove 列挙された 1 手。
type rbMove struct {
	src    RussianBankSource
	toFnd  bool // true=ファウンデーション, false=タブロー
	toCol  int  // タブロー列 (toFnd=false 時)
	toFidx int  // ファウンデーション番号 (toFnd=true 時)
}

// enumerateMoves 現在の手番で可能な全盤面手を列挙する。
func (g *RussianBank) enumerateMoves() []rbMove {
	var moves []rbMove
	sources := []RussianBankSource{
		{Zone: RussianBankZoneReserve},
		{Zone: RussianBankZoneWaste},
		{Zone: RussianBankZoneReserve, FromOpponent: true},
		{Zone: RussianBankZoneWaste, FromOpponent: true},
	}
	for c := 0; c < RussianBankTableauCnt; c++ {
		sources = append(sources, RussianBankSource{Zone: RussianBankZoneTableau, Col: c})
	}
	for _, src := range sources {
		card := g.peekSource(src)
		if card == nil {
			continue
		}
		if fIdx := g.rbFoundationFor(card); fIdx >= 0 {
			moves = append(moves, rbMove{src: src, toFnd: true, toFidx: fIdx})
		}
		for col := 0; col < RussianBankTableauCnt; col++ {
			if src.Zone == RussianBankZoneTableau && src.Col == col {
				continue
			}
			if g.rbCanPlaceTableau(card, col) {
				moves = append(moves, rbMove{src: src, toCol: col})
			}
		}
	}
	return moves
}

// RussianBankHint 推奨手 1 つを表す。Zone/FromOpponent/Col が移動元、
// ToFoundation=true ならファウンデーションへ、false なら ToCol のタブローへ。
type RussianBankHint struct {
	Zone         RussianBankZone
	FromOpponent bool
	Col          int
	ToFoundation bool
	ToCol        int
}

// GetHint 人間の手番で推奨手を 1 つ返す (なければ nil)。
// 強制ファウンデーション手を最優先し、次にリザーブを減らすタブロー手を選ぶ。
func (g *RussianBank) GetHint() *RussianBankHint {
	if g.phase != RussianBankPhasePlaying || g.current != 0 {
		return nil
	}
	moves := g.enumerateMoves()
	best := -1
	bestScore := -1
	for i, m := range moves {
		if s := g.scoreMove(m); s > bestScore {
			bestScore = s
			best = i
		}
	}
	if best < 0 {
		return nil
	}
	m := moves[best]
	return &RussianBankHint{
		Zone:         m.src.Zone,
		FromOpponent: m.src.FromOpponent,
		Col:          m.src.Col,
		ToFoundation: m.toFnd,
		ToCol:        m.toCol,
	}
}

// scoreMove 手の優先度を採点する。自リザーブ→ファウンデーションが最高。
func (g *RussianBank) scoreMove(m rbMove) int {
	score := 0
	if m.toFnd {
		score += 4
	}
	switch {
	case m.src.Zone == RussianBankZoneReserve && !m.src.FromOpponent:
		score += 3 // 自リザーブを減らすのが勝利目標
	case m.src.Zone == RussianBankZoneWaste && !m.src.FromOpponent:
		score += 1
	case m.src.FromOpponent:
		score -= 2 // 相手のリザーブ/廃札を減らす手は基本不利
	}
	return score
}

// 状態のシリアライズ・Undo・CPU 手番は RussianBankState.go を参照。
