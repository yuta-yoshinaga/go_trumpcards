package domain

import "encoding/json"

// TrickHolder トリック管理の共通構造体
type TrickHolder struct {
	tricksTaken [][]*Card // 獲得したトリック
}

// GetTricksTaken 獲得したトリック一覧を取得
func (h *TrickHolder) GetTricksTaken() [][]*Card { return h.tricksTaken }

// GetTrickCount 獲得したトリック数を取得
func (h *TrickHolder) GetTrickCount() int { return len(h.tricksTaken) }

// AddTrick トリックを追加
func (h *TrickHolder) AddTrick(cards []*Card) {
	h.tricksTaken = append(h.tricksTaken, cards)
}

// ResetTricks トリックをリセット
func (h *TrickHolder) ResetTricks() {
	h.tricksTaken = nil
}

// trickHolderJSON is the JSON wire format for TrickHolder.
type trickHolderJSON struct {
	TricksTaken [][]*Card `json:"tt"` // tricksTaken
}

// MarshalJSON implements json.Marshaler.
func (h *TrickHolder) MarshalJSON() ([]byte, error) {
	return json.Marshal(trickHolderJSON{TricksTaken: h.tricksTaken})
}

// UnmarshalJSON implements json.Unmarshaler.
func (h *TrickHolder) UnmarshalJSON(data []byte) error {
	var j trickHolderJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	h.tricksTaken = j.TricksTaken
	if h.tricksTaken == nil {
		h.tricksTaken = make([][]*Card, 0)
	}
	return nil
}
