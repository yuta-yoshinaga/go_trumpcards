//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// AnacondaPlayer はアナコンダ (Anaconda) のプレイヤー。人間 (seat 0) と CPU が同じ型を共有する。
// ポット系のチップ・ゲームであり、チップを保持しラウンドをまたいで賭ける。
type AnacondaPlayer struct {
	*GamePlayer
	ChipHolder
	folded    bool // このラウンドのロールベッティングでフォールドしたか
	out       bool // チップ不足でゲームから脱落したか
	roundBet  int  // このラウンドで支払った累計額 (アンティ + ベット; 表示用)
	streetBet int  // 現在のベッティングラウンド (ストリート) での拠出額
}

// NewAnacondaPlayer はコンストラクタ。初期チップを付与する。
func NewAnacondaPlayer(isHuman bool, startingChips int) *AnacondaPlayer {
	p := &AnacondaPlayer{GamePlayer: NewGamePlayer(isHuman)}
	p.SetChips(startingChips)
	return p
}

// GetFolded フォールド状態を取得。
func (p *AnacondaPlayer) GetFolded() bool { return p.folded }

// SetFolded フォールド状態を設定。
func (p *AnacondaPlayer) SetFolded(v bool) { p.folded = v }

// GetOut 脱落状態を取得。
func (p *AnacondaPlayer) GetOut() bool { return p.out }

// SetOut 脱落状態を設定。
func (p *AnacondaPlayer) SetOut(v bool) { p.out = v }

// GetRoundBet このラウンドの累計支払い額を取得。
func (p *AnacondaPlayer) GetRoundBet() int { return p.roundBet }

// SetRoundBet このラウンドの累計支払い額を設定。
func (p *AnacondaPlayer) SetRoundBet(v int) { p.roundBet = v }

// AddRoundBet このラウンドの累計支払い額に加算。
func (p *AnacondaPlayer) AddRoundBet(v int) { p.roundBet += v }

// GetStreetBet 現在のベッティングラウンドの拠出額を取得。
func (p *AnacondaPlayer) GetStreetBet() int { return p.streetBet }

// SetStreetBet 現在のベッティングラウンドの拠出額を設定。
func (p *AnacondaPlayer) SetStreetBet(v int) { p.streetBet = v }

// AddStreetBet 現在のベッティングラウンドの拠出額に加算。
func (p *AnacondaPlayer) AddStreetBet(v int) { p.streetBet += v }

// ClearHand 手札をクリアする (ディール準備・テスト用)。
func (p *AnacondaPlayer) ClearHand() { p.Reset() }

// ResetForRound はラウンド単位の状態をリセットする (手札・フォールド・累計賭け額)。
func (p *AnacondaPlayer) ResetForRound() {
	p.folded = false
	p.roundBet = 0
	p.streetBet = 0
	p.Reset()
}

// anacondaPlayerJSON is the JSON wire format for AnacondaPlayer.
type anacondaPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Chips      int         `json:"ch"`
	Folded     bool        `json:"fd"`
	Out        bool        `json:"ot"`
	RoundBet   int         `json:"rb"`
	StreetBet  int         `json:"sb"`
}

// MarshalJSON implements json.Marshaler.
func (p *AnacondaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(anacondaPlayerJSON{
		GamePlayer: p.GamePlayer,
		Chips:      p.GetChips(),
		Folded:     p.folded,
		Out:        p.out,
		RoundBet:   p.roundBet,
		StreetBet:  p.streetBet,
	})
}

// UnmarshalJSON implements json.Unmarshaler. 非負のチップ・賭け額を検証する。
func (p *AnacondaPlayer) UnmarshalJSON(data []byte) error {
	var j anacondaPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Chips < 0 || j.RoundBet < 0 || j.StreetBet < 0 {
		return errAnacondaInvalidPlayer
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.SetChips(j.Chips)
	p.folded = j.Folded
	p.out = j.Out
	p.roundBet = j.RoundBet
	p.streetBet = j.StreetBet
	return nil
}
