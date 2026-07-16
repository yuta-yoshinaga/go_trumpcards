//go:build !js || !wasm || classic

package domain

import (
	"errors"
	"fmt"
	"math/rand"
)

// La Belle Lucie (ラ・ベル・ルーシー) 盤面定数。
const (
	// LaBelleLucieFoundationCnt ファウンデーションの数 (4 スート)。
	LaBelleLucieFoundationCnt = 4
	// LaBelleLucieFanSize 1 つの扇 (fan) の初期枚数。
	LaBelleLucieFanSize = 3
	// LaBelleLucieMaxRedeals 再シャッフル (集めて配り直し) の最大回数。
	LaBelleLucieMaxRedeals = 3
	// laBelleLucieMaxSliceLen JSON 復元時のスライス長上限。
	laBelleLucieMaxSliceLen = 10000
)

// LaBelleLuciePhase ゲームフェーズ。
type LaBelleLuciePhase int

// La Belle Lucie のフェーズ定数。
const (
	// LaBelleLuciePhasePlaying プレイ中。
	LaBelleLuciePhasePlaying LaBelleLuciePhase = iota
	// LaBelleLuciePhaseGameClear クリア (52 枚すべてファウンデーション)。
	LaBelleLuciePhaseGameClear
	// LaBelleLuciePhaseGameOver 失敗 (再シャッフル使い切り後に手詰まり / ギブアップ)。
	LaBelleLuciePhaseGameOver
)

// LaBelleLucieHint 推奨手。ToFoundation=true ならファウンデーションへ、
// false なら ToFan の扇へ移す。
type LaBelleLucieHint struct {
	FromFan      int
	ToFan        int
	ToFoundation bool
}

// LaBelleLucie La Belle Lucie (ファン・ソリティア) 本体。状態のみを保持する。
type LaBelleLucie struct {
	trumpCards  *TrumpCards
	fans        [][]*Card
	foundation  [LaBelleLucieFoundationCnt][]*Card
	redealsLeft int
	phase       LaBelleLuciePhase
	moveCount   int
	actionLog   []*ActionLogEntry
	history     []*laBelleLucieSnapshot
}

// laBelleLucieSnapshot Undo 用スナップショット。
type laBelleLucieSnapshot struct {
	fans        [][]*Card
	foundation  [LaBelleLucieFoundationCnt][]*Card
	redealsLeft int
	phase       LaBelleLuciePhase
	moveCount   int
}

// NewLaBelleLucie コンストラクタ。
func NewLaBelleLucie(trumpCards *TrumpCards) *LaBelleLucie {
	return &LaBelleLucie{trumpCards: trumpCards}
}

// NewDefaultLaBelleLucie 標準 52 枚デッキで生成する。
func NewDefaultLaBelleLucie() *LaBelleLucie {
	return NewLaBelleLucie(NewTrumpCards(0))
}

