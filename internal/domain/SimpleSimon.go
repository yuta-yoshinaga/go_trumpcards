//go:build !js || !wasm || classic

package domain

import (
	"errors"
	"fmt"
)

// Simple Simon 盤面定数。
const (
	// SimpleSimonColCnt タブロー列数。
	SimpleSimonColCnt = 10
	// SimpleSimonFoundationCnt 完成させるスート数。
	SimpleSimonFoundationCnt = 4
	// simpleSimonMaxSliceLen JSON 復元時のスライス長上限。
	simpleSimonMaxSliceLen = 10000
)

// simpleSimonDeal 各列の初期配布枚数 (合計 52)。
var simpleSimonDeal = [SimpleSimonColCnt]int{8, 8, 8, 7, 6, 5, 4, 3, 2, 1}

// SimpleSimonPhase ゲームフェーズ。
type SimpleSimonPhase int

// Simple Simon のフェーズ定数。
const (
	// SimpleSimonPhasePlaying プレイ中。
	SimpleSimonPhasePlaying SimpleSimonPhase = iota
	// SimpleSimonPhaseGameClear クリア (4 スート完成)。
	SimpleSimonPhaseGameClear
	// SimpleSimonPhaseGameOver 失敗 (合法手なし) / ギブアップ。
	SimpleSimonPhaseGameOver
)

// SimpleSimonHint 推奨手。FromCol の CardIndex から始まる同スート降順列を ToCol へ。
type SimpleSimonHint struct {
	FromCol   int
	CardIndex int
	ToCol     int
}

// SimpleSimon Simple Simon (スパイダー系ソリティア) 本体。状態のみを保持する。
type SimpleSimon struct {
	trumpCards     *TrumpCards
	columns        [SimpleSimonColCnt][]*Card
	completedSuits int
	phase          SimpleSimonPhase
	moveCount      int
	actionLogBase
	history []*simpleSimonSnapshot
}

// simpleSimonSnapshot Undo 用スナップショット。
type simpleSimonSnapshot struct {
	columns        [SimpleSimonColCnt][]*Card
	completedSuits int
	phase          SimpleSimonPhase
	moveCount      int
}

// NewSimpleSimon コンストラクタ。
func NewSimpleSimon(trumpCards *TrumpCards) *SimpleSimon {
	return &SimpleSimon{trumpCards: trumpCards}
}

// NewDefaultSimpleSimon 標準 52 枚デッキで生成する。
func NewDefaultSimpleSimon() *SimpleSimon {
	return NewSimpleSimon(NewTrumpCards(0))
}

// Reset ゲームを初期化する。全 52 枚を表向きに 10 列へ配る。
func (g *SimpleSimon) Reset() {
	g.trumpCards.Shuffle()
	g.phase = SimpleSimonPhasePlaying
	g.moveCount = 0
	g.completedSuits = 0
	g.actionLog = nil
	g.history = nil
	for i := 0; i < SimpleSimonColCnt; i++ {
		g.columns[i] = make([]*Card, 0, simpleSimonDeal[i])
		for j := 0; j < simpleSimonDeal[i]; j++ {
			g.columns[i] = append(g.columns[i], g.trumpCards.DrawCard())
		}
	}
	g.appendLog("deal", "新しいゲームを開始しました", nil)
}

// --- rules ---

// isMovableRun cards が同スート降順の連続列か (単体は常に可)。
// SimpleSimonMovableFrom は列のうち「まとめて動かせる塊」が始まる位置を返す。
//
// **動かせるのは末尾から同スート降順に途切れず続く部分だけ** (simpleSimonIsRun)。
// 画面はどこからが掴めるのかを示すので、判定を書き写さずに済むよう位置を返す
// (#5679)。空列は 0。
func SimpleSimonMovableFrom(column []*Card) int {
	if len(column) == 0 {
		return 0
	}
	idx := len(column) - 1
	for idx > 0 && simpleSimonIsRun(column[idx-1:idx+1]) {
		idx--
	}
	return idx
}

