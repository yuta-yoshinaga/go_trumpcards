//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ClockSolitairePhase クロックソリティアゲームフェーズ
type ClockSolitairePhase int

// ClockSolitaireのフェーズ定数
const (
	// ClockSolitairePhasePlaying プレイ中
	ClockSolitairePhasePlaying ClockSolitairePhase = iota
	// ClockSolitairePhaseGameClear ゲームクリア
	ClockSolitairePhaseGameClear
	// ClockSolitairePhaseGameOver ゲームオーバー
	ClockSolitairePhaseGameOver
)

// ClockSolitairePileCount パイル数（時計の12位置＋中央）
const ClockSolitairePileCount = 13

// ClockSolitaireCardsPerPile 各パイルのカード枚数
const ClockSolitaireCardsPerPile = 4

// ClockSolitaireKingPileIdx 中央（K）パイルのインデックス
const ClockSolitaireKingPileIdx = 12

// ClockSolitaireCard パイル上のカード
type ClockSolitaireCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// ClockSolitaireMaxUndoDepth caps the undo history so the serialised state
// stays bounded. A full game places at most 52 cards, so this ceiling is never
// hit in practice; it exists purely as a safety bound on the KV payload.
const ClockSolitaireMaxUndoDepth = 50

// ClockSolitaire クロックソリティアゲームクラス
type ClockSolitaire struct {
	trumpCards  *TrumpCards
	piles       [ClockSolitairePileCount][]*ClockSolitaireCard
	faceUpCount [ClockSolitairePileCount]int
	currentCard *Card
	phase       ClockSolitairePhase
	stepCount   int
	actionLog   []*ActionLogEntry
	history     []*clockSolitaireSnapshot
}

// clockSolitaireSnapshot アンドゥ用スナップショット
type clockSolitaireSnapshot struct {
	piles       [ClockSolitairePileCount][]*ClockSolitaireCard
	faceUpCount [ClockSolitairePileCount]int
	currentCard *Card
	phase       ClockSolitairePhase
	stepCount   int
}

// NewClockSolitaire コンストラクタ
func NewClockSolitaire(trumpCards *TrumpCards) *ClockSolitaire {
	return &ClockSolitaire{
		trumpCards: trumpCards,
	}
}

// NewDefaultClockSolitaire returns ClockSolitaire with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultClockSolitaire() *ClockSolitaire {
	return NewClockSolitaire(NewTrumpCards(0))
}

// Reset ゲームリセット
func (cs *ClockSolitaire) Reset() {
	cs.trumpCards.Shuffle()
	cs.phase = ClockSolitairePhasePlaying
	cs.stepCount = 0
	cs.actionLog = nil
	cs.currentCard = nil
	cs.history = nil

	// 13パイルに4枚ずつ裏向きで配る
	for i := range ClockSolitairePileCount {
		cs.piles[i] = make([]*ClockSolitaireCard, 0, ClockSolitaireCardsPerPile)
		cs.faceUpCount[i] = 0
		for range ClockSolitaireCardsPerPile {
			card := cs.trumpCards.DrawCard()
			cs.piles[i] = append(cs.piles[i], &ClockSolitaireCard{
				Card:   card,
				FaceUp: false,
			})
		}
	}

	// 中央パイルの一番上をめくって開始
	cs.flipTopFaceDown(ClockSolitaireKingPileIdx)
}

// Step 1ステップ実行
func (cs *ClockSolitaire) Step() error {
	if cs.phase != ClockSolitairePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if cs.currentCard == nil {
		return errors.New("no current card")
	}

	// Snapshot the pre-move state so this step can be undone (#3121).
	cs.takeSnapshot()

	// カードの値から配置先パイルを決定（A=1→pile0, ..., Q=12→pile11, K=13→pile12）
	destIdx := cs.currentCard.GetValue() - 1

	// 配置先パイルの末尾に表向きで追加
	cs.piles[destIdx] = append(cs.piles[destIdx], &ClockSolitaireCard{
		Card:   cs.currentCard,
		FaceUp: true,
	})
	cs.faceUpCount[destIdx]++

	cs.stepCount++
	cs.appendLog("step", fmt.Sprintf("カードを%d番パイルに配置", destIdx+1),
		[]*Card{cs.currentCard})

	cs.currentCard = nil

	// ゲームクリア判定: 時計の12位置すべてが4枚表向き
	if cs.checkGameClear() {
		cs.phase = ClockSolitairePhaseGameClear
		return nil
	}

	// ゲームオーバー判定: 中央パイルのKが4枚表向き
	if cs.faceUpCount[ClockSolitaireKingPileIdx] >= ClockSolitaireCardsPerPile {
		cs.phase = ClockSolitairePhaseGameOver
		return nil
	}

	// 配置先パイルの一番上の裏向きカードをめくる
	cs.flipTopFaceDown(destIdx)

	return nil
}

