//go:build !js || !wasm || solo

package domain

import (
	"errors"
	"fmt"
)

// Black Hole 盤面定数。
const (
	// BlackHoleFanCnt 扇 (fan) の数。51 枚を 3 枚ずつ配る。
	BlackHoleFanCnt = 17
	// BlackHoleFanSize 1 つの扇の初期枚数。
	BlackHoleFanSize = 3
	// BlackHoleTotalCards 1 デッキの総枚数。
	BlackHoleTotalCards = 52
	// blackHoleMaxSliceLen JSON 復元時のスライス長上限。
	blackHoleMaxSliceLen = 10000
)

// BlackHolePhase ゲームフェーズ。
type BlackHolePhase int

// Black Hole のフェーズ定数。
const (
	// BlackHolePhasePlaying プレイ中。
	BlackHolePhasePlaying BlackHolePhase = iota
	// BlackHolePhaseGameClear クリア (全 52 枚を吸い込んだ)。
	BlackHolePhaseGameClear
	// BlackHolePhaseGameOver 行き詰まり / ギブアップ。
	BlackHolePhaseGameOver
)

// BlackHoleHint 推奨手。Fan は移動可能な扇のインデックス。
type BlackHoleHint struct {
	Fan int
}

// BlackHole Black Hole (ブラックホール) 本体。状態のみを保持する。
type BlackHole struct {
	trumpCards *TrumpCards
	fans       [][]*Card
	blackHole  []*Card
	phase      BlackHolePhase
	moveCount  int
	actionLog  []*ActionLogEntry
	history    []*blackHoleSnapshot
}

// blackHoleSnapshot Undo 用スナップショット。
type blackHoleSnapshot struct {
	fans      [][]*Card
	blackHole []*Card
	phase     BlackHolePhase
	moveCount int
}

// NewBlackHole コンストラクタ。
func NewBlackHole(trumpCards *TrumpCards) *BlackHole {
	return &BlackHole{trumpCards: trumpCards}
}

// NewDefaultBlackHole 標準 52 枚デッキで生成する。
func NewDefaultBlackHole() *BlackHole {
	return NewBlackHole(NewTrumpCards(0))
}

// Reset ゲームを初期化する。♠A を中央のブラックホールに固定し、残り 51 枚を
// 17 の扇 (3 枚ずつ) に配る。
func (g *BlackHole) Reset() {
	g.trumpCards.Shuffle()
	g.phase = BlackHolePhasePlaying
	g.moveCount = 0
	g.actionLog = nil
	g.history = nil

	cards := make([]*Card, 0, BlackHoleTotalCards)
	var start *Card
	for g.trumpCards.GetRemainingCount() > 0 {
		c := g.trumpCards.DrawCard()
		if start == nil && c.GetDesign() == CardDesignSpade && c.GetValue() == 1 {
			start = c // ♠A を中央へ固定する。
			continue
		}
		cards = append(cards, c)
	}
	if start == nil && len(cards) > 0 {
		// 念のため (♠A が見つからない構成) 先頭を中央に置く。
		start = cards[0]
		cards = cards[1:]
	}
	if start != nil {
		g.blackHole = []*Card{start}
	} else {
		g.blackHole = nil
	}
	g.dealFans(cards)
	g.appendLog("deal", "新しいゲームを開始しました", nil)
}

// dealFans カード列を 3 枚ずつの扇に配る。
func (g *BlackHole) dealFans(cards []*Card) {
	g.fans = make([][]*Card, 0, BlackHoleFanCnt)
	for i := 0; i < len(cards); i += BlackHoleFanSize {
		end := i + BlackHoleFanSize
		if end > len(cards) {
			end = len(cards)
		}
		fan := make([]*Card, end-i)
		copy(fan, cards[i:end])
		g.fans = append(g.fans, fan)
	}
}

// --- rules ---

// blackHoleAdjacent 2 枚のランクが ±1 か (スート不問・K-A ラップなし)。
func blackHoleAdjacent(a, b *Card) bool {
	if a == nil || b == nil {
		return false
	}
	diff := a.GetValue() - b.GetValue()
	if diff < 0 {
		diff = -diff
	}
	return diff == 1
}

// blackHoleTop ブラックホールの最上位カード。
func (g *BlackHole) blackHoleTop() *Card {
	if len(g.blackHole) == 0 {
		return nil
	}
	return g.blackHole[len(g.blackHole)-1]
}

// fanTop 扇 idx のトップカードを返す (空・範囲外は nil)。
func (g *BlackHole) fanTop(idx int) *Card {
	if idx < 0 || idx >= len(g.fans) || len(g.fans[idx]) == 0 {
		return nil
	}
	return g.fans[idx][len(g.fans[idx])-1]
}

// canPlay 扇 idx のトップをブラックホールへ積めるか。
func (g *BlackHole) canPlay(idx int) bool {
	return blackHoleAdjacent(g.fanTop(idx), g.blackHoleTop())
}

// AcceptableRanks はいまブラックホールが受け付けるランクを昇順で返す。
// 穴のトップの ±1（K-A ラップなし、1..13 でクランプ）。穴が空なら nil。
//
// **CUI は穴のトップしか出しておらず、±1 を毎回暗算させていた (#4818)。**
func (g *BlackHole) AcceptableRanks() []int {
	top := g.blackHoleTop()
	if top == nil {
		return nil
	}
	var out []int
	for _, v := range []int{top.GetValue() - 1, top.GetValue() + 1} {
		if v >= 1 && v <= CardValueMax {
			out = append(out, v)
		}
	}
	return out
}

