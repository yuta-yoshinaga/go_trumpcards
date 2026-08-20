//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// TrenteEtQuarantePlayer はトラント・エ・カラント (Trente et Quarante) のプレイヤー。
// カジノ系のバンキングゲームであり、プレイヤーはチップを保持しラウンドごとに賭ける。
// カードの判断は一切なく、ベット種別とステークのみを選択する。
type TrenteEtQuarantePlayer struct {
	chips        ChipHolder
	roundsPlayed int // これまでに賭けたラウンド数
	roundsWon    int // これまでに勝ったラウンド数
}

// NewTrenteEtQuarantePlayer は TrenteEtQuarantePlayer を構築する。
func NewTrenteEtQuarantePlayer(chips int) *TrenteEtQuarantePlayer {
	p := &TrenteEtQuarantePlayer{}
	p.chips.SetChips(chips)
	return p
}

// GetChips は保有チップ数を返す。
func (p *TrenteEtQuarantePlayer) GetChips() int { return p.chips.GetChips() }

// SetChips は保有チップ数を設定する。
func (p *TrenteEtQuarantePlayer) SetChips(chips int) { p.chips.SetChips(chips) }

// AddChips はチップを加算する (配当)。
func (p *TrenteEtQuarantePlayer) AddChips(amount int) { p.chips.AddChips(amount) }

// SubtractChips はチップを減算する。不足時は false を返す。
func (p *TrenteEtQuarantePlayer) SubtractChips(amount int) bool {
	return p.chips.SubtractChips(amount)
}

// GetRoundsPlayed は賭けたラウンド数を返す。
func (p *TrenteEtQuarantePlayer) GetRoundsPlayed() int { return p.roundsPlayed }

// GetRoundsWon は勝ったラウンド数を返す。
func (p *TrenteEtQuarantePlayer) GetRoundsWon() int { return p.roundsWon }

// RecordRound はラウンド結果を統計に反映する (won=true なら勝ち)。
func (p *TrenteEtQuarantePlayer) RecordRound(won bool) {
	p.roundsPlayed++
	if won {
		p.roundsWon++
	}
}

// ResetStats はラウンド統計を初期化する (新規ゲーム開始時)。
func (p *TrenteEtQuarantePlayer) ResetStats() {
	p.roundsPlayed = 0
	p.roundsWon = 0
}

// trenteEtQuarantePlayerJSON is the JSON wire format for TrenteEtQuarantePlayer.
type trenteEtQuarantePlayerJSON struct {
	Chips        *ChipHolder `json:"ch"`
	RoundsPlayed int         `json:"rp"`
	RoundsWon    int         `json:"rw"`
}

// MarshalJSON implements json.Marshaler.
func (p *TrenteEtQuarantePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(trenteEtQuarantePlayerJSON{
		Chips:        &p.chips,
		RoundsPlayed: p.roundsPlayed,
		RoundsWon:    p.roundsWon,
	})
}

// UnmarshalJSON implements json.Unmarshaler. 非負の統計値を検証する。
func (p *TrenteEtQuarantePlayer) UnmarshalJSON(data []byte) error {
	var j trenteEtQuarantePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.RoundsPlayed < 0 || j.RoundsWon < 0 || j.RoundsWon > j.RoundsPlayed {
		return errTrenteEtQuaranteInvalidStats
	}
	if j.Chips != nil {
		p.chips = *j.Chips
	}
	if p.chips.GetChips() < 0 {
		return errTrenteEtQuaranteNegativeChips
	}
	p.roundsPlayed = j.RoundsPlayed
	p.roundsWon = j.RoundsWon
	return nil
}
