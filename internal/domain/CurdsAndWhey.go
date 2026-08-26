//go:build !js || !wasm || classic

package domain

import (
	"errors"
	"fmt"
)

// Curds and Whey 盤面定数。
const (
	// CurdsAndWheyColCnt タブロー列数。
	CurdsAndWheyColCnt = 13
	// CurdsAndWheyFoundationCnt 完成させるスート数。
	CurdsAndWheyFoundationCnt = 4
	// curdsAndWheyMaxSliceLen JSON 復元時のスライス長上限。
	curdsAndWheyMaxSliceLen = 10000
)

// curdsAndWheyDeal 各列の初期配布枚数 (合計 52)。
// curdsAndWheyDeal は各列の枚数。52 / 13 = 4 で割り切れるので全列が同じ。
// Simple Simon は 8,8,8,7,6,5,4,3,2,1 の階段状に配る。
var curdsAndWheyDeal = [CurdsAndWheyColCnt]int{4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4}

// CurdsAndWheyPhase ゲームフェーズ。
type CurdsAndWheyPhase int

// Simple Simon のフェーズ定数。
const (
	// CurdsAndWheyPhasePlaying プレイ中。
	CurdsAndWheyPhasePlaying CurdsAndWheyPhase = iota
	// CurdsAndWheyPhaseGameClear クリア (4 スート完成)。
	CurdsAndWheyPhaseGameClear
	// CurdsAndWheyPhaseGameOver 失敗 (合法手なし) / ギブアップ。
	CurdsAndWheyPhaseGameOver
)

// CurdsAndWheyHint 推奨手。FromCol の CardIndex から始まる同スート降順列を ToCol へ。
type CurdsAndWheyHint struct {
	FromCol   int
	CardIndex int
	ToCol     int
}

// CurdsAndWhey カーズ・アンド・ホエイ本体。状態のみを保持する。
type CurdsAndWhey struct {
	trumpCards     *TrumpCards
	columns        [CurdsAndWheyColCnt][]*Card
	completedSuits int
	phase          CurdsAndWheyPhase
	moveCount      int
	actionLogBase
	history []*curdsAndWheySnapshot
}

// curdsAndWheySnapshot Undo 用スナップショット。
type curdsAndWheySnapshot struct {
	columns        [CurdsAndWheyColCnt][]*Card
	completedSuits int
	phase          CurdsAndWheyPhase
	moveCount      int
}

// NewCurdsAndWhey コンストラクタ。
func NewCurdsAndWhey(trumpCards *TrumpCards) *CurdsAndWhey {
	return &CurdsAndWhey{trumpCards: trumpCards}
}

// NewDefaultCurdsAndWhey 標準 52 枚デッキで生成する。
func NewDefaultCurdsAndWhey() *CurdsAndWhey {
	return NewCurdsAndWhey(NewTrumpCards(0))
}

// Reset ゲームを初期化する。全 52 枚を表向きに 10 列へ配る。
func (g *CurdsAndWhey) Reset() {
	g.trumpCards.Shuffle()
	g.phase = CurdsAndWheyPhasePlaying
	g.moveCount = 0
	g.completedSuits = 0
	g.actionLog = nil
	g.history = nil
	for i := 0; i < CurdsAndWheyColCnt; i++ {
		g.columns[i] = make([]*Card, 0, curdsAndWheyDeal[i])
		for j := 0; j < curdsAndWheyDeal[i]; j++ {
			g.columns[i] = append(g.columns[i], g.trumpCards.DrawCard())
		}
	}
	g.appendLog("deal", "新しいゲームを開始しました", nil)
}

// --- rules ---

// isMovableRun cards が同スート降順の連続列か (単体は常に可)。
// CurdsAndWheyMovableFrom は列のうち「まとめて動かせる塊」が始まる位置を返す。
//
// **動かせるのは末尾から同スート降順に途切れず続く部分だけ** (curdsAndWheyIsRun)。
// 画面はどこからが掴めるのかを示すので、判定を書き写さずに済むよう位置を返す
// (#5679)。空列は 0。
func CurdsAndWheyMovableFrom(column []*Card) int {
	if len(column) == 0 {
		return 0
	}
	idx := len(column) - 1
	for idx > 0 && curdsAndWheyIsRun(column[idx-1:idx+1]) {
		idx--
	}
	return idx
}

func curdsAndWheyIsRun(cards []*Card) bool {
	for i := 1; i < len(cards); i++ {
		if cards[i].GetDesign() != cards[i-1].GetDesign() || cards[i].GetValue() != cards[i-1].GetValue()-1 {
			return false
		}
	}
	return true
}

// canPlace card を列 toCol に置けるか。空列は任意。
//
// それ以外は次の 2 通り:
//
//	同スートで 1 つ上   ♠7 の上に ♠6
//	同ランク           ♠7 の上に ♥7（同ランクの仮置き）
//
// Simple Simon は「スート不問で 1 つ上」なので、**スート条件と同ランク許容の
// 両方が違う**。片方だけ直すと、もう片方が Simple Simon のまま残る。
func (g *CurdsAndWhey) canPlace(card *Card, toCol int) bool {
	pile := g.columns[toCol]
	if len(pile) == 0 {
		return true
	}
	top := pile[len(pile)-1]
	if top.GetValue() == card.GetValue() {
		return true
	}
	return top.GetDesign() == card.GetDesign() && top.GetValue() == card.GetValue()+1
}

