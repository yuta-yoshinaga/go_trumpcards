//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"fmt"
)

// HoneymoonBridgePlayer はハネムーンブリッジのプレイヤー。
type HoneymoonBridgePlayer struct {
	*GamePlayer
	TrickHolder
	// bidLevel は宣言したレベル（0 = 未宣言/降り、1..7）。
	bidLevel int
	// bidSuit は宣言したスート（0 = ノートランプ、1..4 = スート）。
	bidSuit int
	// score は累計得点。**勝敗はこれで決まる。**
	score int
}

// NewHoneymoonBridgePlayer はコンストラクタ。
func NewHoneymoonBridgePlayer(isHuman bool) *HoneymoonBridgePlayer {
	return &HoneymoonBridgePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame はゲーム全体をリセットする。
func (p *HoneymoonBridgePlayer) ResetGame() {
	p.ResetRound()
	p.score = 0
}

// ResetRound は 1 ディール分の状態を初期化する。
func (p *HoneymoonBridgePlayer) ResetRound() {
	resetPlayerRound(p)
	p.bidLevel = 0
	p.bidSuit = 0
}

// GetBidLevel は宣言したレベルを返す（0 = 未宣言/降り）。
func (p *HoneymoonBridgePlayer) GetBidLevel() int { return p.bidLevel }

// GetBidSuit は宣言したスートを返す（0 = ノートランプ）。
func (p *HoneymoonBridgePlayer) GetBidSuit() int { return p.bidSuit }

// SetBid は宣言を設定する。
func (p *HoneymoonBridgePlayer) SetBid(level, suit int) { p.bidLevel, p.bidSuit = level, suit }

// GetScore は累計得点を返す。
func (p *HoneymoonBridgePlayer) GetScore() int { return p.score }

// AddScore は得点を加算する。
func (p *HoneymoonBridgePlayer) AddScore(n int) { p.score += n }

// SetScore は得点を設定する（復元・テスト用）。
func (p *HoneymoonBridgePlayer) SetScore(n int) { p.score = n }

// honeymoonBridgePlayerJSON is the JSON wire format for HoneymoonBridgePlayer.
type honeymoonBridgePlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	BidLevel    int          `json:"bl"`
	BidSuit     int          `json:"bs"`
	Score       int          `json:"sc"`
}

// MarshalJSON implements json.Marshaler.
func (p *HoneymoonBridgePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(honeymoonBridgePlayerJSON{
		GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder,
		BidLevel: p.bidLevel, BidSuit: p.bidSuit, Score: p.score,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **宣言は契約そのものなので範囲を検証する** (#5302〜#5305、#5309、#5311 と同じ形)。
// レベルが壊れると必要トリック数が変わり、成否判定がそのまま入れ替わります。
func (p *HoneymoonBridgePlayer) UnmarshalJSON(data []byte) error {
	var j honeymoonBridgePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	// 0 は「未宣言または降り」の正当な値。
	if j.BidLevel < 0 || j.BidLevel > HoneymoonBridgeMaxLevel {
		return fmt.Errorf("invalid bid level: %d", j.BidLevel)
	}
	// スートは 0（ノートランプ）か 1..4。**レベル 0 ならスートも 0。**
	if j.BidSuit < 0 || j.BidSuit > CardDesignDiamond {
		return fmt.Errorf("invalid bid suit: %d", j.BidSuit)
	}
	if j.BidLevel == 0 && j.BidSuit != 0 {
		return fmt.Errorf("bid suit %d without a level", j.BidSuit)
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	p.bidLevel, p.bidSuit, p.score = j.BidLevel, j.BidSuit, j.Score
	return nil
}
