//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// SpoonsPlayer はスプーンのプレイヤー。
//
// 手札 (GamePlayer 由来の cards) に加えて、溜まった文字数 (letters) と
// 脱落フラグ (eliminated)、当該ラウンドでスプーンを取得済みかどうか
// (hasSpoon) を保持する。eliminated になったプレイヤーは以降の配札・
// パス・グラブから除外される。
type SpoonsPlayer struct {
	*GamePlayer
	letters    int
	eliminated bool
	hasSpoon   bool
}

// NewSpoonsPlayer はコンストラクタ。
func NewSpoonsPlayer(isHuman bool) *SpoonsPlayer {
	return &SpoonsPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// GetLetters は溜まった文字数を返す。
func (p *SpoonsPlayer) GetLetters() int { return p.letters }

// SetLetters は文字数を設定する (主にテスト/復元用)。
func (p *SpoonsPlayer) SetLetters(n int) { p.letters = n }

// AddLetter は文字を 1 つ加える。SpoonsMaxLetters に達したら脱落させ true を返す。
func (p *SpoonsPlayer) AddLetter() bool {
	p.letters++
	if p.letters >= SpoonsMaxLetters {
		p.letters = SpoonsMaxLetters
		p.eliminated = true
		p.SetIsFinished(true)
		return true
	}
	return false
}

// GetEliminated は脱落済みかどうかを返す。
func (p *SpoonsPlayer) GetEliminated() bool { return p.eliminated }

// SetEliminated は脱落状態を設定する (主にテスト/復元用)。
func (p *SpoonsPlayer) SetEliminated(v bool) {
	p.eliminated = v
	p.SetIsFinished(v)
}

// GetHasSpoon は当該ラウンドでスプーンを取得済みかを返す。
func (p *SpoonsPlayer) GetHasSpoon() bool { return p.hasSpoon }

// SetHasSpoon はスプーン取得状態を設定する。
func (p *SpoonsPlayer) SetHasSpoon(v bool) { p.hasSpoon = v }

// HasFourOfAKind は手札 4 枚がすべて同じランクかどうかを返す。
// 手札が SpoonsHandSize 枚未満のときは false。
func (p *SpoonsPlayer) HasFourOfAKind() bool {
	if p.GetCardsSize() < SpoonsHandSize {
		return false
	}
	first := p.GetCard(0)
	if first == nil {
		return false
	}
	v := first.GetValue()
	count := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c != nil && c.GetValue() == v {
			count++
		}
	}
	return count >= SpoonsHandSize
}

// spoonsPlayerJSON is the JSON wire format for SpoonsPlayer.
type spoonsPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Letters    int         `json:"lt"`
	Eliminated bool        `json:"el"`
	HasSpoon   bool        `json:"hs"`
}

// MarshalJSON implements json.Marshaler.
func (p *SpoonsPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(spoonsPlayerJSON{
		GamePlayer: p.GamePlayer,
		Letters:    p.letters,
		Eliminated: p.eliminated,
		HasSpoon:   p.hasSpoon,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SpoonsPlayer) UnmarshalJSON(data []byte) error {
	var j spoonsPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	// 文字数はレンジ内にクランプして不正な復元値を防ぐ。
	if j.Letters < 0 {
		j.Letters = 0
	}
	if j.Letters > SpoonsMaxLetters {
		j.Letters = SpoonsMaxLetters
	}
	p.letters = j.Letters
	p.eliminated = j.Eliminated
	p.hasSpoon = j.HasSpoon
	return nil
}
