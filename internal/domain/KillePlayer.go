//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// KilleMaxReentries は買い戻しできる回数。
//
// **ライフ制ではない。**負ければ脱落し、掛け金を払って戻るのは 3 回まで。
// 額は 1 回目が 1 口、2 回目がポットの半分、3 回目がポット全額と上がる。
const KilleMaxReentries = 3

// KillePlayer はキッレのプレイヤークラス。
type KillePlayer struct {
	*GamePlayer
	// chips は通算の収支。
	chips int
	// reentries はこれまでに買い戻した回数。KilleMaxReentries で打ち止め。
	reentries int
	// out はこのラウンドで脱落したか。Hussar / Pig で落とされた場合も含む。
	out bool
	// knockedBy は誰の効果で落とされたかの識別子 ("" なら最弱による脱落)。
	knockedBy string
	// harlequinSwapped は Harlequin を**交換で**受け取ったか。
	//
	// **同じ札でも向きで強さが反転する。**配られた／山から引いた Harlequin は
	// 最強、交換で渡ってきた Harlequin は最弱になる。
	harlequinSwapped bool
	// satisfied はこのラウンドで「交換しない」と宣言したか。
	satisfied bool
}

// NewKillePlayer コンストラクタ
func NewKillePlayer(isHuman bool) *KillePlayer {
	return &KillePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// GetChips は収支を返す。
func (p *KillePlayer) GetChips() int { return p.chips }

// AddChips は収支を増減する。
func (p *KillePlayer) AddChips(n int) { p.chips += n }

// GetReentries は買い戻した回数を返す。
func (p *KillePlayer) GetReentries() int { return p.reentries }

// AddReentry は買い戻し回数を 1 増やす。
func (p *KillePlayer) AddReentry() { p.reentries++ }

// CanReenter はまだ買い戻せるかを返す。
func (p *KillePlayer) CanReenter() bool { return p.reentries < KilleMaxReentries }

// IsOut はこのラウンドで脱落しているかを返す。
func (p *KillePlayer) IsOut() bool { return p.out }

// SetOut は脱落させる。by は落とした効果の識別子。
func (p *KillePlayer) SetOut(by string) {
	p.out = true
	p.knockedBy = by
}

// GetKnockedBy は落とした効果の識別子を返す ("" なら最弱による脱落)。
func (p *KillePlayer) GetKnockedBy() string { return p.knockedBy }

// IsHarlequinSwapped は手札の Harlequin が交換で渡ってきたかを返す。
func (p *KillePlayer) IsHarlequinSwapped() bool { return p.harlequinSwapped }

// SetHarlequinSwapped は Harlequin の向きを設定する。
func (p *KillePlayer) SetHarlequinSwapped(v bool) { p.harlequinSwapped = v }

// IsSatisfied は「交換しない」と宣言済みかを返す。
func (p *KillePlayer) IsSatisfied() bool { return p.satisfied }

// SetSatisfied は「交換しない」を記録する。
func (p *KillePlayer) SetSatisfied(v bool) { p.satisfied = v }

// ResetRound はラウンド開始時に手札と一時状態を初期化する。収支は残す。
func (p *KillePlayer) ResetRound() {
	p.Reset()
	p.SetIsFinished(false)
	p.out = false
	p.knockedBy = ""
	p.harlequinSwapped = false
	p.satisfied = false
}

// killePlayerJSON is the JSON wire format for KillePlayer.
type killePlayerJSON struct {
	GamePlayer       *GamePlayer `json:"gp"`
	Chips            int         `json:"ch"`
	Reentries        int         `json:"re"`
	Out              bool        `json:"ot"`
	KnockedBy        string      `json:"kb"`
	HarlequinSwapped bool        `json:"hs"`
	Satisfied        bool        `json:"sa"`
}

// MarshalJSON implements json.Marshaler.
func (p *KillePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(killePlayerJSON{
		GamePlayer: p.GamePlayer, Chips: p.chips, Reentries: p.reentries,
		Out: p.out, KnockedBy: p.knockedBy,
		HarlequinSwapped: p.harlequinSwapped, Satisfied: p.satisfied,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *KillePlayer) UnmarshalJSON(data []byte) error {
	var j killePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.chips = j.Chips
	p.reentries = j.Reentries
	if p.reentries < 0 {
		p.reentries = 0
	}
	if p.reentries > KilleMaxReentries {
		p.reentries = KilleMaxReentries
	}
	p.out = j.Out
	p.knockedBy = j.KnockedBy
	p.harlequinSwapped = j.HarlequinSwapped
	p.satisfied = j.Satisfied
	return nil
}
