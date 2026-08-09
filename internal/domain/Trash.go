package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
)

// TrashPhase トラッシュゲームフェーズ
type TrashPhase int

// Trashのフェーズ定数
const (
	// TrashPhasePlayerTurn 現プレイヤーのドロー待ち
	TrashPhasePlayerTurn TrashPhase = iota
	// TrashPhaseAwaitWild ワイルドカードの配置位置入力待ち
	TrashPhaseAwaitWild
	// TrashPhaseGameOver 決着済み
	TrashPhaseGameOver
)

// TrashSlotCnt 1プレイヤーのスロット数
const TrashSlotCnt = 10

// TrashPlayerCnt プレイヤー数 (人間 + CPU)
const TrashPlayerCnt = 2

// TrashJokerCnt 使用ジョーカー数
const TrashJokerCnt = 2

// TrashHumanIdx 人間プレイヤーのインデックス
const TrashHumanIdx = 0

// TrashCpuIdx CPUプレイヤーのインデックス
const TrashCpuIdx = 1

// TrashSlot プレイヤーの1スロット (1〜10いずれかの位置)
type TrashSlot struct {
	Card   *Card
	FaceUp bool
}

// TrashPlayer プレイヤー状態
type TrashPlayer struct {
	Slots [TrashSlotCnt]TrashSlot
	IsCpu bool
}

// Trash トラッシュゲームクラス
type Trash struct {
	trumpCards *TrumpCards
	stock      []*Card
	discard    []*Card
	players    [TrashPlayerCnt]TrashPlayer
	current    int
	phase      TrashPhase
	pending    *Card
	moveCount  int
	winner     int
	actionLog  []*ActionLogEntry
}

// NewTrash コンストラクタ
func NewTrash(trumpCards *TrumpCards) *Trash {
	return &Trash{trumpCards: trumpCards, winner: -1}
}

// NewDefaultTrash 標準1デッキ + 2ジョーカーのトラッシュを返す
func NewDefaultTrash() *Trash {
	return NewTrash(NewTrumpCards(TrashJokerCnt))
}

// Reset ゲーム初期化
func (t *Trash) Reset() {
	t.trumpCards.Shuffle()
	t.phase = TrashPhasePlayerTurn
	t.moveCount = 0
	t.winner = -1
	t.pending = nil
	t.actionLog = nil
	t.discard = nil
	t.current = TrashHumanIdx

	for i := range t.players {
		for j := range t.players[i].Slots {
			t.players[i].Slots[j] = TrashSlot{Card: t.trumpCards.DrawCard(), FaceUp: false}
		}
	}
	t.players[TrashHumanIdx].IsCpu = false
	t.players[TrashCpuIdx].IsCpu = true

	t.stock = t.stock[:0]
	for t.trumpCards.GetRemainingCount() > 0 {
		t.stock = append(t.stock, t.trumpCards.DrawCard())
	}
}

// Draw 山札から1枚引いて連鎖を解決する
func (t *Trash) Draw() error {
	if t.phase != TrashPhasePlayerTurn {
		return errors.New("not in player turn phase")
	}
	if t.pending != nil {
		return errors.New("pending card must be resolved first")
	}
	if len(t.stock) == 0 {
		t.refillStock()
	}
	if len(t.stock) == 0 {
		return errors.New("no cards available in stock or discard")
	}
	t.pending = t.stock[0]
	t.stock = t.stock[1:]
	t.moveCount++
	t.appendLog("draw", fmt.Sprintf("プレイヤー%dがカードを引いた", t.current), []*Card{t.pending})
	t.resolveChain()
	return nil
}

// PlaceWild ワイルドカードを指定位置に配置する (1〜10)
func (t *Trash) PlaceWild(pos int) error {
	if t.phase != TrashPhaseAwaitWild {
		return errors.New("not awaiting wild placement")
	}
	if pos < 1 || pos > TrashSlotCnt {
		return errors.New("invalid position")
	}
	idx := pos - 1
	if t.players[t.current].Slots[idx].FaceUp {
		return errors.New("slot already filled")
	}
	t.placeWildAt(idx)
	if t.phase == TrashPhasePlayerTurn {
		t.resolveChain()
	}
	return nil
}

