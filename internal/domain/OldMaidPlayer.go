package domain

import "encoding/json"

// OldMaidPlayer ババ抜きプレイヤークラス
type OldMaidPlayer struct {
	*GamePlayer
	memLastDrawPos int  // 最後に引いたカードの位置 (-1=なし)
	memGotPair     bool // 最後に引いたカードでペアができたか
}

// oldMaidPlayerJSON is the JSON wire format for OldMaidPlayer.
type oldMaidPlayerJSON struct {
	GamePlayer     *GamePlayer `json:"gp"`
	MemLastDrawPos int         `json:"mp"`
	MemGotPair     bool        `json:"mg"`
}

// MarshalJSON implements json.Marshaler.
func (p *OldMaidPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(oldMaidPlayerJSON{
		GamePlayer:     p.GamePlayer,
		MemLastDrawPos: p.memLastDrawPos,
		MemGotPair:     p.memGotPair,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *OldMaidPlayer) UnmarshalJSON(data []byte) error {
	var j oldMaidPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.memLastDrawPos = j.MemLastDrawPos
	p.memGotPair = j.MemGotPair
	return nil
}

// NewOldMaidPlayer コンストラクタ
func NewOldMaidPlayer(isHuman bool) *OldMaidPlayer {
	return &OldMaidPlayer{
		GamePlayer:     NewGamePlayer(isHuman),
		memLastDrawPos: -1,
	}
}

// GetMemLastDrawPos 最後に引いたカードの位置取得
func (p *OldMaidPlayer) GetMemLastDrawPos() int { return p.memLastDrawPos }

// SetMemLastDrawPos 最後に引いたカードの位置設定
func (p *OldMaidPlayer) SetMemLastDrawPos(pos int) { p.memLastDrawPos = pos }

// GetMemGotPair 最後に引いたカードでペアができたか取得
func (p *OldMaidPlayer) GetMemGotPair() bool { return p.memGotPair }

// SetMemGotPair 最後に引いたカードでペアができたか設定
func (p *OldMaidPlayer) SetMemGotPair(v bool) { p.memGotPair = v }

// ResetDrawMemory 引きの記憶をリセット
func (p *OldMaidPlayer) ResetDrawMemory() {
	p.memLastDrawPos = -1
	p.memGotPair = false
}

// DiscardPairs ペアのカードを捨てる (捨てたカードとペア数を返す)
func (p *OldMaidPlayer) DiscardPairs() ([]*Card, int) {
	discardedCards := make([]*Card, 0)
	pairs := 0
	for {
		found := false
		for i := 0; i < len(p.cards); i++ {
			c1 := p.cards[i]
			if IsJoker(c1) {
				continue
			}
			for j := i + 1; j < len(p.cards); j++ {
				c2 := p.cards[j]
				if IsJoker(c2) {
					continue
				}
				if c1.GetValue() == c2.GetValue() {
					newCards := make([]*Card, 0, len(p.cards)-2)
					for k, c := range p.cards {
						if k != i && k != j {
							newCards = append(newCards, c)
						}
					}
					p.cards = newCards
					discardedCards = append(discardedCards, c1, c2)
					pairs++
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			break
		}
	}
	return discardedCards, pairs
}
