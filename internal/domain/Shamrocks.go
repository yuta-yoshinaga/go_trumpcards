//go:build !js || !wasm || classic

package domain

import (
	"errors"
	"fmt"
)

// Shamrocks (シャムロックス) 盤面定数。
const (
	// ShamrocksFoundationCnt ファウンデーションの数 (4 スート)。
	ShamrocksFoundationCnt = 4
	// ShamrocksFanSize 1 つの扇 (fan) の初期枚数。
	ShamrocksFanSize = 3
	// shamrocksMaxSliceLen JSON 復元時のスライス長上限。
	shamrocksMaxSliceLen = 10000
)

// ShamrocksPhase ゲームフェーズ。
type ShamrocksPhase int

// Shamrocks のフェーズ定数。
const (
	// ShamrocksPhasePlaying プレイ中。
	ShamrocksPhasePlaying ShamrocksPhase = iota
	// ShamrocksPhaseGameClear クリア (52 枚すべてファウンデーション)。
	ShamrocksPhaseGameClear
	// ShamrocksPhaseGameOver 失敗 (再シャッフル使い切り後に手詰まり / ギブアップ)。
	ShamrocksPhaseGameOver
)

// ShamrocksHint 推奨手。ToFoundation=true ならファウンデーションへ、
// false なら ToFan の扇へ移す。
type ShamrocksHint struct {
	FromFan      int
	ToFan        int
	ToFoundation bool
}

// Shamrocks Shamrocks (ファン・ソリティア) 本体。状態のみを保持する。
type Shamrocks struct {
	trumpCards *TrumpCards
	fans       [][]*Card
	foundation [ShamrocksFoundationCnt][]*Card
	phase      ShamrocksPhase
	moveCount  int
	actionLogBase
	history []*shamrocksSnapshot
}

// shamrocksSnapshot Undo 用スナップショット。
type shamrocksSnapshot struct {
	fans       [][]*Card
	foundation [ShamrocksFoundationCnt][]*Card
	phase      ShamrocksPhase
	moveCount  int
}

// NewShamrocks コンストラクタ。
func NewShamrocks(trumpCards *TrumpCards) *Shamrocks {
	return &Shamrocks{trumpCards: trumpCards}
}

// NewDefaultShamrocks 標準 52 枚デッキで生成する。
func NewDefaultShamrocks() *Shamrocks {
	return NewShamrocks(NewTrumpCards(0))
}

// Reset ゲームを初期化する。
func (g *Shamrocks) Reset() {
	g.trumpCards.Shuffle()
	g.phase = ShamrocksPhasePlaying
	g.moveCount = 0
	g.actionLog = nil
	g.history = nil
	for i := range g.foundation {
		g.foundation[i] = nil
	}
	cards := make([]*Card, 0, 52)
	for g.trumpCards.GetRemainingCount() > 0 {
		cards = append(cards, g.trumpCards.DrawCard())
	}
	g.dealFans(cards)
	g.appendLog("deal", "新しいゲームを開始しました", nil)
}

// dealFans カード列を 3 枚ずつの扇に配る (末尾の扇は 1〜3 枚)。
func (g *Shamrocks) dealFans(cards []*Card) {
	g.fans = nil
	for i := 0; i < len(cards); i += ShamrocksFanSize {
		end := i + ShamrocksFanSize
		if end > len(cards) {
			end = len(cards)
		}
		fan := make([]*Card, end-i)
		copy(fan, cards[i:end])
		g.fans = append(g.fans, fan)
	}
}

// --- rules ---

// shamrocksCanStack 扇トップ dstTop に card を積めるか (スート不問・階級±1)。
func shamrocksCanStack(card, dstTop *Card) bool {
	if card == nil || dstTop == nil {
		return false
	}
	// **スートは見ない。1つ上でも1つ下でもよい。**
	// La Belle Lucie は「同スート・降順のみ」なので、両方が違う。
	diff := card.GetValue() - dstTop.GetValue()
	return diff == 1 || diff == -1
}

// findFoundation card を置けるファウンデーション番号を返す (なければ -1)。
func (g *Shamrocks) findFoundation(card *Card) int {
	for i := 0; i < ShamrocksFoundationCnt; i++ {
		pile := g.foundation[i]
		if len(pile) == 0 {
			if card.GetValue() == 1 {
				return i
			}
			continue
		}
		top := pile[len(pile)-1]
		if card.GetDesign() == top.GetDesign() && card.GetValue() == top.GetValue()+1 {
			return i
		}
	}
	return -1
}

