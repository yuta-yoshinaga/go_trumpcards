//go:build !js || !wasm || casino

package domain

import "encoding/json"

// IndianPokerPlayer インディアンポーカープレイヤークラス
type IndianPokerPlayer struct {
	Player                            // 親クラス
	ChipHolder                        // チップ管理
	bettingPlayerBase                 // ベッティング共通状態
	isHuman           bool            // 人間フラグ
	playStyle         HoldemPlayStyle // CPUプレイスタイル
}

// NewIndianPokerPlayer コンストラクタ
func NewIndianPokerPlayer(isHuman bool, style HoldemPlayStyle) *IndianPokerPlayer {
	return &IndianPokerPlayer{
		Player:    Player{cards: make([]*Card, 0)},
		isHuman:   isHuman,
		playStyle: style,
	}
}

// GetIsHuman 人間フラグ取得
func (p *IndianPokerPlayer) GetIsHuman() bool { return p.isHuman }

// GetPlayStyle プレイスタイル取得
func (p *IndianPokerPlayer) GetPlayStyle() HoldemPlayStyle { return p.playStyle }

// GetPlayStyleName プレイスタイル名取得
func (p *IndianPokerPlayer) GetPlayStyleName() string {
	return playStyleName(int(p.playStyle), HoldemPlayStyleNames)
}

// indianPokerPlayerJSON is the JSON wire format for IndianPokerPlayer.
type indianPokerPlayerJSON struct {
	Player            *Player            `json:"p"`
	ChipHolder        *ChipHolder        `json:"ch"`
	BettingPlayerBase *bettingPlayerBase `json:"bp"`
	IsHuman           bool               `json:"ih"`
	PlayStyle         HoldemPlayStyle    `json:"ps"`
}

// MarshalJSON implements json.Marshaler.
func (p *IndianPokerPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(indianPokerPlayerJSON{
		Player:            &p.Player,
		ChipHolder:        &p.ChipHolder,
		BettingPlayerBase: &p.bettingPlayerBase,
		IsHuman:           p.isHuman,
		PlayStyle:         p.playStyle,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *IndianPokerPlayer) UnmarshalJSON(data []byte) error {
	var j indianPokerPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Player != nil {
		p.Player = *j.Player
	}
	if j.ChipHolder != nil {
		p.ChipHolder = *j.ChipHolder
	}
	if j.BettingPlayerBase != nil {
		p.bettingPlayerBase = *j.BettingPlayerBase
	}
	p.isHuman = j.IsHuman
	p.playStyle = j.PlayStyle
	return nil
}

// GetComparisonCards ハンド比較用カード取得 (BettingPlayerインターフェース)
func (p *IndianPokerPlayer) GetComparisonCards() []*Card {
	if p.GetCardsSize() == 0 {
		return nil
	}
	return []*Card{p.GetCard(0)}
}
