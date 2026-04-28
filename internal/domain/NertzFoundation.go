package domain

import (
	"encoding/json"
	"errors"
)

// NertzFoundationMax 1ファウンデーションが完成する枚数 (A..K = 13)
const NertzFoundationMax = CardValueMax

// NertzFoundation Nertz の中央共有ファウンデーション。
// 任意のプレイヤーがスート一致 + 値+1 で積めるが、各カードの貢献者
// (ContributorDeckIdx) を保持してラウンド終了時の得点計算に用いる。
//
// 空のファウンデーションは A のみ受け付け、最初に A を置いたカードのスートが
// そのファウンデーションのスートとして固定される。
type NertzFoundation struct {
	cards        []*Card
	contributors []int // cards[i] を置いたプレイヤーの DeckIdx
	suit         int   // -1 if empty
}

// NewNertzFoundation コンストラクタ
func NewNertzFoundation() *NertzFoundation {
	return &NertzFoundation{suit: -1}
}

// IsEmpty 空かどうか
func (f *NertzFoundation) IsEmpty() bool { return len(f.cards) == 0 }

// IsComplete A..K まで完成しているか
func (f *NertzFoundation) IsComplete() bool { return len(f.cards) >= NertzFoundationMax }

// Size 現在の枚数
func (f *NertzFoundation) Size() int { return len(f.cards) }

// Top トップ (最後に積まれたカード)
func (f *NertzFoundation) Top() *Card {
	if len(f.cards) == 0 {
		return nil
	}
	return f.cards[len(f.cards)-1]
}

// Suit ファウンデーションのスート (空の場合 -1)
func (f *NertzFoundation) Suit() int { return f.suit }

// ContributorAt i 番目のカードの貢献者 DeckIdx
func (f *NertzFoundation) ContributorAt(i int) int {
	if i < 0 || i >= len(f.contributors) {
		return -1
	}
	return f.contributors[i]
}

// CanAccept 指定カードを受け入れ可能か。
// 空: A のみ。非空: 同じスート + 値が top+1 で完成前。
// ジョーカーは常に拒否。
func (f *NertzFoundation) CanAccept(card *Card) bool {
	if card == nil || card.GetDesign() == CardDesignJoker {
		return false
	}
	if f.IsComplete() {
		return false
	}
	if f.IsEmpty() {
		return card.GetValue() == 1
	}
	top := f.Top()
	return card.GetDesign() == f.suit && card.GetValue() == top.GetValue()+1
}

// Push カードを積む。受け入れ不能なら error を返し状態は変化しない。
func (f *NertzFoundation) Push(card *Card, contributorDeckIdx int) error {
	if !f.CanAccept(card) {
		return errors.New("nertz foundation: card not acceptable")
	}
	if f.IsEmpty() {
		f.suit = card.GetDesign()
	}
	f.cards = append(f.cards, card)
	f.contributors = append(f.contributors, contributorDeckIdx)
	return nil
}

// CountByContributor 指定 DeckIdx のプレイヤーがこのファウンデーションに置いた枚数。
func (f *NertzFoundation) CountByContributor(deckIdx int) int {
	n := 0
	for _, c := range f.contributors {
		if c == deckIdx {
			n++
		}
	}
	return n
}

// GetCards カードのスナップショット (末尾が Top)
func (f *NertzFoundation) GetCards() []*Card {
	out := make([]*Card, len(f.cards))
	copy(out, f.cards)
	return out
}

// GetContributors 貢献者一覧のスナップショット
func (f *NertzFoundation) GetContributors() []int {
	out := make([]int, len(f.contributors))
	copy(out, f.contributors)
	return out
}

// nertzFoundationJSON is the JSON wire format for NertzFoundation.
type nertzFoundationJSON struct {
	Cards        []*Card `json:"c"`
	Contributors []int   `json:"o"`
	Suit         int     `json:"s"`
}

// MarshalJSON implements json.Marshaler.
func (f *NertzFoundation) MarshalJSON() ([]byte, error) {
	return json.Marshal(nertzFoundationJSON{
		Cards:        f.cards,
		Contributors: f.contributors,
		Suit:         f.suit,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (f *NertzFoundation) UnmarshalJSON(data []byte) error {
	var j nertzFoundationJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	f.cards = j.Cards
	f.contributors = j.Contributors
	f.suit = j.Suit
	return nil
}
