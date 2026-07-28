//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// MichiganPlayer はミシガン (Michigan) のプレイヤー。人間 (seat 0) と CPU が同じ型を共有する。
// チップを保持しラウンドをまたいでブードルに賭ける「ストップ系」パーティゲームのプレイヤー。
type MichiganPlayer struct {
	*GamePlayer
	ChipHolder
	roundBet int // このラウンドでブードルに賭けた累計額 (表示用)
}

// NewMichiganPlayer はコンストラクタ。初期チップを付与する。
func NewMichiganPlayer(isHuman bool, startingChips int) *MichiganPlayer {
	p := &MichiganPlayer{GamePlayer: NewGamePlayer(isHuman)}
	p.SetChips(startingChips)
	return p
}

// GetRoundBet このラウンドの累計賭け額を取得。
func (p *MichiganPlayer) GetRoundBet() int { return p.roundBet }

// SetRoundBet このラウンドの累計賭け額を設定。
func (p *MichiganPlayer) SetRoundBet(v int) { p.roundBet = v }

// AddRoundBet このラウンドの累計賭け額に加算。
func (p *MichiganPlayer) AddRoundBet(v int) { p.roundBet += v }

// ClearHand 手札をクリアする (ディール準備・テスト用)。
func (p *MichiganPlayer) ClearHand() { p.Reset() }

// ResetForRound はラウンド単位の状態をリセットする (手札・累計賭け額)。
func (p *MichiganPlayer) ResetForRound() {
	p.roundBet = 0
	p.Reset()
}

// michiganPlayerJSON is the JSON wire format for MichiganPlayer.
type michiganPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Chips      int         `json:"ch"`
	RoundBet   int         `json:"rb"`
}

// MarshalJSON implements json.Marshaler.
func (p *MichiganPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(michiganPlayerJSON{
		GamePlayer: p.GamePlayer,
		Chips:      p.GetChips(),
		RoundBet:   p.roundBet,
	})
}

// UnmarshalJSON implements json.Unmarshaler. 非負のチップ・賭け額を検証する。
func (p *MichiganPlayer) UnmarshalJSON(data []byte) error {
	var j michiganPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Chips < 0 || j.RoundBet < 0 {
		return errMichiganInvalidPlayer
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.SetChips(j.Chips)
	p.roundBet = j.RoundBet
	return nil
}
