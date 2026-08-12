//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// CucumberPlayer はキューカンバーのプレイヤー。
//
// **点は取るものではなく、付くもの。** penalty は失点で、少ないほうが good です。
type CucumberPlayer struct {
	*GamePlayer
	penalty int
}

// NewCucumberPlayer はコンストラクタ。
func NewCucumberPlayer(isHuman bool) *CucumberPlayer {
	return &CucumberPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// GetPenalty は失点を返す。
func (p *CucumberPlayer) GetPenalty() int { return p.penalty }

// SetPenalty は失点を設定する (主にテスト/復元用)。
func (p *CucumberPlayer) SetPenalty(n int) { p.penalty = n }

// AddPenalty は失点を加える。
func (p *CucumberPlayer) AddPenalty(n int) { p.penalty += n }

// ResetGame はゲーム開始時の状態に戻す。
func (p *CucumberPlayer) ResetGame() {
	p.Reset()
	p.penalty = 0
	p.SetIsFinished(false)
}

// cucumberPlayerJSON is the JSON wire format for CucumberPlayer.
type cucumberPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Penalty    int         `json:"pn"`
}

// MarshalJSON implements json.Marshaler.
func (p *CucumberPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(cucumberPlayerJSON{GamePlayer: p.GamePlayer, Penalty: p.penalty})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *CucumberPlayer) UnmarshalJSON(data []byte) error {
	var j cucumberPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer == nil {
		return errors.New("cucumber player is missing its base player")
	}
	// **失点は減りません。** 負の失点は書き込み側が壊れた証拠です。
	if j.Penalty < 0 {
		return fmt.Errorf("penalty cannot be negative, got %d", j.Penalty)
	}
	p.GamePlayer = j.GamePlayer
	p.penalty = j.Penalty
	return nil
}