func simpleSimonIsRun(cards []*Card) bool {
	for i := 1; i < len(cards); i++ {
		if cards[i].GetDesign() != cards[i-1].GetDesign() || cards[i].GetValue() != cards[i-1].GetValue()-1 {
			return false
		}
	}
	return true
}

// canPlace card を列 toCol に置けるか。空列は任意、それ以外は ToCol トップが
// card より 1 つ大きい (スート不問)。
func (g *SimpleSimon) canPlace(card *Card, toCol int) bool {
	pile := g.columns[toCol]
	if len(pile) == 0 {
		return true
	}
	top := pile[len(pile)-1]
	return top.GetValue() == card.GetValue()+1
}

// MoveSequence 列 fromCol の cardIndex 以降を列 toCol へ移す。
func (g *SimpleSimon) MoveSequence(fromCol, cardIndex, toCol int) error {
	if g.phase != SimpleSimonPhasePlaying {
		return errors.New("simplesimon: game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= SimpleSimonColCnt || toCol < 0 || toCol >= SimpleSimonColCnt {
		return errors.New("simplesimon: invalid column index")
	}
	if fromCol == toCol {
		return errors.New("simplesimon: source and destination are the same")
	}
	src := g.columns[fromCol]
	if cardIndex < 0 || cardIndex >= len(src) {
		return errors.New("simplesimon: invalid card index")
	}
	moving := src[cardIndex:]
	if !simpleSimonIsRun(moving) {
		return errors.New("simplesimon: cards are not a same-suit descending run")
	}
	if !g.canPlace(moving[0], toCol) {
		return errors.New("simplesimon: illegal move")
	}
	g.takeSnapshot()
	g.columns[fromCol] = src[:cardIndex]
	g.columns[toCol] = append(g.columns[toCol], moving...)
	g.moveCount++
	g.appendLog("move", fmt.Sprintf("列%d[%d]→列%d", fromCol, cardIndex, toCol), nil)
	g.checkAndRemoveCompleted(toCol)
	g.checkGameOver()
	return nil
}

// checkAndRemoveCompleted 列 col の末尾 13 枚が同スート K→A なら除去する。
func (g *SimpleSimon) checkAndRemoveCompleted(col int) {
	cards := g.columns[col]
	if len(cards) < CardValueMax {
		return
	}
	seq := cards[len(cards)-CardValueMax:]
	suit := seq[0].GetDesign()
	for i, c := range seq {
		if c.GetDesign() != suit || c.GetValue() != CardValueMax-i {
			return
		}
	}
	g.columns[col] = cards[:len(cards)-CardValueMax]
	g.completedSuits++
	g.appendLog("complete", fmt.Sprintf("列%dでスートが完成しました", col), nil)
	g.checkGameClear()
}

// checkGameClear 4 スート完成でクリア。
func (g *SimpleSimon) checkGameClear() {
	if g.completedSuits >= SimpleSimonFoundationCnt {
		g.phase = SimpleSimonPhaseGameClear
		g.appendLog("clear", "クリア！", nil)
	}
}

// hasAnyLegalMove 1 つ以上の合法手があるか。
func (g *SimpleSimon) hasAnyLegalMove() bool {
	for from := 0; from < SimpleSimonColCnt; from++ {
		src := g.columns[from]
		for idx := 0; idx < len(src); idx++ {
			if !simpleSimonIsRun(src[idx:]) {
				continue
			}
			card := src[idx]
			for to := 0; to < SimpleSimonColCnt; to++ {
				if to == from {
					continue
				}
				// Moving an entire column onto an empty one exposes nothing and
				// makes no progress, so it does not count as escaping a stalemate.
				if idx == 0 && len(g.columns[to]) == 0 {
					continue
				}
				if g.canPlace(card, to) {
					return true
				}
			}
		}
	}
	return false
}

// checkGameOver 合法手が無ければ失敗。
func (g *SimpleSimon) checkGameOver() {
	if g.phase == SimpleSimonPhasePlaying && !g.hasAnyLegalMove() {
		g.phase = SimpleSimonPhaseGameOver
		g.appendLog("gameover", "手詰まりです", nil)
	}
}

// GiveUp 投了する。
func (g *SimpleSimon) GiveUp() {
	if g.phase == SimpleSimonPhasePlaying {
		g.phase = SimpleSimonPhaseGameOver
		g.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 推奨手を 1 つ返す (なければ nil)。完成を狙えるカード列を優先する。
func (g *SimpleSimon) GetHint() *SimpleSimonHint {
	if g.phase != SimpleSimonPhasePlaying {
		return nil
	}
	for from := 0; from < SimpleSimonColCnt; from++ {
		src := g.columns[from]
		for idx := 0; idx < len(src); idx++ {
			if !simpleSimonIsRun(src[idx:]) {
				continue
			}
			card := src[idx]
			for to := 0; to < SimpleSimonColCnt; to++ {
				if to == from {
					continue
				}
				if idx == 0 && len(g.columns[to]) == 0 {
					continue // moving a whole column to an empty one is not progress
				}
				if g.canPlace(card, to) {
					return &SimpleSimonHint{FromCol: from, CardIndex: idx, ToCol: to}
				}
			}
		}
	}
	return nil
}

func (g *SimpleSimon) appendLog(action, detail string, cards []*Card) {
	g.appendLogAt(g.moveCount, 0, action, detail, cards)
}

// --- undo ---

func (g *SimpleSimon) takeSnapshot() {
	snap := &simpleSimonSnapshot{completedSuits: g.completedSuits, phase: g.phase, moveCount: g.moveCount}
	for i := 0; i < SimpleSimonColCnt; i++ {
		snap.columns[i] = make([]*Card, len(g.columns[i]))
		copy(snap.columns[i], g.columns[i])
	}
	g.history = append(g.history, snap)
}

func (g *SimpleSimon) restoreSnapshot(snap *simpleSimonSnapshot) {
	g.columns = snap.columns
	g.completedSuits = snap.completedSuits
	g.phase = snap.phase
	g.moveCount = snap.moveCount
}

// Undo 直近の 1 手を取り消す。
func (g *SimpleSimon) Undo() error {
	if len(g.history) == 0 {
		return errors.New("simplesimon: nothing to undo")
	}
	snap := g.history[len(g.history)-1]
	g.history = g.history[:len(g.history)-1]
	g.restoreSnapshot(snap)
	return nil
}

// CanUndo Undo 可能か。
func (g *SimpleSimon) CanUndo() bool { return len(g.history) > 0 }

// UndoN n 回 Undo する。
func (g *SimpleSimon) UndoN(n int) error {
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

// --- accessors ---

// GetPhase 現在のフェーズ。
func (g *SimpleSimon) GetPhase() SimpleSimonPhase { return g.phase }

// SetPhase フェーズを設定する (テスト用)。
func (g *SimpleSimon) SetPhase(p SimpleSimonPhase) { g.phase = p }

// GetGameEndFlag プレイ中でなければ true。
func (g *SimpleSimon) GetGameEndFlag() bool { return g.phase != SimpleSimonPhasePlaying }

// GetMoveCount 累計手数。
func (g *SimpleSimon) GetMoveCount() int { return g.moveCount }

// GetCompletedSuits 完成スート数。
func (g *SimpleSimon) GetCompletedSuits() int { return g.completedSuits }

// GetColumns タブロー列を返す。
func (g *SimpleSimon) GetColumns() [SimpleSimonColCnt][]*Card { return g.columns }

// GetActionLog アクションログを返す。
func (g *SimpleSimon) GetActionLog() []*ActionLogEntry { return g.actionLog }
