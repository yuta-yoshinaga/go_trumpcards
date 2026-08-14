//go:build !js || !wasm || classic

package domain

import "encoding/json"

// GermanWhistPlayer ジャーマンホイストのプレイヤー
type GermanWhistPlayer struct {
	*GamePlayer
	TrickHolder
	// scoringTricks は**後半 13 トリックだけ**の獲得数。前半のトリックは
	// 山札の札を取るためのもので、得点にならない。
	scoringTricks int
}

// NewGermanWhistPlayer コンストラクタ
func NewGermanWhistPlayer(isHuman bool) *GermanWhistPlayer {
	return &GermanWhistPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame ゲームをリセット (手札/トリック/得点を初期化)
func (p *GermanWhistPlayer) ResetGame() {
	resetPlayerRound(p)
	p.scoringTricks = 0
}

// GetScoringTricks 後半に取ったトリック数
func (p *GermanWhistPlayer) GetScoringTricks() int { return p.scoringTricks }

// AddScoringTrick 後半のトリック獲得を 1 つ数える
func (p *GermanWhistPlayer) AddScoringTrick() { p.scoringTricks++ }

// SetScoringTricks 後半のトリック数を設定する（復元・テスト用）
func (p *GermanWhistPlayer) SetScoringTricks(n int) { p.scoringTricks = n }

// germanWhistPlayerJSON is the JSON wire format for GermanWhistPlayer.
type germanWhistPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	// ScoringTricks must persist: the Worker rebuilds the game from KV on every
	// request, and without it the score silently resets mid-hand (#4478).
	ScoringTricks int `json:"sc"`
}

// MarshalJSON implements json.Marshaler.
func (p *GermanWhistPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(germanWhistPlayerJSON{
		GamePlayer:    p.GamePlayer,
		TrickHolder:   &p.TrickHolder,
		ScoringTricks: p.scoringTricks,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *GermanWhistPlayer) UnmarshalJSON(data []byte) error {
	var j germanWhistPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	p.scoringTricks = j.ScoringTricks
	return nil
}