// placeWildAt is the shared core of explicit wild placement (PlaceWild) and
// the single-empty-slot auto-placement triggered from resolveChain (issue
// #1565). It assumes idx is in range and refers to a face-down slot — both
// callers validate before invoking. The phase is left at PlayerTurn on a
// non-winning placement so the caller decides whether to keep resolving the
// chain (auto path stays in the loop; explicit path re-enters resolveChain).
func (t *Trash) placeWildAt(idx int) {
	p := &t.players[t.current]
	wild := t.pending
	old := p.Slots[idx].Card
	p.Slots[idx].Card = wild
	p.Slots[idx].FaceUp = true
	t.pending = old
	t.appendLog("placeWild", fmt.Sprintf("プレイヤー%dが位置%dにワイルドを配置", t.current, idx+1), []*Card{wild})

	if t.isWin(t.current) {
		t.winner = t.current
		t.phase = TrashPhaseGameOver
		if t.pending != nil {
			t.discard = append(t.discard, t.pending)
			t.pending = nil
		}
		t.appendLog("win", fmt.Sprintf("プレイヤー%dの勝利", t.current), nil)
		return
	}
	t.phase = TrashPhasePlayerTurn
}

// onlyEmptySlot returns the slot index (0..TrashSlotCnt-1) when the player
// at playerIdx has exactly one face-down slot left, or -1 if there are zero
// or more than one. Used to gate auto-placement of wild cards (issue #1565):
// with no empty slot we should never have reached this code path, and with
// two or more we want the human to keep their tactical choice.
//
// Unexported and only called from resolveChain with t.current, which the
// game state machine guarantees to be in [0, TrashPlayerCnt) — no bounds
// guard needed (PR #1584 review).
func (t *Trash) onlyEmptySlot(playerIdx int) int {
	found := -1
	for i, s := range t.players[playerIdx].Slots {
		if s.FaceUp {
			continue
		}
		if found != -1 {
			return -1
		}
		found = i
	}
	return found
}

// IsCpuTurn 現在のターンがCPUか
func (t *Trash) IsCpuTurn() bool {
	if t.phase == TrashPhaseGameOver {
		return false
	}
	return t.players[t.current].IsCpu
}

// CpuStep CPUのターンを1ステップ進める
func (t *Trash) CpuStep() error {
	if !t.IsCpuTurn() {
		return errors.New("not cpu turn")
	}
	switch t.phase {
	case TrashPhasePlayerTurn:
		return t.Draw()
	case TrashPhaseAwaitWild:
		p := &t.players[t.current]
		for i := TrashSlotCnt - 1; i >= 0; i-- {
			if !p.Slots[i].FaceUp {
				return t.PlaceWild(i + 1)
			}
		}
		return errors.New("no face-down slot available")
	default:
		return errors.New("game is over")
	}
}

// --- Getters ---

// GetPhase フェーズ取得
func (t *Trash) GetPhase() TrashPhase { return t.phase }

// SetPhase フェーズ設定 (テスト用)
func (t *Trash) SetPhase(phase TrashPhase) { t.phase = phase }

// GetCurrent 現在ターンのプレイヤーインデックス
func (t *Trash) GetCurrent() int { return t.current }

// SetCurrent 現在プレイヤー設定 (テスト用)
func (t *Trash) SetCurrent(idx int) { t.current = idx }

// GetMoveCount ドロー回数 (両プレイヤー合算)
func (t *Trash) GetMoveCount() int { return t.moveCount }

// GetStockSize 山札残り枚数
func (t *Trash) GetStockSize() int { return len(t.stock) }

// GetDiscardTop 捨て札の一番上
func (t *Trash) GetDiscardTop() *Card {
	return discardTop(t.discard)
}

// GetDiscardSize 捨て札の枚数
func (t *Trash) GetDiscardSize() int { return len(t.discard) }

// GetPending 直前のドロー/拾い上げカード (連鎖中)
func (t *Trash) GetPending() *Card { return t.pending }

// SetPending pendingカード設定 (テスト用)
func (t *Trash) SetPending(c *Card) { t.pending = c }

// GetPlayerSlots プレイヤーのスロット一覧
func (t *Trash) GetPlayerSlots(idx int) []TrashSlot {
	if idx < 0 || idx >= TrashPlayerCnt {
		return nil
	}
	out := make([]TrashSlot, TrashSlotCnt)
	copy(out, t.players[idx].Slots[:])
	return out
}