// fanTop 扇 idx のトップカードを返す (空・範囲外は nil)。
func (g *Shamrocks) fanTop(idx int) *Card {
	if idx < 0 || idx >= len(g.fans) || len(g.fans[idx]) == 0 {
		return nil
	}
	return g.fans[idx][len(g.fans[idx])-1]
}

// --- public actions ---

// canMoveFanToFan は扇 from のトップを扇 to へ動かせるかを一箇所で判定する。
//
// **MoveFanToFan / hasAnyLegalMove / GetHint の3つが同じ規則を見る必要がある。**
// クローン元 (La Belle Lucie) は探索側が shamrocksCanStack だけを見ており、
// Shamrocks で足した2つの条件を知らない。そのままだと
// 空の扇への手を「無い」と数え（再配りが無いので即ゲームオーバーになる）、
// 満杯の扇への手を「ある」と数える（ヒントが打てない手を指す）。
func (g *Shamrocks) canMoveFanToFan(from, to int) bool {
	if from == to || from < 0 || to < 0 || from >= len(g.fans) || to >= len(g.fans) {
		return false
	}
	card := g.fanTop(from)
	if card == nil {
		return false
	}
	if len(g.fans[to]) >= ShamrocksFanSize {
		return false
	}
	if len(g.fans[to]) == 0 {
		return true
	}
	return shamrocksCanStack(card, g.fanTop(to))
}

// MoveFanToFan 扇 from のトップを扇 to に積む。
func (g *Shamrocks) MoveFanToFan(from, to int) error {
	if g.phase != ShamrocksPhasePlaying {
		return errors.New("shamrocks: game is not in playing phase")
	}
	if from < 0 || from >= len(g.fans) || to < 0 || to >= len(g.fans) {
		return errors.New("shamrocks: invalid fan index")
	}
	if from == to {
		return errors.New("shamrocks: cannot move a fan onto itself")
	}
	card := g.fanTop(from)
	if card == nil {
		return errors.New("shamrocks: source fan is empty")
	}
	if len(g.fans[to]) >= ShamrocksFanSize {
		return errors.New("shamrocks: destination fan is full")
	}
	if !g.canMoveFanToFan(from, to) {
		return errors.New("shamrocks: illegal move")
	}
	g.takeSnapshot()
	g.fans[from] = g.fans[from][:len(g.fans[from])-1]
	g.fans[to] = append(g.fans[to], card)
	g.moveCount++
	g.appendLog("move", fmt.Sprintf("扇%d→扇%d", from, to), []*Card{card})
	g.checkGameOver()
	return nil
}

// MoveFanToFoundation 扇 from のトップをファウンデーションに移す。
func (g *Shamrocks) MoveFanToFoundation(from int) error {
	if g.phase != ShamrocksPhasePlaying {
		return errors.New("shamrocks: game is not in playing phase")
	}
	if from < 0 || from >= len(g.fans) {
		return errors.New("shamrocks: invalid fan index")
	}
	card := g.fanTop(from)
	if card == nil {
		return errors.New("shamrocks: source fan is empty")
	}
	fIdx := g.findFoundation(card)
	if fIdx < 0 {
		return errors.New("shamrocks: cannot place card on foundation")
	}
	g.takeSnapshot()
	g.fans[from] = g.fans[from][:len(g.fans[from])-1]
	g.foundation[fIdx] = append(g.foundation[fIdx], card)
	g.moveCount++
	g.appendLog("foundation", fmt.Sprintf("扇%d→ファウンデーション", from), []*Card{card})
	g.checkGameClear()
	g.checkGameOver()
	return nil
}

// GiveUp 投了する。
func (g *Shamrocks) GiveUp() {
	if g.phase == ShamrocksPhasePlaying {
		g.phase = ShamrocksPhaseGameOver
		g.appendLog("giveup", "ギブアップしました", nil)
	}
}

// hasAnyLegalMove ファウンデーション手・扇間移動のいずれかが存在するか。
func (g *Shamrocks) hasAnyLegalMove() bool {
	for i := range g.fans {
		card := g.fanTop(i)
		if card == nil {
			continue
		}
		if g.findFoundation(card) >= 0 {
			return true
		}
		for j := range g.fans {
			if g.canMoveFanToFan(i, j) {
				return true
			}
		}
	}
	return false
}

// checkGameClear 52 枚すべてファウンデーションに揃ったらクリア。
func (g *Shamrocks) checkGameClear() {
	total := 0
	for _, pile := range g.foundation {
		total += len(pile)
	}
	if total == 52 {
		g.phase = ShamrocksPhaseGameClear
		g.appendLog("clear", "クリア！", nil)
	}
}

