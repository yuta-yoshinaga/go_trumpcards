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

// ClockSolitaire クロックソリティアゲームクラス
type ClockSolitaire struct {
	trumpCards  *TrumpCards
	piles       [ClockSolitairePileCount][]*ClockSolitaireCard
	faceUpCount [ClockSolitairePileCount]int
	currentCard *Card
	phase       ClockSolitairePhase
	stepCount   int
	actionLog   []*ActionLogEntry
}

// NewClockSolitaire コンストラクタ
func NewClockSolitaire(trumpCards *TrumpCards) *ClockSolitaire {
	return &ClockSolitaire{
		trumpCards: trumpCards,
	}
}

// Reset ゲームリセット
func (cs *ClockSolitaire) Reset() {
	cs.trumpCards.Shuffle()
	cs.phase = ClockSolitairePhasePlaying
	cs.stepCount = 0
	cs.actionLog = nil
	cs.currentCard = nil

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

// clockSolitaireJSON is the JSON wire format for ClockSolitaire.
type clockSolitaireJSON struct {
	TrumpCards  *TrumpCards                                    `json:"tc"`
	Piles       [ClockSolitairePileCount][]*ClockSolitaireCard `json:"pi"`
	FaceUpCount [ClockSolitairePileCount]int                   `json:"fu"`
	CurrentCard *Card                                          `json:"cc"`
	Phase       ClockSolitairePhase                            `json:"ps"`
	StepCount   int                                            `json:"sc"`
	ActionLog   []*ActionLogEntry                              `json:"al"`
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
	if len(j.ActionLog) > clockSolitaireMaxSliceLen {
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
	return nil
}
