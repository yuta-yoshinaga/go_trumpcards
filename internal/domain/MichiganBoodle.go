//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// MichiganBoodle は中央に置かれる賭け対象の「ブードル (boodle)」札。プレイヤーは
// ベッティングフェーズでチップをブードルに積み、プレイフェーズでブードルと同じ
// スート・ランクのカードを出したプレイヤーがそのチップを総取りする。チップは
// 獲得されるまでラウンドをまたいで持ち越される。
type MichiganBoodle struct {
	card      *Card // ブードル札 (固定: A♥ / K♣ / Q♦ / J♠ のいずれか)
	chips     int   // 現在ブードルに積まれているチップ (獲得されるまで持ち越し)
	claimedBy int   // このラウンドでブードルを獲得した席 (-1 = 未獲得)
}

// GetCard はブードル札を返す。
func (b *MichiganBoodle) GetCard() *Card { return b.card }

// GetChips は現在積まれているチップを返す。
func (b *MichiganBoodle) GetChips() int { return b.chips }

// GetClaimedBy はこのラウンドの獲得者の席を返す (-1 = 未獲得)。
func (b *MichiganBoodle) GetClaimedBy() int { return b.claimedBy }

// SetClaimedBy は獲得者の席を設定する (テスト用。-1 = 未獲得)。
func (b *MichiganBoodle) SetClaimedBy(seat int) { b.claimedBy = seat }

// michiganBoodleJSON is the JSON wire format for MichiganBoodle.
type michiganBoodleJSON struct {
	Card      *Card `json:"c"`
	Chips     int   `json:"ch"`
	ClaimedBy int   `json:"cb"`
}

// MarshalJSON implements json.Marshaler.
func (b *MichiganBoodle) MarshalJSON() ([]byte, error) {
	return json.Marshal(michiganBoodleJSON{
		Card:      b.card,
		Chips:     b.chips,
		ClaimedBy: b.claimedBy,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *MichiganBoodle) UnmarshalJSON(data []byte) error {
	var j michiganBoodleJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	b.card = j.Card
	b.chips = j.Chips
	b.claimedBy = j.ClaimedBy
	return nil
}

// michiganBoodleCards は 4 つのブードル札 (A♥, K♣, Q♦, J♠) を固定順で返す。
func michiganBoodleCards() []*Card {
	return []*Card{
		NewCard(CardDesignHeart, 1, false),    // A♥
		NewCard(CardDesignClover, 13, false),  // K♣
		NewCard(CardDesignDiamond, 12, false), // Q♦
		NewCard(CardDesignSpade, 11, false),   // J♠
	}
}

// newMichiganBoodles は空 (チップ 0・未獲得) の 4 つのブードルを生成する。
func newMichiganBoodles() []*MichiganBoodle {
	cards := michiganBoodleCards()
	boodles := make([]*MichiganBoodle, len(cards))
	for i, c := range cards {
		boodles[i] = &MichiganBoodle{card: c, chips: 0, claimedBy: -1}
	}
	return boodles
}
