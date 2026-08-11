//go:build !js || !wasm || extra

package domain

import "encoding/json"

// BhabhiPlayer はバービーのプレイヤー。
//
// **順位は「上がった順」であって強さの順ではない。** 最後まで手札が残った
// 1 人だけが Bhabhi（敗者）で、それ以外は全員ただの勝ち抜けです。
type BhabhiPlayer struct {
	*RankedGamePlayer
	// pickups はこのゲームで場札を引き取った回数。勝敗には効かないが、
	// フォローできなかった回数がそのまま出るので進行の手掛かりになる。
	pickups int
}

// NewBhabhiPlayer はコンストラクタ。
func NewBhabhiPlayer(isHuman bool) *BhabhiPlayer {
	return &BhabhiPlayer{RankedGamePlayer: NewRankedGamePlayer(isHuman)}
}

// ResetGame はゲーム開始時の初期化。
func (p *BhabhiPlayer) ResetGame() {
	p.Reset()
	p.SetRank(-1)
	p.SetIsFinished(false)
	p.pickups = 0
}

// IsOut は勝ち抜け済み（手札を出し切った）かを返す。
//
// **rank が振られているかで見る。** 手札 0 枚で見ると、引き取り直前の
// 「出したばかりで一時的に 0 枚」を上がりと取り違えます。
func (p *BhabhiPlayer) IsOut() bool { return p.GetRank() > 0 }

// GetPickups は場札を引き取った回数を返す。
func (p *BhabhiPlayer) GetPickups() int { return p.pickups }

// AddPickup は引き取り回数を 1 増やす。
func (p *BhabhiPlayer) AddPickup() { p.pickups++ }

// SetPickups は引き取り回数を設定する（復元・テスト用）。
func (p *BhabhiPlayer) SetPickups(n int) { p.pickups = n }

// bhabhiPlayerJSON is the JSON wire format for BhabhiPlayer.
//
// **rank と pickups は非公開なので明示的に載せる。** 抜けると Worker が
// リクエストごとに状態を作り直したときに勝ち抜けが全部取り消され、一度
// 上がった人にまた手番が回ります (#4478)。
type bhabhiPlayerJSON struct {
	RankedGamePlayer *RankedGamePlayer `json:"rp"`
	Pickups          int               `json:"pu"`
}

// MarshalJSON implements json.Marshaler.
func (p *BhabhiPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(bhabhiPlayerJSON{RankedGamePlayer: p.RankedGamePlayer, Pickups: p.pickups})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BhabhiPlayer) UnmarshalJSON(data []byte) error {
	var j bhabhiPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.RankedGamePlayer != nil {
		p.RankedGamePlayer = j.RankedGamePlayer
	} else {
		p.RankedGamePlayer = NewRankedGamePlayer(false)
	}
	p.pickups = j.Pickups
	return nil
}