// SetPlayerSlots プレイヤースロット設定 (テスト用)
func (t *Trash) SetPlayerSlots(idx int, slots [TrashSlotCnt]TrashSlot) {
	if idx < 0 || idx >= TrashPlayerCnt {
		return
	}
	t.players[idx].Slots = slots
}

// SetStock 山札設定 (テスト用)
func (t *Trash) SetStock(cards []*Card) {
	t.stock = append(t.stock[:0], cards...)
}

// SetDiscard 捨て札設定 (テスト用)
func (t *Trash) SetDiscard(cards []*Card) {
	t.discard = append(t.discard[:0], cards...)
}

// IsCpuPlayer プレイヤーがCPUか
func (t *Trash) IsCpuPlayer(idx int) bool {
	if idx < 0 || idx >= TrashPlayerCnt {
		return false
	}
	return t.players[idx].IsCpu
}

// GetWinner 勝者インデックス (-1 なら未決着)
func (t *Trash) GetWinner() int { return t.winner }

// GetActionLog 棋譜取得
func (t *Trash) GetActionLog() []*ActionLogEntry { return t.actionLog }

// --- Private helpers ---

// resolveChain pendingカードを順次解決する。連鎖中にAwaitWild/GameOver/EndTurnのいずれかに到達したら終了。
//
// Auto-placement (issue #1565): when a wild card surfaces and the current
// player has exactly one face-down slot left, the wild is placed directly
// instead of transitioning to TrashPhaseAwaitWild. This removes a forced
// click for an outcome that has no meaningful choice — the only legal
// destination — and matches the tempo Trash relies on. With multiple
// empty slots the human still chooses; CPUs continue to auto-place via
// CpuStep regardless of count.
func (t *Trash) resolveChain() {
	for t.pending != nil && t.phase == TrashPhasePlayerTurn {
		c := t.pending
		switch {
		case isTrashWild(c):
			if idx := t.onlyEmptySlot(t.current); idx >= 0 {
				t.placeWildAt(idx)
				continue
			}
			t.phase = TrashPhaseAwaitWild
			return
		case isTrashEndTurn(c):
			t.endTurn()
			return
		default:
			pos := trashCardPosition(c)
			if pos == 0 {
				// 想定外のカード: 安全に終了
				t.endTurn()
				return
			}
			p := &t.players[t.current]
			idx := pos - 1
			if p.Slots[idx].FaceUp {
				// 既に表向き → 入れ替えできずターン終了
				t.endTurn()
				return
			}
			old := p.Slots[idx].Card
			p.Slots[idx].Card = c
			p.Slots[idx].FaceUp = true
			t.pending = old
			t.appendLog("place", fmt.Sprintf("プレイヤー%dが位置%dにカードを配置", t.current, pos), []*Card{c})
			if t.isWin(t.current) {
				t.winner = t.current
				t.phase = TrashPhaseGameOver
				if t.pending != nil {
					t.discard = append(t.discard, t.pending)
					t.pending = nil
				}
				t.appendLog("win", fmt.Sprintf("プレイヤー%dの勝利", t.current), nil)
				return
			}
		}
	}
}

// endTurn ターン終了処理
func (t *Trash) endTurn() {
	if t.pending != nil {
		t.discard = append(t.discard, t.pending)
		t.appendLog("end", fmt.Sprintf("プレイヤー%dのターン終了", t.current), []*Card{t.pending})
		t.pending = nil
	} else {
		t.appendLog("end", fmt.Sprintf("プレイヤー%dのターン終了", t.current), nil)
	}
	t.current = (t.current + 1) % TrashPlayerCnt
	t.phase = TrashPhasePlayerTurn
}

// refillStock 山札が空になったら捨て札 (一番上を残す) をシャッフルして山札に戻す
func (t *Trash) refillStock() {
	if len(t.discard) <= 1 {
		return
	}
	top := t.discard[len(t.discard)-1]
	rest := make([]*Card, len(t.discard)-1)
	copy(rest, t.discard[:len(t.discard)-1])
	rand.Shuffle(len(rest), func(i, j int) { rest[i], rest[j] = rest[j], rest[i] })
	t.stock = append(t.stock, rest...)
	t.discard = []*Card{top}
}

// isWin 全スロット表向きなら勝利
func (t *Trash) isWin(idx int) bool {
	for _, s := range t.players[idx].Slots {
		if !s.FaceUp {
			return false
		}
	}
	return true
}

