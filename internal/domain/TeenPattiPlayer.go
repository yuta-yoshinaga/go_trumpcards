//go:build !js || !wasm || casino

package domain

import "encoding/json"

// TeenPattiPlayerCnt Three Card Brag プレイヤー数
const TeenPattiPlayerCnt = 4

// TeenPattiHandSize 各プレイヤーの手札枚数
const TeenPattiHandSize = 3

// TeenPattiPlayer Three Card Brag プレイヤークラス
type TeenPattiPlayer struct {
	*GamePlayer
	ChipHolder
	seen     bool // 手札を見た (Seen) か、見ていない (Blind) か
	folded   bool // このディールで降りたか
	out      bool // チップ切れでゲームから脱落したか
	roundBet int  // このディールでポットに入れた累計額
}

// NewTeenPattiPlayer コンストラクタ。初期チップを付与する。
func NewTeenPattiPlayer(isHuman bool, startingChips int) *TeenPattiPlayer {
	p := &TeenPattiPlayer{GamePlayer: NewGamePlayer(isHuman)}
	p.SetChips(startingChips)
	return p
}

// GetSeen Seen 状態を取得
func (p *TeenPattiPlayer) GetSeen() bool { return p.seen }

// SetSeen Seen 状態を設定
func (p *TeenPattiPlayer) SetSeen(v bool) { p.seen = v }

// GetFolded フォールド状態を取得
func (p *TeenPattiPlayer) GetFolded() bool { return p.folded }

// SetFolded フォールド状態を設定
func (p *TeenPattiPlayer) SetFolded(v bool) { p.folded = v }

// GetOut 脱落状態を取得
func (p *TeenPattiPlayer) GetOut() bool { return p.out }

// SetOut 脱落状態を設定
func (p *TeenPattiPlayer) SetOut(v bool) { p.out = v }

// GetRoundBet このディールの累計賭け額を取得
func (p *TeenPattiPlayer) GetRoundBet() int { return p.roundBet }

// SetRoundBet このディールの累計賭け額を設定
func (p *TeenPattiPlayer) SetRoundBet(v int) { p.roundBet = v }

// AddRoundBet このディールの累計賭け額に加算
func (p *TeenPattiPlayer) AddRoundBet(v int) { p.roundBet += v }

// ResetForDeal ディール単位の状態をリセットする (手札・seen・folded・roundBet)。
func (p *TeenPattiPlayer) ResetForDeal() {
	p.seen = false
	p.folded = false
	p.roundBet = 0
	p.Reset()
	p.SetIsFinished(false)
}

// teenPattiPlayerJSON is the JSON wire format for TeenPattiPlayer.
type teenPattiPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Chips      int         `json:"ch"`
	Seen       bool        `json:"sn"`
	Folded     bool        `json:"fd"`
	Out        bool        `json:"ot"`
	RoundBet   int         `json:"rb"`
}

// MarshalJSON implements json.Marshaler.
func (p *TeenPattiPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(teenPattiPlayerJSON{
		GamePlayer: p.GamePlayer,
		Chips:      p.GetChips(),
		Seen:       p.seen,
		Folded:     p.folded,
		Out:        p.out,
		RoundBet:   p.roundBet,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TeenPattiPlayer) UnmarshalJSON(data []byte) error {
	var j teenPattiPlayerJSON
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