// Reset ゲームを初期化する。
func (g *LaBelleLucie) Reset() {
	g.trumpCards.Shuffle()
	g.phase = LaBelleLuciePhasePlaying
	g.moveCount = 0
	g.redealsLeft = LaBelleLucieMaxRedeals
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
func (g *LaBelleLucie) dealFans(cards []*Card) {
	g.fans = nil
	for i := 0; i < len(cards); i += LaBelleLucieFanSize {
		end := i + LaBelleLucieFanSize
		if end > len(cards) {
			end = len(cards)
		}
		fan := make([]*Card, end-i)
		copy(fan, cards[i:end])
		g.fans = append(g.fans, fan)
	}
}

// --- rules ---

// canPlaceOnFan 扇トップ dstTop に card を積めるか (同スート・降順)。
func laBelleLucieCanStack(card, dstTop *Card) bool {
	if card == nil || dstTop == nil {
		return false
	}
	return card.GetDesign() == dstTop.GetDesign() && card.GetValue() == dstTop.GetValue()-1
}

// findFoundation card を置けるファウンデーション番号を返す (なければ -1)。
func (g *LaBelleLucie) findFoundation(card *Card) int {
	for i := 0; i < LaBelleLucieFoundationCnt; i++ {
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
func (g *LaBelleLucie) fanTop(idx int) *Card {
	if idx < 0 || idx >= len(g.fans) || len(g.fans[idx]) == 0 {
		return nil
	}
	return g.fans[idx][len(g.fans[idx])-1]
}

// --- public actions ---

// MoveFanToFan 扇 from のトップを扇 to に積む。
func (g *LaBelleLucie) MoveFanToFan(from, to int) error {
	if g.phase != LaBelleLuciePhasePlaying {
		return errors.New("labellelucie: game is not in playing phase")
	}
	if from < 0 || from >= len(g.fans) || to < 0 || to >= len(g.fans) {
		return errors.New("labellelucie: invalid fan index")
	}
	if from == to {
		return errors.New("labellelucie: cannot move a fan onto itself")
	}
	card := g.fanTop(from)
	if card == nil {
		return errors.New("labellelucie: source fan is empty")
	}
	if !laBelleLucieCanStack(card, g.fanTop(to)) {
		return errors.New("labellelucie: illegal move")
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
func (g *LaBelleLucie) MoveFanToFoundation(from int) error {
	if g.phase != LaBelleLuciePhasePlaying {
		return errors.New("labellelucie: game is not in playing phase")
	}
	if from < 0 || from >= len(g.fans) {
		return errors.New("labellelucie: invalid fan index")
	}
	card := g.fanTop(from)
	if card == nil {
		return errors.New("labellelucie: source fan is empty")
	}
	fIdx := g.findFoundation(card)
	if fIdx < 0 {
		return errors.New("labellelucie: cannot place card on foundation")
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

// Redeal ファウンデーション以外のカードを集めてシャッフルし配り直す。
func (g *LaBelleLucie) Redeal() error {
	if g.phase != LaBelleLuciePhasePlaying {
		return errors.New("labellelucie: game is not in playing phase")
	}
	if g.redealsLeft <= 0 {
		return errors.New("labellelucie: no redeals left")
	}
	g.takeSnapshot()
	var cards []*Card
	for _, fan := range g.fans {
		cards = append(cards, fan...)
	}
	rand.Shuffle(len(cards), func(i, j int) { cards[i], cards[j] = cards[j], cards[i] })
	g.redealsLeft--
	g.dealFans(cards)
	g.moveCount++
	g.appendLog("redeal", fmt.Sprintf("集めてシャッフル (残り%d回)", g.redealsLeft), nil)
	g.checkGameOver()
	return nil
}

// GiveUp 投了する。
func (g *LaBelleLucie) GiveUp() {
	if g.phase == LaBelleLuciePhasePlaying {
		g.phase = LaBelleLuciePhaseGameOver
		g.appendLog("giveup", "ギブアップしました", nil)
	}
}

// hasAnyLegalMove ファウンデーション手・扇間移動のいずれかが存在するか。
func (g *LaBelleLucie) hasAnyLegalMove() bool {
	for i := range g.fans {
		card := g.fanTop(i)
		if card == nil {
			continue
		}
		if g.findFoundation(card) >= 0 {
			return true
		}
		for j := range g.fans {
			if i != j && laBelleLucieCanStack(card, g.fanTop(j)) {
				return true
			}
		}
	}
	return false
}

// checkGameClear 52 枚すべてファウンデーションに揃ったらクリア。
func (g *LaBelleLucie) checkGameClear() {
	total := 0
	for _, pile := range g.foundation {
		total += len(pile)
	}
	if total == 52 {
		g.phase = LaBelleLuciePhaseGameClear
		g.appendLog("clear", "クリア！", nil)
	}
}

// checkGameOver 再シャッフルを使い切り、かつ合法手が無ければ失敗。
func (g *LaBelleLucie) checkGameOver() {
	if g.phase != LaBelleLuciePhasePlaying {
		return
	}
	if g.redealsLeft == 0 && !g.hasAnyLegalMove() {
		g.phase = LaBelleLuciePhaseGameOver
		g.appendLog("gameover", "手詰まりです", nil)
	}
}

// GetHint 推奨手を 1 つ返す (なければ nil)。ファウンデーション手を優先する。
func (g *LaBelleLucie) GetHint() *LaBelleLucieHint {
	if g.phase != LaBelleLuciePhasePlaying {
		return nil
	}
	for i := range g.fans {
		if card := g.fanTop(i); card != nil && g.findFoundation(card) >= 0 {
			return &LaBelleLucieHint{FromFan: i, ToFoundation: true}
		}
	}
	for i := range g.fans {
		card := g.fanTop(i)
		if card == nil {
			continue
		}
		for j := range g.fans {
			if i != j && laBelleLucieCanStack(card, g.fanTop(j)) {
				return &LaBelleLucieHint{FromFan: i, ToFan: j}
			}
		}
	}
	return nil
}

// AutoComplete 出せるファウンデーション手が無くなるまで自動で出し切る。
func (g *LaBelleLucie) AutoComplete() error {
	if g.phase != LaBelleLuciePhasePlaying {
		return errors.New("labellelucie: game is not in playing phase")
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
		if !moved || g.phase != LaBelleLuciePhasePlaying {
			return nil
		}
	}
}

// --- undo ---

// takeSnapshot 現在の状態を保存する。
func (g *LaBelleLucie) takeSnapshot() {
	snap := &laBelleLucieSnapshot{redealsLeft: g.redealsLeft, phase: g.phase, moveCount: g.moveCount}
	snap.fans = make([][]*Card, len(g.fans))
	for i, fan := range g.fans {
		snap.fans[i] = make([]*Card, len(fan))
		copy(snap.fans[i], fan)
	}
	for i := range g.foundation {
		snap.foundation[i] = make([]*Card, len(g.foundation[i]))
		copy(snap.foundation[i], g.foundation[i])
	}
	g.history = append(g.history, snap)
}

func (g *LaBelleLucie) restoreSnapshot(snap *laBelleLucieSnapshot) {
	g.fans = snap.fans
	g.foundation = snap.foundation
	g.redealsLeft = snap.redealsLeft
	g.phase = snap.phase
	g.moveCount = snap.moveCount
}

// Undo 直近の 1 手を取り消す。
func (g *LaBelleLucie) Undo() error {
	if len(g.history) == 0 {
		return errors.New("labellelucie: nothing to undo")
	}
	snap := g.history[len(g.history)-1]
	g.history = g.history[:len(g.history)-1]
	g.restoreSnapshot(snap)
	return nil
}

// CanUndo Undo 可能か。
func (g *LaBelleLucie) CanUndo() bool { return len(g.history) > 0 }

// UndoN n 回 Undo する。
func (g *LaBelleLucie) UndoN(n int) error {
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

func (g *LaBelleLucie) appendLog(action, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.moveCount,
		PlayerIdx:  0,
		ActionType: action,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- accessors ---

// GetPhase 現在のフェーズ。
func (g *LaBelleLucie) GetPhase() LaBelleLuciePhase { return g.phase }

// SetPhase フェーズを設定する (テスト用)。
func (g *LaBelleLucie) SetPhase(p LaBelleLuciePhase) { g.phase = p }

// GetGameEndFlag プレイ中でなければ true。
func (g *LaBelleLucie) GetGameEndFlag() bool { return g.phase != LaBelleLuciePhasePlaying }

// GetMoveCount 累計手数。
func (g *LaBelleLucie) GetMoveCount() int { return g.moveCount }

// GetRedealsLeft 残り再シャッフル回数。
func (g *LaBelleLucie) GetRedealsLeft() int { return g.redealsLeft }

// HasAnyLegalMove はファウンデーション手・扇間移動のいずれかが存在するかを返す
// (合法手がなければリディールが必要)。
func (g *LaBelleLucie) HasAnyLegalMove() bool { return g.hasAnyLegalMove() }

// GetFans 扇の一覧を返す。
func (g *LaBelleLucie) GetFans() [][]*Card { return g.fans }

// GetFoundation ファウンデーションを返す。
func (g *LaBelleLucie) GetFoundation() [LaBelleLucieFoundationCnt][]*Card { return g.foundation }

// GetActionLog アクションログを返す。
func (g *LaBelleLucie) GetActionLog() []*ActionLogEntry { return g.actionLog }
