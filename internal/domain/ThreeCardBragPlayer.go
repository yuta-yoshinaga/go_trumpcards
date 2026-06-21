//go:build !js || !wasm || casino

package domain

import "encoding/json"

// ThreeCardBragPlayerCnt Three Card Brag プレイヤー数
const ThreeCardBragPlayerCnt = 4

// ThreeCardBragHandSize 各プレイヤーの手札枚数
const ThreeCardBragHandSize = 3

// ThreeCardBragPlayer Three Card Brag プレイヤークラス
type ThreeCardBragPlayer struct {
	*GamePlayer
	ChipHolder
	seen     bool // 手札を見た (Seen) か、見ていない (Blind) か
	folded   bool // このディールで降りたか
	out      bool // チップ切れでゲームから脱落したか
	roundBet int  // このディールでポットに入れた累計額
}

// NewThreeCardBragPlayer コンストラクタ。初期チップを付与する。
func NewThreeCardBragPlayer(isHuman bool, startingChips int) *ThreeCardBragPlayer {
	p := &ThreeCardBragPlayer{GamePlayer: NewGamePlayer(isHuman)}
	p.SetChips(startingChips)
	return p
}

// GetSeen Seen 状態を取得
func (p *ThreeCardBragPlayer) GetSeen() bool { return p.seen }

// SetSeen Seen 状態を設定
func (p *ThreeCardBragPlayer) SetSeen(v bool) { p.seen = v }

// GetFolded フォールド状態を取得
func (p *ThreeCardBragPlayer) GetFolded() bool { return p.folded }

// SetFolded フォールド状態を設定
func (p *ThreeCardBragPlayer) SetFolded(v bool) { p.folded = v }

// GetOut 脱落状態を取得
func (p *ThreeCardBragPlayer) GetOut() bool { return p.out }

// SetOut 脱落状態を設定
func (p *ThreeCardBragPlayer) SetOut(v bool) { p.out = v }

// GetRoundBet このディールの累計賭け額を取得
func (p *ThreeCardBragPlayer) GetRoundBet() int { return p.roundBet }

// SetRoundBet このディールの累計賭け額を設定
func (p *ThreeCardBragPlayer) SetRoundBet(v int) { p.roundBet = v }

// AddRoundBet このディールの累計賭け額に加算
func (p *ThreeCardBragPlayer) AddRoundBet(v int) { p.roundBet += v }

// ResetForDeal ディール単位の状態をリセットする (手札・seen・folded・roundBet)。
func (p *ThreeCardBragPlayer) ResetForDeal() {
	p.seen = false
	p.folded = false
	p.roundBet = 0
	p.Reset()
	p.SetIsFinished(false)
}

// threeCardBragPlayerJSON is the JSON wire format for ThreeCardBragPlayer.
type threeCardBragPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Chips      int         `json:"ch"`
	Seen       bool        `json:"sn"`
	Folded     bool        `json:"fd"`
	Out        bool        `json:"ot"`
	RoundBet   int         `json:"rb"`
}

// MarshalJSON implements json.Marshaler.
func (p *ThreeCardBragPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(threeCardBragPlayerJSON{
		GamePlayer: p.GamePlayer,
		Chips:      p.GetChips(),
		Seen:       p.seen,
		Folded:     p.folded,
		Out:        p.out,
		RoundBet:   p.roundBet,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ThreeCardBragPlayer) UnmarshalJSON(data []byte) error {
	var j threeCardBragPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.SetChips(j.Chips)
	p.seen = j.Seen
	p.folded = j.Folded
	p.out = j.Out
	p.roundBet = j.RoundBet
	return nil
}
