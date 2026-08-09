//go:build !js || !wasm || extra

package domain

import "encoding/json"

// ChinchonPlayer チンチョンプレイヤークラス
//
// 手札 (GamePlayer 内) に加えて、ラウンド点・累積点 (RoundScoreHolder) と、
// 脱落フラグ (eliminated) を保持する。チンチョンは累積点が脱落上限を超えたプレイヤーを
// 順次脱落させ、最後に残ったプレイヤーがマッチ勝者となる。
type ChinchonPlayer struct {
	*GamePlayer
	RoundScoreHolder
	eliminated bool // 累積点が上限を超えて脱落したか
}

// NewChinchonPlayer コンストラクタ
func NewChinchonPlayer(isHuman bool) *ChinchonPlayer {
	return &ChinchonPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// GetEliminated 脱落しているかを返す
func (p *ChinchonPlayer) GetEliminated() bool { return p.eliminated }

// SetEliminated 脱落状態を設定する
func (p *ChinchonPlayer) SetEliminated(v bool) { p.eliminated = v }

// ResetRound ラウンドをリセットする (手札・ラウンド点・終了状態を初期化、累積点と脱落状態は維持)。
func (p *ChinchonPlayer) ResetRound() {
	resetRoundScored(p)
}

// chinchonPlayerJSON is the JSON wire format for ChinchonPlayer.
type chinchonPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	Eliminated       bool              `json:"el"`
}

// MarshalJSON implements json.Marshaler.
func (p *ChinchonPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(chinchonPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		Eliminated:       p.eliminated,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ChinchonPlayer) UnmarshalJSON(data []byte) error {
	var j chinchonPlayerJSON
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
	p.eliminated = j.Eliminated
	return nil
}
