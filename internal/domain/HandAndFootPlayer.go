//go:build !js || !wasm || solo

package domain

import "encoding/json"

// HandAndFootPlayer ハンドアンドフットプレイヤークラス。
// メルドはチーム単位で共有されるため Hand and Foot ゲーム本体が保持する。
// プレイヤーは手札 (GamePlayer) と「フット」(脇に置いた予備手札) のみを持つ。
type HandAndFootPlayer struct {
	*GamePlayer
	RoundScoreHolder
	foot   []*Card // フット（手札を出し切るまで脇に置く予備手札）
	inFoot bool    // フットを手札に取り込み済みか
}

// NewHandAndFootPlayer コンストラクタ
func NewHandAndFootPlayer(isHuman bool) *HandAndFootPlayer {
	return &HandAndFootPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		foot:       make([]*Card, 0),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・フットを初期化）
func (p *HandAndFootPlayer) ResetRound() {
	p.SetRoundScore(0)
	p.Reset()
	p.SetIsFinished(false)
	p.foot = make([]*Card, 0)
	p.inFoot = false
}

// GetFoot フットを取得
func (p *HandAndFootPlayer) GetFoot() []*Card { return p.foot }

// SetFoot フットを設定 (テスト用)
func (p *HandAndFootPlayer) SetFoot(foot []*Card) { p.foot = foot }

// AddFootCard フットにカードを追加
func (p *HandAndFootPlayer) AddFootCard(card *Card) { p.foot = append(p.foot, card) }

// GetFootSize フットの枚数を取得
func (p *HandAndFootPlayer) GetFootSize() int { return len(p.foot) }

// GetInFoot フット取り込み済みフラグ取得
func (p *HandAndFootPlayer) GetInFoot() bool { return p.inFoot }

// SetInFoot フット取り込み済みフラグ設定 (テスト用)
func (p *HandAndFootPlayer) SetInFoot(v bool) { p.inFoot = v }

// handAndFootPlayerJSON is the JSON wire format for HandAndFootPlayer.
type handAndFootPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	Foot             []*Card           `json:"ft"`
	InFoot           bool              `json:"if,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (p *HandAndFootPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(handAndFootPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		Foot:             p.foot,
		InFoot:           p.inFoot,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *HandAndFootPlayer) UnmarshalJSON(data []byte) error {
	var j handAndFootPlayerJSON
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
	p.foot = j.Foot
	if p.foot == nil {
		p.foot = make([]*Card, 0)
	}
	p.inFoot = j.InFoot
	return nil
}
