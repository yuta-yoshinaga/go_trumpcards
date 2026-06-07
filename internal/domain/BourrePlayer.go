//go:build !js || !wasm || casino

package domain

import "encoding/json"

// BourrePlayer ブーレプレイヤークラス
type BourrePlayer struct {
	*GamePlayer
	ChipHolder
	TrickHolder
	folded   bool // このハンドでフォールド（降り）したか
	decided  bool // このハンドで参加/フォールドを決めたか
	drawn    bool // このハンドでドロー（手札交換）を済ませたか
	bourreed bool // 直近ハンドでブーレ（参加して0トリック）になったか（表示用）
}

// NewBourrePlayer コンストラクタ。初期チップを付与する。
func NewBourrePlayer(isHuman bool) *BourrePlayer {
	p := &BourrePlayer{GamePlayer: NewGamePlayer(isHuman)}
	p.SetChips(BourreInitChips)
	return p
}

// GetFolded フォールド状態を取得
func (p *BourrePlayer) GetFolded() bool { return p.folded }

// SetFolded フォールド状態を設定
func (p *BourrePlayer) SetFolded(v bool) { p.folded = v }

// GetDecided 参加可否の決定済みフラグを取得
func (p *BourrePlayer) GetDecided() bool { return p.decided }

// SetDecided 参加可否の決定済みフラグを設定
func (p *BourrePlayer) SetDecided(v bool) { p.decided = v }

// GetDrawn ドロー済みフラグを取得
func (p *BourrePlayer) GetDrawn() bool { return p.drawn }

// SetDrawn ドロー済みフラグを設定
func (p *BourrePlayer) SetDrawn(v bool) { p.drawn = v }

// GetBourreed ブーレ状態を取得（表示用）
func (p *BourrePlayer) GetBourreed() bool { return p.bourreed }

// SetBourreed ブーレ状態を設定（表示用）
func (p *BourrePlayer) SetBourreed(v bool) { p.bourreed = v }

// ResetHand ハンド開始時に手札・トリック・各フラグを初期化する
func (p *BourrePlayer) ResetHand() {
	p.Reset()
	p.ResetTricks()
	p.folded = false
	p.decided = false
	p.drawn = false
	p.bourreed = false
}

// bourrePlayerJSON is the JSON wire format for BourrePlayer.
type bourrePlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	ChipHolder  *ChipHolder  `json:"ch"`
	TrickHolder *TrickHolder `json:"th"`
	Folded      bool         `json:"fd"`
	Decided     bool         `json:"de"`
	Drawn       bool         `json:"dr"`
	Bourreed    bool         `json:"bo"`
}

// MarshalJSON implements json.Marshaler.
func (p *BourrePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(bourrePlayerJSON{
		GamePlayer:  p.GamePlayer,
		ChipHolder:  &p.ChipHolder,
		TrickHolder: &p.TrickHolder,
		Folded:      p.folded,
		Decided:     p.decided,
		Drawn:       p.drawn,
		Bourreed:    p.bourreed,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BourrePlayer) UnmarshalJSON(data []byte) error {
	var j bourrePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.ChipHolder != nil {
		p.ChipHolder = *j.ChipHolder
	}
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	p.folded = j.Folded
	p.decided = j.Decided
	p.drawn = j.Drawn
	p.bourreed = j.Bourreed
	return nil
}
