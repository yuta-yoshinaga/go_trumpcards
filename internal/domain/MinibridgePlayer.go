//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"fmt"
)

// MinibridgePlayer はミニブリッジのプレイヤー。
type MinibridgePlayer struct {
	*GamePlayer
	TrickHolder
	// hcp はこのディールで公開申告したハイカードポイント。
	// **競りが無いこのゲームでは、これが唯一の公開情報。**
	hcp int
	// team は 0 か 1。**席 0/2 が team 0、席 1/3 が team 1。**
	team int
}

// NewMinibridgePlayer はコンストラクタ。
func NewMinibridgePlayer(isHuman bool, team int) *MinibridgePlayer {
	return &MinibridgePlayer{GamePlayer: NewGamePlayer(isHuman), team: team}
}

// ResetGame はゲーム全体をリセットする。
func (p *MinibridgePlayer) ResetGame() { p.ResetRound() }

// ResetRound は 1 ディール分の状態を初期化する。
func (p *MinibridgePlayer) ResetRound() {
	resetPlayerRound(p)
	p.hcp = 0
}

// GetHcp は申告済みのハイカードポイントを返す。
func (p *MinibridgePlayer) GetHcp() int { return p.hcp }

// SetHcp はハイカードポイントを設定する。
func (p *MinibridgePlayer) SetHcp(n int) { p.hcp = n }

// GetTeam はチーム番号を返す。
func (p *MinibridgePlayer) GetTeam() int { return p.team }

// minibridgePlayerJSON is the JSON wire format for MinibridgePlayer.
type minibridgePlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Hcp         int          `json:"h"`
	Team        int          `json:"t"`
}

// MarshalJSON implements json.Marshaler.
func (p *MinibridgePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(minibridgePlayerJSON{
		GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder,
		Hcp: p.hcp, Team: p.team,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **HCP は席を跨いだ制約を持つ。** 総和は必ず 40 なので、席単体の上限も 40 で、
// 復元後の合計は Minibridge.UnmarshalJSON 側で突き合わせます (#5312 と同じ形の
// 「個々は範囲内だが組み合わせがあり得ない」を防ぐため)。
func (p *MinibridgePlayer) UnmarshalJSON(data []byte) error {
	var j minibridgePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Hcp < 0 || j.Hcp > MinibridgeTotalHcp {
		return fmt.Errorf("invalid hcp: %d", j.Hcp)
	}
	if j.Team < 0 || j.Team >= MinibridgeTeamCnt {
		return fmt.Errorf("invalid team: %d", j.Team)
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	p.hcp, p.team = j.Hcp, j.Team
	return nil
}
