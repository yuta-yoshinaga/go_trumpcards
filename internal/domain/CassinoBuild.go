package domain

import "encoding/json"

// CassinoBuild は場の上に作られたビルド (宣言値で後からまとめて取るコミット)。
//
// 単独ビルド: Groups に 1 グループだけ。同宣言値のカードを別途手札に持つこと。
// 複合ビルド (IsMulti=true): 同じ宣言値のグループを複数束ねた状態。
// 例) 宣言値 8 のビルドに、さらに 5+3 を重ねて "8 のペアビルド" にできる。
// 複合になったビルドは値を変更できず、宣言値 8 の手札でしか捕獲できない。
type CassinoBuild struct {
	OwnerIdx int       // このビルドのオーナー (値 8 の手札を持つと宣言したプレイヤー)
	Value    int       // 宣言値 (ビルドを捕獲するために必要な手札値)
	Groups   [][]*Card // Groups[i] は合計が Value になる部分集合。1 要素 = 単独ビルド、2+ = 複合
	IsMulti  bool      // 複合ビルド (ペアビルド) かどうか
}

// NewCassinoBuild は単独ビルドを作成する。Groups は 1 グループから始まる。
func NewCassinoBuild(owner, value int, cards []*Card) *CassinoBuild {
	return &CassinoBuild{
		OwnerIdx: owner,
		Value:    value,
		Groups:   [][]*Card{append([]*Card(nil), cards...)},
		IsMulti:  false,
	}
}

// AllCards はビルド内の全カードをフラットに返す。
func (b *CassinoBuild) AllCards() []*Card {
	if b == nil {
		return nil
	}
	total := 0
	for _, g := range b.Groups {
		total += len(g)
	}
	out := make([]*Card, 0, total)
	for _, g := range b.Groups {
		out = append(out, g...)
	}
	return out
}

// AddGroup は既存ビルドに新しいグループを足して複合ビルドにする。
// cards の合計値は Value と一致している必要がある (呼び出し側で検証する)。
func (b *CassinoBuild) AddGroup(cards []*Card) {
	if b == nil || len(cards) == 0 {
		return
	}
	b.Groups = append(b.Groups, append([]*Card(nil), cards...))
	if len(b.Groups) >= 2 {
		b.IsMulti = true
	}
}

// cassinoBuildJSON is the JSON wire format for CassinoBuild.
type cassinoBuildJSON struct {
	OwnerIdx int       `json:"o"`
	Value    int       `json:"v"`
	Groups   [][]*Card `json:"g"`
	IsMulti  bool      `json:"m"`
}

// MarshalJSON implements json.Marshaler.
func (b *CassinoBuild) MarshalJSON() ([]byte, error) {
	return json.Marshal(cassinoBuildJSON{
		OwnerIdx: b.OwnerIdx,
		Value:    b.Value,
		Groups:   b.Groups,
		IsMulti:  b.IsMulti,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *CassinoBuild) UnmarshalJSON(data []byte) error {
	var j cassinoBuildJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	b.OwnerIdx = j.OwnerIdx
	b.Value = j.Value
	b.Groups = j.Groups
	b.IsMulti = j.IsMulti
	return nil
}
