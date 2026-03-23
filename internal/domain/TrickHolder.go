package domain

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