// AutoPlay 自動プレイ（全ステップ実行）
func (cs *ClockSolitaire) AutoPlay() error {
	if cs.phase != ClockSolitairePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	const maxSteps = 1000
	for i := range maxSteps {
		if cs.phase != ClockSolitairePhasePlaying {
			return nil
		}
		if err := cs.Step(); err != nil {
			return fmt.Errorf("step %d: %w", i+1, err)
		}
	}
	return nil
}

// Undo 直前のステップを取り消す。ゲーム終了後でも巻き戻せる（スナップショットが
// プレイ中フェーズを復元するため）。
func (cs *ClockSolitaire) Undo() error {
	if len(cs.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := cs.history[len(cs.history)-1]
	cs.history = cs.history[:len(cs.history)-1]
	cs.restoreSnapshot(snap)
	cs.appendLog("undo", "1手戻す", nil)
	return nil
}

// CanUndo アンドゥ可能か
func (cs *ClockSolitaire) CanUndo() bool {
	return len(cs.history) > 0
}

// takeSnapshot 現在の可変状態をアンドゥ履歴に積む。履歴は
// ClockSolitaireMaxUndoDepth 件で頭打ちにする。
func (cs *ClockSolitaire) takeSnapshot() {
	snap := &clockSolitaireSnapshot{
		faceUpCount: cs.faceUpCount,
		currentCard: cs.currentCard,
		phase:       cs.phase,
		stepCount:   cs.stepCount,
	}
	for i := range ClockSolitairePileCount {
		snap.piles[i] = make([]*ClockSolitaireCard, len(cs.piles[i]))
		for j, pc := range cs.piles[i] {
			snap.piles[i][j] = &ClockSolitaireCard{Card: pc.Card, FaceUp: pc.FaceUp}
		}
	}
	cs.history = append(cs.history, snap)
	if len(cs.history) > ClockSolitaireMaxUndoDepth {
		cs.history = cs.history[len(cs.history)-ClockSolitaireMaxUndoDepth:]
	}
}

// restoreSnapshot スナップショットの状態を復元する。actionLog と history は
// 復元対象外（アンドゥ操作自体を棋譜に残すため）。
func (cs *ClockSolitaire) restoreSnapshot(snap *clockSolitaireSnapshot) {
	cs.piles = snap.piles
	cs.faceUpCount = snap.faceUpCount
	cs.currentCard = snap.currentCard
	cs.phase = snap.phase
	cs.stepCount = snap.stepCount
}

// checkGameClear 全12時計位置がクリアかチェック
func (cs *ClockSolitaire) checkGameClear() bool {
	for i := range ClockSolitaireKingPileIdx {
		if cs.faceUpCount[i] < ClockSolitaireCardsPerPile {
			return false
		}
	}
	return true
}

// flipTopFaceDown 指定パイルの一番上の裏向きカードをパイルから取り出してcurrentCardにする
func (cs *ClockSolitaire) flipTopFaceDown(pileIdx int) {
	pile := cs.piles[pileIdx]
	for i := len(pile) - 1; i >= 0; i-- {
		if !pile[i].FaceUp {
			cs.currentCard = pile[i].Card
			cs.piles[pileIdx] = append(pile[:i], pile[i+1:]...)
			return
		}
	}
}

// appendLog 棋譜エントリを追加
func (cs *ClockSolitaire) appendLog(actionType, detail string, cards []*Card) {
	cs.actionLog = append(cs.actionLog, &ActionLogEntry{
		TurnNumber: cs.stepCount,
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// GetPhase フェーズ取得
func (cs *ClockSolitaire) GetPhase() ClockSolitairePhase {
	return cs.phase
}

// GetPiles パイル取得
func (cs *ClockSolitaire) GetPiles() [ClockSolitairePileCount][]*ClockSolitaireCard {
	return cs.piles
}

// GetFaceUpCount 表向き枚数取得
func (cs *ClockSolitaire) GetFaceUpCount() [ClockSolitairePileCount]int {
	return cs.faceUpCount
}

// GetStepCount ステップ数取得
func (cs *ClockSolitaire) GetStepCount() int {
	return cs.stepCount
}

// GetCurrentCard 現在のカード取得
func (cs *ClockSolitaire) GetCurrentCard() *Card {
	return cs.currentCard
}

// GetActionLog 棋譜取得
func (cs *ClockSolitaire) GetActionLog() []*ActionLogEntry {
	return cs.actionLog
}

// GetGameEndFlag returns true once the game has left the playing phase.
func (cs *ClockSolitaire) GetGameEndFlag() bool { return cs.phase != ClockSolitairePhasePlaying }

// clockSolitaireJSON is the JSON wire format for ClockSolitaire.
type clockSolitaireJSON struct {
	TrumpCards  *TrumpCards                                    `json:"tc"`
	Piles       [ClockSolitairePileCount][]*ClockSolitaireCard `json:"pi"`
	FaceUpCount [ClockSolitairePileCount]int                   `json:"fu"`
	CurrentCard *Card                                          `json:"cc"`
	Phase       ClockSolitairePhase                            `json:"ps"`
	StepCount   int                                            `json:"sc"`
	ActionLog   []*ActionLogEntry                              `json:"al"`
	History     []*clockSolitaireSnapshot                      `json:"hi,omitempty"`
}

// clockSolitaireSnapshotJSON is the wire format for a single undo snapshot.
// clockSolitaireSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods.
type clockSolitaireSnapshotJSON struct {
	Piles       [ClockSolitairePileCount][]*ClockSolitaireCard `json:"pi"`
	FaceUpCount [ClockSolitairePileCount]int                   `json:"fu"`
	CurrentCard *Card                                          `json:"cc"`
	Phase       ClockSolitairePhase                            `json:"ps"`
	StepCount   int                                            `json:"sc"`
}

// MarshalJSON implements json.Marshaler for clockSolitaireSnapshot.
func (s *clockSolitaireSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(clockSolitaireSnapshotJSON{
		Piles:       s.piles,
		FaceUpCount: s.faceUpCount,
		CurrentCard: s.currentCard,
		Phase:       s.phase,
		StepCount:   s.stepCount,
	})
}

// UnmarshalJSON implements json.Unmarshaler for clockSolitaireSnapshot.
func (s *clockSolitaireSnapshot) UnmarshalJSON(data []byte) error {
	var j clockSolitaireSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for i := range ClockSolitairePileCount {
		if len(j.Piles[i]) > clockSolitaireMaxSliceLen {
			return fmt.Errorf("clocksolitaire: snapshot pile %d exceeds maximum allowed size", i)
		}
	}
	s.piles = j.Piles
	for i := range ClockSolitairePileCount {
		if s.piles[i] == nil {
			s.piles[i] = make([]*ClockSolitaireCard, 0)
		}
	}
	s.faceUpCount = j.FaceUpCount
	s.currentCard = j.CurrentCard
	s.phase = j.Phase
	s.stepCount = j.StepCount
	return nil
}

// MarshalJSON implements json.Marshaler.
func (cs *ClockSolitaire) MarshalJSON() ([]byte, error) {
	return json.Marshal(clockSolitaireJSON{
		TrumpCards:  cs.trumpCards,
		Piles:       cs.piles,
		FaceUpCount: cs.faceUpCount,
		CurrentCard: cs.currentCard,
		Phase:       cs.phase,
		StepCount:   cs.stepCount,
		ActionLog:   cs.actionLog,
		History:     cs.history,
	})
}

// clockSolitaireMaxSliceLen caps slice sizes during deserialisation.
const clockSolitaireMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (cs *ClockSolitaire) UnmarshalJSON(data []byte) error {
	var j clockSolitaireJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > clockSolitaireMaxSliceLen || len(j.History) > clockSolitaireMaxSliceLen {
		return fmt.Errorf("clocksolitaire: input array exceeds maximum allowed size")
	}
	for i := range ClockSolitairePileCount {
		if len(j.Piles[i]) > clockSolitaireMaxSliceLen {
			return fmt.Errorf("clocksolitaire: pile %d exceeds maximum allowed size", i)
		}
	}

	cs.trumpCards = j.TrumpCards
	if cs.trumpCards == nil {
		cs.trumpCards = NewTrumpCards(0)
	}
	cs.piles = j.Piles
	for i := range ClockSolitairePileCount {
		if cs.piles[i] == nil {
			cs.piles[i] = make([]*ClockSolitaireCard, 0)
		}
	}
	cs.faceUpCount = j.FaceUpCount
	cs.currentCard = j.CurrentCard
	cs.phase = j.Phase
	cs.stepCount = j.StepCount
	cs.actionLog = j.ActionLog
	if cs.actionLog == nil {
		cs.actionLog = make([]*ActionLogEntry, 0)
	}
	cs.history = j.History
	if cs.history == nil {
		cs.history = make([]*clockSolitaireSnapshot, 0)
	}
	return nil
}
