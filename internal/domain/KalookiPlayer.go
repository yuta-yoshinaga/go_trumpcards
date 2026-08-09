//go:build !js || !wasm || extra

package domain

import "encoding/json"

// KalookiPlayer は Kalooki のプレイヤー。手札・場のメルド・オープン状態を保持する。
type KalookiPlayer struct {
	*GamePlayer
	RoundScoreHolder
	melds     [][]*Card // 場に出したメルド
	hasOpened bool      // オープニング要件（>= しきい値）を満たしてメルドを出したか
}

// NewKalookiPlayer コンストラクタ
func NewKalookiPlayer(isHuman bool) *KalookiPlayer {
	return &KalookiPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		melds:      nil,
	}
}

// GetMelds 場に出したメルド一覧を返す
func (p *KalookiPlayer) GetMelds() [][]*Card { return p.melds }

// SetMelds メルドを差し替える（テスト用）
func (p *KalookiPlayer) SetMelds(m [][]*Card) { p.melds = m }

// GetMeldCount メルド数を取得
func (p *KalookiPlayer) GetMeldCount() int { return len(p.melds) }

// GetMeld 指定インデックスのメルドを取得
func (p *KalookiPlayer) GetMeld(i int) []*Card {
	if i < 0 || i >= len(p.melds) {
		return nil
	}
	return p.melds[i]
}

// AppendMeld 新しいメルドを追加する
func (p *KalookiPlayer) AppendMeld(meld []*Card) {
	cp := make([]*Card, len(meld))
	copy(cp, meld)
	p.melds = append(p.melds, cp)
}

// AddCardToMeld 指定インデックスのメルドに 1 枚追加する
func (p *KalookiPlayer) AddCardToMeld(i int, card *Card) bool {
	if i < 0 || i >= len(p.melds) {
		return false
	}
	p.melds[i] = append(p.melds[i], card)
	return true
}

// ClearMelds メルドとオープン状態を空にする
func (p *KalookiPlayer) ClearMelds() {
	p.melds = nil
	p.hasOpened = false
}

// HasOpened オープニング要件を満たしてメルドを出したか
func (p *KalookiPlayer) HasOpened() bool { return p.hasOpened }

// SetHasOpened オープン状態を設定する
func (p *KalookiPlayer) SetHasOpened(opened bool) { p.hasOpened = opened }

// ResetRound ラウンドをリセット（手札・スコア・メルド・終了状態を初期化）
func (p *KalookiPlayer) ResetRound() {
	resetRoundScored(p)
	p.ClearMelds()
}

// kalookiPlayerJSON は KalookiPlayer の JSON 表現。
type kalookiPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	Melds            [][]*Card         `json:"md"`
	HasOpened        bool              `json:"op"`
}

// MarshalJSON implements json.Marshaler.
func (p *KalookiPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(kalookiPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		Melds:            p.melds,
		HasOpened:        p.hasOpened,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *KalookiPlayer) UnmarshalJSON(data []byte) error {
	var j kalookiPlayerJSON
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
	p.melds = j.Melds
	p.hasOpened = j.HasOpened
	return nil
}
