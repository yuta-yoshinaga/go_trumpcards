package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// cardMemoryEntry 記憶したカード1枚分の情報
type cardMemoryEntry struct {
	value    int // カードの値 (1-13)
	turnSeen int // 記憶したターン番号
}

// GetTurnSeen MemoryEntryインターフェース実装
func (e cardMemoryEntry) GetTurnSeen() int { return e.turnSeen }

// cardMemoryEntryJSON is the wire format for cardMemoryEntry. Fields on the
// underlying struct are unexported so we need an explicit marshaller.
type cardMemoryEntryJSON struct {
	Value    int `json:"v"`
	TurnSeen int `json:"t"`
}

// MarshalJSON implements json.Marshaler.
func (e cardMemoryEntry) MarshalJSON() ([]byte, error) {
	return json.Marshal(cardMemoryEntryJSON{
		Value:    e.value,
		TurnSeen: e.turnSeen,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *cardMemoryEntry) UnmarshalJSON(data []byte) error {
	var j cardMemoryEntryJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	e.value = j.Value
	e.turnSeen = j.TurnSeen
	return nil
}

// DoubtPlayer ダウトプレイヤークラス
type DoubtPlayer struct {
	*GamePlayer
	memoryManager[cardMemoryEntry]
}

// doubtPlayerMaxMemoryLen caps the deserialised memory slice length so a
// hostile session blob cannot allocate unbounded memory.
const doubtPlayerMaxMemoryLen = 1000

// doubtPlayerJSON is the JSON wire format for DoubtPlayer. Persisting
// cardMemories keeps Hard-difficulty CPUs from forgetting every revealed
// value when a session is restored (see ADR-0028 / issue #1655).
type doubtPlayerJSON struct {
	GamePlayer   *GamePlayer       `json:"gp"`
	CardMemories []cardMemoryEntry `json:"cm,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (p *DoubtPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(doubtPlayerJSON{
		GamePlayer:   p.GamePlayer,
		CardMemories: p.cardMemories,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *DoubtPlayer) UnmarshalJSON(data []byte) error {
	var j doubtPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.CardMemories) > doubtPlayerMaxMemoryLen {
		return fmt.Errorf("doubtPlayer: input array exceeds maximum allowed size")
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.cardMemories = j.CardMemories
	return nil
}

// NewDoubtPlayer コンストラクタ
func NewDoubtPlayer(isHuman bool) *DoubtPlayer {
	return &DoubtPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// RecordRevealedCard 公開されたカードを記憶する (retentionChance の確率で記録)
func (p *DoubtPlayer) RecordRevealedCard(value int, retentionChance float64, turnNumber int) {
	if rand.Float64() < retentionChance {
		p.AddMemory(cardMemoryEntry{value: value, turnSeen: turnNumber})
	}
}

// CountKnownCards 指定した値のカードを何枚知っているか返す (記憶 + 手札)
func (p *DoubtPlayer) CountKnownCards(value int) int {
	count := 0
	for _, entry := range p.cardMemories {
		if entry.value == value {
			count++
		}
	}
	for _, card := range p.cards {
		if card.GetValue() == value {
			count++
		}
	}
	return count
}
