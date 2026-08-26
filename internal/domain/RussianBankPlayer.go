//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// RussianBankReserveSize 各プレイヤーのリザーブ (ストック) 初期枚数。
const RussianBankReserveSize = 13

// RussianBankPlayer Russian Bank (Crapette) の 1 プレイヤーが保持する 3 つの山。
//   - reserve: リザーブ (ストック)。トップ (末尾) のみ表向きで参照・移動可能。
//     これを空にしたプレイヤーが勝利する。
//   - hand: 手札 (タロン)。裏向き。合法手が尽きたときトップを waste へ送る。
//   - waste: 廃札。トップ (末尾) のみ移動可能。
type RussianBankPlayer struct {
	name    string
	isCPU   bool
	seat    int
	reserve []*Card
	hand    []*Card
	waste   []*Card
}

// NewRussianBankPlayer プレイヤーを生成する。
func NewRussianBankPlayer(name string, isCPU bool, seat int) *RussianBankPlayer {
	return &RussianBankPlayer{name: name, isCPU: isCPU, seat: seat}
}

// GetName プレイヤー名を返す。
func (p *RussianBankPlayer) GetName() string { return p.name }

// IsCPU CPU プレイヤーなら true。
func (p *RussianBankPlayer) IsCPU() bool { return p.isCPU }

// GetIsHuman 人間プレイヤーなら true (CUI 表示ヘルパ用)。
func (p *RussianBankPlayer) GetIsHuman() bool { return !p.isCPU }

// GetSeat 座席番号を返す。
func (p *RussianBankPlayer) GetSeat() int { return p.seat }

// resetPiles 全ての山を空にする。
func (p *RussianBankPlayer) resetPiles() {
	p.reserve = nil
	p.hand = nil
	p.waste = nil
}

// --- reserve ---

// ReserveSize リザーブの残枚数。
func (p *RussianBankPlayer) ReserveSize() int { return len(p.reserve) }

// ReserveTop リザーブのトップ (末尾) を返す。空なら nil。
func (p *RussianBankPlayer) ReserveTop() *Card { return rbTopCard(p.reserve) }

// GetReserve リザーブのカード列を返す (末尾 = トップ)。
func (p *RussianBankPlayer) GetReserve() []*Card { return p.reserve }

func (p *RussianBankPlayer) pushReserve(c *Card) { p.reserve = append(p.reserve, c) }

func (p *RussianBankPlayer) popReserve() *Card { return rbPopCard(&p.reserve) }

// --- hand ---

// HandSize 手札の残枚数。
func (p *RussianBankPlayer) HandSize() int { return len(p.hand) }

// HandTop 手札のトップ (末尾) を返す。空なら nil。
func (p *RussianBankPlayer) HandTop() *Card { return rbTopCard(p.hand) }

func (p *RussianBankPlayer) pushHand(c *Card) { p.hand = append(p.hand, c) }

func (p *RussianBankPlayer) popHand() *Card { return rbPopCard(&p.hand) }

// --- waste ---

// WasteSize 廃札の残枚数。
func (p *RussianBankPlayer) WasteSize() int { return len(p.waste) }

// WasteTop 廃札のトップ (末尾) を返す。空なら nil。
func (p *RussianBankPlayer) WasteTop() *Card { return rbTopCard(p.waste) }

// GetWaste 廃札のカード列を返す (末尾 = トップ)。
func (p *RussianBankPlayer) GetWaste() []*Card { return p.waste }

func (p *RussianBankPlayer) pushWaste(c *Card) { p.waste = append(p.waste, c) }

func (p *RussianBankPlayer) popWaste() *Card { return rbPopCard(&p.waste) }

// russianBankPlayerJSON は RussianBankPlayer の JSON ワイヤ形式。
type russianBankPlayerJSON struct {
	Name    string  `json:"n"`
	IsCPU   bool    `json:"c"`
	Seat    int     `json:"s"`
	Reserve []*Card `json:"r"`
	Hand    []*Card `json:"h"`
	Waste   []*Card `json:"w"`
}

// MarshalJSON implements json.Marshaler.
func (p *RussianBankPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(russianBankPlayerJSON{
		Name: p.name, IsCPU: p.isCPU, Seat: p.seat,
		Reserve: p.reserve, Hand: p.hand, Waste: p.waste,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *RussianBankPlayer) UnmarshalJSON(data []byte) error {
	var j russianBankPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.name, p.isCPU, p.seat = j.Name, j.IsCPU, j.Seat
	p.reserve, p.hand, p.waste = j.Reserve, j.Hand, j.Waste
	return nil
}

// rbTopCard スライスの末尾要素を返す。空なら nil。
func rbTopCard(cards []*Card) *Card {
	if len(cards) == 0 {
		return nil
	}
	return cards[len(cards)-1]
}

// rbPopCard スライス末尾を取り除いて返す。参照を切って GC を妨げない。
func rbPopCard(cards *[]*Card) *Card {
	s := *cards
	if len(s) == 0 {
		return nil
	}
	c := s[len(s)-1]
	s[len(s)-1] = nil
	*cards = s[:len(s)-1]
	return c
}