// appendLog 棋譜エントリを追加
func (t *Trash) appendLog(actionType, detail string, cards []*Card) {
	t.actionLog = append(t.actionLog, &ActionLogEntry{
		TurnNumber: t.moveCount,
		PlayerIdx:  t.current,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// isTrashWild ワイルド (キング または ジョーカー) か
func isTrashWild(c *Card) bool {
	if c == nil {
		return false
	}
	if c.GetDesign() == CardDesignJoker {
		return true
	}
	return c.GetValue() == 13
}

// isTrashEndTurn ターン終了カード (J/Q) か
func isTrashEndTurn(c *Card) bool {
	if c == nil || c.GetDesign() == CardDesignJoker {
		return false
	}
	v := c.GetValue()
	return v == 11 || v == 12
}

// trashCardPosition 1〜10のランクなら対応するスロット位置を返す
func trashCardPosition(c *Card) int {
	if c == nil || c.GetDesign() == CardDesignJoker {
		return 0
	}
	v := c.GetValue()
	if v >= 1 && v <= TrashSlotCnt {
		return v
	}
	return 0
}

// --- JSON ---

// trashJSON is the JSON wire format for Trash.
type trashJSON struct {
	TrumpCards *TrumpCards                   `json:"tc"`
	Stock      []*Card                       `json:"st"`
	Discard    []*Card                       `json:"ds"`
	Players    [TrashPlayerCnt]trashPlayerJS `json:"pl"`
	Current    int                           `json:"cu"`
	Phase      TrashPhase                    `json:"ph"`
	Pending    *Card                         `json:"pe"`
	MoveCount  int                           `json:"mc"`
	Winner     int                           `json:"wn"`
	ActionLog  []*ActionLogEntry             `json:"al"`
}

type trashPlayerJS struct {
	Slots [TrashSlotCnt]trashSlotJS `json:"sl"`
	IsCpu bool                      `json:"ic"`
}

type trashSlotJS struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// MarshalJSON implements json.Marshaler.
func (t *Trash) MarshalJSON() ([]byte, error) {
	var j trashJSON
	j.TrumpCards = t.trumpCards
	j.Stock = t.stock
	j.Discard = t.discard
	for i, p := range t.players {
		j.Players[i].IsCpu = p.IsCpu
		for k, s := range p.Slots {
			j.Players[i].Slots[k] = trashSlotJS(s)
		}
	}
	j.Current = t.current
	j.Phase = t.phase
	j.Pending = t.pending
	j.MoveCount = t.moveCount
	j.Winner = t.winner
	j.ActionLog = t.actionLog
	return json.Marshal(j)
}

// trashMaxSliceLen caps slice sizes during deserialisation. The deck has 54
// cards; stock/discard/log can each hold at most that many entries plus a small
// margin. 200 gives a comfortable bound while preventing hostile payloads from
// allocating unbounded memory.
const trashMaxSliceLen = 200

// UnmarshalJSON implements json.Unmarshaler.
func (t *Trash) UnmarshalJSON(data []byte) error {
	var j trashJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > trashMaxSliceLen || len(j.Discard) > trashMaxSliceLen || len(j.ActionLog) > trashMaxSliceLen {
		return fmt.Errorf("trash: input array exceeds maximum allowed size")
	}
	t.trumpCards = j.TrumpCards
	if t.trumpCards == nil {
		t.trumpCards = NewTrumpCards(TrashJokerCnt)
	}
	t.stock = j.Stock
	if t.stock == nil {
		t.stock = make([]*Card, 0)
	}
	t.discard = j.Discard
	if t.discard == nil {
		t.discard = make([]*Card, 0)
	}
	for i, p := range j.Players {
		t.players[i].IsCpu = p.IsCpu
		for k, s := range p.Slots {
			t.players[i].Slots[k] = TrashSlot(s)
		}
	}
	t.current = j.Current
	t.phase = j.Phase
	t.pending = j.Pending
	t.moveCount = j.MoveCount
	t.winner = j.Winner
	if t.phase != TrashPhaseGameOver {
		// Authoritative server state: winner is only meaningful after GameOver.
		// Reject any value carried in the payload for earlier phases so a hostile
		// or malformed snapshot cannot restore into a "we have a winner while
		// still playing" state.
		t.winner = -1
	}
	t.actionLog = j.ActionLog
	if t.actionLog == nil {
		t.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
