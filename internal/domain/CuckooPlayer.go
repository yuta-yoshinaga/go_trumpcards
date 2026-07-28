//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// CuckooPlayer Cuckoo (カッコー) プレイヤー。
//
// 各プレイヤーは常に手札を 1 枚だけ持つ。ライフは初期値から始まり、
// ラウンド終了時に最低位のカードを持っていると 1 減る。ライフが 0 に
// なったプレイヤーは脱落 (IsEliminated)。
type CuckooPlayer struct {
	*GamePlayer
	lives int
}

// NewCuckooPlayer コンストラクタ
func NewCuckooPlayer(isHuman bool) *CuckooPlayer {
	return &CuckooPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// Card プレイヤーの 1 枚の手札を返す (未配牌なら nil)
func (p *CuckooPlayer) Card() *Card {
	if p.GetCardsSize() == 0 {
		return nil
	}
	return p.GetCard(0)
}

// SetCard プレイヤーの手札を 1 枚に差し替える
func (p *CuckooPlayer) SetCard(c *Card) {
	p.Reset()
	if c != nil {
		p.AddCard(c)
	}
}

// CardValue プレイヤーの手札のランク値 (A=1(最弱) … K=13(最強)) を返す。
// 手札がない場合は 0 を返す。
func (p *CuckooPlayer) CardValue() int {
	c := p.Card()
	if c == nil {
		return 0
	}
	return c.GetValue()
}

// HasKing プレイヤーが King (値 13) を持っているかを返す。
// King を持つ隣人はスワップを拒否 (PlayerRefuse) できる。
func (p *CuckooPlayer) HasKing() bool {
	return p.CardValue() == CuckooKingValue

}

// GetLives 残りライフ数を返す
func (p *CuckooPlayer) GetLives() int { return p.lives }

// SetLives ライフ数を設定する
func (p *CuckooPlayer) SetLives(n int) { p.lives = n }

// LoseLife ライフを 1 減らす
func (p *CuckooPlayer) LoseLife() { p.lives-- }

// IsEliminated 脱落済みか (ライフが 0 以下) を返す
func (p *CuckooPlayer) IsEliminated() bool { return p.lives <= 0 }

// cuckooPlayerJSON is the JSON wire format for CuckooPlayer.
type cuckooPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Lives      int         `json:"lv"`
}

// MarshalJSON implements json.Marshaler.
func (p *CuckooPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(cuckooPlayerJSON{GamePlayer: p.GamePlayer, Lives: p.lives})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *CuckooPlayer) UnmarshalJSON(data []byte) error {
	var j cuckooPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.lives = j.Lives
	return nil
}
