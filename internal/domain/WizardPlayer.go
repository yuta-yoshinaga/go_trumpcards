package domain

import "encoding/json"

// WizardPlayer ウィザードプレイヤークラス
type WizardPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
	bid int // 宣言したトリック数 (-1 = 未ビッド)
}

// NewWizardPlayer コンストラクタ
func NewWizardPlayer(isHuman bool) *WizardPlayer {
	return &WizardPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		bid:        -1,
	}
}

// GetBid ビッド取得 (-1 = 未ビッド)
func (p *WizardPlayer) GetBid() int { return p.bid }

// SetBid ビッド設定
func (p *WizardPlayer) SetBid(bid int) { p.bid = bid }

// ResetRound ラウンドをリセット（ビッド・トリック・手札・終了状態を初期化）
func (p *WizardPlayer) ResetRound() {
	p.bid = -1
	resetRoundWithTricks(p)
}

// wizardPlayerJSON is the JSON wire format for WizardPlayer.
type wizardPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
	Bid              int               `json:"bd"`
}

// MarshalJSON implements json.Marshaler.
func (p *WizardPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(wizardPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
		Bid:              p.bid,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *WizardPlayer) UnmarshalJSON(data []byte) error {
	var j wizardPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.RoundScoreHolder != nil {
		p.RoundScoreHolder = *j.RoundScoreHolder
	}
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	p.bid = j.Bid
	return nil
}
