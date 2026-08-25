//go:build !js || !wasm || classic

package domain

import "encoding/json"

// DilotiDeclaration は場に積まれた宣言 (δήλωση)。
//
// **宣言こそがディロティという名前の由来。** ギリシャ語の δηλώνω (宣言する)
// から来ていて、#5458 はこの仕組みに一言も触れていない ── 触れないまま
// 実装すると、残るのはクセリ (Xeri) そのものであってディロティではない。
//
// 単一宣言 (IsGroup=false): Groups が 1 束。相手は札を足して値を上げられる。
// グループ宣言 (IsGroup=true): 同じ値の束が 2 つ以上。**上げられず、値ちょうどの
// 1 枚で丸ごとしか取れない** ── 部分的に取ることはできない。
type DilotiDeclaration struct {
	// OwnerIdx は宣言した席。捕獲の義務を負う。
	OwnerIdx int
	// Value は宣言値。これと同じ値の手札でしか取れない。
	Value int
	// Groups は各束。どの束も合計が Value になる。
	Groups [][]*Card
	// IsGroup はグループ宣言かどうか。
	IsGroup bool
}

// NewDilotiDeclaration は単一宣言を作る。
func NewDilotiDeclaration(owner, value int, cards []*Card) *DilotiDeclaration {
	return &DilotiDeclaration{
		OwnerIdx: owner,
		Value:    value,
		Groups:   [][]*Card{append([]*Card(nil), cards...)},
	}
}

// AllCards は宣言に含まれる全札を平らに返す。
func (d *DilotiDeclaration) AllCards() []*Card {
	if d == nil {
		return nil
	}
	n := 0
	for _, g := range d.Groups {
		n += len(g)
	}
	out := make([]*Card, 0, n)
	for _, g := range d.Groups {
		out = append(out, g...)
	}
	return out
}

// AddGroup は束を足してグループ宣言にする。cards の合計が Value と一致することは
// 呼び出し側が保証する。
func (d *DilotiDeclaration) AddGroup(cards []*Card) {
	if d == nil || len(cards) == 0 {
		return
	}
	d.Groups = append(d.Groups, append([]*Card(nil), cards...))
	if len(d.Groups) >= 2 {
		d.IsGroup = true
	}
}

// dilotiDeclarationJSON は保存用の形。
type dilotiDeclarationJSON struct {
	OwnerIdx int       `json:"o"`
	Value    int       `json:"v"`
	Groups   [][]*Card `json:"g"`
	IsGroup  bool      `json:"m"`
}

// MarshalJSON implements json.Marshaler.
//
// **非公開フィールドしか無い型は `{}` になる。** ここは公開フィールドだけだが、
// キーを短くして保存サイズを抑えるために明示する。
func (d *DilotiDeclaration) MarshalJSON() ([]byte, error) {
	return json.Marshal(dilotiDeclarationJSON{
		OwnerIdx: d.OwnerIdx, Value: d.Value, Groups: d.Groups, IsGroup: d.IsGroup,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *DilotiDeclaration) UnmarshalJSON(data []byte) error {
	var j dilotiDeclarationJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	d.OwnerIdx, d.Value, d.Groups, d.IsGroup = j.OwnerIdx, j.Value, j.Groups, j.IsGroup
	return nil
}
