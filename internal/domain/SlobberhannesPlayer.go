//go:build !js || !wasm || classic

package domain

import "encoding/json"

// SlobberhannesPlayer スロバーハンネスのプレイヤー
type SlobberhannesPlayer struct {
	*GamePlayer
	TrickHolder
	// score は全ラウンドの累計。**罰点が負、全回避ボーナスが正**なので、
	// 合計が一番「大きい」プレイヤーが勝つ。
	score int
	// このラウンドで受けた 3 つの罰の内訳。全部 false なら +1 のボーナス。
	tookFirstTrick bool
	tookLastTrick  bool
	tookQueen      bool
}

// NewSlobberhannesPlayer コンストラクタ
func NewSlobberhannesPlayer(isHuman bool) *SlobberhannesPlayer {
	return &SlobberhannesPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame ゲーム全体をリセットする（累計得点も 0 に戻す）
func (p *SlobberhannesPlayer) ResetGame() {
	p.ResetRound()
	p.score = 0
}

// ResetRound 1 ラウンド分の状態だけを初期化する。累計得点は残す。
func (p *SlobberhannesPlayer) ResetRound() {
	resetPlayerRound(p)
	p.tookFirstTrick = false
	p.tookLastTrick = false
	p.tookQueen = false
}

// GetScore 累計得点を返す
func (p *SlobberhannesPlayer) GetScore() int { return p.score }

// AddScore 累計得点に加算する（罰点は負の値で渡す）
func (p *SlobberhannesPlayer) AddScore(n int) { p.score += n }

// SetScore 累計得点を設定する（復元・テスト用）
func (p *SlobberhannesPlayer) SetScore(n int) { p.score = n }

// GetTookFirstTrick 最初のトリックを取ったか
func (p *SlobberhannesPlayer) GetTookFirstTrick() bool { return p.tookFirstTrick }

// GetTookLastTrick 最後のトリックを取ったか
func (p *SlobberhannesPlayer) GetTookLastTrick() bool { return p.tookLastTrick }

// GetTookQueen クラブのクイーンを取ったか
func (p *SlobberhannesPlayer) GetTookQueen() bool { return p.tookQueen }

// PenaltyCount このラウンドで受けた罰の数（0..3）
func (p *SlobberhannesPlayer) PenaltyCount() int {
	n := 0
	for _, took := range []bool{p.tookFirstTrick, p.tookLastTrick, p.tookQueen} {
		if took {
			n++
		}
	}
	return n
}

// slobberhannesPlayerJSON is the JSON wire format for SlobberhannesPlayer.
type slobberhannesPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	// 得点と罰の内訳は必ず往復させる。Worker はリクエストごとに KV から
	// 作り直すので、抜けると得点がラウンド途中で 0 に戻る (#4478)。
	Score          int  `json:"sc"`
	TookFirstTrick bool `json:"tf"`
	TookLastTrick  bool `json:"tl"`
	TookQueen      bool `json:"tq"`
}

// MarshalJSON implements json.Marshaler.
func (p *SlobberhannesPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(slobberhannesPlayerJSON{
		GamePlayer:     p.GamePlayer,
		TrickHolder:    &p.TrickHolder,
		Score:          p.score,
		TookFirstTrick: p.tookFirstTrick,
		TookLastTrick:  p.tookLastTrick,
		TookQueen:      p.tookQueen,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SlobberhannesPlayer) UnmarshalJSON(data []byte) error {
	var j slobberhannesPlayerJSON
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
	p.tookFirstTrick = j.TookFirstTrick
	p.tookLastTrick = j.TookLastTrick
	p.tookQueen = j.TookQueen
	return nil
}