// PlayableFans はいま積める扇の番号を返す。canPlay と同じ判定を使う。
func (g *BlackHole) PlayableFans() []int {
	var out []int
	for i := range g.fans {
		if g.canPlay(i) {
			out = append(out, i)
		}
	}
	return out
}

// --- public actions ---

// MoveFanToBlackHole 扇 idx のトップカードをブラックホールへ積む。
func (g *BlackHole) MoveFanToBlackHole(idx int) error {
	if g.phase != BlackHolePhasePlaying {
		return errors.New("blackhole: game is not in playing phase")
	}
	if idx < 0 || idx >= len(g.fans) {
		return errors.New("blackhole: invalid fan index")
	}
	card := g.fanTop(idx)
	if card == nil {
		return errors.New("blackhole: fan is empty")
	}
	if !g.canPlay(idx) {
		return errors.New("blackhole: card is not adjacent to the black hole top")
	}
	g.takeSnapshot()
	g.fans[idx] = g.fans[idx][:len(g.fans[idx])-1]
	g.blackHole = append(g.blackHole, card)
	g.moveCount++
	g.appendLog("move", fmt.Sprintf("扇%d→ブラックホール", idx), []*Card{card})
	g.checkGameEnd()
	return nil
}

// GiveUp 投了する。
func (g *BlackHole) GiveUp() {
	if g.phase == BlackHolePhasePlaying {
		g.phase = BlackHolePhaseGameOver
		g.appendLog("giveup", "投了しました", nil)
	}
}

// GetHint 次に積める扇のインデックスを返す (なければ nil)。
func (g *BlackHole) GetHint() *BlackHoleHint {
	for i := range g.fans {
		if g.canPlay(i) {
			return &BlackHoleHint{Fan: i}
		}
	}
	return nil
}

// hasAnyLegalMove 合法手が 1 つでもあるか。
func (g *BlackHole) hasAnyLegalMove() bool {
	for i := range g.fans {
		if g.canPlay(i) {
			return true
		}
	}
	return false
}

// IsStalemate プレイ中で合法手がない状態か。
func (g *BlackHole) IsStalemate() bool {
	return g.phase == BlackHolePhasePlaying && !g.hasAnyLegalMove()
}

// checkGameEnd クリア / 行き詰まりを判定してフェーズを更新する。
func (g *BlackHole) checkGameEnd() {
	if len(g.blackHole) == BlackHoleTotalCards {
		g.phase = BlackHolePhaseGameClear
		g.appendLog("clear", "ブラックホールに全カードを吸い込みました", nil)
		return
	}
	if !g.hasAnyLegalMove() {
		g.phase = BlackHolePhaseGameOver
		g.appendLog("gameover", "合法手がなくなりました", nil)
	}
}

// --- undo ---

// takeSnapshot 現在の状態を履歴に積む。
func (g *BlackHole) takeSnapshot() {
	snap := &blackHoleSnapshot{phase: g.phase, moveCount: g.moveCount}
	snap.fans = make([][]*Card, len(g.fans))
	for i := range g.fans {
		snap.fans[i] = make([]*Card, len(g.fans[i]))
		copy(snap.fans[i], g.fans[i])
	}
	snap.blackHole = make([]*Card, len(g.blackHole))
	copy(snap.blackHole, g.blackHole)
	g.history = append(g.history, snap)
}

// restoreSnapshot スナップショットを復元する。
func (g *BlackHole) restoreSnapshot(snap *blackHoleSnapshot) {
	g.fans = snap.fans
	g.blackHole = snap.blackHole
	g.phase = snap.phase
	g.moveCount = snap.moveCount
}

// Undo 直近の 1 手を取り消す。
func (g *BlackHole) Undo() error {
	if len(g.history) == 0 {
		return errors.New("blackhole: nothing to undo")
	}
	snap := g.history[len(g.history)-1]
	g.history = g.history[:len(g.history)-1]
	g.restoreSnapshot(snap)
	return nil
}

// CanUndo Undo 可能か。
func (g *BlackHole) CanUndo() bool { return len(g.history) > 0 }

// UndoN n 回 Undo する。
func (g *BlackHole) UndoN(n int) error {
	for i := 0; i < n; i++ {
		if len(g.history) == 0 {
			break
		}
		_ = g.Undo() // guarded above: Undo only errors on an empty history.
	}
	return nil
}

// --- accessors ---

// GetPhase 現在のフェーズ。
func (g *BlackHole) GetPhase() BlackHolePhase { return g.phase }

// SetPhase フェーズを設定する (テスト用)。
func (g *BlackHole) SetPhase(p BlackHolePhase) { g.phase = p }

// GetGameEndFlag プレイ中でなければ true。
func (g *BlackHole) GetGameEndFlag() bool { return g.phase != BlackHolePhasePlaying }

// GetMoveCount 累計手数。
func (g *BlackHole) GetMoveCount() int { return g.moveCount }

// GetFans 扇の一覧を返す。
func (g *BlackHole) GetFans() [][]*Card { return g.fans }

// GetBlackHole ブラックホールの積み上げを返す。
func (g *BlackHole) GetBlackHole() []*Card { return g.blackHole }

// GetActionLog アクションログを返す。
func (g *BlackHole) GetActionLog() []*ActionLogEntry { return g.actionLog }

// appendLog アクションログを追加する。
func (g *BlackHole) appendLog(action, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.moveCount,
		PlayerIdx:  0,
		ActionType: action,
		Detail:     detail,
		Cards:      cards,
	})
}