// checkGameOver 再シャッフルを使い切り、かつ合法手が無ければ失敗。
func (g *Shamrocks) checkGameOver() {
	if g.phase != ShamrocksPhasePlaying {
		return
	}
	if !g.hasAnyLegalMove() {
		g.phase = ShamrocksPhaseGameOver
		g.appendLog("gameover", "手詰まりです", nil)
	}
}

// GetHint 推奨手を 1 つ返す (なければ nil)。ファウンデーション手を優先する。
func (g *Shamrocks) GetHint() *ShamrocksHint {
	if g.phase != ShamrocksPhasePlaying {
		return nil
	}
	for i := range g.fans {
		if card := g.fanTop(i); card != nil && g.findFoundation(card) >= 0 {
			return &ShamrocksHint{FromFan: i, ToFoundation: true}
		}
	}
	for i := range g.fans {
		card := g.fanTop(i)
		if card == nil {
			continue
		}
		for j := range g.fans {
			if g.canMoveFanToFan(i, j) {
				return &ShamrocksHint{FromFan: i, ToFan: j}
			}
		}
	}
	return nil
}

// AutoComplete 出せるファウンデーション手が無くなるまで自動で出し切る。
func (g *Shamrocks) AutoComplete() error {
	if g.phase != ShamrocksPhasePlaying {
		return errors.New("shamrocks: game is not in playing phase")
	}
	for {
		moved := false
		for i := range g.fans {
			if card := g.fanTop(i); card != nil && g.findFoundation(card) >= 0 {
				if err := g.MoveFanToFoundation(i); err != nil {
					return err
				}
				moved = true
				break
			}
		}
		if !moved || g.phase != ShamrocksPhasePlaying {
			return nil
		}
	}
}

// --- undo ---

// takeSnapshot 現在の状態を保存する。
func (g *Shamrocks) takeSnapshot() {
	snap := &shamrocksSnapshot{phase: g.phase, moveCount: g.moveCount}
	snap.fans = make([][]*Card, len(g.fans))
	for i, fan := range g.fans {
		snap.fans[i] = make([]*Card, len(fan))
		copy(snap.fans[i], fan)
	}
	for i := range g.foundation {
		snap.foundation[i] = make([]*Card, len(g.foundation[i]))
		copy(snap.foundation[i], g.foundation[i])
	}
	g.history = appendSnapshot(g.history, snap)
}

func (g *Shamrocks) restoreSnapshot(snap *shamrocksSnapshot) {
	g.fans = snap.fans
	g.foundation = snap.foundation
	g.phase = snap.phase
	g.moveCount = snap.moveCount
}

// Undo 直近の 1 手を取り消す。
func (g *Shamrocks) Undo() error {
	if len(g.history) == 0 {
		return errors.New("shamrocks: nothing to undo")
	}
	snap := g.history[len(g.history)-1]
	g.history = g.history[:len(g.history)-1]
	g.restoreSnapshot(snap)
	return nil
}

// CanUndo Undo 可能か。
func (g *Shamrocks) CanUndo() bool { return len(g.history) > 0 }

// UndoN n 回 Undo する。
func (g *Shamrocks) UndoN(n int) error {
	for i := 0; i < n; i++ {
		if len(g.history) == 0 {
			break
		}
		if err := g.Undo(); err != nil {
			return err
		}
	}
	return nil
}

func (g *Shamrocks) appendLog(action, detail string, cards []*Card) {
	g.appendLogAt(g.moveCount, 0, action, detail, cards)
}

// --- accessors ---

// GetPhase 現在のフェーズ。
func (g *Shamrocks) GetPhase() ShamrocksPhase { return g.phase }

// SetPhase フェーズを設定する (テスト用)。
func (g *Shamrocks) SetPhase(p ShamrocksPhase) { g.phase = p }

// GetGameEndFlag プレイ中でなければ true。
func (g *Shamrocks) GetGameEndFlag() bool { return g.phase != ShamrocksPhasePlaying }

// GetMoveCount 累計手数。
func (g *Shamrocks) GetMoveCount() int { return g.moveCount }

// HasAnyLegalMove はファウンデーション手・扇間移動のいずれかが存在するかを返す
// (合法手がなければリディールが必要)。
func (g *Shamrocks) HasAnyLegalMove() bool { return g.hasAnyLegalMove() }

// GetFans 扇の一覧を返す。
func (g *Shamrocks) GetFans() [][]*Card { return g.fans }

// GetFoundation ファウンデーションを返す。
func (g *Shamrocks) GetFoundation() [ShamrocksFoundationCnt][]*Card { return g.foundation }

// GetActionLog アクションログを返す。
func (g *Shamrocks) GetActionLog() []*ActionLogEntry { return g.actionLog }
