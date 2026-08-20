//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"errors"
)

// LingerLongerPlayer はリンガーロンガーのプレイヤー。
//
// **獲得置き場はありません。** 取ったトリックは得点にならず、意味があるのは
// 「山札から 1 枚補充できる」ことだけです。
type LingerLongerPlayer struct {
	*GamePlayer
	// tricksWon は取ったトリック数（表示用）。
	tricksWon int
	// eliminatedAt は脱落した順番（0 = まだ在席）。
	eliminatedAt int
}

// NewLingerLongerPlayer はコンストラクタ。
func NewLingerLongerPlayer(isHuman bool) *LingerLongerPlayer {
	return &LingerLongerPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame はゲーム全体をリセットする。
func (p *LingerLongerPlayer) ResetGame() {
	p.Reset()
	p.tricksWon = 0
	p.eliminatedAt = 0
}

// GetTricksWon は取ったトリック数を返す。
func (p *LingerLongerPlayer) GetTricksWon() int { return p.tricksWon }

// AddTrickWon は取ったトリック数を 1 増やす。
func (p *LingerLongerPlayer) AddTrickWon() { p.tricksWon++ }

// GetEliminatedAt は脱落した順番を返す（0 = まだ在席）。
func (p *LingerLongerPlayer) GetEliminatedAt() int { return p.eliminatedAt }

// SetEliminatedAt は脱落した順番を設定する。
func (p *LingerLongerPlayer) SetEliminatedAt(order int) { p.eliminatedAt = order }

// IsEliminated は脱落したかを返す。
func (p *LingerLongerPlayer) IsEliminated() bool { return p.eliminatedAt > 0 }

// lingerLongerPlayerJSON is the JSON wire format for LingerLongerPlayer.
type lingerLongerPlayerJSON struct {
	GamePlayer   *GamePlayer `json:"gp"`
	TricksWon    int         `json:"tw"`
	EliminatedAt int         `json:"ea"`
}

// MarshalJSON implements json.Marshaler.
func (p *LingerLongerPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(lingerLongerPlayerJSON{
		GamePlayer: p.GamePlayer, TricksWon: p.tricksWon, EliminatedAt: p.eliminatedAt,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **脱落した席は手札を持たない。** ただし逆は成り立ちません——最後の 1 枚を出した
// 席は、そのトリックが解決して**補充が済むまで**脱落が決まらないからです。勝てば
// 山札から 1 枚もらって残るので、「手札が空 ⇒ 脱落」と書くと、トリックの途中の
// 正当な盤面を拒否します（毎手ごとに往復させるテストが検出しました）。
func (p *LingerLongerPlayer) UnmarshalJSON(data []byte) error {
	var j lingerLongerPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.TricksWon < 0 {
		return errors.New("negative trick count")
	}
	if j.EliminatedAt < 0 || j.EliminatedAt > LingerLongerPlayerCntMax {
		return errors.New("invalid elimination order")
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.EliminatedAt > 0 && p.GetCardsSize() > 0 {
		return errors.New("an eliminated seat cannot hold cards")
	}
	p.tricksWon, p.eliminatedAt = j.TricksWon, j.EliminatedAt
	return nil
}
