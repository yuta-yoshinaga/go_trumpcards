//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// SevenTwentySevenHandSize は各プレイヤーの手札枚数 (セブン・トゥエンティセブンは 2 枚)。
const SevenTwentySevenHandSize = 2

// SevenTwentySevenPlayer はセブン・トゥエンティセブン (SevenTwentySeven) のプレイヤー。人間 (seat 0) と CPU が同じ型を共有する。
// カジノ系のポット・ゲームであり、チップを保持しラウンドをまたいで賭ける。
type SevenTwentySevenPlayer struct {
	*GamePlayer
	ChipHolder
	// standing は「もう引かない」と宣言済みか。**一度立てたら戻らない。**
	// 止まった人にはそのラウンド二度と配られない。
	standing bool
	out      bool // チップ不足でゲームから脱落したか
	roundBet int  // このラウンドで支払った累計額 (アンティ + マッチ; 表示用)
}

// NewSevenTwentySevenPlayer はコンストラクタ。初期チップを付与する。
func NewSevenTwentySevenPlayer(isHuman bool, startingChips int) *SevenTwentySevenPlayer {
	p := &SevenTwentySevenPlayer{GamePlayer: NewGamePlayer(isHuman)}
	p.SetChips(startingChips)
	return p
}

// GetStanding は「もう引かない」と宣言済みかを返す。
func (p *SevenTwentySevenPlayer) GetStanding() bool { return p.standing }

// SetStanding は「もう引かない」状態を設定する。
func (p *SevenTwentySevenPlayer) SetStanding(v bool) { p.standing = v }

// GetOut 脱落状態を取得。
func (p *SevenTwentySevenPlayer) GetOut() bool { return p.out }

// SetOut 脱落状態を設定。
func (p *SevenTwentySevenPlayer) SetOut(v bool) { p.out = v }

// GetRoundBet このラウンドの累計支払い額を取得。
func (p *SevenTwentySevenPlayer) GetRoundBet() int { return p.roundBet }

// SetRoundBet このラウンドの累計支払い額を設定。
func (p *SevenTwentySevenPlayer) SetRoundBet(v int) { p.roundBet = v }

// AddRoundBet このラウンドの累計支払い額に加算。
func (p *SevenTwentySevenPlayer) AddRoundBet(v int) { p.roundBet += v }

// ClearHand 手札をクリアする (ディール準備・テスト用)。
func (p *SevenTwentySevenPlayer) ClearHand() { p.Reset() }

// ResetForRound はラウンド単位の状態をリセットする (手札・イン宣言・累計賭け額)。
func (p *SevenTwentySevenPlayer) ResetForRound() {
	p.standing = false
	p.roundBet = 0
	p.Reset()
}

// sevenTwentySevenPlayerJSON is the JSON wire format for SevenTwentySevenPlayer.
type sevenTwentySevenPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Chips      int         `json:"ch"`
	Standing   bool        `json:"st"`
	Out        bool        `json:"ot"`
	RoundBet   int         `json:"rb"`
}

// MarshalJSON implements json.Marshaler.
func (p *SevenTwentySevenPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(sevenTwentySevenPlayerJSON{
		GamePlayer: p.GamePlayer,
		Chips:      p.GetChips(),
		Standing:   p.standing,
		Out:        p.out,
		RoundBet:   p.roundBet,
	})
}

// UnmarshalJSON implements json.Unmarshaler. 非負のチップ・賭け額を検証する。
func (p *SevenTwentySevenPlayer) UnmarshalJSON(data []byte) error {
	var j sevenTwentySevenPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Chips < 0 || j.RoundBet < 0 {
		return errSevenTwentySevenInvalidPlayer
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.SetChips(j.Chips)
	p.standing = j.Standing
	p.out = j.Out
	p.roundBet = j.RoundBet
	return nil
}
