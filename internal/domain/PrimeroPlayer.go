//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// PrimeroHandSize は各プレイヤーの手札枚数 (プリメロは 4 枚)。
const PrimeroHandSize = 4

// PrimeroPlayer はプリメロ (Primero) のプレイヤー。人間 (seat 0) と CPU が同じ型を共有する。
// ルネサンス期のヴァイイング系ポットゲームであり、チップを保持しラウンドをまたいで賭ける。
type PrimeroPlayer struct {
	*GamePlayer
	ChipHolder
	folded   bool // このラウンドで降りた (フォールド) か
	out      bool // チップ不足でゲームから脱落したか
	roundBet int  // このラウンドで支払った累計額 (アンティ + コール/レイズ)
}

// NewPrimeroPlayer はコンストラクタ。初期チップを付与する。
func NewPrimeroPlayer(isHuman bool, startingChips int) *PrimeroPlayer {
	p := &PrimeroPlayer{GamePlayer: NewGamePlayer(isHuman)}
	p.SetChips(startingChips)
	return p
}

// GetFolded フォールド状態を取得。
func (p *PrimeroPlayer) GetFolded() bool { return p.folded }

// SetFolded フォールド状態を設定。
func (p *PrimeroPlayer) SetFolded(v bool) { p.folded = v }

// GetOut 脱落状態を取得。
func (p *PrimeroPlayer) GetOut() bool { return p.out }

// SetOut 脱落状態を設定。
func (p *PrimeroPlayer) SetOut(v bool) { p.out = v }

// GetRoundBet このラウンドの累計支払い額を取得。
func (p *PrimeroPlayer) GetRoundBet() int { return p.roundBet }

// SetRoundBet このラウンドの累計支払い額を設定。
func (p *PrimeroPlayer) SetRoundBet(v int) { p.roundBet = v }

// AddRoundBet このラウンドの累計支払い額に加算。
func (p *PrimeroPlayer) AddRoundBet(v int) { p.roundBet += v }

// ClearHand 手札をクリアする (ディール準備・テスト用)。
func (p *PrimeroPlayer) ClearHand() { p.Reset() }

// ResetForRound はラウンド単位の状態をリセットする (手札・フォールド・累計賭け額)。
func (p *PrimeroPlayer) ResetForRound() {
	p.folded = false
	p.roundBet = 0
	p.Reset()
}

// primeroPlayerJSON is the JSON wire format for PrimeroPlayer.
type primeroPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Chips      int         `json:"ch"`
	Folded     bool        `json:"fd"`
	Out        bool        `json:"ot"`
	RoundBet   int         `json:"rb"`
}

// MarshalJSON implements json.Marshaler.
func (p *PrimeroPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(primeroPlayerJSON{
		GamePlayer: p.GamePlayer,
		Chips:      p.GetChips(),
		Folded:     p.folded,
		Out:        p.out,
		RoundBet:   p.roundBet,
	})
}

// UnmarshalJSON implements json.Unmarshaler. 非負のチップ・賭け額を検証する。
func (p *PrimeroPlayer) UnmarshalJSON(data []byte) error {
	var j primeroPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Chips < 0 || j.RoundBet < 0 {
		return errPrimeroInvalidPlayer
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
	return nil
}
