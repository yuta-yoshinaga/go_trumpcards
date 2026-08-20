//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// GutsHandSize は各プレイヤーの手札枚数 (ガッツは 2 枚)。
const GutsHandSize = 2

// GutsPlayer はガッツ (Guts) のプレイヤー。人間 (seat 0) と CPU が同じ型を共有する。
// カジノ系のポット・ゲームであり、チップを保持しラウンドをまたいで賭ける。
type GutsPlayer struct {
	*GamePlayer
	ChipHolder
	in       bool // このラウンドで「イン (勝負に残る)」を宣言したか
	out      bool // チップ不足でゲームから脱落したか
	roundBet int  // このラウンドで支払った累計額 (アンティ + マッチ; 表示用)
}

// NewGutsPlayer はコンストラクタ。初期チップを付与する。
func NewGutsPlayer(isHuman bool, startingChips int) *GutsPlayer {
	p := &GutsPlayer{GamePlayer: NewGamePlayer(isHuman)}
	p.SetChips(startingChips)
	return p
}

// GetIn イン宣言状態を取得。
func (p *GutsPlayer) GetIn() bool { return p.in }

// SetIn イン宣言状態を設定。
func (p *GutsPlayer) SetIn(v bool) { p.in = v }

// GetOut 脱落状態を取得。
func (p *GutsPlayer) GetOut() bool { return p.out }

// SetOut 脱落状態を設定。
func (p *GutsPlayer) SetOut(v bool) { p.out = v }

// GetRoundBet このラウンドの累計支払い額を取得。
func (p *GutsPlayer) GetRoundBet() int { return p.roundBet }

// SetRoundBet このラウンドの累計支払い額を設定。
func (p *GutsPlayer) SetRoundBet(v int) { p.roundBet = v }

// AddRoundBet このラウンドの累計支払い額に加算。
func (p *GutsPlayer) AddRoundBet(v int) { p.roundBet += v }

// ClearHand 手札をクリアする (ディール準備・テスト用)。
func (p *GutsPlayer) ClearHand() { p.Reset() }

// ResetForRound はラウンド単位の状態をリセットする (手札・イン宣言・累計賭け額)。
func (p *GutsPlayer) ResetForRound() {
	p.in = false
	p.roundBet = 0
	p.Reset()
}

// gutsPlayerJSON is the JSON wire format for GutsPlayer.
type gutsPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Chips      int         `json:"ch"`
	In         bool        `json:"in"`
	Out        bool        `json:"ot"`
	RoundBet   int         `json:"rb"`
}

// MarshalJSON implements json.Marshaler.
func (p *GutsPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(gutsPlayerJSON{
		GamePlayer: p.GamePlayer,
		Chips:      p.GetChips(),
		In:         p.in,
		Out:        p.out,
		RoundBet:   p.roundBet,
	})
}

// UnmarshalJSON implements json.Unmarshaler. 非負のチップ・賭け額を検証する。
func (p *GutsPlayer) UnmarshalJSON(data []byte) error {
	var j gutsPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Chips < 0 || j.RoundBet < 0 {
		return errGutsInvalidPlayer
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.SetChips(j.Chips)
	p.in = j.In
	p.out = j.Out
	p.roundBet = j.RoundBet
	return nil
}