// MoveSequence 列 fromCol の cardIndex 以降を列 toCol へ移す。
func (g *CurdsAndWhey) MoveSequence(fromCol, cardIndex, toCol int) error {
	if g.phase != CurdsAndWheyPhasePlaying {
		return errors.New("curdsandwhey: game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= CurdsAndWheyColCnt || toCol < 0 || toCol >= CurdsAndWheyColCnt {
		return errors.New("curdsandwhey: invalid column index")
	}
	if fromCol == toCol {
		return errors.New("curdsandwhey: source and destination are the same")
	}
	src := g.columns[fromCol]
	if cardIndex < 0 || cardIndex >= len(src) {
		return errors.New("curdsandwhey: invalid card index")
	}
	moving := src[cardIndex:]
	if !curdsAndWheyIsRun(moving) {
		return errors.New("curdsandwhey: cards are not a same-suit descending run")
	}
	if !g.canPlace(moving[0], toCol) {
		return errors.New("curdsandwhey: illegal move")
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
func (g *CurdsAndWhey) checkAndRemoveCompleted(col int) {
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
func (g *CurdsAndWhey) checkGameClear() {
	if g.completedSuits >= CurdsAndWheyFoundationCnt {
		g.phase = CurdsAndWheyPhaseGameClear
		g.appendLog("clear", "クリア！", nil)
	}
}

// hasAnyLegalMove 1 つ以上の合法手があるか。
func (g *CurdsAndWhey) hasAnyLegalMove() bool {
	for from := 0; from < CurdsAndWheyColCnt; from++ {
		src := g.columns[from]
		for idx := 0; idx < len(src); idx++ {
			if !curdsAndWheyIsRun(src[idx:]) {
				continue
			}
			card := src[idx]
			for to := 0; to < CurdsAndWheyColCnt; to++ {
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
func (g *CurdsAndWhey) checkGameOver() {
	if g.phase == CurdsAndWheyPhasePlaying && !g.hasAnyLegalMove() {
		g.phase = CurdsAndWheyPhaseGameOver
		g.appendLog("gameover", "手詰まりです", nil)
	}
}

// GiveUp 投了する。
func (g *CurdsAndWhey) GiveUp() {
	if g.phase == CurdsAndWheyPhasePlaying {
		g.phase = CurdsAndWheyPhaseGameOver
		g.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 推奨手を 1 つ返す (なければ nil)。完成を狙えるカード列を優先する。
func (g *CurdsAndWhey) GetHint() *CurdsAndWheyHint {
	if g.phase != CurdsAndWheyPhasePlaying {
		return nil
	}
	for from := 0; from < CurdsAndWheyColCnt; from++ {
		src := g.columns[from]
		for idx := 0; idx < len(src); idx++ {
			if !curdsAndWheyIsRun(src[idx:]) {
				continue
			}
			card := src[idx]
			for to := 0; to < CurdsAndWheyColCnt; to++ {
				if to == from {
					continue
				}
				if idx == 0 && len(g.columns[to]) == 0 {
					continue // moving a whole column to an empty one is not progress
				}
				if g.canPlace(card, to) {
					return &CurdsAndWheyHint{FromCol: from, CardIndex: idx, ToCol: to}
				}
			}
		}
	}
	return nil
}

func (g *CurdsAndWhey) appendLog(action, detail string, cards []*Card) {
	g.appendLogAt(g.moveCount, 0, action, detail, cards)
}

// --- undo ---

func (g *CurdsAndWhey) takeSnapshot() {
	snap := &curdsAndWheySnapshot{completedSuits: g.completedSuits, phase: g.phase, moveCount: g.moveCount}
	for i := 0; i < CurdsAndWheyColCnt; i++ {
		snap.columns[i] = make([]*Card, len(g.columns[i]))
		copy(snap.columns[i], g.columns[i])
	}
	g.history = append(g.history, snap)
}

func (g *CurdsAndWhey) restoreSnapshot(snap *curdsAndWheySnapshot) {
	g.columns = snap.columns
	g.completedSuits = snap.completedSuits
	g.phase = snap.phase
	g.moveCount = snap.moveCount
}

// Undo 直近の 1 手を取り消す。
func (g *CurdsAndWhey) Undo() error {
	if len(g.history) == 0 {
		return errors.New("curdsandwhey: nothing to undo")
	}
	snap := g.history[len(g.history)-1]
	g.history = g.history[:len(g.history)-1]
	g.restoreSnapshot(snap)
	return nil
}

// CanUndo Undo 可能か。
func (g *CurdsAndWhey) CanUndo() bool { return len(g.history) > 0 }

// UndoN n 回 Undo する。
func (g *CurdsAndWhey) UndoN(n int) error {
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
func (g *CurdsAndWhey) GetPhase() CurdsAndWheyPhase { return g.phase }

// SetPhase フェーズを設定する (テスト用)。
func (g *CurdsAndWhey) SetPhase(p CurdsAndWheyPhase) { g.phase = p }

// GetGameEndFlag プレイ中でなければ true。
func (g *CurdsAndWhey) GetGameEndFlag() bool { return g.phase != CurdsAndWheyPhasePlaying }

// GetMoveCount 累計手数。
func (g *CurdsAndWhey) GetMoveCount() int { return g.moveCount }

// GetCompletedSuits 完成スート数。
func (g *CurdsAndWhey) GetCompletedSuits() int { return g.completedSuits }

// GetColumns タブロー列を返す。
func (g *CurdsAndWhey) GetColumns() [CurdsAndWheyColCnt][]*Card { return g.columns }

// GetActionLog アクションログを返す。
func (g *CurdsAndWhey) GetActionLog() []*ActionLogEntry { return g.actionLog }
