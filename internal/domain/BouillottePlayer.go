//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// BouillotteHandSize は各プレイヤーの手札枚数 (ブイヨットは 3 枚)。
const BouillotteHandSize = 3

// BouillottePlayer はブイヨット (Bouillotte) のプレイヤー。人間 (seat 0) と CPU が同じ型を共有する。
// カジノ系のポット・ベッティングゲームであり、チップを保持しラウンドをまたいで賭ける。
type BouillottePlayer struct {
	*GamePlayer
	ChipHolder
	folded   bool // このラウンドで降りた (フォールド) か
	out      bool // チップ不足でゲームから脱落したか
	roundBet int  // このラウンドで支払った累計額 (アンティ + コール/レイズ)
}

// NewBouillottePlayer はコンストラクタ。初期チップを付与する。
func NewBouillottePlayer(isHuman bool, startingChips int) *BouillottePlayer {
	p := &BouillottePlayer{GamePlayer: NewGamePlayer(isHuman)}
	p.SetChips(startingChips)
	return p
}

// GetFolded フォールド状態を取得。
func (p *BouillottePlayer) GetFolded() bool { return p.folded }

// SetFolded フォールド状態を設定。
func (p *BouillottePlayer) SetFolded(v bool) { p.folded = v }

// GetOut 脱落状態を取得。
func (p *BouillottePlayer) GetOut() bool { return p.out }

// SetOut 脱落状態を設定。
func (p *BouillottePlayer) SetOut(v bool) { p.out = v }

// GetRoundBet このラウンドの累計支払い額を取得。
func (p *BouillottePlayer) GetRoundBet() int { return p.roundBet }

// SetRoundBet このラウンドの累計支払い額を設定。
func (p *BouillottePlayer) SetRoundBet(v int) { p.roundBet = v }

// AddRoundBet このラウンドの累計支払い額に加算。
func (p *BouillottePlayer) AddRoundBet(v int) { p.roundBet += v }

// ClearHand 手札をクリアする (ディール準備・テスト用)。
func (p *BouillottePlayer) ClearHand() { p.Reset() }

// ResetForRound はラウンド単位の状態をリセットする (手札・フォールド・累計賭け額)。
func (p *BouillottePlayer) ResetForRound() {
	p.folded = false
	p.roundBet = 0
	p.Reset()
}

// bouillottePlayerJSON is the JSON wire format for BouillottePlayer.
type bouillottePlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Chips      int         `json:"ch"`
	Folded     bool        `json:"fd"`
	Out        bool        `json:"ot"`
	RoundBet   int         `json:"rb"`
}

// MarshalJSON implements json.Marshaler.
func (p *BouillottePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(bouillottePlayerJSON{
		GamePlayer: p.GamePlayer,
		Chips:      p.GetChips(),
		Folded:     p.folded,
		Out:        p.out,
		RoundBet:   p.roundBet,
	})
}

// UnmarshalJSON implements json.Unmarshaler. 非負のチップ・賭け額を検証する。
func (p *BouillottePlayer) UnmarshalJSON(data []byte) error {
	var j bouillottePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Chips < 0 || j.RoundBet < 0 {
		return errBouillotteInvalidPlayer
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
