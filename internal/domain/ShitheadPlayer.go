package domain

import "encoding/json"

// ShitheadPlayer シットヘッドのプレイヤー
// 3層構造のカードを保持する: 裏向きの場札 (faceDown) / 表向きの場札 (faceUp) / 手札 (hand=cards)
// 上位の Player.cards を「手札」として使う。
type ShitheadPlayer struct {
	*GamePlayer
	// faceDown 裏向きの場札 (3枚から減っていく)。プレイ時は中身を見ずに選ぶ。
	faceDown []*Card
	// faceUp 表向きの場札 (3枚から減っていく)。手札・山札を使い切ってから出す。
	faceUp []*Card
	// rank 上がり順位 (-1=未確定, 1..N)
	rank int
}

// NewShitheadPlayer constructs a ShitheadPlayer with empty stacks.
func NewShitheadPlayer(isHuman bool) *ShitheadPlayer {
	return &ShitheadPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		faceDown:   make([]*Card, 0, 3),
		faceUp:     make([]*Card, 0, 3),
		rank:       -1,
	}
}

// AddFaceDown 裏向きの場札を追加
func (p *ShitheadPlayer) AddFaceDown(c *Card) { p.faceDown = append(p.faceDown, c) }

// AddFaceUp 表向きの場札を追加
func (p *ShitheadPlayer) AddFaceUp(c *Card) { p.faceUp = append(p.faceUp, c) }

// GetFaceDownSize 裏向きの場札の枚数
func (p *ShitheadPlayer) GetFaceDownSize() int { return len(p.faceDown) }

// GetFaceUpSize 表向きの場札の枚数
func (p *ShitheadPlayer) GetFaceUpSize() int { return len(p.faceUp) }

// GetFaceDownCard 指定インデックスの裏向きの場札を返す (範囲外は nil)
func (p *ShitheadPlayer) GetFaceDownCard(idx int) *Card {
	if idx < 0 || idx >= len(p.faceDown) {
		return nil
	}
	return p.faceDown[idx]
}

// GetFaceUpCard 指定インデックスの表向きの場札を返す (範囲外は nil)
func (p *ShitheadPlayer) GetFaceUpCard(idx int) *Card {
	if idx < 0 || idx >= len(p.faceUp) {
		return nil
	}
	return p.faceUp[idx]
}

// GetFaceDownCards 裏向き場札のコピー (公開用は枚数のみだが、内部処理用に提供)
func (p *ShitheadPlayer) GetFaceDownCards() []*Card {
	out := make([]*Card, len(p.faceDown))
	copy(out, p.faceDown)
	return out
}

// GetFaceUpCards 表向き場札のコピー
func (p *ShitheadPlayer) GetFaceUpCards() []*Card {
	out := make([]*Card, len(p.faceUp))
	copy(out, p.faceUp)
	return out
}

// RemoveFaceDownCard 指定インデックスの裏向きの場札を取り出して返す。範囲外は nil。
func (p *ShitheadPlayer) RemoveFaceDownCard(idx int) *Card {
	if idx < 0 || idx >= len(p.faceDown) {
		return nil
	}
	c := p.faceDown[idx]
	p.faceDown = append(p.faceDown[:idx], p.faceDown[idx+1:]...)
	return c
}

// RemoveFaceUpCard 指定インデックスの表向きの場札を取り出して返す。範囲外は nil。
func (p *ShitheadPlayer) RemoveFaceUpCard(idx int) *Card {
	if idx < 0 || idx >= len(p.faceUp) {
		return nil
	}
	c := p.faceUp[idx]
	p.faceUp = append(p.faceUp[:idx], p.faceUp[idx+1:]...)
	return c
}

// GetRank ランク取得 (-1 = 未確定)
func (p *ShitheadPlayer) GetRank() int { return p.rank }

// SetRank ランク設定
func (p *ShitheadPlayer) SetRank(r int) { p.rank = r }

// Reset 手札・場札・ランクを初期化する
func (p *ShitheadPlayer) Reset() {
	p.GamePlayer.Reset()
	p.faceDown = make([]*Card, 0, 3)
	p.faceUp = make([]*Card, 0, 3)
	p.rank = -1
}

// HasHandCards 手札 (Player.cards) が残っているか
func (p *ShitheadPlayer) HasHandCards() bool { return p.GetCardsSize() > 0 }

// HasAnyCards 手札・表場札・裏場札のいずれかにカードが残っているか
func (p *ShitheadPlayer) HasAnyCards() bool {
	return p.GetCardsSize() > 0 || len(p.faceUp) > 0 || len(p.faceDown) > 0
}

// shitheadPlayerJSON is the JSON wire format for ShitheadPlayer.
type shitheadPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	FaceDown   []*Card     `json:"fd"`
	FaceUp     []*Card     `json:"fu"`
	Rank       int         `json:"r"`
}

// MarshalJSON implements json.Marshaler.
func (p *ShitheadPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(shitheadPlayerJSON{
		GamePlayer: p.GamePlayer,
		FaceDown:   p.faceDown,
		FaceUp:     p.faceUp,
		Rank:       p.rank,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ShitheadPlayer) UnmarshalJSON(data []byte) error {
	var j shitheadPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.faceDown = j.FaceDown
	if p.faceDown == nil {
		p.faceDown = make([]*Card, 0)
	}
	p.faceUp = j.FaceUp
	if p.faceUp == nil {
		p.faceUp = make([]*Card, 0)
	}
	p.rank = j.Rank
	return nil
}
