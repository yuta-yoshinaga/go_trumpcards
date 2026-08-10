package domain

import "encoding/json"

// GongZhuPlayer 拱猪（Gong Zhu）プレイヤークラス
type GongZhuPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
}

// NewGongZhuPlayer コンストラクタ
func NewGongZhuPlayer(isHuman bool) *GongZhuPlayer {
	return &GongZhuPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（スコア・トリック・手札・終了状態を初期化）
func (p *GongZhuPlayer) ResetRound() {
	resetRoundWithTricks(p)
}

// gongZhuPlayerJSON is the JSON wire format for GongZhuPlayer.
type gongZhuPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *GongZhuPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(gongZhuPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *GongZhuPlayer) UnmarshalJSON(data []byte) error {
	var j gongZhuPlayerJSON
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
	return nil
}
