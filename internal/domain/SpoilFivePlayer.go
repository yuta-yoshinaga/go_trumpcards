//go:build !js || !wasm || classic

package domain

import "encoding/json"

// SpoilFivePlayer スポイル・ファイブのプレイヤークラス。手札（GamePlayer）と獲得
// トリック（TrickHolder）に加え、累積得点と現ラウンドのトリック数を保持する。
type SpoilFivePlayer struct {
	*GamePlayer
	TrickHolder
	score       int
	roundTricks int
}

// NewSpoilFivePlayer コンストラクタ
func NewSpoilFivePlayer(isHuman bool) *SpoilFivePlayer {
	return &SpoilFivePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット（トリック・手札・ラウンドトリック数を初期化）。
// 累積得点はラウンドをまたいで保持する。
func (p *SpoilFivePlayer) ResetRound() {
	resetPlayerRound(p)
	p.roundTricks = 0
}

// GetScore 累積得点を返す。
func (p *SpoilFivePlayer) GetScore() int { return p.score }

// SetScore 累積得点を設定する。
func (p *SpoilFivePlayer) SetScore(n int) { p.score = n }

// GetRoundTricks 現ラウンドの獲得トリック数を返す。
func (p *SpoilFivePlayer) GetRoundTricks() int { return p.roundTricks }

// IncRoundTricks 現ラウンドの獲得トリック数を 1 増やす。
func (p *SpoilFivePlayer) IncRoundTricks() { p.roundTricks++ }

// spoilFivePlayerJSON is the JSON wire format for SpoilFivePlayer.
type spoilFivePlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Score       int          `json:"sc"`
	RoundTricks int          `json:"rt"`
}

// MarshalJSON implements json.Marshaler.
func (p *SpoilFivePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(spoilFivePlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Score:       p.score,
		RoundTricks: p.roundTricks,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SpoilFivePlayer) UnmarshalJSON(data []byte) error {
	var j spoilFivePlayerJSON
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
	p.score = j.Score
	p.roundTricks = j.RoundTricks
	return nil
}
