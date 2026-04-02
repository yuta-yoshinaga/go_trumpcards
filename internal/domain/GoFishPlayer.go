package domain

import "encoding/json"

// GoFishPlayer Go Fishプレイヤークラス
type GoFishPlayer struct {
	*GamePlayer
	books [][]*Card // 完成したブック (4枚1組)
}

// NewGoFishPlayer コンストラクタ
func NewGoFishPlayer(isHuman bool) *GoFishPlayer {
	return &GoFishPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		books:      make([][]*Card, 0),
	}
}

// GetBooks 完成したブック一覧を取得する
func (p *GoFishPlayer) GetBooks() [][]*Card { return p.books }

// GetBookCount 完成したブック数を取得する
func (p *GoFishPlayer) GetBookCount() int { return len(p.books) }

// AddBook ブックを追加する
func (p *GoFishPlayer) AddBook(cards []*Card) {
	p.books = append(p.books, cards)
}

// ResetBooks ブックをリセットする
func (p *GoFishPlayer) ResetBooks() {
	p.books = make([][]*Card, 0)
}

// HasRank 指定ランクのカードを手札に持っているかを返す
func (p *GoFishPlayer) HasRank(rank int) bool {
	for _, c := range p.cards {
		if c.GetValue() == rank {
			return true
		}
	}
	return false
}

// CountRank 指定ランクのカード枚数を返す
func (p *GoFishPlayer) CountRank(rank int) int {
	cnt := 0
	for _, c := range p.cards {
		if c.GetValue() == rank {
			cnt++
		}
	}
	return cnt
}

// RemoveAllOfRank 指定ランクのカードを全て手札から取り除いて返す
func (p *GoFishPlayer) RemoveAllOfRank(rank int) []*Card {
	removed := make([]*Card, 0)
	kept := make([]*Card, 0, len(p.cards))
	for _, c := range p.cards {
		if c.GetValue() == rank {
			removed = append(removed, c)
		} else {
			kept = append(kept, c)
		}
	}
	p.cards = kept
	return removed
}

// GetDistinctRanks 手札に含まれるユニークなランク一覧を返す
func (p *GoFishPlayer) GetDistinctRanks() []int {
	seen := make(map[int]bool)
	ranks := make([]int, 0)
	for _, c := range p.cards {
		v := c.GetValue()
		if !seen[v] {
			seen[v] = true
			ranks = append(ranks, v)
		}
	}
	return ranks
}

// goFishPlayerJSON is the JSON wire format for GoFishPlayer.
type goFishPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Books      [][]*Card   `json:"bk"`
}

// MarshalJSON implements json.Marshaler.
func (p *GoFishPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(goFishPlayerJSON{
		GamePlayer: p.GamePlayer,
		Books:      p.books,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *GoFishPlayer) UnmarshalJSON(data []byte) error {
	var j goFishPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.books = j.Books
	if p.books == nil {
		p.books = make([][]*Card, 0)
	}
	return nil
}
