package domain

import "encoding/json"

// SpiteAndMalicePlayer Spite & Malice のプレイヤー
type SpiteAndMalicePlayer struct {
	hand  []*Card
	goal  []*Card // 末尾 (最後の要素) が表向きトップ
	sides [SpiteAndMaliceSideCnt][]*Card
	isCpu bool
}

// NewSpiteAndMalicePlayer コンストラクタ
func NewSpiteAndMalicePlayer(isCpu bool) *SpiteAndMalicePlayer {
	return &SpiteAndMalicePlayer{
		hand:  make([]*Card, 0, SpiteAndMaliceHandMax),
		goal:  make([]*Card, 0, SpiteAndMaliceGoalSizeMax),
		isCpu: isCpu,
	}
}

// Reset プレイヤー状態をリセット
func (p *SpiteAndMalicePlayer) Reset() {
	p.hand = p.hand[:0]
	p.goal = p.goal[:0]
	for i := range p.sides {
		p.sides[i] = nil
	}
}

// GetIsCpu CPU かどうか
func (p *SpiteAndMalicePlayer) GetIsCpu() bool { return p.isCpu }

// SetIsCpu テスト用 CPU フラグ設定
func (p *SpiteAndMalicePlayer) SetIsCpu(v bool) { p.isCpu = v }

// --- Hand ---

// GetHand 手札のスナップショット
func (p *SpiteAndMalicePlayer) GetHand() []*Card {
	out := make([]*Card, len(p.hand))
	copy(out, p.hand)
	return out
}

// HandSize 手札枚数
func (p *SpiteAndMalicePlayer) HandSize() int { return len(p.hand) }

// AddToHand 手札に追加
func (p *SpiteAndMalicePlayer) AddToHand(c *Card) {
	p.hand = append(p.hand, c)
}

// RemoveFromHand 手札から i 番目を取り除く
func (p *SpiteAndMalicePlayer) RemoveFromHand(i int) *Card {
	if i < 0 || i >= len(p.hand) {
		return nil
	}
	c := p.hand[i]
	p.hand = append(p.hand[:i], p.hand[i+1:]...)
	return c
}

// --- Goal ---

// AddToGoal ゴールパイルの底に追加 (Reset 用)
func (p *SpiteAndMalicePlayer) AddToGoal(c *Card) {
	// 新しく積むカードはゴールパイルの「下 (底)」に置かれ、
	// 既に積まれているカードが一番上 (先にめくられる) になる構造。
	// 末尾がトップを表すため先頭へ prepend する。
	p.goal = append([]*Card{c}, p.goal...)
}

// GoalTop ゴールパイルの一番上 (次にプレイされるカード)。末尾がトップ。
func (p *SpiteAndMalicePlayer) GoalTop() *Card {
	if len(p.goal) == 0 {
		return nil
	}
	return p.goal[len(p.goal)-1]
}

// PopGoal ゴールパイルのトップを取り出す
func (p *SpiteAndMalicePlayer) PopGoal() *Card {
	if len(p.goal) == 0 {
		return nil
	}
	c := p.goal[len(p.goal)-1]
	p.goal = p.goal[:len(p.goal)-1]
	return c
}

// GoalSize ゴールパイル残り枚数
func (p *SpiteAndMalicePlayer) GoalSize() int { return len(p.goal) }

// GetGoal ゴールパイルのスナップショット (末尾がトップ)
func (p *SpiteAndMalicePlayer) GetGoal() []*Card {
	out := make([]*Card, len(p.goal))
	copy(out, p.goal)
	return out
}

// --- Side ---

// SideTop サイドパイル i のトップ
func (p *SpiteAndMalicePlayer) SideTop(i int) *Card {
	if i < 0 || i >= SpiteAndMaliceSideCnt {
		return nil
	}
	if len(p.sides[i]) == 0 {
		return nil
	}
	return p.sides[i][len(p.sides[i])-1]
}

// SideSize サイドパイル i の枚数
func (p *SpiteAndMalicePlayer) SideSize(i int) int {
	if i < 0 || i >= SpiteAndMaliceSideCnt {
		return 0
	}
	return len(p.sides[i])
}

// PushSide サイドパイル i にカードを積む
func (p *SpiteAndMalicePlayer) PushSide(i int, c *Card) {
	if i < 0 || i >= SpiteAndMaliceSideCnt {
		return
	}
	p.sides[i] = append(p.sides[i], c)
}

// PopSide サイドパイル i のトップを取り除く
func (p *SpiteAndMalicePlayer) PopSide(i int) *Card {
	if i < 0 || i >= SpiteAndMaliceSideCnt {
		return nil
	}
	if len(p.sides[i]) == 0 {
		return nil
	}
	c := p.sides[i][len(p.sides[i])-1]
	p.sides[i] = p.sides[i][:len(p.sides[i])-1]
	return c
}

// GetSide サイドパイル i のスナップショット (末尾がトップ)
func (p *SpiteAndMalicePlayer) GetSide(i int) []*Card {
	if i < 0 || i >= SpiteAndMaliceSideCnt {
		return nil
	}
	out := make([]*Card, len(p.sides[i]))
	copy(out, p.sides[i])
	return out
}

// --- JSON ---

// spiteAndMalicePlayerJSON is the JSON wire format for SpiteAndMalicePlayer.
type spiteAndMalicePlayerJSON struct {
	Hand  []*Card                        `json:"hd"`
	Goal  []*Card                        `json:"gl"`
	Sides [SpiteAndMaliceSideCnt][]*Card `json:"sd"`
	IsCpu bool                           `json:"ic"`
}

// MarshalJSON implements json.Marshaler.
func (p *SpiteAndMalicePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(spiteAndMalicePlayerJSON{
		Hand:  p.hand,
		Goal:  p.goal,
		Sides: p.sides,
		IsCpu: p.isCpu,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SpiteAndMalicePlayer) UnmarshalJSON(data []byte) error {
	var j spiteAndMalicePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.hand = j.Hand
	if p.hand == nil {
		p.hand = make([]*Card, 0)
	}
	p.goal = j.Goal
	if p.goal == nil {
		p.goal = make([]*Card, 0)
	}
	p.sides = j.Sides
	for i := range p.sides {
		if p.sides[i] == nil {
			p.sides[i] = make([]*Card, 0)
		}
	}
	p.isCpu = j.IsCpu
	return nil
}
